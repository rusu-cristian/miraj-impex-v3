package main

import (
	"net/http"
	"slices"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/prices"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/series"
)

// routeRecorder is the httpx.Router of `api serve`: a ServeMux that also
// remembers every pattern registered on it, so the §3.3 matrix test can fail
// on a route its table doesn't cover (SPEC §11.8, plan D6).
type routeRecorder struct {
	mux      *http.ServeMux
	patterns []string
}

func newRouteRecorder() *routeRecorder {
	return &routeRecorder{mux: http.NewServeMux()}
}

// Handle registers h on the mux and records pattern.
func (r *routeRecorder) Handle(pattern string, h http.Handler) {
	r.mux.Handle(pattern, h)
	r.patterns = append(r.patterns, pattern)
}

// Patterns are the registered patterns ("METHOD /path"), in registration order.
func (r *routeRecorder) Patterns() []string { return slices.Clone(r.patterns) }

// registerRoutes registers every route of `api serve`. Per-route middleware
// runs inside the mux (plan D9, D10):
//   - public: the JSON Content-Type rule (httpx.ContentType), for the one
//     /api route without a session, POST /api/auth/login;
//   - authed: public plus the session check (auth.RequireSession). Every
//     other module receives authed as its Routes chain (SPEC §12.3).
//
// Routes with another body type (the guide preview's multipart, voice STT's
// audio) build their own chain when they land.
func registerRoutes(r httpx.Router, h *health, authService *auth.Service, hs handlers) {
	// Outside /api, public, logged at debug level (SPEC §11.12).
	r.Handle("GET /healthz", http.HandlerFunc(h.live))
	r.Handle("GET /readyz", http.HandlerFunc(h.ready))

	public := httpx.NewChain(httpx.ContentType(httpx.MediaJSON))
	authed := public.Append(authService.RequireSession)
	hs.auth.Routes(r, public, authed) // POST /api/auth/login, POST /api/auth/logout, GET /api/auth/me
	hs.table.Routes(r, authed)        // GET /api/table
	hs.farms.Routes(r, authed)        // GET /api/farms/{id}/page, GET /api/farms/{id}/summary, GET /api/bootstrap
	hs.entries.Routes(r, authed)      // GET /api/forms/{feed-daily,delivery,daily}, POST /api/entries/{feed-daily,delivery,daily}, GET/PUT/DELETE /api/entries/{id}, POST /api/entries/{id}/{undo,restore}
	hs.reports.Routes(r, authed)      // GET /api/reports/feed
	hs.series.Routes(r, authed)       // GET /api/series/{id}/page
	hs.prices.Routes(r, authed)       // GET /api/deliveries, POST /api/deliveries/prices, POST /api/deliveries/prices/undo, PATCH /api/deliveries/{id}/price
}

// handlers are the module handlers registerRoutes mounts.
type handlers struct {
	auth    *auth.Handler
	table   *periodtable.Handler
	farms   *farms.Handler
	entries *entries.Handler
	reports *reports.Handler
	series  *series.Handler
	prices  *prices.Handler
}
