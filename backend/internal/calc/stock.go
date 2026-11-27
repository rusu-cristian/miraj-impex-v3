package calc

import (
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Oprit is the lamp of a figure that is not compared: no consumption to
// divide by, or a farm without an open cycle (§5.8).
const Oprit Status = "oprit"

// StockReason says why a stock row's lamp is not InNorma (§5.8). The
// caller words it (messages); InNorma rows have StockReasonNone.
type StockReason string

// StockReason values, in the order DaysLeft checks them.
const (
	StockReasonNone StockReason = ""
	// StockNegative: the stock is below 0, "stoc sub 0 · verifică livrările".
	StockNegative StockReason = "stock_negative"
	// StockBelowMinDays: days left below alert.stock_days_min, "sub pragul de
	// {n} zile".
	StockBelowMinDays StockReason = "below_min_days"
	// StockNoConsumption: no consumption over the window, "– fără consum în
	// ultimele {n} zile".
	StockNoConsumption StockReason = "no_consumption"
)

// Text words the reason as a stock row's "reason" (§5.8), with the settings
// it names; "" for StockReasonNone. The Stoc report, the summary's
// feedDaysLeft and the Hrană zilnică form state all word it here.
func (r StockReason) Text(s AppSettings) string {
	switch r {
	case StockNegative:
		return messages.StockBelowZero()
	case StockBelowMinDays:
		return messages.StockBelowDays(s.StockDaysMin)
	case StockNoConsumption:
		return messages.StockNoConsumption(s.StockAvgDays)
	}
	return ""
}

// Window is the day range of the average daily consumption (§5.8): the last
// alert.stock_avg_days complete days, both ends included.
type Window struct {
	From timex.Day
	To   timex.Day
}

// DaysLeftResult is the days left of one feed type and its lamp (§5.8).
type DaysLeftResult struct {
	// Days is stock ÷ avg at full precision, 0 when the stock is below 0, nil
	// when there is no consumption to divide by. Round it once, to 1 decimal.
	Days *decimal.Decimal
	// RunOut is the window end + ⌊Days⌋ days, nil with Days.
	RunOut *timex.Day
	Lamp   Status
	Reason StockReason
}

// DaysLeft is §5.8's days left and run-out date of a stock against the
// average daily consumption avg over a window ending on windowEnd. Checked in
// order, at full precision:
//   - stock < 0 → Abatere (StockNegative), Days 0, RunOut = windowEnd ("se
//     termină azi"), even when avg is 0;
//   - avg = 0 → Oprit (StockNoConsumption), Days and RunOut nil;
//   - Days < minDays (alert.stock_days_min) → Abatere (StockBelowMinDays);
//   - else InNorma.
func DaysLeft(stock, avg decimal.Decimal, windowEnd timex.Day, minDays decimal.Decimal) DaysLeftResult {
	if stock.IsNegative() {
		days, end := decimal.Zero, windowEnd
		return DaysLeftResult{Days: &days, RunOut: &end, Lamp: Abatere, Reason: StockNegative}
	}
	if avg.Sign() <= 0 {
		return DaysLeftResult{Lamp: Oprit, Reason: StockNoConsumption}
	}
	days := stock.Div(avg)
	runOut := windowEnd.AddDays(int(days.Floor().IntPart()))
	r := DaysLeftResult{Days: &days, RunOut: &runOut, Lamp: InNorma}
	if days.LessThan(minDays) {
		r.Lamp, r.Reason = Abatere, StockBelowMinDays
	}
	return r
}
