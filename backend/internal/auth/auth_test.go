package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

const goodPassword = "parola-corecta"

// The hash of goodPassword, computed once: argon2 under -race is slow.
var goodHash = sync.OnceValue(func() string {
	h, err := NewHasher().Hash(context.Background(), goodPassword)
	if err != nil {
		panic(err)
	}
	return h
})

type env struct {
	t   *testing.T
	d   *testdb.DB
	svc *Service
	h   http.Handler
}

func newEnv(t *testing.T) *env {
	t.Helper()
	d := testdb.New(t)
	svc := NewService(db.NewStore(d.App), d.Clock)
	mux := http.NewServeMux()
	NewHandler(svc).Routes(mux, httpx.NewChain(), httpx.NewChain(svc.RequireSession))
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	return &env{t: t, d: d, svc: svc, h: httpx.NewChain(httpx.ClientIP(trusted)).Then(mux)}
}

func (e *env) exec(sql string, args ...any) {
	e.t.Helper()
	if _, err := e.d.App.Exec(context.Background(), sql, args...); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) farm(name, key string) uuid.UUID {
	e.t.Helper()
	var id uuid.UUID
	if err := e.d.App.QueryRow(context.Background(),
		`insert into farm (name, name_key, type) values ($1, $2, 'laying') returning id`, name, key).Scan(&id); err != nil {
		e.t.Fatal(err)
	}
	return id
}

func (e *env) user(login, name string, role dbq.UserRole, farms ...uuid.UUID) uuid.UUID {
	e.t.Helper()
	return e.userWithHash(login, name, role, goodHash(), farms...)
}

func (e *env) userWithHash(login, name string, role dbq.UserRole, hash string, farms ...uuid.UUID) uuid.UUID {
	e.t.Helper()
	var id uuid.UUID
	if err := e.d.App.QueryRow(context.Background(),
		`insert into app_user (name, login, password_hash, role) values ($1, $2, $3, $4) returning id`,
		name, login, hash, role).Scan(&id); err != nil {
		e.t.Fatal(err)
	}
	for _, f := range farms {
		e.exec(`insert into user_farm (user_id, farm_id) values ($1, $2)`, id, f)
	}
	return id
}

type req struct {
	method, path, body, cookie, remote string
}

func (e *env) do(r req) *httptest.ResponseRecorder {
	e.t.Helper()
	hr := httptest.NewRequest(r.method, r.path, strings.NewReader(r.body))
	hr.Header.Set("Content-Type", "application/json")
	if r.cookie != "" {
		hr.AddCookie(&http.Cookie{Name: CookieName, Value: r.cookie})
	}
	if r.remote != "" {
		hr.RemoteAddr = r.remote
	}
	w := httptest.NewRecorder()
	e.h.ServeHTTP(w, hr)
	return w
}

func (e *env) login(login, password string) *httptest.ResponseRecorder {
	e.t.Helper()
	b, _ := json.Marshal(map[string]string{"login": login, "password": password})
	return e.do(req{method: "POST", path: "/api/auth/login", body: string(b)})
}

// signIn logs in and returns the cookie token.
func (e *env) signIn(login string) string {
	e.t.Helper()
	w := e.login(login, goodPassword)
	if w.Code != 200 {
		e.t.Fatalf("login %s: %d %s", login, w.Code, w.Body)
	}
	return sessionToken(e.t, w)
}

var setCookiePattern = regexp.MustCompile(`^il_session=([A-Za-z0-9_-]{43}); Path=/; Max-Age=2592000; HttpOnly; Secure; SameSite=Lax$`)

func sessionToken(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	cs := w.Result().Header.Values("Set-Cookie")
	if len(cs) != 1 {
		t.Fatalf("Set-Cookie %q, want one", cs)
	}
	m := setCookiePattern.FindStringSubmatch(cs[0])
	if m == nil {
		t.Fatalf("Set-Cookie %q", cs[0])
	}
	return m[1]
}

type attempt struct {
	Login   string
	IP      string
	At      time.Time
	Success bool
}

func (e *env) attempts() []attempt {
	e.t.Helper()
	rows, err := e.d.App.Query(context.Background(), `select login, coalesce(host(ip), ''), at, success from login_attempt order by id`)
	if err != nil {
		e.t.Fatal(err)
	}
	defer rows.Close()
	var out []attempt
	for rows.Next() {
		var a attempt
		if err := rows.Scan(&a.Login, &a.IP, &a.At, &a.Success); err != nil {
			e.t.Fatal(err)
		}
		out = append(out, a)
	}
	return out
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, status int, body string) {
	t.Helper()
	if w.Code != status || strings.TrimSuffix(w.Body.String(), "\n") != body {
		t.Errorf("got %d %s\nwant %d %s", w.Code, w.Body, status, body)
	}
}

