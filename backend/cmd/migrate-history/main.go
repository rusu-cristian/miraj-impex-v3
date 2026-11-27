// Command migrate-history runs the one-time history migration (SPEC §10.3):
//
//	migrate-history --extracts ../design/data/extracts [--db DSN] [--replace] [--seed-dev]
//
// It reads the workbook extracts, writes every row in one transaction, checks
// the golden values and the §10.3 figures, and commits only when all pass. It
// refuses a database that already holds the migration. --db defaults to
// DATABASE_URL (role il_app). --replace (user answer Q4) and --seed-dev are
// dev only: they run only when APP_ENV is dev, and are refused for staging,
// production and an unset APP_ENV. --replace empties every domain table first
// and needs the owner role, so its --db defaults to MIGRATE_DATABASE_URL.
// --seed-dev (§11.9, M1-D7) adds the dev users and the illustrative events in
// the same transaction, and prints each dev login with its generated password
// once on stdout (never logged). Counts go to stdout; messages go to stderr.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"
	_ "time/tzdata" // embed the zone database (SPEC §11.1): the distroless image has none

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/migratehistory"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// timeout bounds the whole run (read, one transaction, verification).
const timeout = 10 * time.Minute

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.LookupEnv, timex.RealClock{}, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run is the command; everything it needs comes in as arguments so tests can
// call it directly.
func run(ctx context.Context, args []string, lookup config.LookupFunc, clock timex.Clock, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("migrate-history", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "usage: migrate-history --extracts ../design/data/extracts [--db DSN] [--replace] [--seed-dev]")
		fs.PrintDefaults()
	}
	dsn := fs.String("db", "", "connection string (default $DATABASE_URL; with --replace $MIGRATE_DATABASE_URL)")
	extracts := fs.String("extracts", "", "the extracts directory (design/data/extracts), required")
	replace := fs.Bool("replace", false, "dev only (APP_ENV=dev): empty every domain table, then import again (needs the owner role)")
	seedDev := fs.Bool("seed-dev", false, "dev only (APP_ENV=dev): add the dev users and the illustrative events; prints their passwords once")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitOK
		}
		return exitUsage
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "migrate-history: unexpected argument %q\n", fs.Arg(0))
		fs.Usage()
		return exitUsage
	}
	if *extracts == "" {
		fmt.Fprintln(stderr, "migrate-history: --extracts is required")
		fs.Usage()
		return exitUsage
	}
	env, err := appEnv(lookup)
	if err != nil {
		fmt.Fprintln(stderr, "migrate-history:", err)
		return exitUsage
	}
	for _, f := range []struct {
		on   bool
		name string
	}{{*replace, "--replace"}, {*seedDev, "--seed-dev"}} {
		// An allowlist: only dev. Staging and production hold real data, and
		// an unset APP_ENV may be either.
		if f.on && env != config.Dev {
			fmt.Fprintf(stderr, "migrate-history: %s is dev only: it runs only with APP_ENV=dev (APP_ENV is %s)\n", f.name, envName(env))
			return exitUsage
		}
	}
	if *dsn == "" {
		key := "DATABASE_URL"
		if *replace {
			key = "MIGRATE_DATABASE_URL"
		}
		v, _ := lookup(key)
		if *dsn = strings.TrimSpace(v); *dsn == "" {
			fmt.Fprintf(stderr, "migrate-history: pass --db or set %s\n", key)
			return exitUsage
		}
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	pc, err := db.PoolConfig(*dsn, 2, nil)
	if err != nil {
		fmt.Fprintln(stderr, "migrate-history: --db:", err)
		return exitUsage
	}
	// The pool connects lazily: Load reads and checks the extracts first, so a
	// bad --extracts never reaches the database.
	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		fmt.Fprintln(stderr, "migrate-history: database:", err)
		return exitError
	}
	defer pool.Close()

	counts, err := migratehistory.Load(ctx, pool, clock, migratehistory.Options{
		Extracts: *extracts, Replace: *replace, SeedDev: *seedDev, SeedOut: stdout,
	})
	switch {
	case errors.Is(err, migratehistory.ErrAlreadyMigrated):
		fmt.Fprintln(stderr, "migrate-history: this database already holds the history migration; nothing was changed (it runs once; in dev, --replace redoes it)")
		return exitError
	case err != nil:
		fmt.Fprintln(stderr, "migrate-history:", strings.TrimPrefix(err.Error(), "migratehistory: "))
		return exitError
	}

	tw := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintln(tw, "table\trows\t")
	for _, c := range counts {
		fmt.Fprintf(tw, "%s\t%d\t\n", c.Table, c.Rows)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	fmt.Fprintln(stderr, "migrate-history: imported and verified")
	return exitOK
}

// appEnv is APP_ENV with §11.10's values; "" when unset.
func appEnv(lookup config.LookupFunc) (config.Env, error) {
	v, _ := lookup("APP_ENV")
	switch env := config.Env(strings.TrimSpace(v)); env {
	case "", config.Production, config.Staging, config.Dev:
		return env, nil
	default:
		return "", errors.New("APP_ENV must be production, staging or dev")
	}
}

// envName names env in a message.
func envName(env config.Env) string {
	if env == "" {
		return "unset"
	}
	return string(env)
}
