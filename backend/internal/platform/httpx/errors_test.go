package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// moduleError stands for a module's own envelope error (e.g. a REFUSED error).
type moduleError struct{}

func (moduleError) Error() string { return "module" }
func (moduleError) HTTPError() (int, APIError) {
	return http.StatusConflict, APIError{Code: CodeStale, Message: "stale"}
}

func TestWriteErrorMapsTypedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"httpx error", NewError(http.StatusForbidden, CodeAccountDisabled, messages.AccountDisabled()), 403,
			`{"error":{"code":"ACCOUNT_DISABLED","message":"Contul tău este dezactivat. Vorbește cu administratorul."}}`},
		{"wrapped httpx error", fmt.Errorf("login: %w", Unauthenticated()), 401,
			`{"error":{"code":"UNAUTHENTICATED","message":"Sesiunea a expirat. Intră din nou în cont."}}`},
		{"forbidden", Forbidden(), 403,
			`{"error":{"code":"FORBIDDEN","message":"Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."}}`},
		{"module error", fmt.Errorf("x: %w", moduleError{}), 409, `{"error":{"code":"STALE","message":"stale"}}`},
		{"refused with nothing", Refused("Transfer refuzat.", nil, nil), 422,
			`{"error":{"code":"REFUSED","message":"Transfer refuzat.","refusals":[],"warnings":[]}}`},
		{"refused with a refusal", Refused("R", []Refusal{{SectionCode: "H2", Message: "r"}}, nil), 422,
			`{"error":{"code":"REFUSED","message":"R","refusals":[{"sectionCode":"H2","message":"r"}],"warnings":[]}}`},
		{"max bytes", fmt.Errorf("read: %w", &http.MaxBytesError{Limit: 10}), 413,
			`{"error":{"code":"PAYLOAD_TOO_LARGE","message":"Datele trimise sunt prea mari."}}`},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), 503,
			`{"error":{"code":"UNAVAILABLE","message":"Serviciul nu răspunde acum. Încearcă din nou."}}`},
		{"validation", Validation(Issue{Path: "login", Code: IssueRequired, Message: messages.LoginRequired()}), 400,
			`{"error":{"code":"VALIDATION","message":"Scrie loginul.","issues":[{"path":"login","code":"required","message":"Scrie loginul."}]}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.err)
			if w.Code != tt.status {
				t.Errorf("status %d want %d", w.Code, tt.status)
			}
			if got := strings.TrimSuffix(w.Body.String(), "\n"); got != tt.body {
				t.Errorf("body %s\nwant %s", got, tt.body)
			}
			assertJSONHeaders(t, w.Header())
		})
	}
}

func TestWriteErrorRateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, RateLimited(messages.LoginRateLimited(), 217))
	if w.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") != "217" {
		t.Errorf("got %d Retry-After %q", w.Code, w.Header().Get("Retry-After"))
	}
	want := `{"error":{"code":"RATE_LIMITED","message":"Prea multe încercări. Încearcă din nou peste 5 minute.","retryAfterSeconds":217}}` + "\n"
	if w.Body.String() != want {
		t.Errorf("body %s", w.Body)
	}
}

// logCapture is a JSON logger writing to a buffer, and its lines.
type logCapture struct{ buf bytes.Buffer }

func (c *logCapture) logger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&c.buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func (c *logCapture) lines(t *testing.T) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(c.buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("log line %q: %v", l, err)
		}
		out = append(out, m)
	}
	return out
}

func TestWriteErrorUnknownIs500AndLogged(t *testing.T) {
	var logs logCapture
	h := NewChain(RequestID(), AccessLog(logs.logger(), timex.FixedClock(fixedNow))).ThenFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, fmt.Errorf("load farm: %w", errors.New("pool closed at 10.0.0.5")))
		})
	r := httptest.NewRequest(http.MethodGet, "/api/x", nil)
	r.Header.Set(HeaderRequestID, "req-500")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", w.Code)
	}
	if got, want := w.Body.String(), `{"error":{"code":"INTERNAL","message":"A apărut o eroare. Încearcă din nou."}}`+"\n"; got != want {
		t.Errorf("body %s want %s", got, want)
	}
	lines := logs.lines(t)
	if len(lines) != 1 {
		t.Fatalf("log lines %v", lines)
	}
	l := lines[0]
	if l["level"] != "ERROR" || l["request_id"] != "req-500" || l["error"] != "load farm: pool closed at 10.0.0.5" || l["status"] != 500.0 {
		t.Errorf("log line %v", l)
	}
}

func TestWriteErrorNamesConstraintInLog(t *testing.T) {
	var logs logCapture
	h := NewChain(RequestID(), AccessLog(logs.logger(), timex.FixedClock(fixedNow))).ThenFunc(
		func(w http.ResponseWriter, r *http.Request) {
			WriteError(w, fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505", ConstraintName: "app_user_login_uq",
				Detail: "Key (login)=(ceban) already exists."}))
		})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/x", nil))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", w.Code)
	}
	l := logs.lines(t)[0]
	msg, _ := l["error"].(string)
	if !strings.Contains(msg, "23505 on app_user_login_uq") || strings.Contains(logs.buf.String(), "ceban") {
		t.Errorf("log line %v", l)
	}
}

func TestWriteErrorAfterHeadersOnlyLogs(t *testing.T) {
	var logs logCapture
	h := NewChain(AccessLog(logs.logger(), timex.FixedClock(fixedNow))).ThenFunc(
		func(w http.ResponseWriter, r *http.Request) {
			JSON(w, http.StatusOK, map[string]int{"a": 1})
			WriteError(w, errors.New("late failure"))
		})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/x", nil))
	if w.Code != http.StatusOK || w.Body.String() != `{"a":1}`+"\n" {
		t.Errorf("response changed: %d %s", w.Code, w.Body)
	}
}

func TestValidator(t *testing.T) {
	var v Validator
	if !v.Valid() || v.Err() != nil {
		t.Fatal("empty validator should be valid")
	}
	// A Validate method checks fields in document order; the issues keep it.
	v.Check(false, "login", IssueRequired, messages.LoginRequired())
	v.Check(true, "x", IssueRequired, "never")
	v.Add("sections[0].F.mortality", IssueNotInteger, "a")
	v.Add("sections[1].M.culls", IssueOutOfRange, "b")
	v.Check(false, "password", IssueRequired, messages.PasswordRequired())

	var e *Error
	if !errors.As(v.Err(), &e) {
		t.Fatalf("Err %T", v.Err())
	}
	if e.Status != 400 || e.Code != CodeValidation || e.Message != "Scrie loginul." {
		t.Errorf("got %d %s %q", e.Status, e.Code, e.Message)
	}
	var paths []string
	for _, is := range e.Issues {
		paths = append(paths, is.Path)
	}
	if got := strings.Join(paths, " "); got != "login sections[0].F.mortality sections[1].M.culls password" {
		t.Errorf("issue order %s", got)
	}
}

type loginBody struct{ Login, Password string }

func (b loginBody) Validate(v *Validator) {
	v.Check(b.Login != "", "login", IssueRequired, messages.LoginRequired())
	v.Check(b.Password != "", "password", IssueRequired, messages.PasswordRequired())
}

func TestValidate(t *testing.T) {
	if err := Validate(loginBody{"a", "b"}); err != nil {
		t.Errorf("valid body: %v", err)
	}
	var e *Error
	if !errors.As(Validate(loginBody{}), &e) || len(e.Issues) != 2 || e.Message != "Scrie loginul." {
		t.Errorf("empty body: %+v", e)
	}
	if !errors.As(Validate(loginBody{Login: "a"}), &e) || len(e.Issues) != 1 || e.Message != "Scrie parola." {
		t.Errorf("no password: %+v", e)
	}
}