func TestLoginSuccess(t *testing.T) {
	e := newEnv(t)
	hirb := e.farm("Hîrbovăț", "hirbovat")
	e.farm("Telița", "telita")
	id := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet, hirb)

	b, _ := json.Marshal(map[string]string{"login": "  Ceban ", "password": goodPassword})
	w := e.do(req{method: "POST", path: "/api/auth/login", body: string(b), remote: "10.0.0.2:5000"})
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	want := fmt.Sprintf(`{"user":{"id":"%s","name":"Dr. Ceban","login":"ceban","role":"vet"},`+
		`"farms":[{"id":"%s","name":"Hîrbovăț","type":"laying"}],"today":"2026-11-18"}`+"\n", id, hirb)
	if w.Body.String() != want {
		t.Errorf("body %s\nwant %s", w.Body, want)
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("login response must be no-store")
	}
	token := sessionToken(t, w)

	// The DB stores sha256 of the 32 token bytes, never the token.
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		t.Fatalf("token %q is not 32 bytes of base64url", token)
	}
	sum := sha256.Sum256(raw)
	var expires, lastSeen time.Time
	var ua *string
	var ip string
	if err := e.d.App.QueryRow(context.Background(),
		`select expires_at, last_seen_at, user_agent, host(ip) from user_session where token_hash = $1 and user_id = $2`,
		sum[:], id).Scan(&expires, &lastSeen, &ua, &ip); err != nil {
		t.Fatalf("session row: %v", err)
	}
	now := testdb.Now()
	if !expires.Equal(now.Add(30*24*time.Hour)) || !lastSeen.Equal(now) || ip != "10.0.0.2" {
		t.Errorf("session expires %v last seen %v ip %s", expires, lastSeen, ip)
	}

	var last time.Time
	if err := e.d.App.QueryRow(context.Background(), `select last_login_at from app_user where id = $1`, id).Scan(&last); err != nil || !last.Equal(now) {
		t.Errorf("last_login_at %v (%v), want %v", last, err, now)
	}
	if got := e.attempts(); len(got) != 1 || got[0] != (attempt{"ceban", "10.0.0.2", got[0].At, true}) || !got[0].At.Equal(now) {
		t.Errorf("attempts %+v", got)
	}
}

func TestLoginRefusals(t *testing.T) {
	const (
		emptyLogin = `{"error":{"code":"VALIDATION","message":"Scrie loginul.","issues":[{"path":"login","code":"required","message":"Scrie loginul."}]}}`
		emptyPass  = `{"error":{"code":"VALIDATION","message":"Scrie parola.","issues":[{"path":"password","code":"required","message":"Scrie parola."}]}}`
		bothEmpty  = `{"error":{"code":"VALIDATION","message":"Scrie loginul.","issues":[{"path":"login","code":"required","message":"Scrie loginul."},{"path":"password","code":"required","message":"Scrie parola."}]}}`
		wrong      = `{"error":{"code":"UNAUTHENTICATED","message":"Login sau parolă greșită."}}`
		disabled   = `{"error":{"code":"ACCOUNT_DISABLED","message":"Contul tău este dezactivat. Vorbește cu administratorul."}}`
	)
	tests := []struct {
		name, login, password string
		status                int
		body                  string
		recorded              *bool // nil = no login_attempt row
		idKeys                int64
	}{
		{"empty login", "  ", goodPassword, 400, emptyLogin, nil, 0},
		{"empty password", "ceban", "", 400, emptyPass, nil, 0},
		{"both empty", "", "", 400, bothEmpty, nil, 0},
		{"wrong password", "ceban", "gresita", 401, wrong, new(false), 1},
		{"unknown login runs the dummy hash", "nimeni", goodPassword, 401, wrong, new(false), 1},
		{"login that can't exist: 401, no hint", "Ce ban!", goodPassword, 401, wrong, new(false), 0},
		{"deactivated, right password", "fost", goodPassword, 403, disabled, new(false), 1},
		{"deactivated, wrong password", "fost", "gresita", 401, wrong, new(false), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t)
			e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
			fost := e.user("fost", "Fost Angajat", dbq.UserRoleManager)
			e.exec(`update app_user set active = false where id = $1`, fost)

			before := e.svc.hash.runs.Load()
			w := e.login(tt.login, tt.password)
			assertError(t, w, tt.status, tt.body)
			if w.Header().Get("Set-Cookie") != "" {
				t.Error("a refused login set a cookie")
			}
			if n := e.svc.hash.runs.Load() - before; n != tt.idKeys {
				t.Errorf("IDKey runs %d, want %d", n, tt.idKeys)
			}
			got := e.attempts()
			if tt.recorded == nil {
				if len(got) != 0 {
					t.Errorf("attempts %+v, want none", got)
				}
				return
			}
			if len(got) != 1 || got[0].Success != *tt.recorded || got[0].Login != NormalizeLogin(tt.login) {
				t.Errorf("attempts %+v", got)
			}
			var last *time.Time
			_ = e.d.App.QueryRow(context.Background(), `select max(last_login_at) from app_user`).Scan(&last)
			if last != nil {
				t.Errorf("last_login_at set to %v by a refused login", last)
			}
		})
	}
}

