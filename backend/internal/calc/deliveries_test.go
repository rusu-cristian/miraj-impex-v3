package calc

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestDeliveryRange(t *testing.T) {
	day := func(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }
	ptr := func(d timex.Day) *timex.Day { return &d }
	today := day(2026, time.November, 18)
	seria2025 := SeriesAge{PlacementDate: day(2024, time.November, 1)}
	seria2026 := SeriesAge{PlacementDate: day(2025, time.October, 7)}
	cycle := func(start timex.Day, end *timex.Day, age SeriesAge) CycleInfo {
		return CycleInfo{Cycle: Cycle{ID: uuid.New(), StartDate: start, EndDate: end, Age: age}}
	}
	// Hîrbovăț: Seria 2025 (an earlier cycle ended 10.02.2025, then 07.03.2025–28.01.2026), Seria 2026 from
	// 22.02.2026, open. Newest first, as FarmCycles returns them.
	hOld := cycle(day(2024, time.November, 1), ptr(day(2025, time.February, 10)), seria2025)
	h2025 := cycle(day(2025, time.March, 7), ptr(day(2026, time.January, 28)), seria2025)
	h2026 := cycle(day(2026, time.February, 22), nil, seria2026)
	hirbovat := []CycleInfo{h2026, h2025, hOld}
	// Telița: Seria 2026 07.10.2025–28.02.2026, its only cycle.
	t2026 := cycle(day(2025, time.October, 7), ptr(day(2026, time.February, 28)), seria2026)
	telita := []CycleInfo{t2026}

	cyclePeriod := Period{Kind: PeriodCycle}
	for _, tc := range []struct {
		name   string
		cycles []CycleInfo
		shown  *CycleInfo
		p      Period
		want   DayRange
	}{
		// (28.01.2026, 18.11.2026]: the 17.02 pre-stocking delivery is inside.
		{"open cycle", hirbovat, &h2026, cyclePeriod, DayRange{ptr(day(2026, time.January, 29)), today}},
		{"closed cycle", hirbovat, &h2025, cyclePeriod, DayRange{ptr(day(2025, time.February, 11)), day(2026, time.January, 28)}},
		{"no previous cycle", telita, &t2026, cyclePeriod, DayRange{nil, day(2026, time.February, 28)}},
		{"the shown cycle need not be in cycles", nil, &h2026, cyclePeriod, DayRange{nil, today}},
		{"no cycle at all", nil, nil, cyclePeriod, DayRange{nil, today}},
		{"no cycle at all, weeks", nil, nil, Period{Kind: PeriodWeeks, Weeks: 4}, DayRange{nil, today}},
		// SAPT 56–59 of Seria 2026: 27.10–18.11.2026.
		{"weeks:4", hirbovat, &h2026, Period{Kind: PeriodWeeks, Weeks: 4}, DayRange{ptr(day(2026, time.October, 27)), today}},
		// Clipped to the cycle start (22.02, SAPT 20), so no pre-stocking day.
		{"weeks beyond the cycle start", hirbovat, &h2026, Period{Kind: PeriodWeeks, Weeks: 100},
			DayRange{ptr(day(2026, time.February, 22)), today}},
		// The S11 empty board: Telița closed on 28.02, the interval is kept as typed.
		{"interval", telita, &t2026, Period{Kind: PeriodInterval, From: day(2026, time.March, 1), To: today},
			DayRange{ptr(day(2026, time.March, 1)), today}},
		{"interval without a cycle", nil, nil, Period{Kind: PeriodInterval, From: day(2025, time.January, 1), To: day(2025, time.January, 31)},
			DayRange{ptr(day(2025, time.January, 1)), day(2025, time.January, 31)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeliveryRange(tc.cycles, tc.shown, tc.p, today)
			if err != nil {
				t.Fatal(err)
			}
			if !sameRange(got, tc.want) {
				t.Errorf("got %s, want %s", rangeText(got), rangeText(tc.want))
			}
		})
	}

	for _, tc := range []struct {
		name string
		p    Period
		err  error
	}{
		{"interval ending after today", Period{Kind: PeriodInterval, From: today, To: today.AddDays(1)}, ErrIntervalFuture},
		{"interval starting after it ends", Period{Kind: PeriodInterval, From: today, To: today.AddDays(-1)}, ErrPeriod},
		{"interval without days", Period{Kind: PeriodInterval}, ErrPeriod},
		{"weeks:0", Period{Kind: PeriodWeeks}, ErrPeriod},
		{"unknown kind", Period{Kind: "luna"}, ErrPeriod},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := DeliveryRange(hirbovat, &h2026, tc.p, today); !errors.Is(err, tc.err) {
				t.Errorf("got %s, %v; want %v", rangeText(got), err, tc.err)
			}
		})
	}
}

