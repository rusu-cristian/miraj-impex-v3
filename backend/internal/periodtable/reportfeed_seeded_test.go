package periodtable_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// consumFigures is a Consum row at the S11 board's precision: kg ♀ · ♂ and
// g ♀ · ♂ at 1 decimal, guide ♀ · ♂ whole, deviation ♀ · ♂ at 1 decimal,
// the status of g ♀ · ♂ and the missing keys.
func consumFigures(c periodtable.Cells, status map[string]string, m *periodtable.Missing) string {
	at := func(key string, places int32) string {
		v, ok := c[key]
		switch {
		case !ok:
			return "absent"
		case v == nil:
			return "–"
		}
		return num.Round(v.Decimal(), places).StringFixed(places)
	}
	s := fmt.Sprintf("kg %s · %s, g %s · %s, ghid %s · %s, %s%% · %s%%, %s · %s",
		at("feed.kg.F", 1), at("feed.kg.M", 1), at("feed.g.F", 1), at("feed.g.M", 1),
		at("feed.g.guide.F", 0), at("feed.g.guide.M", 0), at("feed.g.diffPct.F", 1), at("feed.g.diffPct.M", 1),
		status["feed.g.F"], status["feed.g.M"])
	if m != nil {
		var keys []string
		for _, k := range m.Keys {
			keys = append(keys, k.Code+" "+string(k.Sex)+" "+k.Date.String())
		}
		s += fmt.Sprintf(", lipsă %d: %s", m.Rows, strings.Join(keys, ", "))
	}
	return s
}

func hirbovat(t *testing.T, d *testdb.DB) uuid.UUID {
	t.Helper()
	var farm uuid.UUID
	if err := d.App.QueryRow(t.Context(), `select id from farm where name_key = 'hirbovat'`).Scan(&farm); err != nil {
		t.Fatal(err)
	}
	return farm
}

