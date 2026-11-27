package calc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// alloc inserts a live allocation of g/pasăre for a section × sex on d, in its own feed_daily entry.
func (f *fixture) alloc(d timex.Day, code, sex string, ft uuid.UUID, g string) {
	f.t.Helper()
	e := f.entry("feed_daily", f.farmA, nil, d)
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, $5, $6, $7)`, e, f.farmA, f.sec[code], d, sex, ft, g)
}

// dec31Feed adds the TestFeedKg allocations of 31.12.2026, after app.today: H5.1 ♀ 100 g × 44 = 4,4 kg (Breeder T),
// H6.1 ♂ 150 g × 6 = 0,9 kg (Cocoși T).
func (f *fixture) dec31Feed() {
	f.t.Helper()
	f.alloc(dec31, "H5.1", "F", f.breeder, "100")
	f.alloc(dec31, "H6.1", "M", f.roosters, "150")
}

func TestStockAt(t *testing.T) {
	f := newFixture(t)
	f.feed()
	f.dec31Feed()
	// A delivery whose entry is deleted never counts.
	e := f.entry("delivery", f.farmA, nil, jan(12))
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 999)`,
		e, f.farmA, f.breeder, jan(12))
	user := f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, e, testdb.Now(), user)
	type want struct {
		name, kg string
		forSex   *dbq.Sex
		last     timex.Day
	}
	F, M := dbq.SexF, dbq.SexM
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		d    timex.Day
		want []want
	}{
		{"after an adjustment and a consumption day", f.farmA, jan(12),
			[]want{{"Breeder T", "685.836", &F, jan(12)}, {"Cocoși T", "198.5", &M, jan(12)}}},
		{"latest movement day on or before", f.farmA, jan(20),
			[]want{{"Breeder T", "1180.336", &F, jan(13)}, {"Cocoși T", "197.15", &M, jan(13)}}},
		{"the adjustment day", f.farmA, jan(11),
			[]want{{"Breeder T", "700", &F, jan(11)}, {"Cocoși T", "200", &M, jan(10)}}},
		{"before any movement", f.farmA, jan(9), nil},
		{"another farm's stock is its own", f.farmB, jan(12), []want{{"Breeder T", "50", &F, jan(10)}}},
		// 1 180,336 − 4,4 and 197,15 − 0,9: the allocations count their real kg (v_feed_stock would count 0).
		{"after app.today", f.farmA, dec31,
			[]want{{"Breeder T", "1175.936", &F, dec31}, {"Cocoși T", "196.25", &M, dec31}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StockAt(t.Context(), f.d.App, tc.farm, tc.d)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
			for i, w := range tc.want {
				g := got[i]
				if g.Name != w.name || !g.Kg.Equal(decimal.RequireFromString(w.kg)) || g.LastMove != w.last ||
					g.ForSex == nil || *g.ForSex != *w.forSex {
					t.Errorf("row %d = %s %s kg (last %s, for %v), want %s %s kg (last %s, for %s)",
						i, g.Name, g.Kg, g.LastMove, g.ForSex, w.name, w.kg, w.last, *w.forSex)
				}
			}
		})
	}

	// The dense view stops at app.today: v_feed_stock counts 0 kg for 31.12, so the case above tells the two apart.
	var viewKg decimal.Decimal
	if err := f.d.App.QueryRow(t.Context(), `select stock_end from v_feed_stock
		where farm_id = $1 and feed_type_id = $2 and date = $3`, f.farmA, f.breeder, dec31).Scan(&viewKg); err != nil ||
		!viewKg.Equal(decimal.RequireFromString("1180.336")) {
		t.Fatalf("v_feed_stock Breeder T on 31.12 = %s (%v); the test expects the view's cut-off (1180.336)", viewKg, err)
	}

	// Up to app.today StockAt agrees with v_feed_stock (the §5.8 definition it replaces) on every day.
	for d := jan(9); !d.After(jan(20)); d = d.AddDays(1) {
		got, err := StockAt(t.Context(), f.d.App, f.farmA, d)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := f.d.App.Query(t.Context(), `select distinct on (feed_type_id) feed_type_id, date, stock_end
			from v_feed_stock where farm_id = $1 and date <= $2 order by feed_type_id, date desc`, f.farmA, d)
		if err != nil {
			t.Fatal(err)
		}
		view := map[uuid.UUID]Stock{}
		for rows.Next() {
			var s Stock
			if err := rows.Scan(&s.FeedTypeID, &s.LastMove, &s.Kg); err != nil {
				t.Fatal(err)
			}
			view[s.FeedTypeID] = s
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if len(got) != len(view) {
			t.Fatalf("%s: StockAt %+v, v_feed_stock %+v", d, got, view)
		}
		for _, g := range got {
			v := view[g.FeedTypeID]
			if !g.Kg.Equal(v.Kg) || g.LastMove != v.LastMove {
				t.Errorf("%s %s: StockAt %s (last %s), v_feed_stock %s (last %s)", d, g.Name, g.Kg, g.LastMove, v.Kg, v.LastMove)
			}
		}
	}
}

