package farms

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// summaryData adds to the fixture what the Rezumat summary reads (today
// 18.11.2026 = SAPT 59 = 17.11–23.11; SAPT 58 = 10.11–16.11):
//
//	01.11 move H5.1 → H7 ♀ 3; 14.11 sale H7 ♀ 3 (H7 empty since 14.11).
//	10.11 weighing H5.1 ♀ 3 960 g · ♂ 4 870 g (guide week 57 has no weight).
//	17.11 eggs H5 38, H6 3; 18.11 eggs H5 40 (hens 50 + 35), H6 3 (hens 5).
//	18.11 mortality H5.1 ♀ 1, cull H5.1 ♀ 1; 0 rows for H5.1 ♂, H6.1 ♀ ♂; none for H5.2 ♀ (lipsă).
//	18.11 feed H5.1 ♀ 160 g · ♂ 144 g, H5.2 ♀ 160 g, H6.1 ♀ 200 g (+27 %, abatere) · ♂ 140 g.
//	01.11 deliveries Breeder 1 50 kg (never fed), Breeder 2 100 kg, Cocoși 6 kg.
//	Guide "Ross 308 ouat" SAPT 58–59: ♀ feed 157, lay 58,8 / 57,6, mortality 7,10 / 7,25; ♂ feed 156.
//
// At the end of 18.11: H5.1 ♀ 48 · ♂ 10, H5.2 ♀ 35, H6.1 ♀ 5 · ♂ 10, farm ♀ 88 · ♂ 20.
// Stock (every row with birds is fed today, so the window is 16–18.11):
// Breeder 1 50 kg, no consumption; Breeder 2 100 − 14,6 = 85,4 kg, avg
// 4,8667 kg/zi, 17,55 days; Cocoși 6 − 2,84 = 3,16 kg, avg 0,9467, 3,34 days.
func (f *fixture) summaryData() {
	f.t.Helper()
	H := f.hirbovat
	d := day(2026, time.November, 1)
	e := f.entry("transfer", H, &H, d)
	f.bird(e, H, "H5.1", d, "F", "transfer", -3, "H7")
	f.bird(e, H, "H7", d, "F", "transfer", 3, "H5.1")
	d = day(2026, time.November, 14)
	e = f.entry("sale", H, nil, d)
	f.bird(e, H, "H7", d, "F", "sale", -3, "")

	d = day(2026, time.November, 10)
	e = f.entry("weighing", H, nil, d)
	w := `insert into weighing (entry_id, farm_id, section_id, date, sex, avg_weight_g) values ($1, $2, $3, $4, $5, $6)`
	f.exec(w, e, H, f.sec["H5.1"], d, "F", "3960")
	f.exec(w, e, H, f.sec["H5.1"], d, "M", "4870")

	eggs := `insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs)
		values ($1, $2, $3, $4, 1, 0, 0, 0, $5)`
	for _, x := range []struct {
		d      timex.Day
		h5, h6 int
	}{{day(2026, time.November, 17), 38, 3}, {today, 40, 3}} {
		e = f.entry("daily", H, nil, x.d)
		f.exec(eggs, e, H, f.sec["H5"], x.d, x.h5)
		f.exec(eggs, e, H, f.sec["H6"], x.d, x.h6)
	}

	e = f.entry("daily", H, nil, today)
	f.bird(e, H, "H5.1", today, "F", "mortality", -1, "")
	f.bird(e, H, "H5.1", today, "M", "mortality", 0, "")
	f.bird(e, H, "H6.1", today, "F", "mortality", 0, "")
	f.bird(e, H, "H6.1", today, "M", "mortality", 0, "")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
		values ($1, $2, $3, $4, 'F', 'cull', -1, 'boala')`, e, H, f.sec["H5.1"], today)

	breeder1 := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 1', 'breeder 1', 'F')`)
	breeder := f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')`)
	roosters := f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M')`)
	d = day(2026, time.November, 1)
	e = f.entry("delivery", H, nil, d)
	delivery := `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`
	f.exec(delivery, e, H, breeder1, d, "50")
	f.exec(delivery, e, H, breeder, d, "100")
	f.exec(delivery, e, H, roosters, d, "6")
	e = f.entry("feed_daily", H, nil, today)
	alloc := `insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, $5, $6, $7)`
	f.exec(alloc, e, H, f.sec["H5.1"], today, "F", breeder, "160")
	f.exec(alloc, e, H, f.sec["H5.1"], today, "M", roosters, "144")
	f.exec(alloc, e, H, f.sec["H5.2"], today, "F", breeder, "160")
	f.exec(alloc, e, H, f.sec["H6.1"], today, "F", breeder, "200")
	f.exec(alloc, e, H, f.sec["H6.1"], today, "M", roosters, "140")

	gw := `insert into guide_week (guide_id, week, sex, feed_g, mortality_cum_pct, lay_pct) values ($1, $2, $3, $4, $5, $6)`
	f.exec(gw, f.guideLay, 58, "F", "157", "7.10", "58.8")
	f.exec(gw, f.guideLay, 59, "F", "157", "7.25", "57.6")
	f.exec(gw, f.guideLay, 58, "M", "156", nil, nil)
	f.exec(gw, f.guideLay, 59, "M", "156", nil, nil)
}