// TestConsumSeeded (M3 T7 criterion 4): page 1 of Hîrbovăț's Consum (SAPT
// 56–59) against the S11 board and goldens 19, 21, 24 and 26, each at its
// board precision. A farm day's g/pasăre is the register's (§5.7 day row,
// "zi = g din registru"); weeks, sections and the totals are kg ÷ bird-days.
func TestConsumSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := hirbovat(t, d)
	resp, err := periodtable.Build(ctx, d.App, periodtable.Input{
		Query: periodtable.TableQuery{View: "report-feed", FarmID: &farm}, Today: testdb.Today, Now: testdb.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Loaded == nil || resp.Loaded.SaptFrom != 56 || resp.Loaded.SaptTo != 59 || len(resp.Weeks) != 4 {
		t.Fatalf("loaded %+v, %d weeks", resp.Loaded, len(resp.Weeks))
	}
	if resp.More == nil || resp.More.Label != "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40" {
		t.Errorf("more = %+v", resp.More)
	}
	if !cmp.Equal(resp.Chips, []string{"hrana"}) || resp.NoGuide {
		t.Errorf("chips %v, noGuide %v", resp.Chips, resp.NoGuide)
	}

	got := map[string]string{}
	var sections59, sections58 []string
	for _, w := range resp.Weeks {
		got[fmt.Sprint("SAPT ", w.Sapt)] = consumFigures(w.Cells, w.Status, w.Missing)
		for _, day := range w.Days {
			got[day.Date.String()] = consumFigures(day.Cells, day.Status, day.Missing)
		}
		for _, s := range w.Sections {
			key := fmt.Sprint(s.Code, " SAPT ", w.Sapt)
			got[key] = consumFigures(s.Cells, s.Status, s.Missing)
			switch w.Sapt {
			case 59:
				sections59 = append(sections59, s.Code+" "+s.Label)
				if s.Code == "H7" {
					got[key] = fmt.Sprint(s.Reasons, len(s.Cells))
				}
			case 58:
				sections58 = append(sections58, s.Code)
			}
		}
	}
	got["totals"] = resp.Totals.Label + ": " + consumFigures(resp.Totals.Cells, resp.Totals.Status, resp.Totals.Missing)

	for key, want := range map[string]string{
		"SAPT 59":      "kg 4675.3 · 329.7, g 160.0 · 142.7, ghid 157 · 156, 1.9% · -8.5%, in_norma · in_norma",
		"2026-11-17":   "kg 2372.1 · 175.0, g 160.0 · 142.6, ghid 157 · 156, 1.9% · -8.6%, in_norma · in_norma",
		"2026-11-18":   "kg 2303.2 · 154.7, g 160.0 · 142.8, ghid 157 · 156, 1.9% · -8.4%, in_norma · in_norma",
		"H5.1 SAPT 59": "kg 1364.4 · 98.4, g 160.0 · 143.9, ghid 157 · 156, 1.9% · -7.8%, in_norma · in_norma",
		"H6.2 SAPT 59": "kg 589.4 · 46.5, g 160.0 · 140.9, ghid 157 · 156, 1.9% · -9.7%, in_norma · in_norma",
		"H7 SAPT 59":   "map[*:empty_since:2026-11-14] 0",
		"SAPT 58": "kg 22230.9 · 1682.7, g 155.9 · 142.2, ghid 158 · 155, -1.3% · -8.3%, in_norma · in_norma, " +
			"lipsă 2: H7.1 F 2026-11-13, H7.2 F 2026-11-13",
		"2026-11-13": "kg 2723.3 · 281.1, g 160.0 · 142.0, ghid 158 · 155, 1.3% · -8.4%, in_norma · in_norma, " +
			"lipsă 2: H7.1 F 2026-11-13, H7.2 F 2026-11-13",
		"2026-11-14": "kg 2723.0 · 190.8, g 160.0 · 142.5, ghid 158 · 155, 1.3% · -8.1%, in_norma · in_norma",
		"H7.1 SAPT 58": "kg 1923.7 · 176.0, g 151.7 · 141.0, ghid 158 · 155, -4.0% · -9.0%, in_norma · in_norma, " +
			"lipsă 1: H7.1 F 2026-11-13",
		"SAPT 57":      "kg 23284.2 · 1970.6, g 130.5 · 140.8, ghid 158 · 154, -17.4% · -8.6%, abatere · in_norma",
		"H7.2 SAPT 57": "kg 3600.3 · 327.0, g 123.7 · 140.0, ghid 158 · 154, -21.7% · -9.1%, abatere · in_norma",
		"2026-11-09":   "kg 3770.2 · 279.5, g 148.4 · 140.8, ghid 158 · 154, -6.1% · -8.5%, in_norma · in_norma",
		// Days with a register kg far from g × birds (D163) or on a sale day
		// (H6.2 ♀ 15.11): the day keeps the register's g, as on the board.
		"2026-11-02":   "kg 3691.7 · 285.5, g 161.2 · 140.8, ghid 158 · 153, 2.1% · -8.0%, in_norma · in_norma",
		"2026-11-03":   "kg 2760.0 · 284.2, g 161.2 · 140.8, ghid 158 · 154, 2.0% · -8.6%, in_norma · in_norma",
		"2026-11-04":   "kg 2760.0 · 282.7, g 161.2 · 140.8, ghid 158 · 154, 2.0% · -8.6%, in_norma · in_norma",
		"2026-11-05":   "kg 3309.9 · 281.9, g 145.5 · 140.8, ghid 158 · 154, -7.9% · -8.5%, in_norma · in_norma",
		"2026-11-06":   "kg 3394.1 · 281.3, g 149.2 · 140.8, ghid 158 · 154, -5.5% · -8.5%, in_norma · in_norma",
		"2026-11-10":   "kg 3895.8 · 281.7, g 153.4 · 142.0, ghid 158 · 155, -2.9% · -8.4%, in_norma · in_norma",
		"2026-11-15":   "kg 2471.5 · 183.2, g 160.0 · 142.6, ghid 158 · 155, 1.3% · -8.0%, in_norma · in_norma",
		"SAPT 56":      "kg 28708.2 · 2013.6, g 158.6 · 140.9, ghid 158 · 153, 0.4% · -7.9%, in_norma · in_norma",
		"H7.2 SAPT 56": "kg 4816.9 · 333.2, g 163.0 · 140.1, ghid 158 · 153, 3.2% · -8.5%, in_norma · in_norma",
		"totals": "Total SAPT 56–59: kg 78898.6 · 5996.6, g 148.5 · 141.3, ghid 158 · 154, -6.0% · -8.3%, in_norma · in_norma, " +
			"lipsă 2: H7.1 F 2026-11-13, H7.2 F 2026-11-13",
	} {
		if got[key] != want {
			t.Errorf("%s:\n got %s\nwant %s", key, got[key], want)
		}
	}
	if want := []string{"H5.1 H5.1", "H5.2 H5.2", "H6.1 H6.1", "H6.2 H6.2", "H7 goală din 14.11.2026 · H7.1, H7.2"}; !cmp.Equal(sections59, want) {
		t.Errorf("SAPT 59 sections %q, want %q", sections59, want)
	}
	if want := []string{"H5.1", "H5.2", "H6.1", "H6.2", "H7.1", "H7.2"}; !cmp.Equal(sections58, want) {
		t.Errorf("SAPT 58 sections %q, want %q", sections58, want)
	}
	wantNotes := []string{
		"g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre (♀ după găini, ♂ după cocoși). Toleranță ±10%.",
		"♀ 460 kg/compartiment în registru (g × păsări ≈ 690 kg): SAPT 57: H5, H6 03.11–04.11 · H7 03.11–06.11; SAPT 56: H6 02.11.",
	}
	if !cmp.Equal(resp.Notes, wantNotes) {
		t.Errorf("notes:\n got %q\nwant %q", resp.Notes, wantNotes)
	}
}

// TestConsumMatchesRezumat (ARC-7): Consum's day query and the Rezumat's are
// two readings of §5.7. On every day of the seeded Seria 2026 cycle the farm
// day kg agree, and on every week so do the g/pasăre (kg ÷ fed bird-days,
// D38); golden 26 through calc.FeedKg on 18.11. (A day's g is the
// register's in Consum only, so it is not compared.)
func TestConsumMatchesRezumat(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := hirbovat(t, d)
	resp, err := periodtable.Build(ctx, d.App, periodtable.Input{
		Query: periodtable.TableQuery{View: "report-feed", FarmID: &farm, Count: 60}, Today: testdb.Today, Now: testdb.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	days, err := periodtable.ReadDays(ctx, d.App, periodtable.DaysQuery{
		FarmID: farm, Scope: periodtable.ScopeFarm, From: resp.Period.From, To: resp.Period.To,
	})
	if err != nil {
		t.Fatal(err)
	}
	rezumat := map[timex.Day]calc.DayFigures{}
	for _, r := range days.Rows {
		rezumat[r.Date] = r.DayFigures
	}
	// compare checks one cell against the Rezumat figure; a row without a
	// ration is "–" in Consum and 0 kg (no g) in the Rezumat query.
	compare := func(row, key string, cells periodtable.Cells, want *decimal.Decimal) {
		t.Helper()
		got, w := "–", "–"
		if v := cells[key]; v != nil {
			got = v.Decimal().String()
		}
		if want != nil && (cells[key] != nil || !want.IsZero()) {
			w = want.String()
		}
		if got != w {
			t.Errorf("%s %s = %s, Rezumat %s", row, key, got, w)
		}
	}
	n := 0
	for _, w := range resp.Weeks {
		var week calc.Rollup
		for _, day := range w.Days {
			n++
			r := rezumat[day.Date]
			week.Add(r)
			compare(day.Date.String(), "feed.kg.F", day.Cells, &r.FeedKg.F)
			compare(day.Date.String(), "feed.kg.M", day.Cells, &r.FeedKg.M)
		}
		g := week.FeedGPerBird()
		compare(w.Label, "feed.g.F", w.Cells, g.F)
		compare(w.Label, "feed.g.M", w.Cells, g.M)
	}
	if want := testdb.Today.DaysSince(resp.Period.From) + 1; n != want {
		t.Errorf("%d day rows, want %d", n, want)
	}
	kg, err := calc.FeedKg(ctx, d.App, farm, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	if kg.F.String() != "2303.2" || kg.M.String() != "154.7" {
		t.Errorf("calc.FeedKg 18.11 = %s · %s, want 2303.2 · 154.7 (golden 26)", kg.F, kg.M)
	}
}
