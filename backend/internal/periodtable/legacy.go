package periodtable

import (
	"context"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// A past cycle without a flock count (Seria 2025 at Hîrbovăț, D52, D54): the
// register has eggs and deaths per compartment but no birds alive, so the
// Rezumat has no vii, no % ouat and no guide, and its deaths come from
// legacy_mortality.

// TableCycle is the cycle as the PeriodTable counts its weeks, which the
// summary and WeekSeries share so that they agree with the table. For a past
// cycle without a flock count (legacy true) its start moves to the first
// register day, so its weeks and "din N" count the register's weeks (M2 Q7:
// SAPT 23–65, 43 weeks, where the cycle itself starts in SAPT 19); without
// any register record it stays as it is (its period then has no records).
// Any other cycle is returned as it is, without a read.
func TableCycle(ctx context.Context, db dbq.DBTX, c calc.CycleInfo) (cycle calc.Cycle, legacy bool, err error) {
	cycle = c.Cycle
	if c.HasBirdEvents || c.EndDate == nil {
		return cycle, false, nil
	}
	start, err := registerStart(ctx, db, c.FarmID, c.StartDate, *c.EndDate)
	if err != nil {
		return calc.Cycle{}, false, err
	}
	if start != nil {
		cycle.StartDate = *start
	}
	return cycle, true, nil
}

// legacyDeathCells writes Morți from the register (§5.5): Σ deaths of the
// rows that are not emptying. A row with only emptying rows for a sex is null
// with the reason "emptying" ("– golire, nu mortalitate"): removals when the
// halls were emptied are never mortality (D52, D54, D92). A row with no
// register row at all is null.
func legacyDeathCells(w *cellWriter, a acc) {
	for _, s := range []struct {
		key                   string
		deaths, rows, emptied int
	}{
		{KeyDeathsF, a.legacyDeaths.F, a.legacyRows.F, a.emptyingRows.F},
		{KeyDeathsM, a.legacyDeaths.M, a.legacyRows.M, a.emptyingRows.M},
	} {
		switch {
		case s.rows > 0:
			w.set(s.key, numInt(s.deaths))
		case s.emptied > 0:
			w.set(s.key, nil)
			w.reason(s.key, ReasonEmptying)
		default:
			w.set(s.key, nil)
		}
	}
}
