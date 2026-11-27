package calc

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

// fixture is a small hand-built ledger (all days in January 2026):
//
//	Farm B (growth): section T1. Placement 05.01 ♀ 100 · ♂ 20, cycle 05.01–10.01.
//	Farm A (laying): halls H5 (H5.1, H5.2) and H6 (H6.1), parentless leaf H7 (never used); cycle from 10.01, open.
//	10.01 transfer B → A: T1 → H5.1 ♀ 60, T1 → H6.1 ♂ 10 (peer sections set);
//	      T1 → H5.2 ♀ 40 in a transfer entry without peer sections (migration style);
//	      adjustment H5.1 ♀ +2 (migration entry).
//	11.01 move inside A: H5.1 → H6.1 ♀ 5 (transfer entry, peers set);
//	      H5.2 → H6.1 ♀ 3 in a migration entry without peer sections.
//	12.01 mortality H5.1 ♀ 2, H5.2 ♀ 0 (entered as zero), H6.1 ♂ 1; cull H5.2 ♀ 1;
//	      a superseded (deleted) mortality row H6.1 ♀ 50.
//	13.01 sale H5.1 ♀ 10, H6.1 ♂ 3; mortality H5.1 ♀ 1.
//
// Alive at the end of the day at A:
//
//	        H5.1 ♀  H5.2 ♀  H6.1 ♀  H6.1 ♂   farm ♀ · ♂
//	10.01     62      40       0      10      102 · 10
//	11.01     57      37       8      10      102 · 10
//	12.01     55      36       8       9       99 · 9
//	13.01     44      36       8       6       88 · 6
type fixture struct {
	t        *testing.T
	d        *testdb.DB
	series   uuid.UUID
	farmA    uuid.UUID
	farmB    uuid.UUID
	cycleA   uuid.UUID
	cycleB   uuid.UUID
	sec      map[string]uuid.UUID
	breeder  uuid.UUID
	roosters uuid.UUID
}

func jan(d int) timex.Day { return timex.MustDay(2026, time.January, d) }

func (f *fixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *fixture) section(farm uuid.UUID, code string, parent *uuid.UUID, sort int) uuid.UUID {
	f.t.Helper()
	id := f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, farm, code, parent, sort)
	f.sec[code] = id
	return id
}

func (f *fixture) entry(kind string, farm uuid.UUID, peerFarm *uuid.UUID, d timex.Day) uuid.UUID {
	f.t.Helper()
	var series *uuid.UUID
	if kind == "placement" {
		series = &f.series
	}
	source := "manual"
	if kind == "migration" {
		source = "migration"
	}
	return f.id(`insert into entry (kind, farm_id, peer_farm_id, date, series_id, source)
		values ($1, $2, $3, $4, $5, $6)`, kind, farm, peerFarm, d, series, source)
}

