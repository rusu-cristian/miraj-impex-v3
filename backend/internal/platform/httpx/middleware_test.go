package httpx

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

var fixedNow = time.Date(2026, 11, 18, 12, 0, 0, 0, timex.Chisinau)

func TestChainOrder(t *testing.T) {
	var got []string
	mw := func(name string) Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	base := NewChain(mw("a"), mw("b"))
	withC := base.Append(mw("c"))
	withD := base.Append(mw("d"))

	withC.ThenFunc(func(http.ResponseWriter, *http.Request) { got = append(got, "h") }).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if s := strings.Join(got, ""); s != "abch" {
		t.Errorf("order %s, want abch", s)
	}

	got = nil
	withD.ThenFunc(func(http.ResponseWriter, *http.Request) {}).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if s := strings.Join(got, ""); s != "abd" {
		t.Errorf("Append changed a sibling chain: %s", s)
	}
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestRequestID(t *testing.T) {
	tests := []struct {
		name, in string
		keep     bool
	}{
		{"passed through", "4f9a0c1e-7d2b-4e5f-8a6c-1b2d3e4f5a6b", true},
		{"proxy hex id", "a1b2c3d4e5f6", true},
		{"absent", "", false},
		{"bad characters", "x\ny <script>", false},
		{"too long", strings.Repeat("a", 129), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen string
			h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = RequestIDFrom(r.Context())
			}))
			r := httptest.NewRequest(http.MethodGet, "/api/x", nil)
			if tt.in != "" {
				r.Header.Set(HeaderRequestID, tt.in)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			echoed := w.Header().Get(HeaderRequestID)
			if echoed != seen {
				t.Errorf("echoed %q, context %q", echoed, seen)
			}
			if tt.keep && seen != tt.in {
				t.Errorf("id %q, want %q passed through", seen, tt.in)
			}
			if !tt.keep && !uuidPattern.MatchString(seen) {
				t.Errorf("generated id %q is not a random UUID", seen)
			}
		})
	}
	if RequestIDFrom(httptest.NewRequest(http.MethodGet, "/", nil).Context()) != "" {
		t.Error("RequestIDFrom without the middleware should be empty")
	}
}

// edge is the chain in the order T6 wires it, around a mux.
func edge(logs *logCapture, clock timex.Clock, mux http.Handler) http.Handler {
	l := logs.logger()
	return NewChain(RequestID(), ClientIP(nil), AccessLog(l, clock), Recover(l), SecurityHeaders()).Then(mux)
}

func TestAccessLog(t *testing.T) {
	clock := timex.FixedClock(fixedNow)
	user := uuid.MustParse("6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d")
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /api/entries/{id}", func(w http.ResponseWriter, r *http.Request) {
		SetUserID(w, user)
		clock.Advance(42 * time.Millisecond)
		JSON(w, http.StatusOK, map[string]string{"toast": "secret-response"})
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) })
	var logs logCapture
	h := edge(&logs, clock, mux)

	r := httptest.NewRequest(http.MethodPut, "/api/entries/9?token=secret-query", strings.NewReader(`{"motiv":"secret-body"}`))
	r.Header.Set("Cookie", "il_session=secret-cookie")
	r.Header.Set("Authorization", "Bearer secret-auth")
	r.Header.Set(HeaderRequestID, "req-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/readyz", nil))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/nope", nil))

	lines := logs.lines(t)
	if len(lines) != 4 {
		t.Fatalf("want 4 lines, got %v", lines)
	}
	want := map[string]any{
		"level": "INFO", "msg": "request", "request_id": "req-1", "method": "PUT", "route": "PUT /api/entries/{id}",
		"status": 200.0, "duration_ms": 42.0, "user_id": user.String(), "bytes": float64(w.Body.Len()),
	}
	first := lines[0]
	delete(first, "time")
	if len(first) != len(want) {
		t.Errorf("fields %v, want exactly %v", first, want)
	}
	for k, v := range want {
		if first[k] != v {
			t.Errorf("%s = %v, want %v", k, first[k], v)
		}
	}
	for _, l := range lines[1:3] {
		if l["level"] != "DEBUG" {
			t.Errorf("health line at %v: %v", l["level"], l)
		}
	}
	if l := lines[3]; l["level"] != "INFO" || l["status"] != 404.0 || l["route"] != "" || l["user_id"] != "" {
		t.Errorf("unmatched route line %v", l)
	}
	if out := logs.buf.String(); strings.Contains(out, "secret") {
		t.Errorf("log leaks a body, cookie, header or query: %s", out)
	}
}

func TestRecover(t *testing.T) {
	var logs logCapture
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/boom", func(http.ResponseWriter, *http.Request) { panic("nil map in farms") })
	h := edge(&logs, timex.FixedClock(fixedNow), mux)

	r := httptest.NewRequest(http.MethodGet, "/api/boom", nil)
	r.Header.Set(HeaderRequestID, "req-panic")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", w.Code)
	}
	if got, want := w.Body.String(), `{"error":{"code":"INTERNAL","message":"A apărut o eroare. Încearcă din nou."}}`+"\n"; got != want {
		t.Errorf("body %s", got)
	}
	if w.Header().Get(HeaderRequestID) != "req-panic" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("headers set before the panic are lost: %v", w.Header())
	}
	lines := logs.lines(t)
	if len(lines) != 2 {
		t.Fatalf("want a panic line and an access line, got %v", lines)
	}
	p, access := lines[0], lines[1]
	stack, _ := p["stack"].(string)
	if p["msg"] != "panic" || p["level"] != "ERROR" || p["request_id"] != "req-panic" || p["panic"] != "nil map in farms" ||
		!strings.Contains(stack, "middleware_test.go") {
		t.Errorf("panic line %v", p)
	}
	if access["status"] != 500.0 || access["level"] != "ERROR" || access["error"] != "panic: nil map in farms" ||
		access["route"] != "GET /api/boom" {
		t.Errorf("access line %v", access)
	}
}

func TestRecoverReraisesAbortHandler(t *testing.T) {
	var logs logCapture
	h := Recover(logs.logger())(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }))
	defer func() {
		if p := recover(); p != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler", p)
		}
		if logs.buf.Len() != 0 {
			t.Errorf("abort logged: %s", logs.buf.String())
		}
	}()
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRecorderKeepsFlushAndUnwrap(t *testing.T) {
	var logs logCapture
	h := AccessLog(logs.logger(), timex.FixedClock(fixedNow))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := http.NewResponseController(w).Flush(); err != nil {
			t.Errorf("Flush through the recorder: %v", err)
		}
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/live", nil))
	if !w.Flushed {
		t.Error("not flushed")
	}
}
