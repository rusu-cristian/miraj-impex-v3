package migratehistory

import (
	"cmp"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// extractsDir is design/data/extracts: $IMPEX_EXTRACTS_DIR, else seen from
// this package's directory (M1-D6).
func extractsDir() string {
	if d := os.Getenv("IMPEX_EXTRACTS_DIR"); d != "" {
		return d
	}
	return filepath.Join("..", "..", "..", "design", "data", "extracts")
}

var testNow = time.Date(2026, 11, 18, 12, 0, 0, 0, timex.Chisinau)

func readExtracts(t *testing.T) *Extracts {
	t.Helper()
	x, err := ReadExtracts(extractsDir())
	if err != nil {
		t.Fatal(err)
	}
	return x
}

var (
	planOnce sync.Once
	planP    *Plan
	planErr  error
)

// testPlan is the plan of the real extracts, built once and shared read-only.
func testPlan(t *testing.T) *Plan {
	t.Helper()
	planOnce.Do(func() {
		x, err := ReadExtracts(extractsDir())
		if err != nil {
			planErr = err
			return
		}
		planP, planErr = BuildPlan(x, testNow)
	})
	if planErr != nil {
		t.Fatal(planErr)
	}
	return planP
}

func day(s string) timex.Day {
	d, err := timex.ParseDay(s)
	if err != nil {
		panic(err)
	}
	return d
}

// names resolves ids of a plan to readable labels.
type names struct {
	p       *Plan
	code    map[uuid.UUID]string
	farm    map[uuid.UUID]string
	entries map[uuid.UUID]dbq.Entry
}

func namesOf(p *Plan) names {
	n := names{p: p, code: map[uuid.UUID]string{}, farm: map[uuid.UUID]string{}, entries: map[uuid.UUID]dbq.Entry{}}
	for _, s := range p.Sections {
		n.code[s.ID] = s.Code
	}
	for _, key := range []string{FarmTelita, FarmHirbovat} {
		id, _ := p.FarmID(key)
		n.farm[id] = key
	}
	for _, e := range p.Entries {
		n.entries[e.ID] = e
	}
	return n
}

// move is a bird_event row in words.
type move struct {
	Farm, Section string
	Sex           dbq.Sex
	Date          string
	Qty           int32
	Peer          string
	Note          string
}

func (n names) move(e dbq.BirdEvent) move {
	m := move{Farm: n.farm[e.FarmID], Section: n.code[e.SectionID], Sex: e.Sex, Date: e.Date.String(), Qty: e.Qty}
	if e.PeerSectionID != nil {
		m.Peer = n.code[*e.PeerSectionID]
	}
	if e.Note != nil {
		m.Note = *e.Note
	}
	return m
}

func sortMoves(ms []move) {
	slices.SortFunc(ms, func(a, b move) int {
		return cmp.Or(cmp.Compare(a.Date, b.Date), cmp.Compare(a.Farm, b.Farm), cmp.Compare(a.Section, b.Section),
			cmp.Compare(a.Sex, b.Sex), cmp.Compare(a.Peer, b.Peer), cmp.Compare(a.Qty, b.Qty))
	})
}

func TestPlanGuides(t *testing.T) {
	p := testPlan(t)
	// Golden 36: ouat SAPT 19–65 (94 rows), creștere SAPT 1–25 (50 rows).
	byGuide := map[string][]dbq.GuideWeek{}
	guideName := map[uuid.UUID]string{}
	for _, g := range p.Guides {
		guideName[g.ID] = g.Name
	}
	for _, w := range p.GuideWeeks {
		byGuide[guideName[w.GuideID]] = append(byGuide[guideName[w.GuideID]], w)
	}
	for _, tc := range []struct {
		name     string
		rows     int
		from, to int16
	}{
		{"Ross 308 ouat", 94, 19, 65},
		{"Ross 308 creștere", 50, 1, 25},
	} {
		ws := byGuide[tc.name]
		if len(ws) != tc.rows {
			t.Errorf("%s: %d rows, want %d", tc.name, len(ws), tc.rows)
			continue
		}
		lo, hi := ws[0].Week, ws[0].Week
		for _, w := range ws {
			lo, hi = min(lo, w.Week), max(hi, w.Week)
			if w.Sex == dbq.SexM && (w.LayPct.Valid || w.MortalityCumPct.Valid) {
				t.Errorf("%s SAPT %d ♂ has a lay or mortality guide", tc.name, w.Week)
			}
			if w.HatchingPct.Valid || w.EggWeightG.Valid {
				t.Errorf("%s SAPT %d: hatching / egg weight are not imported (D184)", tc.name, w.Week)
			}
		}
		if lo != tc.from || hi != tc.to {
			t.Errorf("%s spans SAPT %d–%d, want %d–%d", tc.name, lo, hi, tc.from, tc.to)
		}
	}
	// Golden 13 inputs: guide lay SAPT 59 57,6 · SAPT 53 64,5 · SAPT 56 61,1.
	lay := map[int16]string{}
	for _, w := range byGuide["Ross 308 ouat"] {
		if w.Sex == dbq.SexF && w.LayPct.Valid {
			lay[w.Week] = w.LayPct.Decimal.String()
		}
	}
	if lay[59] != "57.6" || lay[53] != "64.5" || lay[56] != "61.1" {
		t.Errorf("guide lay SAPT 59/53/56 = %s/%s/%s, want 57.6/64.5/61.1", lay[59], lay[53], lay[56])
	}
	if p.Guides[0].NameKey != "ross308crestere" {
		t.Errorf("growth guide name_key = %q", p.Guides[0].NameKey)
	}
}

func TestPlanFarmsAndSections(t *testing.T) {
	p := testPlan(t)
	type sec struct {
		Farm, Code, Parent string
		Sort               int16
	}
	n := namesOf(p)
	var got []sec
	for _, s := range p.Sections {
		parent := ""
		if s.ParentID != nil {
			parent = n.code[*s.ParentID]
		}
		got = append(got, sec{n.farm[s.FarmID], s.Code, parent, s.SortOrder})
	}
	want := []sec{
		{"telita", "H1", "", 0}, {"telita", "H2", "", 1}, {"telita", "H3", "", 2}, {"telita", "H6", "", 3},
		{"telita", "H5-1", "", 4}, {"telita", "H5-2", "", 5},
		{"hirbovat", "H5", "", 0}, {"hirbovat", "H5.1", "H5", 0}, {"hirbovat", "H5.2", "H5", 1},
		{"hirbovat", "H6", "", 1}, {"hirbovat", "H6.1", "H6", 0}, {"hirbovat", "H6.2", "H6", 1},
		{"hirbovat", "H7", "", 2}, {"hirbovat", "H7.1", "H7", 0}, {"hirbovat", "H7.2", "H7", 1},
	}
	if d := gocmp.Diff(want, got); d != "" {
		t.Errorf("sections (-want +got):\n%s", d)
	}
	type farm struct {
		Name, Key string
		Type      dbq.FarmType
		Guide     string
	}
	guideName := map[uuid.UUID]string{}
	for _, g := range p.Guides {
		guideName[g.ID] = g.Name
	}
	var farms []farm
	for _, f := range p.Farms {
		farms = append(farms, farm{f.Name, f.NameKey, f.Type, guideName[*f.GuideID]})
	}
	wantFarms := []farm{
		{"Telița", "telita", dbq.FarmTypeGrowth, "Ross 308 creștere"},
		{"Hîrbovăț", "hirbovat", dbq.FarmTypeLaying, "Ross 308 ouat"},
	}
	if d := gocmp.Diff(wantFarms, farms); d != "" {
		t.Errorf("farms (-want +got):\n%s", d)
	}
}

func TestPlanSetup(t *testing.T) {
	p := testPlan(t)
	type ft struct {
		Name, Key string
		ForSex    string
		Active    bool
	}
	var fts []ft
	for _, f := range p.FeedTypes {
		sex := ""
		if f.ForSex != nil {
			sex = string(*f.ForSex)
		}
		fts = append(fts, ft{f.Name, f.NameKey, sex, f.Active})
	}
	wantFT := []ft{
		{"Starter", "starter", "", false}, {"Grower", "grower", "", false}, {"Pre-ouat", "pre-ouat", "", false},
		{"Breeder 2", "breeder2", "F", true}, {"Cocoși", "cocosi", "M", true},
	}
	if d := gocmp.Diff(wantFT, fts); d != "" {
		t.Errorf("feed types (-want +got):\n%s", d)
	}

	n := namesOf(p)
	type series struct {
		Name, Key, Placement string
		Year                 int16
		Status               dbq.SeriesStatus
		Inferred             bool
		PlacementFarm        string
		Closed               bool
	}
	var ss []series
	seriesName := map[uuid.UUID]string{}
	for _, s := range p.Series {
		seriesName[s.ID] = s.Name
		pf := ""
		if s.PlacementFarmID != nil {
			pf = n.farm[*s.PlacementFarmID]
		}
		if s.ClosedAt != nil && !s.ClosedAt.Equal(testNow) {
			t.Errorf("%s closed_at = %s, want the clock's now", s.Name, s.ClosedAt)
		}
		ss = append(ss, series{s.Name, s.NameKey, s.PlacementDate.String(), s.Year, s.Status, s.PlacementInferred, pf, s.ClosedAt != nil})
	}
	wantSeries := []series{
		{"Seria 2026", "seria2026", "2025-10-07", 2026, dbq.SeriesStatusActive, false, "telita", false},
		{"Seria 2025", "seria2025", "2024-11-01", 2025, dbq.SeriesStatusClosed, true, "", true},
	}
	if d := gocmp.Diff(wantSeries, ss); d != "" {
		t.Errorf("series (-want +got):\n%s", d)
	}

	// M1-D1: §10.3 wins over cycles.json.
	type cycle struct {
		Series, Farm, Start, End string
		StartedBy                dbq.CycleStartKind
		Inferred                 bool
		EndedBy                  string
	}
	var cs []cycle
	for _, c := range p.Cycles {
		e, eb := "", ""
		if c.EndDate != nil {
			e = c.EndDate.String()
		}
		if c.EndedBy != nil {
			eb = string(*c.EndedBy)
		}
		if c.GuideID == nil || c.CreatedBy != nil {
			t.Errorf("cycle %s %s: guide must be set and author null", n.farm[c.FarmID], c.StartDate)
		}
		cs = append(cs, cycle{seriesName[c.SeriesID], n.farm[c.FarmID], c.StartDate.String(), e, c.StartedBy, c.StartInferred, eb})
	}
	wantCycles := []cycle{
		{"Seria 2026", "telita", "2025-10-07", "2026-02-28", dbq.CycleStartKindAuto, false, "emptied"},
		{"Seria 2026", "hirbovat", "2026-02-22", "", dbq.CycleStartKindAuto, false, ""},
		{"Seria 2025", "hirbovat", "2025-03-07", "2026-01-28", dbq.CycleStartKindMigration, true, "migration"},
	}
	if d := gocmp.Diff(wantCycles, cs); d != "" {
		t.Errorf("cycles (-want +got):\n%s", d)
	}
}

// transferRows returns the bird rows of transfer entries matching keep, in words.
func transferRows(p *Plan, keep func(dbq.Entry) bool) []move {
	n := namesOf(p)
	var out []move
	for _, e := range p.BirdEvents {
		en := n.entries[e.EntryID]
		if en.Kind == dbq.EntryKindTransfer && keep(en) {
			out = append(out, n.move(e))
		}
	}
	sortMoves(out)
	return out
}

func TestPlanTransferLinks(t *testing.T) {
	p := testPlan(t)
	telita, _ := p.FarmID(FarmTelita)
	hirbovat, _ := p.FarmID(FarmHirbovat)
	// Every link is one entry from Telița to Hîrbovăț on the arrival day that
	// keeps the departure day in its note (H2: out 27.02, in 28.02).
	type entry struct{ Date, Note string }
	var entries []entry
	for _, e := range p.Entries {
		if e.Kind == dbq.EntryKindTransfer && e.FarmID == telita && e.PeerFarmID != nil && *e.PeerFarmID == hirbovat {
			note := ""
			if e.Note != nil {
				note = *e.Note
			}
			entries = append(entries, entry{e.Date.String(), note})
		}
	}
	wantEntries := []entry{
		{"2026-02-22", "plecare 22.02.2026"}, {"2026-02-25", "plecare 25.02.2026"}, {"2026-02-26", "plecare 26.02.2026"},
		{"2026-02-27", "plecare 27.02.2026"}, {"2026-02-28", "plecare 27.02.2026"},
	}
	if d := gocmp.Diff(wantEntries, entries); d != "" {
		t.Errorf("link entries (-want +got):\n%s", d)
	}

	got := transferRows(p, func(e dbq.Entry) bool { return e.FarmID == telita && *e.PeerFarmID == hirbovat })
	pair := func(date, from, to string, sex dbq.Sex, q int32) []move {
		return []move{
			{Farm: "telita", Section: from, Sex: sex, Date: date, Qty: -q, Peer: to},
			{Farm: "hirbovat", Section: to, Sex: sex, Date: date, Qty: q, Peer: from},
		}
	}
	// Code order split (§10.3); per Q1 (a) the pairs carry the arrived
	// quantity: ♂ H5-2 → H7.2 534 (504 out + 30), ♀ H2 → H7.2 4 383 (4 381 + 2).
	want := slices.Concat(
		pair("2026-02-22", "H5-1", "H5.1", "M", 498), pair("2026-02-22", "H5-1", "H5.2", "M", 498),
		pair("2026-02-22", "H5-1", "H6.1", "M", 391), pair("2026-02-22", "H5-2", "H6.1", "M", 107),
		pair("2026-02-22", "H5-2", "H6.2", "M", 498), pair("2026-02-22", "H5-2", "H7.1", "M", 498),
		pair("2026-02-22", "H5-2", "H7.2", "M", 534),
		pair("2026-02-25", "H6", "H5.1", "F", 4500), pair("2026-02-25", "H6", "H5.2", "F", 1283),
		pair("2026-02-26", "H3", "H5.2", "F", 3217), pair("2026-02-26", "H3", "H6.1", "F", 3855),
		pair("2026-02-27", "H1", "H6.1", "F", 645), pair("2026-02-27", "H1", "H6.2", "F", 4500),
		pair("2026-02-27", "H1", "H7.1", "F", 1882),
		pair("2026-02-28", "H2", "H7.1", "F", 2618), pair("2026-02-28", "H2", "H7.2", "F", 4383),
	)
	sortMoves(want)
	if d := gocmp.Diff(want, got); d != "" {
		t.Errorf("link pairs (-want +got):\n%s", d)
	}
	// Golden 9: the Hîrbovăț base (transfer in) is the arrivals.
	in := map[dbq.Sex]int32{}
	out := map[dbq.Sex]int32{}
	for _, m := range got {
		if m.Farm == "hirbovat" {
			in[m.Sex] += m.Qty
		} else {
			out[m.Sex] -= m.Qty
		}
	}
	if in[dbq.SexF] != 26883 || in[dbq.SexM] != 3024 {
		t.Errorf("Hîrbovăț arrivals ♀ %d · ♂ %d, want 26 883 · 3 024", in[dbq.SexF], in[dbq.SexM])
	}
	// Telița's net out stays the register's 26 881 · 2 994 with the +2/+30.
	adj := map[dbq.Sex]int32{}
	for _, e := range p.BirdEvents {
		if e.FarmID == telita && e.Type == dbq.BirdEventTypeAdjustment &&
			((e.Date == day("2026-02-22") && e.Sex == dbq.SexM && e.Qty == 30) || (e.Date == day("2026-02-28") && e.Sex == dbq.SexF && e.Qty == 2)) {
			adj[e.Sex] += e.Qty
		}
	}
	if out[dbq.SexF]-adj[dbq.SexF] != 26881 || out[dbq.SexM]-adj[dbq.SexM] != 2994 {
		t.Errorf("Telița net out ♀ %d · ♂ %d, want 26 881 · 2 994", out[dbq.SexF]-adj[dbq.SexF], out[dbq.SexM]-adj[dbq.SexM])
	}
}

func TestPlanTelitaMoves(t *testing.T) {
	p := testPlan(t)
	telita, _ := p.FarmID(FarmTelita)
	moves := 0
	for _, e := range p.Entries {
		if e.Kind == dbq.EntryKindTransfer && e.FarmID == telita && *e.PeerFarmID == telita {
			moves++
		}
	}
	if moves != 16 {
		t.Errorf("Telița internal move entries = %d, want 16", moves)
	}
	got := transferRows(p, func(e dbq.Entry) bool { return e.FarmID == telita && *e.PeerFarmID == telita })
	type mv struct {
		date, from, to string
		sf, st         dbq.Sex
		q              int32
	}
	var want []move
	for _, m := range []mv{
		{"2025-11-05", "H1", "H3", "F", "F", 3338}, {"2025-11-06", "H2", "H3", "F", "F", 3341},
		{"2025-11-11", "H5-1", "H5-2", "M", "M", 1652}, {"2025-11-13", "H1", "H3", "F", "F", 341},
		{"2025-11-24", "H1", "H3", "F", "F", 100}, {"2025-11-24", "H2", "H3", "F", "F", 200},
		{"2025-12-09", "H5-1", "H5-2", "M", "M", 26}, {"2025-12-16", "H1", "H6", "F", "F", 2040},
		{"2025-12-17", "H2", "H6", "F", "F", 1495}, {"2025-12-18", "H3", "H6", "F", "F", 2443},
		{"2025-12-23", "H2", "H3", "F", "F", 830}, {"2025-12-24", "H2", "H3", "F", "F", 660},
		{"2025-12-25", "H1", "H3", "F", "F", 931}, {"2026-01-13", "H3", "H6", "F", "F", 131},
		// Q2: one pair with a sex change, as recorded.
		{"2026-02-23", "H6", "H5-1", "F", "M", 277},
		{"2026-02-24", "H5-1", "H5-2", "M", "M", 286},
	} {
		want = append(want,
			move{Farm: "telita", Section: m.from, Sex: m.sf, Date: m.date, Qty: -m.q, Peer: m.to},
			move{Farm: "telita", Section: m.to, Sex: m.st, Date: m.date, Qty: m.q, Peer: m.from})
	}
	sortMoves(want)
	if d := gocmp.Diff(want, got); d != "" {
		t.Errorf("Telița moves (-want +got):\n%s", d)
	}
}

func TestPlanTelitaAdjustments(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	telita, _ := p.FarmID(FarmTelita)
	var got []move
	for _, e := range p.BirdEvents {
		if e.Type == dbq.BirdEventTypeAdjustment {
			if e.FarmID != telita {
				t.Errorf("adjustment outside Telița: %+v", n.move(e))
			}
			got = append(got, n.move(e))
		}
	}
	sortMoves(got)
	adj := func(date, sec string, sex dbq.Sex, q int32) move {
		return move{Farm: "telita", Section: sec, Sex: sex, Date: date, Qty: q, Note: "diferență registru"}
	}
	want := []move{
		// other_corrections_telita
		adj("2026-02-21", "H5-1", "M", 11), adj("2026-02-24", "H1", "F", -39), adj("2026-02-24", "H2", "F", -48),
		adj("2026-02-24", "H3", "F", -46), adj("2026-02-24", "H6", "F", -23), adj("2026-02-25", "H1", "F", -16),
		adj("2026-02-25", "H3", "F", -14), adj("2026-02-25", "H5-1", "M", 144), adj("2026-02-25", "H5-2", "M", -42),
		adj("2026-02-26", "H5-1", "M", 16), adj("2026-02-27", "H5-1", "M", 14), adj("2026-02-28", "H2", "F", -14),
		// link differences at the source (Q1 (a))
		adj("2026-02-22", "H5-2", "M", 30), adj("2026-02-28", "H2", "F", 2),
		// residuals: the day a section would go below 0, then the rest on 28.02
		adj("2026-02-23", "H5-2", "M", 2), adj("2026-02-25", "H6", "F", 12), adj("2026-02-26", "H3", "F", 9),
		adj("2026-02-28", "H1", "F", -23), adj("2026-02-28", "H2", "F", -16),
		adj("2026-02-28", "H5-1", "M", -454), adj("2026-02-28", "H5-2", "M", -244),
	}
	sortMoves(want)
	if d := gocmp.Diff(want, got); d != "" {
		t.Errorf("Telița adjustments (-want +got):\n%s", d)
	}
	// Telița is empty at the end of 28.02.2026, so its cycle ends emptied.
	alive := map[string]int64{}
	for _, e := range p.BirdEvents {
		if e.FarmID == telita {
			alive[n.code[e.SectionID]+" "+string(e.Sex)] += int64(e.Qty)
		}
	}
	for k, v := range alive {
		if v != 0 {
			t.Errorf("Telița %s ends at %d, want 0", k, v)
		}
	}
}

func TestPlanHirbovatBirds(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	hirbovat, _ := p.FarmID(FarmHirbovat)

	// 7 sale entries (DATA04): 14.11 is H6.2, then H7.1 + H7.2.
	type sale struct {
		Date     string
		Sections string
		F, M     int32
	}
	var sales []sale
	idx := map[uuid.UUID]int{}
	for _, e := range p.Entries {
		if e.Kind == dbq.EntryKindSale {
			if e.FarmID != hirbovat || e.Source != dbq.RecordSourceMigration || e.CreatedBy != nil {
				t.Errorf("sale entry %s: farm, source or author wrong", e.Date)
			}
			idx[e.ID] = len(sales)
			sales = append(sales, sale{Date: e.Date.String()})
		}
	}
	var soldF, soldM int32
	for _, e := range p.BirdEvents {
		i, ok := idx[e.EntryID]
		if !ok {
			continue
		}
		if e.Type != dbq.BirdEventTypeSale {
			t.Errorf("sale entry holds a %s row", e.Type)
		}
		s := &sales[i]
		if code := n.code[e.SectionID]; !strings.Contains(s.Sections, code) {
			s.Sections = strings.TrimPrefix(s.Sections+"+"+code, "+")
		}
		if e.Sex == dbq.SexF {
			s.F -= e.Qty
			soldF -= e.Qty
		} else {
			s.M -= e.Qty
			soldM -= e.Qty
		}
	}
	wantSales := []sale{
		{"2026-11-12", "H7.2", 1442, 0},
		{"2026-11-14", "H6.2", 305, 53},
		{"2026-11-14", "H7.1+H7.2", 6915, 639},
		{"2026-11-15", "H6.2", 1077, 0},
		{"2026-11-16", "H6.2", 807, 57},
		{"2026-11-17", "H6.2", 424, 144},
		{"2026-11-18", "H6.1+H6.2", 1831, 289},
	}
	if d := gocmp.Diff(wantSales, sales); d != "" {
		t.Errorf("sale entries (-want +got):\n%s", d)
	}
	if soldF != 12801 || soldM != 1182 { // golden 11
		t.Errorf("sold ♀ %d · ♂ %d, want 12 801 · 1 182", soldF, soldM)
	}

	// The 22.10.2026 move H7.2 → H7.1 ♂ 19: one entry inside Hîrbovăț.
	got := transferRows(p, func(e dbq.Entry) bool { return e.FarmID == hirbovat })
	want := []move{
		{Farm: "hirbovat", Section: "H7.1", Sex: "M", Date: "2026-10-22", Qty: 19, Peer: "H7.2"},
		{Farm: "hirbovat", Section: "H7.2", Sex: "M", Date: "2026-10-22", Qty: -19, Peer: "H7.1"},
	}
	if d := gocmp.Diff(want, got); d != "" {
		t.Errorf("Hîrbovăț move (-want +got):\n%s", d)
	}

	// The running count equals the register's alive on every compartment-day.
	net := dailyNet(p.BirdEvents, func(e dbq.BirdEvent) bool { return e.FarmID == hirbovat })
	days := map[sexSection][]timex.Day{}
	for k, m := range net {
		days[k] = sortedDays(m)
	}
	aliveOn := func(k sexSection, d timex.Day) int64 {
		var sum int64
		for _, dd := range days[k] {
			if dd.After(d) {
				break
			}
			sum += net[k][dd]
		}
		return sum
	}
	bad := 0
	for _, r := range readExtracts(t).hirbovat.sectionDay {
		sec, _ := p.SectionID(FarmHirbovat, r.Section)
		k := sexSection{sec, r.Sex}
		if got := aliveOn(k, r.Date); got != int64(r.Alive) && bad < 5 {
			bad++
			t.Errorf("Hîrbovăț %s %s %s: alive %d, register %d", r.Section, r.Sex, r.Date, got, r.Alive)
		}
	}

	// Goldens 5 and 7: alive at the end of 18.11.2026.
	type ac struct {
		Section string
		F, M    int64
	}
	var alive []ac
	var totF, totM int64
	for _, code := range []string{"H5.1", "H5.2", "H6.1", "H6.2", "H7.1", "H7.2"} {
		sec, _ := p.SectionID(FarmHirbovat, code)
		f, m := aliveOn(sexSection{sec, dbq.SexF}, day("2026-11-18")), aliveOn(sexSection{sec, dbq.SexM}, day("2026-11-18"))
		alive = append(alive, ac{code, f, m})
		totF += f
		totM += m
	}
	wantAlive := []ac{{"H5.1", 4263, 342}, {"H5.2", 4281, 323}, {"H6.1", 3542, 124}, {"H6.2", 475, 5}, {"H7.1", 0, 0}, {"H7.2", 0, 0}}
	if d := gocmp.Diff(wantAlive, alive); d != "" {
		t.Errorf("alive 18.11 (-want +got):\n%s", d)
	}
	if totF != 12561 || totM != 794 {
		t.Errorf("alive 18.11 ♀ %d · ♂ %d, want 12 561 · 794", totF, totM)
	}

	// Golden 10: cycle deaths ♀ 1 521 · ♂ 1 048 (mortality from 22.02.2026).
	deaths := map[dbq.Sex]int32{}
	for _, e := range p.BirdEvents {
		if e.FarmID == hirbovat && e.Type == dbq.BirdEventTypeMortality && !e.Date.Before(day("2026-02-22")) {
			deaths[e.Sex] -= e.Qty
		}
	}
	if deaths[dbq.SexF] != 1521 || deaths[dbq.SexM] != 1048 {
		t.Errorf("cycle deaths ♀ %d · ♂ %d, want 1 521 · 1 048", deaths[dbq.SexF], deaths[dbq.SexM])
	}
}

func TestPlanEggs(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	var rows2026, rows2025 int
	var sum2026 int64
	halls := map[string]bool{}
	for _, r := range p.EggProductions {
		code := n.code[r.SectionID]
		halls[code] = true
		total := int64(*r.Cat1 + *r.Cat2 + *r.Cat3 + *r.HatchingEggs)
		if r.Date.Before(day("2026-02-22")) {
			rows2025++
			continue
		}
		rows2026++
		if !r.Date.After(day("2026-11-18")) {
			sum2026 += total
		}
	}
	if rows2026 != 757 || sum2026 != 5073155 {
		t.Errorf("Seria 2026 eggs: %d rows Σ %d, want 757 rows Σ 5 073 155", rows2026, sum2026)
	}
	if rows2025 != 888 {
		t.Errorf("Seria 2025 egg rows = %d, want 888", rows2025)
	}
	if d := gocmp.Diff(map[string]bool{"H5": true, "H6": true, "H7": true}, halls); d != "" {
		t.Errorf("egg units (-want +got):\n%s", d)
	}
	for _, r := range p.EggProductions {
		if r.Date == day("2026-11-18") && n.code[r.SectionID] == "H5" {
			// Golden 16: H5 13 · 417 · 27 + 3 415 = 3 872.
			if *r.Cat1 != 13 || *r.Cat2 != 417 || *r.Cat3 != 27 || *r.HatchingEggs != 3415 {
				t.Errorf("H5 18.11 Cat = %d · %d · %d + %d", *r.Cat1, *r.Cat2, *r.Cat3, *r.HatchingEggs)
			}
		}
		if r.AvgEggWeightG.Valid && r.AvgEggWeightG.Decimal.IsZero() {
			t.Errorf("%s %s: a 0 g egg weight must be null", n.code[r.SectionID], r.Date)
		}
	}

	// Legacy mortality: per compartment-day and sex at H5.1…H7.2, emptying
	// from each hall's depopulation day (H5 15.01, H6 16.01, H7 20.01.2026).
	depop := map[string]timex.Day{"H5": day("2026-01-15"), "H6": day("2026-01-16"), "H7": day("2026-01-20")}
	emptying := 0
	for _, r := range p.LegacyMortality {
		code := n.code[r.SectionID]
		hall := code[:2]
		if want := !r.Date.Before(depop[hall]); r.Emptying != want {
			t.Errorf("legacy %s %s: emptying %v, want %v", code, r.Date, r.Emptying, want)
		}
		if r.Emptying {
			emptying++
		}
	}
	// 888 hall-days × 2 compartments × 2 sexes, minus the 3 negative ♂ cells.
	if len(p.LegacyMortality) != 888*4-3 {
		t.Errorf("legacy_mortality rows = %d, want %d", len(p.LegacyMortality), 888*4-3)
	}
	if emptying != 24*4 {
		t.Errorf("emptying rows = %d, want 96 (24 depopulation hall-days)", emptying)
	}
	for _, r := range p.LegacyMortality {
		k := n.code[r.SectionID] + " " + r.Date.String() + " " + string(r.Sex)
		if k == "H5.2 2025-09-10 M" || k == "H5.2 2025-11-09 M" || k == "H7.2 2025-11-09 M" {
			t.Errorf("negative register cell %s imported", k)
		}
	}
}

func TestPlanFeed(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	ftName := map[uuid.UUID]string{}
	for _, f := range p.FeedTypes {
		ftName[f.ID] = f.Name
	}
	type key struct{ Farm, Section, Date, Sex string }
	alloc := map[key]dbq.FeedAllocation{}
	for _, r := range p.FeedAllocations {
		alloc[key{n.farm[r.FarmID], n.code[r.SectionID], r.Date.String(), string(r.Sex)}] = r
	}
	// Q6: rows with g but no kg are not imported (14 at Hîrbovăț, 2 at Telița).
	skipped := []key{
		{"hirbovat", "H7.1", "2026-11-13", "F"}, {"hirbovat", "H7.2", "2026-11-13", "F"},
		{"hirbovat", "H5.1", "2026-02-22", "M"}, {"hirbovat", "H7.2", "2026-02-22", "M"},
		{"hirbovat", "H5.1", "2026-02-25", "F"}, {"hirbovat", "H7.2", "2026-02-28", "F"},
		{"telita", "H5-2", "2026-02-22", "M"}, {"telita", "H5-2", "2026-02-23", "M"},
	}
	for _, k := range skipped {
		if _, ok := alloc[k]; ok {
			t.Errorf("%+v has g but no kg: must not be imported", k)
		}
	}
	if len(p.FeedAllocations) != 718-2+3198-14 {
		t.Errorf("feed_allocation rows = %d, want %d", len(p.FeedAllocations), 718-2+3198-14)
	}
	// The emptying day 14.11 H7 has kg 0 without g: imported as 0 kg.
	if r, ok := alloc[key{"hirbovat", "H7.1", "2026-11-14", "F"}]; !ok || r.GPerBird.Valid || !r.KgRecorded.Valid || !r.KgRecorded.Decimal.IsZero() {
		t.Errorf("H7.1 ♀ 14.11: want a 0 kg row without g, got %+v (present %v)", r, ok)
	}
	// Telița's first days have kg without g.
	if r := alloc[key{"telita", "H1", "2025-10-07", "F"}]; r.GPerBird.Valid || r.KgRecorded.Decimal.String() != "330" {
		t.Errorf("Telița H1 07.10.2025: g %v kg %s, want no g and 330 kg", r.GPerBird, r.KgRecorded.Decimal)
	}
	for _, tc := range []struct {
		k    key
		want string
	}{
		{key{"telita", "H1", "2025-11-13", "F"}, "Starter"},
		{key{"telita", "H1", "2025-11-14", "F"}, "Grower"},
		{key{"telita", "H5-1", "2026-01-19", "M"}, "Grower"},
		{key{"telita", "H5-1", "2026-01-20", "M"}, "Pre-ouat"},
		{key{"hirbovat", "H5.1", "2026-11-18", "F"}, "Breeder 2"},
		{key{"hirbovat", "H5.1", "2026-11-18", "M"}, "Cocoși"},
	} {
		r, ok := alloc[tc.k]
		if !ok || ftName[r.FeedTypeID] != tc.want {
			t.Errorf("%+v: feed type %q, want %q", tc.k, ftName[r.FeedTypeID], tc.want)
		}
	}
	// Golden 26: feed 18.11 ♀ 2 303,2 kg · ♂ 154,7 kg.
	kg := map[dbq.Sex]decimal.Decimal{}
	for _, r := range p.FeedAllocations {
		if r.Date == day("2026-11-18") {
			kg[r.Sex] = kg[r.Sex].Add(r.KgRecorded.Decimal)
		}
	}
	if kg[dbq.SexF].String() != "2303.2" || kg[dbq.SexM].String() != "154.7" {
		t.Errorf("kg 18.11 ♀ %s · ♂ %s, want 2303.2 · 154.7", kg[dbq.SexF], kg[dbq.SexM])
	}

	// Deliveries (golden 29) and the 3 stock adjustments.
	hirbovat, _ := p.FarmID(FarmHirbovat)
	var count, unpricedCount int
	var lei decimal.Decimal
	var unpriced []string
	for _, r := range p.FeedDeliveries {
		if r.FarmID != hirbovat {
			continue
		}
		count++
		if !r.TotalLei.Valid {
			unpricedCount++
			unpriced = append(unpriced, r.Date.String()+" "+ftName[r.FeedTypeID]+" "+r.QtyKg.String())
			if r.PriceFromRegister {
				t.Errorf("unpriced delivery %s marked din registru", r.Date)
			}
			continue
		}
		lei = lei.Add(r.TotalLei.Decimal)
		if !r.PriceFromRegister || r.PricedAt != nil || r.PricedBy != nil {
			t.Errorf("priced delivery %s: want price_from_register, no priced_at/by", r.Date)
		}
	}
	wantUnpriced := []string{"2026-02-17 Breeder 2 2000", "2026-02-17 Cocoși 1000", "2026-03-14 Cocoși 6160", "2026-11-12 Breeder 2 9000"}
	if count != 69 || unpricedCount != 4 || lei.String() != "9562437" {
		t.Errorf("Hîrbovăț deliveries %d (%d unpriced) %s lei, want 69 (4) 9562437", count, unpricedCount, lei)
	}
	if d := gocmp.Diff(wantUnpriced, unpriced); d != "" {
		t.Errorf("unpriced deliveries (-want +got):\n%s", d)
	}
	var adjs []string
	for _, a := range p.FeedStockAdjustments {
		adjs = append(adjs, a.Date.String()+" "+ftName[a.FeedTypeID]+" "+a.QtyKg.String()+" "+a.Note)
	}
	wantAdjs := []string{
		"2026-02-21 Breeder 2 -300 diferență registru",
		"2026-02-26 Breeder 2 -3000 mutare ♀ → ♂",
		"2026-02-26 Cocoși 3000 mutare ♀ → ♂",
	}
	if d := gocmp.Diff(wantAdjs, adjs); d != "" {
		t.Errorf("stock adjustments (-want +got):\n%s", d)
	}

	// Golden 27: stock end 18.11 = delivered + adjusted − consumed.
	stock := map[string]decimal.Decimal{}
	for _, r := range p.FeedDeliveries {
		if r.FarmID == hirbovat {
			stock[ftName[r.FeedTypeID]] = stock[ftName[r.FeedTypeID]].Add(r.QtyKg)
		}
	}
	for _, a := range p.FeedStockAdjustments {
		stock[ftName[a.FeedTypeID]] = stock[ftName[a.FeedTypeID]].Add(a.QtyKg)
	}
	for _, r := range p.FeedAllocations {
		if r.FarmID == hirbovat && !r.Date.After(day("2026-11-18")) {
			stock[ftName[r.FeedTypeID]] = stock[ftName[r.FeedTypeID]].Sub(r.KgRecorded.Decimal)
		}
	}
	if stock["Breeder 2"].String() != "31664.9" || stock["Cocoși"].String() != "1781" {
		t.Errorf("stock 18.11 Breeder 2 %s · Cocoși %s, want 31664.9 · 1781", stock["Breeder 2"], stock["Cocoși"])
	}
}

// TestPlanRunningCounts recomputes, independently of the plan's own check,
// that no section × sex ever ends a day below 0.
func TestPlanRunningCounts(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	type k struct {
		sec uuid.UUID
		sex dbq.Sex
	}
	rows := slices.Clone(p.BirdEvents)
	slices.SortStableFunc(rows, func(a, b dbq.BirdEvent) int { return a.Date.Compare(b.Date) })
	run := map[k]int64{}
	for i, e := range rows {
		run[k{e.SectionID, e.Sex}] += int64(e.Qty)
		if i+1 < len(rows) && rows[i+1].Date == e.Date {
			continue // check at the end of the day
		}
		for kk, v := range run {
			if v < 0 {
				t.Fatalf("%s %s is %d at the end of %s", n.code[kk.sec], kk.sex, v, e.Date)
			}
		}
	}
}

// TestPlanEntries checks entry ownership (user answer Q3): transfers and
// sales have their own entries, every other row points to its farm's
// migration entry, and every row's entry is planned.
func TestPlanEntries(t *testing.T) {
	p := testPlan(t)
	n := namesOf(p)
	kinds := map[dbq.EntryKind]int{}
	for _, e := range p.Entries {
		kinds[e.Kind]++
		if e.Source != dbq.RecordSourceMigration || e.CreatedBy != nil || !e.CreatedAt.Equal(testNow) {
			t.Errorf("entry %s %s: want source migration, no author, created_at = now", e.Kind, e.Date)
		}
	}
	if d := gocmp.Diff(map[dbq.EntryKind]int{dbq.EntryKindMigration: 2, dbq.EntryKindTransfer: 16 + 5 + 1, dbq.EntryKindSale: 7}, kinds); d != "" {
		t.Errorf("entries per kind (-want +got):\n%s", d)
	}
	for _, farm := range []string{FarmTelita, FarmHirbovat} {
		id, ok := p.MigrationEntryID(farm)
		e := n.entries[id]
		if !ok || e.Kind != dbq.EntryKindMigration || n.farm[e.FarmID] != farm {
			t.Errorf("%s: migration entry missing or wrong", farm)
		}
	}
	tel, _ := p.MigrationEntryID(FarmTelita)
	hir, _ := p.MigrationEntryID(FarmHirbovat)
	if n.entries[tel].Date != day("2025-10-07") || n.entries[hir].Date != day("2025-04-04") {
		t.Errorf("migration entry dates %s / %s, want each farm's first row 2025-10-07 / 2025-04-04", n.entries[tel].Date, n.entries[hir].Date)
	}

	for _, e := range p.BirdEvents {
		en, ok := n.entries[e.EntryID]
		switch {
		case !ok:
			t.Fatalf("bird row %+v: entry not planned", n.move(e))
		case e.Type == dbq.BirdEventTypeTransfer:
			if en.Kind != dbq.EntryKindTransfer || (en.FarmID != e.FarmID && *en.PeerFarmID != e.FarmID) || en.Date != e.Date {
				t.Errorf("transfer row %+v not on its transfer entry", n.move(e))
			}
		case e.Type == dbq.BirdEventTypeSale:
			if en.Kind != dbq.EntryKindSale || en.FarmID != e.FarmID || en.Date != e.Date {
				t.Errorf("sale row %+v not on its sale entry", n.move(e))
			}
		default:
			if en.Kind != dbq.EntryKindMigration || en.FarmID != e.FarmID {
				t.Errorf("%s row %+v not on its farm's migration entry", e.Type, n.move(e))
			}
		}
	}
	check := func(table string, entry, farm uuid.UUID) {
		en, ok := n.entries[entry]
		if !ok || en.Kind != dbq.EntryKindMigration || en.FarmID != farm {
			t.Errorf("%s row not on its farm's migration entry", table)
		}
	}
	for _, r := range p.EggProductions {
		check("egg_production", r.EntryID, r.FarmID)
	}
	for _, r := range p.FeedAllocations {
		check("feed_allocation", r.EntryID, r.FarmID)
	}
	for _, r := range p.FeedDeliveries {
		check("feed_delivery", r.EntryID, r.FarmID)
	}
	for _, r := range p.FeedStockAdjustments {
		check("feed_stock_adjustment", r.EntryID, r.FarmID)
	}
	for _, r := range p.Weighings {
		check("weighing", r.EntryID, r.FarmID)
	}
	for _, r := range p.LegacyMortality {
		check("legacy_mortality", r.EntryID, r.FarmID)
	}
	if len(p.Weighings) != 98 {
		t.Errorf("weighings = %d, want 98", len(p.Weighings))
	}
	// Golden 35: placement ♀ 28 080 · ♂ 3 510.
	placed := map[dbq.Sex]int32{}
	for _, e := range p.BirdEvents {
		if e.Type == dbq.BirdEventTypePlacement {
			placed[e.Sex] += e.Qty
		}
	}
	if placed[dbq.SexF] != 28080 || placed[dbq.SexM] != 3510 {
		t.Errorf("placement ♀ %d · ♂ %d, want 28 080 · 3 510", placed[dbq.SexF], placed[dbq.SexM])
	}
}

func TestReadExtractsErrors(t *testing.T) {
	if _, err := ReadExtracts(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("a missing directory must fail")
	}
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadExtracts(file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("a file instead of a directory: err = %v", err)
	}
	// An empty directory names the first missing file.
	if _, err := ReadExtracts(t.TempDir()); err == nil || !strings.Contains(err.Error(), "farms-sections.json") {
		t.Errorf("empty directory: err = %v", err)
	}
	// A malformed file names itself.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "farms-sections.json"), []byte(`{"farms": [`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadExtracts(dir); err == nil || !strings.Contains(err.Error(), "farms-sections.json") {
		t.Errorf("malformed JSON: err = %v", err)
	}
}

