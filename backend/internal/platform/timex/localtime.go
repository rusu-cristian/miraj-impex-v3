package timex

import (
	"encoding/json"
	"reflect"
	"time"
)

// LocalTime is an instant written in JSON as Europe/Chisinau local time with
// its offset, whole seconds and never "Z": "2026-11-18T13:16:00+02:00"
// (SPEC §7.0a). Used for asOf, at, createdAt, now.
type LocalTime struct {
	time.Time
}

// At wraps t as a LocalTime.
func At(t time.Time) LocalTime { return LocalTime{Time: t} }

// String returns the RFC 3339 text of the instant in Europe/Chisinau, without
// fractional seconds. Chisinau's offset is never zero, so it never ends in Z.
func (t LocalTime) String() string {
	return t.Time.In(Chisinau).Truncate(time.Second).Format(time.RFC3339)
}

// MarshalJSON writes String as a JSON string.
func (t LocalTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

// MarshalText writes String (overrides the promoted time.Time method).
func (t LocalTime) MarshalText() ([]byte, error) { return []byte(t.String()), nil }

// UnmarshalText reads an RFC 3339 timestamp.
func (t *LocalTime) UnmarshalText(b []byte) error {
	parsed, err := time.Parse(time.RFC3339, string(b))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

var localTimeType = reflect.TypeFor[LocalTime]()

// UnmarshalJSON reads an RFC 3339 timestamp (any offset). JSON null leaves t
// unchanged; anything else returns a *json.UnmarshalTypeError.
func (t *LocalTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return &json.UnmarshalTypeError{Value: jsonKind(data), Type: localTimeType}
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return &json.UnmarshalTypeError{Value: "timestamp " + s, Type: localTimeType}
	}
	t.Time = parsed
	return nil
}
