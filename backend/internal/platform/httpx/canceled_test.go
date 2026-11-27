package httpx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// pgxCanceled returns the error pgx gives when ctx is canceled in the middle
// of talking to the server: it connects to a listener that never answers and
// cancels ctx once the connection is accepted (NOTES-verify "pgx and
// context.Canceled").
func pgxCanceled(t *testing.T, ctx context.Context, cancel context.CancelFunc) error {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		t.Cleanup(func() { _ = c.Close() })
		cancel() // the client goes away while pgx waits for the server
	}()
	_, err = pgconn.Connect(ctx, "postgres://u:p@"+ln.Addr().String()+"/x?sslmode=disable")
	if err == nil {
		t.Fatal("pgconn.Connect to a silent server succeeded")
	}
	return err
}

func TestCanceled(t *testing.T) {
	tests := []struct {
		name string
		// handler runs with the request; cancel cancels the request context.
		handler func(t *testing.T, r *http.Request, cancel context.CancelFunc) error
		gone    bool // the request context is done when WriteError runs
	}{
		{"ctx.Err() after the client went away", func(t *testing.T, r *http.Request, cancel context.CancelFunc) error {
			cancel()
			return r.Context().Err()
		}, true},
		{"pgx-wrapped cancel", func(t *testing.T, r *http.Request, cancel context.CancelFunc) error {
			err := pgxCanceled(t, r.Context(), cancel)
			var ce *pgconn.ConnectError
			if !errors.As(err, &ce) || !errors.Is(err, context.Canceled) {
				t.Fatalf("pgx error %T %v: want a *pgconn.ConnectError wrapping context.Canceled", err, err)
			}
			return fmt.Errorf("auth: load session: %w", err)
		}, true},
		{"internal cancel, request still live", func(t *testing.T, r *http.Request, _ context.CancelFunc) error {
			ctx, stop := context.WithCancel(r.Context())
			stop()
			return fmt.Errorf("farms: %w", ctx.Err())
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs logCapture
			mux := http.NewServeMux()
			var cancel context.CancelFunc
			mux.HandleFunc("GET /api/x", func(w http.ResponseWriter, r *http.Request) {
				WriteError(w, tt.handler(t, r, cancel))
			})
			h := edge(&logs, timex.FixedClock(fixedNow), mux)
			ctx, c := context.WithCancel(context.Background())
			defer c()
			cancel = c
			r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/x", nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			lines := logs.lines(t)
			if len(lines) != 1 {
				t.Fatalf("want one access line, got %v", lines)
			}
			l := lines[0]
			if tt.gone {
				if l["level"] != "INFO" || l["status"] != 499.0 || l["error"] != "context canceled" || l["route"] != "GET /api/x" {
					t.Errorf("access line %v, want INFO 499 with error \"context canceled\"", l)
				}
				if w.Body.Len() != 0 || w.Header().Get("Content-Type") != "" {
					t.Errorf("wrote %q (Content-Type %q) to a client that is gone", w.Body, w.Header().Get("Content-Type"))
				}
				return
			}
			if l["level"] != "ERROR" || l["status"] != 500.0 || l["error"] != "farms: context canceled" {
				t.Errorf("access line %v, want ERROR 500 with the error chain", l)
			}
			if w.Code != 500 || w.Body.String() != `{"error":{"code":"INTERNAL","message":"A apărut o eroare. Încearcă din nou."}}`+"\n" {
				t.Errorf("response %d %s, want the 500 envelope", w.Code, w.Body)
			}
		})
	}
}

// A cancellation without AccessLog (no recorder) can't be told apart: it
// stays a 500, as before.
func TestCanceledWithoutAccessLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	w := httptest.NewRecorder()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { WriteError(w, r.Context().Err()) })
	h.ServeHTTP(w, httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/x", nil))
	if w.Code != 500 {
		t.Errorf("status %d, want 500", w.Code)
	}
}
