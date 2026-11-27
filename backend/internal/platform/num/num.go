// Package num is the decimal type of the API (SPEC §7.0a, §11.2): kg, lei,
// grams and percentages are never float64 on the server.
package num

import (
	"encoding/json"
	"reflect"

	"github.com/shopspring/decimal"
)

// JSONPlaces is the most decimals a Num carries on the wire.
const JSONPlaces = 4

// Limits for numbers read from JSON: real inputs have a handful of digits, and
// a huge exponent ("1e9999") would make every later comparison expensive.
const (
	maxJSONLen      = 40
	maxJSONExponent = 18
)

// Num is a decimal number. In JSON it is a number (never a string), written
// rounded half away from zero to at most 4 decimals with trailing zeros
// dropped ("2303.2", "40.4024").
type Num struct {
	d decimal.Decimal
}

// Of wraps d.
func Of(d decimal.Decimal) Num { return Num{d: d} }

// FromInt returns i as a Num.
func FromInt(i int64) Num { return Num{d: decimal.NewFromInt(i)} }

// Parse reads a decimal in Go/JSON notation ("40.4024", "-2.5").
func Parse(s string) (Num, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Num{}, err
	}
	return Num{d: d}, nil
}

// MustParse is Parse for constants and tests; it panics on bad input.
func MustParse(s string) Num {
	n, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return n
}

// Decimal returns the full-precision value.
func (n Num) Decimal() decimal.Decimal { return n.d }

// String is the full-precision value in plain notation.
func (n Num) String() string { return n.d.String() }

// Round returns d rounded to places decimals, half away from zero (D82,
// SPEC §5.15): 90.35 → 90.4, −2.5 → −3. Round once, from full precision.
func Round(d decimal.Decimal, places int32) decimal.Decimal {
	// decimal.Decimal.Round is half away from zero (verified, NOTES-verify.md).
	return d.Round(places)
}

// MarshalJSON writes a JSON number with at most 4 decimals, half away from
// zero, trailing zeros dropped.
func (n Num) MarshalJSON() ([]byte, error) {
	return []byte(Round(n.d, JSONPlaces).String()), nil
}

var numType = reflect.TypeFor[Num]()

// UnmarshalJSON accepts a JSON number only. A string (or any other kind) is a
// *json.UnmarshalTypeError, so the request is a 400. JSON null leaves n
// unchanged (optional numbers are *Num).
func (n *Num) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if len(data) == 0 || (data[0] != '-' && (data[0] < '0' || data[0] > '9')) {
		return &json.UnmarshalTypeError{Value: kind(data), Type: numType}
	}
	if len(data) > maxJSONLen {
		return &json.UnmarshalTypeError{Value: "number (too long)", Type: numType}
	}
	d, err := decimal.NewFromString(string(data))
	if err != nil {
		return &json.UnmarshalTypeError{Value: "number " + string(data), Type: numType}
	}
	if e := d.Exponent(); e > maxJSONExponent || e < -maxJSONExponent {
		return &json.UnmarshalTypeError{Value: "number " + string(data), Type: numType}
	}
	n.d = d
	return nil
}

func kind(data []byte) string {
	if len(data) == 0 {
		return "value"
	}
	switch data[0] {
	case '"':
		return "string"
	case '{':
		return "object"
	case '[':
		return "array"
	case 't', 'f':
		return "bool"
	default:
		return "value"
	}
}
