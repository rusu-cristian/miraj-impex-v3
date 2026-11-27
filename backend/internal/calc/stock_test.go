package calc

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// T1 criterion 5: DaysLeft on the golden 28 figures (Hîrbovăț 18.11.2026, window 16–18.11) and the lamp edges.
func TestDaysLeft(t *testing.T) {
	nov18 := timex.MustDay(2026, time.November, 18)
	three := decimal.NewFromInt(3)
	five := decimal.NewFromInt(5)
	d := decimal.RequireFromString
	for _, tc := range []struct {
		name       string
		stock, avg decimal.Decimal
		days       string // rounded to 1 decimal, "nil" for none
		runOut     string // "nil" for none
		lamp       Status
		reason     StockReason
	}{
		{"Breeder 2: 31 664,9 ÷ 2 392,2", d("31664.9"), d("7176.6").Div(three), "13.2", "2026-12-01", InNorma, StockReasonNone},
		{"Cocoși: 1 781,0 ÷ 170,93…", d("1781.0"), d("512.8").Div(three), "10.4", "2026-11-28", InNorma, StockReasonNone},
		{"stock below 0: clamped at 0, runs out on the window end", d("-1240"), d("100"), "0.0", "2026-11-18", Abatere, StockNegative},
		{"stock below 0 wins over no consumption", d("-1"), decimal.Zero, "0.0", "2026-11-18", Abatere, StockNegative},
		{"4,96 days: below 5, floor 4", d("496"), d("100"), "5.0", "2026-11-22", Abatere, StockBelowMinDays},
		{"4,9999 days shows 5,0 but is below 5 at full precision", d("499.99"), d("100"), "5.0", "2026-11-22", Abatere, StockBelowMinDays},
		{"exactly 5 days is not below the threshold", d("500"), d("100"), "5.0", "2026-11-23", InNorma, StockReasonNone},
		{"stock 0 with consumption: 0 days", decimal.Zero, d("10"), "0.0", "2026-11-18", Abatere, StockBelowMinDays},
		{"no consumption", d("1000"), decimal.Zero, "nil", "nil", Oprit, StockNoConsumption},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := DaysLeft(tc.stock, tc.avg, nov18, five)
			days, runOut := "nil", "nil"
			if got.Days != nil {
				days = got.Days.Round(1).StringFixed(1)
			}
			if got.RunOut != nil {
				runOut = got.RunOut.String()
			}
			if days != tc.days || runOut != tc.runOut || got.Lamp != tc.lamp || got.Reason != tc.reason {
				t.Errorf("DaysLeft = {%s, %s, %q, %q}, want {%s, %s, %q, %q}",
					days, runOut, got.Lamp, got.Reason, tc.days, tc.runOut, tc.lamp, tc.reason)
			}
			if (got.Days == nil) != (got.RunOut == nil) {
				t.Errorf("Days and RunOut must be nil together: %+v", got)
			}
		})
	}
	// Full precision: 4,96 days stays 4,96 (not rounded inside DaysLeft).
	if got := DaysLeft(d("496"), d("100"), nov18, five); !got.Days.Equal(d("4.96")) {
		t.Errorf("Days = %s, want 4.96", got.Days)
	}
	// Breeder 2 at full precision: 13,237… (golden 28 rounds it once).
	if got := DaysLeft(d("31664.9"), d("2392.2"), nov18, five); got.Days.Round(3).String() != "13.237" {
		t.Errorf("Days = %s, want 13.237…", got.Days)
	}
}

func TestLowest(t *testing.T) {
	d := decimal.RequireFromString
	if _, ok := Lowest(nil); ok {
		t.Error("Lowest(nil) is ok, want false")
	}
	s := []StockPoint{{jan(11), d("700")}, {jan(12), d("-5")}, {jan(13), d("-5")}, {jan(14), d("3")}}
	if got, ok := Lowest(s); !ok || got.Day != jan(12) || !got.Kg.Equal(d("-5")) {
		t.Errorf("Lowest = %+v, %v; want the first −5 (12.01)", got, ok)
	}
}

// M3 ARC-1: the one wording of a stock row's reason, shared by the Stoc
// report, the summary's feedDaysLeft and the Hrană zilnică form state.
func TestStockReasonText(t *testing.T) {
	settings := AppSettings{StockDaysMin: decimal.NewFromInt(5), StockAvgDays: 3}
	for r, want := range map[StockReason]string{
		StockReasonNone:    "",
		StockNegative:      "stoc sub 0 · verifică livrările",
		StockBelowMinDays:  "sub pragul de 5 zile",
		StockNoConsumption: "– fără consum în ultimele 3 zile",
		"unknown":          "",
	} {
		if got := r.Text(settings); got != want {
			t.Errorf("%q.Text = %q, want %q", r, got, want)
		}
	}
}

func TestStockForSexText(t *testing.T) {
	f := dbq.SexF
	if got := (Stock{}).ForSexText(); got != nil {
		t.Errorf("ambele sexe: ForSexText = %q, want nil", *got)
	}
	if got := (Stock{ForSex: &f}).ForSexText(); got == nil || *got != "F" {
		t.Errorf("♀: ForSexText = %v, want F", got)
	}
}
