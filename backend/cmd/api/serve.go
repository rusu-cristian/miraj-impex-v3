package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// http.Server timeouts for ordinary requests (SPEC §9.1 caveat; recorded in
// NOTES-verify.md). The SSE and voice handlers (M11, M12) lift the deadlines
// per response.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	writeTimeout      = 60 * time.Second
	idleTimeout       = 120 * time.Second
)

// shutdownTimeout is the budget of http.Server.Shutdown (SPEC §11.12).
const shutdownTimeout = 20 * time.Second

// runServe is `api serve`: config → slog → pools → app (wiring.go) → listen,
// until ctx is cancelled (SIGINT, SIGTERM).
func runServe(ctx context.Context, args []string, lookup config.LookupFunc, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: api serve")
		return exitUsage
	}
	cfg, ok := loadConfig(lookup, config.Serve, stderr)
	if !ok {
		return exitError
	}
	log := newLogger(cfg, stdout)

	pools, err := db.Open(ctx, cfg)
	if err != nil {
		log.Error("serve: open database pools", "err", err)
		return exitError
	}
	a, err := newApp(cfg, log, timex.RealClock{}, pools)
	if err != nil {
		pools.Close()
		log.Error("serve: wiring", "err", err)
		return exitError
	}
	ln, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		pools.Close()
		log.Error("serve: listen", "addr", cfg.ListenAddr, "err", err)
		return exitError
	}
	log.Info("api serve listening", "addr", ln.Addr().String(), "config", cfg)
	if err := a.serve(ctx, ln); err != nil {
		return exitError
	}
	log.Info("api serve stopped")
	return exitOK
}

// serve answers on ln until ctx is cancelled, then shuts down in the order of
// SPEC §11.12: /readyz turns 503, http.Server.Shutdown waits up to
// shutdownTimeout for in-flight requests (and closes the listener), then the
// pools close. It always closes the pools. A Shutdown that runs out of time
// closes the remaining connections and returns the error.
func (a *app) serve(ctx context.Context, ln net.Listener) error {
	defer a.pools.Close()

	srv := &http.Server{
		Handler:           a.handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(a.log.Handler(), slog.LevelWarn),
	}
	served := make(chan error, 1)
	go func() { served <- srv.Serve(ln) }()

	select {
	case err := <-served: // Serve stopped on its own: the listener failed
		a.log.Error("serve", "err", err)
		return err
	case <-ctx.Done():
	}

	a.log.Info("shutting down", "budget", a.shutdownTimeout.String())
	a.health.drain()
	sctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.shutdownTimeout)
	defer cancel()
	err := srv.Shutdown(sctx)
	if err != nil {
		a.log.Error("shutdown: requests still running when the budget ran out", "err", err)
		_ = srv.Close()
	}
	if serr := <-served; !errors.Is(serr, http.ErrServerClosed) {
		a.log.Error("serve", "err", serr)
	}
	return err
}
