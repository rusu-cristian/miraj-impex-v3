// Package golden is the registry of the golden values of SPEC §11.7: the 38
// figures the real Hîrbovăț data must reproduce, each with the service call
// that produces it.
//
// The expected values live here once. TestGolden runs every row against the
// seeded test database, and migratehistory.Verify runs them inside the import
// transaction before it commits (M1-D4, M1-D5). A row whose service does not
// exist yet has Got == nil and names the milestone that brings it.
//
// Got functions call the services (calc; periodtable and farms from M2; the
// Consum table and the stock readers from M3) and round once, at the
// precision §11.7 states, into plain values (ints and fixed-point strings),
// so Want and Got compare exactly. They never copy a formula into SQL.
package golden

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/textkey"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// AsOf is the day the golden values are stated at (§11.7). Rows pass it, and
// every other day they need, explicitly: none relies on app_today().
var AsOf = timex.MustDay(2026, time.November, 18)

// FarmName is the farm the golden values (and the §10.3 checks) describe.
const FarmName = "Hîrbovăț"

// FarmKey is FarmName's name key (§10.2).
var FarmKey = textkey.Key(FarmName)

// Env is what a row reads from.
type Env struct {
	// DB is a pool or the import transaction.
	DB dbq.DBTX
}

// Row is one golden value of §11.7.
type Row struct {
	ID        int
	Name      string // the start of the §11.7 "Value" column
	Milestone string // the milestone that implements Got (SPEC §12.6)
	Want      any    // nil until Got exists
	Got       func(ctx context.Context, env Env) (any, error)
}

// ErrNotImplemented is returned by Check for a row whose Got is nil.
var ErrNotImplemented = errors.New("golden: not implemented yet")

// Mismatch is the error Check returns when a row's value differs from Want.
type Mismatch struct {
	ID   int
	Name string
	Got  any
	Want any
}

func (m *Mismatch) Error() string {
	return fmt.Sprintf("golden %d %s: got %+v want %+v", m.ID, m.Name, m.Got, m.Want)
}

// Check runs the row's Got and compares the result with Want. It returns
// ErrNotImplemented when Got is nil, a *Mismatch when the values differ, and
// Got's own error (wrapped with the row id and name) when the read fails.
func (r Row) Check(ctx context.Context, env Env) error {
	if r.Got == nil {
		return fmt.Errorf("golden %d %s (%s): %w", r.ID, r.Name, r.Milestone, ErrNotImplemented)
	}
	got, err := r.Got(ctx, env)
	if err != nil {
		return fmt.Errorf("golden %d %s: %w", r.ID, r.Name, err)
	}
	if !reflect.DeepEqual(got, r.Want) {
		return &Mismatch{ID: r.ID, Name: r.Name, Got: got, Want: r.Want}
	}
	return nil
}

// Value types of the rows. Days are "YYYY-MM-DD" strings and decimals are
// fixed-point strings at the row's precision ("5.66", "2303.2").

// AgeWeek is the age in days and the age week (SAPT) of the series on Day.
type AgeWeek struct {
	Day     string
	AgeDays int
	SAPT    int
}

// WeekSpan is the first and last day of age week SAPT.
type WeekSpan struct {
	SAPT         int
	First, Last  string
	FirstWeekday time.Weekday
	LastWeekday  time.Weekday
}

// SectionCount is a bird-holding section's ♀ · ♂ count.
type SectionCount struct {
	Code  string
	Count calc.PerSex[int]
}

// DeathsSold is a day's deaths (mortality only) and sales at the farm, and the
// sales of each section that sold birds that day (hall, then compartment).
type DeathsSold struct {
	Deaths        calc.PerSex[int]
	Sold          calc.PerSex[int]
	SoldBySection []SectionCount
}

// CumMortality is the cycle's deaths and the cumulative mortality % (2 dp).
type CumMortality struct {
	Deaths calc.PerSex[int]
	Pct    calc.PerSex[string] // "" when the base is 0
}

// SoldSpan is golden 11 read two ways: Σ sold over the cycle up to AsOf
// (the sales start on 12.11) and Σ sold over the stated range 12–18.11.
type SoldSpan struct {
	Cycle calc.PerSex[int]
	Range calc.PerSex[int]
}

// LayDay is the day's eggs, hens at the start of the day and lay % (1 dp).
type LayDay struct {
	Eggs      int
	HensStart int
	Pct       string // "" when there are no hens
}

