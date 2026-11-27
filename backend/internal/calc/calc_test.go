package calc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

func day(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }

// Seria 2026: placed 07.10.2025 (a Tuesday) at age 0 (SPEC §5.1).
var seria2026 = SeriesAge{PlacementDate: day(2025, time.October, 7)}

// The same series after the §5.1 age change 0 → 7.
var seria2026Age7 = SeriesAge{PlacementDate: day(2025, time.October, 7), AgeAtPlacementDays: 7}

func TestAgeDaysAndSAPT(t *testing.T) {
	for _, tc := range []struct {
		name      string
		s         SeriesAge
		d         timex.Day
		age, sapt int
	}{
		{"golden 1: 18.11.2026", seria2026, day(2026, time.November, 18), 407, 59},
		{"golden 2: 24.02.2026", seria2026, day(2026, time.February, 24), 140, 21},
		{"golden 2: Hîrbovăț cycle start 22.02.2026", seria2026, day(2026, time.February, 22), 138, 20},
		{"golden 3: 13 days", seria2026, day(2025, time.October, 20), 13, 2},
		{"golden 3: 14 days", seria2026, day(2025, time.October, 21), 14, 3},
		{"placement day", seria2026, day(2025, time.October, 7), 0, 1},
		{"6 days", seria2026, day(2025, time.October, 13), 6, 1},
		{"7 days", seria2026, day(2025, time.October, 14), 7, 2},
		{"day before placement floors", seria2026, day(2025, time.October, 6), -1, 0},
		{"7 days before placement", seria2026, day(2025, time.September, 30), -7, 0},
		{"8 days before placement", seria2026, day(2025, time.September, 29), -8, -1},
		{"age change 0 → 7: 18.11.2026", seria2026Age7, day(2026, time.November, 18), 414, 60},
		{"age 7 on the placement day", seria2026Age7, day(2025, time.October, 7), 7, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := AgeDays(tc.s, tc.d); got != tc.age {
				t.Errorf("AgeDays = %d, want %d", got, tc.age)
			}
			if got := SAPT(tc.s, tc.d); got != tc.sapt {
				t.Errorf("SAPT = %d, want %d", got, tc.sapt)
			}
		})
	}
}

func TestWeekBounds(t *testing.T) {
	for _, tc := range []struct {
		name        string
		s           SeriesAge
		k           int
		first, last timex.Day
	}{
		{"golden 4: SAPT 59 at age 0", seria2026, 59, day(2026, time.November, 17), day(2026, time.November, 23)},
		{"SAPT 1 at age 0", seria2026, 1, day(2025, time.October, 7), day(2025, time.October, 13)},
		{"SAPT 20 at age 0", seria2026, 20, day(2026, time.February, 17), day(2026, time.February, 23)},
		{"the same days are SAPT 60 at age 7", seria2026Age7, 60, day(2026, time.November, 17), day(2026, time.November, 23)},
		{"SAPT 59 at age 7", seria2026Age7, 59, day(2026, time.November, 10), day(2026, time.November, 16)},
		{"SAPT 1 at age 7 is before the placement", seria2026Age7, 1, day(2025, time.September, 30), day(2025, time.October, 6)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := WeekFirstDay(tc.s, tc.k); got != tc.first {
				t.Errorf("WeekFirstDay(%d) = %s, want %s", tc.k, got, tc.first)
			}
			if got := WeekLastDay(tc.s, tc.k); got != tc.last {
				t.Errorf("WeekLastDay(%d) = %s, want %s", tc.k, got, tc.last)
			}
		})
	}
	if wd := WeekFirstDay(seria2026, 59).Weekday(); wd != time.Tuesday {
		t.Errorf("SAPT 59 starts on %s, want Tuesday", wd)
	}

	// Every day lies inside the week SAPT gives it, for both ages.
	for _, s := range []SeriesAge{seria2026, seria2026Age7} {
		for d := day(2025, time.September, 1); d.Before(day(2027, time.January, 1)); d = d.AddDays(1) {
			k := SAPT(s, d)
			if first, last := WeekFirstDay(s, k), WeekLastDay(s, k); d.Before(first) || d.After(last) {
				t.Fatalf("age %d: %s is SAPT %d but that week is %s–%s", s.AgeAtPlacementDays, d, k, first, last)
			}
		}
	}
}

