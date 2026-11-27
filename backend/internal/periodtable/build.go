package periodtable

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Input is what Build needs besides the database: the parsed query, the
// business day and the instant of the response (asOf).
type Input struct {
	Query TableQuery
	Today timex.Day
	Now   time.Time
}

var (
	// ErrUnsupported is returned for a view, tab, detail or grain that M2
	// does not serve (M2 D2); the handler refuses those first.
	ErrUnsupported = errors.New("periodtable: view, tab, detail or grain not served")
)

const (
	viewFarm   = "farm"
	tabRezumat = "rezumat"
	detailFarm = "farm"
	grainDay   = "day"
	grainWeek  = "week"
)

// variant is the shape of a cycle's Rezumat (§7.4).
type variant int

const (
	variantLaying variant = iota
	variantGrowth         // g/pasăre and Greutate ♀ instead of ouat (growth.go)
	variantLegacy         // a past cycle without a flock count (legacy.go)
)

// Build returns the PeriodTable of in.Query (§7.4) as of in.Today. The page
// read is the cycle read, the day query (ReadDays) and, when the cycle has a
// guide, the guide weeks and the thresholds; a past cycle without a flock
// count adds the read of its first register day. A cycleId of another farm is
// calc.ErrNoCycle; a period or paging parameter out of range is the calc
// error (calc.ErrPeriod, calc.ErrIntervalFuture, *calc.PageError).
//
// view=report-feed is the Consum table (buildReportFeed, reportfeed.go).
func Build(ctx context.Context, db dbq.DBTX, in Input) (TableResponse, error) {
	q := in.Query
	if q.View == viewReportFeed {
		return buildReportFeed(ctx, db, in)
	}
	if q.View != viewFarm || q.FarmID == nil || (q.Tab != "" && q.Tab != tabRezumat) ||
		(q.Detail != "" && q.Detail != detailFarm) || (q.Grain != "" && q.Grain != grainDay && q.Grain != grainWeek) {
		return TableResponse{}, ErrUnsupported
	}
	farmID := *q.FarmID
	resp := TableResponse{
		Scope:  Scope{View: viewFarm, FarmID: &farmID, Label: messages.ScopeFarmTotal()},
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
	if !ok { // a farm that never had a cycle
		resp.Period = PeriodInfo{Kind: period.Kind, Label: noCycleLabel(period), From: in.Today, To: in.Today}
		return empty(resp), nil
	}
	resp.Scope.CycleID, resp.Scope.SeriesID = &cyc.ID, &cyc.SeriesID
	kind := variantLaying
	cycle, legacy, err := TableCycle(ctx, db, cyc)
	if err != nil {
		return TableResponse{}, err
	}
	if legacy {
		kind = variantLegacy
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
	days, err := ReadDays(ctx, db, DaysQuery{
		FarmID: farmID, Scope: ScopeFarm, From: page.Loaded.From, To: page.Loaded.To,
		PeriodFrom: rp.From, PeriodTo: rp.To, WithChips: true,
	})
	if err != nil {
		return TableResponse{}, err
	}
	if len(days.Chips) == 0 { // no record in the period
		return empty(resp), nil
	}
	if kind == variantLaying && cyc.Type(dbq.FarmType(days.FarmType)) == dbq.FarmTypeGrowth {
		kind = variantGrowth
	}

	lay := layout{variant: kind}
	if kind != variantLegacy {
		// Growth compares weights with the guide of SAPT − 1 (§5.10).
		if lay.guide, err = readGuide(ctx, db, cyc, page.Loaded.SaptFrom-1, page.Loaded.SaptTo); err != nil {
			return TableResponse{}, err
		}
	}
	if lay.guide.weeks != nil {
		settings, err := calc.ReadAppSettings(ctx, db)
		if err != nil {
			return TableResponse{}, err
		}
		lay.tolLay, lay.tolFeed, lay.tolWeight = settings.ToleranceLayPts, settings.ToleranceFeedPct, settings.ToleranceWeightPct
	}
	g := lay.guide

	resp.Chips = days.Chips
	switch kind {
	case variantLegacy:
		resp.NoGuide = cyc.Guide == nil
		resp.Groups = legacyRezumatGroups(days.Chips, q.Types)
	case variantGrowth:
		resp.NoGuide = g.weeks == nil
		resp.Groups = growthRezumatGroups(days.Chips, q.Types, resp.NoGuide)
		resp.Notes = []string{messages.NoteDeathsIncludeCulls()}
	default:
		resp.NoGuide = g.weeks == nil
		resp.Groups = layingRezumatGroups(days.Chips, q.Types, resp.NoGuide)
		resp.Notes = []string{messages.NoteDeathsIncludeCulls()}
	}
	lay.leaves = leafSet(resp.Groups)

	var total acc
	for _, d := range days.Rows {
		total.add(g.fill(rp.Age, d))
	}
	for _, w := range page.Weeks {
		resp.Weeks = append(resp.Weeks, lay.weekRow(rp.Age, w, days.Rows, q.Grain != grainWeek))
	}
	resp.Loaded = &Loaded{SaptFrom: page.Loaded.SaptFrom, SaptTo: page.Loaded.SaptTo, Weeks: page.Loaded.Weeks}
	if page.More != nil {
		resp.More = &More{Before: page.More.Before, Remaining: page.More.Remaining, Label: page.More.Label}
	}
	if page.Start != "" {
		resp.Start = &StartNote{Label: page.Start}
	}
	cells, status, reasons := lay.cells(total, nil, rowTotals)
	resp.Totals = &Totals{Label: page.TotalsLabel, Cells: cells, Status: status, Reasons: reasons}
	return resp, nil
}

// empty turns resp into the no-data state (§7.4 States).
func empty(resp TableResponse) TableResponse {
	resp.Empty = &EmptyState{Next: messages.EmptyNext()}
	return resp
}

// noCycleLabel is the period label when the farm has no cycle to resolve it on.
func noCycleLabel(p calc.Period) string {
	switch p.Kind {
	case calc.PeriodWeeks:
		return messages.PeriodWeeks(p.Weeks)
	case calc.PeriodInterval:
		if !p.From.IsZero() && !p.To.IsZero() {
			return messages.PeriodInterval(p.From, p.To)
		}
	}
	return messages.PeriodCurrentCycle()
}

// pickCycle returns the asked cycle of the farm (calc.ErrNoCycle when it is
// another farm's), or the open cycle, else the latest; ok is false when the
// farm has no cycle.
func pickCycle(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, cycleID *uuid.UUID) (calc.CycleInfo, bool, error) {
	if cycleID != nil {
		c, err := calc.CycleByID(ctx, db, farmID, *cycleID)
		if err != nil {
			return calc.CycleInfo{}, false, err
		}
		return c, true, nil
	}
	cycles, err := calc.FarmCycles(ctx, db, farmID)
	if err != nil {
		return calc.CycleInfo{}, false, err
	}
	c, ok := calc.DefaultCycle(cycles)
	return c, ok, nil
}

// guide is the cycle's guide over the weeks read; weeks is nil for a cycle
// without a guide.
type guide struct {
	weeks calc.GuideWeeks
}

func readGuide(ctx context.Context, db dbq.DBTX, c calc.CycleInfo, from, to int) (guide, error) {
	if c.Guide == nil {
		return guide{}, nil
	}
	w, err := calc.ReadGuideWeeks(ctx, db, c.Guide.ID, from, to)
	if err != nil {
		return guide{}, err
	}
	if w == nil {
		w = calc.GuideWeeks{}
	}
	return guide{weeks: w}, nil
}

// fill sets the guide fields of a day from the guide of its SAPT.
func (g guide) fill(age calc.SeriesAge, d ScopeDay) ScopeDay {
	k := calc.SAPT(age, d.Date)
	d.GuideLay = g.weeks.LayPct(k)
	d.GuideFeedG.F = g.weeks.FeedG(k, dbq.SexF)
	d.GuideFeedG.M = g.weeks.FeedG(k, dbq.SexM)
	return d
}

// acc adds up the days of one row: the calc rollup plus what decides a null
// cell.
type acc struct {
	r             calc.Rollup
	mortalityRows int
	eggRows       int
	feedRows      int
	birdDays      int                     // days with birds at the start
	weight        calc.PerSex[calc.Ratio] // Σ weight × birds ÷ Σ birds over the weighings (§5.10, A8)
	legacyDeaths  calc.PerSex[int]
	legacyRows    calc.PerSex[int]
	emptyingRows  calc.PerSex[int]
}

func (a *acc) add(d ScopeDay) {
	a.r.Add(d.DayFigures)
	a.mortalityRows += d.MortalityRows
	a.eggRows += d.EggRows
	a.feedRows += d.FeedRows
	if d.AliveStart.F+d.AliveStart.M > 0 {
		a.birdDays++
	}
	a.weight.F.Add(d.WeightGrams.F, decimal.NewFromInt(int64(d.WeighedBirds.F)))
	a.weight.M.Add(d.WeightGrams.M, decimal.NewFromInt(int64(d.WeighedBirds.M)))
	a.legacyDeaths = sumSex(a.legacyDeaths, d.LegacyDeaths)
	a.legacyRows = sumSex(a.legacyRows, d.LegacyRows)
	a.emptyingRows = sumSex(a.emptyingRows, d.EmptyingRows)
}

func sumSex(a, b calc.PerSex[int]) calc.PerSex[int] {
	return calc.PerSex[int]{F: a.F + b.F, M: a.M + b.M}
}

// rowKind tells how a row takes its guide (§5.14): a day the guide of its
// SAPT, a week weighted by hen-days, the totals weighted over every loaded day.
type rowKind int

const (
	rowDay rowKind = iota
	rowWeek
	rowTotals
)

// layFigures is a row's lay % and lay guide, and whether the guide is
// missing ("– fără date în ghid"). sapt is the row's week (unused for totals).
//   - lay % is nil for a row without an egg entry (weeks without eggs stay
//     "–", M2 Q1), else Σ eggs ÷ Σ hen-days;
//   - the guide of a day is its SAPT's; of a week, weighted by hen-days, or
//     its SAPT's when it has no hen-days; of the totals, weighted by hen-days.
func (g guide) layFigures(a acc, sapt int, kind rowKind) (lay, guidePct *decimal.Decimal, gap bool) {
	if a.eggRows > 0 {
		lay = a.r.LayPct()
	}
	if g.weeks == nil {
		return lay, nil, false
	}
	switch {
	case kind == rowDay || (kind == rowWeek && a.r.HenDays == 0):
		guidePct = g.weeks.LayPct(sapt)
		return lay, guidePct, guidePct == nil
	default:
		guidePct = a.r.GuideLay()
		return lay, guidePct, guidePct == nil && a.r.HenDays > 0
	}
}

// layout writes rows: the leaves shown, the guide and the lay tolerance.
type layout struct {
	variant   variant
	leaves    map[string]bool
	guide     guide
	tolLay    decimal.Decimal // D19 tolerances from app_setting
	tolFeed   decimal.Decimal
	tolWeight decimal.Decimal
}

// weekRow builds week w from the days of the loaded span, with its day rows
// when withDays.
func (l layout) weekRow(age calc.SeriesAge, w calc.WeekSpan, days []ScopeDay, withDays bool) WeekRow {
	var week acc
	var dayRows []DayRow
	for _, d := range days {
		if !w.Contains(d.Date) {
			continue
		}
		d = l.guide.fill(age, d)
		week.add(d)
		if withDays {
			var one acc
			one.add(d)
			k := calc.SAPT(age, d.Date)
			cells, status, reasons := l.cells(one, &k, rowDay)
			dayRows = append(dayRows, DayRow{
				Date: d.Date, AgeDays: calc.AgeDays(age, d.Date), Sapt: k, Weekday: messages.Weekday(d.Date.Weekday()),
				Cells: cells, Status: status, Reasons: reasons,
			})
		}
	}
	cells, status, reasons := l.cells(week, &w.Sapt, rowWeek)
	return WeekRow{
		Sapt: w.Sapt, From: w.From, To: w.To, DaysInRow: w.DaysInRow(), Partial: w.Partial(), Label: w.Label(),
		Cells: cells, Status: status, Reasons: reasons, Days: dayRows,
	}
}

// cellWriter collects a row's cells, status and reasons, keeping only the
// leaves of the groups shown.
type cellWriter struct {
	leaves  map[string]bool
	c       Cells
	status  map[string]string
	reasons map[string]string
}

func (w *cellWriter) set(key string, v *num.Num) {
	if w.leaves[key] {
		w.c[key] = v
	}
}

func (w *cellWriter) reason(key, r string) {
	if w.leaves[key] {
		w.reasons[key] = r
	}
}

// compare writes the status of key from its difference with the guide.
func (w *cellWriter) compare(key string, diff *decimal.Decimal, tolerance decimal.Decimal) {
	if st, ok := calc.StatusOf(diff, tolerance); ok && w.leaves[key] {
		w.status[key] = string(st)
	}
}

// cells writes the Rezumat cells of a row (§5.14): counts are sums, alive is
// the level at the row's last day, rates are Σ ÷ Σ. A count with no entry in
// the row is null, "missing" when the row had birds. Status and reasons are
// nil when empty, so they are left out of the JSON.
func (l layout) cells(a acc, sapt *int, kind rowKind) (Cells, map[string]string, map[string]string) {
	w := &cellWriter{leaves: l.leaves, c: Cells{}, status: map[string]string{}, reasons: map[string]string{}}
	missing := func(key string) {
		if a.birdDays > 0 {
			w.reason(key, ReasonMissing)
		}
	}
	if alive, ok := a.r.AliveEnd(); ok {
		w.set(KeyAliveF, numInt(alive.F))
		w.set(KeyAliveM, numInt(alive.M))
	}
	switch {
	case l.variant == variantLegacy:
		legacyDeathCells(w, a)
	case a.mortalityRows > 0:
		deaths := a.r.Deaths()
		w.set(KeyDeathsF, numInt(deaths.F))
		w.set(KeyDeathsM, numInt(deaths.M))
	default:
		w.set(KeyDeathsF, nil)
		w.set(KeyDeathsM, nil)
		missing(KeyDeathsF)
		missing(KeyDeathsM)
	}

	k := 0
	if sapt != nil {
		k = *sapt
	}
	lay, guidePct, gap := l.guide.layFigures(a, k, kind)
	if a.eggRows > 0 {
		w.set(KeyEggs, numInt(a.r.Eggs))
		w.set(KeyInc, numInt(a.r.Hatching))
		w.set(KeyJos, numInt(a.r.Floor))
		w.set(KeyIncPct, numDec(a.r.HatchingPct()))
		w.set(KeyJosPct, numDec(a.r.FloorPct()))
	} else {
		for _, key := range []string{KeyEggs, KeyInc, KeyJos, KeyIncPct, KeyJosPct} {
			w.set(key, nil)
		}
	}
	w.set(KeyLayPct, numDec(lay))
	if l.guide.weeks != nil {
		w.set(KeyLayGuide, numDec(guidePct))
		if gap {
			w.reason(KeyLayGuide, ReasonNoGuide)
		}
		w.compare(KeyLayPct, calc.DiffPts(lay, guidePct), l.tolLay)
	}

	if a.feedRows > 0 {
		w.set(KeyFeedKgF, numDecV(a.r.FeedKg.F))
		w.set(KeyFeedKgM, numDecV(a.r.FeedKg.M))
	} else {
		w.set(KeyFeedKgF, nil)
		w.set(KeyFeedKgM, nil)
		missing(KeyFeedKgF)
		missing(KeyFeedKgM)
	}
	if l.variant == variantGrowth {
		l.growthCells(w, a, k, kind)
	}
	w.set(KeySoldF, numInt(a.r.Sold.F))
	w.set(KeySoldM, numInt(a.r.Sold.M))
	return w.c, nilIfEmpty(w.status), nilIfEmpty(w.reasons)
}

func nilIfEmpty(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	return m
}

func numInt(i int) *num.Num {
	n := num.FromInt(int64(i))
	return &n
}

func numDec(d *decimal.Decimal) *num.Num {
	if d == nil {
		return nil
	}
	return numDecV(*d)
}

func numDecV(d decimal.Decimal) *num.Num {
	n := num.Of(d)
	return &n
}

// SeriesPoint is one week of WeekSeries.
type SeriesPoint struct {
	Sapt  int
	Lay   *decimal.Decimal // lay %; nil for a week without an egg entry
	Guide *decimal.Decimal // the lay guide weighted by hen-days; nil without a guide value
	// Rollup is the week's sums (eggs, hen-days, …) from the same rollup as
	// the table's week rows.
	Rollup calc.Rollup
}

// WeekSeries returns per SAPT of the period, oldest first, the lay % and the
// lay guide of the farm (§7.6 readout series, golden 23), over the whole
// period, from the same day query and rollups as Build's week rows. The
// period is resolved on TableCycle, as Build resolves it.
func WeekSeries(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, cycle calc.CycleInfo, period calc.Period, today timex.Day) ([]SeriesPoint, error) {
	if period.Kind == "" {
		period.Kind = calc.PeriodCycle
	}
	tc, _, err := TableCycle(ctx, db, cycle)
	if err != nil {
		return nil, err
	}
	rp, err := calc.ResolvePeriod(tc, period, today)
	if err != nil {
		return nil, err
	}
	if rp.Empty {
		return []SeriesPoint{}, nil
	}
	days, err := ReadDays(ctx, db, DaysQuery{FarmID: farmID, Scope: ScopeFarm, From: rp.From, To: rp.To})
	if err != nil {
		return nil, err
	}
	g, err := readGuide(ctx, db, cycle, rp.SaptFrom, rp.SaptTo)
	if err != nil {
		return nil, err
	}
	return Series(rp, days.Rows, g.weeks), nil
}

// Series is WeekSeries on rows already read: rows are the farm's days of rp
// (ReadDays with ScopeFarm), weeks the cycle's guide over at least
// rp.SaptFrom–rp.SaptTo (nil for a cycle without a guide). An Empty rp gives
// no point.
func Series(rp calc.ResolvedPeriod, rows []ScopeDay, weeks calc.GuideWeeks) []SeriesPoint {
	out := []SeriesPoint{}
	if rp.Empty {
		return out
	}
	g := guide{weeks: weeks}
	byWeek := map[int]*acc{}
	for _, d := range rows {
		k := calc.SAPT(rp.Age, d.Date)
		if byWeek[k] == nil {
			byWeek[k] = &acc{}
		}
		byWeek[k].add(g.fill(rp.Age, d))
	}
	for k := rp.SaptFrom; k <= rp.SaptTo; k++ {
		a := byWeek[k]
		if a == nil {
			a = &acc{}
		}
		lay, guidePct, _ := g.layFigures(*a, k, rowWeek)
		out = append(out, SeriesPoint{Sapt: k, Lay: lay, Guide: guidePct, Rollup: a.r})
	}
	return out
}