func TestStockSeries(t *testing.T) {
	f := newFixture(t)
	f.feed()
	f.dec31Feed()
	unused := f.id(`insert into feed_type (name, name_key) values ('Fără mișcări', 'fara miscari')`)
	type pt struct {
		d  timex.Day
		kg string
	}
	for _, tc := range []struct {
		name     string
		types    []uuid.UUID
		from, to timex.Day
		want     map[uuid.UUID][]pt
	}{
		{"from's stock, then every movement day", []uuid.UUID{f.breeder, f.roosters}, jan(11), jan(13), map[uuid.UUID][]pt{
			f.breeder:  {{jan(11), "700"}, {jan(12), "685.836"}, {jan(13), "1180.336"}},
			f.roosters: {{jan(11), "200"}, {jan(12), "198.5"}, {jan(13), "197.15"}},
		}},
		{"before any movement: 0, then the farm's own deliveries", []uuid.UUID{f.breeder}, jan(9), jan(10), map[uuid.UUID][]pt{
			f.breeder: {{jan(9), "0"}, {jan(10), "1000"}},
		}},
		{"one day", []uuid.UUID{f.roosters}, jan(12), jan(12), map[uuid.UUID][]pt{
			f.roosters: {{jan(12), "198.5"}},
		}},
		{"a type without movements, asked twice", []uuid.UUID{unused, unused}, jan(11), jan(13), map[uuid.UUID][]pt{
			unused: {{jan(11), "0"}},
		}},
		{"after app.today", []uuid.UUID{f.breeder, f.roosters}, jan(20), dec31, map[uuid.UUID][]pt{
			f.breeder:  {{jan(20), "1180.336"}, {dec31, "1175.936"}},
			f.roosters: {{jan(20), "197.15"}, {dec31, "196.25"}},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := StockSeries(t.Context(), f.d.App, f.farmA, tc.types, tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d types (%+v), want %d", len(got), got, len(tc.want))
			}
			for id, want := range tc.want {
				g := got[id]
				if len(g) != len(want) {
					t.Fatalf("type %s: got %+v, want %+v", id, g, want)
				}
				for i, w := range want {
					if g[i].Day != w.d || !g[i].Kg.Equal(decimal.RequireFromString(w.kg)) {
						t.Errorf("type %s point %d = %s %s, want %s %s", id, i, g[i].Day, g[i].Kg, w.d, w.kg)
					}
				}
			}
		})
	}

	// H5 takes the lowest point: Breeder T over 11–13.01 is lowest on 12.01.
	s, err := StockSeries(t.Context(), f.d.App, f.farmA, []uuid.UUID{f.breeder}, jan(11), jan(13))
	if err != nil {
		t.Fatal(err)
	}
	if low, ok := Lowest(s[f.breeder]); !ok || low.Day != jan(12) || !low.Kg.Equal(decimal.RequireFromString("685.836")) {
		t.Errorf("Lowest = %+v, %v; want 685.836 on 12.01", low, ok)
	}
	if _, err := StockSeries(t.Context(), f.d.App, f.farmA, []uuid.UUID{f.breeder}, jan(13), jan(12)); !errors.Is(err, errRange) {
		t.Errorf("from after to: err = %v, want errRange", err)
	}
}