// bird inserts a live bird_event row; peer is a section code or "".
func (f *fixture) bird(entry, farm uuid.UUID, code string, d timex.Day, sex, typ string, qty int, peer string) {
	f.t.Helper()
	var peerID *uuid.UUID
	if peer != "" {
		p := f.sec[peer]
		peerID = &p
	}
	var reason *string
	if typ == "cull" {
		r := "boala"
		reason = &r
	}
	f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason, peer_section_id)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`, entry, farm, f.sec[code], d, sex, typ, qty, reason, peerID)
}

func (f *fixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria T', 'seria t', 2030, $1)`, jan(5))
	f.farmA = f.id(`insert into farm (name, name_key, type) values ('Ferma A', 'ferma a', 'laying')`)
	f.farmB = f.id(`insert into farm (name, name_key, type) values ('Ferma B', 'ferma b', 'growth')`)
	f.cycleB = f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, $3, $4, 'auto', 'emptied')`, f.series, f.farmB, jan(5), jan(10))
	f.cycleA = f.id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`,
		f.series, f.farmA, jan(10))

	f.section(f.farmB, "T1", nil, 1)
	h5 := f.section(f.farmA, "H5", nil, 1)
	h6 := f.section(f.farmA, "H6", nil, 2)
	f.section(f.farmA, "H7", nil, 3)
	f.section(f.farmA, "H6.1", &h6, 1)
	f.section(f.farmA, "H5.2", &h5, 2) // inserted before H5.1: order comes from sort_order, not insertion
	f.section(f.farmA, "H5.1", &h5, 1)

	A, B := f.farmA, f.farmB
	e := f.entry("placement", B, nil, jan(5))
	f.bird(e, B, "T1", jan(5), "F", "placement", 100, "")
	f.bird(e, B, "T1", jan(5), "M", "placement", 20, "")

	e = f.entry("transfer", B, &A, jan(10))
	f.bird(e, B, "T1", jan(10), "F", "transfer", -60, "H5.1")
	f.bird(e, A, "H5.1", jan(10), "F", "transfer", 60, "T1")
	f.bird(e, B, "T1", jan(10), "M", "transfer", -10, "H6.1")
	f.bird(e, A, "H6.1", jan(10), "M", "transfer", 10, "T1")
	e = f.entry("transfer", B, &A, jan(10))
	f.bird(e, B, "T1", jan(10), "F", "transfer", -40, "")
	f.bird(e, A, "H5.2", jan(10), "F", "transfer", 40, "")
	e = f.entry("migration", A, nil, jan(10))
	f.bird(e, A, "H5.1", jan(10), "F", "adjustment", 2, "")

	e = f.entry("transfer", A, &A, jan(11))
	f.bird(e, A, "H5.1", jan(11), "F", "transfer", -5, "H6.1")
	f.bird(e, A, "H6.1", jan(11), "F", "transfer", 5, "H5.1")
	e = f.entry("migration", A, nil, jan(11))
	f.bird(e, A, "H5.2", jan(11), "F", "transfer", -3, "")
	f.bird(e, A, "H6.1", jan(11), "F", "transfer", 3, "")

	e = f.entry("daily", A, nil, jan(12))
	f.bird(e, A, "H5.1", jan(12), "F", "mortality", -2, "")
	f.bird(e, A, "H5.2", jan(12), "F", "mortality", 0, "")
	f.bird(e, A, "H6.1", jan(12), "M", "mortality", -1, "")
	f.bird(e, A, "H5.2", jan(12), "F", "cull", -1, "")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, deleted_at)
		values ($1, $2, $3, $4, 'F', 'mortality', -50, $5)`, e, A, f.sec["H6.1"], jan(12), testdb.Now())

	e = f.entry("sale", A, nil, jan(13))
	f.bird(e, A, "H5.1", jan(13), "F", "sale", -10, "")
	f.bird(e, A, "H6.1", jan(13), "M", "sale", -3, "")
	e = f.entry("daily", A, nil, jan(13))
	f.bird(e, A, "H5.1", jan(13), "F", "mortality", -1, "")

	return f
}

// eggs adds the egg rows: hall-days with the four parts filled, one with a part missing (no total) and one
// superseded row.
//
//	12.01 H5: 10 + 20 + 3 + 30 = 63 eggs, hens at the start 57 + 37 = 94
//	12.01 H6: Cat. 3 missing → no total, not a hen-day
//	13.01 H5: 1 + 1 + 1 + 47 = 50 eggs, hens 55 + 36 = 91; H6: 4 eggs, hens 8
//	11.01 H5: deleted row
func (f *fixture) eggs() {
	f.t.Helper()
	ins := `insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, deleted_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	A := f.farmA
	e := f.entry("daily", A, nil, jan(11))
	f.exec(ins, e, A, f.sec["H5"], jan(11), 0, 500, 500, 500, 500, testdb.Now())
	e = f.entry("daily", A, nil, jan(12))
	f.exec(ins, e, A, f.sec["H5"], jan(12), 4, 10, 20, 3, 30, nil)
	f.exec(ins, e, A, f.sec["H6"], jan(12), 1, 1, 1, nil, 5, nil)
	e = f.entry("daily", A, nil, jan(13))
	f.exec(ins, e, A, f.sec["H5"], jan(13), 0, 1, 1, 1, 47, nil)
	f.exec(ins, e, A, f.sec["H6"], jan(13), 0, 0, 0, 0, 4, nil)
}

