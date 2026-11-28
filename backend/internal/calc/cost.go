package calc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// EggsBetween is the farm's Σ eggs over r (§5.6, §5.9 lei_per_egg): the live
// egg rows whose total is filled (Cat. 1 + 2 + 3 + Incubabile, Jos apart). A
// nil r.From counts from the first egg row.
func EggsBetween(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, r DayRange) (int, error) {
	if r.From != nil && r.From.After(r.To) {
		return 0, errRange
	}
	n, err := dbq.New(db).CalcEggsBetween(ctx, dbq.CalcEggsBetweenParams{FarmID: farmID, FromDay: r.From, ToDay: r.To})
	if err != nil {
		return 0, fmt.Errorf("calc: eggs up to %s: %w", r.To, err)
	}
	return int(n), nil
}

// FarmFilter limits a reader to some farms: every farm when All, else the
// listed IDs only (none when IDs is empty). Pass the caller's: the owner sees
// every farm, anyone else his user_farm farms.
type FarmFilter struct {
	All bool
	IDs []uuid.UUID
}

// AllFarms is the FarmFilter of every farm.
var AllFarms = FarmFilter{All: true}

// SeriesPlacement is the series placement per sex (§5.9 series total, §7.3
// series page): the live placement rows dated inside a cycle of the series at
// their farm, over the farms only admits (the migrated placements have no
// series on their entry; the cycle links them). Seria 2026: ♀ 28 080 · ♂ 3 510 at
// Telița. A series without placement rows on those farms reads 0 · 0.
func SeriesPlacement(ctx context.Context, db dbq.DBTX, seriesID uuid.UUID, only FarmFilter) (PerSex[int], error) {
	ids := only.IDs
	if ids == nil {
		ids = []uuid.UUID{}
	}
	r, err := dbq.New(db).CalcSeriesPlacement(ctx, dbq.CalcSeriesPlacementParams{
		SeriesID: seriesID, AllFarms: only.All, FarmIds: ids,
	})
	if err != nil {
		return PerSex[int]{}, fmt.Errorf("calc: series placement: %w", err)
	}
	return pair(r.PlacedF, r.PlacedM), nil
}

// EggReason says why a cost has no lei per egg (§5.9). The caller words it
// (messages).
type EggReason string

// EggReason values.
const (
	EggReasonNone EggReason = ""
	// EggReasonGrowth: a growth cycle lays no eggs, "creștere, fără ouă".
	EggReasonGrowth EggReason = "growth"
	// EggReasonNoEggs: a laying cycle with no eggs in the range yet.
	EggReasonNoEggs EggReason = "no_eggs"
)

// Cost is a cycle's (or a series') feed cost per bird and per egg at full
// precision (§5.9). Round once, to 2 decimals, for the wire.
type Cost struct {
	PerBird   *decimal.Decimal // lei ÷ birds; nil when birds is 0
	PerEgg    *decimal.Decimal // lei ÷ eggs; nil for growth or no eggs (EggReason says which)
	EggReason EggReason
}

// CostOf is §5.9's lei_per_bird = lei ÷ birds (the birds at the cycle start,
// or the series placement) and lei_per_egg = lei ÷ eggs. A growth cycle has
// no lei per egg (EggReasonGrowth) whatever eggs says; 0 eggs gives
// EggReasonNoEggs; 0 birds gives a nil PerBird (a rate with no base is "–").
func CostOf(lei decimal.Decimal, birds, eggs int, growth bool) Cost {
	var c Cost
	if birds > 0 {
		v := lei.Div(decimal.NewFromInt(int64(birds)))
		c.PerBird = &v
	}
	switch {
	case growth:
		c.EggReason = EggReasonGrowth
	case eggs <= 0:
		c.EggReason = EggReasonNoEggs
	default:
		v := lei.Div(decimal.NewFromInt(int64(eggs)))
		c.PerEgg = &v
	}
	return c
}
