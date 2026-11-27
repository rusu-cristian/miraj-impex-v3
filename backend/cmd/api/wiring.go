package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
)

// app is `api serve` wired by hand: config → slog → pools (done by runServe)
// → services → handlers → routes → middleware.
type app struct {
	log             *slog.Logger
	pools           *db.Pools
	health          *health
	handler         http.Handler // the full chain: middleware around the mux
	routes          []string     // every registered pattern (routeRecorder)
	shutdownTimeout time.Duration
}

// newApp builds the services, handlers and routes on pools. It does not
// touch the database (readiness is /readyz's job). Tests call it with a
// testdb pool and a frozen clock.
func newApp(cfg config.Config, log *slog.Logger, clock timex.Clock, pools *db.Pools) (*app, error) {
	latest, err := db.LatestVersion()
	if err != nil {
		return nil, fmt.Errorf("wiring: %w", err)
	}

	// Services.
	store := db.NewStore(pools.App)
	authService := auth.NewService(store, clock)

	// Handlers and routes.
	h := newHealth(pools.App, latest, log)
	rec := newRouteRecorder()
	registerRoutes(rec, h, authService, handlers{
		auth:    auth.NewHandler(authService),
		table:   periodtable.NewHandler(periodtable.NewService(store, clock)),
		farms:   farms.NewHandler(farms.NewService(store, clock)),
		entries: entries.NewHandler(entries.NewService(store, clock)),
		reports: reports.NewHandler(reports.NewService(store, clock)),
	})

	return &app{
		log:             log,
		pools:           pools,
		health:          h,
		handler:         middleware(cfg, log, clock).Then(httpx.MuxFallback(rec.mux, "/api")),
		routes:          rec.Patterns(),
		shutdownTimeout: shutdownTimeout,
	}, nil
}

// middleware is the chain around the mux, outermost first. The order matters:
//
//  1. RequestID: first, so every later log line and every response (errors
//     included) carries the request id.
//  2. ClientIP: resolves the caller's address behind TRUSTED_PROXIES and
//     whether the client is on HTTPS (HSTS, cookies).
//  3. AccessLog: inside 1 and 2 (it reads their context values) and outside
//     everything else. 1 and 2 replace the request (r.WithContext); nothing
//     between AccessLog and the mux may, because the mux writes r.Pattern on
//     the request it receives and AccessLog reads it from its own
//     (NOTES-verify: "net/http ServeMux sets r.Pattern").
//  4. Recover: inside AccessLog, so a panic is logged as a 500 line.
//  5. SecurityHeaders: on every response, refusals and 404/405 included.
//  6. CORS (when CORS_ENABLED): answers preflights from APP_ORIGIN and adds
//     its headers before CSRF can refuse, so the browser can read a 403.
//  7. CSRF: refuses cross-site writes (the Origin rule) before any handler
//     runs.
//
// Inside the chain, httpx.MuxFallback hands the request to the mux unchanged
// (so r.Pattern still reaches AccessLog) and answers /api paths no route
// serves with the 404/405 envelope. Per-route middleware (the Content-Type
// rule, the session check, RequireRole; registerRoutes) runs inside the mux
// and may replace the request; it names the user with httpx.SetUserID.
func middleware(cfg config.Config, log *slog.Logger, clock timex.Clock) httpx.Chain {
	return httpx.NewChain(
		httpx.RequestID(),
		httpx.ClientIP(cfg.TrustedProxies),
		httpx.AccessLog(log, clock),
		httpx.Recover(log),
		httpx.SecurityHeaders(),
		httpx.CORS(cfg.AppOrigin, cfg.CORSEnabled),
		httpx.CSRF(cfg.AppOrigin),
	)
}
