package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The auth rows of the SPEC §3.3 permission matrix (§11.8, plan D6): every
// route `api serve` registers, called by every kind of caller through the
// full handler chain. A route registered without a row here fails
// TestAuthMatrix with the route named; later milestones add their rows.

// caller is who makes a request.
type caller string

const (
	anonymous   caller = "anonymous"   // no cookie; on login it signs in with the vet's credentials
	vet         caller = "vet"         // open session; farm Hîrbovăț
	viewer      caller = "viewer"      // open session; farm Telița
	manager     caller = "manager"     // open session; farms Hîrbovăț and Telița
	owner       caller = "owner"       // open session; no user_farm rows, sees every farm
	revoked     caller = "revoked"     // a revoked session of an active vet
	expired     caller = "expired"     // an expired session of an active vet
	deactivated caller = "deactivated" // an open session of a deactivated vet
)

var callers = []caller{anonymous, vet, viewer, manager, owner, revoked, expired, deactivated}

// matrixRow is one route and the status each caller gets.
type matrixRow struct {
	route string // the pattern exactly as registered ("POST /api/auth/login")
	path  string // a concrete path when route has wildcards; "" = route's path
	// signIn: the body is the caller's own credentials (login, password).
	signIn bool
	// account: a 200 body is the caller's account (user, farms).
	account bool
	want    map[caller]int
}

func everyone(status int) map[caller]int {
	m := make(map[caller]int, len(callers))
	for _, c := range callers {
		m[c] = status
	}
	return m
}

var authMatrix = []matrixRow{
	// Public, outside /api (SPEC §11.12): 200 whatever the cookie.
	{route: "GET /healthz", want: everyone(http.StatusOK)},
	{route: "GET /readyz", want: everyone(http.StatusOK)},

	// The only /api route without a session (§7.0). Every active account signs
	// in, whatever state its old session is in; a deactivated one is refused
	// after its correct password (§6.16).
	{route: "POST /api/auth/login", signIn: true, account: true, want: map[caller]int{
		anonymous: 200, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 200, expired: 200, deactivated: http.StatusForbidden,
	}},
	{route: "POST /api/auth/logout", want: map[caller]int{
		anonymous: 401, vet: 204, viewer: 204, manager: 204, owner: 204,
		revoked: 401, expired: 401, deactivated: 401,
	}},
	{route: "GET /api/auth/me", account: true, want: map[caller]int{
		anonymous: 401, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 401, expired: 401, deactivated: 401,
	}},
}

// matrixGaps compares the registered routes with the table: missing are
// routes without a row, stale are rows naming no registered route, partial
// are rows that lack a caller.
func matrixGaps(routes []string, rows []matrixRow) (missing, stale, partial []string) {
	have := make(map[string]bool, len(rows))
	for _, r := range rows {
		have[r.route] = true
		if !slices.Contains(routes, r.route) {
			stale = append(stale, r.route)
		}
		for _, c := range callers {
			if _, ok := r.want[c]; !ok {
				partial = append(partial, r.route+" × "+string(c))
			}
		}
	}
	for _, route := range routes {
		if !have[route] {
			missing = append(missing, route)
		}
	}
	return missing, stale, partial
}

// account is what login and me must return for a caller.
type account struct {
	id    uuid.UUID
	login string
	role  string
	farms []string // names, in name_key order
}

type accountBody struct {
	User struct {
		ID    uuid.UUID `json:"id"`
		Login string    `json:"login"`
		Role  string    `json:"role"`
	} `json:"user"`
	Farms []struct {
		Name string `json:"name"`
	} `json:"farms"`
	Today string `json:"today"`
}

