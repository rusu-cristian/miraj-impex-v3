package periodtable

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// newGrowthFarm is a hand-built growth farm. Seria G is placed 27.10.2026 at
// age 0 (SAPT 1 = 27.10–02.11); its cycle closed 16.11 (SAPT 1–3, "Tot ciclul").
//
//	Hall H1 with H1.1 and H1.2; 27.10 placement ♀ 100 ♂ 20 in each.
//	Mortality rows every day, 0 except 30.10 H1.1 ♀ 4.
//	Feed every day: ♀ 50 g, ♂ 70 g (27.10 has no birds at the start: 0 kg,
//	not a fed day).
//	Weighings 10.11 (SAPT 3): H1.1 ♀ 520 g (96 birds), H1.2 ♀ 480 g (100 birds)
//	→ (520·96 + 480·100) ÷ 196 = 499,59 g.
//	Guide: feed ♀ 40, 45, 50, 55 g · ♂ 55, 58, 60, 62 g; ♀ weight at the end
//	of SAPT 1–4: 150, 480, 700, 900 g.
func newGrowthFarm(t *testing.T) *farmT {
	t.Helper()
	f := &farmT{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria G', 'seria g', 2031, $1)`, day(time.October, 27))
	f.guide = f.id(`insert into guide (name, name_key, type) values ('Ghid G', 'ghid g', 'growth')`)
	f.exec(`insert into guide_week (guide_id, week, sex, feed_g, body_weight_g)
		select $1::uuid, w, 'F'::sex, (array[40, 45, 50, 55])[w], (array[150, 480, 700, 900])[w] from generate_series(1, 4) w
		union all select $1::uuid, w, 'M'::sex, (array[55, 58, 60, 62])[w], null from generate_series(1, 4) w`, f.guide)
	f.farm = f.id(`insert into farm (name, name_key, type, guide_id) values ('Ferma G', 'ferma g', 'growth', $1)`, f.guide)
	f.cycle = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, $3, $4, $5, 'auto', 'emptied')`, f.series, f.farm, f.guide, day(time.October, 27), day(time.November, 16))
	h1 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'H1', 1)`, f.farm)
	f.sec["H1.1"] = f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, 'H1.1', $2, 1)`, f.farm, h1)
	f.sec["H1.2"] = f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, 'H1.2', $2, 2)`, f.farm, h1)
	leaves := []uuid.UUID{f.sec["H1.1"], f.sec["H1.2"]}

	e := f.entry("placement")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, s, '2026-10-27', x.sex::sex, 'placement', x.qty from unnest($3::uuid[]) s, (values ('F', 100), ('M', 20)) x(sex, qty)`,
		e, f.farm, leaves)
	e = f.entry("daily")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, s, d::date, x.sex::sex, 'mortality',
		       case when s = $4 and x.sex = 'F' and d::date = '2026-10-30' then -4 else 0 end
		from unnest($3::uuid[]) s, generate_series('2026-10-27'::date, '2026-11-16'::date, interval '1 day') d, (values ('F'), ('M')) x(sex)`,
		e, f.farm, leaves, f.sec["H1.1"])
	ft := f.id(`insert into feed_type (name, name_key) values ('Starter G', 'starter g')`)
	e = f.entry("feed_daily")
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		select $1, $2, s, d::date, x.sex::sex, $4, x.g
		from unnest($3::uuid[]) s, generate_series('2026-10-27'::date, '2026-11-16'::date, interval '1 day') d, (values ('F', 50), ('M', 70)) x(sex, g)`,
		e, f.farm, leaves, ft)
	e = f.entry("weighing")
	f.exec(`insert into weighing (entry_id, farm_id, section_id, date, sex, avg_weight_g)
		values ($1, $2, $3, '2026-11-10', 'F', 520), ($1, $2, $4, '2026-11-10', 'F', 480)`, e, f.farm, f.sec["H1.1"], f.sec["H1.2"])
	return f
}

func groupIDs(gs []Group) []string {
	var ids []string
	for _, g := range gs {
		ids = append(ids, g.ID)
	}
	return ids
}

func leafIDs(g Group) []string {
	var ids []string
	for _, l := range g.Leaves {
		ids = append(ids, l.ID+" "+l.Label)
	}
	return ids
}

func TestBuildGrowth(t *testing.T) {
	f := newGrowthFarm(t)
	resp := f.build(TableQuery{CycleID: &f.cycle})

	diff(t, "period", PeriodInfo{Kind: "cycle", Label: "Tot ciclul", SaptFrom: 1, SaptTo: 3,
		From: day(time.October, 27), To: day(time.November, 16), WeeksTotal: 3}, resp.Period)
	diff(t, "chips", []string{"mortalitate", "populare", "hrana", "cantarire"}, resp.Chips)
	diff(t, "groups", []string{"vii", "morti", "gPasare", "greutateF", "hranaKg"}, groupIDs(resp.Groups))
	diff(t, "g/pasăre", []string{"feed.g.F ♀ efectiv", "feed.g.guide.F ♀ ghid", "feed.g.M ♂ efectiv", "feed.g.guide.M ♂ ghid"}, leafIDs(resp.Groups[2]))
	diff(t, "greutate", []string{"weight.F medie g", "weight.guide.F ghid", "weight.diffPct.F față de ghid"}, leafIDs(resp.Groups[3]))
	if resp.Groups[2].Label != "g/pasăre" || resp.Groups[3].Label != "Greutate ♀" || resp.NoGuide {
		t.Errorf("labels %q, %q; noGuide %v", resp.Groups[2].Label, resp.Groups[3].Label, resp.NoGuide)
	}
	if resp.More != nil || resp.Start == nil || resp.Start.Label != "Începutul ciclului · SAPT 1" {
		t.Errorf("more %+v, start %+v", resp.More, resp.Start)
	}

	// SAPT 3 (10–16.11): g = 1000 × kg ÷ bird-days; the weighing week.
	w3 := resp.Weeks[0]
	diff(t, "SAPT 3", []string{"alive.F=196", "alive.M=40", "deaths.F=0", "feed.g.F=50", "feed.g.guide.F=50",
		"feed.g.M=70", "feed.g.guide.M=60", "weight.F=499.5918", "weight.guide.F=480", "weight.diffPct.F=4.0816",
		"feed.kg.F=68.6", "feed.kg.M=19.6", "sold.F=absent", "egg.total=absent"},
		cellsOf(w3.Cells, KeyAliveF, KeyAliveM, KeyDeathsF, KeyFeedGF, KeyFeedGGuideF, KeyFeedGM, KeyFeedGGuideM,
			KeyWeightF, KeyWeightGuideF, KeyWeightDiffF, KeyFeedKgF, KeyFeedKgM, KeySoldF, KeyEggs))
	diff(t, "SAPT 3 status", map[string]string{"feed.g.F": "in_norma", "feed.g.M": "abatere", "weight.F": "in_norma"}, w3.Status)
	d10 := w3.Days[0]
	diff(t, "10.11", []string{"weight.F=499.5918", "weight.guide.F=480", "feed.g.F=50", "feed.g.guide.F=50"},
		cellsOf(d10.Cells, KeyWeightF, KeyWeightGuideF, KeyFeedGF, KeyFeedGGuideF))
	if d11 := w3.Days[1]; cell(d11.Cells, KeyWeightF) != "null" || d11.Reasons != nil {
		t.Errorf("11.11 (no weighing) = %s, reasons %v", cell(d11.Cells, KeyWeightF), d11.Reasons)
	}

	// SAPT 1: 27.10 is no fed day (no birds at its start); no weighing.
	w1 := resp.Weeks[2]
	diff(t, "SAPT 1", []string{"deaths.F=4", "feed.g.F=50", "feed.g.guide.F=40", "weight.F=null", "weight.guide.F=null"},
		cellsOf(w1.Cells, KeyDeathsF, KeyFeedGF, KeyFeedGGuideF, KeyWeightF, KeyWeightGuideF))
	diff(t, "SAPT 1 reasons", map[string]string{"weight.F": "no_weighing"}, w1.Reasons)
	diff(t, "SAPT 1 status", map[string]string{"feed.g.F": "abatere", "feed.g.M": "abatere"}, w1.Status)
	if c := w1.Days[0].Cells; cell(c, KeyFeedGF) != "null" || cell(c, KeyFeedKgF) != "0" {
		t.Errorf("27.10 = %v", cellsOf(c, KeyFeedGF, KeyFeedKgF))
	}

	// Totals: no weight (§5.14), g guide weighted by fed bird-days:
	// (40·1 188 + 45·1 372 + 50·1 372) ÷ 3 932.
	diff(t, "totals", []string{"feed.g.F=50", "feed.g.guide.F=45.234", "weight.F=null", "weight.guide.F=null", "deaths.F=4", "alive.F=196"},
		cellsOf(resp.Totals.Cells, KeyFeedGF, KeyFeedGGuideF, KeyWeightF, KeyWeightGuideF, KeyDeathsF, KeyAliveF))
	if resp.Totals.Reasons != nil {
		t.Errorf("totals reasons = %v", resp.Totals.Reasons)
	}

	// Without a guide: the guide leaves go, noGuide is set, no status.
	f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
	resp = f.build(TableQuery{CycleID: &f.cycle})
	if !resp.NoGuide || resp.Weeks[0].Status != nil {
		t.Errorf("noGuide %v, status %v", resp.NoGuide, resp.Weeks[0].Status)
	}
	diff(t, "g/pasăre without a guide", []string{"feed.g.F ♀ efectiv", "feed.g.M ♂ efectiv"}, leafIDs(resp.Groups[2]))
	diff(t, "greutate without a guide", []string{"weight.F medie g"}, leafIDs(resp.Groups[3]))
}

// newLegacyFarm is a hand-built past cycle without a flock count, like Seria
// 2025 at Hîrbovăț: Seria L placed 01.11.2024 at age 0, cycle 07.03.2025
// (SAPT 19) – 28.01.2026 (SAPT 65), no bird event; the register starts
// 04.04.2025 (SAPT 23).
//
//	04.04.2025 eggs H1: Cat. 2 263, jos 100.
//	10.01.2026 legacy H1.1 ♀ 102 · ♂ 1.
//	16.01.2026 legacy H1.1 ♀ 300 · ♂ 5 emptying; H4.1 ♀ 2 · ♂ 0; eggs H4 100 incubabile, jos 3.
//	23.01.2026 legacy H4.1 ♀ 400 · ♂ 10 emptying; eggs H4 50 incubabile.
func newLegacyFarm(t *testing.T) *farmT {
	t.Helper()
	f := &farmT{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series = f.id(`insert into series (name, name_key, year, placement_date, placement_inferred, status, closed_at)
		values ('Seria L', 'seria l', 2032, '2024-11-01', true, 'closed', $1)`, testdb.Now())
	f.farm = f.id(`insert into farm (name, name_key, type) values ('Ferma L', 'ferma l', 'laying')`)
	f.cycle = f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by, start_inferred)
		values ($1, $2, '2025-03-07', '2026-01-28', 'migration', 'migration', true)`, f.series, f.farm)
	h1 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'H1', 1)`, f.farm)
	h4 := f.id(`insert into section (farm_id, code, sort_order) values ($1, 'H4', 2)`, f.farm)
	f.sec["H1"], f.sec["H4"] = h1, h4
	f.sec["H1.1"] = f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, 'H1.1', $2, 1)`, f.farm, h1)
	f.sec["H4.1"] = f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, 'H4.1', $2, 1)`, f.farm, h4)
	e := f.id(`insert into entry (kind, farm_id, date, source) values ('migration', $1, '2025-04-04', 'migration')`, f.farm)
	eggs := `insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs) values ($1, $2, $3, $4, $5, 0, $6, 0, $7)`
	f.exec(eggs, e, f.farm, h1, timex.MustDay(2025, 4, 4), 100, 263, 0)
	f.exec(eggs, e, f.farm, h4, timex.MustDay(2026, 1, 16), 3, 0, 100)
	f.exec(eggs, e, f.farm, h4, timex.MustDay(2026, 1, 23), 0, 0, 50)
	f.exec(`insert into legacy_mortality (entry_id, farm_id, section_id, date, sex, deaths, emptying) values
		($1, $2, $3, '2026-01-10', 'F', 102, false), ($1, $2, $3, '2026-01-10', 'M', 1, false),
		($1, $2, $3, '2026-01-16', 'F', 300, true), ($1, $2, $3, '2026-01-16', 'M', 5, true),
		($1, $2, $4, '2026-01-16', 'F', 2, false), ($1, $2, $4, '2026-01-16', 'M', 0, false),
		($1, $2, $4, '2026-01-23', 'F', 400, true), ($1, $2, $4, '2026-01-23', 'M', 10, true)`,
		e, f.farm, f.sec["H1.1"], f.sec["H4.1"])
	return f
}

func TestBuildLegacy(t *testing.T) {
	f := newLegacyFarm(t)
	resp := f.build(TableQuery{CycleID: &f.cycle})

	// The register's weeks: SAPT 23–65, "din 43" (M2 Q7).
	diff(t, "period", PeriodInfo{Kind: "cycle", Label: "Tot ciclul", SaptFrom: 23, SaptTo: 65,
		From: timex.MustDay(2025, 4, 4), To: timex.MustDay(2026, 1, 28), WeeksTotal: 43}, resp.Period)
	diff(t, "more", &More{Before: 62, Remaining: 39, Label: "Încarcă încă 4 săptămâni · SAPT 62–65 · 4 din 43"}, resp.More)
	diff(t, "chips", []string{"mortalitate", "oua"}, resp.Chips)
	diff(t, "groups", []string{"morti", "ouat"}, groupIDs(resp.Groups))
	diff(t, "ouat", []string{"egg.total ouă", "egg.inc incub.", "egg.incPct % incub.", "egg.jos jos", "egg.josPct % jos"}, leafIDs(resp.Groups[1]))
	if !resp.NoGuide || len(resp.Notes) != 0 {
		t.Errorf("noGuide %v, notes %v", resp.NoGuide, resp.Notes)
	}

	byWeek := map[int]WeekRow{}
	for _, w := range resp.Weeks {
		byWeek[w.Sapt] = w
	}
	// SAPT 65 (23–28.01, 6 din 7): only emptying rows.
	w65 := byWeek[65]
	if w65.Label != "SAPT 65 · 23.01–28.01 · 6 din 7 zile" {
		t.Errorf("SAPT 65 label %q", w65.Label)
	}
	diff(t, "SAPT 65", []string{"deaths.F=null", "deaths.M=null", "egg.total=50", "egg.incPct=100", "alive.F=absent", "egg.layPct=absent"},
		cellsOf(w65.Cells, KeyDeathsF, KeyDeathsM, KeyEggs, KeyIncPct, KeyAliveF, KeyLayPct))
	diff(t, "SAPT 65 reasons", map[string]string{"deaths.F": "emptying", "deaths.M": "emptying"}, w65.Reasons)
	if w65.Status != nil {
		t.Errorf("SAPT 65 status %v", w65.Status)
	}
	// SAPT 64 (16–22.01): H1 emptying, H4 still counting.
	w64 := byWeek[64]
	diff(t, "SAPT 64", []string{"deaths.F=2", "deaths.M=0", "egg.total=100", "egg.jos=3"},
		cellsOf(w64.Cells, KeyDeathsF, KeyDeathsM, KeyEggs, KeyJos))
	if w64.Reasons != nil {
		t.Errorf("SAPT 64 reasons %v", w64.Reasons)
	}
	if d := w64.Days[0]; d.Date != timex.MustDay(2026, 1, 16) || cell(d.Cells, KeyDeathsF) != "2" {
		t.Errorf("16.01 = %s %v", d.Date, cellsOf(d.Cells, KeyDeathsF))
	}
	if d := w64.Days[1]; cell(d.Cells, KeyDeathsF) != "null" || d.Reasons != nil {
		t.Errorf("17.01 (no register row) = %v %v", cellsOf(d.Cells, KeyDeathsF), d.Reasons)
	}
	diff(t, "SAPT 63", []string{"deaths.F=102", "deaths.M=1", "egg.total=null"}, cellsOf(byWeek[63].Cells, KeyDeathsF, KeyDeathsM, KeyEggs))
	diff(t, "SAPT 62", []string{"deaths.F=null", "egg.total=null"}, cellsOf(byWeek[62].Cells, KeyDeathsF, KeyEggs))

	// The first page of the register: the start note names SAPT 23.
	resp = f.build(TableQuery{CycleID: &f.cycle, Before: intp(27)})
	if resp.Start == nil || resp.Start.Label != "Începutul ciclului · SAPT 23" || resp.More != nil {
		t.Errorf("start %+v, more %+v", resp.Start, resp.More)
	}
	diff(t, "SAPT 23", []string{"egg.total=263", "egg.jos=100", "deaths.F=null"}, cellsOf(resp.Weeks[len(resp.Weeks)-1].Cells, KeyEggs, KeyJos, KeyDeathsF))

	// A guide on the cycle doesn't bring guide leaves: there are no hen-days.
	g := f.id(`insert into guide (name, name_key, type) values ('Ghid L', 'ghid l', 'laying')`)
	f.exec(`update cycle set guide_id = $1 where id = $2`, g, f.cycle)
	resp = f.build(TableQuery{CycleID: &f.cycle})
	if resp.NoGuide || slices.ContainsFunc(resp.Groups[1].Leaves, func(l Leaf) bool { return l.Kind == KindGuide }) {
		t.Errorf("legacy with a guide: noGuide %v, ouat %v", resp.NoGuide, leafIDs(resp.Groups[1]))
	}

	// Statements: the cycle, its first register day, the day query.
	pool, st := tracedPool(t, f)
	if _, err := Build(context.Background(), pool, Input{Query: TableQuery{View: "farm", FarmID: &f.farm, CycleID: &f.cycle}, Today: testdb.Today, Now: testdb.Now()}); err != nil {
		t.Fatal(err)
	}
	diff(t, "statements", []string{"CalcCycleByID", "PeriodTableRegisterStart", "PeriodTableDays"}, st.take())
}
