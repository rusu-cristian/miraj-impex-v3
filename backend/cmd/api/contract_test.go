package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The HTTP contract of the M0 routes (SPEC §11.8 "HTTP contract", §6.16,
// §7.0, §7.0a, §7.1, §11.4; plan T7). Every request goes over a socket to the
// real `api serve` handler (newApp: the whole middleware chain, the mux and
// the auth module) behind an httptest.Server, on a fresh test database with
// the frozen clock. Every /api response is checked for Cache-Control:
// no-store and the echoed X-Request-ID, and every JSON response for null
// lists (stack.do).

// ---------------------------------------------------------------------------
// Harness (shared with matrix_test.go)
// ---------------------------------------------------------------------------

const testPassword = "parola-de-test"

// passwordHash is the argon2id hash of testPassword, computed once per
// process (argon2 under -race is slow). Hashing doesn't touch the database.
var passwordHash = sync.OnceValues(func() (string, error) {
	svc := auth.NewService(db.NewStore(nil), timex.FixedClock(testdb.Now()))
	return svc.HashPassword(context.Background(), testPassword)
})

// loopback are the addresses the httptest client connects from; trusting
// them makes the test client a reverse proxy.
var loopback = []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8"), netip.MustParsePrefix("::1/128")}

// withCORS is the cross-origin deployment of SPEC §1.
func withCORS(c *config.Config) { c.CORSEnabled = true }

// behindProxy trusts the test client as a proxy (TRUSTED_PROXIES), so its
// X-Forwarded-For names the client.
func behindProxy(c *config.Config) { c.TrustedProxies = loopback }

// stack is the full server on its own test database.
type stack struct {
	d   *testdb.DB
	app *app
	srv *httptest.Server
	ids atomic.Int64 // X-Request-ID counter
}

// newStack wires newApp on a fresh database (frozen clock, APP_ORIGIN
// testOrigin, CORS off, no trusted proxies unless opts say otherwise) and
// serves it over HTTP.
func newStack(t *testing.T, opts ...func(*config.Config)) *stack {
	t.Helper()
	d := testdb.New(t)
	cfg := config.Config{Env: config.Dev, AppOrigin: testOrigin, LogFormat: config.FormatJSON}
	for _, o := range opts {
		o(&cfg)
	}
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	a, err := newApp(cfg, log, d.Clock, &db.Pools{App: d.App, Maint: d.Pool(testdb.RoleMaint)})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(a.handler)
	t.Cleanup(srv.Close)
	return &stack{d: d, app: a, srv: srv}
}

// call is one request. Writes (POST) get Origin: APP_ORIGIN and
// Content-Type: application/json by default; every request gets a fresh
// X-Request-ID. header is applied last; an empty value removes the header.
type call struct {
	method, path, body string
	cookie             string // il_session value; "" sends no cookie
	header             map[string]string
}

type response struct {
	status int
	header http.Header
	body   string // as received, trailing newline included
}

// text is the body without the encoder's trailing newline.
func (r response) text() string { return strings.TrimSuffix(r.body, "\n") }

func (r response) String() string { return fmt.Sprintf("%d %s", r.status, r.text()) }

// do sends c and checks the contract every response must meet: on /api,
// Cache-Control: no-store and the X-Request-ID echoed (a UUID generated when
// none was sent; a test that overrides the header checks other values
// itself); for a JSON body, no null list (SPEC §7.0a).
func (s *stack) do(t *testing.T, c call) response {
	t.Helper()
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	req, err := http.NewRequest(c.method, s.srv.URL+c.path, body)
	if err != nil {
		t.Fatal(err)
	}
	if c.method == http.MethodPost {
		req.Header.Set("Origin", testOrigin)
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Request-ID", fmt.Sprintf("t7-%d", s.ids.Add(1)))
	if c.cookie != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: c.cookie})
	}
	for k, v := range c.header {
		if v == "" {
			req.Header.Del(k)
		} else {
			req.Header.Set(k, v)
		}
	}
	res, err := s.srv.Client().Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", c.method, c.path, err)
	}
	b, err := io.ReadAll(res.Body)
	_ = res.Body.Close()
	if err != nil {
		t.Fatalf("%s %s: read body: %v", c.method, c.path, err)
	}
	r := response{status: res.StatusCode, header: res.Header, body: string(b)}

	what := c.method + " " + c.path
	if strings.HasPrefix(c.path, "/api/") {
		if cc := r.header.Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s → %d: Cache-Control %q, want no-store", what, r.status, cc)
		}
		sent, got := req.Header.Get("X-Request-ID"), r.header.Get("X-Request-ID")
		switch {
		case strings.HasPrefix(sent, "t7-") && got != sent: // ours; a test that sends its own checks it
			t.Errorf("%s → %d: X-Request-ID %q, want the sent %q echoed", what, r.status, got, sent)
		case sent == "" && !uuidPattern.MatchString(got):
			t.Errorf("%s → %d: generated X-Request-ID %q is not a UUID", what, r.status, got)
		}
	}
	if strings.HasPrefix(r.header.Get("Content-Type"), "application/json") {
		if nulls, err := nullLists(r.body); err != nil {
			t.Errorf("%s → %d: JSON body %q: %v", what, r.status, r.body, err)
		} else if len(nulls) > 0 {
			t.Errorf("%s → %d: null where a list (or other non-nullable value) belongs at %v: %s", what, r.status, nulls, r.body)
		}
	}
	return r
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// nullableFields are the response paths (array indexes written "[]") that
// may hold a JSON null scalar (SPEC §7.0a: "endDate": null, …). JSON alone
// can't tell a null list from a null scalar, so nullLists reports every null
// that is not listed here. No M0 response has a nullable field; later
// milestones add theirs.
var nullableFields = map[string]bool{}

