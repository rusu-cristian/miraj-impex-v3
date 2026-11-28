package reports

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// costCandidate is a farm whose shown cycle gets a Cost row.
type costCandidate struct {
	farm  farmScope
	cycle calc.CycleInfo
	meta  cycleMeta
	dels  calc.FarmDeliveries
}

// buildCost builds the CostBlock (§5.9, M4 Q6) from the farms' deliveries
// (dels, in the scope's farm order):
//   - one row per farm whose shown cycle has priced deliveries in its range,
//     ordered by cycle start, then farm name (Telița, then Hîrbovăț);
//   - lei per bird = the priced lei ÷ the birds at the cycle start (the §5.5
//     farm base over the cycle up to today, calc.FarmBase);
//   - lei per egg = lei ÷ the farm's eggs in the range (calc.EggsBetween);
//     none for a growth cycle ("creștere, fără ouă") or without eggs ("fără
//     ouă în perioadă");
//   - the total only when every row is of one series and the rows cover
//     every farm with a cycle of it (a manager of one farm never gets a total
//     built from farms he can't see): Σ lei ÷ the series placement
//     (calc.SeriesPlacement) and ÷ Σ eggs of the rows.
//
// The bases are read in one statement (ReportsCostInputs), only when there is
// a row.
func buildCost(ctx context.Context, db dbq.DBTX, sc scope, dels []calc.FarmDeliveries, today timex.Day) (Cost, error) {
	res := Cost{AsOf: today, Cycles: []CostCycle{}}
	var cands []costCandidate
	for i, f := range sc.farms {
		shown := f.shown()
		if shown == nil || dels[i].Priced.Count == 0 {
			continue
		}
		cands = append(cands, costCandidate{farm: f, cycle: *shown, meta: f.meta[shown.ID], dels: dels[i]})
	}
	if len(cands) == 0 {
		return res, nil
	}
	slices.SortStableFunc(cands, func(a, b costCandidate) int { return a.cycle.StartDate.Compare(b.cycle.StartDate) })

	var seriesID *uuid.UUID
	if s := cands[0].cycle.SeriesID; len(cands) == cands[0].meta.SeriesFarms &&
		!slices.ContainsFunc(cands, func(c costCandidate) bool { return c.cycle.SeriesID != s }) {
		seriesID = &s
	}
	inputs, err := readCostInputs(ctx, db, cands, seriesID, today)
	if err != nil {
		return res, err
	}

	var (
		totalLei, totalKg, excluded = decimal.Zero, decimal.Zero, decimal.Zero
		totalEggs                   int
		allGrowth                   = true
	)
	for i, c := range cands {
		in := inputs[i]
		growth := c.meta.Type == dbq.FarmTypeGrowth
		birds := httpx.SexPair[int]{F: int(in.BaseF), M: int(in.BaseM)}
		eggs := int(in.Eggs)
		cost := calc.CostOf(c.dels.Priced.Lei, birds.F+birds.M, eggs, growth)
		row := CostCycle{
			Label:       messages.CostRowLabel(c.farm.Name, c.cycle.SeriesName),
			Span:        costSpan(c.cycle, today),
			PricedCount: c.dels.Priced.Count,
			Kg:          num.Of(c.dels.Priced.Kg),
			Lei:         num.Of(c.dels.Priced.Lei),
			Birds:       birds,
			LeiPerBird:  optNum(cost.PerBird),
			LeiPerEgg:   optNum(cost.PerEgg),
			EggReason:   eggReasonText(cost.EggReason),
		}
		if !growth {
			row.Eggs = &eggs
			totalEggs += eggs
			allGrowth = false
		}
		if first, last, ok := c.dels.Unpriced.Span(); ok {
			row.Unpriced = &CostUnpriced{
				Count: c.dels.Unpriced.Count, Kg: num.Of(c.dels.Unpriced.Kg), Span: messages.UnpricedSpan(first, last),
			}
		}
		res.Cycles = append(res.Cycles, row)
		totalLei, totalKg = totalLei.Add(c.dels.Priced.Lei), totalKg.Add(c.dels.Priced.Kg)
		excluded = excluded.Add(c.dels.Unpriced.Kg)
	}
	if seriesID != nil {
		placed := httpx.SexPair[int]{F: int(inputs[0].PlacedF), M: int(inputs[0].PlacedM)}
		cost := calc.CostOf(totalLei, placed.F+placed.M, totalEggs, allGrowth)
		res.Total = &CostTotal{
			Label: messages.CostSeriesTotal(cands[0].cycle.SeriesName), Kg: num.Of(totalKg), Lei: num.Of(totalLei),
			Birds: placed, Eggs: totalEggs, LeiPerBird: optNum(cost.PerBird), LeiPerEgg: optNum(cost.PerEgg),
			EggReason: eggReasonText(cost.EggReason), ExcludedKg: num.Of(excluded),
		}
	}
	return res, nil
}

// readCostInputs reads each candidate's farm base over its cycle up to today,
// its eggs over its delivery range and, with seriesID, the series placement:
// one row per candidate, in order.
func readCostInputs(ctx context.Context, db dbq.DBTX, cands []costCandidate, seriesID *uuid.UUID, today timex.Day) ([]dbq.ReportsCostInputsRow, error) {
	ranges := make([]calc.FarmRange, len(cands))
	p := dbq.ReportsCostInputsParams{
		CycleStarts: make([]timex.Day, len(cands)), BaseTos: make([]timex.Day, len(cands)), SeriesID: seriesID,
	}
	for i, c := range cands {
		ranges[i] = c.dels.FarmRange
		p.CycleStarts[i], p.BaseTos[i] = c.cycle.StartDate, today
		if c.cycle.EndDate != nil && c.cycle.EndDate.Before(today) {
			p.BaseTos[i] = *c.cycle.EndDate
		}
	}
	p.FarmIds, p.HasEggFrom, p.EggFroms, p.EggTos = rangeArrays(ranges)
	rows, err := dbq.New(db).ReportsCostInputs(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("reports: cost inputs: %w", err)
	}
	if len(rows) != len(cands) {
		return nil, fmt.Errorf("reports: cost inputs: %d rows for %d cycles", len(rows), len(cands))
	}
	return rows, nil
}

// costSpan is a cost row's span: "07.10.2025–28.02.2026" for a closed cycle,
// "din 22.02.2026 · SAPT 59" for the open one.
func costSpan(c calc.CycleInfo, today timex.Day) string {
	if c.EndDate != nil {
		return messages.CostSpanClosed(c.StartDate, *c.EndDate)
	}
	return messages.CostSpanOpen(c.StartDate, calc.SAPT(c.Age, today))
}

// eggReasonText words calc's reason for a missing lei per egg.
func eggReasonText(r calc.EggReason) string {
	switch r {
	case calc.EggReasonGrowth:
		return messages.CostNoEggsGrowth()
	case calc.EggReasonNoEggs:
		return messages.CostNoEggs()
	}
	return ""
}

func optNum(d *decimal.Decimal) *num.Num {
	if d == nil {
		return nil
	}
	n := num.Of(*d)
	return &n
}
