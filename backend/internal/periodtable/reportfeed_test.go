package periodtable

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// newFeedFarm is a hand-built laying farm for Consum, shaped like Hîrbovăț's
// last weeks. Seria F is placed 06.10.2026 at age 140, so its weeks run
// Tue–Mon: SAPT 24 27.10–02.11, 25 03–09.11, 26 10–16.11, 27 17–18.11 (2 din
// 7). The cycle opens 08.10 (SAPT 21): SAPT 21–27, 7 weeks.
//
//	Halls H5 (H5.1, H5.2) and H7 (H7.1, H7.2). 08.10 placement ♀ 100 · ♂ 10
//	in each; no deaths. 14.11 sale: H7.1 and H7.2 all their birds (H7 is
//	empty since 14.11), H5.2 ♂ 5.
//	Hrană zilnică every day to 18.11, ♀ 100 g, ♂ 120 g, except:
//	  - H7 from 14.11 (empty), and H7.1 ♀, H7.2 ♀ on 13.11 (missing);
//	  - H5.1 ♀, H5.2 ♀ on 03–04.11 and H5.2 ♂ on 14.11, which are migrated
//	    rows: g 100 with 6 kg (g × birds = 10 kg, D163) and g 120 with 0,5 kg
//	    on the sale day (g × birds at the start = 1,2 kg: a sale day, not a
//	    D163 row);
//	  - H7 on 14.11: migrated rows of 0 kg without g (the sell-out day: not
//	    fed, not missing).
//	Guide: feed ♀ 110 g · ♂ 125 g in SAPT 21–27.
func newFeedFarm(t *testing.T) *farmT {
	t.Helper()
	f := &farmT{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series = f.id(`insert into series (name, name_key, year, placement_date, age_at_placement_days)
		values ('Seria F', 'seria f', 2030, $1, 140)`, day(time.October, 6))
	f.guide = f.id(`insert into guide (name, name_key, type) values ('Ghid F', 'ghid f', 'laying')`)
	f.exec(`insert into guide_week (guide_id, week, sex, feed_g, lay_pct)
		select $1::uuid, w, 'F'::sex, 110, 80 from generate_series(21, 27) w
		union all select $1::uuid, w, 'M'::sex, 125, null from generate_series(21, 27) w`, f.guide)
	f.farm = f.id(`insert into farm (name, name_key, type, guide_id) values ('Ferma F', 'ferma f', 'laying', $1)`, f.guide)
	f.cycle = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, started_by) values ($1, $2, $3, $4, 'auto')`,
		f.series, f.farm, f.guide, day(time.October, 8))
	sec := func(code string, parent *uuid.UUID, sort int) uuid.UUID {
		id := f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, f.farm, code, parent, sort)
		f.sec[code] = id
		return id
	}
	h5 := sec("H5", nil, 1)
	sec("H5.1", &h5, 1)
	sec("H5.2", &h5, 2)
	h7 := sec("H7", nil, 2)
	sec("H7.1", &h7, 1)
	sec("H7.2", &h7, 2)
	leaves := []uuid.UUID{f.sec["H5.1"], f.sec["H5.2"], f.sec["H7.1"], f.sec["H7.2"]}

	e := f.entry("placement")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1, $2, s, $3, x.sex::sex, 'placement', x.qty
		from unnest($4::uuid[]) s, (values ('F', 100), ('M', 10)) x(sex, qty)`, e, f.farm, day(time.October, 8), leaves)
	e = f.entry("sale")
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
		select $1::uuid, $2::uuid, s, '2026-11-14'::date, x.sex::sex, 'sale'::bird_event_type, x.qty
		from unnest($3::uuid[]) s, (values ('F', -100), ('M', -10)) x(sex, qty)
		union all select $1::uuid, $2::uuid, $4::uuid, '2026-11-14'::date, 'M'::sex, 'sale'::bird_event_type, -5`,
		e, f.farm, []uuid.UUID{f.sec["H7.1"], f.sec["H7.2"]}, f.sec["H5.2"])

	ft := f.id(`insert into feed_type (name, name_key) values ('Breeder F', 'breeder f')`)
	e = f.entry("feed_daily")
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		select $1, $2, s, d::date, x.sex::sex, $4, x.g
		from unnest($3::uuid[]) s, generate_series('2026-10-08'::date, '2026-11-18'::date, interval '1 day') d,
		     (values ('F', 100), ('M', 120)) x(sex, g)
		where not (s in ($5, $6) and d::date >= '2026-11-14')
		  and not (s in ($5, $6) and x.sex = 'F' and d::date = '2026-11-13')
		  and not (s in ($7, $8) and x.sex = 'F' and d::date in ('2026-11-03', '2026-11-04'))
		  and not (s = $8 and x.sex = 'M' and d::date = '2026-11-14')`,
		e, f.farm, leaves, ft, f.sec["H7.1"], f.sec["H7.2"], f.sec["H5.1"], f.sec["H5.2"])
	e = f.entry("migration")
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird, kg_recorded)
		select $1::uuid, $2::uuid, s, d::date, 'F'::sex, $3::uuid, 100::numeric, 6::numeric
		from unnest($4::uuid[]) s, (values ('2026-11-03'), ('2026-11-04')) d(d)
		union all
		select $1::uuid, $2::uuid, s, '2026-11-14'::date, x.sex::sex, $3::uuid, null, 0
		from unnest($5::uuid[]) s, (values ('F'), ('M')) x(sex)
		union all
		select $1::uuid, $2::uuid, $6::uuid, '2026-11-14'::date, 'M'::sex, $3::uuid, 120, 0.5`,
		e, f.farm, ft, []uuid.UUID{f.sec["H5.1"], f.sec["H5.2"]}, []uuid.UUID{f.sec["H7.1"], f.sec["H7.2"]}, f.sec["H5.2"])
	return f
}

func (f *farmT) consum(q TableQuery) TableResponse {
	f.t.Helper()
	resp, err := f.consumAt(q, testdb.Now())
	if err != nil {
		f.t.Fatalf("Build: %v", err)
	}
	return resp
}

func (f *farmT) consumAt(q TableQuery, now time.Time) (TableResponse, error) {
	q.View, q.FarmID = "report-feed", &f.farm
	return Build(context.Background(), f.d.App, Input{Query: q, Today: testdb.Today, Now: now})
}

// feedLine is a Consum row as text: kg, g, guide and deviation ♀ · ♂ at 4
// decimals, the g status and the missing keys.
func feedLine(c Cells, status map[string]string, m *Missing) string {
	s := strings.Join(cellsOf(c, KeyFeedKgF, KeyFeedKgM, KeyFeedGF, KeyFeedGM, KeyFeedGGuideF, KeyFeedGGuideM,
		KeyFeedGDiffF, KeyFeedGDiffM), " ")
	s += " " + status[KeyFeedGF] + "/" + status[KeyFeedGM]
	if m != nil {
		s += " missing " + itoa(m.Rows) + fmt.Sprint(m.Days)
		for _, k := range m.Keys {
			s += " " + k.Code + " " + string(k.Sex) + " " + k.Date.String()
		}
	}
	return s
}

// TestConsumBuild (M3 T7 criterion 3): the default page of the hand-built
// farm, row by row, with the figures worked out in newFeedFarm's terms:
// ♀ 100 birds × 100 g = 10 kg, ♂ 10 × 120 g = 1,2 kg a section and day.
func TestConsumBuild(t *testing.T) {
	f := newFeedFarm(t)
	resp := f.consum(TableQuery{})

	diff(t, "scope", Scope{View: "report-feed", FarmID: &f.farm, CycleID: &f.cycle, SeriesID: &f.series, Label: "Total fermă"}, resp.Scope)
	diff(t, "chips", []string{"hrana"}, resp.Chips)
	diff(t, "groups", []string{"hranaKg", "gPasare"}, groupIDs(resp.Groups))
	diff(t, "kg leaves", []string{"feed.kg.F ♀", "feed.kg.M ♂"}, leafIDs(resp.Groups[0]))
	diff(t, "g leaves", []string{"feed.g.F ♀ efectiv", "feed.g.guide.F ♀ ghid", "feed.g.diffPct.F ♀ față de ghid",
		"feed.g.M ♂ efectiv", "feed.g.guide.M ♂ ghid", "feed.g.diffPct.M ♂ față de ghid", "feed.state stare"}, leafIDs(resp.Groups[1]))
	if resp.Groups[0].Label != "Hrană · kg" || resp.Groups[1].Label != "Hrană · g/pasăre" || resp.Groups[1].Leaves[6].Kind != "status" {
		t.Errorf("groups: %+v", resp.Groups)
	}
	diff(t, "loaded", &Loaded{SaptFrom: 24, SaptTo: 27, Weeks: 4}, resp.Loaded)
	if resp.More == nil || resp.More.Before != 24 || resp.More.Remaining != 3 || resp.Start != nil {
		t.Errorf("more %+v, start %+v", resp.More, resp.Start)
	}
	diff(t, "notes", []string{
		"g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre (♀ după găini, ♂ după cocoși). Toleranță ±10%.",
		"♀ 6 kg/compartiment în registru (g × păsări ≈ 10 kg): SAPT 25: H5 03.11–04.11.",
	}, resp.Notes)

	var weeks []string
	for _, w := range resp.Weeks {
		weeks = append(weeks, w.Label)
	}
	diff(t, "weeks, newest first", []string{"SAPT 27 · 17.11–18.11 · 2 din 7 zile", "SAPT 26 · 10.11–16.11",
		"SAPT 25 · 03.11–09.11", "SAPT 24 · 27.10–02.11"}, weeks)

	got := map[string]string{}
	var sections []string
	for _, w := range resp.Weeks {
		got[w.Label] = feedLine(w.Cells, w.Status, w.Missing)
		if len(w.Days) != w.DaysInRow {
			t.Errorf("%s: %d day rows, want %d", w.Label, len(w.Days), w.DaysInRow)
		}
		for i, d := range w.Days {
			if i > 0 && !d.Date.After(w.Days[i-1].Date) {
				t.Errorf("%s: days out of order", w.Label)
			}
			got[d.Date.String()] = feedLine(d.Cells, d.Status, d.Missing)
		}
		for _, s := range w.Sections {
			sections = append(sections, fmt.Sprint(w.Sapt, " ", s.Code, " ", s.Label))
			got[fmt.Sprint(s.Code, " SAPT ", w.Sapt)] = feedLine(s.Cells, s.Status, s.Missing)
			if s.Sapt != w.Sapt || s.From != w.From || s.To != w.To || s.Days != nil || s.Sections != nil {
				t.Errorf("section row %s: %+v", s.Code, s)
			}
		}
	}
	got["totals"] = resp.Totals.Label + " " + feedLine(resp.Totals.Cells, resp.Totals.Status, resp.Totals.Missing)

	diff(t, "section rows", []string{
		"27 H5.1 H5.1", "27 H5.2 H5.2", "27 H7 goală din 14.11.2026 · H7.1, H7.2",
		"26 H5.1 H5.1", "26 H5.2 H5.2", "26 H7.1 H7.1", "26 H7.2 H7.2",
		"25 H5.1 H5.1", "25 H5.2 H5.2", "25 H7.1 H7.1", "25 H7.2 H7.2",
		"24 H5.1 H5.1", "24 H5.2 H5.2", "24 H7.1 H7.1", "24 H7.2 H7.2",
	}, sections)
	empty := resp.Weeks[0].Sections[2]
	diff(t, "emptied hall line", WeekRow{Code: "H7", Sapt: 27, From: day(time.November, 17), To: day(time.November, 18),
		DaysInRow: 2, Partial: true, Label: "goală din 14.11.2026 · H7.1, H7.2", Cells: Cells{},
		Reasons: map[string]string{"*": "empty_since:2026-11-14"}}, empty)

	const (
		ok    = " in_norma/in_norma"
		guide = " feed.g.guide.F=110 feed.g.guide.M=125"
	)
	for key, want := range map[string]string{
		// SAPT 27: H5 only; ♂ H5.1 10 birds, H5.2 5.
		"SAPT 27 · 17.11–18.11 · 2 din 7 zile": "feed.kg.F=40 feed.kg.M=3.6 feed.g.F=100 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		"2026-11-17":   "feed.kg.F=20 feed.kg.M=1.8 feed.g.F=100 feed.g.M=120" + guide + " feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		"H5.2 SAPT 27": "feed.kg.F=20 feed.kg.M=1.2 feed.g.F=100 feed.g.M=120" + guide + " feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		// SAPT 26: H7 ♀ missing on 13.11; on 14.11 H7's 0 kg rows are not fed
		// and H5.2 ♂ ate 0,5 kg for its 10 birds at the start.
		"SAPT 26 · 10.11–16.11": "feed.kg.F=200 feed.kg.M=24.5 feed.g.F=100 feed.g.M=116.6667" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-6.6667" + ok +
			" missing 2[2026-11-13] H7.1 F 2026-11-13 H7.2 F 2026-11-13",
		"2026-11-13": "feed.kg.F=20 feed.kg.M=4.8 feed.g.F=100 feed.g.M=120" + guide + " feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok +
			" missing 2[2026-11-13] H7.1 F 2026-11-13 H7.2 F 2026-11-13",
		// A day's g is the register's: H5.2 ♂ 0,5 kg at 120 g is for 4,17
		// birds, so the day is 1,7 kg ÷ 14,17 birds = 120 g, not 1,7 ÷ 20.
		"2026-11-14": "feed.kg.F=20 feed.kg.M=1.7 feed.g.F=100 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		"2026-11-15": "feed.kg.F=20 feed.kg.M=1.8 feed.g.F=100 feed.g.M=120" + guide + " feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		"H5.2 SAPT 26": "feed.kg.F=70 feed.kg.M=6.5 feed.g.F=100 feed.g.M=108.3333" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-13.3333 in_norma/abatere",
		"H7.1 SAPT 26": "feed.kg.F=30 feed.kg.M=4.8 feed.g.F=100 feed.g.M=120" + guide + " feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok +
			" missing 1[2026-11-13] H7.1 F 2026-11-13",
		// SAPT 25: 6 kg instead of 10 for H5 ♀ on 03–04.11. The week is kg ÷
		// bird-days; the day keeps the register's 100 g (6 kg for 60 birds).
		"SAPT 25 · 03.11–09.11": "feed.kg.F=264 feed.kg.M=33.6 feed.g.F=94.2857 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-14.2857 feed.g.diffPct.M=-4 abatere/in_norma",
		"2026-11-03": "feed.kg.F=32 feed.kg.M=4.8 feed.g.F=100 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		"H5.1 SAPT 25": "feed.kg.F=62 feed.kg.M=8.4 feed.g.F=88.5714 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-19.4805 feed.g.diffPct.M=-4 abatere/in_norma",
		"SAPT 24 · 27.10–02.11": "feed.kg.F=280 feed.kg.M=33.6 feed.g.F=100 feed.g.M=120" + guide +
			" feed.g.diffPct.F=-9.0909 feed.g.diffPct.M=-4" + ok,
		// 784 kg ÷ 8 000 ♀ bird-days, 95,3 kg ÷ 800 ♂ bird-days.
		"totals": "Total SAPT 24–27 feed.kg.F=784 feed.kg.M=95.3 feed.g.F=98 feed.g.M=119.125" + guide +
			" feed.g.diffPct.F=-10.9091 feed.g.diffPct.M=-4.7 abatere/in_norma" +
			" missing 2[2026-11-13] H7.1 F 2026-11-13 H7.2 F 2026-11-13",
	} {
		if got[key] != want {
			t.Errorf("%s:\n got %s\nwant %s", key, got[key], want)
		}
	}
}

// TestConsumNullCells: a row without a ration is "–", "lipsă" when a section
// × sex with birds has none; a sold-out section's 0 kg row is 0 kg, not fed.
func TestConsumNullCells(t *testing.T) {
	f := newFeedFarm(t)
	// H7.1 ♀ has no ration in SAPT 26 at all (its 0 kg row of 14.11 gone
	// too): kg and g are "lipsă".
	f.exec(`update feed_allocation set deleted_at = now() where section_id = $1 and sex = 'F' and date between '2026-11-10' and '2026-11-14'`,
		f.sec["H7.1"])
	resp := f.consum(TableQuery{Sapt: intp(26)})
	if len(resp.Weeks) != 1 || resp.Totals.Label != "Total SAPT 26" {
		t.Fatalf("sapt=26: %d weeks, totals %q", len(resp.Weeks), resp.Totals.Label)
	}
	var h71 WeekRow
	for _, s := range resp.Weeks[0].Sections {
		if s.Code == "H7.1" {
			h71 = s
		}
	}
	diff(t, "H7.1 without ♀ ration", []string{"feed.kg.F=null", "feed.g.F=null", "feed.g.guide.F=110", "feed.g.diffPct.F=null"},
		cellsOf(h71.Cells, KeyFeedKgF, KeyFeedGF, KeyFeedGGuideF, KeyFeedGDiffF))
	diff(t, "H7.1 reasons", map[string]string{"feed.kg.F": "missing", "feed.g.F": "missing"}, h71.Reasons)
	if h71.Missing == nil || h71.Missing.Rows != 4 {
		t.Errorf("H7.1 missing = %+v, want 4 rows (10–13.11)", h71.Missing)
	}
	// 14.11 empties H7.1: its ♀ row without a ration is not missing.
	day14 := resp.Weeks[0].Days[4]
	if day14.Date != day(time.November, 14) || cell(day14.Cells, KeyFeedKgF) != "20" || day14.Missing != nil {
		t.Errorf("14.11 = %+v", day14)
	}
}

// TestConsumMissingToday: today's gaps count only from missing.from_hour
// (12:00, §5.11).
func TestConsumMissingToday(t *testing.T) {
	f := newFeedFarm(t)
	f.exec(`update feed_allocation set deleted_at = now() where section_id = $1 and sex = 'F' and date = '2026-11-18'`, f.sec["H5.1"])
	for _, tc := range []struct {
		hour int
		want string
	}{{10, "<nil>"}, {12, "&{1 [2026-11-18] [{H5.1 F 2026-11-18}]}"}} {
		resp, err := f.consumAt(TableQuery{}, time.Date(2026, 11, 18, tc.hour, 0, 0, 0, timex.Chisinau))
		if err != nil {
			t.Fatal(err)
		}
		if got := fmt.Sprint(resp.Weeks[0].Days[1].Missing); got != tc.want {
			t.Errorf("%d:00: 18.11 missing %s, want %s", tc.hour, got, tc.want)
		}
	}
}

// TestConsumGrainPagingAndStates: grain=week drops the day rows, load-more
// reaches the cycle start, a hall emptied before the loaded days keeps its
// line, no guide drops the guide leaves, and a period
// without a ration is the empty state.
func TestConsumGrainPagingAndStates(t *testing.T) {
	f := newFeedFarm(t)
	resp := f.consum(TableQuery{Grain: "week"})
	for _, w := range resp.Weeks {
		if w.Days != nil || len(w.Sections) == 0 {
			t.Errorf("grain=week %s: %d days, %d sections", w.Label, len(w.Days), len(w.Sections))
		}
	}
	resp = f.consum(TableQuery{Before: intp(24)})
	if len(resp.Weeks) != 3 || resp.More != nil || resp.Start == nil || resp.Start.Label != "Începutul ciclului · SAPT 21" ||
		resp.Totals.Label != "Total SAPT 21–27" {
		t.Errorf("load more: %d weeks, more %+v, start %+v, totals %q", len(resp.Weeks), resp.More, resp.Start, resp.Totals.Label)
	}
	// H7 emptied before the loaded days: its line still says since when.
	resp = f.consum(TableQuery{Period: Period{Kind: "interval", From: day(time.November, 17), To: day(time.November, 18)}})
	if s := resp.Weeks[0].Sections; len(s) != 3 || s[2].Label != "goală din 14.11.2026 · H7.1, H7.2" {
		t.Errorf("interval 17–18.11 sections: %+v", s)
	}

	f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
	resp = f.consum(TableQuery{})
	diff(t, "g leaves without a guide", []string{"feed.g.F ♀ efectiv", "feed.g.M ♂ efectiv", "feed.state stare"}, leafIDs(resp.Groups[1]))
	if !resp.NoGuide || resp.Weeks[0].Status != nil {
		t.Errorf("noGuide %v, status %v", resp.NoGuide, resp.Weeks[0].Status)
	}
	diff(t, "cells without a guide", []string{"feed.kg.F", "feed.kg.M", "feed.g.F", "feed.g.M"}, presentKeys(resp.Weeks[0].Cells))

	f.exec(`update feed_allocation set deleted_at = now() where farm_id = $1`, f.farm)
	resp = f.consum(TableQuery{})
	if resp.Empty == nil || len(resp.Weeks) != 0 || len(resp.Chips) != 0 || resp.Totals != nil || resp.Period.WeeksTotal != 7 {
		t.Errorf("no ration: %+v", resp)
	}

	if _, err := f.consumAt(TableQuery{Detail: "halls"}, testdb.Now()); !errors.Is(err, ErrUnsupported) {
		t.Errorf("detail: %v, want ErrUnsupported", err)
	}
	if _, err := f.consumAt(TableQuery{Types: []string{ChipFeed}}, testdb.Now()); !errors.Is(err, ErrUnsupported) {
		t.Errorf("types: %v, want ErrUnsupported", err)
	}
}

// TestConsumStatements (M3 T7 criterion 2): a page is the cycle read, one
// Consum day query and the guide and settings reads.
func TestConsumStatements(t *testing.T) {
	f := newFeedFarm(t)
	pool, st := tracedPool(t, f)
	run := func(t *testing.T, q TableQuery, want ...string) {
		t.Helper()
		q.View, q.FarmID = "report-feed", &f.farm
		if _, err := Build(context.Background(), pool, Input{Query: q, Today: testdb.Today, Now: testdb.Now()}); err != nil {
			t.Fatal(err)
		}
		if got := st.take(); !slices.Equal(got, want) {
			t.Errorf("statements %q, want %q", got, want)
		}
	}
	t.Run("page 1", func(t *testing.T) {
		run(t, TableQuery{}, "CalcFarmCycles", "PeriodTableFeedDays", "CalcGuideWeeks", "CalcSettings")
	})
	t.Run("whole cycle", func(t *testing.T) {
		run(t, TableQuery{Count: 20}, "CalcFarmCycles", "PeriodTableFeedDays", "CalcGuideWeeks", "CalcSettings")
	})
	t.Run("no guide", func(t *testing.T) {
		f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
		run(t, TableQuery{}, "CalcFarmCycles", "PeriodTableFeedDays", "CalcSettings")
	})
}

// TestConsumResponseJSON: code, missing and sections are written for Consum
// and left out of the Rezumat rows.
func TestConsumResponseJSON(t *testing.T) {
	f := newFeedFarm(t)
	b, err := json.Marshal(f.consum(TableQuery{}))
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	for _, want := range []string{
		`"code":"H7","sapt":27`, `"cells":{},"reasons":{"*":"empty_since:2026-11-14"}`,
		`"missing":{"rows":2,"days":["2026-11-13"],"keys":[{"code":"H7.1","sex":"F","date":"2026-11-13"},{"code":"H7.2","sex":"F","date":"2026-11-13"}]}`,
		`"sections":[{"code":"H5.1"`, `"view":"report-feed"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("JSON lacks %s", want)
		}
	}
	b, err = json.Marshal(newFarmT(t).build(TableQuery{}))
	if err != nil {
		t.Fatal(err)
	}
	for _, absent := range []string{`"code"`, `"missing"`, `"sections"`} {
		if strings.Contains(string(b), absent) {
			t.Errorf("Rezumat JSON has %s", absent)
		}
	}
}