// FeedDay is the day's feed in kg per sex and in total (1 dp).
type FeedDay struct {
	Kg    calc.PerSex[string]
	Total string
}

// StockKg is the stock of one feed type at the end of the day, in kg (1 dp).
type StockKg struct {
	Name string
	Kg   string
}

// Rows returns the 38 rows of §11.7 in id order. Each call returns a new
// slice.
func Rows() []Row {
	return []Row{
		{ID: 1, Name: "age, SAPT on 18.11.2026", Milestone: "M1",
			Want: []AgeWeek{{"2026-11-18", 407, 59}},
			Got:  ageWeeks(AsOf)},
		{ID: 2, Name: "age, SAPT on 24.02.2026 and 22.02.2026", Milestone: "M1",
			Want: []AgeWeek{{"2026-02-24", 140, 21}, {"2026-02-22", 138, 20}},
			Got:  ageWeeks(day(2026, time.February, 24), day(2026, time.February, 22))},
		{ID: 3, Name: "SAPT on 20.10.2025 (13 days) and 21.10.2025 (14 days)", Milestone: "M1",
			Want: []AgeWeek{{"2025-10-20", 13, 2}, {"2025-10-21", 14, 3}},
			Got:  ageWeeks(day(2025, time.October, 20), day(2025, time.October, 21))},
		{ID: 4, Name: "SAPT 59 days", Milestone: "M1",
			Want: WeekSpan{SAPT: 59, First: "2026-11-17", Last: "2026-11-23", FirstWeekday: time.Tuesday, LastWeekday: time.Monday},
			Got:  weekSpan(59)},
		{ID: 5, Name: "alive end 18.11", Milestone: "M1",
			Want: calc.PerSex[int]{F: 12561, M: 794},
			Got:  farmDay(func(d calc.FarmDay) any { return d.AliveEnd })},
		{ID: 6, Name: "alive start 18.11", Milestone: "M1",
			Want: calc.PerSex[int]{F: 14395, M: 1083},
			Got:  farmDay(func(d calc.FarmDay) any { return d.AliveStart })},
		{ID: 7, Name: "alive end per section 18.11", Milestone: "M1",
			Want: []SectionCount{
				{"H5.1", calc.PerSex[int]{F: 4263, M: 342}},
				{"H5.2", calc.PerSex[int]{F: 4281, M: 323}},
				{"H6.1", calc.PerSex[int]{F: 3542, M: 124}},
				{"H6.2", calc.PerSex[int]{F: 475, M: 5}},
				{"H7.1", calc.PerSex[int]{F: 0, M: 0}},
				{"H7.2", calc.PerSex[int]{F: 0, M: 0}},
			},
			Got: sectionsAliveEnd},
		{ID: 8, Name: "deaths, sold 18.11", Milestone: "M1",
			Want: DeathsSold{
				Deaths: calc.PerSex[int]{F: 3, M: 0},
				Sold:   calc.PerSex[int]{F: 1831, M: 289},
				SoldBySection: []SectionCount{
					{"H6.1", calc.PerSex[int]{F: 676, M: 201}},
					{"H6.2", calc.PerSex[int]{F: 1155, M: 88}},
				},
			},
			Got: deathsSold},
		{ID: 9, Name: "arrivals", Milestone: "M1",
			Want: calc.PerSex[int]{F: 26883, M: 3024},
			Got:  arrivals},
		{ID: 10, Name: "cycle deaths, cum. mortality", Milestone: "M1",
			Want: CumMortality{Deaths: calc.PerSex[int]{F: 1521, M: 1048}, Pct: calc.PerSex[string]{F: "5.66", M: "34.66"}},
			Got:  cumMortality},
		{ID: 11, Name: "sold 12–18.11", Milestone: "M1",
			Want: SoldSpan{Cycle: calc.PerSex[int]{F: 12801, M: 1182}, Range: calc.PerSex[int]{F: 12801, M: 1182}},
			Got:  sold},
		{ID: 12, Name: "lay 18.11", Milestone: "M1",
			Want: LayDay{Eggs: 5914, HensStart: 14395, Pct: "41.1"},
			Got:  layDay},
		{ID: 13, Name: "guide lay SAPT 59, 53, 56", Milestone: "M2",
			Want: GuideLay{Guide59: "57.6", DiffToday: "-16.5", StatusToday: calc.Abatere, Guide53: "64.5", Guide56: "61.1"},
			Got:  guideLay},
		// Golden 14 is split (M2 Q2): M2 checks the lay per hall; M8 adds
		// "H5 below guide since 01.11.2026" (alert lay.since) to Want.
		{ID: 14, Name: "lay per hall 18.11 (alert since: M8)", Milestone: "M2",
			Want: []HallLay{{Hall: "H5", Pct: "45.3", DiffPts: "-12.3"}, {Hall: "H6", Pct: "34.9", DiffPts: "-22.7"}},
			Got:  layPerHall},
		{ID: 15, Name: "eggs 18.11", Milestone: "M2",
			Want: EggsDay{Eggs: 5914, Inc: 5308, IncPct: "89.8", Jos: 158, JosPct: "2.7"},
			Got:  eggsDay},
		{ID: 16, Name: "Cat 18.11", Milestone: "M2",
			Want: []HallEggs{
				{Hall: "H5", Cat: [3]int{13, 417, 27}, Inc: 3415, Total: 3872},
				{Hall: "H6", Cat: [3]int{8, 110, 31}, Inc: 1893, Total: 2042},
			},
			Got: hallEggs},
		{ID: 17, Name: "egg weight", Milestone: "M2",
			Want: EggWeight{FarmDay: "66.4", H7Week58: "66.3", FarmWeek58: "66.6"},
			Got:  eggWeight},
		{ID: 18, Name: "SAPT 59 week row", Milestone: "M2",
			Want: LayWeek{Label: "SAPT 59 · 17.11–18.11 · 2 din 7 zile", DaysInRow: 2, Eggs: 11806, HenDays: 29221,
				Lay: "40.4", Guide: "57.6", DiffPts: "-17.2", IncPct: "92.5", JosPct: "2.6", CatPct: [3]string{"0.3", "6.2", "1.1"}},
			Got: layWeek59},
		{ID: 19, Name: "SAPT 59 deaths, sold, kg", Milestone: "M2",
			Want: WeekRemovals{Deaths: calc.PerSex[int]{F: 10, M: 0}, Sold: calc.PerSex[int]{F: 2255, M: 433},
				FeedKg: calc.PerSex[string]{F: "4675.3", M: "329.7"}},
			Got: weekRemovals(59)},
		// Goldens 20 and 24 hold under the hen-day rule of M2 Q1 (b): a unit-day
		// counts when it has eggs or the unit is emptied that day (H7 on
		// 14.11.2026, 6 915 hens). §5.6's old rule gives 149 503 → 31,8 % and
		// 538 123 → 42,1 %. Golden 24's hen-days (545 038) are not in §11.7: they
		// are the M2 plan's check of the 41,6 %.
		{ID: 20, Name: "SAPT 58 week row", Milestone: "M2",
			Want: Week58{Eggs: 47512, HenDays: 156418, Lay: "30.4", Guide: "58.8", IncPct: "96.3", CatPct: [3]string{"0.3", "2.6", "0.8"},
				Deaths: calc.PerSex[int]{F: 36, M: 8}, Sold: calc.PerSex[int]{F: 10546, M: 749}, AliveEnd: calc.PerSex[int]{F: 14826, M: 1227}},
			Got: week58},
		{ID: 21, Name: "SAPT 58 kg, g", Milestone: "M2",
			Want: WeekFeed{Kg: calc.PerSex[string]{F: "22230.9", M: "1682.7"}, G: calc.PerSex[string]{F: "155.9", M: "142.2"}},
			Got:  weekFeed(58)},
		{ID: 22, Name: "SAPT 57, SAPT 56 week rows", Milestone: "M2",
			Want: Weeks5756{
				Weeks:    []WeekLay{{Sapt: 57, Eggs: 49225, Lay: "27.6", Guide: "59.9"}, {Sapt: 56, Eggs: 117929, Lay: "65.2", Guide: "61.1"}},
				IncPct56: "98.3",
			},
			Got: weeks5756},
		{ID: 23, Name: "peak lay SAPT 30", Milestone: "M2",
			Want: PeakLay{Sapt: 30, Eggs: 168122, HenDays: 186088, Lay: "90.3"},
			Got:  peakLay},
		{ID: 24, Name: "foot Total SAPT 56–59", Milestone: "M2",
			Want: Foot{Label: "Total SAPT 56–59", Eggs: 226472, HenDays: 545038, Lay: "41.6", Guide: "59.9",
				Deaths: calc.PerSex[int]{F: 534, M: 77}, Sold: calc.PerSex[int]{F: 12801, M: 1182}, IncPct: "97.7",
				FeedKg: calc.PerSex[string]{F: "78898.6", M: "5996.6"}},
			Got: foot},
		{ID: 25, Name: "load-more label", Milestone: "M2",
			Want: LoadMore{Label: "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40", SaptFrom: 20, SaptTo: 59, WeeksTotal: 40},
			Got:  loadMore},
		// Goldens 26 and 27 are M1's (SPEC §12.6); M3 re-points them (M3 D12):
		// 26 reads the Consum day row, 27 the stock reader without dense views.
		{ID: 26, Name: "feed 18.11", Milestone: "M1",
			Want: FeedDay{Kg: calc.PerSex[string]{F: "2303.2", M: "154.7"}, Total: "2457.9"},
			Got:  feedDay},
		{ID: 27, Name: "stock end 18.11", Milestone: "M1",
			Want: []StockKg{{"Breeder 2", "31664.9"}, {"Cocoși", "1781.0"}},
			Got:  stockEnd},
		{ID: 28, Name: "days left, run-out", Milestone: "M3",
			Want: StockDaysLeft{From: "2026-11-16", To: "2026-11-18", Items: []TypeDaysLeft{
				{Type: "Breeder 2", AvgDaily: "2392.2", Days: "13.2", RunOut: "2026-12-01"},
				{Type: "Cocoși", AvgDaily: "170.9", Days: "10.4", RunOut: "2026-11-28"},
			}},
			Got: daysLeft},
		{ID: 29, Name: "Hîrbovăț deliveries", Milestone: "M4"},
		{ID: 30, Name: "13.11.2026 delivery price", Milestone: "M4"},
		{ID: 31, Name: "cost", Milestone: "M4"},
		{ID: 32, Name: "weighing Telița H1 vs guide", Milestone: "M9"},
		{ID: 33, Name: "3× warning H6.1", Milestone: "M5"},
		{ID: 34, Name: "missing Hrană zilnică H7.1, H7.2", Milestone: "M8"},
		{ID: 35, Name: "placement Seria 2026, Telița cycle", Milestone: "M4"},
		{ID: 36, Name: "guide spans", Milestone: "M10"},
		{ID: 37, Name: "correction 17.11 H5.2", Milestone: "M5"},
		{ID: 38, Name: "H7 empty since 14.11.2026", Milestone: "M2",
			Want: []EmptyHall{{Hall: "H7", Since: "2026-11-14", Tile: "goală din 14.11.2026 · H7.1, H7.2", Note: "H7 goală din 14.11"}},
			Got:  emptyHalls},
	}
}

