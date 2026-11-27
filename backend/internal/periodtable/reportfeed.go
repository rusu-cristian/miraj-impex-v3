package periodtable

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Consum table (S11, §7.4 view=report-feed, §7.12; M3 D11, Q2, Q8): per
// week the farm week row, its day rows and the section week rows, with kg ♀ ·
// ♂ and g/pasăre against the guide.

const viewReportFeed = "report-feed"

// buildReportFeed is Build for view=report-feed. The page read is the cycle
// read, the Consum day query (feedsql.go), the guide weeks when the cycle has
// a guide and the settings (the tolerance of the foot note and the hour from
// which today's gaps count).
//
// Figures (§5.7, D38, D165):
//   - kg = Σ kg of the live allocations; null when the row has none ("lipsă"
//     when one of its section × sex rows is missing);
//   - g/pasăre of a week, a section week and the totals = 1000 × Σ kg ÷ Σ
//     birds at the start over the fed rows (kg > 0); of a farm day, the
//     register's g ("zi = g din registru", dayRegisterG);
//   - the guide of a day is its SAPT's; of a week, a section week and the
//     totals, weighted by fed bird-days; status per sex from the deviation
//     against tolerance.feed_pct;
//   - missing: the section × sex × day rows with birds at the start and the
//     end of the day (§5.11 has_birds) and no live allocation, on every day
//     of the period; today's only from missing.from_hour.
func buildReportFeed(ctx context.Context, db dbq.DBTX, in Input) (TableResponse, error) {
	q := in.Query
	if q.FarmID == nil || q.Detail != "" || q.Types != nil || (q.Grain != "" && q.Grain != grainDay && q.Grain != grainWeek) {
		return TableResponse{}, ErrUnsupported
	}
	farmID := *q.FarmID
	resp := TableResponse{
		Scope:  Scope{View: viewReportFeed, FarmID: &farmID, Label: messages.ScopeFarmTotal()},
		Chips:  []string{},
		Groups: []Group{},
		Weeks:  []WeekRow{},
		Notes:  []string{},
		AsOf:   timex.At(in.Now),
	}
	period := calc.Period(q.Period)
	if period.Kind == "" {
		period.Kind = calc.PeriodCycle
	}

	cyc, ok, err := pickCycle(ctx, db, farmID, q.CycleID)
	if err != nil {
		return TableResponse{}, err
	}
	if !ok {
		resp.Period = PeriodInfo{Kind: period.Kind, Label: noCycleLabel(period), From: in.Today, To: in.Today}
		return empty(resp), nil
	}
	resp.Scope.CycleID, resp.Scope.SeriesID = &cyc.ID, &cyc.SeriesID
	cycle, _, err := TableCycle(ctx, db, cyc)
	if err != nil {
		return TableResponse{}, err
	}
	rp, err := calc.ResolvePeriod(cycle, period, in.Today)
	if err != nil {
		return TableResponse{}, err
	}
	resp.Period = PeriodInfo{
		Kind: rp.Kind, Label: rp.Label, SaptFrom: rp.SaptFrom, SaptTo: rp.SaptTo,
		From: rp.From, To: rp.To, WeeksTotal: rp.WeeksTotal(),
	}
	if rp.Empty {
		return empty(resp), nil
	}
	page, err := calc.Paginate(rp, calc.PageRequest{Before: q.Before, Count: q.Count, Sapt: q.Sapt, FromSapt: q.FromSapt})
	if err != nil {
		return TableResponse{}, err
	}
	feed, err := readFeedDays(ctx, db, feedDaysQuery{
		FarmID: farmID, From: page.Loaded.From, To: page.Loaded.To, PeriodFrom: rp.From, PeriodTo: rp.To,
	})
	if err != nil {
		return TableResponse{}, err
	}
	if !feed.PeriodHasFeed { // no ration in the period
		return empty(resp), nil
	}
	g, err := readGuide(ctx, db, cyc, page.Loaded.SaptFrom, page.Loaded.SaptTo)
	if err != nil {
		return TableResponse{}, err
	}
	settings, err := calc.ReadAppSettings(ctx, db)
	if err != nil {
		return TableResponse{}, err
	}

	resp.Chips = []string{ChipFeed}
	resp.NoGuide = g.weeks == nil
	resp.Groups = consumGroups(resp.NoGuide)
	c := newConsum(feed.Rows, consumRules{
		keys:         leafSet(resp.Groups),
		guide:        g,
		tolFeed:      settings.ToleranceFeedPct,
		age:          rp.Age,
		today:        in.Today,
		missingToday: in.Now.In(timex.Chisinau).Hour() >= settings.MissingFromHour,
		cycleStart:   cyc.StartDate,
	})
	for _, w := range page.Weeks {
		resp.Weeks = append(resp.Weeks, c.weekRow(w, q.Grain != grainWeek))
	}
	resp.Loaded = &Loaded{SaptFrom: page.Loaded.SaptFrom, SaptTo: page.Loaded.SaptTo, Weeks: page.Loaded.Weeks}
	if page.More != nil {
		resp.More = &More{Before: page.More.Before, Remaining: page.More.Remaining, Label: page.More.Label}
	}
	if page.Start != "" {
		resp.Start = &StartNote{Label: page.Start}
	}
	var total feedAcc
	for _, r := range feed.Rows {
		c.add(&total, r)
	}
	cells, status, reasons := c.cells(total, 0, rowTotals)
	resp.Totals = &Totals{Label: page.TotalsLabel, Cells: cells, Status: status, Reasons: reasons, Missing: total.missingOut()}
	resp.Notes = append([]string{messages.NoteConsumG(settings.ToleranceFeedPct)}, c.recordedKgNotes()...)
	return resp, nil
}

