package periodtable

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func day(m time.Month, d int) timex.Day { return timex.MustDay(2026, m, d) }

// farmT is a hand-built laying farm. Seria T is placed 06.10.2026 at age 140,
// so its weeks run Tue–Mon and 18.11.2026 is age 183, SAPT 27 (17–18.11, 2
// din 7). The cycle opens 08.10 (SAPT 21, 08–12.10, 5 din 7): SAPT 21–27,
// 7 weeks.
//
//	Halls H1 (H1.1, H1.2) and H2 (no compartments: its own egg unit).
//	08.10 placement H1.1 ♀ 100 ♂ 10, H1.2 ♀ 100 ♂ 10, H2 ♀ 50 ♂ 5.
//	Mortality rows every day for every section × sex, 0 except 10.10 H1.1 ♀ 2,
//	17.11 H1.2 ♀ 1, 18.11 H2 ♀ 3. Cull 17.11 H1.1 ♀ 1. Sale 18.11 H2 ♀ 47 ♂ 5
//	(H2 is empty at the end of 18.11).
//	Eggs from 09.10: H1 1 + 2 + 3 + 150 = 156 (jos 4, 60 g) every day; H2
//	0 + 1 + 1 + 30 = 32 (jos 2, 62 g) to 17.11, none on 18.11.
//	Feed every day for every section × sex, ♀ 100 g, ♂ 120 g; none for H2 on 18.11.
//	Weighing 05.11 H1.1 ♀ 2 000 g.
//	Guide: lay % SAPT 21–27 = 50, 60, 70, 80, 85, 86, 87; feed ♀ 110 g, ♂ 125 g.
type farmT struct {
	t      *testing.T
	d      *testdb.DB
	farm   uuid.UUID
	cycle  uuid.UUID
	guide  uuid.UUID
	series uuid.UUID
	sec    map[string]uuid.UUID
}

func (f *farmT) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *farmT) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

func (f *farmT) entry(kind string) uuid.UUID {
	f.t.Helper()
	var series *uuid.UUID
	if kind == "placement" {
		series = &f.series
	}
	return f.id(`insert into entry (kind, farm_id, date, series_id) values ($1, $2, $3, $4)`, kind, f.farm, day(time.October, 8), series)
}

