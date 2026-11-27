package timex

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const dayLayout = "2006-01-02"

// Day is a civil date with no time zone (a local day in Europe/Chisinau,
// SPEC §7.0a). The zero value is "no day" (IsZero); it is never a valid date.
// Days compare with ==.
type Day struct {
	y int
	m time.Month
	d int
}

// ErrBadDay is returned (wrapped) for text that is not a real YYYY-MM-DD date.
var ErrBadDay = errors.New("timex: not a valid YYYY-MM-DD date")

// NewDay returns the day y-m-d, or an error when that date does not exist.
func NewDay(y int, m time.Month, d int) (Day, error) {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if ty, tm, td := t.Date(); ty != y || tm != m || td != d || y < 1 || y > 9999 {
		return Day{}, fmt.Errorf("%w: %04d-%02d-%02d", ErrBadDay, y, int(m), d)
	}
	return Day{y: y, m: m, d: d}, nil
}

// MustDay is NewDay for constants and tests; it panics on an invalid date.
func MustDay(y int, m time.Month, d int) Day {
	day, err := NewDay(y, m, d)
	if err != nil {
		panic(err)
	}
	return day
}

// ParseDay parses "YYYY-MM-DD" and rejects dates that don't exist
// ("2026-02-30").
func ParseDay(s string) (Day, error) {
	t, err := time.Parse(dayLayout, s)
	if err != nil {
		return Day{}, fmt.Errorf("%w: %q", ErrBadDay, s)
	}
	return NewDay(t.Date())
}

func dayFromUTC(t time.Time) Day {
	y, m, d := t.Date()
	return Day{y: y, m: m, d: d}
}

// IsZero reports whether d is the zero Day.
func (d Day) IsZero() bool { return d == Day{} }

// Date returns the year, month and day.
func (d Day) Date() (int, time.Month, int) { return d.y, d.m, d.d }

// String formats the day as "YYYY-MM-DD" ("" for the zero Day).
func (d Day) String() string {
	if d.IsZero() {
		return ""
	}
	return fmt.Sprintf("%04d-%02d-%02d", d.y, int(d.m), d.d)
}

// UTCMidnight is 00:00 UTC on this date, for date arithmetic only; it is not
// the instant the local day starts.
func (d Day) UTCMidnight() time.Time { return time.Date(d.y, d.m, d.d, 0, 0, 0, 0, time.UTC) }

// Start is the instant the day starts in Europe/Chisinau (00:00 local).
func (d Day) Start() time.Time { return time.Date(d.y, d.m, d.d, 0, 0, 0, 0, Chisinau) }

// AddDays returns the day n days later (earlier when n < 0).
func (d Day) AddDays(n int) Day { return dayFromUTC(d.UTCMidnight().AddDate(0, 0, n)) }

// DaysSince returns the number of days from o to d (negative when d is before o).
func (d Day) DaysSince(o Day) int {
	return int(d.UTCMidnight().Sub(o.UTCMidnight()).Hours() / 24)
}

// Compare returns -1, 0 or +1 as d is before, equal to or after o.
func (d Day) Compare(o Day) int { return d.UTCMidnight().Compare(o.UTCMidnight()) }

// Before reports whether d is before o.
func (d Day) Before(o Day) bool { return d.Compare(o) < 0 }

// After reports whether d is after o.
func (d Day) After(o Day) bool { return d.Compare(o) > 0 }

// Weekday is the day of the week.
func (d Day) Weekday() time.Weekday { return d.UTCMidnight().Weekday() }

var dayType = reflect.TypeFor[Day]()

// MarshalJSON writes "YYYY-MM-DD". The zero Day is an error: optional days
// are *Day.
func (d Day) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return nil, errors.New("timex: marshal zero Day")
	}
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON reads "YYYY-MM-DD". JSON null leaves d unchanged. Anything
// else, or a date that doesn't exist, returns a *json.UnmarshalTypeError so
// the decoder reports the JSON path of the field.
func (d *Day) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return &json.UnmarshalTypeError{Value: jsonKind(data), Type: dayType}
	}
	day, err := ParseDay(s)
	if err != nil {
		return &json.UnmarshalTypeError{Value: "date " + s, Type: dayType}
	}
	*d = day
	return nil
}

// MarshalText writes "YYYY-MM-DD" (query parameters, map keys).
func (d Day) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, errors.New("timex: marshal zero Day")
	}
	return []byte(d.String()), nil
}

// UnmarshalText reads "YYYY-MM-DD".
func (d *Day) UnmarshalText(b []byte) error {
	day, err := ParseDay(string(b))
	if err != nil {
		return err
	}
	*d = day
	return nil
}

var (
	_ pgtype.DateScanner = (*Day)(nil)
	_ pgtype.DateValuer  = Day{}
)

// ScanDate implements pgtype.DateScanner (pgx/v5) so a PostgreSQL date scans
// into a Day. NULL and ±infinity are errors; nullable columns scan into *Day.
func (d *Day) ScanDate(v pgtype.Date) error {
	if !v.Valid {
		return errors.New("timex: cannot scan NULL into Day (use *Day)")
	}
	if v.InfinityModifier != pgtype.Finite {
		return errors.New("timex: cannot scan an infinite date into Day")
	}
	*d = dayFromUTC(v.Time)
	return nil
}

// DateValue implements pgtype.DateValuer (pgx/v5) so a Day is sent as a
// PostgreSQL date. The zero Day is an error rather than a silent NULL.
func (d Day) DateValue() (pgtype.Date, error) {
	if d.IsZero() {
		return pgtype.Date{}, errors.New("timex: zero Day as a query parameter (use *Day for NULL)")
	}
	return pgtype.Date{Time: d.UTCMidnight(), Valid: true}, nil
}

// jsonKind names the JSON type of a raw value for decode errors.
func jsonKind(data []byte) string {
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
	case 'n':
		return "null"
	default:
		return "number"
	}
}
