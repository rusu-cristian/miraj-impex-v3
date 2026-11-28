package main

import (
	"cmp"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The HTTP contract of the M0 routes and the M2 farm reads (SPEC §11.8 "HTTP contract", §6.16,
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
	return auth.NewHasher().Hash(context.Background(), testPassword)
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
	return newStackOn(t, testdb.New(t), opts...)
}

// newSeededStack is newStack on a copy of the seeded database: the imported
// history plus the dev seed (§11.9), with its farms and users.
func newSeededStack(t testing.TB) *stack {
	t.Helper()
	return newStackOn(t, testdb.NewSeeded(t))
}

func newStackOn(t testing.TB, d *testdb.DB, opts ...func(*config.Config)) *stack {
	t.Helper()
	cfg := config.Config{Env: config.Dev, AppOrigin: testOrigin, AppSecret: testAppSecret, LogFormat: config.FormatJSON}
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

// call is one request. Writes (POST, PUT, PATCH, DELETE) get Origin:
// APP_ORIGIN and Content-Type: application/json by default; every request
// gets a fresh X-Request-ID. header is applied last; an empty value removes
// the header.
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
func (s *stack) do(t testing.TB, c call) response {
	t.Helper()
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	req, err := http.NewRequest(c.method, s.srv.URL+c.path, body)
	if err != nil {
		t.Fatal(err)
	}
	switch c.method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
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
// that is not listed here. A path ending in ".*" allows a null under any key
// of that object (the PeriodTable's cells, keyed "alive.F", "egg.layPct", …:
// a null cell is "–"). The paths are those of every route's DTO, whichever
// route returns them; no M0 response has one.
var nullableFields = map[string]bool{
	// GET /api/table (§7.4): the load-more row and the start note; cells.
	"more":                   true,
	"start":                  true,
	"weeks[].cells.*":        true,
	"weeks[].days[].cells.*": true,
	"totals.cells.*":         true,

	// GET /api/farms/{id}/page (§7.3): no cycle, no flock count, no guide.
	"farm.guide":                     true,
	"cycle":                          true,
	"cycle.endDate":                  true,
	"alive":                          true,
	"track.lanes[].cycles[].endDate": true,

	// GET /api/bootstrap (§7.2).
	"farms[].guide":      true,
	"farms[].openCycle":  true,
	"feedTypes[].forSex": true,

	// GET /api/farms/{id}/summary (§7.6): the readout and its series; the
	// stats' figures without a base, guide or weighing; the tiles' g on an
	// unfed day, weighing and lay % without an entry.
	"readout":                         true,
	"readout.guide":                   true,
	"readout.series[].value":          true,
	"readout.series[].guide":          true,
	"stats[].F":                       true,
	"stats[].M":                       true,
	"stats[].guide":                   true,
	"stats[].guideF":                  true,
	"stats[].guideM":                  true,
	"stats[].at":                      true,
	"stats[].section":                 true,
	"stats[].value":                   true,
	"stats[].diffPts":                 true,
	"tiles[].sections[].gToday.F":     true,
	"tiles[].sections[].gToday.M":     true,
	"tiles[].sections[].weight":       true,
	"tiles[].sections[].weight.guide": true,
	"tiles[].sections[].layHall":      true,

	// The summary's counts on a cycle without a flock count (M3 Q10); its
	// feedDaysLeft items of a type for both sexes, or without consumption
	// (M3 Q7).
	"stats[].arrivals.F":     true,
	"stats[].arrivals.M":     true,
	"stats[].sold.F":         true,
	"stats[].sold.M":         true,
	"stats[].deaths.F":       true,
	"stats[].deaths.M":       true,
	"stats[].items[].forSex": true,
	"stats[].items[].days":   true,
	"stats[].items[].runOut": true,

	// GET /api/table?view=report-feed (§7.4, M3 Q2): the Consum section week
	// rows' cells, like the week rows'.
	"weeks[].sections[].cells.*": true,

	// GET /api/forms/feed-daily (§7.8, M3 D8, Q1): no cycle on the date, no
	// guide (or no guide row for a sex), nothing locked; a row without a
	// guide, a suggestion or a saved allocation (a migrated kg-only one has
	// no g); a migration entry has no author.
	"sapt":                         true,
	"ageDays":                      true,
	"guide":                        true,
	"guide.F":                      true,
	"guide.M":                      true,
	"lockedReason":                 true,
	"rows[].guideG":                true,
	"rows[].suggestion.g":          true,
	"rows[].suggestion.feedTypeId": true,
	"rows[].existing":              true,
	"rows[].existing.g":            true,
	"existing[].author":            true,

	// The stock rows of the form state and of GET /api/reports/feed (§5.8,
	// §7.12, M3 D7): a type for both sexes; no consumption in the window (no
	// days left, no run-out).
	"stock[].forSex":   true,
	"stock[].daysLeft": true,
	"stock[].runOut":   true,

	// GET /api/reports/feed Livrări and Cost (§7.12, M4 D2, D6, Q6–Q8): the
	// last page has no cursor; lastBefore only when the range is empty; a
	// type for both sexes; an unpriced row has no price, no pricer and no
	// time, a migrated price no time; a growth cycle has no eggs and no lei
	// per egg, a cycle without birds no lei per bird; every delivery of a
	// cycle priced.
	"deliveries.nextCursor":        true,
	"deliveries.lastBefore":        true,
	"deliveries.rows[].forSex":     true,
	"deliveries.rows[].pricePerKg": true,
	"deliveries.rows[].totalLei":   true,
	"deliveries.rows[].pricedBy":   true,
	"deliveries.rows[].pricedAt":   true,
	"cost.cycles[].eggs":           true,
	"cost.cycles[].leiPerBird":     true,
	"cost.cycles[].leiPerEgg":      true,
	"cost.cycles[].unpriced":       true,
	"cost.total.leiPerBird":        true,
	"cost.total.leiPerEgg":         true,

	// GET /api/deliveries?priced=false (§7.13, M4 Q10): a type for both sexes.
	"items[].forSex": true,

	// GET /api/series/{id}/page (§7.3, M4 Q2–Q5): a closed series has no
	// week today; no placement event (Seria 2025); a farm panel without a
	// cycle of the series, without a guide, its cycle still open. The
	// panels' stats are the §7.6 stats: init below adds their paths.
	"saptToday":             true,
	"placement.placed":      true,
	"farms[].cycle":         true,
	"farms[].cycle.endDate": true,
}

// init allows under the series page's farms[].stats[] every null
// the summary allows under stats[] (M4 Q3: the panels carry that cycle's
// §7.6 stats unchanged).
func init() {
	for path := range maps.Clone(nullableFields) {
		if strings.HasPrefix(path, "stats[].") {
			nullableFields["farms[]."+path] = true
		}
	}
}

// nullable reports whether nullableFields allows a null at path (array
// indexes already written "[]"): the path itself, or a ".*" entry for one of
// its parent objects.
func nullable(path string) bool {
	if nullableFields[path] {
		return true
	}
	for i := range len(path) {
		if path[i] == '.' && nullableFields[path[:i]+".*"] {
			return true
		}
	}
	return false
}

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
			if !nullable(arrayIndex.ReplaceAllString(path, "[]")) {
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

// priceKeys returns the path of every price in the JSON document body: a key
// pricePerKg, totalLei, cost, price… or lei…, and any string holding "fără
// preț". A viewer's (and a vet's) responses must have none (SPEC §7.0a "never
// sent", D64, D200, M4 D2).
func priceKeys(body string) ([]string, error) {
	var v any
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return nil, err
	}
	var out []string
	var walk func(v any, path string)
	walk = func(v any, path string) {
		switch v := v.(type) {
		case string:
			if strings.Contains(v, "fără preț") {
				out = append(out, path)
			}
		case map[string]any:
			for k, e := range v {
				p := k
				if path != "" {
					p = path + "." + k
				}
				if k == "cost" || k == "totalLei" || strings.HasPrefix(k, "price") || strings.HasPrefix(k, "lei") {
					out = append(out, p)
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

// noPrices fails if r's JSON body carries a price (priceKeys); a body of
// another type (/healthz's "ok") has none.
func noPrices(t testing.TB, what string, r response) {
	t.Helper()
	if !strings.HasPrefix(r.header.Get("Content-Type"), "application/json") {
		return
	}
	keys, err := priceKeys(r.body)
	if err != nil {
		t.Errorf("%s: JSON body %q: %v", what, r.body, err)
	} else if len(keys) > 0 {
		t.Errorf("%s: a price for a caller who may not see prices at %v: %s", what, keys, r.body)
	}
}

// exec runs fixture SQL as il_app.
func (s *stack) exec(t testing.TB, sql string, args ...any) {
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
func (s *stack) session(t testing.TB, userID uuid.UUID, state sessionState) string {
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
	bodyNotFound     = `{"error":{"code":"NOT_FOUND","message":"Adresa cerută nu există."}}`
	bodyNoMethod     = `{"error":{"code":"METHOD_NOT_ALLOWED","message":"Operația nu este permisă la această adresă."}}`
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

// /api paths no route serves answer with the envelope (user answer Q5); the
// rest of the server keeps the mux's plain answers.
func TestAPINotFoundAndMethodNotAllowed(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	envelope := func(what string, r response) {
		t.Helper()
		if ct := r.header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("%s: Content-Type %q", what, ct)
		}
		if cc := r.header.Get("Cache-Control"); cc != "no-store" {
			t.Errorf("%s: Cache-Control %q", what, cc)
		}
	}

	r := s.do(t, call{method: http.MethodGet, path: "/api/nope"})
	wantBody(t, "GET /api/nope", r, http.StatusNotFound, bodyNotFound)
	envelope("GET /api/nope", r)
	r = s.do(t, call{method: http.MethodPost, path: "/api/nope", body: "{}"})
	wantBody(t, "POST /api/nope", r, http.StatusNotFound, bodyNotFound)

	r = s.do(t, call{method: http.MethodGet, path: "/api/auth/login"})
	wantBody(t, "GET /api/auth/login", r, http.StatusMethodNotAllowed, bodyNoMethod)
	envelope("GET /api/auth/login", r)
	if allow := r.header.Values("Allow"); !slices.Equal(allow, []string{"POST"}) {
		t.Errorf("GET /api/auth/login: Allow %q, want POST", allow)
	}
	r = s.do(t, call{method: http.MethodPost, path: "/api/auth/me", body: "{}"})
	wantBody(t, "POST /api/auth/me", r, http.StatusMethodNotAllowed, bodyNoMethod)
	if allow := r.header.Get("Allow"); allow != "GET, HEAD" {
		t.Errorf("POST /api/auth/me: Allow %q, want GET, HEAD", allow)
	}

	// The global Origin rule still runs first: a cross-site write is a 403
	// whether or not the path exists.
	r = s.do(t, call{method: http.MethodPost, path: "/api/nope", body: "{}", header: map[string]string{"Origin": "https://evil.example"}})
	wantBody(t, "cross-site POST /api/nope", r, http.StatusForbidden, bodyForbidden)

	// Outside /api the mux's defaults stay.
	for _, c := range []call{{method: http.MethodGet, path: "/nope"}, {method: http.MethodPost, path: "/healthz", body: "{}"}} {
		r := s.do(t, c)
		if r.status != http.StatusNotFound && r.status != http.StatusMethodNotAllowed || strings.HasPrefix(r.header.Get("Content-Type"), "application/json") {
			t.Errorf("%s %s: %s (Content-Type %q), want the mux's plain-text answer", c.method, c.path, r, r.header.Get("Content-Type"))
		}
	}
}

// The Content-Type rule runs per route (plan D9): the authed chain refuses a
// text/plain logout before the session check, and the session stays open.
func TestContentTypeOnAuthedRoutes(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	id := s.user(t, "ceban", "Dr. Ceban", dbq.UserRoleVet, true)
	token := s.session(t, id, sessionOpen)

	r := s.do(t, call{method: http.MethodPost, path: "/api/auth/logout", cookie: token, header: map[string]string{"Content-Type": "text/plain"}})
	wantBody(t, "text/plain logout", r, http.StatusForbidden, bodyForbidden)
	r = s.do(t, call{method: http.MethodPost, path: "/api/auth/logout", header: map[string]string{"Content-Type": "text/plain"}})
	wantBody(t, "text/plain logout without a session", r, http.StatusForbidden, bodyForbidden)
	if r := s.do(t, call{method: http.MethodGet, path: "/api/auth/me", cookie: token}); r.status != http.StatusOK {
		t.Errorf("me after the refused logout: %s, want 200", r)
	}
}

// otherMediaRoutes are the /api write routes whose chain allows a media type
// other than JSON (the guide preview's multipart, voice STT's audio when they
// land): pattern → the type the route takes. Empty today.
var otherMediaRoutes = map[string]string{}

// wildcard is a {name}, {name...} or {$} segment of a ServeMux pattern.
var wildcard = regexp.MustCompile(`\{[^}]*\}`)

// writeRoute splits a registered pattern into the write method and the path
// to request, or ok false for a read route or one outside /api. A pattern
// without a method matches every method, so it is a write route too (POST).
// Wildcards become a UUID; {$} becomes nothing.
func writeRoute(pattern string) (method, path string, ok bool) {
	method, path, found := strings.Cut(pattern, " ")
	if !found {
		method, path = http.MethodPost, pattern
	}
	if i := strings.Index(path, "/"); i > 0 { // a host before the path
		path = path[i:]
	}
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return "", "", false
	}
	if path != "/api" && !strings.HasPrefix(path, "/api/") {
		return "", "", false
	}
	path = wildcard.ReplaceAllStringFunc(path, func(w string) string {
		if w == "{$}" {
			return ""
		}
		return "0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90"
	})
	return method, path, true
}

// Every write route under /api carries the Content-Type rule (SPEC §11.4,
// plan D9): a text/plain body with a valid Origin and a valid session is
// refused with 403 FORBIDDEN before anything else runs. The walk covers
// every pattern the real wiring registers, so a write route registered
// without httpx.ContentType fails here.
func TestEveryWriteRouteChecksContentType(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	id := s.user(t, "proprietar", "Proprietar", dbq.UserRoleOwner, true)
	token := s.session(t, id, sessionOpen)

	var checked []string
	for _, pattern := range s.app.routes {
		method, path, ok := writeRoute(pattern)
		if !ok {
			continue
		}
		checked = append(checked, pattern)
		r := s.do(t, call{method: method, path: path, body: "{}", cookie: token,
			header: map[string]string{"Origin": testOrigin, "Content-Type": "text/plain"}})
		wantBody(t, pattern+" with text/plain", r, http.StatusForbidden, bodyForbidden)
		if media, ok := otherMediaRoutes[pattern]; ok {
			r := s.do(t, call{method: method, path: path, cookie: token,
				header: map[string]string{"Origin": testOrigin, "Content-Type": media}})
			if r.status == http.StatusForbidden && r.text() == bodyForbidden {
				t.Errorf("%s with its own media type %s: %s, want it past the Content-Type rule", pattern, media, r)
			}
		}
	}
	for _, want := range []string{
		"POST /api/auth/login", "POST /api/auth/logout", "POST /api/entries/feed-daily", "POST /api/entries/delivery",
		"POST /api/deliveries/prices", "POST /api/deliveries/prices/undo", "PATCH /api/deliveries/{id}/price",
	} {
		if !slices.Contains(checked, want) {
			t.Errorf("checked %q; the walk missed %s", checked, want)
		}
	}
	for pattern := range otherMediaRoutes {
		if !slices.Contains(checked, pattern) {
			t.Errorf("otherMediaRoutes names %q, which is not a registered /api write route", pattern)
		}
	}
}

func TestWriteRoute(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		pattern, method, path string
		ok                    bool
	}{
		{"POST /api/auth/login", "POST", "/api/auth/login", true},
		{"DELETE /api/entries/{id}", "DELETE", "/api/entries/0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90", true},
		{"PATCH /api/farms/{farm}/sections/{rest...}", "PATCH", "/api/farms/0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90/sections/0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90", true},
		{"PUT example.com/api/x/{$}", "PUT", "/api/x/", true},
		{"/api/any", "POST", "/api/any", true},
		{"GET /api/auth/me", "", "", false},
		{"HEAD /api/auth/me", "", "", false},
		{"POST /healthz", "", "", false},
		{"POST /apix", "", "", false},
	} {
		method, path, ok := writeRoute(tc.pattern)
		if method != tc.method || path != tc.path || ok != tc.ok {
			t.Errorf("writeRoute(%q) = %q %q %v, want %q %q %v", tc.pattern, method, path, ok, tc.method, tc.path, tc.ok)
		}
	}
}

// ---------------------------------------------------------------------------
// M2 farm reads: table, farm page, summary, bootstrap (§7.2–§7.4, §7.6)
// ---------------------------------------------------------------------------

// bodyUnknownParam is the 400 for an unknown query parameter (§7.0a, M2 D2).
func bodyUnknownParam(name string) string {
	return `{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",` +
		`"issues":[{"path":"` + name + `","code":"unknown_field","message":"Datele trimise nu sunt valide."}]}}`
}

// An unknown query parameter is refused with an issue naming it, on every
// M2 read; the farm page takes no period, the bootstrap no parameter at all.
func TestFarmReadsRejectUnknownParams(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	farm := s.farm(t, "Hîrbovăț", "hirbovat").String()
	token := s.session(t, s.user(t, "proprietar", "Proprietar", dbq.UserRoleOwner, true), sessionOpen)

	for _, tc := range []struct{ path, param string }{
		{"/api/table?view=farm&farmId=" + farm + "&foo=1", "foo"},
		{"/api/table?farmId=" + farm + "&view=farm&farmid=" + farm, "farmid"},
		{"/api/farms/" + farm + "/page?period=cycle", "period"},
		{"/api/farms/" + farm + "/summary?zi=2026-11-18", "zi"},
		{"/api/bootstrap?farmId=" + farm, "farmId"},
	} {
		r := s.do(t, call{method: http.MethodGet, path: tc.path, cookie: token})
		wantBody(t, tc.path, r, http.StatusBadRequest, bodyUnknownParam(tc.param))
	}
	// The same requests without the stray parameter pass.
	for _, path := range []string{
		"/api/table?view=farm&farmId=" + farm, "/api/farms/" + farm + "/page",
		"/api/farms/" + farm + "/summary", "/api/bootstrap",
	} {
		if r := s.do(t, call{method: http.MethodGet, path: path, cookie: token}); r.status != http.StatusOK {
			t.Errorf("%s: %s, want 200", path, r)
		}
	}
}

// seededReader is a seeded user's open session.
type seededReader struct {
	login string
	token string
}

// seededFarm returns the id of the seeded farm with that name_key.
func seededFarm(t testing.TB, s *stack, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := s.d.App.QueryRow(context.Background(), `select id from farm where name_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("farm %s: %v", key, err)
	}
	return id
}

// seededSession opens a session for the seeded user with that login.
func seededSession(t testing.TB, s *stack, login string) seededReader {
	t.Helper()
	var id uuid.UUID
	if err := s.d.App.QueryRow(context.Background(), `select id from app_user where login = $1`, login).Scan(&id); err != nil {
		t.Fatalf("user %s: %v", login, err)
	}
	return seededReader{login: login, token: s.session(t, id, sessionOpen)}
}

// On the real data (the imported history and the dev seed), every M2 read
// and the M3 reads (Consum by day and by week, the Hrană zilnică form, the
// Stoc report) answer 200 for the owner, the manager and the viewer, on both
// farms and for each of their cycles (laying with a guide, growth, the
// Seria 2025 cycle without a flock count), and every body passes stack.do's
// contract: no null list, a null only where nullableFields allows it,
// Cache-Control: no-store. The viewer is refused the form (FORBIDDEN). The
// vet gets the bootstrap and his farm's form, and is refused the pages,
// Consum, the Stoc report and another farm's form (FARM_FORBIDDEN).
//
// M4: the same callers read both series pages, every page of Livrări (82
// deliveries, cursor to cursor), Telița's empty interval (lastBefore), the
// Livrare hrană form (manager, owner; the viewer is refused) and the S12
// list (the owner only); the vet is refused all of them. No response to the
// viewer or the vet carries a price (noPrices, M4 D2).
func TestFarmReadsOnSeededData(t *testing.T) {
	t.Parallel()
	s := newSeededStack(t)
	hirb, tel := seededFarm(t, s, "hirbovat"), seededFarm(t, s, "telita")
	series := seededSeries(t, s)
	get := func(t *testing.T, u seededReader, path string, want int) response {
		t.Helper()
		r := s.do(t, call{method: http.MethodGet, path: path, cookie: u.token})
		if r.status != want {
			t.Fatalf("%s as %s: %s, want %d", path, u.login, r, want)
		}
		if u.login == "observator" || u.login == "dr.ceban" {
			noPrices(t, path+" as "+u.login, r)
		}
		return r
	}
	refused := func(t *testing.T, u seededReader, path, code string) {
		t.Helper()
		if got := errorCode(t, get(t, u, path, http.StatusForbidden)); got != code {
			t.Errorf("%s as %s: code %s, want %s", path, u.login, got, code)
		}
	}

	for _, login := range []string{"proprietar", "ion.rusu", "observator"} {
		u := seededSession(t, s, login)
		t.Run(login, func(t *testing.T) {
			get(t, u, "/api/bootstrap", http.StatusOK)
			for _, farm := range []uuid.UUID{hirb, tel} {
				base := "/api/farms/" + farm.String()
				var page struct {
					Cycles []struct {
						ID uuid.UUID `json:"id"`
					} `json:"cycles"`
				}
				if err := json.Unmarshal([]byte(get(t, u, base+"/page", http.StatusOK).body), &page); err != nil {
					t.Fatal(err)
				}
				if len(page.Cycles) == 0 {
					t.Fatalf("%s/page as %s: no cycles on the seeded farm", base, login)
				}
				for _, c := range page.Cycles {
					q := "cycleId=" + c.ID.String()
					get(t, u, base+"/page?"+q, http.StatusOK)
					get(t, u, base+"/summary?"+q, http.StatusOK)
					get(t, u, "/api/table?view=farm&farmId="+farm.String()+"&"+q, http.StatusOK)
					get(t, u, "/api/table?view=farm&farmId="+farm.String()+"&grain=week&"+q, http.StatusOK)
					// M3: Consum per cycle, by day and by week; the Hrană
					// zilnică form on a day of the cycle (a closed series' day
					// comes back locked).
					get(t, u, "/api/table?view=report-feed&farmId="+farm.String()+"&"+q, http.StatusOK)
					get(t, u, "/api/table?view=report-feed&farmId="+farm.String()+"&grain=week&"+q, http.StatusOK)
					form := "/api/forms/feed-daily?farmId=" + farm.String() + "&date=" + cycleDay(t, s, c.ID).String()
					if login == "observator" {
						if code := errorCode(t, get(t, u, form, http.StatusForbidden)); code != httpx.CodeForbidden {
							t.Errorf("%s as the viewer: code %s, want %s", form, code, httpx.CodeForbidden)
						}
					} else {
						get(t, u, form, http.StatusOK)
					}
				}
				get(t, u, "/api/reports/feed?farmId="+farm.String(), http.StatusOK)
				if login != "observator" {
					get(t, u, "/api/forms/feed-daily?farmId="+farm.String()+"&date=2026-11-18", http.StatusOK)
				}
			}
			get(t, u, "/api/reports/feed", http.StatusOK)
			get(t, u, "/api/reports/feed?farmId=all&period=weeks:4", http.StatusOK)
			// Hîrbovăț's open cycle: load-more, the start note, a short and an
			// interval period, one week.
			table := "/api/table?view=farm&farmId=" + hirb.String()
			for _, q := range []string{
				"&before=56", "&before=24", "&period=weeks:4", "&period=interval:2026-11-01,2026-11-15", "&sapt=58",
				"&types=oua,hrana",
			} {
				get(t, u, table+q, http.StatusOK)
			}
			get(t, u, "/api/farms/"+hirb.String()+"/summary?period=weeks:4", http.StatusOK)

			// M4: series pages, Livrări pages, the empty interval, Livrare
			// hrană form, S12 list.
			for _, id := range series {
				get(t, u, "/api/series/"+id.String()+"/page", http.StatusOK)
			}
			if n := deliveryPages(t, func(path string) response { return get(t, u, path, http.StatusOK) }, "/api/reports/feed?"); n != 82 {
				t.Errorf("Livrări as %s: %d rows over every page, want 82", login, n)
			}
			empty := decodeBody[struct {
				Deliveries struct {
					Total      int `json:"total"`
					LastBefore *struct {
						Date string `json:"date"`
						Type string `json:"type"`
					} `json:"lastBefore"`
				} `json:"deliveries"`
			}](t, get(t, u, "/api/reports/feed?farmId="+tel.String()+"&period=interval:2026-03-01,2026-11-18", http.StatusOK))
			if d := empty.Deliveries; d.Total != 0 || d.LastBefore == nil || d.LastBefore.Date != "2026-02-10" || d.LastBefore.Type != "Pre-ouat" {
				t.Errorf("Telița 01.03–18.11 as %s: %+v, want none and the last one on 10.02 (Pre-ouat)", login, d)
			}
			for _, farm := range []uuid.UUID{hirb, tel} {
				form := "/api/forms/delivery?farmId=" + farm.String()
				if login == "observator" {
					refused(t, u, form, httpx.CodeForbidden)
				} else {
					get(t, u, form, http.StatusOK)
				}
			}
			if login == "proprietar" {
				get(t, u, "/api/deliveries?priced=false", http.StatusOK)
			} else {
				refused(t, u, "/api/deliveries?priced=false", httpx.CodeForbidden)
			}
		})
	}

	vet := seededSession(t, s, "dr.ceban")
	get(t, vet, "/api/bootstrap", http.StatusOK)
	for _, path := range []string{
		"/api/farms/" + hirb.String() + "/page", "/api/farms/" + hirb.String() + "/summary",
		"/api/table?view=farm&farmId=" + hirb.String(),
		"/api/table?view=report-feed&farmId=" + hirb.String(), "/api/reports/feed",
	} {
		if code := errorCode(t, get(t, vet, path, http.StatusForbidden)); code != httpx.CodeForbidden {
			t.Errorf("%s as the vet: code %s, want %s", path, code, httpx.CodeForbidden)
		}
	}
	// The vet's Hrană zilnică: today on his farm, a past day (locked), and
	// Telița, not his.
	get(t, vet, "/api/forms/feed-daily?farmId="+hirb.String()+"&date=2026-11-18", http.StatusOK)
	get(t, vet, "/api/forms/feed-daily?farmId="+hirb.String()+"&date=2026-11-10", http.StatusOK)
	telForm := "/api/forms/feed-daily?farmId=" + tel.String() + "&date=2026-11-18"
	if code := errorCode(t, get(t, vet, telForm, http.StatusForbidden)); code != httpx.CodeFarmForbidden {
		t.Errorf("%s as the vet: code %s, want %s", telForm, code, httpx.CodeFarmForbidden)
	}
	// M4: no series page, Livrare hrană or Prețuri for the vet.
	for _, path := range []string{
		"/api/series/" + series[0].String() + "/page", "/api/forms/delivery?farmId=" + hirb.String(), "/api/deliveries?priced=false",
	} {
		refused(t, vet, path, httpx.CodeForbidden)
	}
}

// seededSeries are the ids of the seeded series, oldest first (Seria 2025,
// Seria 2026).
func seededSeries(t testing.TB, s *stack) []uuid.UUID {
	t.Helper()
	rows, err := s.d.App.Query(context.Background(), `select id from series order by year`)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil || len(ids) != 2 {
		t.Fatalf("seeded series: %v, %v; want 2", ids, err)
	}
	return ids
}

// deliveryPages reads Livrări from base ("/api/reports/feed?…&") page by
// page through nextCursor and returns how many rows it saw; it fails on a
// row seen twice.
func deliveryPages(t testing.TB, get func(path string) response, base string) int {
	t.Helper()
	seen := map[uuid.UUID]bool{}
	path := base
	for range 100 {
		page := decodeBody[struct {
			Deliveries struct {
				Rows []struct {
					ID uuid.UUID `json:"id"`
				} `json:"rows"`
				NextCursor *string `json:"nextCursor"`
			} `json:"deliveries"`
		}](t, get(path))
		for _, r := range page.Deliveries.Rows {
			if seen[r.ID] {
				t.Errorf("%s: delivery %s again", path, r.ID)
			}
			seen[r.ID] = true
		}
		if page.Deliveries.NextCursor == nil {
			return len(seen)
		}
		path = base + "cursor=" + url.QueryEscape(*page.Deliveries.NextCursor)
	}
	t.Fatalf("%s: more than 100 pages", base)
	return 0
}

// cycleDay is a day inside the seeded cycle: its start + 30 days, or its last
// day (its end, or today) when the cycle is shorter.
func cycleDay(t *testing.T, s *stack, cycleID uuid.UUID) timex.Day {
	t.Helper()
	var day timex.Day
	if err := s.d.App.QueryRow(context.Background(), `select least(start_date + 30, coalesce(end_date, $2::date))
		from cycle where id = $1`, cycleID, testdb.Today).Scan(&day); err != nil {
		t.Fatalf("cycle %s: %v", cycleID, err)
	}
	return day
}

// BenchmarkFarmReads times the M2 reads and the M3 and M4 routes through the
// whole server (socket, middleware, session check, service, SQL, JSON) on the
// seeded database, the owner on Hîrbovăț's open cycle (SAPT 20–59) and
// Telița's growth cycle. It reports the p50 and p95 per request against SPEC
// §11.6 (table page ≤ 300 ms, load-more ≤ 200 ms, summary ≤ 400 ms, save ≤
// 300 ms; the series page and Rapoarte › Hrană are page reads, held to the
// summary's 400 ms). The feed save is a real save of Hîrbovăț's 18.11.2026
// rows, its dry run the same save rolled back: before each one (outside the
// timed request) the day's live entry is soft-deleted. The delivery save adds
// a new 9 000 kg delivery each time (two of one day are legal); the price
// save prices the seeded 12.11 delivery, whose price is cleared before each
// one (outside the timed request). The writes run last, so the reads see the
// seed as it is.
// Run it without -race:
//
//	go test -run '^$' -bench FarmReads -benchtime 200x ./cmd/api
func BenchmarkFarmReads(b *testing.B) {
	s := newSeededStack(b)
	hirbID := seededFarm(b, s, "hirbovat")
	hirb, tel := hirbID.String(), seededFarm(b, s, "telita").String()
	owner := seededSession(b, s, "proprietar")
	feedSave := seededFeedDay(b, s, hirbID)
	seria2026 := seededSeries(b, s)[1].String()
	ctx := context.Background()
	var cursor, breeder2 string
	var unpriced uuid.UUID
	if err := s.d.App.QueryRow(ctx, `select fd.id, ft.id::text from feed_delivery fd join feed_type ft on ft.id = fd.feed_type_id
		where fd.farm_id = $1 and fd.date = '2026-11-12' and ft.name = 'Breeder 2' and fd.total_lei is null and fd.deleted_at is null`,
		hirbID).Scan(&unpriced, &breeder2); err != nil {
		b.Fatal(err)
	}
	first := decodeBody[struct {
		Deliveries struct {
			NextCursor *string `json:"nextCursor"`
		} `json:"deliveries"`
	}](b, s.do(b, call{method: http.MethodGet, path: "/api/reports/feed", cookie: owner.token}))
	if first.Deliveries.NextCursor == nil {
		b.Fatal("reports/feed: no second page")
	}
	cursor = url.QueryEscape(*first.Deliveries.NextCursor)
	unprice := func() {
		s.exec(b, `update feed_delivery set price_per_kg = null, total_lei = null, price_input = null, priced_by = null,
			priced_at = null where id = $1`, unpriced)
	}

	for _, bc := range []struct {
		name, method, path, body string
		before                   func() // untimed, before each request
		status                   int
	}{
		{name: "table-page", path: "/api/table?view=farm&farmId=" + hirb},
		{name: "table-load-more", path: "/api/table?view=farm&farmId=" + hirb + "&before=56"},
		{name: "table-whole-cycle-weeks", path: "/api/table?view=farm&farmId=" + hirb + "&grain=week&count=40"},
		{name: "farm-page", path: "/api/farms/" + hirb + "/page"},
		{name: "summary", path: "/api/farms/" + hirb + "/summary"},
		{name: "summary-growth", path: "/api/farms/" + tel + "/summary"},
		{name: "bootstrap", path: "/api/bootstrap"},
		// M3.
		{name: "feed-form", path: "/api/forms/feed-daily?farmId=" + hirb + "&date=2026-11-18"},
		{name: "feed-save", method: http.MethodPost, path: feedDailyPath, body: feedSave,
			before: func() { deleteFeedDay(b, s, hirbID) }, status: http.StatusCreated},
		{name: "feed-save-dry-run", method: http.MethodPost, path: feedDailyPath,
			body:   strings.Replace(feedSave, "{", `{"dryRun":true,`, 1),
			before: func() { deleteFeedDay(b, s, hirbID) }}, // the whole save path, rolled back
		{name: "reports-feed", path: "/api/reports/feed"},
		{name: "consum-page", path: "/api/table?view=report-feed&farmId=" + hirb},
		{name: "consum-load-more", path: "/api/table?view=report-feed&farmId=" + hirb + "&before=56"},
		// M4.
		{name: "series-page", path: "/api/series/" + seria2026 + "/page"},
		{name: "reports-feed-cursor", path: "/api/reports/feed?cursor=" + cursor},
		{name: "delivery-form", path: "/api/forms/delivery?farmId=" + hirb},
		{name: "unpriced-list", path: "/api/deliveries?priced=false"},
		{name: "delivery-save", method: http.MethodPost, path: deliveryPath, status: http.StatusCreated,
			body: `{"farmId":"` + hirb + `","date":"2026-11-18","feedTypeId":"` + breeder2 + `","qtyKg":9000}`},
		{name: "price-save", method: http.MethodPost, path: "/api/deliveries/prices", before: unprice,
			body: `{"items":[{"id":"` + unpriced.String() + `","pricePerKg":7.74}]}`},
	} {
		b.Run(bc.name, func(b *testing.B) {
			method, status := cmp.Or(bc.method, http.MethodGet), cmp.Or(bc.status, http.StatusOK)
			client := s.srv.Client()
			var took []time.Duration
			for b.Loop() {
				if bc.before != nil {
					bc.before()
				}
				var body io.Reader
				if bc.body != "" {
					body = strings.NewReader(bc.body)
				}
				req, err := http.NewRequest(method, s.srv.URL+bc.path, body)
				if err != nil {
					b.Fatal(err)
				}
				req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: owner.token})
				if method == http.MethodPost {
					req.Header.Set("Origin", testOrigin)
					req.Header.Set("Content-Type", "application/json")
				}
				start := time.Now()
				res, err := client.Do(req)
				if err != nil {
					b.Fatal(err)
				}
				_, err = io.Copy(io.Discard, res.Body)
				_ = res.Body.Close()
				took = append(took, time.Since(start))
				if err != nil || res.StatusCode != status {
					b.Fatalf("%s %s: %d %v, want %d", method, bc.path, res.StatusCode, err, status)
				}
			}
			slices.Sort(took)
			ms := func(q float64) float64 {
				return float64(took[int(q*float64(len(took)-1))]) / float64(time.Millisecond)
			}
			b.ReportMetric(ms(0.50), "p50-ms")
			b.ReportMetric(ms(0.95), "p95-ms")
		})
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
		// Nullable scalars pass; the same name elsewhere does not.
		{`{"cycle":null,"alive":null,"more":null}`, nil},
		{`{"farm":{"guide":null},"cycle":{"guide":null}}`, []string{"cycle.guide"}},
		// A ".*" entry allows any key under that object, dotted keys included,
		// but not the object itself nor a deeper list.
		{`{"weeks":[{"cells":{"alive.F":null,"egg.layPct":null}}],"totals":{"cells":{"alive.F":null}}}`, nil},
		{`{"weeks":[{"cells":null,"days":null}]}`, []string{"weeks[0].cells", "weeks[0].days"}},
		{`{"chips":null,"groups":[{"leaves":null}]}`, []string{"chips", "groups[0].leaves"}},
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

func TestPriceKeys(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want []string
	}{
		{`{"deliveries":{"rows":[{"id":"x","kg":9000,"forSex":null}],"unpriced":{"count":4}}}`, nil},
		{`{"deliveries":{"rows":[{"kg":9000,"pricePerKg":null,"totalLei":null,"pricedBy":null,"pricedAt":null}]}}`,
			[]string{"deliveries.rows[0].pricePerKg", "deliveries.rows[0].pricedAt", "deliveries.rows[0].pricedBy", "deliveries.rows[0].totalLei"}},
		{`{"cost":{"cycles":[]},"stock":[]}`, []string{"cost"}},
		{`{"cycles":[{"lei":1,"leiPerBird":2,"kg":3}]}`, []string{"cycles[0].lei", "cycles[0].leiPerBird"}},
		{`{"rows":[{"tag":"fără preț"}],"toast":"Livrare adăugată · 9 000 kg · fără preț"}`, []string{"rows[0].tag", "toast"}},
		// Words that only look alike pass: a "pret" chip, a "layPct" key.
		{`{"recordTypes":["pret"],"layPct":81.2,"label":"Prețuri"}`, nil},
	} {
		got, err := priceKeys(tc.body)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, tc.want) {
			t.Errorf("priceKeys(%s) = %q, want %q", tc.body, got, tc.want)
		}
	}
	if _, err := priceKeys(`{"a":`); err == nil {
		t.Error("priceKeys accepted malformed JSON")
	}
}
