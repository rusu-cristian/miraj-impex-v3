// Package fmtx formats numbers, dates and labels as SPEC §11.3 prescribes, for
// the few strings the server writes itself (Toasts, messages, voice answers,
// xlsx head lines). The client applies the same rules to raw API numbers.
package fmtx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/shopspring/decimal"
)

const (
	// ThinSpace is the thousands separator: U+202F narrow no-break space.
	ThinSpace = "\u202f"
	// Minus is the true minus sign U+2212.
	Minus = "\u2212"
	// EnDash joins the two ends of a date range, without spaces.
	EnDash = "\u2013"
)

// Int formats a count: "14 080" (U+202F), "−82" (U+2212).
func Int(n int64) string {
	neg := n < 0
	digits := strconv.FormatInt(n, 10)
	if neg {
		digits = digits[1:]
	}
	return sign(neg, false) + group(digits)
}

// Dec formats d rounded half away from zero to exactly places decimals, with
// a decimal comma and grouped thousands: Dec(2.41, 2) = "2,41",
// Dec(2457.9, 1) = "2 457,9", Dec(−82, 0) = "−82".
func Dec(d decimal.Decimal, places int32) string {
	return formatDec(d, places, false)
}

// Signed is Dec with a "+" on positive values, for differences: "+7,7".
// Zero (after rounding) has no sign.
func Signed(d decimal.Decimal, places int32) string {
	return formatDec(d, places, true)
}

// Percent formats a percentage with 1 decimal and no space: "41,1%".
func Percent(d decimal.Decimal) string { return Dec(d, 1) + "%" }

// PercentDiff formats a percentage difference with its sign: "+7,7%", "−1,4%".
func PercentDiff(d decimal.Decimal) string { return Signed(d, 1) + "%" }

// Points formats a difference in percentage points: "−16,5 pct".
func Points(d decimal.Decimal) string { return Signed(d, 1) + " pct" }

func formatDec(d decimal.Decimal, places int32, plus bool) string {
	r := num.Round(d, places)
	neg := r.Sign() < 0
	text := r.Abs().StringFixed(places)
	intPart, frac, hasFrac := strings.Cut(text, ".")
	out := sign(neg, plus && r.Sign() > 0) + group(intPart)
	if hasFrac {
		out += "," + frac
	}
	return out
}

func sign(neg, plus bool) string {
	switch {
	case neg:
		return Minus
	case plus:
		return "+"
	default:
		return ""
	}
}

// group inserts U+202F every three digits from the right.
func group(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var b strings.Builder
	head := len(digits) % 3
	if head > 0 {
		b.WriteString(digits[:head])
	}
	for i := head; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteString(ThinSpace)
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

// Day formats a day as "18.11.2026".
func Day(d timex.Day) string {
	y, m, dd := d.Date()
	return fmt.Sprintf("%02d.%02d.%04d", dd, int(m), y)
}

// dayMonth formats "18.11".
func dayMonth(d timex.Day) string {
	_, m, dd := d.Date()
	return fmt.Sprintf("%02d.%02d", dd, int(m))
}

// DateTime formats an instant in Europe/Chisinau as "18.11.2026, 13:16" (D129).
func DateTime(t time.Time) string {
	local := t.In(timex.Chisinau)
	return Day(timex.DayOf(local)) + ", " + local.Format("15:04")
}

// Range formats a day range with an en dash and no spaces; the start drops
// its year only when the end has the same year: "10.11–16.11.2026",
// "07.10.2025–28.02.2026".
func Range(from, to timex.Day) string {
	fy, _, _ := from.Date()
	ty, _, _ := to.Date()
	if fy == ty {
		return dayMonth(from) + EnDash + Day(to)
	}
	return Day(from) + EnDash + Day(to)
}

// RangeNoYear formats an in-context window that drops both years:
// "16.11–18.11".
func RangeNoYear(from, to timex.Day) string {
	return dayMonth(from) + EnDash + dayMonth(to)
}

// WeekLabel is the week in labels, column heads and Mono context lines:
// "SAPT 59".
func WeekLabel(week int) string { return "SAPT " + strconv.Itoa(week) }

// WeekText is the week in running text, Toasts and notes: "sapt. 59".
func WeekText(week int) string { return "sapt. " + strconv.Itoa(week) }

// SexWord is the plural sex word: "F" → "găini", "M" → "cocoși". It takes any
// string type (e.g. the sqlc enum dbq.Sex) and panics on another value, which
// is a programming error: sex values are checked at the boundary.
func SexWord[S ~string](sex S) string {
	switch sex {
	case "F":
		return "găini"
	case "M":
		return "cocoși"
	default:
		panic(fmt.Sprintf("fmtx: unknown sex %q", string(sex)))
	}
}
