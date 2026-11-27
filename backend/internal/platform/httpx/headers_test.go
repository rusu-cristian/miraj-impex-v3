package httpx

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	base := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "same-origin",
		"Permissions-Policy":      "microphone=(self)",
		"Content-Security-Policy": "default-src 'none'; frame-ancestors 'none'",
		"Cache-Control":           "no-store",
	}
	h := ClientIP(nil)(SecurityHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	t.Run("plain HTTP: no HSTS", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil))
		for k, v := range base {
			if w.Header().Get(k) != v {
				t.Errorf("%s = %q, want %q", k, w.Header().Get(k), v)
			}
		}
		if v := w.Header().Get("Strict-Transport-Security"); v != "" {
			t.Errorf("HSTS over plain HTTP: %q", v)
		}
	})
	t.Run("HTTPS: HSTS", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		r.TLS = &tls.ConnectionState{}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if v := w.Header().Get("Strict-Transport-Security"); v != "max-age=31536000" {
			t.Errorf("HSTS %q", v)
		}
	})
}