// summaryJSON builds the summary and returns it as JSON.
func (f *fixture) summaryJSON(t *testing.T, in Input) string {
	t.Helper()
	res, err := BuildSummary(t.Context(), f.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func expectJSON(t *testing.T, got string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(got, p) {
			t.Errorf("JSON lacks %s\n%s", p, got)
		}
	}
}

func TestBuildSummaryLayingFarm(t *testing.T) {
	f := newFixture(t)
	f.summaryData()
	got := f.summaryJSON(t, f.input(principal(auth.Viewer, f.hirbovat), f.hirbovat, nil))

	// Readout: SAPT 59 = 84 eggs ÷ 180 hen-days; weeks without eggs are null, the guide still shown.
	expectJSON(t, got,
		`"readout":{"kind":"lay","label":"SAPT 59 · efectiv 46,7% · ghid 57,6% · −10,9 pct","value":46.6667,"guide":57.6,"status":"abatere","series":[{"sapt":20,"value":null,"guide":null},`,
		`{"sapt":58,"value":null,"guide":58.8},{"sapt":59,"value":46.6667,"guide":57.6}]}`,
	)
	// Stats.
	expectJSON(t, got,
		`{"id":"alive","F":88,"M":20,"arrivals":{"F":100,"M":20},"sold":{"F":8,"M":0}}`,
		`{"id":"cumMortality","F":4,"M":0,"guideF":7.25,"status":"in_norma","note":"(ghid din intrarea în ouat)","deaths":{"F":4,"M":0}}`,
		`{"id":"weight","F":3960,"M":4870,"guide":null,"reason":"no_guide","at":"2026-11-10","section":"H5.1"}`,
		// ♀ 1000 × (8 + 5,6 + 1) kg ÷ 90 = 162,22 g; ♂ 1000 × (1,44 + 1,4) ÷ 20 = 142 g.
		`{"id":"feedToday","F":162.2222,"M":142,"guideF":157,"guideM":156,"status":"in_norma","note":"H7 goală din 14.11"}`,
		// 43 eggs ÷ 90 hens.
		`{"id":"layToday","value":47.7778,"guide":57.6,"diffPts":-9.8222,"status":"abatere","eggs":43,"halls":2}`,
		// The stock as of today between feedToday and layToday (M3 D10, Q7), in type name order: Breeder 1
		// never fed; Breeder 2 85,4 ÷ (14,6 ÷ 3) = 17,55 days → 18.11 + 17; Cocoși 3,16 ÷ (2,84 ÷ 3) = 3,34
		// days, under 5 → 18.11 + 3.
		`"note":"H7 goală din 14.11"},{"id":"feedDaysLeft","items":[`+
			`{"type":"Breeder 1","forSex":"F","days":null,"runOut":null,"kg":50,"lamp":"oprit","reason":"– fără consum în ultimele 3 zile"},`+
			`{"type":"Breeder 2","forSex":"F","days":17.5479,"runOut":"2026-12-05","kg":85.4,"lamp":"in_norma"},`+
			`{"type":"Cocoși","forSex":"M","days":3.338,"runOut":"2026-11-21","kg":3.16,"lamp":"abatere","reason":"sub pragul de 5 zile"}]},`+
			`{"id":"layToday"`,
	)
	// Tiles: H5 below the guide (47,06 %), H6 above it (60 %) but H6.1 ♀ eats 27 % over; H5.2 lacks its
	// mortality row after 12:00; H7 is one empty line.
	expectJSON(t, got,
		`"tiles":[{"hall":"H5","sections":[`+
			`{"code":"H5.1","alive":{"F":48,"M":10},"deathsWeek":{"F":2,"M":0},"gToday":{"F":160,"M":144},"weight":{"F":3960,"at":"2026-11-10","guide":null,"reason":"no_guide"},"lamps":["abatere"],"layHall":47.0588},`+
			`{"code":"H5.2","alive":{"F":35,"M":0},"deathsWeek":{"F":0,"M":0},"gToday":{"F":160,"M":null},"weight":null,"lamps":["lipsa","abatere"],"layHall":47.0588}]},`+
			`{"hall":"H6","sections":[{"code":"H6.1","alive":{"F":5,"M":10},"deathsWeek":{"F":0,"M":0},"gToday":{"F":200,"M":140},"weight":null,"lamps":["abatere"],"layHall":60}]},`+
			`{"hall":"H7","emptySince":"2026-11-14","label":"goală din 14.11.2026 · H7"}]`,
		`"alerts":[],"latest":[]`,
	)

	// Before 12:00 the gap is not shown yet (missing.from_hour).
	in := f.input(owner(), f.hirbovat, nil)
	in.Now = time.Date(2026, time.November, 18, 11, 59, 0, 0, timex.Chisinau)
	if got := f.summaryJSON(t, in); strings.Contains(got, `"lipsa"`) {
		t.Errorf("lipsă before missing.from_hour: %s", got)
	}

	// The period moves the readout series: 4 weeks = SAPT 56–59.
	in = f.input(owner(), f.hirbovat, nil)
	in.Period = calc.Period{Kind: calc.PeriodWeeks, Weeks: 4}
	got = f.summaryJSON(t, in)
	expectJSON(t, got, `"series":[{"sapt":56,"value":null,"guide":null},{"sapt":57,`)
}

// periodSummary is what TestBuildSummaryPeriods checks of one summary.
type periodSummary struct {
	Sapts             []int // the readout series' weeks
	ReadoutSapt       int   // the latest week with a lay %
	AliveF, AliveM    int
	SoldF, DeathsF    int
	LayValue          string // layToday.value, "null" without an egg entry
	LayEggs           int
	DeathsWeek        map[string]string // section → "♀ · ♂"
	EmptyHalls, Halls []string
}

func summarize(t *testing.T, res SummaryResponse) periodSummary {
	t.Helper()
	var out periodSummary
	if res.Readout == nil {
		t.Fatal("no readout")
	}
	for _, p := range res.Readout.Series {
		out.Sapts = append(out.Sapts, p.Sapt)
		if p.Value != nil {
			out.ReadoutSapt = p.Sapt
		}
	}
	for _, st := range res.Stats {
		switch st := st.(type) {
		case AliveStat:
			if st.F == nil || st.M == nil || st.Sold.F == nil {
				t.Fatalf("alive = %+v", st)
			}
			out.AliveF, out.AliveM, out.SoldF = *st.F, *st.M, *st.Sold.F
		case CumMortalityStat:
			if st.Deaths.F == nil {
				t.Fatalf("cumMortality = %+v", st)
			}
			out.DeathsF = *st.Deaths.F
		case LayTodayStat:
			out.LayValue, out.LayEggs = "null", st.Eggs
			if st.Value != nil {
				out.LayValue = st.Value.String()
			}
		}
	}
	out.DeathsWeek = map[string]string{}
	for _, tile := range res.Tiles {
		switch tile := tile.(type) {
		case HallTile[LayingTileSection]:
			out.Halls = append(out.Halls, tile.Hall)
			for _, sec := range tile.Sections {
				out.DeathsWeek[sec.Code] = fmt.Sprintf("%d · %d", sec.DeathsWeek.F, sec.DeathsWeek.M)
			}
		case EmptyHallTile:
			out.EmptyHalls = append(out.EmptyHalls, tile.Hall)
		default:
			t.Fatalf("tile %T on a laying farm", tile)
		}
	}
	return out
}

// TestBuildSummaryFeedDaysLeft: the feed days left are shown only for the
// open cycle with a period that ends today (M3 D10); a stock below 0 shows
// 0 days, running out at the window end.
func TestBuildSummaryFeedDaysLeft(t *testing.T) {
	f := newFixture(t)
	f.summaryData()
	shown := func(t *testing.T, in Input) *FeedDaysLeftStat {
		t.Helper()
		res, err := BuildSummary(t.Context(), f.d.App, in)
		if err != nil {
			t.Fatal(err)
		}
		var out *FeedDaysLeftStat
		for i, st := range res.Stats {
			if st, ok := st.(FeedDaysLeftStat); ok {
				out = &st
				if _, isFeedToday := res.Stats[i-1].(FeedTodayStat); !isFeedToday {
					t.Errorf("feedDaysLeft follows %T, want feedToday", res.Stats[i-1])
				}
			}
		}
		return out
	}
	period := func(p calc.Period) Input {
		in := f.input(owner(), f.hirbovat, nil)
		in.Period = p
		return in
	}
	for _, tc := range []struct {
		name  string
		in    Input
		shown bool
	}{
		{"open cycle, whole cycle", f.input(owner(), f.hirbovat, nil), true},
		{"open cycle asked by id, weeks:4", func() Input {
			in := period(calc.Period{Kind: calc.PeriodWeeks, Weeks: 4})
			in.CycleID = &f.cycleH26
			return in
		}(), true},
		{"interval ending today", period(calc.Period{Kind: calc.PeriodInterval, From: day(2026, time.November, 1), To: today}), true},
		{"interval ending yesterday", period(calc.Period{Kind: calc.PeriodInterval, From: day(2026, time.November, 1), To: day(2026, time.November, 17)}), false},
		{"past cycle", f.input(owner(), f.hirbovat, &f.cycleH25), false},
		{"farm without an open cycle", f.input(owner(), f.telita, nil), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shown(t, tc.in); (got != nil) != tc.shown {
				t.Errorf("feedDaysLeft = %+v, want shown %v", got, tc.shown)
			}
		})
	}

	// Consumption past the deliveries: 0 days, running out at the window end.
	e := f.entry("feed_daily", f.hirbovat, nil, day(2026, time.November, 17))
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, kg_recorded)
		select $1, $2, $3, $4, 'M', id, 10 from feed_type where name_key = 'cocosi'`,
		e, f.hirbovat, f.sec["H5.1"], day(2026, time.November, 17))
	st := shown(t, f.input(owner(), f.hirbovat, nil))
	if st == nil || len(st.Items) != 3 {
		t.Fatalf("feedDaysLeft = %+v, want 3 items", st)
	}
	b, err := json.Marshal(st.Items[2])
	if err != nil {
		t.Fatal(err)
	}
	// 6 − 2,84 − 10 = −6,84 kg.
	const want = `{"type":"Cocoși","forSex":"M","days":0,"runOut":"2026-11-18","kg":-6.84,"lamp":"abatere","reason":"stoc sub 0 · verifică livrările"}`
	if string(b) != want {
		t.Errorf("Cocoși = %s, want %s", b, want)
	}
}

// TestBuildSummaryPeriods: the period sets the readout series and the day
// the stats and tiles are taken on (§7.6): its last day, with the counts
// that build up from the cycle start and deathsWeek over SAPT(day) ∩ cycle.
func TestBuildSummaryPeriods(t *testing.T) {
	f := newFixture(t)
	f.summaryData()
	// An egg entry inside the interval, so its readout has a week to show.
	d := day(2026, time.November, 11)
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs)
		values ($1, $2, $3, $4, 0, 0, 0, 0, 30)`, f.entry("daily", f.hirbovat, nil, d), f.hirbovat, f.sec["H5"], d)
	for _, tc := range []struct {
		name   string
		period calc.Period
		want   periodSummary
	}{
		{
			// 18.11 (SAPT 59: 17–18.11): mortality 1 + cull 1 in H5.1.
			name:   "weeks:4",
			period: calc.Period{Kind: calc.PeriodWeeks, Weeks: 4},
			want: periodSummary{
				Sapts: []int{56, 57, 58, 59}, ReadoutSapt: 59, AliveF: 88, AliveM: 20, SoldF: 8, DeathsF: 4,
				LayValue: "47.7777777777777778", LayEggs: 43,
				DeathsWeek: map[string]string{"H5.1": "2 · 0", "H5.2": "0 · 0", "H6.1": "0 · 0"},
				EmptyHalls: []string{"H7"}, Halls: []string{"H5", "H6"},
			},
		},
		{
			// 12.11 (SAPT 58: 10–12.11): the 10.11 mortality ♀ 2 in H5.1, the
			// 12.11 sale ♀ 5; H7 still holds the 3 hens moved on 01.11 (sold 14.11).
			name:   "interval ending mid-cycle",
			period: calc.Period{Kind: calc.PeriodInterval, From: day(2026, time.November, 1), To: day(2026, time.November, 12)},
			want: periodSummary{
				Sapts: []int{56, 57, 58}, ReadoutSapt: 58, AliveF: 93, AliveM: 20, SoldF: 5, DeathsF: 2,
				LayValue: "null", LayEggs: 0,
				DeathsWeek: map[string]string{"H5.1": "2 · 0", "H5.2": "0 · 0", "H6.1": "0 · 0", "H7": "0 · 0"},
				Halls:      []string{"H5", "H6", "H7"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := f.input(owner(), f.hirbovat, nil)
			in.Period = tc.period
			res, err := BuildSummary(t.Context(), f.d.App, in)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, summarize(t, res)); diff != "" {
				t.Errorf("summary (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildSummaryWithoutGuide(t *testing.T) {
	f := newFixture(t)
	f.summaryData()
	f.exec(`update cycle set guide_id = null where id = $1`, f.cycleH26)
	got := f.summaryJSON(t, f.input(owner(), f.hirbovat, nil))
	expectJSON(t, got,
		`"readout":{"kind":"lay","label":"SAPT 59 · efectiv 46,7% · ghid – fără ghid","value":46.6667,"guide":null,"series":[`,
		`{"id":"cumMortality","F":4,"M":0,"guideF":null,"reason":"no_guide","note":"(ghid din intrarea în ouat)","deaths":{"F":4,"M":0}}`,
		`{"id":"weight","F":3960,"M":4870,"guide":null,"reason":"no_guide","at":"2026-11-10","section":"H5.1"}`,
		`{"id":"feedToday","F":162.2222,"M":142,"guideF":null,"guideM":null,"reason":"no_guide","note":"H7 goală din 14.11"}`,
		`{"id":"layToday","value":47.7778,"guide":null,"diffPts":null,"reason":"no_guide","eggs":43,"halls":2}`,
		// No guide, no abatere; the gap still shows.
		`"lamps":["lipsa"],"layHall":47.0588}`,
	)
	// The feed days left compare no guide: their own lamps are left out here.
	got = regexp.MustCompile(`\{"id":"feedDaysLeft".*?\]\}`).ReplaceAllString(got, "")
	if strings.Contains(got, `"abatere"`) || strings.Contains(got, `"status"`) {
		t.Errorf("no guide, no comparison: %s", got)
	}
}

// nnbsp is U+202F, the thousands separator of §11.3.
const nnbsp = "\u202f"

func TestBuildSummaryGrowthCycle(t *testing.T) {
	f := newFixture(t)
	d := day(2026, time.February, 24) // SAPT 21: compared with guide week 20
	e := f.entry("weighing", f.telita, nil, d)
	f.exec(`insert into weighing (entry_id, farm_id, section_id, date, sex, avg_weight_g) values ($1, $2, $3, $4, 'F', 2553)`,
		e, f.telita, f.sec["T1"], d)
	f.exec(`insert into guide_week (guide_id, week, sex, body_weight_g) values ($1, 20, 'F', 2370)`, f.guideGrow)

	got := f.summaryJSON(t, f.input(owner(), f.telita, nil))
	expectJSON(t, got,
		`"readout":{"kind":"weight","label":"SAPT 21 · efectiv 2`+nnbsp+`553 g · ghid 2`+nnbsp+`370 g · +7,7%","value":2553,"guide":2370,"status":"abatere","series":[]}`,
		`{"id":"alive","F":0,"M":0,"arrivals":{"F":100,"M":20},"sold":{"F":0,"M":0}}`,
		`{"id":"weight","F":2553,"M":null,"guide":2370,"status":"abatere","at":"2026-02-24","section":"T1"}`,
		`"tiles":[{"hall":"T1","emptySince":"2026-02-28","label":"goală din 28.02.2026 · T1"}]`,
	)
	if strings.Contains(got, `"layToday"`) || strings.Contains(got, `"note":"(ghid`) {
		t.Errorf("a growth cycle has no lay stat and no lay-start note: %s", got)
	}
}

func TestBuildSummaryEdges(t *testing.T) {
	f := newFixture(t)
	const empty = `{"readout":null,"stats":[],"tiles":[],"alerts":[],"latest":[]}`
	if got := f.summaryJSON(t, f.input(owner(), f.noua, nil)); got != empty {
		t.Errorf("farm without a cycle = %s, want %s", got, empty)
	}
	in := f.input(owner(), f.hirbovat, nil)
	in.Period = calc.Period{Kind: calc.PeriodInterval, From: day(2025, time.January, 1), To: day(2025, time.January, 31)}
	if got := f.summaryJSON(t, in); got != empty {
		t.Errorf("a period outside the cycle = %s, want %s", got, empty)
	}

	for _, tc := range []struct {
		name string
		in   Input
		want error
	}{
		{"unknown farm", f.input(owner(), f.cycT, nil), ErrFarmNotFound},
		{"another farm's cycle", f.input(owner(), f.hirbovat, &f.cycT), ErrCycleNotFound},
		{"interval in the future", func() Input {
			in := f.input(owner(), f.hirbovat, nil)
			in.Period = calc.Period{Kind: calc.PeriodInterval, From: day(2026, time.November, 1), To: day(2026, time.November, 20)}
			return in
		}(), calc.ErrIntervalFuture},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildSummary(t.Context(), f.d.App, tc.in); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestSummaryHandler(t *testing.T) {
	f := newFixture(t)
	hirbovat := f.hirbovat.String()
	vet := principal(auth.Vet, f.hirbovat)
	viewer := principal(auth.Viewer, f.telita)
	manager := principal(auth.Manager, f.hirbovat)
	path := "/api/farms/" + hirbovat + "/summary"

	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		path      string
		status    int
		code      string
		issuePath string
		issueCode string
		message   string
	}{
		{"vet", &vet, path, http.StatusForbidden, "FORBIDDEN", "", "", ""},
		{"viewer on another farm", &viewer, path, http.StatusForbidden, "FARM_FORBIDDEN", "", "", ""},
		{"another farm's cycle", &manager, path + "?cycleId=" + f.cycT.String(), http.StatusNotFound, "NOT_FOUND", "", "", ""},
		{"unknown parameter", &manager, path + "?tab=rezumat", http.StatusBadRequest, "VALIDATION", "tab", "unknown_field", ""},
		{"bad period", &manager, path + "?period=weeks:x", http.StatusBadRequest, "VALIDATION", "period", "bad_type", ""},
		{"repeated period", &manager, path + "?period=cycle&period=cycle", http.StatusBadRequest, "VALIDATION", "period", "bad_type", ""},
		{"interval in the future", &manager, path + "?period=interval:2026-11-01,2026-11-20", http.StatusBadRequest, "VALIDATION",
			"period", "out_of_range", "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{"manager", &manager, path + "?period=weeks:4&cycleId=" + f.cycleH26.String(), http.StatusOK, "", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := get(t, f.server(tc.caller), tc.path)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			if tc.status == http.StatusOK {
				if !strings.Contains(rec.Body.String(), `"stats":[{"id":"alive"`) {
					t.Errorf("body = %s", rec.Body)
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
