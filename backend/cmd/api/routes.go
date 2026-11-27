package main

import (
	"net/http"
	"slices"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
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
// (the session check) is added by each module inside the mux; the chain
// passed here is empty at M0.
func registerRoutes(r httpx.Router, h *health, authHandler *auth.Handler) {
	// Outside /api, public, logged at debug level (SPEC §11.12).
	r.Handle("GET /healthz", http.HandlerFunc(h.live))
	r.Handle("GET /readyz", http.HandlerFunc(h.ready))

	perRoute := httpx.NewChain()
	authHandler.Routes(r, perRoute) // POST /api/auth/login, POST /api/auth/logout, GET /api/auth/me
}