// T1 criterion 3. On 14.01 farm A has, at the start of the day, H5.1 ♀ 44, H5.2 ♀ 36, H6.1 ♀ 8 and H6.1 ♂ 6.
func TestFeedWindow(t *testing.T) {
	all := func(f *fixture, d timex.Day) {
		f.alloc(d, "H5.1", "F", f.breeder, "100")
		f.alloc(d, "H5.2", "F", f.breeder, "100")
		f.alloc(d, "H6.1", "F", f.breeder, "100")
		f.alloc(d, "H6.1", "M", f.roosters, "150")
	}
	for _, tc := range []struct {
		name  string
		setup func(f *fixture)
		today timex.Day
		want  Window
	}{
		{"today complete", func(f *fixture) { all(f, jan(14)) }, jan(14), Window{jan(12), jan(14)}},
		{"one section × sex missing today", func(f *fixture) {
			f.alloc(jan(14), "H5.1", "F", f.breeder, "100")
			f.alloc(jan(14), "H5.2", "F", f.breeder, "100")
			f.alloc(jan(14), "H6.1", "F", f.breeder, "100")
		}, jan(14), Window{jan(11), jan(13)}},
		{"a deleted allocation is missing", func(f *fixture) {
			all(f, jan(14))
			f.exec(`update feed_allocation set deleted_at = $2 where section_id = $1 and sex = 'M' and date = $3`,
				f.sec["H6.1"], testdb.Now(), jan(14))
		}, jan(14), Window{jan(11), jan(13)}},
		{"a section sold out today is not missing", func(f *fixture) {
			f.alloc(jan(14), "H5.1", "F", f.breeder, "100")
			f.alloc(jan(14), "H5.2", "F", f.breeder, "100")
			f.alloc(jan(14), "H6.1", "F", f.breeder, "100")
			e := f.entry("sale", f.farmA, nil, jan(14))
			f.bird(e, f.farmA, "H6.1", jan(14), "M", "sale", -6, "")
		}, jan(14), Window{jan(12), jan(14)}},
		{"after app.today: complete", func(f *fixture) { all(f, dec31) }, dec31,
			Window{timex.MustDay(2026, time.December, 29), dec31}},
		{"after app.today: missing", func(f *fixture) { f.dec31Feed() }, dec31,
			Window{timex.MustDay(2026, time.December, 28), timex.MustDay(2026, time.December, 30)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			f.feed()
			tc.setup(f)
			got, err := FeedWindow(t.Context(), f.d.App, f.farmA, tc.today, 3)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("FeedWindow = %s–%s, want %s–%s", got.From, got.To, tc.want.From, tc.want.To)
			}
		})
	}

	f := newFixture(t)
	ctx := t.Context()
	// Farm B still has T1 ♂ 10 on 14.01 and no allocation: the window ends the day before.
	if got, err := FeedWindow(ctx, f.d.App, f.farmB, jan(14), 3); err != nil || got != (Window{jan(11), jan(13)}) {
		t.Errorf("farm B: %+v, %v; want 11–13.01", got, err)
	}
	// A farm without birds ends the window on today.
	if got, err := FeedWindow(ctx, f.d.App, uuid.New(), jan(14), 3); err != nil || got != (Window{jan(12), jan(14)}) {
		t.Errorf("no birds: %+v, %v; want 12–14.01", got, err)
	}
	if got, err := FeedWindow(ctx, f.d.App, uuid.New(), jan(14), 1); err != nil || got != (Window{jan(14), jan(14)}) {
		t.Errorf("one day: %+v, %v; want 14–14.01", got, err)
	}
	if _, err := FeedWindow(ctx, f.d.App, f.farmA, jan(14), 0); !errors.Is(err, errWindowDays) {
		t.Errorf("0 days: err = %v, want errWindowDays", err)
	}
}

func TestConsumption(t *testing.T) {
	f := newFixture(t)
	f.feed()
	f.dec31Feed()
	for _, tc := range []struct {
		name     string
		farm     uuid.UUID
		from, to timex.Day
		want     map[uuid.UUID]string
	}{
		// 5,7 + 7,5 (typed) + 0,964 + 5,5; 1,5 + 1,35. The deleted 13.01 row doesn't count.
		{"two days", f.farmA, jan(12), jan(13), map[uuid.UUID]string{f.breeder: "19.664", f.roosters: "2.85"}},
		{"one day", f.farmA, jan(12), jan(12), map[uuid.UUID]string{f.breeder: "14.164", f.roosters: "1.5"}},
		{"no allocation", f.farmA, jan(14), jan(20), map[uuid.UUID]string{}},
		{"after app.today", f.farmA, dec31, dec31, map[uuid.UUID]string{f.breeder: "4.4", f.roosters: "0.9"}},
		{"another farm", f.farmB, jan(1), dec31, map[uuid.UUID]string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Consumption(t.Context(), f.d.App, tc.farm, tc.from, tc.to)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for id, kg := range tc.want {
				if g, ok := got[id]; !ok || !g.Equal(decimal.RequireFromString(kg)) {
					t.Errorf("type %s = %s, want %s", id, g, kg)
				}
			}
		})
	}
	// Agrees with FeedKg (§5.7) day by day.
	for d := jan(12); !d.After(jan(13)); d = d.AddDays(1) {
		byType, err := Consumption(t.Context(), f.d.App, f.farmA, d, d)
		if err != nil {
			t.Fatal(err)
		}
		bySex, err := FeedKg(t.Context(), f.d.App, f.farmA, d)
		if err != nil {
			t.Fatal(err)
		}
		if !byType[f.breeder].Equal(bySex.F) || !byType[f.roosters].Equal(bySex.M) {
			t.Errorf("%s: Consumption %v, FeedKg %+v", d, byType, bySex)
		}
	}
	if _, err := Consumption(t.Context(), f.d.App, f.farmA, jan(13), jan(12)); !errors.Is(err, errRange) {
		t.Errorf("from after to: err = %v, want errRange", err)
	}
}

