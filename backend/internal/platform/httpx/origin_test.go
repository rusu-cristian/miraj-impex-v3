package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const appOrigin = "https://app.example"

var forbiddenBody = `{"error":{"code":"FORBIDDEN","message":"Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."}}` + "\n"

func TestCSRF(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		origin      string // "-" = absent
		fetchSite   string // "" = absent
		contentType string
		pass        bool
	}{
		{"same origin JSON", "POST", appOrigin, "same-origin", "application/json", true},
		{"Origin only", "POST", appOrigin, "", "application/json; charset=utf-8", true},
		{"same-site API origin (CORS deployment)", "PUT", appOrigin, "same-site", "application/json", true},
		{"no Origin, Sec-Fetch-Site same-origin", "DELETE", "-", "same-origin", "application/json", true},
		{"Origin differs", "POST", "https://evil.example", "cross-site", "application/json", false},
		{"Origin differs even if Sec-Fetch-Site says same-origin", "POST", "https://evil.example", "same-origin", "application/json", false},
		{"Origin differs by scheme", "POST", "http://app.example", "", "application/json", false},
		{"Origin differs by port", "POST", "https://app.example:8443", "", "application/json", false},
		{"Origin null", "POST", "null", "", "application/json", false},
		{"neither header", "POST", "-", "", "application/json", false},
		{"no Origin, Sec-Fetch-Site none", "POST", "-", "none", "application/json", false},
		{"no Origin, Sec-Fetch-Site same-site", "PATCH", "-", "same-site", "application/json", false},
		{"text/plain", "POST", appOrigin, "same-origin", "text/plain", false},
		{"form post", "POST", appOrigin, "same-origin", "application/x-www-form-urlencoded", false},
		{"multipart", "POST", appOrigin, "same-origin", "multipart/form-data; boundary=x", false},
		{"no Content-Type", "POST", appOrigin, "same-origin", "", false},
		{"GET needs nothing", "GET", "https://evil.example", "cross-site", "", true},
		{"HEAD needs nothing", "HEAD", "-", "", "", true},
		{"OPTIONS needs nothing", "OPTIONS", "-", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reached := false
			h := CSRF(appOrigin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
			r := httptest.NewRequest(tt.method, "/api/auth/login", strings.NewReader(`{}`))
			if tt.origin != "-" {
				r.Header.Set("Origin", tt.origin)
			}
			if tt.fetchSite != "" {
				r.Header.Set("Sec-Fetch-Site", tt.fetchSite)
			}
			if tt.contentType != "" {
				r.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if reached != tt.pass {
				t.Fatalf("reached handler = %v, want %v", reached, tt.pass)
			}
			if !tt.pass && (w.Code != http.StatusForbidden || w.Body.String() != forbiddenBody) {
				t.Errorf("refusal %d %s", w.Code, w.Body)
			}
		})
	}
}