// TestConsumQuery (M3 T7 criterion 1): report-feed takes the D11 parameters
// and ignores tab; detail and types are out_of_range.
func TestConsumQuery(t *testing.T) {
	const base = "view=report-feed&farmId=11111111-1111-4111-8111-111111111111"
	for _, raw := range []string{
		base,
		base + "&cycleId=" + qCycle.String() + "&period=weeks:4&grain=week&before=56&count=8",
		base + "&period=interval:2026-11-01,2026-11-15&sapt=58",
		base + "&before=56&fromSapt=20",
		base + "&tab=pasari",
		base + "&tab=x",
	} {
		q, err := ParseQuery(mustValues(t, raw))
		if err == nil {
			err = httpx.Validate(q)
		}
		if err != nil || q.View != "report-feed" || q.FarmID == nil || *q.FarmID != qFarm {
			t.Errorf("%s: %+v, %v", raw, q, err)
		}
	}
	type issue struct{ Path, Code string }
	for _, tc := range []struct {
		raw  string
		want []issue
	}{
		{base + "&detail=farm", []issue{{"detail", "out_of_range"}}},
		{base + "&detail=halls", []issue{{"detail", "out_of_range"}}},
		{base + "&types=hrana", []issue{{"types", "out_of_range"}}},
		{base + "&types=x", []issue{{"types", "out_of_range"}}},
		{"view=report-feed", []issue{{"farmId", "required"}}},
		{base + "&sectionId=" + qOther.String(), []issue{{"sectionId", "out_of_range"}}},
		{base + "&seriesId=" + qOther.String(), []issue{{"seriesId", "out_of_range"}}},
		{base + "&grain=month", []issue{{"grain", "out_of_range"}}},
		{base + "&sapt=55&before=56", []issue{{"sapt", "out_of_range"}}},
	} {
		q, err := ParseQuery(mustValues(t, tc.raw))
		if err == nil {
			err = httpx.Validate(q)
		}
		var he *httpx.Error
		if !errors.As(err, &he) || he.Status != 400 || he.Code != httpx.CodeValidation {
			t.Errorf("%s: %v, want a VALIDATION error", tc.raw, err)
			continue
		}
		got := []issue{}
		for _, i := range he.Issues {
			got = append(got, issue{i.Path, i.Code})
		}
		diff(t, tc.raw, tc.want, got)
	}
}

