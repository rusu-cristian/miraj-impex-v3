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
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
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
	// path builds the concrete path (wildcards filled, query added) from the
	// farms the test creates at run time; nil = route's path.
	path func(matrixFarms) string
	// signIn: the body is the caller's own credentials (login, password).
	signIn bool
	// account: a 200 body is the caller's account (user, farms).
	account bool
	want    map[caller]int
	// code is the error.code a caller's refusal must carry, where the status
	// alone can't tell (403 FORBIDDEN vs FARM_FORBIDDEN).
	code map[caller]string
	// also are the same route on other targets (another farm, an unknown
	// id), each with its own statuses. They belong to the row, so deleting
	// the row still leaves its route without one.
	also []matrixCase
}

// matrixCase is a row's route on another target.
type matrixCase struct {
	name string // the target, in the subtest name ("Telița")
	path func(matrixFarms) string
	want map[caller]int
	code map[caller]string
}

// matrixFarms are the farms TestAuthMatrix creates; their ids exist only at
// run time, so farm-scoped rows build their paths from them.
type matrixFarms struct {
	hirb, tel, cojusna uuid.UUID
}

// unknownFarm is a farm id that no farm has.
var unknownFarm = uuid.MustParse("0b9e4b8e-6f0a-4c51-9d3e-2f7a1c5d8e90")

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

	// The farm scope pages (M2 D4): viewer, manager and owner, each on the
	// farms they may see. Vets never reach them (403 FORBIDDEN); the viewer
	// of Telița is refused Hîrbovăț and Cojușna (403 FARM_FORBIDDEN), and so
	// is anyone but the owner on an unknown id, which the owner gets as 404.
	farmRow("GET /api/table", func(id uuid.UUID) string { return "/api/table?view=farm&farmId=" + id.String() }),
	farmRow("GET /api/farms/{id}/page", func(id uuid.UUID) string { return "/api/farms/" + id.String() + "/page" }),
	farmRow("GET /api/farms/{id}/summary", func(id uuid.UUID) string { return "/api/farms/" + id.String() + "/summary" }),

	// The shell bootstrap (M2 D7): every active session, scoped to its farms.
	{route: "GET /api/bootstrap", want: map[caller]int{
		anonymous: 401, vet: 200, viewer: 200, manager: 200, owner: 200,
		revoked: 401, expired: 401, deactivated: 401,
	}},
}

// farmRow is the row of a farm-scoped read: path builds the request for a
// farm id. The main target is Hîrbovăț (the viewer's other farm); also are
// Telița (the viewer's farm), Cojușna (the owner's only) and an unknown id.
func farmRow(route string, path func(uuid.UUID) string) matrixRow {
	at := func(pick func(matrixFarms) uuid.UUID) func(matrixFarms) string {
		return func(f matrixFarms) string { return path(pick(f)) }
	}
	// scoped are the statuses with the session check and RequireRole(viewer,
	// manager, owner) in front of the farm check.
	scoped := func(viewerStatus, managerStatus, ownerStatus int) map[caller]int {
		return map[caller]int{
			anonymous: 401, vet: 403, viewer: viewerStatus, manager: managerStatus, owner: ownerStatus,
			revoked: 401, expired: 401, deactivated: 401,
		}
	}
	return matrixRow{
		route: route,
		path:  at(func(f matrixFarms) uuid.UUID { return f.hirb }),
		want:  scoped(403, 200, 200),
		code:  map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden},
		also: []matrixCase{
			{name: "Telița", path: at(func(f matrixFarms) uuid.UUID { return f.tel }),
				want: scoped(200, 200, 200),
				code: map[caller]string{vet: httpx.CodeForbidden}},
			{name: "Cojușna", path: at(func(f matrixFarms) uuid.UUID { return f.cojusna }),
				want: scoped(403, 403, 200),
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden, manager: httpx.CodeFarmForbidden}},
			{name: "unknown farm", path: func(matrixFarms) string { return path(unknownFarm) },
				want: scoped(403, 403, 404),
				code: map[caller]string{vet: httpx.CodeForbidden, viewer: httpx.CodeFarmForbidden,
					manager: httpx.CodeFarmForbidden, owner: httpx.CodeNotFound}},
		},
	}
}

