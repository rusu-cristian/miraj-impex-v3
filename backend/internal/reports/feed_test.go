package reports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// fixture is a hand-built set of three farms, today 18.11.2026:
//
//	Alfa: open cycle (Seria 2026, from 01.11.2026), section A1 with ♀ 100 placed 01.11.
//	  01.11 deliveries Breeder 1 40 kg, Breeder 2 100 kg.
//	  16.11, 17.11, 18.11 Hrană zilnică A1 ♀ Breeder 2 100 g → 10 kg a day (the window 16–18.11
//	  ends today: the only row with birds is fed today).
//	  10.11 Hrană zilnică A1 ♀ Starter (ambele sexe) 5 kg in the register, never delivered.
//	Beta: one cycle (Seria 2025, 01.03.2025–28.02.2026), closed.
//	Gama: never a cycle.
//
// Alfa's stock at the end of 18.11: Breeder 1 40 kg (no consumption), Breeder 2 100 − 30 = 70 kg (avg 10,
// 7 days → 25.11), Starter −5 kg.
type fixture struct {
	t                *testing.T
	d                *testdb.DB
	user             uuid.UUID
	alfa, beta, gama uuid.UUID
}

func day(m time.Month, d int) timex.Day { return timex.MustDay(2026, m, d) }

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t)}
	f.user = f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	s26 := f.id(`insert into series (name, name_key, year, placement_date) values ('Seria 2026', 'seria 2026', 2026, '2026-11-01')`)
	s25 := f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, '2025-03-01', 'closed', $1)`, testdb.Now())
	// Inserted out of name order: the report orders farms by name.
	f.gama = f.id(`insert into farm (name, name_key, type) values ('Gama', 'gama', 'laying')`)
	f.beta = f.id(`insert into farm (name, name_key, type) values ('Beta', 'beta', 'growth')`)
	f.alfa = f.id(`insert into farm (name, name_key, type) values ('Alfa', 'alfa', 'laying')`)
	a1 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'A1', 1)`, f.alfa)
	f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2025-03-01', '2026-02-28', 'auto', 'emptied')`, s25, f.beta)
	f.id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, '2026-11-01', 'auto')`, s26, f.alfa)

	breeder1 := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 1', 'breeder 1', 'F')`)
	breeder2 := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')`)
	starter := f.id(`insert into feed_type (name, name_key) values ('Starter', 'starter')`)

	placed := day(time.November, 1)
	e := f.id(`insert into entry (kind, farm_id, date, series_id, created_by) values ('placement', $1, $2, $3, $4)`,
		f.alfa, placed, s26, f.user)
	f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		values ($1, $2, $3, $4, 'F', 'placement', 100)`, e, f.alfa, a1, placed)
	e = f.entry("delivery", placed)
	f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 40)`,
		e, f.alfa, breeder1, placed)
	f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 100)`,
		e, f.alfa, breeder2, placed)
	for _, d := range []int{16, 17, 18} {
		date := day(time.November, d)
		f.id(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
			values ($1, $2, $3, $4, 'F', $5, 100)`, f.entry("feed_daily", date), f.alfa, a1, date, breeder2)
	}
	date := day(time.November, 10)
	f.id(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, kg_recorded)
		values ($1, $2, $3, $4, 'F', $5, 5)`, f.entry("feed_daily", date), f.alfa, a1, date, starter)
	return f
}

func (f *fixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *fixture) entry(kind string, d timex.Day) uuid.UUID {
	f.t.Helper()
	return f.id(`insert into entry (kind, farm_id, date, created_by) values ($1, $2, $3, $4)`, kind, f.alfa, d, f.user)
}

func (f *fixture) input(p auth.Principal, farm *uuid.UUID) Input {
	return Input{Principal: p, FarmID: farm, Today: testdb.Today, Now: testdb.Now()}
}

func owner() auth.Principal { return auth.Principal{UserID: uuid.New(), Role: auth.Owner} }

func principal(role auth.Role, farms ...uuid.UUID) auth.Principal {
	if farms == nil {
		farms = []uuid.UUID{}
	}
	return auth.Principal{UserID: uuid.New(), Role: role, FarmIDs: farms}
}

// Rows of the fixture's report, as encoded.
const (
	alfaRows = `{"farm":"Alfa","type":"Breeder 1","forSex":"F","kg":40,"daysLeft":null,"avgDaily":0,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":null,"lamp":"oprit","reason":"– fără consum în ultimele 3 zile"},` +
		`{"farm":"Alfa","type":"Breeder 2","forSex":"F","kg":70,"daysLeft":7,"avgDaily":10,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":"2026-11-25","lamp":"in_norma"},` +
		`{"farm":"Alfa","type":"Starter","forSex":null,"kg":-5,"daysLeft":0,"avgDaily":0,` +
		`"window":["2026-11-16","2026-11-18"],"runOut":"2026-11-18","lamp":"abatere","reason":"stoc sub 0 · verifică livrările"}`
	betaRow = `{"farm":"Beta","lamp":"oprit","reason":"fără ciclu deschis","sub":"ultimul ciclu: Seria 2025, închis pe 28.02.2026"}`
	gamaRow = `{"farm":"Gama","lamp":"oprit","reason":"fără ciclu deschis"}`
	asOf    = `"asOf":"2026-11-18T12:00:00+02:00"`
)

