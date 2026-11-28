package audit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// TestValueJSON: a Value is a JSON number (num.Num rules), a string or null.
func TestValueJSON(t *testing.T) {
	n := num.MustParse("67.04")
	for _, tc := range []struct {
		name string
		v    Value
		want string
	}{
		{"zero value", Value{}, `null`},
		{"Null", Null(), `null`},
		{"Int", Int(3), `3`},
		{"Int negative", Int(-4), `-4`},
		{"Number", Number(n), `67.04`},
		{"Dec rounds to 4 decimals", Dec(decimal.RequireFromString("7.761349")), `7.7613`},
		{"Dec trailing zeros", Dec(decimal.RequireFromString("2.50")), `2.5`},
		{"Number zero", Int(0), `0`},
		{"Text", Text("boală"), `"boală"`},
		{"Text quotes", Text(`notă "x"`), `"notă \"x\""`},
		{"Text empty", Text(""), `""`},
		{"OptNumber nil", OptNumber(nil), `null`},
		{"OptNumber", OptNumber(&n), `67.04`},
		{"OptText nil", OptText(nil), `null`},
		{"OptText", OptText(new("Breeder 2")), `"Breeder 2"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.v)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Fatalf("got %s, want %s", b, tc.want)
			}
			// Decoding gives back the value as written (rounded to 4
			// decimals): it encodes to the same bytes.
			var back Value
			if err := json.Unmarshal(b, &back); err != nil {
				t.Fatal(err)
			}
			again, err := json.Marshal(back)
			if err != nil {
				t.Fatal(err)
			}
			if string(again) != tc.want || back.IsNull() != tc.v.IsNull() {
				t.Errorf("round trip: %s → %#v → %s", b, back, again)
			}
		})
	}
}

// TestValueUnmarshalRejects: a bool, an object or an array is not a Value.
func TestValueUnmarshalRejects(t *testing.T) {
	for _, in := range []string{`true`, `false`, `{}`, `[1]`, `"x`, `1e9999`} {
		var v Value
		if err := json.Unmarshal([]byte(in), &v); err == nil {
			t.Errorf("%s: no error, got %#v", in, v)
		}
	}
}

// TestValueUnmarshalNullResets: null over a set Value makes it null again.
func TestValueUnmarshalNullResets(t *testing.T) {
	c := Change{Old: Int(1), New: Text("x")}
	if err := json.Unmarshal([]byte(`{"key":"k","old":null,"new":null}`), &c); err != nil {
		t.Fatal(err)
	}
	if !c.Old.IsNull() || !c.New.IsNull() {
		t.Errorf("old %#v, new %#v, want null", c.Old, c.New)
	}
}

func TestValueAccessorsAndEqual(t *testing.T) {
	if n, ok := Int(3).Num(); !ok || n.String() != "3" {
		t.Errorf("Int(3).Num() = %s %v", n, ok)
	}
	if _, ok := Text("3").Num(); ok {
		t.Error("Text(3).Num() ok")
	}
	if s, ok := Text("altul").Text(); !ok || s != "altul" {
		t.Errorf("Text.Text() = %q %v", s, ok)
	}
	if _, ok := Null().Text(); ok || !Null().IsNull() || Int(0).IsNull() || Text("").IsNull() {
		t.Error("null accessors")
	}
	for _, tc := range []struct {
		a, b Value
		eq   bool
	}{
		{Null(), Null(), true},
		{Dec(decimal.RequireFromString("2.50")), Dec(decimal.RequireFromString("2.5")), true},
		{Int(1), Int(3), false},
		{Int(0), Null(), false},
		{Int(3), Text("3"), false},
		{Text(""), Null(), false},
		{Text("boală"), Text("boală"), true},
	} {
		if got := tc.a.Equal(tc.b); got != tc.eq {
			t.Errorf("%#v.Equal(%#v) = %v", tc.a, tc.b, got)
		}
	}
}

// TestChangeJSON: key, section, sex and field as §4.10 writes them (the
// optional three omitted when empty), old and new always present.
func TestChangeJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		c    Change
		want string
	}{
		{"§4.10 example", Change{Key: "H5.2 · găini moarte", Section: "H5.2", Sex: dbq.SexF, Field: "mortality", Old: Int(1), New: Int(3)},
			`{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":1,"new":3}`},
		{"text values", Change{Key: "H5.2 · motiv sacrificare cocoși", Section: "H5.2", Sex: dbq.SexM, Field: "cullReason", Old: Text("boală"), New: Text("altul")},
			`{"key":"H5.2 · motiv sacrificare cocoși","section":"H5.2","sex":"M","field":"cullReason","old":"boală","new":"altul"}`},
		{"added value", Change{Key: "H5 · Cat. 1", Section: "H5", Field: "cat1", New: Int(15)},
			`{"key":"H5 · Cat. 1","section":"H5","field":"cat1","old":null,"new":15}`},
		{"cleared value", Change{Key: "Livrare · notă", Field: "note", Old: Text("sac rupt")},
			`{"key":"Livrare · notă","field":"note","old":"sac rupt","new":null}`},
		{"key only", Change{Key: "Breeder 2 · 12.11.2026 · lei/kg"}, `{"key":"Breeder 2 · 12.11.2026 · lei/kg","old":null,"new":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.c)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Fatalf("got  %s\nwant %s", b, tc.want)
			}
			var back Change
			if err := json.Unmarshal(b, &back); err != nil {
				t.Fatal(err)
			}
			if !changeEqual(back, tc.c) {
				t.Errorf("round trip: %+v → %s → %+v", tc.c, b, back)
			}
		})
	}
}

// TestChangeSeedShape: the seeded correction's changes (migratehistory's
// auditChange, switched to Change in M5 T8) encode byte for byte the same.
func TestChangeSeedShape(t *testing.T) {
	type auditChange struct { // the M1 seed's private type
		Key     string  `json:"key"`
		Section string  `json:"section"`
		Sex     dbq.Sex `json:"sex"`
		Field   string  `json:"field"`
		Old     int32   `json:"old"`
		New     int32   `json:"new"`
	}
	old, err := json.Marshal([]auditChange{{Key: "H5.2 · găini moarte", Section: "H5.2", Sex: dbq.SexF, Field: "mortality", Old: 1, New: 3}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal([]Change{{Key: "H5.2 · găini moarte", Section: "H5.2", Sex: dbq.SexF, Field: "mortality", Old: Int(1), New: Int(3)}})
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(old) {
		t.Errorf("Change = %s, seed shape %s", got, old)
	}
}

// TestChangesRoundTripJSONB: mixed number / string / null old and new values
// survive audit_log.changes (jsonb) as written by il_app and read back.
func TestChangesRoundTripJSONB(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	user := id(t, d, `insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'manager') returning id`)
	farm := id(t, d, `insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying') returning id`)
	entry := id(t, d, `insert into entry (kind, farm_id, date, created_by) values ('daily', $1, '2026-11-17', $2) returning id`, farm, user)
	w := num.MustParse("67.04")
	changes := []Change{
		{Key: "H5.2 · găini moarte", Section: "H5.2", Sex: dbq.SexF, Field: "mortality", Old: Int(1), New: Int(3)},
		{Key: "H5.2 · motiv sacrificare găini", Section: "H5.2", Sex: dbq.SexF, Field: "cullReason", Old: Text("boală"), New: Text("șchiopătare")},
		{Key: "H5 · masa medie g", Section: "H5", Field: "avgWeightG", Old: Null(), New: OptNumber(&w)},
		{Key: "H5.2 · eroare sexare cocoși", Section: "H5.2", Sex: dbq.SexM, Field: "sexingError", Old: Int(-2), New: Null()},
	}
	kind := dbq.EntryKindDaily
	err := Write(ctx, dbq.New(d.App), Row{
		At: testdb.Now(), UserID: &user, Action: dbq.AuditActionUpdate, Entity: "entry", EntityID: entry, EntryID: &entry,
		EntryKind: &kind, FarmIDs: []uuid.UUID{farm}, Reason: new("numărătoare refăcută"), Changes: changes,
	})
	if err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err := d.App.QueryRow(ctx, `select changes from audit_log where entry_id = $1`, entry).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var back []Change
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("%s: %v", raw, err)
	}
	if len(back) != len(changes) {
		t.Fatalf("got %d changes: %s", len(back), raw)
	}
	for i := range changes {
		if !changeEqual(back[i], changes[i]) {
			t.Errorf("change %d: got %+v, want %+v (jsonb %s)", i, back[i], changes[i], raw)
		}
	}
	// jsonb keeps the text values as strings and the numbers as numbers.
	for _, frag := range []string{`"old": "boală"`, `"new": "șchiopătare"`, `"new": 67.04`, `"old": -2`, `"new": null`} {
		if !strings.Contains(string(raw), frag) {
			t.Errorf("jsonb %s lacks %s", raw, frag)
		}
	}
}

func changeEqual(a, b Change) bool {
	return a.Key == b.Key && a.Section == b.Section && a.Sex == b.Sex && a.Field == b.Field &&
		a.Old.Equal(b.Old) && a.New.Equal(b.New)
}