func TestAuthMatrix(t *testing.T) {
	t.Parallel()
	s := newStack(t)

	missing, stale, partial := matrixGaps(s.app.routes, authMatrix)
	for _, r := range missing {
		t.Errorf("route %q is registered but has no row in authMatrix (SPEC §3.3)", r)
	}
	for _, r := range stale {
		t.Errorf("authMatrix row %q names no registered route", r)
	}
	for _, r := range partial {
		t.Errorf("authMatrix has no expected status for %s", r)
	}

	// Farms and accounts. Cojușna has no user_farm rows: only the owner sees it.
	hirb := s.farm(t, "Hîrbovăț", "hirbovat")
	tel := s.farm(t, "Telița", "telita")
	s.farm(t, "Cojușna", "cojusna")
	accounts := map[caller]account{}
	add := func(c caller, login, name string, role dbq.UserRole, active bool, farms []uuid.UUID, names ...string) {
		accounts[c] = account{id: s.user(t, login, name, role, active, farms...), login: login, role: string(role), farms: names}
	}
	add(vet, "ceban", "Dr. Ceban", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(viewer, "observator", "Ana Observator", dbq.UserRoleViewer, true, []uuid.UUID{tel}, "Telița")
	add(manager, "ionescu", "Ion Ionescu", dbq.UserRoleManager, true, []uuid.UUID{tel, hirb}, "Hîrbovăț", "Telița")
	add(owner, "proprietar", "Petru Proprietar", dbq.UserRoleOwner, true, nil, "Cojușna", "Hîrbovăț", "Telița")
	add(revoked, "revocat", "Vet Revocat", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(expired, "expirat", "Vet Expirat", dbq.UserRoleVet, true, []uuid.UUID{hirb}, "Hîrbovăț")
	add(deactivated, "dezactivat", "Vet Dezactivat", dbq.UserRoleVet, false, []uuid.UUID{hirb}, "Hîrbovăț")
	accounts[anonymous] = accounts[vet]

	// Each cell gets its own session, so cells don't depend on each other
	// (a logout cell revokes only its own).
	newCookie := func(t *testing.T, c caller) string {
		switch c {
		case anonymous:
			return ""
		case revoked:
			return s.session(t, accounts[c].id, sessionRevoked)
		case expired:
			return s.session(t, accounts[c].id, sessionExpired)
		default:
			return s.session(t, accounts[c].id, sessionOpen)
		}
	}

	for _, row := range authMatrix {
		method, path, _ := strings.Cut(row.route, " ")
		if row.path != "" {
			path = row.path
		}
		for _, c := range callers {
			want, ok := row.want[c]
			if !ok {
				continue // reported above
			}
			t.Run(row.route+"/"+string(c), func(t *testing.T) {
				t.Parallel()
				req := call{method: method, path: path, cookie: newCookie(t, c)}
				if row.signIn {
					req.body = loginBody(accounts[c].login, testPassword)
				}
				r := s.do(t, req)
				if r.status != want {
					t.Fatalf("%s as %s: %s, want %d", row.route, c, r, want)
				}
				if row.account && r.status == http.StatusOK {
					checkAccount(t, accounts[c], r)
				}
			})
		}
	}
}

// checkAccount fails unless the body names a's user, role and farms.
func checkAccount(t *testing.T, a account, r response) {
	t.Helper()
	var b accountBody
	if err := json.Unmarshal([]byte(r.body), &b); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	if b.User.ID != a.id || b.User.Login != a.login || b.User.Role != a.role {
		t.Errorf("user %+v, want id %s login %s role %s", b.User, a.id, a.login, a.role)
	}
	var farms []string
	for _, f := range b.Farms {
		farms = append(farms, f.Name)
	}
	if !slices.Equal(farms, a.farms) {
		t.Errorf("farms %q, want %q", farms, a.farms)
	}
	if b.Today != testdb.Today.String() {
		t.Errorf("today %q, want %s", b.Today, testdb.Today)
	}
}

// Deleting any row from authMatrix makes matrixGaps name its route, against
// the routes the real wiring registers.
func TestAuthMatrixCatchesMissingRow(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	a, err := newApp(config.Config{AppOrigin: testOrigin}, log, timex.FixedClock(testdb.Now()), &db.Pools{})
	if err != nil {
		t.Fatal(err)
	}
	for i, row := range authMatrix {
		rows := slices.Delete(slices.Clone(authMatrix), i, i+1)
		missing, stale, partial := matrixGaps(a.routes, rows)
		if !slices.Equal(missing, []string{row.route}) || stale != nil || partial != nil {
			t.Errorf("without the row %q: missing %q, stale %q, partial %q; want only %q missing",
				row.route, missing, stale, partial, row.route)
		}
	}

	// A row naming an unregistered route, and a row without every caller.
	extra := append(slices.Clone(authMatrix),
		matrixRow{route: "GET /api/nu-exista", want: everyone(http.StatusOK)},
		matrixRow{route: "GET /healthz", want: map[caller]int{anonymous: 200}})
	_, stale, partial := matrixGaps(a.routes, extra)
	if !slices.Equal(stale, []string{"GET /api/nu-exista"}) {
		t.Errorf("stale %q, want the unregistered route", stale)
	}
	if len(partial) != len(callers)-1 || !strings.HasPrefix(partial[0], "GET /healthz × ") {
		t.Errorf("partial %q, want the healthz row's %d missing callers", partial, len(callers)-1)
	}
}
