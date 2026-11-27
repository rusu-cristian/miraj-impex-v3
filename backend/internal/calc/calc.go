// Package calc holds the derived calculations of SPEC §5: pure formulas over
// plain values (age, week, percentages) and the SQL readers they need
// (db/queries/calc.sql).
//
// Every formula works at full precision; callers round once, at the end, with
// num.Round (half away from zero, §5.15). Readers take a dbq.DBTX, so they run
// on a pool or inside a transaction alike, and always take the farm and the day
// or day range explicitly: none reads app_today(). Bird counts are running sums
// of the live bird events up to the asked day, so they don't depend on how far
// v_section_sex_day's dense day list reaches.
package calc

import (
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// PerSex holds one value for hens (F, găini) and one for roosters (M, cocoși).
type PerSex[T any] struct {
	F T
	M T
}

// SeriesAge is what the age formulas of §5.1 need from a series.
type SeriesAge struct {
	PlacementDate      timex.Day // the day of the first placement
	AgeAtPlacementDays int       // 0–140 (B3)
}

// AgeDays is the birds' age in days on d (§5.1): 0 on the placement day when
// they were placed at age 0. It is negative before the placement.
func AgeDays(s SeriesAge, d timex.Day) int {
	return s.AgeAtPlacementDays + d.DaysSince(s.PlacementDate)
}

// SAPT is the birds' age week on d (§5.1): floor(age_days / 7) + 1, so 13
// days → 2 and 14 days → 3. Like the SQL sapt(), it floors negative ages too.
func SAPT(s SeriesAge, d timex.Day) int {
	return floorDiv(AgeDays(s, d), 7) + 1
}

// WeekFirstDay is the first day of age week k (§5.1).
func WeekFirstDay(s SeriesAge, k int) timex.Day {
	return s.PlacementDate.AddDays(7*(k-1) - s.AgeAtPlacementDays)
}

// WeekLastDay is the last day of age week k (§5.1): the first day + 6.
func WeekLastDay(s SeriesAge, k int) timex.Day {
	return WeekFirstDay(s, k).AddDays(6)
}

// floorDiv is a / b rounded toward negative infinity (b > 0).
func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && a < 0 {
		q--
	}
	return q
}

var hundred = decimal.NewFromInt(100)

// Pct is 100 × part / whole at full precision, or nil when whole is 0 (a rate
// with no base is "–", never 0). Round the result once, with num.Round.
func Pct(part, whole decimal.Decimal) *decimal.Decimal {
	if whole.IsZero() {
		return nil
	}
	p := part.Mul(hundred).Div(whole)
	return &p
}

// CumMortalityPct is the cumulative mortality % of §5.5: 100 × deaths / base,
// nil when the base is 0. Deaths are mortality + culls, never sales.
func CumMortalityPct(deaths, base int) *decimal.Decimal {
	return Pct(decimal.NewFromInt(int64(deaths)), decimal.NewFromInt(int64(base)))
}
