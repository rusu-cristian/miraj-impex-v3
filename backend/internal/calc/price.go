package calc

import (
	"errors"

	"github.com/shopspring/decimal"
)

// PriceInput is which of the two prices the owner typed (feed_delivery
// .price_input); the other one is derived (§6.8).
type PriceInput string

// PriceInput values, as stored.
const (
	PriceInputPerKg PriceInput = "per_kg"
	PriceInputTotal PriceInput = "total"
)

// Price places and bounds: price_per_kg numeric(12,4), total_lei
// numeric(14,2), both > 0 (§6.8, feed_delivery).
const (
	PricePerKgPlaces = 4
	PriceTotalPlaces = 2
)

var (
	maxPerKg = decimal.New(1, 8)  // numeric(12,4): below 10^8
	maxTotal = decimal.New(1, 12) // numeric(14,2): below 10^12
)

// Errors of DerivePrice. ErrPriceNotPositive, ErrPriceScale and ErrPriceRange
// come wrapped in a *PriceError naming the typed input; the services map them
// to their messages.
var (
	// ErrPriceBoth: the item has both a price per kg and a total.
	ErrPriceBoth = errors.New("calc: both a price per kg and a total")
	// ErrPriceNeither: the item has neither.
	ErrPriceNeither = errors.New("calc: neither a price per kg nor a total")
	// ErrPriceNotPositive: the typed value is 0 or below.
	ErrPriceNotPositive = errors.New("calc: the price is not above 0")
	// ErrPriceScale: more than 4 decimals on a price per kg, 2 on a total.
	ErrPriceScale = errors.New("calc: the price has too many decimals")
	// ErrPriceRange: the typed value, or the one derived from it, does not
	// fit its column (or the derived one rounds to 0).
	ErrPriceRange = errors.New("calc: the price is out of range")
	// ErrPriceKg: the delivery's kg is not above 0 (a caller bug; the
	// column refuses it).
	ErrPriceKg = errors.New("calc: the delivery kg is not above 0")
)

// PriceError is a problem with the typed value of Input.
type PriceError struct {
	Input PriceInput
	Err   error // ErrPriceNotPositive | ErrPriceScale | ErrPriceRange
}

func (e *PriceError) Error() string { return e.Err.Error() + " (" + string(e.Input) + ")" }

func (e *PriceError) Unwrap() error { return e.Err }

// Price is a delivery's price: the two stored values and which one was typed.
type Price struct {
	PerKg decimal.Decimal // lei/kg, 4 decimals
	Total decimal.Decimal // lei, 2 decimals
	Input PriceInput
}

// DerivePrice completes a delivery's price from exactly one typed value
// (§6.8, M4 D3): total = round(kg × perKg, 2) or perKg = round(total ÷ kg, 4),
// half away from zero, exact (no intermediate rounding). 25 640 kg at
// 199 000 lei → 7,7613 lei/kg; 25 640 kg at 7,74 lei/kg → 198 453,60 lei.
// The typed value must be > 0, have at most 4 (per kg) or 2 (total) decimals
// and, like the derived one, fit its column.
func DerivePrice(kg decimal.Decimal, perKg, total *decimal.Decimal) (Price, error) {
	if kg.Sign() <= 0 {
		return Price{}, ErrPriceKg
	}
	switch {
	case perKg != nil && total != nil:
		return Price{}, ErrPriceBoth
	case perKg != nil:
		if err := checkTyped(*perKg, PricePerKgPlaces, maxPerKg); err != nil {
			return Price{}, &PriceError{Input: PriceInputPerKg, Err: err}
		}
		t := kg.Mul(*perKg).Round(PriceTotalPlaces)
		if !fits(t, maxTotal) {
			return Price{}, &PriceError{Input: PriceInputPerKg, Err: ErrPriceRange}
		}
		return Price{PerKg: *perKg, Total: t, Input: PriceInputPerKg}, nil
	case total != nil:
		if err := checkTyped(*total, PriceTotalPlaces, maxTotal); err != nil {
			return Price{}, &PriceError{Input: PriceInputTotal, Err: err}
		}
		p := total.DivRound(kg, PricePerKgPlaces)
		if !fits(p, maxPerKg) {
			return Price{}, &PriceError{Input: PriceInputTotal, Err: ErrPriceRange}
		}
		return Price{PerKg: p, Total: *total, Input: PriceInputTotal}, nil
	}
	return Price{}, ErrPriceNeither
}

// checkTyped checks a typed price: > 0, at most places decimals, below limit.
func checkTyped(v decimal.Decimal, places int32, limit decimal.Decimal) error {
	switch {
	case v.Sign() <= 0:
		return ErrPriceNotPositive
	case !v.Equal(v.Truncate(places)):
		return ErrPriceScale
	case !fits(v, limit):
		return ErrPriceRange
	}
	return nil
}

// fits reports whether 0 < v < limit.
func fits(v, limit decimal.Decimal) bool { return v.Sign() > 0 && v.LessThan(limit) }