var arrayIndex = regexp.MustCompile(`\[\d+\]`)

// nullLists returns the path of every null in the JSON document body that
// nullableFields doesn't allow ("farms", "error.issues", "[2].x").
func nullLists(body string) ([]string, error) {
	var v any
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return nil, err
	}
	var out []string
	var walk func(v any, path string)
	walk = func(v any, path string) {
		switch v := v.(type) {
		case nil:
			if !nullableFields[arrayIndex.ReplaceAllString(path, "[]")] {
				out = append(out, path)
			}
		case map[string]any:
			for k, e := range v {
				p := k
				if path != "" {
					p = path + "." + k
				}
				walk(e, p)
			}
		case []any:
			for i, e := range v {
				walk(e, fmt.Sprintf("%s[%d]", path, i))
			}
		}
	}
	walk(v, "")
	slices.Sort(out)
	return out, nil
}

// exec runs fixture SQL as il_app.
func (s *stack) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := s.d.App.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

// farm inserts a laying farm.
func (s *stack) farm(t *testing.T, name, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := s.d.App.QueryRow(context.Background(),
		`insert into farm (name, name_key, type) values ($1, $2, 'laying') returning id`, name, key).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// user inserts an account whose password is testPassword, with its farms.
func (s *stack) user(t *testing.T, login, name string, role dbq.UserRole, active bool, farms ...uuid.UUID) uuid.UUID {
	t.Helper()
	hash, err := passwordHash()
	if err != nil {
		t.Fatal(err)
	}
	var id uuid.UUID
	if err := s.d.App.QueryRow(context.Background(),
		`insert into app_user (name, login, password_hash, role, active) values ($1, $2, $3, $4, $5) returning id`,
		name, login, hash, role, active).Scan(&id); err != nil {
		t.Fatal(err)
	}
	for _, f := range farms {
		s.exec(t, `insert into user_farm (user_id, farm_id) values ($1, $2)`, id, f)
	}
	return id
}

type sessionState int

const (
	sessionOpen    sessionState = iota // seen now, expires in 30 days
	sessionRevoked                     // revoked a minute ago
	sessionExpired                     // expired a minute ago
)

// session stores a session for userID as login would (SPEC §11.4: 32 random
// bytes, base64url in the cookie, sha256 in the database) and returns the
// cookie value. Times come from the frozen clock.
func (s *stack) session(t *testing.T, userID uuid.UUID, state sessionState) string {
	t.Helper()
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	sum := sha256.Sum256(raw)
	now := s.d.Clock.Now()
	expires, revoked := now.Add(auth.SessionLifetime), (*time.Time)(nil)
	switch state {
	case sessionRevoked:
		at := now.Add(-time.Minute)
		revoked = &at
	case sessionExpired:
		expires = now.Add(-time.Minute)
	}
	s.exec(t, `insert into user_session (user_id, token_hash, created_at, last_seen_at, expires_at, revoked_at)
		values ($1, $2, $3, $3, $4, $5)`, userID, sum[:], now, expires, revoked)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func loginBody(login, password string) string {
	b, err := json.Marshal(map[string]string{"login": login, "password": password})
	if err != nil {
		panic(err)
	}
	return string(b)
}

// login posts credentials, from client (X-Forwarded-For) when not empty.
func (s *stack) login(t *testing.T, login, password, client string) response {
	t.Helper()
	c := call{method: http.MethodPost, path: "/api/auth/login", body: loginBody(login, password)}
	if client != "" {
		c.header = map[string]string{"X-Forwarded-For": client}
	}
	return s.do(t, c)
}

// loginAttempts counts login_attempt rows.
func (s *stack) loginAttempts(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.d.App.QueryRow(context.Background(), `select count(*) from login_attempt`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// wantBody fails unless r has status and exactly body (trailing newline aside).
func wantBody(t *testing.T, what string, r response, status int, body string) {
	t.Helper()
	if r.status != status || r.text() != body {
		t.Errorf("%s:\n got %s\nwant %d %s", what, r, status, body)
	}
}

// The envelopes of SPEC §6.16 and §7.0, verbatim.
const (
	bodyLoginFailed  = `{"error":{"code":"UNAUTHENTICATED","message":"Login sau parolă greșită."}}`
	bodyDisabled     = `{"error":{"code":"ACCOUNT_DISABLED","message":"Contul tău este dezactivat. Vorbește cu administratorul."}}`
	bodyNoSession    = `{"error":{"code":"UNAUTHENTICATED","message":"Sesiunea a expirat. Intră din nou în cont."}}`
	bodyForbidden    = `{"error":{"code":"FORBIDDEN","message":"Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."}}`
	bodyTooLarge     = `{"error":{"code":"PAYLOAD_TOO_LARGE","message":"Datele trimise sunt prea mari."}}`
	rateLimitMessage = "Prea multe încercări. Încearcă din nou peste 5 minute."
)

func bodyRateLimited(seconds int) string {
	return fmt.Sprintf(`{"error":{"code":"RATE_LIMITED","message":%q,"retryAfterSeconds":%d}}`, rateLimitMessage, seconds)
}

// wantRateLimited checks a 429 with retryAfterSeconds and Retry-After.
func wantRateLimited(t *testing.T, what string, r response, seconds int) {
	t.Helper()
	wantBody(t, what, r, http.StatusTooManyRequests, bodyRateLimited(seconds))
	if got := r.header.Get("Retry-After"); got != fmt.Sprint(seconds) {
		t.Errorf("%s: Retry-After %q, want %d", what, got, seconds)
	}
}

// ---------------------------------------------------------------------------
// §6.16 sign-in rows
// ---------------------------------------------------------------------------

func TestSignInErrorRows(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	hirb := s.farm(t, "Hîrbovăț", "hirbovat")
	s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true, hirb)
	s.user(t, "fost.vet", "Fost Veterinar", dbq.UserRoleVet, false, hirb)

	for _, tc := range []struct {
		name            string
		login, password string
		status          int
		body            string
	}{
		{"both fields empty", "", "", 400,
			`{"error":{"code":"VALIDATION","message":"Scrie loginul.","issues":[` +
				`{"path":"login","code":"required","message":"Scrie loginul."},` +
				`{"path":"password","code":"required","message":"Scrie parola."}]}}`},
		{"login only spaces", "   ", testPassword, 400,
			`{"error":{"code":"VALIDATION","message":"Scrie loginul.","issues":[{"path":"login","code":"required","message":"Scrie loginul."}]}}`},
		{"password empty", "ceban", "", 400,
			`{"error":{"code":"VALIDATION","message":"Scrie parola.","issues":[{"path":"password","code":"required","message":"Scrie parola."}]}}`},
		{"unknown login", "nimeni", testPassword, 401, bodyLoginFailed},
		{"login that can't exist (no hint)", "ce ban!", testPassword, 401, bodyLoginFailed},
		{"wrong password", "ceban", "parola-gresita", 401, bodyLoginFailed},
		{"deactivated, correct password", "fost.vet", testPassword, 403, bodyDisabled},
		{"deactivated, wrong password (no probing)", "fost.vet", "parola-gresita", 401, bodyLoginFailed},
	} {
		r := s.login(t, tc.login, tc.password, "")
		wantBody(t, tc.name, r, tc.status, tc.body)
		if len(r.header.Values("Set-Cookie")) != 0 {
			t.Errorf("%s: refused sign-in sets a cookie: %q", tc.name, r.header.Values("Set-Cookie"))
		}
	}
	// The empty-field 400s are not recorded (SPEC §11.4, Q3); the other five are.
	if n := s.loginAttempts(t); n != 5 {
		t.Errorf("login_attempt has %d rows, want 5", n)
	}
	// The sign-in still works: none of the above reached 5 failures for ceban.
	if r := s.login(t, "  CEBAN ", testPassword, ""); r.status != http.StatusOK {
		t.Errorf("trimmed, upper-case login: %s, want 200", r)
	}
}

// ---------------------------------------------------------------------------
// Throttling (§6.16, §11.4, Q3)
// ---------------------------------------------------------------------------

func TestSignInThrottlePerLogin(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)
	s.user(t, "ionescu", "Ion Ionescu", dbq.UserRoleManager, true)

	for i := 1; i <= 5; i++ {
		wantBody(t, fmt.Sprintf("failure %d", i), s.login(t, "ceban", "parola-gresita", ""), 401, bodyLoginFailed)
	}
	// Locked for 5 minutes from the latest failure, correct password or not.
	wantRateLimited(t, "6th attempt, correct password", s.login(t, "ceban", testPassword, ""), 300)
	wantRateLimited(t, "6th attempt, wrong password", s.login(t, "ceban", "parola-gresita", ""), 300)
	// The lock is per login: another account from the same IP signs in.
	if r := s.login(t, "ionescu", testPassword, ""); r.status != http.StatusOK {
		t.Errorf("other login while ceban is locked: %s, want 200", r)
	}

	s.d.Clock.Advance(5*time.Minute - time.Second)
	wantRateLimited(t, "1 s before the lock ends", s.login(t, "ceban", testPassword, ""), 1)
	s.d.Clock.Advance(time.Second)
	if r := s.login(t, "ceban", testPassword, ""); r.status != http.StatusOK {
		t.Errorf("5 min after the latest failure: %s, want 200", r)
	}
}

func TestSignInThrottlePerIP(t *testing.T) {
	t.Parallel()
	s := newStack(t, behindProxy)
	s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)
	const attacker, other = "203.0.113.9", "198.51.100.4"

	// 20 failures from one client, spread over logins (one each, so no login
	// reaches its own limit of 5).
	for i := 1; i <= 20; i++ {
		login := fmt.Sprintf("intrus%02d", i)
		wantBody(t, "failure "+login, s.login(t, login, "parola-gresita", attacker), 401, bodyLoginFailed)
	}
	wantRateLimited(t, "21st attempt from the client", s.login(t, "ceban", testPassword, attacker), 300)
	// The lock is per IP: ceban signs in from elsewhere.
	if r := s.login(t, "ceban", testPassword, other); r.status != http.StatusOK {
		t.Errorf("from another IP: %s, want 200", r)
	}

	s.d.Clock.Advance(5 * time.Minute)
	if r := s.login(t, "ceban", testPassword, attacker); r.status != http.StatusOK {
		t.Errorf("5 min after the latest failure: %s, want 200", r)
	}
}

// Every login_attempt.ip is the client behind the trusted proxy (SPEC §11.4),
// never the proxy and never an address the client wrote itself.
func TestLoginAttemptIPBehindTrustedProxy(t *testing.T) {
	t.Parallel()
	const client = "198.51.100.23"

	t.Run("trusted proxy", func(t *testing.T) {
		t.Parallel()
		s := newStack(t, behindProxy)
		s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)
		s.user(t, "fost.vet", "Fost Veterinar", dbq.UserRoleVet, false)
		for _, tc := range []struct{ login, password, xff string }{
			{"ceban", "parola-gresita", client},                // failure
			{"ceban", testPassword, "6.6.6.6, " + client},      // success; the left entry is the client's own spoof
			{"fost.vet", testPassword, client + ", 127.0.0.1"}, // deactivated; two proxies
			{"x", "y", client},                                 // a login that can't exist
		} {
			s.login(t, tc.login, tc.password, tc.xff)
		}
		if got := s.attemptIPs(t); !slices.Equal(got, []string{client, client, client, client}) {
			t.Errorf("login_attempt.ip = %v, want %s on every row", got, client)
		}
	})

	t.Run("untrusted peer", func(t *testing.T) {
		t.Parallel()
		s := newStack(t) // no TRUSTED_PROXIES: X-Forwarded-For is ignored
		s.login(t, "x", "y", client)
		peer := s.srv.Listener.Addr().(*net.TCPAddr).IP.String()
		if got := s.attemptIPs(t); !slices.Equal(got, []string{peer}) {
			t.Errorf("login_attempt.ip = %v, want the peer %s", got, peer)
		}
	})
}