func TestCSRFTwoOriginHeaders(t *testing.T) {
	reached := false
	h := CSRF(appOrigin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	r := httptest.NewRequest(http.MethodPost, "/api/x", nil)
	r.Header.Add("Origin", appOrigin)
	r.Header.Add("Origin", "https://evil.example")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if reached || w.Code != http.StatusForbidden {
		t.Errorf("two Origin headers: reached %v, status %d", reached, w.Code)
	}
}

// TestCSRFStricterThanCrossOriginProtection shows the requests that
// http.CrossOriginProtection lets through and §11.4 refuses.
func TestCSRFStricterThanCrossOriginProtection(t *testing.T) {
	cop := http.NewCrossOriginProtection()
	if err := cop.AddTrustedOrigin(appOrigin); err != nil {
		t.Fatal(err)
	}
	h := CSRF(appOrigin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	for name, set := range map[string]func(r *http.Request){
		"neither header":            func(r *http.Request) {},
		"Sec-Fetch-Site: none":      func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "none") },
		"Origin matching Host only": func(r *http.Request) { r.Host = "api.internal"; r.Header.Set("Origin", "http://api.internal") },
	} {
		r := httptest.NewRequest(http.MethodPost, "/api/x", nil)
		r.Header.Set("Content-Type", "application/json")
		set(r)
		if err := cop.Check(r); err != nil {
			t.Errorf("%s: CrossOriginProtection refused (%v); the comparison no longer holds", name, err)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusForbidden {
			t.Errorf("%s: CSRF answered %d, want 403", name, w.Code)
		}
	}
}

func TestCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { JSON(w, http.StatusOK, map[string]bool{"ok": true}) })
	corsHeaders := []string{"Access-Control-Allow-Origin", "Access-Control-Allow-Credentials", "Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers", "Access-Control-Max-Age"}
	noCORS := func(t *testing.T, h http.Header) {
		t.Helper()
		for _, k := range corsHeaders {
			if v := h.Get(k); v != "" {
				t.Errorf("%s: %q, want absent", k, v)
			}
		}
	}
	preflight := func(origin string) *http.Request {
		r := httptest.NewRequest(http.MethodOptions, "/api/entries/1", nil)
		r.Header.Set("Origin", origin)
		r.Header.Set("Access-Control-Request-Method", "PUT")
		r.Header.Set("Access-Control-Request-Headers", "content-type")
		return r
	}
	get := func(origin string) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return r
	}

	t.Run("enabled, APP_ORIGIN request", func(t *testing.T) {
		w := httptest.NewRecorder()
		CORS(appOrigin, true)(next).ServeHTTP(w, get(appOrigin))
		h := w.Header()
		if w.Code != 200 || h.Get("Access-Control-Allow-Origin") != appOrigin || h.Get("Access-Control-Allow-Credentials") != "true" ||
			h.Get("Vary") != "Origin" {
			t.Errorf("got %d %v", w.Code, h)
		}
		if h.Get("Access-Control-Max-Age") != "" || h.Get("Access-Control-Allow-Methods") != "" {
			t.Errorf("preflight headers on a plain request: %v", h)
		}
	})
	t.Run("enabled, APP_ORIGIN preflight", func(t *testing.T) {
		w := httptest.NewRecorder()
		reached := false
		CORS(appOrigin, true)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true })).ServeHTTP(w, preflight(appOrigin))
		h := w.Header()
		want := map[string]string{
			"Access-Control-Allow-Origin":      appOrigin,
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, PATCH, DELETE",
			"Access-Control-Allow-Headers":     "Content-Type, Last-Event-ID, X-Voice-Run",
			"Access-Control-Max-Age":           "600",
			"Vary":                             "Origin",
		}
		for k, v := range want {
			if h.Get(k) != v {
				t.Errorf("%s = %q, want %q", k, h.Get(k), v)
			}
		}
		if w.Code != http.StatusNoContent || reached {
			t.Errorf("preflight: status %d, reached handler %v", w.Code, reached)
		}
	})
	t.Run("enabled, other origin request", func(t *testing.T) {
		w := httptest.NewRecorder()
		CORS(appOrigin, true)(next).ServeHTTP(w, get("https://evil.example"))
		noCORS(t, w.Header())
		if w.Header().Get("Vary") != "Origin" {
			t.Errorf("Vary %q", w.Header().Get("Vary"))
		}
	})
	t.Run("enabled, other origin preflight", func(t *testing.T) {
		w := httptest.NewRecorder()
		CORS(appOrigin, true)(next).ServeHTTP(w, preflight("https://app.example.evil.example"))
		noCORS(t, w.Header())
		if w.Code != http.StatusForbidden || w.Body.String() != forbiddenBody {
			t.Errorf("got %d %s", w.Code, w.Body)
		}
	})
	t.Run("enabled, same-origin request without Origin", func(t *testing.T) {
		w := httptest.NewRecorder()
		CORS(appOrigin, true)(next).ServeHTTP(w, get(""))
		noCORS(t, w.Header())
	})
	t.Run("disabled", func(t *testing.T) {
		for _, r := range []*http.Request{get(appOrigin), preflight(appOrigin)} {
			w := httptest.NewRecorder()
			CORS(appOrigin, false)(next).ServeHTTP(w, r)
			noCORS(t, w.Header())
			if w.Header().Get("Vary") != "" || w.Code != 200 {
				t.Errorf("%s: %d %v", r.Method, w.Code, w.Header())
			}
		}
	})
}
