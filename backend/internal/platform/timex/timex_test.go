package timex

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestChisinauLoaded(t *testing.T) {
	if Chisinau == nil || Chisinau.String() != "Europe/Chisinau" {
		t.Fatalf("Chisinau = %v", Chisinau)
	}
}

func TestMustLoadFailsFast(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("mustLoad of an unknown zone did not panic")
		}
	}()
	mustLoad("Nowhere/Atlantis")
}

func TestToday(t *testing.T) {
	want := MustDay(2026, time.November, 18)
	for _, tc := range []struct {
		name string
		now  time.Time
	}{
		{"noon local", time.Date(2026, 11, 18, 12, 0, 0, 0, Chisinau)},
		{"22:30Z is after local midnight", time.Date(2026, 11, 17, 22, 30, 0, 0, time.UTC)},
		{"21:59Z is 23:59 local, still 18.11", time.Date(2026, 11, 18, 21, 59, 0, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := Today(FixedClock(tc.now)); got != want {
				t.Fatalf("Today = %v, want %v", got, want)
			}
		})
	}
	if got := Today(FixedClock(time.Date(2026, 11, 17, 21, 59, 0, 0, time.UTC))); got != MustDay(2026, 11, 17) {
		t.Fatalf("21:59Z on 17.11 = %v, want 2026-11-17", got)
	}
}

func TestFixedClock(t *testing.T) {
	start := time.Date(2026, 11, 18, 12, 0, 0, 0, Chisinau)
	c := FixedClock(start)
	if !c.Now().Equal(start) {
		t.Fatalf("Now = %v", c.Now())
	}
	c.Advance(13 * time.Hour)
	if got := Today(c); got != MustDay(2026, 11, 19) {
		t.Fatalf("after 13h Today = %v", got)
	}
	c.Set(start)
	if !c.Now().Equal(start) {
		t.Fatalf("after Set Now = %v", c.Now())
	}
	// Concurrent use is safe (run with -race).
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			c.Advance(time.Second)
			_ = c.Now()
		})
	}
	wg.Wait()
	if got := c.Now().Sub(start); got != 8*time.Second {
		t.Fatalf("advanced %v, want 8s", got)
	}
	var _ Clock = RealClock{}
	if RealClock.Now(RealClock{}).IsZero() {
		t.Fatal("RealClock.Now is zero")
	}
}

func TestParseDay(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Day
		ok   bool
	}{
		{"2026-11-18", MustDay(2026, 11, 18), true},
		{"2024-02-29", MustDay(2024, 2, 29), true},
		{"2026-02-30", Day{}, false},
		{"2026-02-29", Day{}, false},
		{"2026-13-01", Day{}, false},
		{"2026-1-01", Day{}, false},
		{"18.11.2026", Day{}, false},
		{"2026-11-18T00:00:00Z", Day{}, false},
		{"0000-01-01", Day{}, false},
		{"", Day{}, false},
	} {
		got, err := ParseDay(tc.in)
		if tc.ok != (err == nil) || got != tc.want {
			t.Errorf("ParseDay(%q) = %v, %v", tc.in, got, err)
		}
		if err != nil && !errors.Is(err, ErrBadDay) {
			t.Errorf("ParseDay(%q) error %v is not ErrBadDay", tc.in, err)
		}
	}
}

