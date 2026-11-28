package audit

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
)

// Change is one old → new pair Istoric shows (§4.10, §7.8 M5 D9): the element
// of audit_log.changes and of a correction preview's `changes`. In JSON:
// {"key", "section"?, "sex"?, "field"?, "old", "new"}; old and new are always
// present (null = the value was empty / is cleared).
type Change struct {
	Key     string  `json:"key"`              // "H5.2 · găini moarte" (messages.AuditKey…)
	Section string  `json:"section,omitzero"` // section code: "H5.2"
	Sex     dbq.Sex `json:"sex,omitzero"`     // F | M
	Field   string  `json:"field,omitzero"`   // the JSON field: mortality, cull, cullReason, cat1, g, qtyKg, …
	Old     Value   `json:"old"`
	New     Value   `json:"new"`
}

// valueKind is what a Value holds.
type valueKind uint8

const (
	kindNull valueKind = iota
	kindNumber
	kindText
)

// Value is a Change's old or new value: a JSON number, a string (a shown word:
// "boală", a feed type's name, a note) or null (M5 Q12). The zero Value is
// null.
type Value struct {
	kind valueKind
	n    num.Num
	s    string
}

// Null is the null Value (empty or cleared).
func Null() Value { return Value{} }

// Number is a numeric Value.
func Number(n num.Num) Value { return Value{kind: kindNumber, n: n} }

// Int is a count as a Value.
func Int(i int64) Value { return Number(num.FromInt(i)) }

// Dec is a decimal as a Value (written with at most 4 decimals, as num.Num).
func Dec(d decimal.Decimal) Value { return Number(num.Of(d)) }

// Text is a string Value.
func Text(s string) Value { return Value{kind: kindText, s: s} }

// OptNumber is Number(*n), or null when n is nil (a field left empty).
func OptNumber(n *num.Num) Value {
	if n == nil {
		return Null()
	}
	return Number(*n)
}

// OptText is Text(*s), or null when s is nil.
func OptText(s *string) Value {
	if s == nil {
		return Null()
	}
	return Text(*s)
}

// IsNull reports whether v is null.
func (v Value) IsNull() bool { return v.kind == kindNull }

// Num is v's number and true, or false when v is not a number.
func (v Value) Num() (num.Num, bool) { return v.n, v.kind == kindNumber }

// Text is v's string and true, or false when v is not a string.
func (v Value) Text() (string, bool) { return v.s, v.kind == kindText }

// Equal reports whether v and w hold the same value: numbers compare by value
// (2.50 = 2.5), strings byte for byte, null only equals null.
func (v Value) Equal(w Value) bool {
	if v.kind != w.kind {
		return false
	}
	switch v.kind {
	case kindNumber:
		return v.n.Decimal().Equal(w.n.Decimal())
	case kindText:
		return v.s == w.s
	default:
		return true
	}
}

// MarshalJSON writes a number (num.Num: at most 4 decimals), a string or null.
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.kind {
	case kindNumber:
		return v.n.MarshalJSON()
	case kindText:
		return json.Marshal(v.s)
	default:
		return []byte("null"), nil
	}
}

// UnmarshalJSON reads a number, a string or null; any other JSON value (a
// bool, an object, an array) is an error.
func (v *Value) UnmarshalJSON(data []byte) error {
	switch {
	case string(data) == "null":
		*v = Value{}
		return nil
	case len(data) > 0 && data[0] == '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*v = Text(s)
		return nil
	default:
		var n num.Num
		if err := n.UnmarshalJSON(data); err != nil {
			return fmt.Errorf("audit: change value: %w", err)
		}
		*v = Number(n)
		return nil
	}
}
