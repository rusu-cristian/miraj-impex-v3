package num

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
)

func TestMarshalJSON(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"2303.20", "2303.2"},
		{"40.40241234", "40.4024"},
		{"40.40245", "40.4025"},   // half away from zero
		{"-40.40245", "-40.4025"}, // half away from zero, negative
		{"7.7600", "7.76"},
		{"160.0", "160"},
		{"0", "0"},
		{"-0.00004", "0"}, // rounds to zero: no "-0"
		{"14080", "14080"},
		{"1e3", "1000"},
	} {
		b, err := json.Marshal(MustParse(tc.in))
		if err != nil {
			t.Fatal(err)
		}
		if got := string(b); got != tc.want {
			t.Errorf("Marshal(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}
	if b, _ := json.Marshal(Num{}); string(b) != "0" {
		t.Errorf("Marshal(zero Num) = %s", b)
	}
	type row struct {
		Kg  Num  `json:"kg"`
		Lei *Num `json:"lei"`
	}
	b, _ := json.Marshal(row{Kg: MustParse("2457.9")})
	if got, want := string(b), `{"kg":2457.9,"lei":null}`; got != want {
		t.Errorf("Marshal(row) = %s, want %s", got, want)
	}
}

func TestRoundHalfAwayFromZero(t *testing.T) {
	// SPEC §5.15 / D82: SAPT 30 peak 168 122 ÷ 186 088 = 90,345… → 90,3 (round
	// once from full precision; rounding 90.35 first would give 90.4).
	lay := decimal.NewFromInt(168122).Div(decimal.NewFromInt(186088)).Mul(decimal.NewFromInt(100))
	if got := Round(lay, 1).String(); got != "90.3" {
		t.Errorf("168122/186088×100 at 1 dp = %s, want 90.3", got)
	}
	if got := Round(Round(lay, 2), 1).String(); got != "90.4" {
		t.Errorf("double rounding = %s (the case the rule forbids), want 90.4", got)
	}
	for _, tc := range []struct {
		in     string
		places int32
		want   string
	}{
		{"-2.5", 0, "-3"},
		{"2.5", 0, "3"},
		{"-1.45", 1, "-1.5"},
		{"1.45", 1, "1.5"},
		{"0.05", 1, "0.1"},
		{"-16.54", 1, "-16.5"},
		{"41.149", 1, "41.1"},
		{"7.765", 2, "7.77"},
	} {
		if got := Round(decimal.RequireFromString(tc.in), tc.places).String(); got != tc.want {
			t.Errorf("Round(%s, %d) = %s, want %s", tc.in, tc.places, got, tc.want)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	type body struct {
		G   Num  `json:"g"`
		Opt *Num `json:"opt"`
	}
	var b body
	if err := json.Unmarshal([]byte(`{"g":141.5,"opt":-2}`), &b); err != nil {
		t.Fatal(err)
	}
	if b.G.String() != "141.5" || b.Opt == nil || b.Opt.String() != "-2" {
		t.Fatalf("got %+v", b)
	}
	b = body{}
	if err := json.Unmarshal([]byte(`{"g":2.5,"opt":null}`), &b); err != nil || b.Opt != nil {
		t.Fatalf("null into *Num = %v, %v", b.Opt, err)
	}

	for _, tc := range []struct{ name, in, value string }{
		{"string", `{"g":"141.5"}`, "string"},
		{"bool", `{"g":true}`, "bool"},
		{"array", `{"g":[1]}`, "array"},
		{"huge exponent", `{"g":1e9999}`, "number 1e9999"},
		{"tiny exponent", `{"g":1e-30}`, "number 1e-30"},
		{"too long", `{"g":12345678901234567890123456789012345678901}`, "number (too long)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var b body
			err := json.Unmarshal([]byte(tc.in), &b)
			var te *json.UnmarshalTypeError
			if !errors.As(err, &te) {
				t.Fatalf("error %v (%T), want *json.UnmarshalTypeError", err, err)
			}
			if te.Value != tc.value || te.Type != numType {
				t.Fatalf("UnmarshalTypeError = %+v", te)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	for _, s := range []string{"0", "2303.2", "40.4024", "-82", "0.0001"} {
		b, err := json.Marshal(MustParse(s))
		if err != nil {
			t.Fatal(err)
		}
		var back Num
		if err := json.Unmarshal(b, &back); err != nil {
			t.Fatal(err)
		}
		if !back.Decimal().Equal(MustParse(s).Decimal()) {
			t.Errorf("round trip %s → %s → %s", s, b, back)
		}
	}
	if FromInt(-82).String() != "-82" || !Of(decimal.NewFromInt(3)).Decimal().Equal(decimal.NewFromInt(3)) {
		t.Error("constructors")
	}
	if _, err := Parse("abc"); err == nil {
		t.Error("Parse(abc): no error")
	}
}
