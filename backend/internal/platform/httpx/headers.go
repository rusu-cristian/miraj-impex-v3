package httpx

import "net/http"

// SecurityHeaders sets the API's own security headers (SPEC §11.4 as decided
// in Q5): X-Content-Type-Options, Referrer-Policy, Permissions-Policy, a CSP
// that allows nothing (the API serves only JSON; the page CSP belongs to the
// proxy or front end), HSTS when the client is on HTTPS (IsHTTPS), and
// Cache-Control: no-store, so responses without a JSON body (204, mux 404/405)
// are not cached either (SPEC §7.0a: no-store on every /api response).
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "same-origin")
			h.Set("Permissions-Policy", "microphone=(self)")
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			h.Set("Cache-Control", "no-store")
			if IsHTTPS(r) {
				h.Set("Strict-Transport-Security", "max-age=31536000")
			}
			next.ServeHTTP(w, r)
		})
	}
}