// consumRules are what a Consum page lays its rows out with.
type consumRules struct {
	keys         map[string]bool // the leaf ids of the groups shown
	guide        guide
	tolFeed      decimal.Decimal // tolerance.feed_pct
	age          calc.SeriesAge
	today        timex.Day
	missingToday bool // today's gaps count (missing.from_hour has passed)
	cycleStart   timex.Day
}

// leafInfo is a leaf section with its hall.
type leafInfo struct {
	id, hall       uuid.UUID
	code, hallCode string
}

// consum lays out the rows of one page.
type consum struct {
	consumRules
	rows    []feedRow
	byDate  map[timex.Day][]feedRow
	leaves  []leafInfo                             // tree order
	emptied map[uuid.UUID]map[timex.Day]*timex.Day // hall → day → the day it is empty since (nil: has birds)
}

func newConsum(rows []feedRow, r consumRules) *consum {
	c := &consum{consumRules: r, rows: rows, byDate: map[timex.Day][]feedRow{}, emptied: map[uuid.UUID]map[timex.Day]*timex.Day{}}
	seen := map[uuid.UUID]bool{}
	var dates []timex.Day
	for _, row := range rows {
		if _, ok := c.byDate[row.Date]; !ok {
			dates = append(dates, row.Date)
		}
		c.byDate[row.Date] = append(c.byDate[row.Date], row)
		if !seen[row.SectionID] {
			seen[row.SectionID] = true
			c.leaves = append(c.leaves, leafInfo{id: row.SectionID, hall: row.HallID, code: row.Code, hallCode: row.HallCode})
		}
	}
	c.walkHalls(dates)
	return c
}

// walkHalls records, for each hall and day, the day the hall is empty since
// (§5.3 empty_since, both sexes together): it starts from the day before the
// page (hall_emptied_before when the hall has no bird at the start of the
// first day), a day that ends with none after starting with some empties it,
// a day that ends with birds fills it.
func (c *consum) walkHalls(dates []timex.Day) {
	if len(dates) == 0 {
		return
	}
	state := map[uuid.UUID]*timex.Day{}
	start := map[uuid.UUID]int{}
	for _, r := range c.byDate[dates[0]] {
		start[r.HallID] += r.BirdsStart
		state[r.HallID] = r.HallEmptiedBefore
	}
	for h, n := range start {
		if n > 0 {
			state[h] = nil
		}
	}
	for _, d := range dates {
		begin, end := map[uuid.UUID]int{}, map[uuid.UUID]int{}
		for _, r := range c.byDate[d] {
			begin[r.HallID] += r.BirdsStart
			end[r.HallID] += r.AliveEnd
		}
		for h := range begin {
			switch {
			case end[h] == 0 && begin[h] > 0:
				state[h] = &d
			case end[h] > 0:
				state[h] = nil
			}
			if c.emptied[h] == nil {
				c.emptied[h] = map[timex.Day]*timex.Day{}
			}
			c.emptied[h][d] = state[h]
		}
	}
}

// missing reports whether r is a missing ration (§5.11).
func (c *consum) missing(r feedRow) bool {
	return r.BirdsStart > 0 && r.AliveEnd > 0 && !r.Allocated && (r.Date != c.today || c.missingToday)
}