// feed adds feed types, allocations, deliveries and a stock adjustment.
//
//	12.01 H5.1 ♀ 100 g × 57 = 5,7 kg; H5.2 ♀ typed 7,5 kg (no g); H6.1 ♀ 120,5 g × 8 = 0,964 kg; H6.1 ♂ 150 g × 10 = 1,5 kg
//	13.01 H5.1 ♀ 100 g × 55 = 5,5 kg; H6.1 ♂ 150 g × 9 = 1,35 kg; a deleted H5.2 ♀ row
//	Breeder T at A: +1 000 (10.01), −300 adjusted (11.01), −14,164 (12.01), +500 − 5,5 (13.01)
//	Cocoși T at A: +200 (10.01), −1,5 (12.01), −1,35 (13.01); Breeder T at B: +50 (10.01)
func (f *fixture) feed() {
	f.t.Helper()
	f.breeder = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder T', 'breeder t', 'F')`)
	f.roosters = f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși T', 'cocosi t', 'M')`)
	A, B := f.farmA, f.farmB
	alloc := `insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird, kg_recorded, deleted_at)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	e := f.entry("feed_daily", A, nil, jan(12))
	f.exec(alloc, e, A, f.sec["H5.1"], jan(12), "F", f.breeder, "100", nil, nil)
	f.exec(alloc, e, A, f.sec["H5.2"], jan(12), "F", f.breeder, nil, "7.5", nil)
	f.exec(alloc, e, A, f.sec["H6.1"], jan(12), "F", f.breeder, "120.5", nil, nil)
	f.exec(alloc, e, A, f.sec["H6.1"], jan(12), "M", f.roosters, "150", nil, nil)
	e = f.entry("feed_daily", A, nil, jan(13))
	f.exec(alloc, e, A, f.sec["H5.1"], jan(13), "F", f.breeder, "100", nil, nil)
	f.exec(alloc, e, A, f.sec["H6.1"], jan(13), "M", f.roosters, "150", nil, nil)
	f.exec(alloc, e, A, f.sec["H5.2"], jan(13), "F", f.breeder, "300", nil, testdb.Now())

	deliver := func(farm, ft uuid.UUID, d timex.Day, kg string) {
		e := f.entry("delivery", farm, nil, d)
		f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`,
			e, farm, ft, d, kg)
	}
	deliver(A, f.breeder, jan(10), "1000")
	deliver(A, f.roosters, jan(10), "200")
	deliver(A, f.breeder, jan(13), "500")
	deliver(B, f.breeder, jan(10), "50")
	e = f.entry("migration", A, nil, jan(11))
	f.exec(`insert into feed_stock_adjustment (entry_id, farm_id, feed_type_id, date, qty_kg, note)
		values ($1, $2, $3, $4, -300, 'ajustare')`, e, A, f.breeder, jan(11))
}

func TestReadFarmDay(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		d    timex.Day
		want FarmDay
	}{
		{"deaths and culls", f.farmA, jan(12), FarmDay{
			AliveStart: PerSex[int]{102, 10}, AliveEnd: PerSex[int]{99, 9},
			Deaths: PerSex[int]{2, 1}, Culls: PerSex[int]{1, 0}}},
		{"sales are not deaths", f.farmA, jan(13), FarmDay{
			AliveStart: PerSex[int]{99, 9}, AliveEnd: PerSex[int]{88, 6},
			Deaths: PerSex[int]{1, 0}, Sold: PerSex[int]{10, 3}}},
		{"moves inside the farm net to 0", f.farmA, jan(11), FarmDay{
			AliveStart: PerSex[int]{102, 10}, AliveEnd: PerSex[int]{102, 10}}},
		{"before the first event", f.farmA, jan(9), FarmDay{}},
		{"after app.today: no dense-view cut-off", f.farmA, timex.MustDay(2026, time.December, 31), FarmDay{
			AliveStart: PerSex[int]{88, 6}, AliveEnd: PerSex[int]{88, 6}}},
		{"source farm of the transfer", f.farmB, jan(10), FarmDay{
			AliveStart: PerSex[int]{100, 20}, AliveEnd: PerSex[int]{0, 10}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ReadFarmDay(t.Context(), f.d.App, tc.farm, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %+v\nwant %+v", got, tc.want)
			}
		})
	}
}