func (s *stack) attemptIPs(t *testing.T) []string {
	t.Helper()
	rows, err := s.d.App.Query(context.Background(), `select coalesce(host(ip), '') from login_attempt order by id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			t.Fatal(err)
		}
		out = append(out, ip)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// ---------------------------------------------------------------------------
// Session cookie and the §7.1 shapes
// ---------------------------------------------------------------------------

var (
	sessionCookiePattern = regexp.MustCompile(`^il_session=([A-Za-z0-9_-]{43}); Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax$`)
	clearedCookie        = "il_session=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax"
)

func TestSessionCookieLifecycle(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	hirb := s.farm(t, "Hîrbovăț", "hirbovat")
	s.farm(t, "Telița", "telita")
	id := s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true, hirb)

	// Login: 200 { user, farms, today } (§7.1, D13) and the cookie of §11.4.
	r := s.login(t, "ceban", testPassword, "")
	wantBody(t, "login", r, 200, fmt.Sprintf(`{"user":{"id":"%s","name":"Dr. Ceban","login":"ceban","role":"vet"},`+
		`"farms":[{"id":"%s","name":"Hîrbovăț","type":"laying"}],"today":"2026-11-18"}`, id, hirb))
	if ct := r.header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("login Content-Type %q", ct)
	}
	cookies := r.header.Values("Set-Cookie")
	if len(cookies) != 1 || !sessionCookiePattern.MatchString(cookies[0]) {
		t.Fatalf("login Set-Cookie %q, want one il_session=<43 base64url chars>; Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax", cookies)
	}
	token := sessionCookiePattern.FindStringSubmatch(cookies[0])[1]

	// Me with that cookie: the §7.1 shape.
	r = s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token})
	wantBody(t, "me", r, 200, fmt.Sprintf(`{"user":{"id":"%s","name":"Dr. Ceban","login":"ceban","role":"vet"},`+
		`"farms":[{"id":"%s","name":"Hîrbovăț","type":"laying"}],"today":"2026-11-18",`+
		`"now":"2026-11-18T12:00:00+02:00","startPage":"/acasa"}`, id, hirb))

	// Logout: 204, no body, the cookie cleared.
	r = s.do(t, call{method: http.MethodPost, path: "/api/auth/logout", cookie: token})
	wantBody(t, "logout", r, http.StatusNoContent, "")
	if got := r.header.Values("Set-Cookie"); !slices.Equal(got, []string{clearedCookie}) {
		t.Errorf("logout Set-Cookie %q, want %q", got, clearedCookie)
	}

	// The old cookie is dead.
	wantBody(t, "me after logout", s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token}), 401, bodyNoSession)
	wantBody(t, "logout again", s.do(t, call{method: http.MethodPost, path: "/api/auth/logout", cookie: token}), 401, bodyNoSession)
}

// Lists are [] even when empty (SPEC §7.0a): a user with no farms.
func TestEmptyListsAreArrays(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	s.user(t, "observator", "Observator Nou", dbq.UserRoleViewer, true)

	r := s.login(t, "observator", testPassword, "")
	if r.status != 200 || !strings.Contains(r.body, `"farms":[]`) {
		t.Fatalf("login: %s, want 200 with \"farms\":[]", r)
	}
	token := sessionCookiePattern.FindStringSubmatch(r.header.Get("Set-Cookie"))[1]
	r = s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token})
	if r.status != 200 || !strings.Contains(r.body, `"farms":[]`) {
		t.Errorf("me: %s, want 200 with \"farms\":[]", r)
	}
}

// ---------------------------------------------------------------------------
// Decoder (§7.0a)
// ---------------------------------------------------------------------------

func TestDecoderErrors(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)

	for _, tc := range []struct {
		name, body string
		status     int
		want       string
	}{
		{"unknown field", `{"login":"ceban","password":"` + testPassword + `","remember":true}`, 400,
			`{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",` +
				`"issues":[{"path":"remember","code":"unknown_field","message":"Datele trimise nu sunt valide."}]}}`},
		{"number where a string is expected", `{"login":"ceban","password":12345678}`, 400,
			`{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",` +
				`"issues":[{"path":"password","code":"bad_type","message":"Datele trimise nu sunt valide."}]}}`},
		{"body over 64 KiB", `{"login":"ceban","password":"` + strings.Repeat("a", 64<<10) + `"}`, 413, bodyTooLarge},
	} {
		wantBody(t, tc.name, s.do(t, call{method: http.MethodPost, path: "/api/auth/login", body: tc.body}), tc.status, tc.want)
	}
	// Refused before the service: nothing was recorded or checked.
	if n := s.loginAttempts(t); n != 0 {
		t.Errorf("login_attempt has %d rows, want 0", n)
	}
}