const rateLimited = `{"error":{"code":"RATE_LIMITED","message":"Prea multe încercări. Încearcă din nou peste 5 minute.","retryAfterSeconds":%d}}`

func assertLimited(t *testing.T, w *httptest.ResponseRecorder, secs int) {
	t.Helper()
	assertError(t, w, 429, fmt.Sprintf(rateLimited, secs))
	if got := w.Header().Get("Retry-After"); got != fmt.Sprint(secs) {
		t.Errorf("Retry-After %q, want %d", got, secs)
	}
}

func TestLoginThrottlePerLogin(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	for i := range 5 {
		if w := e.login("ceban", "gresita"); w.Code != 401 {
			t.Fatalf("failure %d: %d", i+1, w.Code)
		}
		e.d.Clock.Advance(time.Minute) // failures at 12:00 … 12:04, now 12:05
	}
	e.d.Clock.Advance(-time.Minute) // now 12:04, latest failure 12:04

	before := e.svc.hash.runs.Load()
	assertLimited(t, e.login("ceban", goodPassword), 300) // even the right password waits
	if e.svc.hash.runs.Load() != before {
		t.Error("a throttled attempt checked the password")
	}
	if n := len(e.attempts()); n != 5 {
		t.Errorf("%d attempts recorded, want 5 (429s are not recorded, Q3a)", n)
	}

	e.d.Clock.Advance(2*time.Minute + 30*time.Second)
	assertLimited(t, e.login("ceban", goodPassword), 150)
	assertLimited(t, e.login("CEBAN", goodPassword), 150) // same login after normalising

	// Another login from the same IP is not affected by the login lock.
	e.user("rusu", "Ion Rusu", dbq.UserRoleOwner)
	if w := e.login("rusu", goodPassword); w.Code != 200 {
		t.Errorf("other login: %d %s", w.Code, w.Body)
	}

	e.d.Clock.Advance(2*time.Minute + 30*time.Second) // latest failure + 5 min
	if w := e.login("ceban", goodPassword); w.Code != 200 {
		t.Errorf("after 5 minutes: %d %s", w.Code, w.Body)
	}
}

