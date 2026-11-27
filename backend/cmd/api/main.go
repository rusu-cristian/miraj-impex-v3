// Command api is the Impex Line API service (SPEC §12.3):
//
//	api serve              serve the API (DATABASE_URL, MAINT_DATABASE_URL, …)
//	api migrate up|status  apply or list the embedded migrations (MIGRATE_DATABASE_URL only)
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	_ "time/tzdata" // embed the zone database (SPEC §11.1): the distroless image has none

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
)

const usage = "usage: api serve | api migrate up|status"

// Exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run dispatches the subcommand. Everything it needs comes in as arguments so
// tests can call it directly.
func run(ctx context.Context, args []string, lookup config.LookupFunc, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage)
		return exitUsage
	}
	switch args[0] {
	case "serve":
		return runServe(ctx, args[1:], lookup, stdout, stderr)
	case "migrate":
		return runMigrate(ctx, args[1:], lookup, stdout, stderr)
	default:
		fmt.Fprintln(stderr, usage)
		return exitUsage
	}
}

// loadConfig reads the configuration for cmd; on failure it writes the one
// line naming every missing or invalid variable (SPEC §11.10).
func loadConfig(lookup config.LookupFunc, cmd config.Command, stderr io.Writer) (config.Config, bool) {
	cfg, err := config.Load(lookup, cmd)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return config.Config{}, false
	}
	return cfg, true
}

// newLogger builds the process logger (SPEC §11.11): JSON in production, text
// when LOG_FORMAT=text (the dev default).
func newLogger(cfg config.Config, w io.Writer) *slog.Logger {
	opts := &slog.HandlerOptions{Level: cfg.LogLevel}
	if cfg.LogFormat == config.FormatText {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}