// ---------------------------------------------------------------------------
// CSRF / Origin (§11.4)
// ---------------------------------------------------------------------------

func TestCSRF(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	id := s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)
	token := s.session(t, id, sessionOpen)
	// An empty sign-in: 400 VALIDATION when it reaches the handler, no hashing.
	empty := `{"login":"","password":""}`

	for _, tc := range []struct {
		name   string
		header map[string]string
		status int
	}{
		{"Origin other than APP_ORIGIN", map[string]string{"Origin": "https://evil.example"}, 403},
		{"APP_ORIGIN over another scheme", map[string]string{"Origin": "http://app.example"}, 403},
		{"no Origin, no Sec-Fetch-Site", map[string]string{"Origin": ""}, 403},
		{"no Origin, Sec-Fetch-Site cross-site", map[string]string{"Origin": "", "Sec-Fetch-Site": "cross-site"}, 403},
		{"Content-Type text/plain", map[string]string{"Content-Type": "text/plain"}, 403},
		{"no Content-Type", map[string]string{"Content-Type": ""}, 403},
		// Controls: the same request passes CSRF when it should.
		{"Origin = APP_ORIGIN", nil, 400},
		{"no Origin, Sec-Fetch-Site same-origin", map[string]string{"Origin": "", "Sec-Fetch-Site": "same-origin"}, 400},
	} {
		r := s.do(t, call{method: http.MethodPost, path: "/api/auth/login", body: empty, header: tc.header})
		if tc.status == 403 {
			wantBody(t, tc.name, r, 403, bodyForbidden)
		} else if r.status != tc.status {
			t.Errorf("%s: %s, want %d", tc.name, r, tc.status)
		}
	}

	// A refused logout never reaches the handler: the session stays open.
	r := s.do(t, call{method: http.MethodPost, path: "/api/auth/logout", cookie: token,
		header: map[string]string{"Origin": "https://evil.example"}})
	wantBody(t, "cross-site logout", r, 403, bodyForbidden)
	if r := s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token}); r.status != 200 {
		t.Errorf("me after the refused logout: %s, want 200", r)
	}
	// Safe methods are not checked.
	r = s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token,
		header: map[string]string{"Origin": "https://evil.example"}})
	if r.status != 200 {
		t.Errorf("GET me with a foreign Origin: %s, want 200", r)
	}
}