func TestDayJSON(t *testing.T) {
	type body struct {
		Day  Day  `json:"day"`
		Opt  *Day `json:"opt"`
		Omit *Day `json:"omit,omitzero"`
	}
	in := body{Day: MustDay(2026, 11, 18)}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"day":"2026-11-18","opt":null}`; got != want {
		t.Fatalf("Marshal = %s, want %s", got, want)
	}
	var out body
	if err := json.Unmarshal([]byte(`{"day":"2026-11-18","opt":"2025-10-07"}`), &out); err != nil {
		t.Fatal(err)
	}
	if out.Day != in.Day || out.Opt == nil || *out.Opt != MustDay(2025, 10, 7) {
		t.Fatalf("Unmarshal = %+v", out)
	}

	for _, tc := range []struct{ name, body, value string }{
		{"impossible date", `{"day":"2026-02-30"}`, "date 2026-02-30"},
		{"number", `{"day":20261118}`, "number"},
		{"wrong layout", `{"day":"18.11.2026"}`, "date 18.11.2026"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out body
			err := json.Unmarshal([]byte(tc.body), &out)
			var te *json.UnmarshalTypeError
			if !errors.As(err, &te) {
				t.Fatalf("error %v (%T), want *json.UnmarshalTypeError", err, err)
			}
			// The JSON path is the decoder's job (httpx): with Go 1.27's
			// default encoding/json (v2-backed, legacy errors) an Unmarshaler's
			// error is returned verbatim, without Field.
			if te.Value != tc.value || te.Type != dayType {
				t.Fatalf("UnmarshalTypeError = %+v", te)
			}
		})
	}

	if _, err := json.Marshal(Day{}); err == nil {
		t.Fatal("marshal zero Day: no error")
	}
}

func TestDayArithmetic(t *testing.T) {
	d := MustDay(2026, 11, 18)
	if got := d.AddDays(14); got != MustDay(2026, 12, 2) {
		t.Fatalf("AddDays(14) = %v", got)
	}
	if got := d.AddDays(-18); got != MustDay(2026, 10, 31) {
		t.Fatalf("AddDays(-18) = %v", got)
	}
	// Across the DST change of 25.10.2026 (25-hour day).
	if got := d.DaysSince(MustDay(2026, 10, 24)); got != 25 {
		t.Fatalf("DaysSince = %d", got)
	}
	if !MustDay(2026, 11, 17).Before(d) || !d.After(MustDay(2026, 11, 17)) || d.Compare(d) != 0 {
		t.Fatal("ordering")
	}
	if d.Weekday() != time.Wednesday {
		t.Fatalf("Weekday = %v", d.Weekday())
	}
	if got := d.Start(); !got.Equal(time.Date(2026, 11, 17, 22, 0, 0, 0, time.UTC)) {
		t.Fatalf("Start = %v", got)
	}
	if !(Day{}).IsZero() || d.IsZero() {
		t.Fatal("IsZero")
	}
}

func TestDayPgx(t *testing.T) {
	m := pgtype.NewMap()
	d := MustDay(2026, 11, 18)
	for _, format := range []int16{pgtype.BinaryFormatCode, pgtype.TextFormatCode} {
		buf, err := m.Encode(pgtype.DateOID, format, d, nil)
		if err != nil {
			t.Fatalf("format %d: encode: %v", format, err)
		}
		var got Day
		if err := m.Scan(pgtype.DateOID, format, buf, &got); err != nil {
			t.Fatalf("format %d: scan: %v", format, err)
		}
		if got != d {
			t.Fatalf("format %d: round trip = %v", format, got)
		}
		var ptr *Day
		if err := m.Scan(pgtype.DateOID, format, nil, &ptr); err != nil || ptr != nil {
			t.Fatalf("format %d: NULL into *Day = %v, %v", format, ptr, err)
		}
		var val Day
		if err := m.Scan(pgtype.DateOID, format, nil, &val); err == nil {
			t.Fatalf("format %d: NULL into Day: no error", format)
		}
	}
	if _, err := m.Encode(pgtype.DateOID, pgtype.BinaryFormatCode, Day{}, nil); err == nil {
		t.Fatal("encode zero Day: no error")
	}
}

func TestLocalTimeJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   time.Time
		want string
	}{
		{"winter +02:00", time.Date(2026, 11, 18, 11, 16, 0, 0, time.UTC), "2026-11-18T13:16:00+02:00"},
		{"summer +03:00", time.Date(2026, 7, 1, 9, 30, 5, 0, time.UTC), "2026-07-01T12:30:05+03:00"},
		{"fraction dropped", time.Date(2026, 11, 18, 11, 16, 0, 987654321, time.UTC), "2026-11-18T13:16:00+02:00"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(At(tc.in))
			if err != nil {
				t.Fatal(err)
			}
			if got := string(b); got != `"`+tc.want+`"` {
				t.Fatalf("Marshal = %s, want %q", got, tc.want)
			}
			if strings.Contains(string(b), "Z") {
				t.Fatal("contains Z")
			}
			var back LocalTime
			if err := json.Unmarshal(b, &back); err != nil {
				t.Fatal(err)
			}
			if !back.Equal(tc.in.Truncate(time.Second)) {
				t.Fatalf("round trip = %v", back.Time)
			}
		})
	}
	var lt LocalTime
	var te *json.UnmarshalTypeError
	if err := json.Unmarshal([]byte(`"18.11.2026"`), &lt); !errors.As(err, &te) {
		t.Fatalf("bad timestamp error = %v", err)
	}
}