func rounded(t *testing.T, p *decimal.Decimal, places int32) string {
	t.Helper()
	if p == nil {
		t.Fatal("got nil, want a value")
	}
	return num.Round(*p, places).String()
}

func TestCumMortalityPct(t *testing.T) {
	// Golden 10 (Hîrbovăț to 18.11.2026).
	f := CumMortalityPct(1521, 26883)
	if got := rounded(t, f, 2); got != "5.66" {
		t.Errorf("♀ at 2 dp = %s, want 5.66", got)
	}
	if got := rounded(t, f, 1); got != "5.7" {
		t.Errorf("♀ at 1 dp = %s, want 5.7", got)
	}
	if got := rounded(t, CumMortalityPct(1048, 3024), 2); got != "34.66" {
		t.Errorf("♂ at 2 dp = %s, want 34.66", got)
	}
	if got := rounded(t, CumMortalityPct(0, 10), 2); got != "0" {
		t.Errorf("no deaths = %s, want 0", got)
	}
	if got := CumMortalityPct(5, 0); got != nil {
		t.Errorf("base 0 = %s, want nil", got)
	}
}

func TestPct(t *testing.T) {
	// Golden 23 / §5.15: round once, from full precision.
	p := Pct(decimal.NewFromInt(168122), decimal.NewFromInt(186088))
	if got := rounded(t, p, 1); got != "90.3" {
		t.Errorf("168122 / 186088 at 1 dp = %s, want 90.3", got)
	}
	if twice := num.Round(num.Round(*p, 2), 1).String(); twice != "90.4" {
		t.Errorf("rounding twice = %s; the test no longer shows the §5.15 trap", twice)
	}
	// Half away from zero.
	if got := rounded(t, Pct(decimal.NewFromInt(1), decimal.NewFromInt(8)), 0); got != "13" {
		t.Errorf("12.5 at 0 dp = %s, want 13", got)
	}
	if got := rounded(t, Pct(decimal.NewFromInt(-1), decimal.NewFromInt(8)), 0); got != "-13" {
		t.Errorf("-12.5 at 0 dp = %s, want -13", got)
	}
	if got := Pct(decimal.NewFromInt(1), decimal.Zero); got != nil {
		t.Errorf("whole 0 = %s, want nil", got)
	}
}

func TestLayPct(t *testing.T) {
	// Golden 12: 5 914 ÷ 14 395 = 41,1 %.
	if got := rounded(t, Lay{Eggs: 5914, HensStart: 14395}.Pct(), 1); got != "41.1" {
		t.Errorf("lay = %s, want 41.1", got)
	}
	if got := (Lay{}).Pct(); got != nil {
		t.Errorf("no hen-days = %s, want nil", got)
	}
}

func TestRangeStartsAfterEnd(t *testing.T) {
	// Refused before any query, so no database is needed.
	from, to := day(2026, time.November, 18), day(2026, time.November, 12)
	if _, err := Sold(t.Context(), nil, uuid.Nil, from, to); err != errRange {
		t.Errorf("Sold: err = %v, want errRange", err)
	}
	if _, err := CycleDeaths(t.Context(), nil, uuid.Nil, from, to); err != errRange {
		t.Errorf("CycleDeaths: err = %v, want errRange", err)
	}
	if _, err := FarmBase(t.Context(), nil, uuid.Nil, from, to); err != errRange {
		t.Errorf("FarmBase: err = %v, want errRange", err)
	}
	if _, err := SectionsRemovals(t.Context(), nil, uuid.Nil, from, to); err != errRange {
		t.Errorf("SectionsRemovals: err = %v, want errRange", err)
	}
}