// feedAcc adds up the leaf rows of one Consum row.
type feedAcc struct {
	r    calc.Rollup      // FeedKg, FedBirdDays and the feed guide weighted by them
	rows calc.PerSex[int] // live allocations
	// registerG is the farm day's g/pasăre "din registru" (dayRegisterG): Σ kg
	// over Σ 1000 × kg ÷ g of the fed rows.
	registerG calc.PerSex[calc.Ratio]
	missing   []MissingKey
	shown     bool // a fed or missing row: a section week row is shown
}

// add adds the leaf row r to a.
func (c *consum) add(a *feedAcc, r feedRow) {
	d := calc.DayFigures{Date: r.Date}
	var kg decimal.Decimal
	fed := 0
	if r.Allocated {
		kg = r.Kg
		if r.fed() {
			fed = r.BirdsStart
		}
	}
	var guideG *decimal.Decimal
	if c.guide.weeks != nil {
		guideG = c.guide.weeks.FeedG(calc.SAPT(c.age, r.Date), r.Sex)
	}
	if r.Sex == dbq.SexF {
		d.FeedKg.F, d.FedBirdDays.F, d.GuideFeedG.F = kg, fed, guideG
	} else {
		d.FeedKg.M, d.FedBirdDays.M, d.GuideFeedG.M = kg, fed, guideG
	}
	a.r.Add(d)
	if r.fed() {
		ratio := &a.registerG.M
		if r.Sex == dbq.SexF {
			ratio = &a.registerG.F
		}
		ratio.Add(kg, registerBirds(r))
	}
	if r.Allocated {
		if r.Sex == dbq.SexF {
			a.rows.F++
		} else {
			a.rows.M++
		}
	}
	missing := c.missing(r)
	if missing {
		a.missing = append(a.missing, MissingKey{Code: r.Code, Sex: r.Sex, Date: r.Date})
	}
	if missing || r.fed() {
		a.shown = true
	}
}

// registerBirds are the birds a fed row's kg was rationed for, 1000 × kg ÷
// g: its birds at the start for an app row (kg = g × birds ÷ 1000), the
// register's own figure for a migrated one. A migrated row with kg and no g
// (early Telița days) counts its birds at the start.
//
// dayRegisterG: a farm day's g/pasăre is Σ kg ÷ Σ registerBirds × 1000 over
// its fed rows (M3, user answer to the T7 question; SPEC §5.7). It is the
// bird-weighted typed g, and equals 1000 × Σ kg ÷ Σ birds at the start
// (D38) except on days with a register kg far from g × birds (D163,
// 02–06.11.2026) or a register kg for fewer birds (H6.2 ♀ 15.11.2026), where
// it keeps the register's g as the S11 board shows it (03.11 161,2).
func registerBirds(r feedRow) decimal.Decimal {
	if !r.G.Valid || !r.KgRecorded.Valid {
		return decimal.NewFromInt(int64(r.BirdsStart))
	}
	return r.Kg.Mul(decimal.NewFromInt(1000)).Div(r.G.Decimal)
}

// missingOut is a's Missing, nil when it has none. Keys are in the order
// they were added: by day, then tree order, ♀ before ♂.
func (a feedAcc) missingOut() *Missing {
	if len(a.missing) == 0 {
		return nil
	}
	m := &Missing{Rows: len(a.missing), Days: []timex.Day{}, Keys: a.missing}
	for _, k := range a.missing {
		if n := len(m.Days); n == 0 || m.Days[n-1] != k.Date {
			m.Days = append(m.Days, k.Date)
		}
	}
	return m
}