// ---------------------------------------------------------------------------
// CORS (§1)
// ---------------------------------------------------------------------------

var corsHeaders = []string{
	"Access-Control-Allow-Origin", "Access-Control-Allow-Credentials", "Access-Control-Allow-Methods",
	"Access-Control-Allow-Headers", "Access-Control-Max-Age",
}

// noCORS fails if r carries any CORS header.
func noCORS(t *testing.T, what string, r response) {
	t.Helper()
	for _, h := range corsHeaders {
		if v := r.header.Values(h); len(v) != 0 {
			t.Errorf("%s: %s: %q, want no CORS headers", what, h, v)
		}
	}
}

func preflight(origin string) call {
	return call{method: http.MethodOptions, path: "/api/auth/login", header: map[string]string{
		"Origin": origin, "Access-Control-Request-Method": "POST", "Access-Control-Request-Headers": "content-type",
	}}
}

func TestCORSEnabled(t *testing.T) {
	t.Parallel()
	s := newStack(t, withCORS)

	// A request from APP_ORIGIN: allowed origin with credentials, never *.
	r := s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"Origin": testOrigin}})
	wantBody(t, "me from APP_ORIGIN", r, 401, bodyNoSession)
	if r.header.Get("Access-Control-Allow-Origin") != testOrigin || r.header.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("me from APP_ORIGIN: CORS headers %v", r.header)
	}
	if !slices.Contains(r.header.Values("Vary"), "Origin") {
		t.Errorf("me from APP_ORIGIN: Vary %q lacks Origin", r.header.Values("Vary"))
	}

	// Its preflight: the exact §1 values.
	r = s.do(t, preflight(testOrigin))
	if r.status != http.StatusNoContent {
		t.Errorf("preflight from APP_ORIGIN: %s, want 204", r)
	}
	for h, want := range map[string]string{
		"Access-Control-Allow-Origin":      testOrigin,
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, PATCH, DELETE",
		"Access-Control-Allow-Headers":     "Content-Type, Last-Event-ID, X-Voice-Run",
		"Access-Control-Max-Age":           "600",
	} {
		if got := r.header.Get(h); got != want {
			t.Errorf("preflight from APP_ORIGIN: %s %q, want %q", h, got, want)
		}
	}

	// Any other origin: no CORS headers; its preflight is refused.
	for _, origin := range []string{"https://evil.example", "http://app.example", "https://app.example:8443"} {
		noCORS(t, "me from "+origin, s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"Origin": origin}}))
		r := s.do(t, preflight(origin))
		wantBody(t, "preflight from "+origin, r, 403, bodyForbidden)
		noCORS(t, "preflight from "+origin, r)
	}
}

