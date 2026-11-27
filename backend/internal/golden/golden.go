// Package golden is the registry of the golden values of SPEC §11.7: the 38
// figures the real Hîrbovăț data must reproduce, each with the service call
// that produces it.
//
// The expected values live here once. TestGolden runs every row against the
// seeded test database, and migratehistory.Verify runs them inside the import
// transaction before it commits (M1-D4, M1-D5). A row whose service does not
// exist yet has Got == nil and names the milestone that brings it.
//
// Got functions call the services (calc for M1) and round once, at the
// precision §11.7 states, into plain values (ints and fixed-point strings), so
// Want and Got compare exactly. They never copy a formula into SQL.
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
		{ID: 13, Name: "guide lay SAPT 59, 53, 56", Milestone: "M2"},
		{ID: 14, Name: "lay per hall, lay alert since", Milestone: "M8"},
		{ID: 15, Name: "eggs 18.11", Milestone: "M2"},
		{ID: 16, Name: "Cat 18.11", Milestone: "M2"},
		{ID: 17, Name: "egg weight", Milestone: "M2"},
		{ID: 18, Name: "SAPT 59 week row", Milestone: "M2"},
		{ID: 19, Name: "SAPT 59 deaths, sold, kg", Milestone: "M2"},
		{ID: 20, Name: "SAPT 58 week row", Milestone: "M2"},
		{ID: 21, Name: "SAPT 58 kg, g", Milestone: "M2"},
		{ID: 22, Name: "SAPT 57, SAPT 56 week rows", Milestone: "M2"},
		{ID: 23, Name: "peak lay SAPT 30", Milestone: "M2"},
		{ID: 24, Name: "foot Total SAPT 56–59", Milestone: "M2"},
		{ID: 25, Name: "load-more label", Milestone: "M2"},
		{ID: 26, Name: "feed 18.11", Milestone: "M1",
			Want: FeedDay{Kg: calc.PerSex[string]{F: "2303.2", M: "154.7"}, Total: "2457.9"},
			Got:  feedDay},
		{ID: 27, Name: "stock end 18.11", Milestone: "M1",
			Want: []StockKg{{"Breeder 2", "31664.9"}, {"Cocoși", "1781.0"}},
			Got:  stockEnd},
		{ID: 28, Name: "days left, run-out", Milestone: "M3"},
		{ID: 29, Name: "Hîrbovăț deliveries", Milestone: "M4"},
		{ID: 30, Name: "13.11.2026 delivery price", Milestone: "M4"},
		{ID: 31, Name: "cost", Milestone: "M4"},
		{ID: 32, Name: "weighing Telița H1 vs guide", Milestone: "M9"},
		{ID: 33, Name: "3× warning H6.1", Milestone: "M5"},
		{ID: 34, Name: "missing Hrană zilnică H7.1, H7.2", Milestone: "M8"},
		{ID: 35, Name: "placement Seria 2026, Telița cycle", Milestone: "M4"},
		{ID: 36, Name: "guide spans", Milestone: "M10"},
		{ID: 37, Name: "correction 17.11 H5.2", Milestone: "M5"},
		{ID: 38, Name: "H7 empty since 14.11.2026", Milestone: "M2"},
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

// feedDay rounds each sex and the total once, from full precision.
func feedDay(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	kg, err := calc.FeedKg(ctx, env.DB, farm, AsOf)
	if err != nil {
		return nil, err
	}
	total := kg.F.Add(kg.M)
	return FeedDay{
		Kg:    calc.PerSex[string]{F: fixed(&kg.F, 1), M: fixed(&kg.M, 1)},
		Total: fixed(&total, 1),
	}, nil
}

func stockEnd(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	stock, err := calc.StockEnd(ctx, env.DB, farm, AsOf)
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