// weekRow builds week w: the farm week row with its day rows when withDays,
// and its section week rows in tree order. A section is shown when it has a
// fed or a missing row in the week; a hall emptied in the cycle whose
// sections are not shown is one line "goală din 14.11.2026 · H7.1, H7.2".
func (c *consum) weekRow(w calc.WeekSpan, withDays bool) WeekRow {
	var week feedAcc
	sections := map[uuid.UUID]*feedAcc{}
	var days []DayRow
	for d := w.From; !d.After(w.To); d = d.AddDays(1) {
		var one feedAcc
		for _, r := range c.byDate[d] {
			c.add(&week, r)
			c.add(&one, r)
			if sections[r.SectionID] == nil {
				sections[r.SectionID] = &feedAcc{}
			}
			c.add(sections[r.SectionID], r)
		}
		if withDays {
			k := calc.SAPT(c.age, d)
			cells, status, reasons := c.cells(one, k, rowDay)
			days = append(days, DayRow{
				Date: d, AgeDays: calc.AgeDays(c.age, d), Sapt: k, Weekday: messages.Weekday(d.Weekday()),
				Cells: cells, Status: status, Reasons: reasons, Missing: one.missingOut(),
			})
		}
	}
	row := func(code, label string) WeekRow {
		return WeekRow{Code: code, Sapt: w.Sapt, From: w.From, To: w.To, DaysInRow: w.DaysInRow(), Partial: w.Partial(), Label: label}
	}
	var secRows []WeekRow
	for _, l := range c.halls() {
		hall := c.hallLeaves(l.hall)
		anyShown := slices.ContainsFunc(hall, func(h leafInfo) bool { return sections[h.id] != nil && sections[h.id].shown })
		if since := c.emptied[l.hall][w.To]; !anyShown && since != nil && !since.Before(c.cycleStart) {
			codes := make([]string, 0, len(hall))
			for _, h := range hall {
				codes = append(codes, h.code)
			}
			r := row(l.hallCode, messages.EmptyHallTile(*since, codes))
			r.Cells = Cells{}
			r.Reasons = map[string]string{ReasonAllCells: ReasonEmptySince + since.String()}
			secRows = append(secRows, r)
			continue
		}
		for _, h := range hall {
			a := sections[h.id]
			if a == nil || !a.shown {
				continue
			}
			r := row(h.code, h.code)
			r.Cells, r.Status, r.Reasons = c.cells(*a, w.Sapt, rowWeek)
			r.Missing = a.missingOut()
			secRows = append(secRows, r)
		}
	}
	out := row("", w.Label())
	out.Cells, out.Status, out.Reasons = c.cells(week, w.Sapt, rowWeek)
	out.Missing, out.Days, out.Sections = week.missingOut(), days, secRows
	return out
}

// hallLeaves are the leaves of hall in tree order.
func (c *consum) hallLeaves(hall uuid.UUID) []leafInfo {
	var out []leafInfo
	for _, l := range c.leaves {
		if l.hall == hall {
			out = append(out, l)
		}
	}
	return out
}

// cells writes the Consum cells of a row, k being its SAPT (unused for the
// totals). Status and reasons are nil when empty.
func (c *consum) cells(a feedAcc, k int, kind rowKind) (Cells, map[string]string, map[string]string) {
	w := &cellWriter{leaves: c.keys, c: Cells{}, status: map[string]string{}, reasons: map[string]string{}}
	for _, s := range []struct {
		sex                 dbq.Sex
		rows                int
		kg                  decimal.Decimal
		registerG           calc.Ratio
		kgKey, gKey, guided string
		diffKey             string
	}{
		{dbq.SexF, a.rows.F, a.r.FeedKg.F, a.registerG.F, KeyFeedKgF, KeyFeedGF, KeyFeedGGuideF, KeyFeedGDiffF},
		{dbq.SexM, a.rows.M, a.r.FeedKg.M, a.registerG.M, KeyFeedKgM, KeyFeedGM, KeyFeedGGuideM, KeyFeedGDiffM},
	} {
		missing := slices.ContainsFunc(a.missing, func(m MissingKey) bool { return m.Sex == s.sex })
		if s.rows > 0 {
			w.set(s.kgKey, numDecV(s.kg))
		} else {
			w.set(s.kgKey, nil)
			if missing {
				w.reason(s.kgKey, ReasonMissing)
			}
		}
		actual, guideG, gap := c.guide.feedGFigures(acc{r: a.r}, k, kind, s.sex)
		if kind == rowDay {
			actual = s.registerG.Times(decimal.NewFromInt(1000))
		}
		w.set(s.gKey, numDec(actual))
		if actual == nil && missing {
			w.reason(s.gKey, ReasonMissing)
		}
		if c.guide.weeks != nil {
			w.set(s.guided, numDec(guideG))
			if gap {
				w.reason(s.guided, ReasonNoGuide)
			}
			diff := calc.DeviationPct(actual, guideG)
			w.set(s.diffKey, numDec(diff))
			w.compare(s.gKey, diff, c.tolFeed)
		}
	}
	return w.c, nilIfEmpty(w.status), nilIfEmpty(w.reasons)
}