func TestSectionsAliveEnd(t *testing.T) {
	f := newFixture(t)
	codes := func(rows []SectionAlive) []string {
		var out []string
		for _, r := range rows {
			out = append(out, r.Code)
		}
		return out
	}
	for _, tc := range []struct {
		d    timex.Day
		want []SectionAlive
	}{
		{jan(12), []SectionAlive{
			{f.sec["H5.1"], "H5.1", PerSex[int]{55, 0}},
			{f.sec["H5.2"], "H5.2", PerSex[int]{36, 0}},
			{f.sec["H6.1"], "H6.1", PerSex[int]{8, 9}},
			{f.sec["H7"], "H7", PerSex[int]{0, 0}},
		}},
		{jan(13), []SectionAlive{
			{f.sec["H5.1"], "H5.1", PerSex[int]{44, 0}},
			{f.sec["H5.2"], "H5.2", PerSex[int]{36, 0}},
			{f.sec["H6.1"], "H6.1", PerSex[int]{8, 6}},
			{f.sec["H7"], "H7", PerSex[int]{0, 0}},
		}},
		{jan(1), []SectionAlive{
			{f.sec["H5.1"], "H5.1", PerSex[int]{}},
			{f.sec["H5.2"], "H5.2", PerSex[int]{}},
			{f.sec["H6.1"], "H6.1", PerSex[int]{}},
			{f.sec["H7"], "H7", PerSex[int]{}},
		}},
	} {
		t.Run(tc.d.String(), func(t *testing.T) {
			got, err := SectionsAliveEnd(t.Context(), f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("got %v %+v\nwant %v %+v", codes(got), got, codes(tc.want), tc.want)
			}
		})
	}
	got, err := SectionsAliveEnd(t.Context(), f.d.App, uuid.New(), jan(12))
	if err != nil || len(got) != 0 {
		t.Errorf("unknown farm: %+v, %v; want no rows", got, err)
	}
}

func TestCycleDeathsAndSold(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	deaths, err := CycleDeaths(ctx, f.d.App, f.farmA, jan(10), jan(12))
	if err != nil {
		t.Fatal(err)
	}
	// mortality ♀ 2 + 0 and cull ♀ 1; mortality ♂ 1. The deleted ♀ 50 never counts.
	if want := (PerSex[int]{3, 1}); deaths != want {
		t.Errorf("deaths 10–12.01 = %+v, want %+v", deaths, want)
	}
	deaths, err = CycleDeaths(ctx, f.d.App, f.farmA, jan(10), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	if want := (PerSex[int]{4, 1}); deaths != want {
		t.Errorf("deaths 10–13.01 = %+v, want %+v (sales are not deaths)", deaths, want)
	}
	deaths, err = CycleDeaths(ctx, f.d.App, f.farmA, jan(13), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	if want := (PerSex[int]{1, 0}); deaths != want {
		t.Errorf("deaths on 13.01 = %+v, want %+v", deaths, want)
	}

	for _, tc := range []struct {
		name     string
		from, to timex.Day
		want     PerSex[int]
	}{
		{"range with the sale", jan(12), jan(13), PerSex[int]{10, 3}},
		{"range without", jan(10), jan(12), PerSex[int]{}},
		{"one day", jan(13), jan(13), PerSex[int]{10, 3}},
	} {
		got, err := Sold(ctx, f.d.App, f.farmA, tc.from, tc.to)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("%s: sold = %+v, want %+v", tc.name, got, tc.want)
		}
	}
}

func TestFarmBase(t *testing.T) {
	f := newFixture(t)
	for _, tc := range []struct {
		name     string
		farm     uuid.UUID
		from, to timex.Day
		want     PerSex[int]
	}{
		// ♀ 60 (peer in B) + 40 (no peer, transfer entry from B); ♂ 10. Not the +2 adjustment, the internal moves
		// (+5 with peers in A, +3 without peers in a migration entry) or anything after the sale.
		{"arrivals from another farm", f.farmA, jan(10), jan(13), PerSex[int]{100, 10}},
		{"the range starts after the arrivals", f.farmA, jan(11), jan(13), PerSex[int]{}},
		{"placement; transfers out never count", f.farmB, jan(5), jan(10), PerSex[int]{100, 20}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FarmBase(t.Context(), f.d.App, tc.farm, tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("base = %+v, want %+v", got, tc.want)
			}
		})
	}

	// §5.5 on the fixture: 3 ÷ 100 = 3 % ♀, 1 ÷ 10 = 10 % ♂ to 12.01.
	ctx := t.Context()
	base, err := FarmBase(ctx, f.d.App, f.farmA, jan(10), jan(12))
	if err != nil {
		t.Fatal(err)
	}
	deaths, err := CycleDeaths(ctx, f.d.App, f.farmA, jan(10), jan(12))
	if err != nil {
		t.Fatal(err)
	}
	if got := rounded(t, CumMortalityPct(deaths.F, base.F), 2); got != "3" {
		t.Errorf("♀ cum. mortality = %s, want 3", got)
	}
	if got := rounded(t, CumMortalityPct(deaths.M, base.M), 2); got != "10" {
		t.Errorf("♂ cum. mortality = %s, want 10", got)
	}
}