func sameRange(a, b DayRange) bool {
	return a.To == b.To && (a.From == nil) == (b.From == nil) && (a.From == nil || *a.From == *b.From)
}

func rangeText(r DayRange) string {
	from := "…"
	if r.From != nil {
		from = r.From.String()
	}
	return "[" + from + ", " + r.To.String() + "]"
}

func TestDayRangeContains(t *testing.T) {
	from := jan(10)
	r := DayRange{From: &from, To: jan(12)}
	for d, want := range map[timex.Day]bool{jan(9): false, jan(10): true, jan(12): true, jan(13): false} {
		if r.Contains(d) != want {
			t.Errorf("Contains(%s) = %t, want %t", d, !want, want)
		}
	}
	if !(DayRange{To: jan(12)}).Contains(timex.MustDay(2000, time.January, 1)) {
		t.Error("a range without From contains every earlier day")
	}
}

// deliveries adds feed types and deliveries to the fixture. Farm B gets a second cycle from 20.01, open.
//
//	A: 08.01 Breeder 100 (pre-stocking)        A: 12.01 Cocoși 5, Breeder 500 (inserted in that order)
//	   10.01 Breeder 1 000 · 8 000 lei (total)    12.01 Breeder 40, soft-deleted
//	   10.01 Cocoși 200                           13.01 Cocoși 20, its entry deleted
//	   10.01 Mix 30 · 75 lei (din registru)       13.01 Breeder 10
//	   25.01 Breeder 70 · 560 lei (per kg)
//	B: 10.01 Breeder 50 (the first cycle's end day), 21.01 Breeder 60
func (f *fixture) deliveries() (bCycles []CycleInfo) {
	f.t.Helper()
	f.breeder = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder T', 'breeder t', 'F')`)
	f.roosters = f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși T', 'cocosi t', 'M')`)
	mix := f.id(`insert into feed_type (name, name_key) values ('Mix T', 'mix t')`)
	user := f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	deliver := func(farm, ft uuid.UUID, d timex.Day, kg string) (entry, id uuid.UUID) {
		entry = f.entry("delivery", farm, nil, d)
		return entry, f.id(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`,
			entry, farm, ft, d, kg)
	}
	price := func(id uuid.UUID, perKg, total, input string) {
		f.exec(`update feed_delivery set price_per_kg = $2, total_lei = $3, price_input = $4, priced_by = $5, priced_at = $6
			where id = $1`, id, perKg, total, input, user, testdb.Now())
	}
	A, B := f.farmA, f.farmB
	deliver(A, f.breeder, jan(8), "100")
	_, id := deliver(A, f.breeder, jan(10), "1000")
	price(id, "8", "8000", "total")
	deliver(A, f.roosters, jan(10), "200")
	_, id = deliver(A, mix, jan(10), "30")
	f.exec(`update feed_delivery set price_per_kg = 2.5, total_lei = 75, price_from_register = true where id = $1`, id)
	deliver(A, f.roosters, jan(12), "5")
	deliver(A, f.breeder, jan(12), "500")
	_, id = deliver(A, f.breeder, jan(12), "40")
	f.exec(`update feed_delivery set deleted_at = $2 where id = $1`, id, testdb.Now())
	e, _ := deliver(A, f.roosters, jan(13), "20")
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, e, testdb.Now(), user)
	deliver(A, f.breeder, jan(13), "10")
	_, id = deliver(A, f.breeder, jan(25), "70")
	price(id, "8", "560", "per_kg")
	deliver(B, f.breeder, jan(10), "50")
	deliver(B, f.breeder, jan(21), "60")

	second := f.id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`,
		f.series, B, jan(20))
	first := jan(10)
	return []CycleInfo{
		{Cycle: Cycle{ID: second, StartDate: jan(20)}, FarmID: B},
		{Cycle: Cycle{ID: f.cycleB, StartDate: jan(5), EndDate: &first}, FarmID: B},
	}
}

func TestReadDeliveries(t *testing.T) {
	f := newFixture(t)
	bCycles := f.deliveries()
	today := jan(31)
	rangeOf := func(farm uuid.UUID, shown CycleInfo) FarmRange {
		t.Helper()
		r, err := DeliveryRange(bCycles, &shown, Period{Kind: PeriodCycle}, today)
		if err != nil {
			t.Fatal(err)
		}
		return FarmRange{FarmID: farm, DayRange: r}
	}
	from14 := jan(14)
	ranges := []FarmRange{
		rangeOf(f.farmB, bCycles[0]),                                      // (10.01, 31.01]
		{FarmID: f.farmA, DayRange: DayRange{To: jan(20)}},                // up to 20.01
		rangeOf(f.farmB, bCycles[1]),                                      // up to 10.01
		{FarmID: f.farmA, DayRange: DayRange{From: &from14, To: jan(20)}}, // none
	}
	got, err := ReadDeliveries(t.Context(), f.d.App, ranges)
	if err != nil {
		t.Fatal(err)
	}
	type item struct {
		date timex.Day
		typ  string
		kg   string
	}
	type want struct {
		count                   int
		kgF, kgM, kgBoth, total string
		pricedCount             int
		pricedKg, lei           string
		unpricedKg              string
		dates                   []timex.Day
		items                   []item
	}
	wants := []want{
		// The 10.01 delivery is on the first cycle's end day: it belongs to that cycle, not this one.
		{1, "60", "0", "0", "60", 0, "0", "0", "60", []timex.Day{jan(21)}, []item{{jan(21), "Breeder T", "60"}}},
		{7, "1610", "205", "30", "1845", 2, "1030", "8075", "815", []timex.Day{jan(8), jan(10), jan(12), jan(13)},
			[]item{{jan(8), "Breeder T", "100"}, {jan(10), "Cocoși T", "200"}, {jan(12), "Breeder T", "500"},
				{jan(12), "Cocoși T", "5"}, {jan(13), "Breeder T", "10"}}},
		{1, "50", "0", "0", "50", 0, "0", "0", "50", []timex.Day{jan(10)}, []item{{jan(10), "Breeder T", "50"}}},
		{0, "0", "0", "0", "0", 0, "0", "0", "0", []timex.Day{}, []item{}},
	}
	if len(got) != len(wants) {
		t.Fatalf("got %d results, want %d", len(got), len(wants))
	}
	d := decimal.RequireFromString
	for i, w := range wants {
		g := got[i]
		if g.FarmID != ranges[i].FarmID || !sameRange(g.DayRange, ranges[i].DayRange) {
			t.Errorf("%d: range %s %s, want %s %s", i, g.FarmID, rangeText(g.DayRange), ranges[i].FarmID, rangeText(ranges[i].DayRange))
		}
		if g.Count != w.count || !g.Kg.F.Equal(d(w.kgF)) || !g.Kg.M.Equal(d(w.kgM)) || !g.KgBoth.Equal(d(w.kgBoth)) ||
			!g.TotalKg().Equal(d(w.total)) {
			t.Errorf("%d: %d deliveries, kg ♀ %s ♂ %s ambele %s (Σ %s); want %d, %s %s %s (%s)",
				i, g.Count, g.Kg.F, g.Kg.M, g.KgBoth, g.TotalKg(), w.count, w.kgF, w.kgM, w.kgBoth, w.total)
		}
		if g.Priced.Count != w.pricedCount || !g.Priced.Kg.Equal(d(w.pricedKg)) || !g.Priced.Lei.Equal(d(w.lei)) {
			t.Errorf("%d: priced %+v, want %d · %s kg · %s lei", i, g.Priced, w.pricedCount, w.pricedKg, w.lei)
		}
		u := g.Unpriced
		if u.Count != len(w.items) || !u.Kg.Equal(d(w.unpricedKg)) || u.Dates == nil || !slices.Equal(u.Dates, w.dates) {
			t.Errorf("%d: unpriced %d · %s kg · %v, want %d · %s · %v", i, u.Count, u.Kg, u.Dates, len(w.items), w.unpricedKg, w.dates)
		}
		if len(u.Items) != len(w.items) {
			t.Fatalf("%d: unpriced items %+v, want %+v", i, u.Items, w.items)
		}
		for j, wi := range w.items {
			gi := u.Items[j]
			if gi.Date != wi.date || gi.Type != wi.typ || !gi.Kg.Equal(d(wi.kg)) || gi.ID == uuid.Nil || gi.FeedTypeID == uuid.Nil {
				t.Errorf("%d: item %d = %+v, want %+v", i, j, gi, wi)
			}
		}
		if first, last, ok := u.Span(); ok != (len(w.dates) > 0) || (ok && (first != w.dates[0] || last != w.dates[len(w.dates)-1])) {
			t.Errorf("%d: span %s–%s (%t)", i, first, last, ok)
		}
	}

	if got, err := ReadDeliveries(t.Context(), f.d.App, nil); err != nil || got == nil || len(got) != 0 {
		t.Errorf("no ranges: %v, %v; want an empty list", got, err)
	}
	from := jan(21)
	if _, err := ReadDeliveries(t.Context(), f.d.App, []FarmRange{{FarmID: f.farmA, DayRange: DayRange{From: &from, To: jan(20)}}}); !errors.Is(err, errRange) {
		t.Errorf("a range that starts after it ends: %v, want errRange", err)
	}
}