func TestBuildFeed(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name string
		in   Input
		want string
	}{
		{"owner, every farm", f.input(owner(), nil), `{"stock":[` + alfaRows + `,` + betaRow + `,` + gamaRow + `],` + asOf + `}`},
		{"manager, own farms", f.input(principal(auth.Manager, f.gama, f.beta), nil), `{"stock":[` + betaRow + `,` + gamaRow + `],` + asOf + `}`},
		{"viewer without farms", f.input(principal(auth.Viewer), nil), `{"stock":[],` + asOf + `}`},
		{"one farm", f.input(owner(), &f.alfa), `{"stock":[` + alfaRows + `],` + asOf + `}`},
		{"one farm without a cycle", f.input(principal(auth.Viewer, f.gama), &f.gama), `{"stock":[` + gamaRow + `],` + asOf + `}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res, err := BuildFeed(t.Context(), f.d.App, tc.in)
			if err != nil {
				t.Fatal(err)
			}
			b, err := json.Marshal(res)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Errorf("got  %s\nwant %s", b, tc.want)
			}
		})
	}

	unknown := uuid.New()
	if _, err := BuildFeed(t.Context(), f.d.App, f.input(owner(), &unknown)); !errors.Is(err, ErrFarmNotFound) {
		t.Errorf("unknown farm: err = %v, want ErrFarmNotFound", err)
	}
}

// server serves the reports routes, the caller being p (nil: no session
// middleware ran).
func (f *fixture) server(p *auth.Principal) http.Handler {
	mux := http.NewServeMux()
	withCaller := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p != nil {
				r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
			}
			next.ServeHTTP(w, r)
		})
	}
	NewHandler(NewService(db.NewStore(f.d.App), f.d.Clock)).Routes(mux, httpx.NewChain(withCaller))
	return mux
}

type errorBody struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Issues  []httpx.Issue `json:"issues"`
	} `json:"error"`
}

func TestFeedHandler(t *testing.T) {
	f := newFixture(t)
	alfa, beta := f.alfa.String(), f.beta.String()
	vet := principal(auth.Vet, f.alfa)
	viewer := principal(auth.Viewer, f.beta)
	manager := principal(auth.Manager, f.alfa, f.beta)
	own := owner()
	const path = "/api/reports/feed"

	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		query     string
		status    int
		code      string
		issuePath string
		issueCode string
		message   string
		body      string // the whole 200 body, without the trailing newline
	}{
		{name: "no session", query: "", status: 401, code: "UNAUTHENTICATED"},
		{name: "vet", caller: &vet, status: 403, code: "FORBIDDEN"},
		{name: "viewer on another farm", caller: &viewer, query: "?farmId=" + alfa, status: 403, code: "FARM_FORBIDDEN",
			message: "Nu ai acces la această fermă."},
		{name: "viewer on an unknown farm", caller: &viewer, query: "?farmId=" + uuid.NewString(), status: 403, code: "FARM_FORBIDDEN"},
		{name: "owner on an unknown farm", caller: &own, query: "?farmId=" + uuid.NewString(), status: 404, code: "NOT_FOUND"},
		{name: "malformed farm id", caller: &manager, query: "?farmId=alfa", status: 400, code: "VALIDATION",
			issuePath: "farmId", issueCode: "bad_type"},
		{name: "repeated farm id", caller: &manager, query: "?farmId=all&farmId=all", status: 400, code: "VALIDATION",
			issuePath: "farmId", issueCode: "bad_type"},
		{name: "unknown parameter", caller: &manager, query: "?view=stock", status: 400, code: "VALIDATION",
			issuePath: "view", issueCode: "unknown_field"},
		{name: "bad period", caller: &manager, query: "?period=weeks:x", status: 400, code: "VALIDATION",
			issuePath: "period", issueCode: "bad_type"},
		{name: "unknown period", caller: &manager, query: "?period=month", status: 400, code: "VALIDATION",
			issuePath: "period", issueCode: "out_of_range"},
		{name: "interval in the future", caller: &manager, query: "?period=interval:2026-11-01,2026-11-20", status: 400,
			code: "VALIDATION", issuePath: "period", issueCode: "out_of_range",
			message: "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{name: "viewer, own farms", caller: &viewer, status: 200, body: `{"stock":[` + betaRow + `],` + asOf + `}`},
		{name: "manager, all, a period", caller: &manager, query: "?farmId=all&period=weeks:4", status: 200,
			body: `{"stock":[` + alfaRows + `,` + betaRow + `],` + asOf + `}`},
		{name: "manager, one farm", caller: &manager, query: "?farmId=" + beta + "&period=interval:2026-11-01,2026-11-18",
			status: 200, body: `{"stock":[` + betaRow + `],` + asOf + `}`},
		{name: "owner", caller: &own, query: "?period=cycle", status: 200,
			body: `{"stock":[` + alfaRows + `,` + betaRow + `,` + gamaRow + `],` + asOf + `}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			f.server(tc.caller).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path+tc.query, nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			if tc.status == http.StatusOK {
				if got := rec.Body.String(); got != tc.body+"\n" {
					t.Errorf("body\n got %s\nwant %s", got, tc.body)
				}
				return
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Errorf("code = %s, want %s", body.Error.Code, tc.code)
			}
			if tc.issuePath != "" && (len(body.Error.Issues) != 1 || body.Error.Issues[0].Path != tc.issuePath ||
				body.Error.Issues[0].Code != tc.issueCode) {
				t.Errorf("issues = %+v, want one %s %s", body.Error.Issues, tc.issuePath, tc.issueCode)
			}
			if tc.message != "" && body.Error.Message != tc.message {
				t.Errorf("message = %q, want %q", body.Error.Message, tc.message)
			}
		})
	}
}
