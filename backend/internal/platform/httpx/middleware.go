package httpx

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Middleware wraps a handler.
type Middleware func(http.Handler) http.Handler

// Chain is an ordered list of middleware; the first one is the outermost.
// Chains are values: Append returns a new chain and never changes the
// receiver.
type Chain struct {
	mws []Middleware
}

// NewChain builds a chain; mws[0] runs first.
func NewChain(mws ...Middleware) Chain {
	return Chain{mws: append([]Middleware(nil), mws...)}
}

// Append returns a new chain with mws added after (inside) c's middleware.
func (c Chain) Append(mws ...Middleware) Chain {
	out := make([]Middleware, 0, len(c.mws)+len(mws))
	out = append(out, c.mws...)
	return Chain{mws: append(out, mws...)}
}

// Then wraps h in the chain.
func (c Chain) Then(h http.Handler) http.Handler {
	for i := len(c.mws) - 1; i >= 0; i-- {
		h = c.mws[i](h)
	}
	return h
}

// ThenFunc wraps f in the chain.
func (c Chain) ThenFunc(f http.HandlerFunc) http.Handler { return c.Then(f) }

// recorder is the ResponseWriter AccessLog hands down: it records the status,
// the bytes written, the user id (SetUserID) and the cause of a 5xx
// (WriteError, Recover) for the access-log line.
type recorder struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
	userID      string
	err         error
}

func (r *recorder) WriteHeader(code int) {
	if !r.wroteHeader && code >= 200 {
		r.status, r.wroteHeader = code, true
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *recorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.status, r.wroteHeader = http.StatusOK, true
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += int64(n)
	return n, err
}

// Flush keeps streaming responses (SSE, §9) working through the recorder.
func (r *recorder) Flush() {
	if !r.wroteHeader {
		r.status, r.wroteHeader = http.StatusOK, true
	}
	_ = http.NewResponseController(r.ResponseWriter).Flush()
}

// Unwrap lets http.ResponseController reach the server's writer.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func recorderOf(w http.ResponseWriter) *recorder {
	for {
		switch x := w.(type) {
		case *recorder:
			return x
		case interface{ Unwrap() http.ResponseWriter }:
			w = x.Unwrap()
		default:
			return nil
		}
	}
}

// SetUserID puts the signed-in user's id on the access-log line of this
// request. The auth middleware calls it once it knows the principal.
func SetUserID(w http.ResponseWriter, id uuid.UUID) {
	if rec := recorderOf(w); rec != nil {
		rec.userID = id.String()
	}
}

type ctxKey int

const (
	requestIDKey ctxKey = iota
	clientKey
)

// HeaderRequestID is the request id header, read from the proxy and echoed.
const HeaderRequestID = "X-Request-ID"

// requestIDPattern is what an incoming X-Request-ID must look like to be
// passed through (UUIDs, hex and base64 ids of proxies); anything else is
// replaced, so a client can't put arbitrary text into the logs.
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:/+=-]{1,128}$`)

// RequestID takes X-Request-ID from the request (set by the proxy) or
// generates one (a random UUID), puts it in the context (RequestIDFrom) and
// echoes it in the response header (SPEC §11.11).
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(HeaderRequestID)
			if !requestIDPattern.MatchString(id) {
				id = uuid.NewString()
			}
			w.Header().Set(HeaderRequestID, id)
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
		})
	}
}

// RequestIDFrom is the request id RequestID put in ctx, or "".
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

// AccessLog writes one line per request (SPEC §11.11): request_id, method,
// route (the matched pattern, r.Pattern), status, duration_ms, user_id and
// bytes; a 5xx adds the error chain given to WriteError and is logged at
// error level. /healthz and /readyz are logged at debug level. Bodies,
// cookies, headers and query strings are never logged.
//
// Place it inside RequestID and ClientIP and outside every middleware that
// replaces the request (r.WithContext) before the ServeMux: the mux sets
// r.Pattern on the request it receives, and AccessLog reads it from its own.
func AccessLog(logger *slog.Logger, clock timex.Clock) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := clock.Now()
			rec := &recorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			level := slog.LevelInfo
			switch {
			case r.URL.Path == "/healthz" || r.URL.Path == "/readyz":
				level = slog.LevelDebug
			case rec.status >= 500:
				level = slog.LevelError
			}
			attrs := []slog.Attr{
				slog.String("request_id", RequestIDFrom(r.Context())),
				slog.String("method", r.Method),
				slog.String("route", r.Pattern),
				slog.Int("status", rec.status),
				slog.Int64("duration_ms", clock.Now().Sub(start).Milliseconds()),
				slog.String("user_id", rec.userID),
				slog.Int64("bytes", rec.bytes),
			}
			if rec.status >= 500 && rec.err != nil {
				attrs = append(attrs, slog.String("error", rec.err.Error()))
			}
			logger.LogAttrs(r.Context(), level, "request", attrs...)
		})
	}
}

// Recover turns a panic into the 500 envelope and logs it with the stack and
// the request id. http.ErrAbortHandler is re-raised (net/http's own signal to
// abort a response).
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				p := recover()
				if p == nil {
					return
				}
				if p == http.ErrAbortHandler {
					panic(p)
				}
				logger.LogAttrs(r.Context(), slog.LevelError, "panic",
					slog.String("request_id", RequestIDFrom(r.Context())),
					slog.String("panic", fmt.Sprint(p)),
					slog.String("stack", string(debug.Stack())),
				)
				WriteError(w, fmt.Errorf("panic: %v", p))
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Router is where modules register routes: *http.ServeMux, or a wrapper that
// also records the patterns (cmd/api's route list for the §3.3 matrix).
type Router interface {
	Handle(pattern string, handler http.Handler)
}
