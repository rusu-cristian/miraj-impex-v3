package calc

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// decStr is the decimal's value as text, "nil" for nil.
func decStr(d *decimal.Decimal) string {
	if d == nil {
		return "nil"
	}
	return d.String()
}

func TestReadGuideWeeks(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	guide := f.id(`insert into guide (name, name_key, type) values ('Ghid T', 'ghid t', 'laying')`)
	other := f.id(`insert into guide (name, name_key, type) values ('Ghid U', 'ghid u', 'laying')`)
	row := `insert into guide_week (guide_id, week, sex, feed_g, body_weight_g, mortality_cum_pct, lay_pct, hatching_pct, egg_weight_g)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	f.exec(row, guide, 57, "F", "158", "3900", "7.00", "61.5", "96", "66.1")
	f.exec(row, guide, 58, "F", "157", nil, "7.10", "60.10", "96.5", "66.3")
	f.exec(row, guide, 58, "M", "156", "4800", nil, nil, nil, nil)
	f.exec(row, guide, 59, "F", "157", nil, "7.25", "57.6", nil, nil)
	f.exec(row, guide, 60, "F", "156", nil, "7.40", "57.0", nil, nil)
	f.exec(row, other, 58, "F", "999", nil, nil, "1", nil, nil)

	got, err := ReadGuideWeeks(ctx, f.d.App, guide, 58, 59)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("weeks = %v, want 58 and 59 only", got)
	}
	for _, tc := range []struct {
		name string
		cell *GuideCells
		want [6]string // feed, body weight, mortality, lay, hatching, egg weight
	}{
		{"58 ♀", got[58].F, [6]string{"157", "nil", "7.1", "60.1", "96.5", "66.3"}},
		{"58 ♂ (null cells)", got[58].M, [6]string{"156", "4800", "nil", "nil", "nil", "nil"}},
		{"59 ♀", got[59].F, [6]string{"157", "nil", "7.25", "57.6", "nil", "nil"}},
	} {
		if tc.cell == nil {
			t.Errorf("%s: no row", tc.name)
			continue
		}
		c := tc.cell
		if g := [6]string{decStr(c.FeedG), decStr(c.BodyWeightG), decStr(c.MortalityCumPct), decStr(c.LayPct),
			decStr(c.HatchingPct), decStr(c.EggWeightG)}; g != tc.want {
			t.Errorf("%s = %v, want %v", tc.name, g, tc.want)
		}
	}
	if got[59].M != nil || got[57].F != nil || got[60].F != nil {
		t.Errorf("59 ♂ %v, 57 ♀ %v, 60 ♀ %v; want nil (no row, or outside the range)", got[59].M, got[57].F, got[60].F)
	}

	if got, err := ReadGuideWeeks(ctx, f.d.App, uuid.New(), 1, 120); err != nil || len(got) != 0 {
		t.Errorf("unknown guide: %v, %v; want no weeks", got, err)
	}
	if _, err := ReadGuideWeeks(ctx, f.d.App, guide, 59, 58); !errors.Is(err, errRange) {
		t.Errorf("reversed range: err = %v, want errRange", err)
	}
}

func TestReadAppSettings(t *testing.T) {
	d := testdb.New(t)
	ctx := t.Context()
	got, err := ReadAppSettings(ctx, d.App)
	if err != nil {
		t.Fatal(err)
	}
	// The baseline values (D19, C9).
	n := decimal.NewFromInt
	if !got.ToleranceWeightPct.Equal(n(5)) || !got.ToleranceLayPts.Equal(n(3)) || !got.ToleranceFeedPct.Equal(n(10)) ||
		!got.StockDaysMin.Equal(n(5)) || got.StockAvgDays != 3 || !got.MortalityFactor.Equal(n(3)) ||
		got.MissingFromHour != 12 || got.MissingDaysBack != 7 {
		t.Errorf("settings = %+v", got)
	}

	exec := func(sql string) {
		t.Helper()
		if _, err := d.App.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	exec(`update app_setting set value = '2.5' where key = 'tolerance.lay_pts'`)
	if got, err := ReadAppSettings(ctx, d.App); err != nil || !got.ToleranceLayPts.Equal(decimal.RequireFromString("2.5")) {
		t.Errorf("a decimal threshold: %s, %v; want 2.5", got.ToleranceLayPts, err)
	}

	for _, tc := range []struct{ name, sql string }{
		{"not a number", `update app_setting set value = '"trei"' where key = 'tolerance.feed_pct'`},
		{"an hour that isn't whole", `update app_setting set value = '12.5' where key = 'missing.from_hour'`},
		{"JSON null", `update app_setting set value = 'null' where key = 'alert.stock_days_min'`},
		{"missing key", `delete from app_setting where key = 'missing.days_back'`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := d.App.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(context.Background()) }()
			if _, err := tx.Exec(ctx, tc.sql); err != nil {
				t.Fatal(err)
			}
			if _, err := ReadAppSettings(ctx, tx); !errors.Is(err, ErrSetting) {
				t.Errorf("err = %v, want ErrSetting", err)
			}
		})
	}
}

func TestCycleByIDAndFarmCycles(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	guide := f.id(`insert into guide (name, name_key, type) values ('Ghid T', 'ghid t', 'laying')`)
	f.exec(`update cycle set guide_id = $1 where id = $2`, guide, f.cycleA)
	// A past cycle at A without a flock count: no live bird event inside it (a deleted one doesn't count).
	legacySeries := f.id(`insert into series (name, name_key, year, placement_date, age_at_placement_days, status, closed_at)
		values ('Seria L', 'seria l', 2025, '2024-11-01', 0, 'closed', $1)`, testdb.Now())
	legacy := f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, start_inferred, ended_by)
		values ($1, $2, '2025-03-07', '2025-12-01', 'migration', true, 'migration')`, legacySeries, f.farmA)
	e := f.entry("migration", f.farmA, nil, timex.MustDay(2025, time.June, 1))
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, deleted_at)
		values ($1, $2, $3, '2025-06-01', 'F', 'placement', 50, $4)`, e, f.farmA, f.sec["H5.1"], testdb.Now())

	got, err := FarmCycles(ctx, f.d.App, f.farmA)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != f.cycleA || got[1].ID != legacy {
		t.Fatalf("cycles of A = %+v, want the open one, then the legacy one", got)
	}
	open, past := got[0], got[1]
	if open.FarmID != f.farmA || open.SeriesID != f.series || open.SeriesName != "Seria T" || open.StartDate != jan(10) ||
		open.EndDate != nil || open.EndedBy != nil || open.StartedBy != dbq.CycleStartKindAuto || open.StartInferred ||
		!open.HasBirdEvents || open.Age != (SeriesAge{PlacementDate: jan(5)}) {
		t.Errorf("open cycle = %+v", open)
	}
	if open.Guide == nil || *open.Guide != (CycleGuide{ID: guide, Name: "Ghid T", Type: dbq.FarmTypeLaying}) {
		t.Errorf("open cycle guide = %+v", open.Guide)
	}
	end := timex.MustDay(2025, time.December, 1)
	if past.SeriesName != "Seria L" || past.EndDate == nil || *past.EndDate != end || past.EndedBy == nil ||
		*past.EndedBy != dbq.CycleEndKindMigration || past.StartedBy != dbq.CycleStartKindMigration ||
		!past.StartInferred || past.HasBirdEvents || past.Guide != nil ||
		past.Age != (SeriesAge{PlacementDate: timex.MustDay(2024, time.November, 1)}) {
		t.Errorf("legacy cycle = %+v", past)
	}

	if got, err := FarmCycles(ctx, f.d.App, uuid.New()); err != nil || len(got) != 0 {
		t.Errorf("unknown farm: %+v, %v; want no cycles", got, err)
	}

	byID, err := CycleByID(ctx, f.d.App, f.farmB, f.cycleB)
	if err != nil {
		t.Fatal(err)
	}
	if byID.ID != f.cycleB || byID.FarmID != f.farmB || byID.EndDate == nil || *byID.EndDate != jan(10) ||
		byID.EndedBy == nil || *byID.EndedBy != dbq.CycleEndKindEmptied || !byID.HasBirdEvents || byID.Guide != nil {
		t.Errorf("cycle B = %+v", byID)
	}
	if byID, err := CycleByID(ctx, f.d.App, f.farmA, legacy); err != nil || byID.HasBirdEvents || !byID.StartInferred {
		t.Errorf("legacy by id = %+v, %v", byID, err)
	}
	for _, tc := range []struct {
		name          string
		farm, cycleID uuid.UUID
	}{
		{"another farm's cycle", f.farmA, f.cycleB},
		{"unknown id", f.farmA, uuid.New()},
	} {
		if _, err := CycleByID(ctx, f.d.App, tc.farm, tc.cycleID); !errors.Is(err, ErrNoCycle) {
			t.Errorf("%s: err = %v, want ErrNoCycle", tc.name, err)
		}
	}
}

func TestReadLatestWeighings(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	ins := `insert into weighing (entry_id, farm_id, section_id, date, sex, avg_weight_g, deleted_at)
		values ($1, $2, $3, $4, $5, $6, $7)`
	weigh := func(farm uuid.UUID, code string, d timex.Day, sex, g string, deleted bool) {
		var at *time.Time
		if deleted {
			now := testdb.Now()
			at = &now
		}
		f.exec(ins, f.entry("weighing", farm, nil, d), farm, f.sec[code], d, sex, g, at)
	}
	weigh(f.farmA, "H5.2", jan(11), "F", "1900", false)
	weigh(f.farmA, "H5.1", jan(12), "F", "2000", false)
	weigh(f.farmA, "H5.1", jan(13), "F", "2100", false)
	weigh(f.farmA, "H6.1", jan(13), "F", "2050", false) // same day as H5.1: H5 comes first in the tree
	weigh(f.farmA, "H6.1", jan(12), "M", "3000", false)
	weigh(f.farmA, "H6.1", jan(13), "M", "9999", true) // superseded
	weigh(f.farmB, "T1", jan(13), "F", "1500", false)  // another farm

	w := func(code string, d timex.Day, g string, alive int) *Weighing {
		return &Weighing{SectionID: f.sec[code], Code: code, Date: d, AvgWeightG: decimal.RequireFromString(g), AliveEnd: alive}
	}
	same := func(a, b *Weighing) bool {
		if a == nil || b == nil {
			return a == b
		}
		return a.SectionID == b.SectionID && a.Code == b.Code && a.Date == b.Date && a.AvgWeightG.Equal(b.AvgWeightG) &&
			a.AliveEnd == b.AliveEnd
	}
	for _, tc := range []struct {
		name     string
		from, to timex.Day
		farm     PerSex[*Weighing]
		sections map[string]PerSex[*Weighing]
	}{
		{"latest per section × sex; deleted rows and other farms never count", jan(10), jan(13),
			PerSex[*Weighing]{F: w("H5.1", jan(13), "2100", 44), M: w("H6.1", jan(12), "3000", 9)},
			map[string]PerSex[*Weighing]{
				"H5.1": {F: w("H5.1", jan(13), "2100", 44)},
				"H5.2": {F: w("H5.2", jan(11), "1900", 37)},
				"H6.1": {F: w("H6.1", jan(13), "2050", 8), M: w("H6.1", jan(12), "3000", 9)},
			}},
		{"on or before the day", jan(10), jan(12),
			PerSex[*Weighing]{F: w("H5.1", jan(12), "2000", 55), M: w("H6.1", jan(12), "3000", 9)},
			map[string]PerSex[*Weighing]{
				"H5.1": {F: w("H5.1", jan(12), "2000", 55)},
				"H5.2": {F: w("H5.2", jan(11), "1900", 37)},
				"H6.1": {M: w("H6.1", jan(12), "3000", 9)},
			}},
		{"the range starts after a section's weighing", jan(12), jan(12),
			PerSex[*Weighing]{F: w("H5.1", jan(12), "2000", 55), M: w("H6.1", jan(12), "3000", 9)},
			map[string]PerSex[*Weighing]{
				"H5.1": {F: w("H5.1", jan(12), "2000", 55)},
				"H6.1": {M: w("H6.1", jan(12), "3000", 9)},
			}},
		{"no weighing in the range", jan(14), jan(20), PerSex[*Weighing]{}, map[string]PerSex[*Weighing]{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadLatestWeighings(ctx, f.d.App, f.farmA, tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if !same(got.Farm.F, tc.farm.F) || !same(got.Farm.M, tc.farm.M) {
				t.Errorf("farm = ♀ %+v · ♂ %+v, want ♀ %+v · ♂ %+v", got.Farm.F, got.Farm.M, tc.farm.F, tc.farm.M)
			}
			if len(got.Sections) != len(tc.sections) {
				t.Errorf("sections = %d, want %d", len(got.Sections), len(tc.sections))
			}
			for code, want := range tc.sections {
				g := got.Sections[f.sec[code]]
				if !same(g.F, want.F) || !same(g.M, want.M) {
					t.Errorf("%s = ♀ %+v · ♂ %+v, want ♀ %+v · ♂ %+v", code, g.F, g.M, want.F, want.M)
				}
			}
		})
	}
	if _, err := ReadLatestWeighings(ctx, f.d.App, f.farmA, jan(13), jan(12)); !errors.Is(err, errRange) {
		t.Errorf("reversed range: err = %v, want errRange", err)
	}
}

func TestSectionsOnDay(t *testing.T) {
	f := newFixture(t)
	f.feed()
	ctx := t.Context()
	// 31.12, after app.today: H5.1 ♀ 100 g × 44 = 4,4 kg (the count at the end of 13.01, nothing since).
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, 'F', $5, 100)`, f.entry("feed_daily", f.farmA, nil, dec31), f.farmA, f.sec["H5.1"], dec31, f.breeder)

	type row struct {
		code, hall       string
		start, end       PerSex[int]
		mortality, fed   PerSex[bool]
		gF, gM, kgF, kgM string
	}
	no := PerSex[bool]{}
	for _, tc := range []struct {
		d    timex.Day
		want []row
	}{
		{jan(12), []row{
			{"H5.1", "H5", PerSex[int]{57, 0}, PerSex[int]{55, 0}, PerSex[bool]{F: true}, PerSex[bool]{F: true},
				"100", "nil", "5.7", "0"},
			// A migrated kg-only row: fed, no typed g, the typed kg. Mortality entered as 0 still counts as entered.
			{"H5.2", "H5", PerSex[int]{37, 0}, PerSex[int]{36, 0}, PerSex[bool]{F: true}, PerSex[bool]{F: true},
				"nil", "nil", "7.5", "0"},
			// The deleted ♀ mortality row is not an entered one.
			{"H6.1", "H6", PerSex[int]{8, 10}, PerSex[int]{8, 9}, PerSex[bool]{M: true}, PerSex[bool]{F: true, M: true},
				"120.5", "150", "0.964", "1.5"},
			{"H7", "H7", PerSex[int]{}, PerSex[int]{}, no, no, "nil", "nil", "0", "0"},
		}},
		{jan(13), []row{
			{"H5.1", "H5", PerSex[int]{55, 0}, PerSex[int]{44, 0}, PerSex[bool]{F: true}, PerSex[bool]{F: true},
				"100", "nil", "5.5", "0"},
			{"H5.2", "H5", PerSex[int]{36, 0}, PerSex[int]{36, 0}, no, no, "nil", "nil", "0", "0"}, // deleted allocation
			{"H6.1", "H6", PerSex[int]{8, 9}, PerSex[int]{8, 6}, no, PerSex[bool]{M: true}, "nil", "150", "0", "1.35"},
			{"H7", "H7", PerSex[int]{}, PerSex[int]{}, no, no, "nil", "nil", "0", "0"},
		}},
		{dec31, []row{
			{"H5.1", "H5", PerSex[int]{44, 0}, PerSex[int]{44, 0}, no, PerSex[bool]{F: true}, "100", "nil", "4.4", "0"},
			{"H5.2", "H5", PerSex[int]{36, 0}, PerSex[int]{36, 0}, no, no, "nil", "nil", "0", "0"},
			{"H6.1", "H6", PerSex[int]{8, 6}, PerSex[int]{8, 6}, no, no, "nil", "nil", "0", "0"},
			{"H7", "H7", PerSex[int]{}, PerSex[int]{}, no, no, "nil", "nil", "0", "0"},
		}},
	} {
		t.Run(tc.d.String(), func(t *testing.T) {
			got, err := SectionsOnDay(ctx, f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d rows %+v, want %d", len(got), got, len(tc.want))
			}
			sumKg := PerSex[decimal.Decimal]{}
			for i, w := range tc.want {
				g := got[i]
				if g.SectionID != f.sec[w.code] || g.Code != w.code || g.HallCode != w.hall || g.HallID != f.sec[w.hall] ||
					g.AliveStart != w.start || g.AliveEnd != w.end || g.MortalityEntered != w.mortality || g.Fed != w.fed ||
					decStr(g.G.F) != w.gF || decStr(g.G.M) != w.gM ||
					!g.Kg.F.Equal(decimal.RequireFromString(w.kgF)) || !g.Kg.M.Equal(decimal.RequireFromString(w.kgM)) {
					t.Errorf("row %d = %+v (g %s · %s)\nwant %+v", i, g, decStr(g.G.F), decStr(g.G.M), w)
				}
				sumKg.F, sumKg.M = sumKg.F.Add(g.Kg.F), sumKg.M.Add(g.Kg.M)
			}
			// The sections add up to the farm's FeedKg.
			farm, err := FeedKg(ctx, f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if !farm.F.Equal(sumKg.F) || !farm.M.Equal(sumKg.M) {
				t.Errorf("Σ sections kg ♀ %s · ♂ %s; farm ♀ %s · ♂ %s", sumKg.F, sumKg.M, farm.F, farm.M)
			}
		})
	}
	if got, err := SectionsOnDay(ctx, f.d.App, uuid.New(), jan(12)); err != nil || len(got) != 0 {
		t.Errorf("unknown farm: %+v, %v; want no rows", got, err)
	}
}