func TestLoginThrottleWindowIs15Minutes(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	// Two old failures fall out of the 15-minute window; three recent ones are not enough.
	now := testdb.Now()
	for _, ago := range []time.Duration{16 * time.Minute, 15 * time.Minute, 3 * time.Minute, 2 * time.Minute, time.Minute} {
		e.exec(`insert into login_attempt (login, ip, at, success) values ('ceban', '192.0.2.1', $1, false)`, now.Add(-ago))
	}
	if w := e.login("ceban", goodPassword); w.Code != 200 {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}

func TestLoginThrottlePerIP(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	const client = "198.51.100.9"
	for i := range 20 {
		// Logins that can't exist fail without hashing; each one counts for the IP.
		w := e.do(req{method: "POST", path: "/api/auth/login", remote: "10.0.0.2:5000",
			body: fmt.Sprintf(`{"login":"x%d!","password":"p"}`, i)})
		if w.Code != 401 {
			t.Fatalf("failure %d: %d", i+1, w.Code)
		}
	}
	// Behind the trusted proxy the client is the forwarded address.
	for _, a := range e.attempts() {
		if a.IP != "10.0.0.2" {
			t.Fatalf("attempt ip %s", a.IP)
		}
	}

	limited := req{method: "POST", path: "/api/auth/login", remote: "10.0.0.2:5000",
		body: `{"login":"ceban","password":"` + goodPassword + `"}`}
	assertLimited(t, e.do(limited), 300)

	// From another address the same login works.
	limited.remote = client + ":5000"
	if w := e.do(limited); w.Code != 200 {
		t.Errorf("other IP: %d %s", w.Code, w.Body)
	}
}

func TestLoginIPIsForwardedClient(t *testing.T) {
	e := newEnv(t)
	hr := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"login":"nimeni","password":"x"}`))
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("X-Forwarded-For", "1.2.3.4, 198.51.100.9")
	hr.RemoteAddr = "10.0.0.2:5000"
	e.h.ServeHTTP(httptest.NewRecorder(), hr)

	hr = httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"login":"nimeni","password":"x"}`))
	hr.Header.Set("Content-Type", "application/json")
	hr.Header.Set("X-Forwarded-For", "1.2.3.4")
	hr.RemoteAddr = "203.0.113.7:5000" // untrusted peer: its XFF is ignored
	e.h.ServeHTTP(httptest.NewRecorder(), hr)

	got := e.attempts()
	if len(got) != 2 || got[0].IP != "198.51.100.9" || got[1].IP != "203.0.113.7" {
		t.Errorf("attempts %+v", got)
	}
}

func TestRehashOnLogin(t *testing.T) {
	e := newEnv(t)
	old := Params{Time: 1, Memory: 8 * 1024, Threads: 1, KeyLen: 32}
	salt := []byte("fedcba9876543210")
	oldHash := encodePHC(old, salt, argon2.IDKey([]byte(goodPassword), salt, old.Time, old.Memory, old.Threads, old.KeyLen))
	id := e.userWithHash("vechi", "Cont Vechi", dbq.UserRoleViewer, oldHash)

	if w := e.login("vechi", goodPassword); w.Code != 200 {
		t.Fatalf("%d %s", w.Code, w.Body)
	}
	var stored string
	if err := e.d.App.QueryRow(context.Background(), `select password_hash from app_user where id = $1`, id).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored == oldHash || !phcPattern.MatchString(stored) {
		t.Errorf("stored hash %q, want a new one with the current parameters", stored)
	}
	if ok, rehash, err := NewHasher().Verify(context.Background(), goodPassword, stored); !ok || rehash || err != nil {
		t.Errorf("new hash: ok=%v rehash=%v err=%v", ok, rehash, err)
	}
}

func TestMe(t *testing.T) {
	e := newEnv(t)
	hirb := e.farm("Hîrbovăț", "hirbovat")
	tel := e.farm("Telița", "telita")
	cioc := e.farm("Ciocana", "ciocana")
	vet := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet, hirb)
	owner := e.user("rusu", "Ion Rusu", dbq.UserRoleOwner)
	mgr := e.user("lungu", "Ana Lungu", dbq.UserRoleManager, tel, cioc)

	farm := func(id uuid.UUID, name string) string {
		return fmt.Sprintf(`{"id":"%s","name":"%s","type":"laying"}`, id, name)
	}
	tests := []struct {
		login, body string
	}{
		{"ceban", fmt.Sprintf(`{"user":{"id":"%s","name":"Dr. Ceban","login":"ceban","role":"vet"},"farms":[%s],`, vet, farm(hirb, "Hîrbovăț"))},
		{"rusu", fmt.Sprintf(`{"user":{"id":"%s","name":"Ion Rusu","login":"rusu","role":"owner"},"farms":[%s,%s,%s],`, owner,
			farm(cioc, "Ciocana"), farm(hirb, "Hîrbovăț"), farm(tel, "Telița"))},
		{"lungu", fmt.Sprintf(`{"user":{"id":"%s","name":"Ana Lungu","login":"lungu","role":"manager"},"farms":[%s,%s],`, mgr,
			farm(cioc, "Ciocana"), farm(tel, "Telița"))},
	}
	for _, tt := range tests {
		t.Run(tt.login, func(t *testing.T) {
			w := e.do(req{method: "GET", path: "/api/auth/me", cookie: e.signIn(tt.login)})
			want := tt.body + `"today":"2026-11-18","now":"2026-11-18T12:00:00+02:00","startPage":"/acasa"}` + "\n"
			if w.Code != 200 || w.Body.String() != want {
				t.Errorf("got %d %s\nwant %s", w.Code, w.Body, want)
			}
		})
	}

	t.Run("user without farms gets []", func(t *testing.T) {
		e.user("gol", "Fără Ferme", dbq.UserRoleViewer)
		w := e.do(req{method: "GET", path: "/api/auth/me", cookie: e.signIn("gol")})
		if !strings.Contains(w.Body.String(), `"farms":[]`) {
			t.Errorf("body %s", w.Body)
		}
	})
}