func TestDecodeTable(t *testing.T) {
	type row struct {
		A int32  `json:"a"`
		B string `json:"b"`
	}
	got, err := decodeTable[row](table{Cols: []string{"a", "b"}, Rows: [][]json.RawMessage{{[]byte("1"), []byte(`"x"`)}}})
	if err != nil || len(got) != 1 || got[0] != (row{1, "x"}) {
		t.Errorf("decodeTable = %+v, %v", got, err)
	}
	if _, err := decodeTable[row](table{Cols: []string{"a", "b"}, Rows: [][]json.RawMessage{{[]byte("1")}}}); err == nil {
		t.Error("a short row must fail")
	}
	if _, err := decodeTable[row](table{Cols: []string{"a"}, Rows: [][]json.RawMessage{{[]byte(`"x"`)}}}); err == nil {
		t.Error("a wrong cell type must fail")
	}
}

// TestBuildPlanRejects breaks the extracts in memory: each break must stop
// the plan with an error naming it, never produce rows.
func TestBuildPlanRejects(t *testing.T) {
	for _, tc := range []struct {
		name   string
		break_ func(x *Extracts)
		want   string
	}{
		{"link out total", func(x *Extracts) { x.transfers.Links[0].QtyOut++ }, "transfers out sum"},
		{"arrival differs from link", func(x *Extracts) { x.transfers.Arrivals[0].Qty++ }, "arrival"},
		{"Telița correction column", func(x *Extracts) { x.telitaDaily.Rows[40].Correction += 5 }, "register correction"},
		{"sale missing from sales", func(x *Extracts) { x.hirbovat.sales = x.hirbovat.sales[1:] }, "sold in section_day"},
		{"Hîrbovăț below 0", func(x *Extracts) { x.hirbovat.sectionDay[0].Mort = 1000 }, "goes below 0"},
		{"unknown section", func(x *Extracts) { x.weighings.Rows[0].Section = "H9" }, `unknown section telita "H9"`},
		{"egg total", func(x *Extracts) {
			for i := range x.hirbovat.hallDay {
				if x.hirbovat.hallDay[i].Eggs != nil {
					v := *x.hirbovat.hallDay[i].Eggs + 1
					x.hirbovat.hallDay[i].Eggs = &v
					return
				}
			}
		}, "total is not Cat. 1–3 + Incubabile"},
		{"g out of range", func(x *Extracts) {
			x.hirbovat.sectionDay[100].G = decimal.NewNullDecimal(decimal.NewFromInt(301))
		}, "outside 1–300"},
		{"priced without lei", func(x *Extracts) { x.deliveries.Rows[0].TotalLei = decimal.NullDecimal{} }, "priced without"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			x := readExtracts(t)
			tc.break_(x)
			p, err := BuildPlan(x, testNow)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
			if p != nil {
				t.Error("a failed plan must be nil")
			}
		})
	}
}
