package periodtable

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// TestResponseJSON: a nil cell is null; loaded, totals, empty, noGuide,
// status, reasons, days and partial are left out when unset; more and start
// are always written (§7.4, §7.0a).
func TestResponseJSON(t *testing.T) {
	v := num.MustParse("40.40245")
	d := timex.MustDay(2026, time.November, 17)
	encode := func(r TableResponse) string {
		t.Helper()
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	resp := TableResponse{
		Chips: []string{}, Groups: []Group{{ID: "vii", Leaves: []Leaf{{ID: "alive.F"}}}}, Notes: []string{},
		Weeks:  []WeekRow{{Sapt: 58, From: d, To: d, DaysInRow: 7, Cells: Cells{"egg.layPct": &v, "egg.guide": nil}}},
		Period: PeriodInfo{From: d, To: d}, AsOf: timex.At(time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)),
	}
	body := encode(resp)
	for _, want := range []string{`"egg.guide":null`, `"egg.layPct":40.4025`, `"more":null`, `"start":null`, `"leaves":[{"id":"alive.F"}]`, `"asOf":"2026-11-18T12:00:00+02:00"`} {
		if !strings.Contains(body, want) {
			t.Errorf("JSON lacks %s: %s", want, body)
		}
	}
	for _, absent := range []string{`"loaded"`, `"totals"`, `"empty"`, `"noGuide"`, `"status"`, `"reasons"`, `"days"`, `"halls"`, `"partial"`} {
		if strings.Contains(body, absent) {
			t.Errorf("JSON has %s: %s", absent, body)
		}
	}
	resp.NoGuide = true
	resp.Loaded = &Loaded{SaptFrom: 56, SaptTo: 59, Weeks: 4}
	resp.Start = &StartNote{Label: "x"}
	resp.Weeks[0].Status = map[string]string{"egg.layPct": "abatere"}
	body = encode(resp)
	for _, want := range []string{`"noGuide":true`, `"loaded":{"saptFrom":56,"saptTo":59,"weeks":4}`, `"start":{"label":"x"}`, `"status":{"egg.layPct":"abatere"}`} {
		if !strings.Contains(body, want) {
			t.Errorf("JSON lacks %s: %s", want, body)
		}
	}
}