const unauthenticated = `{"error":{"code":"UNAUTHENTICATED","message":"Sesiunea a expirat. Intră din nou în cont."}}`

func TestSessionMiddlewarePrincipal(t *testing.T) {
	e := newEnv(t)
	hirb := e.farm("Hîrbovăț", "hirbovat")
	vet := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet, hirb)
	owner := e.user("rusu", "Ion Rusu", dbq.UserRoleOwner)
	viewer := e.user("gol", "Fără Ferme", dbq.UserRoleViewer)

	var got Principal
	probe := e.svc.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = PrincipalFrom(r.Context())
	}))
	call := func(token string) {
		r := httptest.NewRequest("GET", "/api/x", nil)
		r.AddCookie(&http.Cookie{Name: CookieName, Value: token})
		got = Principal{}
		probe.ServeHTTP(httptest.NewRecorder(), r)
	}

	call(e.signIn("ceban"))
	if got.UserID != vet || got.Name != "Dr. Ceban" || got.Role != Vet || !slices.Equal(got.FarmIDs, []uuid.UUID{hirb}) {
		t.Errorf("vet principal %+v", got)
	}
	call(e.signIn("rusu"))
	if got.UserID != owner || got.Role != Owner || got.FarmIDs != nil {
		t.Errorf("owner principal %+v (FarmIDs must be nil)", got)
	}
	call(e.signIn("gol"))
	if got.UserID != viewer || got.FarmIDs == nil || len(got.FarmIDs) != 0 {
		t.Errorf("viewer without farms %+v (FarmIDs must be empty, not nil)", got)
	}
}

// queryCounter is a pgx QueryTracer that records the sqlc queries (their
// "-- name: X" header) a pool runs; connection setup (SET ROLE, app.today)
// is not counted.
type queryCounter struct {
	mu    sync.Mutex
	names []string
}

func (c *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name, _, _ := strings.Cut(rest, " ")
		c.mu.Lock()
		c.names = append(c.names, name)
		c.mu.Unlock()
	}
	return ctx
}

func (c *queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (c *queryCounter) take() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := c.names
	c.names = nil
	return out
}

// TestRequireSessionQueries: the session middleware loads the session, the
// user and the user's farms in one query, and adds the slide only when it is
// due (follow-up ARC-2).
func TestRequireSessionQueries(t *testing.T) {
	e := newEnv(t)
	hirb := e.farm("Hîrbovăț", "hirbovat")
	tel := e.farm("Telița", "telita")
	cioc := e.farm("Ciocana", "ciocana")
	vet := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet, hirb)
	mgr := e.user("lungu", "Ana Lungu", dbq.UserRoleManager, tel, cioc)
	e.user("rusu", "Ion Rusu", dbq.UserRoleOwner)
	e.user("gol", "Fără Ferme", dbq.UserRoleViewer)

	// A second app pool on the same database, traced.
	var counter queryCounter
	pc, err := db.PoolConfig(testdb.DSN(e.d.Name), 2, &testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	prepare := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "set role "+testdb.RoleApp); err != nil {
			return err
		}
		return prepare(ctx, conn)
	}
	pc.ConnConfig.Tracer = &counter
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	svc := NewService(db.NewStore(pool), e.d.Clock)

	var got Principal
	probe := svc.RequireSession(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = PrincipalFrom(r.Context())
	}))
	call := func(t *testing.T, token string, want ...string) {
		t.Helper()
		r := httptest.NewRequest("GET", "/api/x", nil)
		r.AddCookie(&http.Cookie{Name: CookieName, Value: token})
		got = Principal{}
		counter.take()
		probe.ServeHTTP(httptest.NewRecorder(), r)
		if names := counter.take(); !slices.Equal(names, want) {
			t.Errorf("queries %q, want %q", names, want)
		}
	}

	vetToken := e.signIn("ceban")
	t.Run("vet, no slide due", func(t *testing.T) {
		call(t, vetToken, "GetActiveSession")
		if got.UserID != vet || !slices.Equal(got.FarmIDs, []uuid.UUID{hirb}) {
			t.Errorf("principal %+v", got)
		}
	})
	t.Run("manager, farms in name_key order", func(t *testing.T) {
		call(t, e.signIn("lungu"), "GetActiveSession")
		if got.UserID != mgr || !slices.Equal(got.FarmIDs, []uuid.UUID{cioc, tel}) {
			t.Errorf("principal %+v, want farms Ciocana, Telița", got)
		}
	})
	t.Run("owner", func(t *testing.T) {
		call(t, e.signIn("rusu"), "GetActiveSession")
		if got.Role != Owner || got.FarmIDs != nil {
			t.Errorf("owner principal %+v (FarmIDs must be nil)", got)
		}
	})
	t.Run("viewer without farms", func(t *testing.T) {
		call(t, e.signIn("gol"), "GetActiveSession")
		if got.FarmIDs == nil || len(got.FarmIDs) != 0 {
			t.Errorf("principal %+v (FarmIDs must be empty, not nil)", got)
		}
	})
	t.Run("vet, slide due", func(t *testing.T) {
		e.d.Clock.Advance(SlideEvery + time.Minute)
		call(t, vetToken, "GetActiveSession", "SlideSession")
		if got.UserID != vet || !slices.Equal(got.FarmIDs, []uuid.UUID{hirb}) {
			t.Errorf("principal %+v", got)
		}
	})
	t.Run("unknown token", func(t *testing.T) {
		random, _, _ := newToken()
		call(t, random, "GetActiveSession")
		if got.UserID != uuid.Nil {
			t.Errorf("a principal for an unknown token: %+v", got)
		}
	})
}

