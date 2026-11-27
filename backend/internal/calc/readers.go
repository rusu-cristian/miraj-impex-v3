package calc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ErrNoCycle is returned by CycleAt when no cycle of the farm contains the day.
var ErrNoCycle = errors.New("calc: no cycle of the farm contains the day")

// ErrNoSeries is returned by ReadSeriesAge for an unknown series.
var ErrNoSeries = errors.New("calc: series not found")

// errRange is returned when a range starts after it ends.
var errRange = errors.New("calc: the range starts after it ends")

// ReadSeriesAge reads the age parameters of a series (§5.1).
func ReadSeriesAge(ctx context.Context, db dbq.DBTX, seriesID uuid.UUID) (SeriesAge, error) {
	r, err := dbq.New(db).CalcSeriesAge(ctx, seriesID)
	if errors.Is(err, pgx.ErrNoRows) {
		return SeriesAge{}, ErrNoSeries
	}
	if err != nil {
		return SeriesAge{}, fmt.Errorf("calc: series age: %w", err)
	}
	return SeriesAge{PlacementDate: r.PlacementDate, AgeAtPlacementDays: int(r.AgeAtPlacementDays)}, nil
}

// Cycle is one stay of a series in a farm (§4.6) with the series' age
// parameters.
type Cycle struct {
	ID        uuid.UUID
	SeriesID  uuid.UUID
	StartDate timex.Day
	EndDate   *timex.Day // nil while the cycle is open
	Age       SeriesAge
}

// CycleAt returns the cycle of the farm whose [start, end] contains day (an
// open cycle has no end), or ErrNoCycle.
func CycleAt(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) (Cycle, error) {
	r, err := dbq.New(db).CalcCycleAt(ctx, dbq.CalcCycleAtParams{FarmID: farmID, Day: day})
	if errors.Is(err, pgx.ErrNoRows) {
		return Cycle{}, ErrNoCycle
	}
	if err != nil {
		return Cycle{}, fmt.Errorf("calc: cycle at %s: %w", day, err)
	}
	return Cycle{
		ID:        r.ID,
		SeriesID:  r.SeriesID,
		StartDate: r.StartDate,
		EndDate:   r.EndDate,
		Age:       SeriesAge{PlacementDate: r.PlacementDate, AgeAtPlacementDays: int(r.AgeAtPlacementDays)},
	}, nil
}

// FarmDay is one farm's bird figures for one day (§5.3, §5.4).
type FarmDay struct {
	AliveStart PerSex[int] // alive at the start of the day = end of the day before
	AliveEnd   PerSex[int]
	Deaths     PerSex[int] // mortality only
	Culls      PerSex[int]
	Sold       PerSex[int]
}

// ReadFarmDay reads the farm × sex figures of day (§5.3, §5.4).
func ReadFarmDay(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) (FarmDay, error) {
	r, err := dbq.New(db).CalcFarmSexDay(ctx, dbq.CalcFarmSexDayParams{FarmID: farmID, Day: day})
	if err != nil {
		return FarmDay{}, fmt.Errorf("calc: farm day %s: %w", day, err)
	}
	return FarmDay{
		AliveStart: pair(r.AliveStartF, r.AliveStartM),
		AliveEnd:   pair(r.AliveEndF, r.AliveEndM),
		Deaths:     pair(r.DeathsF, r.DeathsM),
		Culls:      pair(r.CullsF, r.CullsM),
		Sold:       pair(r.SoldF, r.SoldM),
	}, nil
}

// SectionAlive is the count of one bird-holding section at the end of a day.
type SectionAlive struct {
	SectionID uuid.UUID
	Code      string
	AliveEnd  PerSex[int]
}

// SectionsAliveEnd returns every bird-holding section of the farm (a section
// without children) with its count at the end of day, 0 when empty, in tree
// order (hall, then compartment).
func SectionsAliveEnd(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]SectionAlive, error) {
	rows, err := dbq.New(db).CalcSectionAliveEnd(ctx, dbq.CalcSectionAliveEndParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: sections alive %s: %w", day, err)
	}
	out := make([]SectionAlive, 0, len(rows))
	for _, r := range rows {
		out = append(out, SectionAlive{SectionID: r.SectionID, Code: r.Code, AliveEnd: pair(r.AliveEndF, r.AliveEndM)})
	}
	return out, nil
}