// dec31 is a day after app.today (18.11.2026) with no bird event: the dense
// view v_section_sex_day has no row for it.
var dec31 = timex.MustDay(2026, time.December, 31)

func TestLayDay(t *testing.T) {
	f := newFixture(t)
	f.eggs()
	// 31.12 H5: 1 + 1 + 1 + 37 = 40 eggs; hens at the start = H5.1 44 + H5.2 36 (end of 13.01, nothing since).
	e := f.entry("daily", f.farmA, nil, dec31)
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, cat1, cat2, cat3, hatching_eggs)
		values ($1, $2, $3, $4, 1, 1, 1, 37)`, e, f.farmA, f.sec["H5"], dec31)
	// The dense view stops at app.today, so v_egg_day has no hens that day: the reader must not read it.
	var viewHens *int
	if err := f.d.App.QueryRow(t.Context(), `select hens_start from v_egg_day where farm_id = $1 and date = $2`,
		f.farmA, dec31).Scan(&viewHens); err != nil || viewHens != nil {
		t.Fatalf("v_egg_day hens on 31.12 = %v (%v); the test expects the view's cut-off (null)", viewHens, err)
	}
	for _, tc := range []struct {
		name string
		d    timex.Day
		want Lay
		pct  string // at 1 dp; "" = nil
	}{
		{"a unit without a total is not a hen-day", jan(12), Lay{63, 94}, "67"},
		{"two halls", jan(13), Lay{54, 99}, "54.5"},
		{"deleted row only", jan(11), Lay{}, ""},
		{"no eggs entered", jan(14), Lay{}, ""},
		{"after app.today: no dense-view cut-off", dec31, Lay{40, 80}, "50"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := LayDay(t.Context(), f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("lay = %+v, want %+v", got, tc.want)
			}
			if tc.pct == "" {
				if p := got.Pct(); p != nil {
					t.Errorf("pct = %s, want nil", p)
				}
			} else if p := rounded(t, got.Pct(), 1); p != tc.pct {
				t.Errorf("pct = %s, want %s", p, tc.pct)
			}
		})
	}
}

func TestFeedKg(t *testing.T) {
	f := newFixture(t)
	f.feed()
	// 31.12: H5.1 ♀ 100 g × 44 = 4,4 kg; H6.1 ♂ 150 g × 6 = 0,9 kg (counts at the end of 13.01, nothing since).
	e := f.entry("feed_daily", f.farmA, nil, dec31)
	alloc := `insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, $5, $6, $7)`
	f.exec(alloc, e, f.farmA, f.sec["H5.1"], dec31, "F", f.breeder, "100")
	f.exec(alloc, e, f.farmA, f.sec["H6.1"], dec31, "M", f.roosters, "150")
	// The dense view stops at app.today, so v_feed_consumption has 0 birds that day: the reader must not read it.
	var viewBirds int
	if err := f.d.App.QueryRow(t.Context(), `select coalesce(sum(birds_start), 0) from v_feed_consumption
		where farm_id = $1 and date = $2`, f.farmA, dec31).Scan(&viewBirds); err != nil || viewBirds != 0 {
		t.Fatalf("v_feed_consumption birds on 31.12 = %d (%v); the test expects the view's cut-off (0)", viewBirds, err)
	}
	for _, tc := range []struct {
		d    timex.Day
		f, m string
	}{
		{jan(12), "14.164", "1.5"}, // 5,7 + 7,5 (typed) + 0,964
		{jan(13), "5.5", "1.35"},   // the deleted row doesn't count
		{jan(14), "0", "0"},
		{dec31, "4.4", "0.9"}, // after app.today: no dense-view cut-off
	} {
		t.Run(tc.d.String(), func(t *testing.T) {
			got, err := FeedKg(t.Context(), f.d.App, f.farmA, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if !got.F.Equal(decimal.RequireFromString(tc.f)) || !got.M.Equal(decimal.RequireFromString(tc.m)) {
				t.Errorf("kg = ♀ %s · ♂ %s, want ♀ %s · ♂ %s", got.F, got.M, tc.f, tc.m)
			}
		})
	}
}

func TestStockEnd(t *testing.T) {
	f := newFixture(t)
	f.feed()
	type want struct {
		name, kg string
		last     timex.Day
	}
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		d    timex.Day
		want []want
	}{
		{"after an adjustment and a consumption day", f.farmA, jan(12),
			[]want{{"Breeder T", "685.836", jan(12)}, {"Cocoși T", "198.5", jan(12)}}},
		{"latest movement day on or before", f.farmA, jan(20),
			[]want{{"Breeder T", "1180.336", jan(13)}, {"Cocoși T", "197.15", jan(13)}}},
		{"the adjustment day", f.farmA, jan(11),
			[]want{{"Breeder T", "700", jan(11)}, {"Cocoși T", "200", jan(10)}}},
		{"before any movement", f.farmA, jan(9), nil},
		{"another farm's stock is its own", f.farmB, jan(12), []want{{"Breeder T", "50", jan(10)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StockEnd(t.Context(), f.d.App, tc.farm, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			for i, w := range tc.want {
				g := got[i]
				if g.Name != w.name || !g.Kg.Equal(decimal.RequireFromString(w.kg)) || g.LastMove != w.last {
					t.Errorf("row %d = %s %s kg (last %s), want %s %s kg (last %s)", i, g.Name, g.Kg, g.LastMove, w.name, w.kg, w.last)
				}
			}
		})
	}
}

func TestCycleAt(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	end := jan(10)
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		d    timex.Day
		want Cycle
	}{
		{"open cycle", f.farmA, jan(12), Cycle{ID: f.cycleA, SeriesID: f.series, StartDate: jan(10),
			Age: SeriesAge{PlacementDate: jan(5)}}},
		{"open cycle long after", f.farmA, timex.MustDay(2027, time.June, 1), Cycle{ID: f.cycleA, SeriesID: f.series,
			StartDate: jan(10), Age: SeriesAge{PlacementDate: jan(5)}}},
		{"closed cycle, end day inclusive", f.farmB, jan(10), Cycle{ID: f.cycleB, SeriesID: f.series,
			StartDate: jan(5), EndDate: &end, Age: SeriesAge{PlacementDate: jan(5)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CycleAt(ctx, f.d.App, tc.farm, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != tc.want.ID || got.SeriesID != tc.want.SeriesID || got.StartDate != tc.want.StartDate ||
				got.Age != tc.want.Age || (got.EndDate == nil) != (tc.want.EndDate == nil) ||
				(got.EndDate != nil && *got.EndDate != *tc.want.EndDate) {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		d    timex.Day
	}{
		{"before the first cycle", f.farmA, jan(9)},
		{"after a closed cycle", f.farmB, jan(11)},
	} {
		if _, err := CycleAt(ctx, f.d.App, tc.farm, tc.d); !errors.Is(err, ErrNoCycle) {
			t.Errorf("%s: err = %v, want ErrNoCycle", tc.name, err)
		}
	}
}

func TestReadSeriesAge(t *testing.T) {
	f := newFixture(t)
	got, err := ReadSeriesAge(t.Context(), f.d.App, f.series)
	if err != nil {
		t.Fatal(err)
	}
	if want := (SeriesAge{PlacementDate: jan(5)}); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if _, err := ReadSeriesAge(t.Context(), f.d.App, uuid.New()); !errors.Is(err, ErrNoSeries) {
		t.Errorf("unknown series: err = %v, want ErrNoSeries", err)
	}
}

func TestReadersFailWithTheContext(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ReadFarmDay(ctx, f.d.App, f.farmA, jan(12)); !errors.Is(err, context.Canceled) {
		t.Errorf("ReadFarmDay: err = %v, want context.Canceled", err)
	}
	if _, err := StockEnd(ctx, f.d.App, f.farmA, jan(12)); !errors.Is(err, context.Canceled) {
		t.Errorf("StockEnd: err = %v, want context.Canceled", err)
	}
}

// Criterion 3 (T3): SQL age_days()/sapt() and calc agree on the golden 1–3 dates (and the edges around them).
func TestAgeMatchesSQL(t *testing.T) {
	d := testdb.New(t)
	ctx := t.Context()
	for _, ageAt := range []int{0, 7} {
		var id uuid.UUID
		if err := d.App.QueryRow(ctx, `insert into series (name, name_key, year, placement_date, age_at_placement_days)
			values ($1, $1, $2, '2025-10-07', $3) returning id`, "Seria 2026 +"+string(rune('0'+ageAt)), 2026+ageAt, ageAt).Scan(&id); err != nil {
			t.Fatal(err)
		}
		s, err := ReadSeriesAge(ctx, d.App, id)
		if err != nil {
			t.Fatal(err)
		}
		for _, dd := range []timex.Day{
			timex.MustDay(2026, time.November, 18), // golden 1
			timex.MustDay(2026, time.February, 24), // golden 2
			timex.MustDay(2026, time.February, 22), // golden 2
			timex.MustDay(2025, time.October, 20),  // golden 3
			timex.MustDay(2025, time.October, 21),  // golden 3
			timex.MustDay(2025, time.October, 7),
			timex.MustDay(2025, time.October, 6),
			timex.MustDay(2025, time.September, 29),
		} {
			var age, sapt int
			if err := d.App.QueryRow(ctx, `select age_days($1, $2), sapt($1, $2)`, id, dd).Scan(&age, &sapt); err != nil {
				t.Fatal(err)
			}
			if got := AgeDays(s, dd); got != age {
				t.Errorf("age %d, %s: calc age %d, SQL %d", ageAt, dd, got, age)
			}
			if got := SAPT(s, dd); got != sapt {
				t.Errorf("age %d, %s: calc SAPT %d, SQL %d", ageAt, dd, got, sapt)
			}
		}
	}
	// The golden values themselves, through the SQL side.
	var age, sapt int
	if err := d.App.QueryRow(ctx, `select age_days(id, '2026-11-18'), sapt(id, '2026-11-18') from series where year = 2026`).
		Scan(&age, &sapt); err != nil {
		t.Fatal(err)
	}
	if age != 407 || sapt != 59 {
		t.Errorf("SQL golden 1 = %d / %d, want 407 / 59", age, sapt)
	}
}

func TestSectionsRemovals(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	sr := func(code string, mortality, culls, sold PerSex[int]) SectionRemovals {
		return SectionRemovals{SectionID: f.sec[code], Code: code, Mortality: mortality, Culls: culls, Sold: sold}
	}
	none := PerSex[int]{}
	for _, tc := range []struct {
		name     string
		from, to timex.Day
		want     []SectionRemovals
	}{
		// The deleted H6.1 ♀ 50 never counts; H5.2 ♀ mortality was entered as 0.
		{"two days", jan(12), jan(13), []SectionRemovals{
			sr("H5.1", PerSex[int]{3, 0}, none, PerSex[int]{10, 0}),
			sr("H5.2", none, PerSex[int]{1, 0}, none),
			sr("H6.1", PerSex[int]{0, 1}, none, PerSex[int]{0, 3}),
			sr("H7", none, none, none),
		}},
		{"one day: sales are not deaths", jan(13), jan(13), []SectionRemovals{
			sr("H5.1", PerSex[int]{1, 0}, none, PerSex[int]{10, 0}),
			sr("H5.2", none, none, none),
			sr("H6.1", none, none, PerSex[int]{0, 3}),
			sr("H7", none, none, none),
		}},
		{"transfers and adjustments are not removals", jan(10), jan(11), []SectionRemovals{
			sr("H5.1", none, none, none),
			sr("H5.2", none, none, none),
			sr("H6.1", none, none, none),
			sr("H7", none, none, none),
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SectionsRemovals(ctx, f.d.App, f.farmA, tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("got  %+v\nwant %+v", got, tc.want)
			}
		})
	}

	// The sections add up to the farm figures.
	rows, err := SectionsRemovals(ctx, f.d.App, f.farmA, jan(10), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	var deaths, sold PerSex[int]
	for _, r := range rows {
		d := r.Deaths()
		deaths.F, deaths.M = deaths.F+d.F, deaths.M+d.M
		sold.F, sold.M = sold.F+r.Sold.F, sold.M+r.Sold.M
	}
	farmDeaths, err := CycleDeaths(ctx, f.d.App, f.farmA, jan(10), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	farmSold, err := Sold(ctx, f.d.App, f.farmA, jan(10), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	if deaths != farmDeaths || sold != farmSold {
		t.Errorf("Σ sections deaths %+v sold %+v; farm deaths %+v sold %+v", deaths, sold, farmDeaths, farmSold)
	}
	if want := (PerSex[int]{4, 1}); deaths != want {
		t.Errorf("Σ deaths = %+v, want %+v", deaths, want)
	}
}