func day(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }

// FarmID finds the golden farm (FarmName) by its name key.
func FarmID(ctx context.Context, db dbq.DBTX) (uuid.UUID, error) {
	farms, err := dbq.New(db).ListFarms(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("list farms: %w", err)
	}
	for _, f := range farms {
		if textkey.Key(f.Name) == FarmKey {
			return f.ID, nil
		}
	}
	return uuid.Nil, fmt.Errorf("no farm with the name key %q", FarmKey)
}

// cycle returns the golden farm and its cycle on AsOf: Seria 2026 at
// Hîrbovăț, open since 22.02.2026.
func cycle(ctx context.Context, db dbq.DBTX) (uuid.UUID, calc.Cycle, error) {
	farm, err := FarmID(ctx, db)
	if err != nil {
		return uuid.Nil, calc.Cycle{}, err
	}
	c, err := calc.CycleAt(ctx, db, farm, AsOf)
	return farm, c, err
}

// ageWeeks reads the age of the golden cycle's series on each day (§5.1).
func ageWeeks(days ...timex.Day) func(context.Context, Env) (any, error) {
	return func(ctx context.Context, env Env) (any, error) {
		_, c, err := cycle(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		out := make([]AgeWeek, 0, len(days))
		for _, d := range days {
			out = append(out, AgeWeek{Day: d.String(), AgeDays: calc.AgeDays(c.Age, d), SAPT: calc.SAPT(c.Age, d)})
		}
		return out, nil
	}
}

// weekSpan reads the days of age week k of the golden cycle's series (§5.1).
func weekSpan(k int) func(context.Context, Env) (any, error) {
	return func(ctx context.Context, env Env) (any, error) {
		_, c, err := cycle(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		first, last := calc.WeekFirstDay(c.Age, k), calc.WeekLastDay(c.Age, k)
		return WeekSpan{SAPT: k, First: first.String(), Last: last.String(),
			FirstWeekday: first.Weekday(), LastWeekday: last.Weekday()}, nil
	}
}

// farmDay reads the golden farm's figures on AsOf and picks a part of them.
func farmDay(pick func(calc.FarmDay) any) func(context.Context, Env) (any, error) {
	return func(ctx context.Context, env Env) (any, error) {
		farm, err := FarmID(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		d, err := calc.ReadFarmDay(ctx, env.DB, farm, AsOf)
		if err != nil {
			return nil, err
		}
		return pick(d), nil
	}
}

func sectionsAliveEnd(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	secs, err := calc.SectionsAliveEnd(ctx, env.DB, farm, AsOf)
	if err != nil {
		return nil, err
	}
	out := make([]SectionCount, 0, len(secs))
	for _, s := range secs {
		out = append(out, SectionCount{Code: s.Code, Count: s.AliveEnd})
	}
	return out, nil
}

// deathsSold reads the farm's deaths and sales on AsOf and the sales of each
// section that sold birds that day.
func deathsSold(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	d, err := calc.ReadFarmDay(ctx, env.DB, farm, AsOf)
	if err != nil {
		return nil, err
	}
	secs, err := calc.SectionsRemovals(ctx, env.DB, farm, AsOf, AsOf)
	if err != nil {
		return nil, err
	}
	bySection := []SectionCount{}
	for _, s := range secs {
		if s.Sold != (calc.PerSex[int]{}) {
			bySection = append(bySection, SectionCount{Code: s.Code, Count: s.Sold})
		}
	}
	return DeathsSold{Deaths: d.Deaths, Sold: d.Sold, SoldBySection: bySection}, nil
}

// arrivals is the base of §5.5 at the farm over the cycle, up to AsOf.
func arrivals(ctx context.Context, env Env) (any, error) {
	farm, c, err := cycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	return calc.FarmBase(ctx, env.DB, farm, c.StartDate, AsOf)
}

func cumMortality(ctx context.Context, env Env) (any, error) {
	farm, c, err := cycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	deaths, err := calc.CycleDeaths(ctx, env.DB, farm, c.StartDate, AsOf)
	if err != nil {
		return nil, err
	}
	base, err := calc.FarmBase(ctx, env.DB, farm, c.StartDate, AsOf)
	if err != nil {
		return nil, err
	}
	return CumMortality{Deaths: deaths, Pct: calc.PerSex[string]{
		F: fixed(calc.CumMortalityPct(deaths.F, base.F), 2),
		M: fixed(calc.CumMortalityPct(deaths.M, base.M), 2),
	}}, nil
}

// sold is Σ sold in the cycle up to AsOf and Σ sold over 12.11–18.11
// (SoldSpan): the two agree because the first sale is on 12.11.
func sold(ctx context.Context, env Env) (any, error) {
	farm, c, err := cycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	cycleSold, err := calc.Sold(ctx, env.DB, farm, c.StartDate, AsOf)
	if err != nil {
		return nil, err
	}
	rangeSold, err := calc.Sold(ctx, env.DB, farm, day(2026, time.November, 12), AsOf)
	if err != nil {
		return nil, err
	}
	return SoldSpan{Cycle: cycleSold, Range: rangeSold}, nil
}

func layDay(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	l, err := calc.LayDay(ctx, env.DB, farm, AsOf)
	if err != nil {
		return nil, err
	}
	return LayDay{Eggs: l.Eggs, HensStart: l.HensStart, Pct: fixed(l.Pct(), 1)}, nil
}

// stockEnd reads the stock per feed type at the end of AsOf (calc.StockAt,
// running sums without the dense views, M3 D7).
func stockEnd(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	stock, err := calc.StockAt(ctx, env.DB, farm, AsOf)
	if err != nil {
		return nil, err
	}
	out := make([]StockKg, 0, len(stock))
	for _, s := range stock {
		out = append(out, StockKg{Name: s.Name, Kg: fixed(&s.Kg, 1)})
	}
	return out, nil
}

// fixed rounds d once, half away from zero, and writes it with exactly places
// decimals; nil (a rate without a base) is "".
func fixed(d *decimal.Decimal, places int32) string {
	if d == nil {
		return ""
	}
	return num.Round(*d, places).StringFixed(places)
}