// emptying adds the removals that empty sections of farm A, after the fixture's 13.01:
//
//	14.01 sale H6.1 ♀ 8 · ♂ 6            → H6.1 and hall H6 empty
//	15.01 sale H5.2 ♀ 36                 → H5.2 empty; H5 keeps H5.1
//	16.01 move H5.1 → H6.1 ♀ 4           → H6.1 and H6 have birds again
//	17.01 sale H6.1 ♀ 4                  → empty again
//	20.01 mortality H6.1 ♀ entered as 0  → an event day that doesn't move the count
func (f *fixture) emptying() {
	f.t.Helper()
	A := f.farmA
	e := f.entry("sale", A, nil, jan(14))
	f.bird(e, A, "H6.1", jan(14), "F", "sale", -8, "")
	f.bird(e, A, "H6.1", jan(14), "M", "sale", -6, "")
	e = f.entry("sale", A, nil, jan(15))
	f.bird(e, A, "H5.2", jan(15), "F", "sale", -36, "")
	e = f.entry("transfer", A, &A, jan(16))
	f.bird(e, A, "H5.1", jan(16), "F", "transfer", -4, "H6.1")
	f.bird(e, A, "H6.1", jan(16), "F", "transfer", 4, "H5.1")
	e = f.entry("sale", A, nil, jan(17))
	f.bird(e, A, "H6.1", jan(17), "F", "sale", -4, "")
	e = f.entry("daily", A, nil, jan(20))
	f.bird(e, A, "H6.1", jan(20), "F", "mortality", 0, "")
}