func newFarmT(t *testing.T) *farmT {
	t.Helper()
	f := &farmT{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series = f.id(`insert into series (name, name_key, year, placement_date, age_at_placement_days)
		values ('Seria T', 'seria t', 2030, $1, 140)`, day(time.October, 6))
	f.guide = f.id(`insert into guide (name, name_key, type) values ('Ghid T', 'ghid t', 'laying')`)
	f.exec(`insert into guide_week (guide_id, week, sex, feed_g, lay_pct)
		select $1::uuid, w, 'F'::sex, 110, (array[50, 60, 70, 80, 85, 86, 87])[w - 20] from generate_series(21, 27) w
		union all select $1::uuid, w, 'M'::sex, 125, null from generate_series(21, 27) w`, f.guide)
	f.farm = f.id(`insert into farm (name, name_key, type, guide_id) values ('Ferma T', 'ferma t', 'laying', $1)`, f.guide)
	f.cycle = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, started_by) values ($1, $2, $3, $4, 'auto')`,
		f.series, f.farm, f.guide, day(time.October, 8))
	sec := func(code string, parent *uuid.UUID, sort int) uuid.UUID {
		id := f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, f.farm, code, parent, sort)
		f.sec[code] = id
		return id
	}
	h1 := sec("H1", nil, 1)
	sec("H1.1", &h1, 1)
	sec("H1.2", &h1, 2)
	sec("H2", nil, 2)
	leaves := []uuid.UUID{f.sec["H1.1"], f.sec["H1.2"], f.sec["H2"]}

	e := f.entry("placement")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, s, $3, x.sex::sex, 'placement', x.qty
		from unnest($4::uuid[], $5::text[], $6::int[]) as x(s, sex, qty)`,
		e, f.farm, day(time.October, 8),
		[]uuid.UUID{f.sec["H1.1"], f.sec["H1.1"], f.sec["H1.2"], f.sec["H1.2"], f.sec["H2"], f.sec["H2"]},
		[]string{"F", "M", "F", "M", "F", "M"}, []int{100, 10, 100, 10, 50, 5})

	e = f.entry("daily")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, s, d::date, x.sex::sex, 'mortality',
		       case when s = $5 and x.sex = 'F' and d::date = '2026-10-10' then -2
		            when s = $6 and x.sex = 'F' and d::date = '2026-11-17' then -1
		            when s = $7 and x.sex = 'F' and d::date = '2026-11-18' then -3
		            else 0 end
		from unnest($3::uuid[]) s, generate_series($4::date, '2026-11-18'::date, interval '1 day') d,
		     (values ('F'), ('M')) x(sex)`,
		e, f.farm, leaves, day(time.October, 8), f.sec["H1.1"], f.sec["H1.2"], f.sec["H2"])
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
		values ($1, $2, $3, '2026-11-17', 'F', 'cull', -1, 'boala')`, e, f.farm, f.sec["H1.1"])
	e = f.entry("sale")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		values ($1, $2, $3, '2026-11-18', 'F', 'sale', -47), ($1, $2, $3, '2026-11-18', 'M', 'sale', -5)`, e, f.farm, f.sec["H2"])

	e = f.entry("daily")
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g)
		select $1::uuid, $2::uuid, $3::uuid, d::date, 4, 1, 2, 3, 150, 60 from generate_series('2026-10-09'::date, '2026-11-18'::date, interval '1 day') d
		union all
		select $1::uuid, $2::uuid, $4::uuid, d::date, 2, 0, 1, 1, 30, 62 from generate_series('2026-10-09'::date, '2026-11-17'::date, interval '1 day') d`,
		e, f.farm, f.sec["H1"], f.sec["H2"])

	ft := f.id(`insert into feed_type (name, name_key) values ('Breeder T', 'breeder t')`)
	e = f.entry("feed_daily")
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		select $1, $2, s, d::date, x.sex::sex, $4, x.g
		from unnest($3::uuid[]) s, generate_series('2026-10-08'::date, '2026-11-18'::date, interval '1 day') d,
		     (values ('F', 100), ('M', 120)) x(sex, g)
		where not (s = $5 and d::date = '2026-11-18')`, e, f.farm, leaves, ft, f.sec["H2"])

	e = f.entry("weighing")
	f.exec(`insert into weighing (entry_id, farm_id, section_id, date, sex, avg_weight_g) values ($1, $2, $3, '2026-11-05', 'F', 2000)`,
		e, f.farm, f.sec["H1.1"])
	return f
}

func (f *farmT) build(q TableQuery) TableResponse {
	f.t.Helper()
	resp, err := f.buildErr(q)
	if err != nil {
		f.t.Fatalf("Build: %v", err)
	}
	return resp
}

func (f *farmT) buildErr(q TableQuery) (TableResponse, error) {
	q.View = "farm"
	q.FarmID = &f.farm
	return Build(context.Background(), f.d.App, Input{Query: q, Today: testdb.Today, Now: testdb.Now()})
}

func intp(i int) *int { return &i }

// cell is c[key] at 4 decimals as text, "null" for a null cell and "absent"
// when the key is missing.
func cell(c Cells, key string) string {
	v, ok := c[key]
	switch {
	case !ok:
		return "absent"
	case v == nil:
		return "null"
	default:
		return num.Round(v.Decimal(), 4).String()
	}
}

func cellsOf(c Cells, keys ...string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+cell(c, k))
	}
	return out
}

func diff(t *testing.T, what string, want, got any) {
	t.Helper()
	if d := cmp.Diff(want, got, cmp.AllowUnexported(timex.Day{})); d != "" {
		t.Errorf("%s (-want +got):\n%s", what, d)
	}
}

func TestBuildDefaultPage(t *testing.T) {
	f := newFarmT(t)
	resp := f.build(TableQuery{})

	diff(t, "scope", Scope{View: "farm", FarmID: &f.farm, CycleID: &f.cycle, SeriesID: &f.series, Label: "Total fermă"}, resp.Scope)
	diff(t, "period", PeriodInfo{Kind: "cycle", Label: "Ciclul curent", SaptFrom: 21, SaptTo: 27,
		From: day(time.October, 8), To: day(time.November, 18), WeeksTotal: 7}, resp.Period)
	diff(t, "chips", []string{"mortalitate", "sacrificare", "populare", "vanzare", "hrana", "cantarire", "oua"}, resp.Chips)
	var ids []string
	for _, g := range resp.Groups {
		ids = append(ids, g.ID)
	}
	diff(t, "groups", []string{"vii", "morti", "ouat", "hranaKg", "vanzare"}, ids)
	var ouat []string
	for _, l := range resp.Groups[2].Leaves {
		ouat = append(ouat, l.ID+" "+l.Label+" "+l.Kind)
	}
	diff(t, "ouat leaves", []string{"egg.total ouă count", "egg.layPct % ouat rate", "egg.guide ghid guide",
		"egg.inc incub. count", "egg.incPct % incub. rate", "egg.jos jos count", "egg.josPct % jos rate"}, ouat)
	diff(t, "notes", []string{"Morți include sacrificările, ca în registru."}, resp.Notes)
	if resp.NoGuide || resp.Empty != nil {
		t.Errorf("NoGuide %v, Empty %v", resp.NoGuide, resp.Empty)
	}

	var sapts []int
	for _, w := range resp.Weeks {
		sapts = append(sapts, w.Sapt)
	}
	diff(t, "weeks newest first", []int{27, 26, 25, 24}, sapts)
	diff(t, "loaded", &Loaded{SaptFrom: 24, SaptTo: 27, Weeks: 4}, resp.Loaded)
	diff(t, "more", &More{Before: 24, Remaining: 3, Label: "Încarcă încă 3 săptămâni · SAPT 24–27 · 4 din 7"}, resp.More)
	if resp.Start != nil {
		t.Errorf("start = %+v, want null", resp.Start)
	}

	// SAPT 27: 17–18.11, the hen-days count H2 on 18.11, the day it is
	// emptied with no egg entry (M2 Q1): 248 + (196 + 50) = 494, not 444.
	w := resp.Weeks[0]
	if w.Label != "SAPT 27 · 17.11–18.11 · 2 din 7 zile" || w.DaysInRow != 2 || !w.Partial {
		t.Errorf("SAPT 27 row = %q, %d days, partial %v", w.Label, w.DaysInRow, w.Partial)
	}
	diff(t, "SAPT 27 cells", []string{
		"alive.F=196", "alive.M=20", "deaths.F=5", "deaths.M=0",
		"egg.total=344", "egg.layPct=69.6356", "egg.guide=87", "egg.inc=330", "egg.incPct=95.9302",
		"egg.jos=10", "egg.josPct=2.907", "feed.kg.F=44.4", "feed.kg.M=5.4", "sold.F=47", "sold.M=5",
	}, cellsOf(w.Cells, KeyAliveF, KeyAliveM, KeyDeathsF, KeyDeathsM, KeyEggs, KeyLayPct, KeyLayGuide, KeyInc, KeyIncPct,
		KeyJos, KeyJosPct, KeyFeedKgF, KeyFeedKgM, KeySoldF, KeySoldM))
	diff(t, "SAPT 27 status", map[string]string{"egg.layPct": "abatere"}, w.Status)
	if w.Reasons != nil {
		t.Errorf("SAPT 27 reasons = %v", w.Reasons)
	}

	// Its days, by date.
	if len(w.Days) != 2 {
		t.Fatalf("SAPT 27 has %d days", len(w.Days))
	}
	d0, d1 := w.Days[0], w.Days[1]
	diff(t, "day heads", []string{"2026-11-17 182 27 Ma", "2026-11-18 183 27 Mi"}, []string{
		d0.Date.String() + " 182 27 " + d0.Weekday, d1.Date.String() + " 183 27 " + d1.Weekday})
	if d0.AgeDays != 182 || d1.AgeDays != 183 || d1.Sapt != 27 {
		t.Errorf("ages %d, %d; SAPT %d", d0.AgeDays, d1.AgeDays, d1.Sapt)
	}
	diff(t, "18.11 cells", []string{"alive.F=196", "deaths.F=3", "egg.total=156", "egg.layPct=63.4146", "egg.guide=87",
		"feed.kg.F=19.6", "feed.kg.M=2.4", "sold.F=47"},
		cellsOf(d1.Cells, KeyAliveF, KeyDeathsF, KeyEggs, KeyLayPct, KeyLayGuide, KeyFeedKgF, KeyFeedKgM, KeySoldF))
	diff(t, "17.11 cells", []string{"alive.F=246", "deaths.F=2", "egg.total=188", "egg.layPct=75.8065"},
		cellsOf(d0.Cells, KeyAliveF, KeyDeathsF, KeyEggs, KeyLayPct))
	diff(t, "18.11 status", map[string]string{"egg.layPct": "abatere"}, d1.Status)

	// A whole week: SAPT 26 = 10–16.11, 7 × 188 eggs over 7 × 248 hen-days.
	w = resp.Weeks[1]
	if w.Label != "SAPT 26 · 10.11–16.11" || w.Partial || len(w.Days) != 7 {
		t.Errorf("SAPT 26 row = %q, partial %v, %d days", w.Label, w.Partial, len(w.Days))
	}
	diff(t, "SAPT 26 cells", []string{"deaths.F=0", "egg.total=1316", "egg.layPct=75.8065", "egg.guide=86", "sold.F=0"},
		cellsOf(w.Cells, KeyDeathsF, KeyEggs, KeyLayPct, KeyLayGuide, KeySoldF))

	// Totals over SAPT 24–27 (27.10–18.11): eggs 22 × 188 + 156 = 4 292 over
	// 21 × 248 + 248 + 246 = 5 702 hen-days; guide (80·1 736 + 85·1 736 +
	// 86·1 736 + 87·494) ÷ 5 702; alive at the latest day.
	if resp.Totals.Label != "Total SAPT 24–27" {
		t.Errorf("totals label %q", resp.Totals.Label)
	}
	diff(t, "totals", []string{"alive.F=196", "alive.M=20", "deaths.F=5", "egg.total=4292", "egg.layPct=75.2718",
		"egg.guide=83.9555", "sold.F=47", "sold.M=5"},
		cellsOf(resp.Totals.Cells, KeyAliveF, KeyAliveM, KeyDeathsF, KeyEggs, KeyLayPct, KeyLayGuide, KeySoldF, KeySoldM))
	diff(t, "totals status", map[string]string{"egg.layPct": "abatere"}, resp.Totals.Status)
}

func TestBuildPaging(t *testing.T) {
	f := newFarmT(t)

	// The cycle's first weeks: SAPT 21 is 08–12.10 (5 din 7), then the start note.
	resp := f.build(TableQuery{Before: intp(24)})
	var sapts []int
	for _, w := range resp.Weeks {
		sapts = append(sapts, w.Sapt)
	}
	diff(t, "weeks", []int{23, 22, 21}, sapts)
	diff(t, "loaded", &Loaded{SaptFrom: 21, SaptTo: 27, Weeks: 7}, resp.Loaded)
	if resp.More != nil || resp.Start == nil || resp.Start.Label != "Începutul ciclului · SAPT 21" {
		t.Errorf("more %+v, start %+v", resp.More, resp.Start)
	}
	if resp.Totals.Label != "Total SAPT 21–27" {
		t.Errorf("totals label %q", resp.Totals.Label)
	}
	w21 := resp.Weeks[2]
	if w21.Label != "SAPT 21 · 08.10–12.10 · 5 din 7 zile" || w21.DaysInRow != 5 || len(w21.Days) != 5 {
		t.Errorf("SAPT 21 = %q, %d days in row, %d day rows", w21.Label, w21.DaysInRow, len(w21.Days))
	}
	// Eggs from 09.10: 4 × 188 over hen-days 250 + 250 + 248 + 248 (deaths
	// 2 on 10.10); 08.10 has birds placed that day, none at its start.
	diff(t, "SAPT 21", []string{"egg.total=752", "egg.layPct=75.502", "egg.guide=50", "deaths.F=2", "alive.F=248"},
		cellsOf(w21.Cells, KeyEggs, KeyLayPct, KeyLayGuide, KeyDeathsF, KeyAliveF))
	if c := w21.Days[0].Cells; cell(c, KeyEggs) != "null" || cell(c, KeyLayPct) != "null" || cell(c, KeyLayGuide) != "50" {
		t.Errorf("08.10 (no egg entry) = %v", cellsOf(c, KeyEggs, KeyLayPct, KeyLayGuide))
	}
	if r := w21.Days[0].Reasons; r != nil {
		t.Errorf("08.10 reasons = %v (eggs have no missing rule)", r)
	}

	// One week only; grain=week has no day rows.
	resp = f.build(TableQuery{Sapt: intp(25), Grain: "week"})
	if len(resp.Weeks) != 1 || resp.Weeks[0].Sapt != 25 || resp.Weeks[0].Days != nil {
		t.Fatalf("sapt=25&grain=week = %+v", resp.Weeks)
	}
	if resp.More != nil || resp.Start != nil || resp.Totals.Label != "Total SAPT 25" {
		t.Errorf("more %+v, start %+v, totals %q", resp.More, resp.Start, resp.Totals.Label)
	}
	diff(t, "SAPT 25 = its totals", cellsOf(resp.Weeks[0].Cells, KeyEggs, KeyLayPct, KeyLayGuide),
		cellsOf(resp.Totals.Cells, KeyEggs, KeyLayPct, KeyLayGuide))

	if _, err := f.buildErr(TableQuery{Before: intp(21)}); !errors.As(err, new(*calc.PageError)) {
		t.Errorf("before=21: err = %v, want a PageError", err)
	}
}

func TestBuildTypesAndPeriodChips(t *testing.T) {
	f := newFarmT(t)
	// A period without sales, culls or weighings: no vanzare group, and the
	// chips are the period's.
	resp := f.build(TableQuery{Period: Period{Kind: "interval", From: day(time.October, 8), To: day(time.October, 20)}})
	diff(t, "chips", []string{"mortalitate", "populare", "hrana", "oua"}, resp.Chips)
	var ids []string
	for _, g := range resp.Groups {
		ids = append(ids, g.ID)
	}
	diff(t, "groups", []string{"vii", "morti", "ouat", "hranaKg"}, ids)
	if _, ok := resp.Weeks[0].Cells[KeySoldF]; ok {
		t.Error("a row writes sold.F without the vanzare group")
	}
	if resp.Period.Label != "08.10–20.10.2026" || resp.Period.SaptFrom != 21 || resp.Period.SaptTo != 23 {
		t.Errorf("period = %+v", resp.Period)
	}

	// One chip (phone): vii is always there.
	resp = f.build(TableQuery{Types: []string{"vanzare"}})
	ids = nil
	for _, g := range resp.Groups {
		ids = append(ids, g.ID)
	}
	diff(t, "types=vanzare", []string{"vii", "vanzare"}, ids)
	diff(t, "chips stay the period's", []string{"mortalitate", "sacrificare", "populare", "vanzare", "hrana", "cantarire", "oua"}, resp.Chips)
	diff(t, "cells", []string{"alive.F=196", "deaths.F=absent", "egg.total=absent", "sold.F=47"},
		cellsOf(resp.Weeks[0].Cells, KeyAliveF, KeyDeathsF, KeyEggs, KeySoldF))
}

func TestBuildNoGuideAndPartialGuide(t *testing.T) {
	f := newFarmT(t)
	f.exec(`delete from guide_week where guide_id = $1 and week = 27`, f.guide)
	resp := f.build(TableQuery{})
	w := resp.Weeks[0]
	if cell(w.Cells, KeyLayGuide) != "null" || w.Reasons[KeyLayGuide] != ReasonNoGuide || w.Status != nil {
		t.Errorf("SAPT 27 without a guide row: guide %s, reasons %v, status %v", cell(w.Cells, KeyLayGuide), w.Reasons, w.Status)
	}
	if d := w.Days[0]; cell(d.Cells, KeyLayGuide) != "null" || d.Reasons[KeyLayGuide] != ReasonNoGuide {
		t.Errorf("17.11 without a guide row: %v %v", cell(d.Cells, KeyLayGuide), d.Reasons)
	}
	if cell(resp.Weeks[1].Cells, KeyLayGuide) != "86" {
		t.Errorf("SAPT 26 guide = %s", cell(resp.Weeks[1].Cells, KeyLayGuide))
	}
	if cell(resp.Totals.Cells, KeyLayGuide) != "null" || resp.Totals.Reasons[KeyLayGuide] != ReasonNoGuide {
		t.Errorf("totals over a partial guide: %s %v", cell(resp.Totals.Cells, KeyLayGuide), resp.Totals.Reasons)
	}

	f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
	resp = f.build(TableQuery{})
	if !resp.NoGuide {
		t.Error("noGuide = false for a cycle without a guide")
	}
	var leaves []string
	for _, l := range resp.Groups[2].Leaves {
		leaves = append(leaves, l.ID)
	}
	diff(t, "ouat leaves", []string{"egg.total", "egg.layPct", "egg.inc", "egg.incPct", "egg.jos", "egg.josPct"}, leaves)
	if _, ok := resp.Weeks[0].Cells[KeyLayGuide]; ok || resp.Weeks[0].Status != nil || resp.Totals.Status != nil {
		t.Errorf("no-guide rows carry a guide or a status: %v %v", resp.Weeks[0].Cells, resp.Weeks[0].Status)
	}
	if cell(resp.Weeks[0].Cells, KeyLayPct) != "69.6356" {
		t.Errorf("lay without a guide = %s", cell(resp.Weeks[0].Cells, KeyLayPct))
	}
}

func TestBuildMissing(t *testing.T) {
	f := newFarmT(t)
	// 16.11: no allocation and no mortality row at all.
	f.exec(`delete from feed_allocation where date = '2026-11-16'`)
	f.exec(`delete from bird_event where type = 'mortality' and date = '2026-11-16'`)
	resp := f.build(TableQuery{})
	w := resp.Weeks[1] // SAPT 26
	d := w.Days[6]
	if d.Date != day(time.November, 16) {
		t.Fatalf("day 6 of SAPT 26 is %s", d.Date)
	}
	diff(t, "16.11 cells", []string{"deaths.F=null", "feed.kg.F=null", "feed.kg.M=null", "alive.F=248"},
		cellsOf(d.Cells, KeyDeathsF, KeyFeedKgF, KeyFeedKgM, KeyAliveF))
	diff(t, "16.11 reasons", map[string]string{"deaths.F": "missing", "deaths.M": "missing", "feed.kg.F": "missing", "feed.kg.M": "missing"}, d.Reasons)
	// The week adds up the days that have entries: 6 × 24,8 kg ♀.
	diff(t, "SAPT 26", []string{"deaths.F=0", "feed.kg.F=148.8"}, cellsOf(w.Cells, KeyDeathsF, KeyFeedKgF))
	if w.Reasons != nil {
		t.Errorf("SAPT 26 reasons = %v", w.Reasons)
	}
}

func TestBuildEmptyStates(t *testing.T) {
	f := newFarmT(t)
	check := func(t *testing.T, resp TableResponse) {
		t.Helper()
		if resp.Empty == nil || resp.Empty.Next != "Alege ciclul curent" || resp.Loaded != nil || resp.More != nil ||
			resp.Start != nil || resp.Totals != nil || len(resp.Weeks) != 0 || len(resp.Chips) != 0 || len(resp.Groups) != 0 {
			t.Errorf("empty state = %+v", resp)
		}
		body, err := json.Marshal(resp)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{`"chips":[]`, `"groups":[]`, `"weeks":[]`, `"notes":[]`, `"more":null`, `"start":null`, `"empty":{"next":"Alege ciclul curent"}`} {
			if !strings.Contains(string(body), want) {
				t.Errorf("JSON lacks %s: %s", want, body)
			}
		}
		for _, absent := range []string{`"loaded"`, `"totals"`, `"noGuide"`} {
			if strings.Contains(string(body), absent) {
				t.Errorf("JSON has %s: %s", absent, body)
			}
		}
	}
	t.Run("interval before the cycle", func(t *testing.T) {
		resp := f.build(TableQuery{Period: Period{Kind: "interval", From: timex.MustDay(2025, 11, 1), To: timex.MustDay(2025, 11, 15)}})
		check(t, resp)
		if resp.Period.Label != "01.11–15.11.2025" || resp.Scope.CycleID == nil {
			t.Errorf("period %+v, scope %+v", resp.Period, resp.Scope)
		}
	})
	t.Run("a period with no records", func(t *testing.T) {
		g := newFarmT(t)
		// Only the placement (08.10) is left.
		g.exec(`delete from bird_event where type <> 'placement'`)
		for _, table := range []string{"egg_production", "feed_allocation", "weighing"} {
			g.exec(`delete from ` + table)
		}
		resp := g.build(TableQuery{Period: Period{Kind: "interval", From: day(time.October, 20), To: day(time.October, 25)}})
		check(t, resp)
	})
	t.Run("a farm without a cycle", func(t *testing.T) {
		other := f.id(`insert into farm (name, name_key, type) values ('Goală', 'goala', 'laying')`)
		resp, err := Build(context.Background(), f.d.App, Input{
			Query: TableQuery{View: "farm", FarmID: &other}, Today: testdb.Today, Now: testdb.Now()})
		if err != nil {
			t.Fatal(err)
		}
		check(t, resp)
		if resp.Scope.CycleID != nil || resp.Period.From != testdb.Today || resp.Period.Label != "Ciclul curent" {
			t.Errorf("scope %+v, period %+v", resp.Scope, resp.Period)
		}
	})
}

func TestBuildCycleAndErrors(t *testing.T) {
	f := newFarmT(t)
	other := newFarmT(t)
	_, err := Build(context.Background(), f.d.App, Input{
		Query: TableQuery{View: "farm", FarmID: &f.farm, CycleID: &other.cycle}, Today: testdb.Today, Now: testdb.Now()})
	if !errors.Is(err, calc.ErrNoCycle) {
		t.Errorf("another farm's cycle: err = %v, want calc.ErrNoCycle", err)
	}
	resp := f.build(TableQuery{CycleID: &f.cycle})
	if *resp.Scope.CycleID != f.cycle {
		t.Errorf("cycleId not used")
	}
	if _, err := f.buildErr(TableQuery{Period: Period{Kind: "interval", From: testdb.Today, To: testdb.Today.AddDays(1)}}); !errors.Is(err, calc.ErrIntervalFuture) {
		t.Errorf("future interval: err = %v", err)
	}
	for _, q := range []TableQuery{{Tab: "pasari"}, {Detail: "halls"}, {Grain: "month"}} {
		if _, err := f.buildErr(q); !errors.Is(err, ErrUnsupported) {
			t.Errorf("%+v: err = %v, want ErrUnsupported", q, err)
		}
	}
	// A cycle without a guide takes its type from the farm.
	f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
	f.exec(`update farm set type = 'growth', guide_id = null where id = $1`, f.farm)
	resp = f.build(TableQuery{})
	if resp.Groups[2].ID != GroupFeedG {
		t.Errorf("a growth farm's cycle without a guide has groups %v", resp.Groups)
	}
}

func TestWeekSeries(t *testing.T) {
	f := newFarmT(t)
	cyc, err := calc.CycleByID(context.Background(), f.d.App, f.farm, f.cycle)
	if err != nil {
		t.Fatal(err)
	}
	pts, err := WeekSeries(context.Background(), f.d.App, f.farm, cyc, calc.Period{Kind: calc.PeriodCycle}, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, p := range pts {
		got = append(got, strings.Join([]string{itoa(p.Sapt), dec4(p.Lay), dec4(p.Guide), itoa(p.Rollup.HenDays)}, " "))
	}
	diff(t, "series", []string{
		"21 75.502 50 996", "22 75.8065 60 1736", "23 75.8065 70 1736", "24 75.8065 80 1736",
		"25 75.8065 85 1736", "26 75.8065 86 1736", "27 69.6356 87 494",
	}, got)

	// The same figures as the table's week rows.
	resp := f.build(TableQuery{Count: 7})
	for _, w := range resp.Weeks {
		p := pts[w.Sapt-21]
		if cell(w.Cells, KeyLayPct) != dec4(p.Lay) || cell(w.Cells, KeyLayGuide) != dec4(p.Guide) {
			t.Errorf("SAPT %d: table %s/%s, series %s/%s", w.Sapt, cell(w.Cells, KeyLayPct), cell(w.Cells, KeyLayGuide), dec4(p.Lay), dec4(p.Guide))
		}
	}
	// weeks:4 on the same cycle.
	pts, err = WeekSeries(context.Background(), f.d.App, f.farm, cyc, calc.Period{Kind: calc.PeriodWeeks, Weeks: 4}, testdb.Today)
	if err != nil || len(pts) != 4 || pts[0].Sapt != 24 {
		t.Errorf("weeks:4 series = %v, %v", pts, err)
	}
}

func itoa(i int) string { return num.FromInt(int64(i)).String() }

func dec4(d *decimal.Decimal) string {
	if d == nil {
		return "null"
	}
	return num.Round(*d, 4).String()
}