func readRemovals(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (dbq.CalcFarmRemovalsRow, error) {
	if from.After(to) {
		return dbq.CalcFarmRemovalsRow{}, errRange
	}
	r, err := dbq.New(db).CalcFarmRemovals(ctx, dbq.CalcFarmRemovalsParams{FarmID: farmID, FromDay: from, ToDay: to})
	if err != nil {
		return r, fmt.Errorf("calc: removals %s–%s: %w", from, to, err)
	}
	return r, nil
}

// CycleDeaths is the deaths of §5.5 at the farm in [cycleStart, day]:
// mortality + culls (the register counts culls as deaths, D165), never sales.
func CycleDeaths(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, cycleStart, day timex.Day) (PerSex[int], error) {
	r, err := readRemovals(ctx, db, farmID, cycleStart, day)
	if err != nil {
		return PerSex[int]{}, err
	}
	return pair(r.DeathsF+r.CullsF, r.DeathsM+r.CullsM), nil
}

// Sold is the birds sold at the farm in [from, to] (§5.4).
func Sold(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (PerSex[int], error) {
	r, err := readRemovals(ctx, db, farmID, from, to)
	if err != nil {
		return PerSex[int]{}, err
	}
	return pair(r.SoldF, r.SoldM), nil
}

// Removals are a farm's deaths (§5.5: mortality + culls, D165) and sales
// (§5.4) over a day range, from one read.
type Removals struct {
	Deaths PerSex[int]
	Sold   PerSex[int]
}

// FarmRemovals is CycleDeaths and Sold of [from, to] in one statement, for a
// caller that needs both.
func FarmRemovals(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (Removals, error) {
	r, err := readRemovals(ctx, db, farmID, from, to)
	if err != nil {
		return Removals{}, err
	}
	return Removals{Deaths: pair(r.DeathsF+r.CullsF, r.DeathsM+r.CullsM), Sold: pair(r.SoldF, r.SoldM)}, nil
}

// SectionRemovals is one bird-holding section's removals over a day range
// (§5.4).
type SectionRemovals struct {
	SectionID uuid.UUID
	Code      string
	Mortality PerSex[int]
	Culls     PerSex[int]
	Sold      PerSex[int] // never deaths (D155)
}

// Deaths is mortality + culls, as the register and §5.5 count them (D165).
func (s SectionRemovals) Deaths() PerSex[int] {
	return PerSex[int]{F: s.Mortality.F + s.Culls.F, M: s.Mortality.M + s.Culls.M}
}

// SectionsRemovals returns every bird-holding section of the farm (a section
// without children) with its mortality, culls and sales in [from, to], 0 when
// it has none, in tree order (hall, then compartment).
func SectionsRemovals(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) ([]SectionRemovals, error) {
	if from.After(to) {
		return nil, errRange
	}
	rows, err := dbq.New(db).CalcSectionRemovals(ctx, dbq.CalcSectionRemovalsParams{FarmID: farmID, FromDay: from, ToDay: to})
	if err != nil {
		return nil, fmt.Errorf("calc: section removals %s–%s: %w", from, to, err)
	}
	out := make([]SectionRemovals, 0, len(rows))
	for _, r := range rows {
		out = append(out, SectionRemovals{
			SectionID: r.SectionID,
			Code:      r.Code,
			Mortality: pair(r.MortalityF, r.MortalityM),
			Culls:     pair(r.CullsF, r.CullsM),
			Sold:      pair(r.SoldF, r.SoldM),
		})
	}
	return out, nil
}

// FarmBase is the base of §5.5 at farm scope in [cycleStart, day]: the birds
// that entered the farm from outside it (placements and transfers in from
// another farm's section). Adjustments and moves inside the farm don't count.
func FarmBase(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, cycleStart, day timex.Day) (PerSex[int], error) {
	if cycleStart.After(day) {
		return PerSex[int]{}, errRange
	}
	r, err := dbq.New(db).CalcFarmBase(ctx, dbq.CalcFarmBaseParams{FarmID: farmID, FromDay: cycleStart, ToDay: day})
	if err != nil {
		return PerSex[int]{}, fmt.Errorf("calc: base %s–%s: %w", cycleStart, day, err)
	}
	return pair(r.BaseF, r.BaseM), nil
}

// Lay is a farm's eggs and hens on one day (§5.6), over the egg units whose
// total is filled.
type Lay struct {
	Eggs      int // Σ total_eggs (Cat. 1 + 2 + 3 + Incubabile; Jos apart)
	HensStart int // Σ hens alive at the start of the day (D156)
}

// Pct is the lay % = 100 × eggs / hens at the start of the day, nil when
// there are no hen-days.
func (l Lay) Pct() *decimal.Decimal {
	return Pct(decimal.NewFromInt(int64(l.Eggs)), decimal.NewFromInt(int64(l.HensStart)))
}

// LayDay reads the farm's lay on day (§5.6).
func LayDay(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) (Lay, error) {
	r, err := dbq.New(db).CalcLayDay(ctx, dbq.CalcLayDayParams{FarmID: farmID, Day: day})
	if err != nil {
		return Lay{}, fmt.Errorf("calc: lay %s: %w", day, err)
	}
	return Lay{Eggs: int(r.Eggs), HensStart: int(r.HensStart)}, nil
}

// FeedKg is the feed the farm consumed on day per sex, in kg at full
// precision (§5.7): g × birds at the start of the day, or the migrated kg.
func FeedKg(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) (PerSex[decimal.Decimal], error) {
	r, err := dbq.New(db).CalcFeedKg(ctx, dbq.CalcFeedKgParams{FarmID: farmID, Day: day})
	if err != nil {
		return PerSex[decimal.Decimal]{}, fmt.Errorf("calc: feed kg %s: %w", day, err)
	}
	return PerSex[decimal.Decimal]{F: r.KgF, M: r.KgM}, nil
}

// Stock is the stock of one feed type at the end of a day (§5.8).
type Stock struct {
	FeedTypeID uuid.UUID
	Name       string
	LastMove   timex.Day       // the latest movement day on or before the asked day
	Kg         decimal.Decimal // full precision; negative when consumption outran deliveries
}

// StockEnd returns the farm's stock per feed type at the end of day, ordered
// by name. Feed types with no movement up to that day are absent.
func StockEnd(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]Stock, error) {
	rows, err := dbq.New(db).CalcStockEnd(ctx, dbq.CalcStockEndParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: stock %s: %w", day, err)
	}
	out := make([]Stock, 0, len(rows))
	for _, r := range rows {
		out = append(out, Stock{FeedTypeID: r.FeedTypeID, Name: r.Name, LastMove: r.Date, Kg: r.StockEnd})
	}
	return out, nil
}

func pair(f, m int32) PerSex[int] { return PerSex[int]{F: int(f), M: int(m)} }