func TestSessionRefusals(t *testing.T) {
	e := newEnv(t)
	id := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	me := func(cookie string) *httptest.ResponseRecorder {
		return e.do(req{method: "GET", path: "/api/auth/me", cookie: cookie})
	}
	random, _, _ := newToken()

	assertError(t, me(""), 401, unauthenticated)
	assertError(t, me("not-a-token"), 401, unauthenticated)
	assertError(t, me(random), 401, unauthenticated) // well-formed, unknown

	t.Run("revoked", func(t *testing.T) {
		tok := e.signIn("ceban")
		e.exec(`update user_session set revoked_at = $1`, testdb.Now())
		assertError(t, me(tok), 401, unauthenticated)
	})
	t.Run("expired", func(t *testing.T) {
		tok := e.signIn("ceban")
		e.exec(`update user_session set expires_at = $1 where revoked_at is null`, e.d.Clock.Now())
		assertError(t, me(tok), 401, unauthenticated)
	})
	t.Run("deactivated user", func(t *testing.T) {
		tok := e.signIn("ceban")
		if w := me(tok); w.Code != 200 {
			t.Fatalf("before: %d", w.Code)
		}
		e.exec(`update app_user set active = false where id = $1`, id)
		assertError(t, me(tok), 401, unauthenticated)
	})
}

func TestSessionSlides(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	tok := e.signIn("ceban")
	session := func() (lastSeen, expires time.Time) {
		t.Helper()
		if err := e.d.App.QueryRow(context.Background(), `select last_seen_at, expires_at from user_session`).Scan(&lastSeen, &expires); err != nil {
			t.Fatal(err)
		}
		return
	}
	login := testdb.Now()

	e.d.Clock.Advance(30 * time.Minute)
	w := e.do(req{method: "GET", path: "/api/auth/me", cookie: tok})
	if w.Code != 200 || w.Header().Get("Set-Cookie") != "" {
		t.Errorf("after 30 min: %d, Set-Cookie %q", w.Code, w.Header().Get("Set-Cookie"))
	}
	if ls, ex := session(); !ls.Equal(login) || !ex.Equal(login.Add(SessionLifetime)) {
		t.Errorf("after 30 min the session moved: last seen %v, expires %v", ls, ex)
	}

	e.d.Clock.Advance(31 * time.Minute) // 61 min after login
	now := e.d.Clock.Now()
	w = e.do(req{method: "GET", path: "/api/auth/me", cookie: tok})
	if w.Code != 200 {
		t.Fatalf("after 61 min: %d", w.Code)
	}
	if got := sessionToken(t, w); got != tok {
		t.Error("the re-sent cookie must carry the same token")
	}
	if ls, ex := session(); !ls.Equal(now) || !ex.Equal(now.Add(SessionLifetime)) {
		t.Errorf("after 61 min: last seen %v, expires %v; want %v and +30 d", ls, ex, now)
	}

	// Within the next hour nothing slides again.
	e.d.Clock.Advance(10 * time.Minute)
	if w := e.do(req{method: "GET", path: "/api/auth/me", cookie: tok}); w.Header().Get("Set-Cookie") != "" {
		t.Error("slid twice within an hour")
	}
}

