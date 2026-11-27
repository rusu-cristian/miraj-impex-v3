// Command create-owner creates an owner account and prints its generated
// password once (SPEC §11.9, §6.10):
//
//	create-owner --name "Ion Rusu" --login ion.rusu [--force]
//
// It connects with DATABASE_URL (role il_app) and needs nothing else (user
// answer Q6). It refuses when an active owner exists, unless --force. The
// password goes to stdout once and nowhere else; messages go to stderr.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // embed the zone database (SPEC §11.1): the distroless image has none

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// timeout bounds the whole run (connect, hash, one transaction).
const timeout = time.Minute

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.LookupEnv, timex.RealClock{}, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run is the command; everything it needs comes in as arguments so tests can
// call it directly.
func run(ctx context.Context, args []string, lookup config.LookupFunc, clock timex.Clock, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("create-owner", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, `usage: create-owner --name "Ion Rusu" --login ion.rusu [--force]`)
		fs.PrintDefaults()
	}
	name := fs.String("name", "", "the owner's name as the app shows it (U1: required, at most 60 characters)")
	login := fs.String("login", "", "the login (U2: 3–40 of a–z, 0–9 and '.'; trimmed and lower-cased)")
	force := fs.Bool("force", false, "create the owner even when an active owner exists")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "create-owner: unexpected argument %q\n", fs.Arg(0))
		fs.Usage()
		return exitUsage
	}
	o, err := newOwnerInput(*name, *login)
	if err != nil {
		fmt.Fprintln(stderr, "create-owner:", err)
		return exitUsage
	}

	cfg, err := config.Load(lookup, config.CreateOwner)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitError
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	pc, err := db.PoolConfig(cfg.DatabaseURL, 2, nil)
	if err != nil {
		fmt.Fprintln(stderr, "create-owner: DATABASE_URL:", err)
		return exitError
	}
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		fmt.Fprintln(stderr, "create-owner: database:", err)
		return exitError
	}
	defer pool.Close()

	created, err := createOwner(ctx, db.NewStore(pool), clock, o, *force)
	switch {
	case errors.Is(err, errOwnerExists):
		fmt.Fprintln(stderr, "create-owner: an active owner already exists; pass --force to add another one")
		return exitError
	case errors.Is(err, errLoginTaken):
		fmt.Fprintf(stderr, "create-owner: the login %q is already used (deactivated accounts keep theirs)\n", o.Login)
		return exitError
	case err != nil:
		fmt.Fprintln(stderr, "create-owner:", err)
		return exitError
	}

	fmt.Fprintf(stderr, "create-owner: created owner %s, login %s (id %s)\n", o.Name, o.Login, created.ID)
	// The password is printed here once, alone on its line, and kept nowhere.
	fmt.Fprintf(stdout, "One-time password for %s (shown only now):\n%s\n", o.Login, created.Password)
	return exitOK
}