// matrixGaps compares the registered routes with the table: missing are
// routes without a row, stale are rows naming no registered route (or a
// route a previous row already has: one row per route, so deleting a row
// always leaves its route missing), partial are rows or cases that lack a
// caller.
func matrixGaps(routes []string, rows []matrixRow) (missing, stale, partial []string) {
	have := make(map[string]bool, len(rows))
	for _, r := range rows {
		if have[r.route] || !slices.Contains(routes, r.route) {
			stale = append(stale, r.route)
		}
		have[r.route] = true
		for _, c := range callers {
			if _, ok := r.want[c]; !ok {
				partial = append(partial, r.route+" × "+string(c))
			}
			for _, a := range r.also {
				if _, ok := a.want[c]; !ok {
					partial = append(partial, r.route+" ("+a.name+") × "+string(c))
				}
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
	ids := matrixFarms{hirb: hirb, tel: tel, cojusna: s.farm(t, "Cojușna", "cojusna")}
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
		if row.path != nil {
			path = row.path(ids)
		}
		type target struct {
			name, path string
			want       map[caller]int
			code       map[caller]string
		}
		cases := []target{{path: path, want: row.want, code: row.code}}
		for _, a := range row.also {
			cases = append(cases, target{name: a.name, path: a.path(ids), want: a.want, code: a.code})
		}
		for _, mc := range cases {
			name := row.route
			if mc.name != "" {
				name += " (" + mc.name + ")"
			}
			for _, c := range callers {
				want, ok := mc.want[c]
				if !ok {
					continue // reported above
				}
				t.Run(name+"/"+string(c), func(t *testing.T) {
					t.Parallel()
					req := call{method: method, path: mc.path, cookie: newCookie(t, c)}
					if row.signIn {
						req.body = loginBody(accounts[c].login, testPassword)
					}
					r := s.do(t, req)
					if r.status != want {
						t.Fatalf("%s as %s: %s, want %d", name, c, r, want)
					}
					if code, ok := mc.code[c]; ok {
						if got := errorCode(t, r); got != code {
							t.Fatalf("%s as %s: %s, want error.code %s", name, c, r, code)
						}
					}
					if row.account && r.status == http.StatusOK {
						checkAccount(t, accounts[c], r)
					}
				})
			}
		}
	}
}

// errorCode is the error.code of an error envelope ("" for another body).
func errorCode(t *testing.T, r response) string {
	t.Helper()
	var b struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(r.body), &b); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	return b.Error.Code
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

	// A row naming an unregistered route, and a second row for a route (both
	// stale), and a row without every caller.
	extra := append(slices.Clone(authMatrix),
		matrixRow{route: "GET /api/nu-exista", want: everyone(http.StatusOK)},
		matrixRow{route: "GET /healthz", want: map[caller]int{anonymous: 200}})
	_, stale, partial := matrixGaps(a.routes, extra)
	if !slices.Equal(stale, []string{"GET /api/nu-exista", "GET /healthz"}) {
		t.Errorf("stale %q, want the unregistered route and the second healthz row", stale)
	}
	if len(partial) != len(callers)-1 || !strings.HasPrefix(partial[0], "GET /healthz × ") {
		t.Errorf("partial %q, want the healthz row's %d missing callers", partial, len(callers)-1)
	}

	// A case without every caller is partial too, named with its target.
	row := farmRow("GET /api/table", func(uuid.UUID) string { return "/api/table" })
	delete(row.also[0].want, viewer)
	_, _, partial = matrixGaps(a.routes, []matrixRow{row})
	if !slices.Equal(partial, []string{"GET /api/table (Telița) × viewer"}) {
		t.Errorf("partial %q, want the Telița case's viewer", partial)
	}
}
