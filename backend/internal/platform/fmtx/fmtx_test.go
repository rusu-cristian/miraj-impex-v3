package fmtx

import (
	"strings"
	"testing"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

// The expected strings spell out U+202F, U+2212 and U+2013 so the test can't
// be fooled by a look-alike character.
func TestNumbers(t *testing.T) {
	for _, tc := range []struct {
		name, got, want string
	}{
		{"thousands", Int(14080), "14\u202f080"},
		{"thousands 1 098", Int(1098), "1\u202f098"},
		{"millions", Int(-1234567), "\u22121\u202f234\u202f567"},
		{"small", Int(407), "407"},
		{"zero", Int(0), "0"},
		{"minus", Int(-82), "\u221282"},
		{"decimal comma", Dec(dec("2.41"), 2), "2,41"},
		{"decimal rounds", Dec(dec("2.405"), 2), "2,41"},
		{"decimal keeps places", Dec(dec("7.7"), 2), "7,70"},
		{"kg grouped", Dec(dec("2457.94"), 1), "2\u202f457,9"},
		{"dec minus", Dec(dec("-82"), 0), "\u221282"},
		{"dec rounds to zero, no sign", Dec(dec("-0.04"), 1), "0,0"},
		{"percent", Percent(dec("41.149")), "41,1%"},
		{"percent diff plus", PercentDiff(dec("7.65")), "+7,7%"},
		{"percent diff minus", PercentDiff(dec("-1.35")), "\u22121,4%"},
		{"percent diff zero", PercentDiff(dec("0.04")), "0,0%"},
		{"points", Points(dec("-16.5")), "\u221216,5 pct"},
		{"signed plus", Signed(dec("1500"), 0), "+1\u202f500"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestSpecExamples(t *testing.T) {
	// The examples of SPEC §11.3. The SPEC text shows "14 080" with a plain
	// space; §11.3 names U+202F as the separator, which is what is written.
	for _, tc := range []struct{ got, want string }{
		{Int(14080), "14\u202f080"},
		{Int(-82), "−82"},
		{PercentDiff(dec("7.7")), "+7,7%"},
		{Percent(dec("41.1")), "41,1%"},
		{Points(dec("-16.5")), "−16,5 pct"},
		{Dec(dec("2.41"), 2), "2,41"},
		{Day(timex.MustDay(2026, 11, 18)), "18.11.2026"},
		{Range(timex.MustDay(2026, 11, 10), timex.MustDay(2026, 11, 16)), "10.11–16.11.2026"},
		{Range(timex.MustDay(2025, 10, 7), timex.MustDay(2026, 2, 28)), "07.10.2025–28.02.2026"},
		{RangeNoYear(timex.MustDay(2026, 11, 16), timex.MustDay(2026, 11, 18)), "16.11–18.11"},
		{WeekLabel(59), "SAPT 59"},
		{WeekText(59), "sapt. 59"},
		{WeekRange(56, 59), "SAPT 56–59"},
		{WeekRange(58, 58), "SAPT 58"},
		{DayMonth(timex.MustDay(2026, 2, 3)), "03.02"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

func TestDates(t *testing.T) {
	winter := time.Date(2026, 11, 18, 11, 16, 0, 0, time.UTC)
	if got := DateTime(winter); got != "18.11.2026, 13:16" {
		t.Errorf("DateTime = %q", got)
	}
	// 22:30Z on 17.11 is 00:30 local on 18.11.
	if got := DateTime(time.Date(2026, 11, 17, 22, 30, 0, 0, time.UTC)); got != "18.11.2026, 00:30" {
		t.Errorf("DateTime across midnight = %q", got)
	}
	if got := DateTime(time.Date(2026, 7, 1, 9, 5, 0, 0, time.UTC)); got != "01.07.2026, 12:05" {
		t.Errorf("DateTime summer = %q", got)
	}
	if got := Range(timex.MustDay(2025, 12, 29), timex.MustDay(2026, 1, 4)); got != "29.12.2025\u201304.01.2026" {
		t.Errorf("Range across years = %q", got)
	}
	if strings.Contains(Range(timex.MustDay(2026, 11, 10), timex.MustDay(2026, 11, 16)), " ") {
		t.Error("range contains a space")
	}
}

func TestSexWord(t *testing.T) {
	type sex string // stands in for the sqlc enum type
	if got := SexWord(sex("F")); got != "găini" {
		t.Errorf("F = %q", got)
	}
	if got := SexWord("M"); got != "cocoși" {
		t.Errorf("M = %q", got)
	}
	if strings.ContainsAny(SexWord("M")+SexWord("F"), "\u015f\u0163") {
		t.Error("cedilla in sex word")
	}
	defer func() {
		if recover() == nil {
			t.Error("unknown sex did not panic")
		}
	}()
	SexWord("X")
}