// TestConsumHandler: GET /api/table?view=report-feed goes through the same
// route, roles and farm check as the Rezumat.
func TestConsumHandler(t *testing.T) {
	f := newFarmL(t)
	q := "view=report-feed&farmId=" + f.farm.String()
	for _, tc := range []struct {
		name   string
		caller *auth.Principal
		query  string
		status int
		code   string
	}{
		{"vet", caller(auth.Vet, f.farm), q, 403, "FORBIDDEN"},
		{"viewer on another farm", caller(auth.Viewer, f.other), q, 403, "FARM_FORBIDDEN"},
		{"owner on an unknown farm", caller(auth.Owner), "view=report-feed&farmId=" + uuid.NewString(), 404, "NOT_FOUND"},
		{"detail", caller(auth.Owner), q + "&detail=farm", 400, "VALIDATION"},
		{"viewer, tab ignored", caller(auth.Viewer, f.farm), q + "&tab=hrana", 200, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := getTable(t, f.server(tc.caller), tc.query)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if tc.status != 200 {
				var body tableError
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body.Error.Code != tc.code {
					t.Errorf("error %s, want %s (%v)", body.Error.Code, tc.code, err)
				}
				return
			}
			// Ferma L has no ration: the empty state.
			var res TableResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatal(err)
			}
			if res.Scope.View != "report-feed" || res.Empty == nil || len(res.Weeks) != 0 {
				t.Errorf("body %s", rec.Body)
			}
		})
	}
}

// presentKeys are the Consum keys of c in group order.
func presentKeys(c Cells) []string {
	keys := make([]string, 0, len(c))
	for _, k := range []string{KeyFeedKgF, KeyFeedKgM, KeyFeedGF, KeyFeedGM, KeyFeedGGuideF, KeyFeedGGuideM, KeyFeedGDiffF, KeyFeedGDiffM} {
		if _, ok := c[k]; ok {
			keys = append(keys, k)
		}
	}
	if len(keys) != len(c) {
		keys = append(keys, "…unexpected keys")
	}
	return keys
}
