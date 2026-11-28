package series_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/series"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded (seed imports farms,
// which the external package may import with it).
func TestMain(m *testing.M) { seed.Main(m) }

// seeded is the seeded database with the ids the tests name.
type seeded struct {
	d                *testdb.DB
	s26, s25         uuid.UUID
	telita, hirbovat uuid.UUID
}

func newSeeded(t *testing.T) seeded {
	t.Helper()
	s := seeded{d: testdb.NewSeeded(t)}
	for _, q := range []struct {
		sql string
		id  *uuid.UUID
	}{
		{`select id from series where name = 'Seria 2026'`, &s.s26},
		{`select id from series where name = 'Seria 2025'`, &s.s25},
		{`select id from farm where name_key = 'telita'`, &s.telita},
		{`select id from farm where name_key = 'hirbovat'`, &s.hirbovat},
	} {
		if err := s.d.App.QueryRow(t.Context(), q.sql).Scan(q.id); err != nil {
			t.Fatalf("%s: %v", q.sql, err)
		}
	}
	return s
}

func (s seeded) page(t *testing.T, p auth.Principal, id uuid.UUID) series.PageResponse {
	t.Helper()
	res, err := series.BuildPage(t.Context(), s.d.App, series.Input{Principal: p, SeriesID: id, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

// summaryStats is the stats of the farm's Rezumat summary for cycle (nil:
// the default one) over the whole cycle, as JSON.
func (s seeded) summaryStats(t *testing.T, farm uuid.UUID, cycle *uuid.UUID) string {
	t.Helper()
	res, err := farms.BuildSummary(t.Context(), s.d.App, farms.Input{FarmID: farm, CycleID: cycle, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		t.Fatal(err)
	}
	return jsonOf(t, res.Stats)
}

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// dayOpt compares timex.Day values with ==.
var dayOpt = cmpopts.EquateComparable(timex.Day{})

func owner() auth.Principal { return auth.Principal{UserID: uuid.New(), Role: auth.Owner} }

func viewerOf(farms ...uuid.UUID) auth.Principal {
	return auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: farms}
}

func day(s string) timex.Day {
	d, err := timex.ParseDay(s)
	if err != nil {
		panic(err)
	}
	return d
}

// statOf is the stat with id in stats, nil when absent.
func statOf(stats []farms.Stat, id string) farms.Stat {
	for _, st := range stats {
		switch st := st.(type) {
		case farms.AliveStat:
			if st.ID == id {
				return st
			}
		case farms.CumMortalityStat:
			if st.ID == id {
				return st
			}
		case farms.FeedDaysLeftStat:
			if st.ID == id {
				return st
			}
		case farms.LayTodayStat:
			if st.ID == id {
				return st
			}
		}
	}
	return nil
}

// TestBuildPageSeria2026Owner: the S3 board's header, track and farm panels
// from the real data (M4 T6 criterion 2).
func TestBuildPageSeria2026Owner(t *testing.T) {
	s := newSeeded(t)
	res := s.page(t, owner(), s.s26)

	if res.Series.Name != "Seria 2026" || res.Series.Status != "active" || res.Series.PlacementDate != day("2025-10-07") ||
		res.Series.PlacementInferred || res.Series.AgeAtPlacementDays != 0 {
		t.Errorf("series = %+v", res.Series)
	}
	if res.Alive == nil || *res.Alive != (httpx.SexPair[int]{F: 12561, M: 794}) {
		t.Errorf("alive = %v, want ♀ 12 561 · ♂ 794", res.Alive)
	}
	if res.SaptToday == nil || *res.SaptToday != 59 || res.SaptSpan != [2]int{1, 59} || res.OwnFarmsOnly {
		t.Errorf("saptToday %v, saptSpan %v, ownFarmsOnly %v; want 59, [1 59], false", res.SaptToday, res.SaptSpan, res.OwnFarmsOnly)
	}
	wantPlacement := &series.Placement{Date: day("2025-10-07"), Farm: "Telița", Placed: &httpx.SexPair[int]{F: 28080, M: 3510}}
	if diff := cmp.Diff(wantPlacement, res.Placement, dayOpt); diff != "" {
		t.Errorf("placement (-want +got):\n%s", diff)
	}
	if res.End != nil {
		t.Errorf("end = %+v on an active series", res.End)
	}
	if want := []string{"pasari", "hrana", "greutate", "productie"}; !cmp.Equal(res.Tabs, want) {
		t.Errorf("tabs = %v, want %v", res.Tabs, want)
	}

	// Track: Telița (growth) then Hîrbovăț (laying).
	type laneCycle struct {
		Farm, Type string
		Span       [2]int
		Open       bool
	}
	var cycles []laneCycle
	var names []string
	for _, l := range res.Track.Lanes {
		names = append(names, l.Farm.Name)
		for _, c := range l.Cycles {
			cycles = append(cycles, laneCycle{l.Farm.Name, string(c.Type), c.SaptSpan, c.Open})
		}
	}
	if want := []string{"Telița", "Hîrbovăț"}; !cmp.Equal(names, want) {
		t.Fatalf("lanes = %v, want %v", names, want)
	}
	wantCycles := []laneCycle{{"Telița", "growth", [2]int{1, 21}, false}, {"Hîrbovăț", "laying", [2]int{20, 59}, true}}
	if diff := cmp.Diff(wantCycles, cycles); diff != "" {
		t.Errorf("lane cycles (-want +got):\n%s", diff)
	}
	if res.Track.Today != testdb.Today {
		t.Errorf("track today = %s", res.Track.Today)
	}
	// The links Telița → Hîrbovăț 22.02–28.02, on Telița's lane only.
	var f, m int
	for _, tr := range res.Track.Lanes[0].Transfers {
		if tr.FromFarm != "Telița" || tr.ToFarm != "Hîrbovăț" || tr.Date.Before(day("2026-02-22")) || tr.Date.After(day("2026-02-28")) {
			t.Errorf("transfer %+v outside Telița → Hîrbovăț 22.02–28.02", tr)
		}
		f, m = f+tr.F, m+tr.M
	}
	if f != 26883 || m != 3024 {
		t.Errorf("transferred ♀ %d · ♂ %d, want 26 883 · 3 024", f, m)
	}
	if n := len(res.Track.Lanes[1].Transfers); n != 0 {
		t.Errorf("Hîrbovăț lane has %d links, want 0 (listed on the source lane)", n)
	}

	// Panels.
	if len(res.Farms) != 2 {
		t.Fatalf("farms = %d, want 2", len(res.Farms))
	}
	telita, ok1 := res.Farms[0].(series.FarmPanel)
	hirbovat, ok2 := res.Farms[1].(series.FarmPanel)
	if !ok1 || !ok2 || telita.Farm.Name != "Telița" || hirbovat.Farm.Name != "Hîrbovăț" || !telita.Access || !hirbovat.Access {
		t.Fatalf("farms = %+v", res.Farms)
	}
	if c := telita.Cycle; c == nil || c.Open || c.StartDate != day("2025-10-07") || *c.EndDate != day("2026-02-28") || c.SaptSpan != [2]int{1, 21} {
		t.Errorf("Telița cycle = %+v", c)
	}
	if telita.Farm.Type != auth.FarmGrowth || telita.Guide == nil || telita.Guide.Name != "Ross 308 creștere" {
		t.Errorf("Telița farm %+v, guide %+v", telita.Farm, telita.Guide)
	}
	alive, _ := statOf(telita.Stats, farms.StatAlive).(farms.AliveStat)
	if a := alive.Arrivals; a.F == nil || a.M == nil || *a.F != 28080 || *a.M != 3510 {
		t.Errorf("Telița arrivals = %+v, want 28 080 · 3 510", alive.Arrivals)
	}
	if statOf(telita.Stats, farms.StatCumMortality) == nil || statOf(telita.Stats, farms.StatFeedDaysLeft) != nil {
		t.Errorf("Telița stats = %s, want cumMortality and no feedDaysLeft", jsonOf(t, telita.Stats))
	}
	telitaCycle := telita.Cycle.ID
	if got, want := jsonOf(t, telita.Stats), s.summaryStats(t, s.telita, &telitaCycle); got != want {
		t.Errorf("Telița stats = %s\nwant the summary's %s", got, want)
	}

	if c := hirbovat.Cycle; c == nil || !c.Open || c.StartDate != day("2026-02-22") || c.EndDate != nil || c.SaptSpan != [2]int{20, 59} {
		t.Errorf("Hîrbovăț cycle = %+v", c)
	}
	// Equal to the farm summary of the open cycle: alive, cumMortality
	// 5.66 / 34.66, feedToday, feedDaysLeft 13.2 / 10.4, layToday 41.08.
	if got, want := jsonOf(t, hirbovat.Stats), s.summaryStats(t, s.hirbovat, nil); got != want {
		t.Errorf("Hîrbovăț stats = %s\nwant the summary's %s", got, want)
	}
	cm, _ := statOf(hirbovat.Stats, farms.StatCumMortality).(farms.CumMortalityStat)
	lay, _ := statOf(hirbovat.Stats, farms.StatLayToday).(farms.LayTodayStat)
	left, _ := statOf(hirbovat.Stats, farms.StatFeedDaysLeft).(farms.FeedDaysLeftStat)
	if cm.F == nil || cm.M == nil || cm.F.Decimal().StringFixed(2) != "5.66" || cm.M.Decimal().StringFixed(2) != "34.66" {
		t.Errorf("Hîrbovăț cumMortality = %s", jsonOf(t, cm))
	}
	if lay.Value == nil || lay.Value.Decimal().StringFixed(2) != "41.08" {
		t.Errorf("Hîrbovăț layToday = %s", jsonOf(t, lay))
	}
	if len(left.Items) != 2 || left.Items[0].Days.Decimal().StringFixed(1) != "13.2" || left.Items[1].Days.Decimal().StringFixed(1) != "10.4" {
		t.Errorf("Hîrbovăț feedDaysLeft = %s", jsonOf(t, left))
	}
}

// TestBuildPageViewerHirbovat: a viewer of Hîrbovăț only (the "fără acces"
// board): one lane, no links, no placement, Telița named without access.
func TestBuildPageViewerHirbovat(t *testing.T) {
	s := newSeeded(t)
	res := s.page(t, viewerOf(s.hirbovat), s.s26)

	if !res.OwnFarmsOnly || res.Placement != nil || res.Alive == nil || *res.Alive != (httpx.SexPair[int]{F: 12561, M: 794}) {
		t.Errorf("ownFarmsOnly %v, placement %+v, alive %v", res.OwnFarmsOnly, res.Placement, res.Alive)
	}
	if res.SaptSpan != [2]int{20, 59} {
		t.Errorf("saptSpan = %v, want [20 59] (the viewer's cycles)", res.SaptSpan)
	}
	if len(res.Track.Lanes) != 1 || res.Track.Lanes[0].Farm.ID != s.hirbovat || len(res.Track.Lanes[0].Transfers) != 0 {
		t.Errorf("lanes = %+v, want Hîrbovăț only without links", res.Track.Lanes)
	}
	if len(res.Farms) != 2 {
		t.Fatalf("farms = %+v", res.Farms)
	}
	if got, want := jsonOf(t, res.Farms[0]), `{"farm":{"name":"Telița"},"access":false}`; got != want {
		t.Errorf("Telița panel = %s, want %s", got, want)
	}
	if p, ok := res.Farms[1].(series.FarmPanel); !ok || p.Farm.ID != s.hirbovat || len(p.Stats) == 0 {
		t.Errorf("Hîrbovăț panel = %+v", res.Farms[1])
	}

	// A viewer of Telița only: Seria 2025 has no cycle there.
	if _, err := series.BuildPage(t.Context(), s.d.App, series.Input{
		Principal: viewerOf(s.telita), SeriesID: s.s25, Today: testdb.Today, Now: testdb.Now(),
	}); err != series.ErrSeriesNotFound {
		t.Errorf("Seria 2025 for a viewer of Telița: err = %v, want ErrSeriesNotFound", err)
	}
}

// TestBuildPageSeria2025Owner: a closed series (the "seria închisă" board).
func TestBuildPageSeria2025Owner(t *testing.T) {
	s := newSeeded(t)
	res := s.page(t, owner(), s.s25)

	if res.Series.Status != "closed" || res.Alive != nil || res.SaptToday != nil || res.SaptSpan != [2]int{19, 65} {
		t.Errorf("status %s, alive %v, saptToday %v, saptSpan %v", res.Series.Status, res.Alive, res.SaptToday, res.SaptSpan)
	}
	wantPlacement := &series.Placement{Date: day("2025-03-07"), Inferred: true, Farm: "Hîrbovăț"}
	if diff := cmp.Diff(wantPlacement, res.Placement, dayOpt); diff != "" {
		t.Errorf("placement (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(&series.End{Date: day("2026-01-28"), By: series.EndEmptied}, res.End, dayOpt); diff != "" {
		t.Errorf("end (-want +got):\n%s", diff)
	}
	if len(res.Track.Lanes) != 2 || res.Track.Lanes[0].Farm.Name != "Telița" || len(res.Track.Lanes[0].Cycles) != 0 {
		t.Errorf("lanes = %+v, want an empty Telița lane first", res.Track.Lanes)
	}
	if got, want := jsonOf(t, res.Farms[0]), `"cycle":null,"stats":[]}`; len(got) < len(want) || got[len(got)-len(want):] != want {
		t.Errorf("Telița panel = %s, want no cycle", got)
	}
	h, ok := res.Farms[1].(series.FarmPanel)
	if !ok || h.Cycle == nil || !h.Cycle.StartInferred || h.Cycle.SaptSpan != [2]int{19, 65} {
		t.Fatalf("Hîrbovăț panel = %+v", res.Farms[1])
	}
	cycle := h.Cycle.ID
	if got, want := jsonOf(t, h.Stats), s.summaryStats(t, s.hirbovat, &cycle); got != want {
		t.Errorf("Hîrbovăț stats = %s\nwant the summary's %s", got, want)
	}
}

// TestPageRouteSeeded: the route's access rules (M4 Q4) and that it serves
// the builder's page in one read snapshot.
func TestPageRouteSeeded(t *testing.T) {
	s := newSeeded(t)
	server := func(p *auth.Principal) http.Handler {
		mux := http.NewServeMux()
		withCaller := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if p != nil {
					r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
				}
				next.ServeHTTP(w, r)
			})
		}
		series.NewHandler(series.NewService(db.NewStore(s.d.App), s.d.Clock)).Routes(mux, httpx.NewChain(withCaller))
		return mux
	}
	own := owner()
	vet := auth.Principal{UserID: uuid.New(), Role: auth.Vet, FarmIDs: []uuid.UUID{s.hirbovat}}
	manager := auth.Principal{UserID: uuid.New(), Role: auth.Manager, FarmIDs: []uuid.UUID{s.hirbovat}}
	telitaViewer := viewerOf(s.telita)
	path := func(id string) string { return "/api/series/" + id + "/page" }

	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		path      string
		status    int
		code      string
		issuePath string
	}{
		{"no session", nil, path(s.s26.String()), 401, "UNAUTHENTICATED", ""},
		{"vet", &vet, path(s.s26.String()), 403, "FORBIDDEN", ""},
		{"owner, unknown series", &own, path(uuid.NewString()), 404, "NOT_FOUND", ""},
		{"viewer, no cycle on his farms", &telitaViewer, path(s.s25.String()), 404, "NOT_FOUND", ""},
		{"malformed id", &own, path("seria-2026"), 400, "VALIDATION", "id"},
		{"unknown parameter", &own, path(s.s26.String()) + "?farmId=" + s.hirbovat.String(), 400, "VALIDATION", "farmId"},
		{"owner", &own, path(s.s26.String()), 200, "", ""},
		{"manager", &manager, path(s.s26.String()), 200, "", ""},
		{"viewer of Telița, Seria 2026", &telitaViewer, path(s.s26.String()), 200, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			server(tc.caller).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			if tc.status == 200 {
				var res struct {
					Series struct{ Name string } `json:"series"`
					Track  struct {
						Lanes []json.RawMessage `json:"lanes"`
					} `json:"track"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil || res.Series.Name != "Seria 2026" || len(res.Track.Lanes) == 0 {
					t.Errorf("body = %s (%v)", rec.Body, err)
				}
				return
			}
			var body struct {
				Error struct {
					Code   string        `json:"code"`
					Issues []httpx.Issue `json:"issues"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Errorf("code = %s, want %s", body.Error.Code, tc.code)
			}
			if tc.issuePath != "" && (len(body.Error.Issues) != 1 || body.Error.Issues[0].Path != tc.issuePath) {
				t.Errorf("issues = %+v, want one on %s", body.Error.Issues, tc.issuePath)
			}
		})
	}
}