func TestEmptySince(t *testing.T) {
	f := newFixture(t)
	f.emptying()
	ctx := t.Context()
	day := func(d timex.Day) *timex.Day { return &d }
	se := func(code string, hall bool, alive int, since *timex.Day) SectionEmpty {
		return SectionEmpty{SectionID: f.sec[code], Code: code, IsHall: hall, AliveEnd: alive, EmptySince: since}
	}
	equal := func(a, b SectionEmpty) bool {
		return a.SectionID == b.SectionID && a.Code == b.Code && a.IsHall == b.IsHall && a.AliveEnd == b.AliveEnd &&
			(a.EmptySince == nil) == (b.EmptySince == nil) && (a.EmptySince == nil || *a.EmptySince == *b.EmptySince)
	}
	str := func(rows []SectionEmpty) []string {
		var out []string
		for _, r := range rows {
			s := "–"
			if r.EmptySince != nil {
				s = r.EmptySince.String()
			}
			out = append(out, r.Code+" "+s)
		}
		return out
	}
	for _, tc := range []struct {
		name string
		d    timex.Day
		want []SectionEmpty
	}{
		// H7 never had birds: 0, but never emptied.
		{"every section has birds", jan(13), []SectionEmpty{
			se("H5", true, 80, nil), se("H5.1", false, 44, nil), se("H5.2", false, 36, nil),
			se("H6", true, 14, nil), se("H6.1", false, 14, nil), se("H7", false, 0, nil),
		}},
		{"the emptying day is the first empty day", jan(14), []SectionEmpty{
			se("H5", true, 80, nil), se("H5.1", false, 44, nil), se("H5.2", false, 36, nil),
			se("H6", true, 0, day(jan(14))), se("H6.1", false, 0, day(jan(14))), se("H7", false, 0, nil),
		}},
		{"a hall with one empty compartment is not empty", jan(15), []SectionEmpty{
			se("H5", true, 44, nil), se("H5.1", false, 44, nil), se("H5.2", false, 0, day(jan(15))),
			se("H6", true, 0, day(jan(14))), se("H6.1", false, 0, day(jan(14))), se("H7", false, 0, nil),
		}},
		{"birds again", jan(16), []SectionEmpty{
			se("H5", true, 40, nil), se("H5.1", false, 40, nil), se("H5.2", false, 0, day(jan(15))),
			se("H6", true, 4, nil), se("H6.1", false, 4, nil), se("H7", false, 0, nil),
		}},
		{"the last emptying wins; a 0 event day changes nothing", jan(20), []SectionEmpty{
			se("H5", true, 40, nil), se("H5.1", false, 40, nil), se("H5.2", false, 0, day(jan(15))),
			se("H6", true, 0, day(jan(17))), se("H6.1", false, 0, day(jan(17))), se("H7", false, 0, nil),
		}},
		{"after app.today: no dense-view cut-off", dec31, []SectionEmpty{
			se("H5", true, 40, nil), se("H5.1", false, 40, nil), se("H5.2", false, 0, day(jan(15))),
			se("H6", true, 0, day(jan(17))), se("H6.1", false, 0, day(jan(17))), se("H7", false, 0, nil),
		}},
		{"before any event", jan(1), []SectionEmpty{
			se("H5", true, 0, nil), se("H5.1", false, 0, nil), se("H5.2", false, 0, nil),
			se("H6", true, 0, nil), se("H6.1", false, 0, nil), se("H7", false, 0, nil),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EmptySince(ctx, f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.EqualFunc(got, tc.want, equal) {
				t.Errorf("got  %v %+v\nwant %v", str(got), got, str(tc.want))
			}
		})
	}

	// Farm B's section was emptied of ♀ but keeps ♂ 10: not empty.
	got, err := EmptySince(ctx, f.d.App, f.farmB, jan(10))
	if err != nil {
		t.Fatal(err)
	}
	if want := []SectionEmpty{se("T1", false, 10, nil)}; !slices.EqualFunc(got, want, equal) {
		t.Errorf("farm B = %v, want %v", str(got), str(want))
	}
	if got, err := EmptySince(ctx, f.d.App, uuid.New(), jan(10)); err != nil || len(got) != 0 {
		t.Errorf("unknown farm: %+v, %v; want no rows", got, err)
	}
}

func TestM2ReadersFailWithTheContext(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for name, call := range map[string]func() error{
		"ReadGuideWeeks":  func() error { _, err := ReadGuideWeeks(ctx, f.d.App, uuid.New(), 1, 2); return err },
		"ReadAppSettings": func() error { _, err := ReadAppSettings(ctx, f.d.App); return err },
		"CycleByID":       func() error { _, err := CycleByID(ctx, f.d.App, f.farmA, f.cycleA); return err },
		"FarmCycles":      func() error { _, err := FarmCycles(ctx, f.d.App, f.farmA); return err },
		"ReadLatestWeighings": func() error {
			_, err := ReadLatestWeighings(ctx, f.d.App, f.farmA, jan(1), jan(20))
			return err
		},
		"SectionsOnDay": func() error { _, err := SectionsOnDay(ctx, f.d.App, f.farmA, jan(12)); return err },
		"EmptySince":    func() error { _, err := EmptySince(ctx, f.d.App, f.farmA, jan(12)); return err },
	} {
		if err := call(); !errors.Is(err, context.Canceled) || errors.Is(err, ErrNoCycle) {
			t.Errorf("%s: err = %v, want context.Canceled", name, err)
		}
	}
}
