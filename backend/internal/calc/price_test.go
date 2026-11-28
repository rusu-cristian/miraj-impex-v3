package calc

import (
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestDerivePrice(t *testing.T) {
	d := decimal.RequireFromString
	p := func(s string) *decimal.Decimal { v := d(s); return &v }
	for _, tc := range []struct {
		name         string
		kg           string
		perKg, total *decimal.Decimal
		want         Price
	}{
		// Golden 30: the 13.11.2026 invoice, stored 7.7613 (shown 7,76).
		{"total → per kg, 4 dp", "25640", nil, p("199000"), Price{d("7.7613"), d("199000"), PriceInputTotal}},
		// The D72 board's change: 25 640 × 7,74.
		{"per kg → total, 2 dp", "25640", p("7.74"), nil, Price{d("7.74"), d("198453.60"), PriceInputPerKg}},
		{"whole numbers", "2000", p("8"), nil, Price{d("8"), d("16000.00"), PriceInputPerKg}},
		// 1 000,5 × 1,2345 = 1 235,117 25 → 1 235,12; 10 ÷ 3 = 3,333… → 3,3333.
		{"per kg rounds the total", "1000.5", p("1.2345"), nil, Price{d("1.2345"), d("1235.12"), PriceInputPerKg}},
		{"total rounds the per kg", "3", nil, p("10"), Price{d("3.3333"), d("10"), PriceInputTotal}},
		// Half away from zero: 0,5 ÷ 10 000 = 0,000 05 → 0,0001; 2 ÷ 3 = 0,6666… → 0,6667; 0,5 × 0,01 = 0,005 → 0,01.
		{"half away from zero, per kg", "10000", nil, p("0.50"), Price{d("0.0001"), d("0.5"), PriceInputTotal}},
		{"half up, per kg", "3", nil, p("2"), Price{d("0.6667"), d("2"), PriceInputTotal}},
		{"half away from zero, total", "0.5", p("0.01"), nil, Price{d("0.01"), d("0.01"), PriceInputPerKg}},
		{"trailing zeros are not decimals", "100", p("7.740000"), nil, Price{d("7.74"), d("774"), PriceInputPerKg}},
		{"largest per kg", "1", p("99999999.9999"), nil, Price{d("99999999.9999"), d("100000000.00"), PriceInputPerKg}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DerivePrice(d(tc.kg), tc.perKg, tc.total)
			if err != nil {
				t.Fatal(err)
			}
			if !got.PerKg.Equal(tc.want.PerKg) || !got.Total.Equal(tc.want.Total) || got.Input != tc.want.Input {
				t.Errorf("got %s lei/kg · %s lei (%s), want %s · %s (%s)",
					got.PerKg, got.Total, got.Input, tc.want.PerKg, tc.want.Total, tc.want.Input)
			}
		})
	}
	// Golden 30's shown figure: 7,7613 → 7,76 at 2 dp.
	if got, _ := DerivePrice(d("25640"), nil, p("199000")); got.PerKg.Round(2).String() != "7.76" {
		t.Errorf("golden 30 shown = %s, want 7.76", got.PerKg.Round(2))
	}
}

func TestDerivePriceErrors(t *testing.T) {
	d := decimal.RequireFromString
	p := func(s string) *decimal.Decimal { v := d(s); return &v }
	for _, tc := range []struct {
		name         string
		kg           string
		perKg, total *decimal.Decimal
		err          error
		input        PriceInput // "" for an error without a typed input
	}{
		{"both", "100", p("7"), p("700"), ErrPriceBoth, ""},
		{"neither", "100", nil, nil, ErrPriceNeither, ""},
		{"no kg", "0", p("7"), nil, ErrPriceKg, ""},
		{"per kg 0", "100", p("0"), nil, ErrPriceNotPositive, PriceInputPerKg},
		{"per kg below 0", "100", p("-7.5"), nil, ErrPriceNotPositive, PriceInputPerKg},
		{"total 0", "100", nil, p("0"), ErrPriceNotPositive, PriceInputTotal},
		{"total below 0", "100", nil, p("-1"), ErrPriceNotPositive, PriceInputTotal},
		{"per kg with 5 decimals", "100", p("7.76131"), nil, ErrPriceScale, PriceInputPerKg},
		{"total with 3 decimals", "100", nil, p("776.131"), ErrPriceScale, PriceInputTotal},
		{"per kg beyond numeric(12,4)", "1", p("100000000"), nil, ErrPriceRange, PriceInputPerKg},
		{"total beyond numeric(14,2)", "1", nil, p("1000000000000"), ErrPriceRange, PriceInputTotal},
		{"derived total beyond numeric(14,2)", "99999999999.9", p("10.5"), nil, ErrPriceRange, PriceInputPerKg},
		{"derived per kg beyond numeric(12,4)", "0.1", nil, p("10000000"), ErrPriceRange, PriceInputTotal},
		// 0,0001 × 0,1 = 0,000 01 lei → 0,00: the column wants > 0.
		{"derived total rounds to 0", "0.1", p("0.0001"), nil, ErrPriceRange, PriceInputPerKg},
		{"derived per kg rounds to 0", "1000", nil, p("0.01"), ErrPriceRange, PriceInputTotal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DerivePrice(d(tc.kg), tc.perKg, tc.total)
			if !errors.Is(err, tc.err) {
				t.Fatalf("got %+v, %v; want %v", got, err, tc.err)
			}
			pe, ok := errors.AsType[*PriceError](err)
			switch {
			case tc.input == "" && ok:
				t.Errorf("%v is a *PriceError, want a plain error", err)
			case tc.input != "" && (!ok || pe.Input != tc.input):
				t.Errorf("%v: want a *PriceError on %s", err, tc.input)
			}
		})
	}
}