func TestReadStockDays(t *testing.T) {
	f := newFixture(t)
	f.feed()
	// 14.01 complete: Breeder T 100 g × (44 + 36 + 8) = 8,8 kg; Cocoși T 150 g × 6 = 0,9 kg.
	f.alloc(jan(14), "H5.1", "F", f.breeder, "100")
	f.alloc(jan(14), "H5.2", "F", f.breeder, "100")
	f.alloc(jan(14), "H6.1", "F", f.breeder, "100")
	f.alloc(jan(14), "H6.1", "M", f.roosters, "150")
	// A type delivered but never fed: no consumption, oprit.
	idle := f.id(`insert into feed_type (name, name_key) values ('Starter T', 'starter t')`)
	e := f.entry("delivery", f.farmA, nil, jan(13))
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 80)`,
		e, f.farmA, idle, jan(13))

	s := AppSettings{StockAvgDays: 3, StockDaysMin: decimal.NewFromInt(5)}
	got, err := ReadStockDays(t.Context(), f.d.App, f.farmA, jan(14), s)
	if err != nil {
		t.Fatal(err)
	}
	if got.Window != (Window{jan(12), jan(14)}) {
		t.Errorf("window = %s–%s, want 12–14.01", got.Window.From, got.Window.To)
	}
	type want struct {
		name, kg, avg, days, runOut string
		lamp                        Status
		reason                      StockReason
	}
	// Breeder T: 1 180,336 − 8,8 = 1 171,536; (14,164 + 5,5 + 8,8) ÷ 3 = 9,488 → 123,47… → 17.05.2026.
	// Cocoși T: 197,15 − 0,9 = 196,25; (1,5 + 1,35 + 0,9) ÷ 3 = 1,25 → 157,0 → 20.06.2026.
	wants := []want{
		{"Breeder T", "1171.536", "9.488", "123.5", "2026-05-17", InNorma, StockReasonNone},
		{"Cocoși T", "196.25", "1.25", "157.0", "2026-06-20", InNorma, StockReasonNone},
		{"Starter T", "80", "0", "nil", "nil", Oprit, StockNoConsumption},
	}
	if len(got.Items) != len(wants) {
		t.Fatalf("items = %+v, want %d", got.Items, len(wants))
	}
	for i, w := range wants {
		g := got.Items[i]
		days, runOut := "nil", "nil"
		if g.Left.Days != nil {
			days = g.Left.Days.Round(1).StringFixed(1)
			runOut = g.Left.RunOut.String()
		}
		if g.Name != w.name || !g.Kg.Equal(decimal.RequireFromString(w.kg)) || !g.AvgDaily.Equal(decimal.RequireFromString(w.avg)) ||
			days != w.days || runOut != w.runOut || g.Left.Lamp != w.lamp || g.Left.Reason != w.reason {
			t.Errorf("item %d = %s %s kg avg %s, %s days, %s, %q %q; want %+v",
				i, g.Name, g.Kg, g.AvgDaily, days, runOut, g.Left.Lamp, g.Left.Reason, w)
		}
	}
}

func TestM3ReadersFailWithTheContext(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for name, call := range map[string]func() error{
		"StockAt": func() error { _, err := StockAt(ctx, f.d.App, f.farmA, jan(12)); return err },
		"StockSeries": func() error {
			_, err := StockSeries(ctx, f.d.App, f.farmA, []uuid.UUID{uuid.New()}, jan(11), jan(12))
			return err
		},
		"FeedWindow":  func() error { _, err := FeedWindow(ctx, f.d.App, f.farmA, jan(12), 3); return err },
		"Consumption": func() error { _, err := Consumption(ctx, f.d.App, f.farmA, jan(11), jan(12)); return err },
		"ReadStockDays": func() error {
			_, err := ReadStockDays(ctx, f.d.App, f.farmA, jan(12), AppSettings{StockAvgDays: 3})
			return err
		},
	} {
		if err := call(); !errors.Is(err, context.Canceled) {
			t.Errorf("%s: err = %v, want context.Canceled", name, err)
		}
	}
}