// recordedKgNotes are the D163 notes of the loaded days (M3 Q8): the migrated
// allocations whose register kg differs from g × birds at the start ÷ 1000 by
// more than tolerance.feed_pct, one sentence per sex and register kg (whole
// kg), ♀ first: "♀ 460 kg/compartiment în registru (g × păsări ≈ 690 kg):
// SAPT 57: H5, H6 03.11–04.11 · H7 03.11–06.11; SAPT 56: H6 02.11.". Weeks
// are newest first; in a week, each hall spans its first to last such day
// and halls with the same span go together, in tree order. g × păsări is the
// mean over the sentence's rows, to the nearest 10 kg. A day the section ×
// sex had birds moved (a sale, a transfer, …) is left out: the register
// feeds fewer birds that day (H6.2 ♀ 15.11.2026, 428 kg).
func (c *consum) recordedKgNotes() []string {
	type cell struct {
		hall     uuid.UUID
		hallCode string
		date     timex.Day
	}
	type group struct {
		sex      dbq.Sex
		kg       decimal.Decimal
		computed decimal.Decimal // Σ g × birds ÷ 1000
		cells    []cell
	}
	var groups []*group
	thousand, hundred := decimal.NewFromInt(1000), decimal.NewFromInt(100)
	for _, r := range c.rows {
		if !r.Allocated || !r.KgRecorded.Valid || !r.G.Valid || r.BirdsStart <= 0 || r.Moved {
			continue
		}
		computed := r.G.Decimal.Mul(decimal.NewFromInt(int64(r.BirdsStart))).Div(thousand)
		dev := r.KgRecorded.Decimal.Div(computed).Sub(decimal.NewFromInt(1)).Abs().Mul(hundred)
		if !dev.GreaterThan(c.tolFeed) {
			continue
		}
		kg := r.KgRecorded.Decimal.Round(0)
		i := slices.IndexFunc(groups, func(g *group) bool { return g.sex == r.Sex && g.kg.Equal(kg) })
		if i < 0 {
			groups = append(groups, &group{sex: r.Sex, kg: kg})
			i = len(groups) - 1
		}
		g := groups[i]
		g.computed = g.computed.Add(computed)
		g.cells = append(g.cells, cell{hall: r.HallID, hallCode: r.HallCode, date: r.Date})
	}
	slices.SortStableFunc(groups, func(a, b *group) int {
		if a.sex != b.sex {
			if a.sex == dbq.SexF {
				return -1
			}
			return 1
		}
		return a.kg.Cmp(b.kg)
	})

	notes := []string{}
	for _, g := range groups {
		bySapt := map[int][]cell{}
		var sapts []int
		for _, x := range g.cells {
			k := calc.SAPT(c.age, x.date)
			if bySapt[k] == nil {
				sapts = append(sapts, k)
			}
			bySapt[k] = append(bySapt[k], x)
		}
		slices.SortFunc(sapts, func(a, b int) int { return b - a })
		weeks := make([]messages.RecordedKgWeek, 0, len(sapts))
		for _, k := range sapts {
			var spans []messages.RecordedKgSpan
			for _, h := range c.halls() {
				var from, to timex.Day
				for _, x := range bySapt[k] {
					if x.hall != h.hall {
						continue
					}
					if from.IsZero() || x.date.Before(from) {
						from = x.date
					}
					if to.IsZero() || x.date.After(to) {
						to = x.date
					}
				}
				if from.IsZero() {
					continue
				}
				if j := slices.IndexFunc(spans, func(s messages.RecordedKgSpan) bool { return s.From == from && s.To == to }); j >= 0 {
					spans[j].Halls = append(spans[j].Halls, h.hallCode)
				} else {
					spans = append(spans, messages.RecordedKgSpan{Halls: []string{h.hallCode}, From: from, To: to})
				}
			}
			weeks = append(weeks, messages.RecordedKgWeek{Sapt: k, Spans: spans})
		}
		ten := decimal.NewFromInt(10)
		mean := g.computed.Div(decimal.NewFromInt(int64(len(g.cells))))
		notes = append(notes, messages.NoteRecordedKg(g.sex, g.kg, mean.Div(ten).Round(0).Mul(ten), weeks))
	}
	return notes
}

// halls are the halls of the page in tree order, one leafInfo each (its
// first leaf).
func (c *consum) halls() []leafInfo {
	var out []leafInfo
	for i, l := range c.leaves {
		if i == 0 || c.leaves[i-1].hall != l.hall {
			out = append(out, l)
		}
	}
	return out
}
