package calc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// SectionDeaths is one bird-holding section's recent deaths before a day, the
// inputs of the Zilnic form's Ieri and media 7 zile and of Z6 (§6.2). Deaths
// are mortality + culls (D165, plan M5 Q3); sales never count.
type SectionDeaths struct {
	SectionID uuid.UUID
	Code      string
	Yesterday PerSex[int] // the day before
	Week      PerSex[int] // Σ over the 7 days before (day − 7 … day − 1)
}

// DeathsWindow returns every bird-holding section of the farm (a section
// without children) with its deaths of the day before day and of the 7 days
// before it, 0 when it has none, in tree order (hall, then compartment).
func DeathsWindow(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]SectionDeaths, error) {
	rows, err := dbq.New(db).CalcDeathsWindow(ctx, dbq.CalcDeathsWindowParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: deaths window %s: %w", day, err)
	}
	out := make([]SectionDeaths, 0, len(rows))
	for _, r := range rows {
		out = append(out, SectionDeaths{
			SectionID: r.SectionID,
			Code:      r.Code,
			Yesterday: pair(r.YesterdayF, r.YesterdayM),
			Week:      pair(r.WeekF, r.WeekM),
		})
	}
	return out, nil
}

var seven = decimal.NewFromInt(7)

// Mean7 is Z6's mean: the deaths of the 7 days before a day ÷ 7, at full
// precision (6 → 0,857…). Callers compare at full precision and show one
// decimal (num.Round, §5.15); a 0 mean is 0, never "–" (D53).
func Mean7(week int) decimal.Decimal {
	return decimal.NewFromInt(int64(week)).Div(seven)
}

// EggUnitHens is one egg unit's hens at the start of a day (§5.6, D156).
type EggUnitHens struct {
	UnitID    uuid.UUID // the hall, or a section without parent and children
	Code      string
	HensStart int // Σ ♀ alive at the start of the day over the unit's compartments
}

// HensStart returns every egg unit of the farm (every section without a
// parent) with its hens at the start of day, 0 when it has none, in tree
// order.
func HensStart(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]EggUnitHens, error) {
	rows, err := dbq.New(db).CalcHensStart(ctx, dbq.CalcHensStartParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: hens at the start of %s: %w", day, err)
	}
	out := make([]EggUnitHens, 0, len(rows))
	for _, r := range rows {
		out = append(out, EggUnitHens{UnitID: r.EggUnitID, Code: r.Code, HensStart: int(r.HensStart)})
	}
	return out, nil
}
