package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

const testOrigin = "https://app.example"

// newTestApp wires the real app on a fresh test database (il_app and
// il_maint pools, frozen clock). Logs go to the returned buffer.
func newTestApp(t *testing.T) (*app, *testdb.DB, *logBuffer) {
	t.Helper()
	d := testdb.New(t)
	logs := &logBuffer{}
	log := slog.New(slog.NewJSONHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := config.Config{Env: config.Dev, AppOrigin: testOrigin, LogFormat: config.FormatJSON}
	a, err := newApp(cfg, log, d.Clock, &db.Pools{App: d.App, Maint: d.Pool(testdb.RoleMaint)})
	if err != nil {
		t.Fatal(err)
	}
	return a, d, logs
}

// logBuffer collects log output from several goroutines.
type logBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *logBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func wantText(t *testing.T, w *httptest.ResponseRecorder, status int, body string) {
	t.Helper()
	if w.Code != status || w.Body.String() != body {
		t.Fatalf("got %d %q, want %d %q", w.Code, w.Body.String(), status, body)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("Content-Type %q", ct)
	}
}

func TestHealthThroughChain(t *testing.T) {
	t.Parallel()
	a, _, logs := newTestApp(t)

	w := get(t, a.handler, "/healthz")
	wantText(t, w, http.StatusOK, "ok")
	if w.Header().Get("X-Request-ID") == "" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("middleware headers missing: %v", w.Header())
	}
	wantText(t, get(t, a.handler, "/readyz"), http.StatusOK, "ok")

	// Probes are logged at debug level only (SPEC §11.12).
	for line := range strings.SplitSeq(strings.TrimSpace(logs.String()), "\n") {
		if strings.Contains(line, `"msg":"request"`) && !strings.Contains(line, `"level":"DEBUG"`) {
			t.Fatalf("probe logged above debug: %s", line)
		}
	}
	// Both are registered and visible to the route recorder.
	for _, p := range []string{"GET /healthz", "GET /readyz"} {
		if !slices.Contains(a.routes, p) {
			t.Fatalf("routes %v lack %q", a.routes, p)
		}
	}
}

func TestRoutesRecorded(t *testing.T) {
	t.Parallel()
	a, _, _ := newTestApp(t)
	want := []string{"GET /healthz", "GET /readyz", "POST /api/auth/login", "POST /api/auth/logout", "GET /api/auth/me",
		"GET /api/table", "GET /api/farms/{id}/page", "GET /api/farms/{id}/summary", "GET /api/bootstrap"}
	if !slices.Equal(a.routes, want) {
		t.Fatalf("routes = %v, want %v", a.routes, want)
	}
	// The returned list is a copy.
	rec := newRouteRecorder()
	rec.Handle("GET /x", http.NotFoundHandler())
	rec.Patterns()[0] = "changed"
	if rec.Patterns()[0] != "GET /x" {
		t.Fatal("Patterns shares its slice")
	}
}

func TestReadyzOlderMigration(t *testing.T) {
	t.Parallel()
	a, d, logs := newTestApp(t)
	// The database is one migration behind the binary.
	conn, err := pgx.Connect(context.Background(), testdb.DSN(d.Name))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	latest, err := db.LatestVersion()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(context.Background(), "delete from goose_db_version where version_id = $1", latest); err != nil {
		t.Fatal(err)
	}

	wantText(t, get(t, a.handler, "/readyz"), http.StatusServiceUnavailable, "migrations")
	wantText(t, get(t, a.handler, "/healthz"), http.StatusOK, "ok")
	if !strings.Contains(logs.String(), `"msg":"not ready"`) || !strings.Contains(logs.String(), `"failing":"migrations"`) {
		t.Fatalf("no 'not ready' log line:\n%s", logs.String())
	}

	// A database ahead of the binary is not ready either.
	if _, err := conn.Exec(context.Background(),
		"insert into goose_db_version (version_id, is_applied) values ($1, true), ($2, true)", latest, latest+1); err != nil {
		t.Fatal(err)
	}
	wantText(t, get(t, a.handler, "/readyz"), http.StatusServiceUnavailable, "migrations")
}

// fakeDB is a database whose ping fails or hangs; it counts its calls.
type fakeDB struct {
	pingErr error
	hang    bool
	calls   int
}

func (f *fakeDB) Ping(ctx context.Context) error {
	f.calls++
	if f.hang {
		<-ctx.Done()
		return ctx.Err()
	}
	return f.pingErr
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	f.calls++
	return errRow{errors.New("fakeDB: no query")}
}

type errRow struct{ err error }

func (r errRow) Scan(...any) error { return r.err }

func TestReadyzDatabase(t *testing.T) {
	t.Parallel()
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, tc := range []struct {
		name string
		db   *fakeDB
	}{
		{"ping fails", &fakeDB{pingErr: errors.New("connection refused")}},
		{"ping takes over 1 s", &fakeDB{hang: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h := newHealth(tc.db, 2, discard)
			start := time.Now()
			w := httptest.NewRecorder()
			h.ready(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			wantText(t, w, http.StatusServiceUnavailable, "database")
			if el := time.Since(start); el > readyTimeout+500*time.Millisecond {
				t.Fatalf("readyz took %v", el)
			}
			if tc.db.calls != 1 {
				t.Fatalf("%d database calls, want only the ping", tc.db.calls)
			}
		})
	}
}

func TestHealthzNeverTouchesDatabase(t *testing.T) {
	t.Parallel()
	f := &fakeDB{pingErr: errors.New("down")}
	h := newHealth(f, 2, slog.New(slog.NewTextHandler(io.Discard, nil)))
	w := httptest.NewRecorder()
	h.live(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	wantText(t, w, http.StatusOK, "ok")
	if f.calls != 0 {
		t.Fatalf("healthz made %d database calls", f.calls)
	}
}

func TestReadyzDraining(t *testing.T) {
	t.Parallel()
	f := &fakeDB{}
	h := newHealth(f, 2, slog.New(slog.NewTextHandler(io.Discard, nil)))
	h.drain()
	w := httptest.NewRecorder()
	h.ready(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	wantText(t, w, http.StatusServiceUnavailable, "shutdown")
	if f.calls != 0 {
		t.Fatalf("draining readyz made %d database calls", f.calls)
	}
}