func TestSessionOutlivesLoginWhileUsed(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	tok := e.signIn("ceban")
	for range 4 { // 4 × 10 days of use: past the first 30 days
		e.d.Clock.Advance(10 * 24 * time.Hour)
		if w := e.do(req{method: "GET", path: "/api/auth/me", cookie: tok}); w.Code != 200 {
			t.Fatalf("day %v: %d", e.d.Clock.Now(), w.Code)
		}
	}
	e.d.Clock.Advance(31 * 24 * time.Hour) // unused for 31 days
	assertError(t, e.do(req{method: "GET", path: "/api/auth/me", cookie: tok}), 401, unauthenticated)
}

func TestLogout(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	tok := e.signIn("ceban")
	other := e.signIn("ceban") // another device stays signed in

	w := e.do(req{method: "POST", path: "/api/auth/logout", cookie: tok})
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("logout: %d %s", w.Code, w.Body)
	}
	if got := w.Header().Get("Set-Cookie"); got != "il_session=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax" {
		t.Errorf("Set-Cookie %q", got)
	}
	var revoked int
	_ = e.d.App.QueryRow(context.Background(), `select count(*) from user_session where revoked_at = $1`, testdb.Now()).Scan(&revoked)
	if revoked != 1 {
		t.Errorf("%d sessions revoked, want 1", revoked)
	}
	assertError(t, e.do(req{method: "GET", path: "/api/auth/me", cookie: tok}), 401, unauthenticated)
	if w := e.do(req{method: "GET", path: "/api/auth/me", cookie: other}); w.Code != 200 {
		t.Errorf("other device: %d", w.Code)
	}
	assertError(t, e.do(req{method: "POST", path: "/api/auth/logout"}), 401, unauthenticated)
}

func TestRevokeUserSessions(t *testing.T) {
	e := newEnv(t)
	id := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	e.user("rusu", "Ion Rusu", dbq.UserRoleOwner)
	a, b, keep := e.signIn("ceban"), e.signIn("ceban"), e.signIn("rusu")

	n, err := e.svc.RevokeUserSessions(context.Background(), dbq.New(e.d.App), id)
	if err != nil || n != 2 {
		t.Fatalf("revoked %d, %v; want 2", n, err)
	}
	for _, tok := range []string{a, b} {
		assertError(t, e.do(req{method: "GET", path: "/api/auth/me", cookie: tok}), 401, unauthenticated)
	}
	if w := e.do(req{method: "GET", path: "/api/auth/me", cookie: keep}); w.Code != 200 {
		t.Errorf("another user's session: %d", w.Code)
	}
}

type routeRecorder struct {
	patterns []string
	handlers []http.Handler
}

func (r *routeRecorder) Handle(p string, h http.Handler) {
	r.patterns = append(r.patterns, p)
	r.handlers = append(r.handlers, h)
}

// TestRoutes: login is on the public chain, logout and me on the authed one
// (plan D10). Each chain ends in a marker that answers without calling the
// handler, so the test needs no database.
func TestRoutes(t *testing.T) {
	marker := func(name string) httpx.Middleware {
		return func(http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(name)) })
		}
	}
	var rr routeRecorder
	NewHandler(&Service{}).Routes(&rr, httpx.NewChain(marker("public")), httpx.NewChain(marker("authed")))
	want := []string{"POST /api/auth/login", "POST /api/auth/logout", "GET /api/auth/me"}
	if !slices.Equal(rr.patterns, want) {
		t.Fatalf("routes %q, want %q", rr.patterns, want)
	}
	for i, chain := range []string{"public", "authed", "authed"} {
		w := httptest.NewRecorder()
		rr.handlers[i].ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Body.String() != chain {
			t.Errorf("%s runs on the %q chain, want %q", rr.patterns[i], w.Body, chain)
		}
	}
}