func TestCORSDisabled(t *testing.T) {
	t.Parallel()
	s := newStack(t) // CORS_ENABLED=false: same-origin deployment

	noCORS(t, "me from APP_ORIGIN", s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"Origin": testOrigin}}))
	r := s.do(t, preflight(testOrigin))
	noCORS(t, "preflight from APP_ORIGIN", r)
	if r.status == http.StatusNoContent {
		t.Errorf("preflight answered 204 with CORS disabled")
	}
	r = s.login(t, "nimeni", testPassword, "")
	noCORS(t, "login from APP_ORIGIN", r)
}

// ---------------------------------------------------------------------------
// Headers on every /api response (§7.0a, §11.11)
// ---------------------------------------------------------------------------

// The success and error responses above are all checked by stack.do; this
// adds the responses no handler writes (mux 404/405) and a request without
// X-Request-ID.
func TestAPIResponseHeaders(t *testing.T) {
	t.Parallel()
	s := newStack(t)

	for _, c := range []call{
		{method: http.MethodGet, path: "/api/nu-exista"},
		{method: http.MethodGet, path: "/api/auth/login"},
		{method: http.MethodPost, path: "/api/auth/me", body: "{}"},
	} {
		r := s.do(t, c)
		if r.status != http.StatusNotFound && r.status != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: %s, want 404 or 405", c.method, c.path, r)
		}
	}
	// No X-Request-ID sent: one is generated and returned (checked in do).
	s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"X-Request-ID": ""}})
	// The proxy's id comes back as sent.
	r := s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"X-Request-ID": "caddy-4f2a:9"}})
	if got := r.header.Get("X-Request-ID"); got != "caddy-4f2a:9" {
		t.Errorf("X-Request-ID %q, want caddy-4f2a:9 echoed", got)
	}
	// An unusable one is replaced, not echoed.
	r = s.do(t, call{method: http.MethodGet, path: "/api/auth/me", header: map[string]string{"X-Request-ID": "not a valid id"}})
	if !uuidPattern.MatchString(r.header.Get("X-Request-ID")) {
		t.Errorf("invalid X-Request-ID came back as %q, want a generated UUID", r.header.Get("X-Request-ID"))
	}
}

// ---------------------------------------------------------------------------
// The null-list walker itself
// ---------------------------------------------------------------------------

func TestNullLists(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want []string
	}{
		{`{"user":{"id":"x"},"farms":[],"today":"2026-11-18"}`, nil},
		{`{"farms":null}`, []string{"farms"}},
		{`{"error":{"code":"REFUSED","refusals":null,"warnings":[]}}`, []string{"error.refusals"}},
		{`{"farms":[{"sections":null},{"sections":[]}]}`, []string{"farms[0].sections"}},
		{`[null]`, []string{"[0]"}},
		{`null`, []string{""}},
	} {
		got, err := nullLists(tc.body)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("nullLists(%s) = %q, want %q", tc.body, got, tc.want)
		}
	}
	if _, err := nullLists(`{"a":`); err == nil {
		t.Error("nullLists accepted malformed JSON")
	}
}
