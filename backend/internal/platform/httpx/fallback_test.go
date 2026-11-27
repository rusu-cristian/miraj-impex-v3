package httpx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

const (
	notFoundBody         = `{"error":{"code":"NOT_FOUND","message":"Adresa cerută nu există."}}` + "\n"
	methodNotAllowedBody = `{"error":{"code":"METHOD_NOT_ALLOWED","message":"Operația nu este permisă la această adresă."}}` + "\n"
)

func TestMuxFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/login", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "login") })
	mux.HandleFunc("GET /api/items/{id}", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "item "+r.PathValue("id")) })
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, "ok") })
	h := MuxFallback(mux, "/api")

	tests := []struct {
		name, method, path string
		status             int
		body               string // exact; "" = don't check
		allow              string // the Allow header; "" = must be absent
		envelope           bool   // a JSON envelope (Content-Type, Cache-Control)
	}{
		{"matched route", "POST", "/api/auth/login", 200, "login", "", false},
		{"matched wildcard route", "GET", "/api/items/7", 200, "item 7", "", false},
		{"unknown /api path", "GET", "/api/nope", 404, notFoundBody, "", true},
		{"unknown /api path, POST", "POST", "/api/nope", 404, notFoundBody, "", true},
		{"/api itself", "GET", "/api", 404, notFoundBody, "", true},
		{"wrong method", "GET", "/api/auth/login", 405, methodNotAllowedBody, "POST", true},
		{"wrong method on a wildcard route", "DELETE", "/api/items/7", 405, methodNotAllowedBody, "GET, HEAD", true},
		{"unclean path keeps the mux redirect", "GET", "/api//nope", http.StatusTemporaryRedirect, "", "", false},
		{"outside /api: matched", "GET", "/healthz", 200, "ok", "", false},
		{"outside /api: mux 404", "GET", "/nope", 404, "404 page not found\n", "", false},
		{"outside /api: mux 405", "POST", "/healthz", 405, "Method Not Allowed\n", "GET, HEAD", false},
		{"prefix look-alike is outside /api", "GET", "/apix", 404, "404 page not found\n", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			if w.Code != tt.status || tt.body != "" && w.Body.String() != tt.body {
				t.Fatalf("got %d %q, want %d %q", w.Code, w.Body, tt.status, tt.body)
			}
			if got := w.Header().Get("Allow"); got != tt.allow {
				t.Errorf("Allow %q, want %q", got, tt.allow)
			}
			ct := w.Header().Get("Content-Type")
			if tt.envelope && (ct != "application/json; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store") {
				t.Errorf("envelope headers: Content-Type %q, Cache-Control %q", ct, w.Header().Get("Cache-Control"))
			}
			if !tt.envelope && strings.HasPrefix(ct, "application/json") {
				t.Errorf("Content-Type %q on a response the mux owns", ct)
			}
		})
	}
}

// The access log still sees the matched pattern through MuxFallback, and an
// unmatched /api path is logged with an empty route at INFO.
func TestMuxFallbackAccessLog(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/items/{id}", func(http.ResponseWriter, *http.Request) {})
	for path, want := range map[string]string{"/api/items/7": "GET /api/items/{id}", "/api/nope": ""} {
		var logs logCapture
		h := edge(&logs, timex.FixedClock(fixedNow), MuxFallback(mux, "/api"))
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
		lines := logs.lines(t)
		if len(lines) != 1 || lines[0]["route"] != want || lines[0]["level"] != "INFO" {
			t.Errorf("%s: lines %v, want one INFO line with route %q", path, lines, want)
		}
	}
}