// TestLoginThrottleConcurrent: a burst of wrong passwords for one login can't
// get past the limit, because each attempt is recorded before its password is
// checked, under the login's advisory lock.
func TestLoginThrottleConcurrent(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)

	const n = 20
	codes := make(chan int, n)
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() { codes <- e.login("ceban", "gresita").Code })
	}
	wg.Wait()
	close(codes)
	count := map[int]int{}
	for c := range codes {
		count[c]++
	}
	if count[401] != maxLoginFailures || count[429] != n-maxLoginFailures || len(count) != 2 {
		t.Errorf("statuses %v, want %d × 401 and %d × 429", count, maxLoginFailures, n-maxLoginFailures)
	}
	if got := len(e.attempts()); got != maxLoginFailures {
		t.Errorf("%d attempts recorded, want %d", got, maxLoginFailures)
	}
}

// TestLoginThrottleBurstRefusesCorrectGuess holds the password checks of a
// burst open: once 5 attempts are in flight, every other attempt, the right
// password included, is refused before any password is checked.
func TestLoginThrottleBurstRefusesCorrectGuess(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	release := make(chan struct{})
	e.svc.hash.idKey = func(_, _ []byte, _, _ uint32, _ uint8, n uint32) []byte {
		<-release
		return make([]byte, n) // matches no stored hash
	}

	const n = 30
	codes := make(chan int, n)
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() { codes <- e.login("ceban", "gresita").Code })
	}
	timeout := time.After(20 * time.Second)
	for i := range n - maxLoginFailures {
		select {
		case c := <-codes:
			if c != 429 {
				t.Errorf("early answer %d: %d, want 429", i+1, c)
			}
		case <-timeout:
			close(release)
			t.Fatalf("only %d of %d attempts were refused while 5 checks were in flight", i, n-maxLoginFailures)
		}
	}

	w := e.login("ceban", goodPassword)
	assertLimited(t, w, 300)

	close(release)
	wg.Wait()
	close(codes)
	for c := range codes {
		if c != 401 {
			t.Errorf("held attempt answered %d, want 401", c)
		}
	}
	if got := len(e.attempts()); got != maxLoginFailures {
		t.Errorf("%d attempts recorded, want %d", got, maxLoginFailures)
	}
}

func TestLoginControlCharacters(t *testing.T) {
	e := newEnv(t)
	e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	for _, login := range []string{"ce\x00ban", "\x00", "ceban\x07", "ce\u0085ban"} {
		w := e.login(login, goodPassword)
		assertError(t, w, 401, `{"error":{"code":"UNAUTHENTICATED","message":"Login sau parolă greșită."}}`)
	}
	want := []string{"ce�ban", "�", "ceban�", "ce�ban"}
	got := e.attempts()
	if len(got) != len(want) {
		t.Fatalf("attempts %+v", got)
	}
	for i, a := range got {
		if a.Login != want[i] || a.Success {
			t.Errorf("attempt %d stored %q success %v, want %q", i, a.Login, a.Success, want[i])
		}
	}
}

func TestStoredText(t *testing.T) {
	long := strings.Repeat("ă", 150) // 300 bytes
	for _, tc := range []struct{ in, want string }{
		{"ceban", "ceban"},
		{"a\x00b\tc\nd", "a�b�c�d"},
		{"bad\xffutf8", "bad�utf8"},
		{long, strings.Repeat("ă", 100)},
	} {
		if got := storedText(tc.in, maxStoredLogin); got != tc.want {
			t.Errorf("storedText(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestLoginNamesUserInAccessLog(t *testing.T) {
	e := newEnv(t)
	id := e.user("ceban", "Dr. Ceban", dbq.UserRoleVet)
	var buf strings.Builder
	var mu sync.Mutex
	logger := slog.New(slog.NewJSONHandler(writerFunc(func(p []byte) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		return buf.Write(p)
	}), nil))
	h := httpx.NewChain(httpx.AccessLog(logger, e.d.Clock)).Then(e.h)

	hr := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"login":"ceban","password":"`+goodPassword+`"}`))
	hr.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, hr)
	if w.Code != 200 {
		t.Fatalf("login %d %s", w.Code, w.Body)
	}
	var line map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &line); err != nil {
		t.Fatalf("log %q: %v", buf.String(), err)
	}
	if line["user_id"] != id.String() || line["status"] != 200.0 {
		t.Errorf("access log %v, want user_id %s", line, id)
	}
	if strings.Contains(buf.String(), goodPassword) || strings.Contains(buf.String(), sessionToken(t, w)) {
		t.Error("the access log carries the password or the session token")
	}
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
