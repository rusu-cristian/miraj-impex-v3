package periodtable

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// farmL is a hand-built laying farm whose weeks are the §7.4 example's: Seria
// L is placed 07.10.2025 at age 0, so 18.11.2026 is age 407, SAPT 59
// (17–18.11). Its cycle opens 01.09.2026 (SAPT 48), so the period `cycle` is
// SAPT 48–59, 12 weeks; SAPT 52 is 29.09–05.10.
//
//	Ferma L (laying, no guide): H1 (a hall without compartments).
//	01.09 placement H1 ♀ 100 · ♂ 10; a mortality row of 0 every day.
//	Eggs every day from 01.09: hatching 100, nothing else (100 a day).
//
// Ferma A has its own cycle (Seria L from 01.09.2026) and Ferma N none.
type farmL struct {
	d                    *testdb.DB
	farm, other, noCycle uuid.UUID
	otherCycle           uuid.UUID
}

func newFarmL(t *testing.T) *farmL {
	t.Helper()
	d := testdb.New(t)
	ctx := context.Background()
	id := func(sql string, args ...any) uuid.UUID {
		t.Helper()
		var id uuid.UUID
		if err := d.App.QueryRow(ctx, sql+" returning id", args...).Scan(&id); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
		return id
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := d.App.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
	start := day(time.September, 1)
	f := &farmL{d: d}
	series := id(`insert into series (name, name_key, year, placement_date, age_at_placement_days)
		values ('Seria L', 'seria l', 2031, '2025-10-07', 0)`)
	f.farm = id(`insert into farm (name, name_key, type) values ('Ferma L', 'ferma l', 'laying')`)
	f.other = id(`insert into farm (name, name_key, type) values ('Ferma A', 'ferma a', 'laying')`)
	f.noCycle = id(`insert into farm (name, name_key, type) values ('Ferma N', 'ferma n', 'laying')`)
	id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`, series, f.farm, start)
	f.otherCycle = id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`,
		series, f.other, start)
	h1 := id(`insert into section (farm_id, code, sort_order) values ($1, 'H1', 1)`, f.farm)

	e := id(`insert into entry (kind, farm_id, date, series_id) values ('placement', $1, $2, $3)`, f.farm, start, series)
	exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		values ($1, $2, $3, $4, 'F', 'placement', 100), ($1, $2, $3, $4, 'M', 'placement', 10)`, e, f.farm, h1, start)
	e = id(`insert into entry (kind, farm_id, date) values ('daily', $1, $2)`, f.farm, start)
	exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, $3, d::date, x.sex::sex, 'mortality', 0
		from generate_series($4::date, '2026-11-18'::date, interval '1 day') d, (values ('F'), ('M')) x(sex)`,
		e, f.farm, h1, start)
	exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g)
		select $1, $2, $3, d::date, 0, 0, 0, 0, 100, 60 from generate_series($4::date, '2026-11-18'::date, interval '1 day') d`,
		e, f.farm, h1, start)
	return f
}

// server serves the PeriodTable route, the caller being p (nil: no session
// middleware ran).
func (f *farmL) server(p *auth.Principal) http.Handler {
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

func caller(role auth.Role, farms ...uuid.UUID) *auth.Principal {
	if role == auth.Owner {
		return &auth.Principal{UserID: uuid.New(), Role: auth.Owner}
	}
	if farms == nil {
		farms = []uuid.UUID{}
	}
	return &auth.Principal{UserID: uuid.New(), Role: role, FarmIDs: farms}
}

type tableError struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Issues  []httpx.Issue `json:"issues"`
	} `json:"error"`
}

func getTable(t *testing.T, h http.Handler, query string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/table?"+query, nil))
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("%s: Cache-Control = %q, want no-store", query, cc)
	}
	return rec
}

func TestTableHandler(t *testing.T) {
	f := newFarmL(t)
	of := func(farm uuid.UUID) string { return "view=farm&farmId=" + farm.String() }
	q := of(f.farm)
	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		query     string
		status    int
		code      string
		message   string
		issuePath string
		issueCode string
	}{
		{"no session", nil, q, 401, "UNAUTHENTICATED", "Sesiunea a expirat. Intră din nou în cont.", "", ""},
		{"vet", caller(auth.Vet, f.farm), q, 403, "FORBIDDEN", "Nu ai acces la această pagină.", "", ""},
		{"viewer on another farm", caller(auth.Viewer, f.other), q, 403, "FARM_FORBIDDEN", "Nu ai acces la această fermă.", "", ""},
		{"manager on an unknown farm", caller(auth.Manager, f.farm), of(uuid.New()), 403, "FARM_FORBIDDEN", "Nu ai acces la această fermă.", "", ""},
		{"owner on an unknown farm", caller(auth.Owner), of(uuid.New()), 404, "NOT_FOUND", "Adresa cerută nu există.", "", ""},
		{"another farm's cycle", caller(auth.Manager, f.farm, f.other), q + "&cycleId=" + f.otherCycle.String(),
			404, "NOT_FOUND", "Adresa cerută nu există.", "", ""},
		{"unknown parameter", caller(auth.Owner), q + "&page=2", 400, "VALIDATION", "Datele trimise nu sunt valide.", "page", "unknown_field"},
		{"malformed farm id", caller(auth.Owner), "view=farm&farmId=L", 400, "VALIDATION", "Datele trimise nu sunt valide.", "farmId", "bad_type"},
		{"tab not served in M2", caller(auth.Owner), q + "&tab=hrana", 400, "VALIDATION", "Datele trimise nu sunt valide.", "tab", "out_of_range"},
		{"interval in the future", caller(auth.Viewer, f.farm), q + "&period=interval:2026-11-01,2026-11-20",
			400, "VALIDATION", "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026.", "period", "out_of_range"},
		{"before at the period's first week", caller(auth.Viewer, f.farm), q + "&before=48",
			400, "VALIDATION", "Datele trimise nu sunt valide.", "before", "out_of_range"},
		{"sapt outside the period", caller(auth.Viewer, f.farm), q + "&period=weeks:4&sapt=50",
			400, "VALIDATION", "Datele trimise nu sunt valide.", "sapt", "out_of_range"},
		{"viewer on own farm", caller(auth.Viewer, f.farm), q, 200, "", "", "", ""},
		{"manager, weeks:12", caller(auth.Manager, f.farm), q + "&period=weeks:12&grain=week", 200, "", "", "", ""},
		{"owner, a farm with no cycle", caller(auth.Owner), of(f.noCycle), 200, "", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := getTable(t, f.server(tc.caller), tc.query)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status == 200 {
				var res TableResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Fatal(err)
				}
				if res.Scope.FarmID == nil || res.Weeks == nil {
					t.Errorf("not a table: %s", rec.Body)
				}
				return
			}
			var body tableError
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code || body.Error.Message != tc.message {
				t.Errorf("error = %s %q, want %s %q", body.Error.Code, body.Error.Message, tc.code, tc.message)
			}
			if tc.issuePath != "" {
				if len(body.Error.Issues) != 1 || body.Error.Issues[0].Path != tc.issuePath || body.Error.Issues[0].Code != tc.issueCode {
					t.Errorf("issues = %+v, want one %s %s", body.Error.Issues, tc.issuePath, tc.issueCode)
				}
			}
		})
	}

	rec := getTable(t, f.server(caller(auth.Viewer, f.other)), q)
	if want := `{"error":{"code":"FARM_FORBIDDEN","message":"Nu ai acces la această fermă."}}` + "\n"; rec.Body.String() != want {
		t.Errorf("FARM_FORBIDDEN body = %s, want %s", rec.Body, want)
	}
}

func TestTableHandlerLoadMore(t *testing.T) {
	f := newFarmL(t)
	h := f.server(caller(auth.Viewer, f.farm))
	table := func(query string) (TableResponse, map[string]json.RawMessage) {
		t.Helper()
		rec := getTable(t, h, "view=farm&farmId="+f.farm.String()+query)
		if rec.Code != 200 {
			t.Fatalf("%s: status %d: %s", query, rec.Code, rec.Body)
		}
		var res TableResponse
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}
		return res, raw
	}
	sapts := func(res TableResponse) []int {
		out := []int{}
		for _, w := range res.Weeks {
			out = append(out, w.Sapt)
		}
		return out
	}

	// Page 1: SAPT 56–59 (27.10–18.11, 23 days of 100 eggs).
	res, raw := table("")
	diff(t, "page 1 weeks", []int{59, 58, 57, 56}, sapts(res))
	diff(t, "page 1 loaded", &Loaded{SaptFrom: 56, SaptTo: 59, Weeks: 4}, res.Loaded)
	diff(t, "page 1 more", &More{Before: 56, Remaining: 8, Label: "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 12"}, res.More)
	diff(t, "page 1 totals", []string{"egg.total=2300"}, cellsOf(res.Totals.Cells, KeyEggs))
	if string(raw["start"]) != "null" {
		t.Errorf("page 1 start = %s, want null", raw["start"])
	}

	// Load more (before=56): only SAPT 52–55, the totals over 52–59
	// (29.09–18.11, 51 days), loaded 52–59.
	res, raw = table("&before=56")
	diff(t, "load-more weeks", []int{55, 54, 53, 52}, sapts(res))
	w52 := res.Weeks[3]
	if w52.From != day(time.September, 29) || w52.To != day(time.October, 5) || len(w52.Days) != 7 {
		t.Errorf("SAPT 52 = %s–%s with %d days", w52.From, w52.To, len(w52.Days))
	}
	diff(t, "load-more loaded", &Loaded{SaptFrom: 52, SaptTo: 59, Weeks: 8}, res.Loaded)
	if res.Totals == nil || res.Totals.Label != "Total SAPT 52–59" {
		t.Fatalf("load-more totals = %+v", res.Totals)
	}
	diff(t, "load-more totals", []string{"egg.total=5100", "alive.F=100", "alive.M=10", "deaths.F=0"},
		cellsOf(res.Totals.Cells, KeyEggs, KeyAliveF, KeyAliveM, KeyDeathsF))
	diff(t, "load-more more", &More{Before: 52, Remaining: 4, Label: "Încarcă încă 4 săptămâni · SAPT 52–59 · 8 din 12"}, res.More)
	if string(raw["start"]) != "null" {
		t.Errorf("load-more start = %s, want null", raw["start"])
	}

	// The last page: SAPT 48–51 and the start note, no more.
	res, raw = table("&before=52")
	diff(t, "last page weeks", []int{51, 50, 49, 48}, sapts(res))
	if string(raw["more"]) != "null" || res.Start == nil || res.Start.Label != "Începutul ciclului · SAPT 48" {
		t.Errorf("last page more %s, start %+v", raw["more"], res.Start)
	}

	// A set-length period: "4 săpt." (SAPT 56–59), then load-more widens it
	// to weeks:8 with before=56 (D190–D192): only SAPT 52–55, the totals over
	// every loaded week 52–59, as on the whole cycle.
	res, _ = table("&period=weeks:4")
	diff(t, "weeks:4 weeks", []int{59, 58, 57, 56}, sapts(res))
	diff(t, "weeks:4 period", []int{56, 59, 4}, []int{res.Period.SaptFrom, res.Period.SaptTo, res.Period.WeeksTotal})
	if res.More == nil || res.More.Before != 56 {
		t.Fatalf("weeks:4 more = %+v, want before 56", res.More)
	}
	res, _ = table("&period=weeks:8&before=56")
	diff(t, "weeks:8 load-more weeks", []int{55, 54, 53, 52}, sapts(res))
	diff(t, "weeks:8 period", []int{52, 59, 8}, []int{res.Period.SaptFrom, res.Period.SaptTo, res.Period.WeeksTotal})
	diff(t, "weeks:8 load-more loaded", &Loaded{SaptFrom: 52, SaptTo: 59, Weeks: 8}, res.Loaded)
	if res.Totals == nil || res.Totals.Label != "Total SAPT 52–59" {
		t.Fatalf("weeks:8 load-more totals = %+v", res.Totals)
	}
	diff(t, "weeks:8 load-more totals", []string{"egg.total=5100", "alive.F=100", "alive.M=10", "deaths.F=0"},
		cellsOf(res.Totals.Cells, KeyEggs, KeyAliveF, KeyAliveM, KeyDeathsF))
}
