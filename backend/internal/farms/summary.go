package farms

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// BuildSummary builds GET /api/farms/{id}/summary (§7.6, §5.13, M2 D6) on db
// for the cycle in.CycleID (nil: the open one, else the latest) and the
// period in.Period (zero: the whole cycle). Everything is taken on the
// period's last day ("day": today for the open cycle by default), the stats
// that sum over time from the cycle start. The lamp "lipsa" is shown only
// when day is in.Today, from the missing.from_hour of in.Now. The period is
// resolved on periodtable.TableCycle, as the table resolves it: a past cycle
// without a flock count runs from its first register day (M2 Q7), and its
// alive and cumulative mortality figures are null (M2 D5).
//
// A farm without a cycle, or a period with no day in the cycle, gives a
// summary with no readout and empty lists. Alerts and the latest records are
// [] (M8, M6), the feed days left stat comes in M3 (M2 Q3). It does not check
// the caller's access (the service does, M2 D4). ErrFarmNotFound and
// ErrCycleNotFound name a missing farm or cycle; a period that doesn't apply
// is calc's error (calc.ErrPeriod, calc.ErrIntervalFuture).
func BuildSummary(ctx context.Context, db dbq.DBTX, in Input) (SummaryResponse, error) {
	res := SummaryResponse{Stats: []Stat{}, Tiles: []Tile{}, Alerts: []any{}, Latest: []any{}}
	farm, err := dbq.New(db).FarmsGetFarm(ctx, in.FarmID)
	if errors.Is(err, pgx.ErrNoRows) {
		return res, ErrFarmNotFound
	}
	if err != nil {
		return res, fmt.Errorf("farms: summary farm: %w", err)
	}
	cycles, err := calc.FarmCycles(ctx, db, in.FarmID)
	if err != nil {
		return res, err
	}
	cycle, err := pickCycle(cycles, in.CycleID)
	if err != nil || cycle == nil {
		return res, err
	}
	period := in.Period
	if period.Kind == "" {
		period.Kind = calc.PeriodCycle
	}
	// The period is resolved on the cycle as the table counts it, so a past
	// cycle without a flock count runs from its first register day (M2 Q7).
	tableCycle, legacy, err := periodtable.TableCycle(ctx, db, *cycle)
	if err != nil {
		return res, err
	}
	rp, err := calc.ResolvePeriod(tableCycle, period, in.Today)
	if err != nil {
		return res, err
	}
	if rp.Empty {
		return res, nil
	}

	laying := cycle.Type(farm.Type) == dbq.FarmTypeLaying
	s, err := readSummary(ctx, db, *cycle, rp, laying, in)
	if err != nil {
		return res, err
	}
	s.noFlockCount = legacy
	if laying {
		res.Readout = s.layReadout(periodtable.Series(rp, s.farmDays, s.guide))
	} else {
		res.Readout = s.growthReadout()
	}
	res.Stats = s.stats()
	res.Tiles = s.tiles()
	return res, nil
}

// summary is what the summary reads for one cycle and day.
type summary struct {
	cycle        calc.CycleInfo
	day          timex.Day
	sapt         int
	laying       bool
	noFlockCount bool // a past cycle without a flock count (Seria 2025, D52, D54)
	settings     calc.AppSettings
	guide        calc.GuideWeeks // nil for a cycle without a guide
	// farmDays are the farm's days of the readout series (the whole period
	// for a laying cycle, the last day for a growth one); farmDay is the last.
	farmDays []periodtable.ScopeDay
	farmDay  periodtable.ScopeDay
	hallDays map[uuid.UUID]periodtable.ScopeDay // by egg unit
	arrivals calc.PerSex[int]
	deaths   calc.PerSex[int]
	sold     calc.PerSex[int]
	weighed  calc.LatestWeighings
	sections []calc.SectionDay
	week     map[uuid.UUID]calc.PerSex[int] // deaths per section in SAPT(day) ∩ cycle
	empty    map[uuid.UUID]calc.SectionEmpty
	missing  bool // today, after missing.from_hour: the §5.11 gaps show
}

// readSummary reads what the summary shows on the period's last day. Each
// figure is read once: the guide weeks cover both the weighings (from the
// week before the cycle's first, §5.10) and the readout series, and the
// farm's last day is the last row of the series' day query.
func readSummary(ctx context.Context, db dbq.DBTX, c calc.CycleInfo, rp calc.ResolvedPeriod, laying bool, in Input) (summary, error) {
	day := rp.To
	s := summary{cycle: c, day: day, sapt: calc.SAPT(c.Age, day), laying: laying}
	var err error
	if s.settings, err = calc.ReadAppSettings(ctx, db); err != nil {
		return s, err
	}
	if c.Guide != nil {
		// Weighings compare with SAPT − 1 (§5.10): one week before the cycle's
		// first. The series' weeks (rp.SaptFrom ≥ the cycle's first) fall inside.
		from := max(1, calc.SAPT(c.Age, c.StartDate)-1)
		if s.guide, err = calc.ReadGuideWeeks(ctx, db, c.Guide.ID, from, max(from, s.sapt)); err != nil {
			return s, err
		}
	}
	from := day // growth: no readout series in M2, so the last day only
	if laying {
		from = rp.From
	}
	farmDays, err := periodtable.ReadDays(ctx, db, periodtable.DaysQuery{FarmID: c.FarmID, Scope: periodtable.ScopeFarm, From: from, To: day})
	if err != nil {
		return s, err
	}
	s.farmDays = farmDays.Rows
	if n := len(s.farmDays); n > 0 {
		s.farmDay = s.farmDays[n-1]
	}
	hallDays, err := periodtable.ReadDays(ctx, db, periodtable.DaysQuery{FarmID: c.FarmID, Scope: periodtable.ScopeHall, From: day, To: day})
	if err != nil {
		return s, err
	}
	s.hallDays = make(map[uuid.UUID]periodtable.ScopeDay, len(hallDays.Rows))
	for _, r := range hallDays.Rows {
		s.hallDays[r.ScopeID] = r
	}
	if s.arrivals, err = calc.FarmBase(ctx, db, c.FarmID, c.StartDate, day); err != nil {
		return s, err
	}
	removals, err := calc.FarmRemovals(ctx, db, c.FarmID, c.StartDate, day)
	if err != nil {
		return s, err
	}
	s.deaths, s.sold = removals.Deaths, removals.Sold
	if s.weighed, err = calc.ReadLatestWeighings(ctx, db, c.FarmID, c.StartDate, day); err != nil {
		return s, err
	}
	if s.sections, err = calc.SectionsOnDay(ctx, db, c.FarmID, day); err != nil {
		return s, err
	}
	weekFrom := calc.WeekFirstDay(c.Age, s.sapt)
	if weekFrom.Before(c.StartDate) {
		weekFrom = c.StartDate
	}
	sectionRemovals, err := calc.SectionsRemovals(ctx, db, c.FarmID, weekFrom, day)
	if err != nil {
		return s, err
	}
	s.week = make(map[uuid.UUID]calc.PerSex[int], len(sectionRemovals))
	for _, r := range sectionRemovals {
		s.week[r.SectionID] = r.Deaths()
	}
	empties, err := calc.EmptySince(ctx, db, c.FarmID, day)
	if err != nil {
		return s, err
	}
	s.empty = make(map[uuid.UUID]calc.SectionEmpty, len(empties))
	for _, e := range empties {
		s.empty[e.SectionID] = e
	}
	s.missing = day == in.Today && in.Now.In(timex.Chisinau).Hour() >= s.settings.MissingFromHour
	return s, nil
}

// guideWeight is the body weight of the guide for sex at the end of the
// week before the weighing's (§5.10).
func (s summary) guideWeight(at timex.Day, sex dbq.Sex) *decimal.Decimal {
	return s.guide.BodyWeightG(calc.SAPT(s.cycle.Age, at)-1, sex)
}

// guideFeed is the guide's g/pasăre per sex for week k.
func (s summary) guideFeed(k int) calc.PerSex[*decimal.Decimal] {
	return calc.PerSex[*decimal.Decimal]{F: s.guide.FeedG(k, dbq.SexF), M: s.guide.FeedG(k, dbq.SexM)}
}

// layReadout is the latest week of the series with a lay % (§5.13), with the
// whole series; nil when no week has one.
func (s summary) layReadout(series []periodtable.SeriesPoint) *Readout {
	points := make([]ReadoutPoint, 0, len(series))
	var last *periodtable.SeriesPoint
	for i, p := range series {
		points = append(points, ReadoutPoint{Sapt: p.Sapt, Value: numPtr(p.Lay), Guide: numPtr(p.Guide)})
		if p.Lay != nil {
			last = &series[i]
		}
	}
	if last == nil {
		return nil
	}
	r := &Readout{Kind: ReadoutKindLay, Value: num.Of(*last.Lay), Guide: numPtr(last.Guide), Series: points}
	diff := calc.DiffPts(last.Lay, last.Guide)
	if diff == nil {
		r.Label = messages.ReadoutLayNoGuide(last.Sapt, *last.Lay, s.cycle.Guide != nil)
		return r
	}
	r.Label = messages.ReadoutLay(last.Sapt, *last.Lay, *last.Guide, *diff)
	r.Status = status(calc.StatusOf(diff, s.settings.ToleranceLayPts))
	return r
}

// growthReadout is the farm's latest ♀ weighing vs the guide of SAPT − 1
// (§5.13, M2 Q6 c); nil without a ♀ weighing. Its series is [] in M2: the
// weekly weights are the Greutate tab's (M6).
func (s summary) growthReadout() *Readout {
	w := s.weighed.Farm.F
	if w == nil {
		return nil
	}
	sapt := calc.SAPT(s.cycle.Age, w.Date)
	guide := s.guideWeight(w.Date, dbq.SexF)
	r := &Readout{Kind: ReadoutKindWeight, Value: num.Of(w.AvgWeightG), Guide: numPtr(guide), Series: []ReadoutPoint{}}
	diff := calc.DeviationPct(&w.AvgWeightG, guide)
	if diff == nil {
		r.Label = messages.ReadoutGrowthNoGuide(sapt, w.AvgWeightG, s.cycle.Guide != nil)
		return r
	}
	r.Label = messages.ReadoutGrowth(sapt, w.AvgWeightG, *guide, *diff)
	r.Status = status(calc.StatusOf(diff, s.settings.ToleranceWeightPct))
	return r
}

func (s summary) stats() []Stat {
	out := []Stat{s.aliveStat(), s.cumMortalityStat(), s.weightStat(), s.feedTodayStat()}
	if s.laying {
		out = append(out, s.layTodayStat())
	}
	return out
}

func (s summary) aliveStat() AliveStat {
	st := AliveStat{ID: StatAlive, Arrivals: sexPair(s.arrivals), Sold: sexPair(s.sold)}
	if !s.noFlockCount {
		f, m := s.farmDay.AliveEnd.F, s.farmDay.AliveEnd.M
		st.F, st.M = &f, &m
	}
	return st
}

func (s summary) cumMortalityStat() CumMortalityStat {
	var pct calc.PerSex[*decimal.Decimal]
	if !s.noFlockCount { // no flock count, no base (§5.5)
		pct.F, pct.M = calc.CumMortalityPct(s.deaths.F, s.arrivals.F), calc.CumMortalityPct(s.deaths.M, s.arrivals.M)
	}
	st := CumMortalityStat{ID: StatCumMortality, F: numPtr(pct.F), M: numPtr(pct.M), Deaths: sexPair(s.deaths)}
	var guide *decimal.Decimal
	if c := s.guide[s.sapt].F; c != nil {
		guide = c.MortalityCumPct
	}
	st.GuideF = numPtr(guide)
	if guide == nil {
		st.Reason = ReasonNoGuide
	}
	st.Status = status(calc.CumMortalityStatus(pct.F, guide))
	if s.laying {
		st.Note = messages.CumMortalityGuideNote()
	}
	return st
}

func (s summary) weightStat() WeightStat {
	st := WeightStat{ID: StatWeight}
	f, m := s.weighed.Farm.F, s.weighed.Farm.M
	if f != nil {
		st.F = numPtr(&f.AvgWeightG)
	}
	if m != nil {
		st.M = numPtr(&m.AvgWeightG)
	}
	shown, sex := f, dbq.SexF
	if shown == nil {
		shown, sex = m, dbq.SexM
	}
	if shown == nil {
		st.Reason = ReasonNoWeighing
		return st
	}
	at, code := shown.Date, shown.Code
	st.At, st.Section = &at, &code
	guide := s.guideWeight(at, sex)
	st.Guide = numPtr(guide)
	if guide == nil {
		st.Reason = ReasonNoGuide
	}
	st.Status = status(calc.StatusOf(calc.DeviationPct(&shown.AvgWeightG, guide), s.settings.ToleranceWeightPct))
	return st
}

func (s summary) feedTodayStat() FeedTodayStat {
	var r calc.Rollup
	r.Add(s.farmDay.DayFigures)
	g := r.FeedGPerBird()
	guide := s.guideFeed(s.sapt)
	st := FeedTodayStat{
		ID: StatFeedToday, F: numPtr(g.F), M: numPtr(g.M), GuideF: numPtr(guide.F), GuideM: numPtr(guide.M),
		Note: s.emptyHallsNote(),
	}
	if guide.F == nil || guide.M == nil {
		st.Reason = ReasonNoGuide
	}
	st.Status = status(calc.Worst(
		calc.Compared(calc.StatusOf(calc.DeviationPct(g.F, guide.F), s.settings.ToleranceFeedPct)),
		calc.Compared(calc.StatusOf(calc.DeviationPct(g.M, guide.M), s.settings.ToleranceFeedPct)),
	))
	return st
}

func (s summary) layTodayStat() LayTodayStat {
	var r calc.Rollup
	r.Add(s.farmDay.DayFigures)
	st := LayTodayStat{ID: StatLayToday, Eggs: r.Eggs}
	var lay *decimal.Decimal
	if s.farmDay.EggRows > 0 {
		lay = r.LayPct()
	}
	guide := s.guide.LayPct(s.sapt)
	diff := calc.DiffPts(lay, guide)
	st.Value, st.Guide, st.DiffPts = numPtr(lay), numPtr(guide), numPtr(diff)
	if guide == nil {
		st.Reason = ReasonNoGuide
	}
	st.Status = status(calc.StatusOf(diff, s.settings.ToleranceLayPts))
	for _, h := range s.hallDays {
		if h.EggRows > 0 {
			st.Halls++
		}
	}
	return st
}

// emptyHalls are the halls (egg units) with no bird at the end of the day
// that were emptied during the cycle, in tree order, with their sections.
func (s summary) emptyHalls() []EmptyHallTile {
	var out []EmptyHallTile
	for _, h := range s.halls() {
		if e, ok := s.emptiedInCycle(h); ok {
			out = append(out, e)
		}
	}
	return out
}

func (s summary) emptyHallsNote() string {
	var notes []string
	for _, e := range s.emptyHalls() {
		notes = append(notes, messages.HallEmptySince(e.Hall, e.EmptySince))
	}
	return strings.Join(notes, ", ")
}

// hall is one egg unit with its bird-holding sections, in tree order.
type hall struct {
	id       uuid.UUID
	code     string
	sections []calc.SectionDay
}

func (s summary) halls() []hall {
	var out []hall
	for _, sec := range s.sections {
		if n := len(out); n > 0 && out[n-1].id == sec.HallID {
			out[n-1].sections = append(out[n-1].sections, sec)
			continue
		}
		out = append(out, hall{id: sec.HallID, code: sec.HallCode, sections: []calc.SectionDay{sec}})
	}
	return out
}

// emptiedInCycle is the empty-hall line of h when it has no bird at the end
// of the day and was emptied on or after the cycle start (golden 38: H7
// "goală din 14.11.2026"); false otherwise.
func (s summary) emptiedInCycle(h hall) (EmptyHallTile, bool) {
	e, ok := s.empty[h.id]
	if !ok || e.AliveEnd != 0 || e.EmptySince == nil || e.EmptySince.Before(s.cycle.StartDate) {
		return EmptyHallTile{}, false
	}
	codes := make([]string, 0, len(h.sections))
	for _, sec := range h.sections {
		codes = append(codes, sec.Code)
	}
	return EmptyHallTile{Hall: h.code, EmptySince: *e.EmptySince, Label: messages.EmptyHallTile(*e.EmptySince, codes)}, true
}

// tiles are the section tiles (D41, M2 D6): per hall in tree order its
// sections that hold birds at the start or end of the day, or one empty-hall
// line. A hall without birds that wasn't emptied in this cycle is left out.
func (s summary) tiles() []Tile {
	out := []Tile{}
	feedGuide := s.guideFeed(s.sapt)
	layGuide := s.guide.LayPct(s.sapt)
	for _, h := range s.halls() {
		if e, ok := s.emptiedInCycle(h); ok {
			out = append(out, e)
			continue
		}
		var layHall, layDiff *decimal.Decimal
		if d, ok := s.hallDays[h.id]; ok && d.EggRows > 0 {
			var r calc.Rollup
			r.Add(d.DayFigures)
			layHall = r.LayPct()
			layDiff = calc.DiffPts(layHall, layGuide)
		}
		layBelow := layDiff != nil && layDiff.LessThan(s.settings.ToleranceLayPts.Neg())
		var growth []TileSection
		var laying []LayingTileSection
		for _, sec := range h.sections {
			if sec.AliveStart.F+sec.AliveStart.M+sec.AliveEnd.F+sec.AliveEnd.M == 0 {
				continue
			}
			t := s.tileSection(sec, feedGuide, layBelow)
			if s.laying {
				laying = append(laying, LayingTileSection{TileSection: t, LayHall: numPtr(layHall)})
			} else {
				growth = append(growth, t)
			}
		}
		switch {
		case len(laying) > 0:
			out = append(out, HallTile[LayingTileSection]{Hall: h.code, Sections: laying})
		case len(growth) > 0:
			out = append(out, HallTile[TileSection]{Hall: h.code, Sections: growth})
		}
	}
	return out
}

func (s summary) tileSection(sec calc.SectionDay, feedGuide calc.PerSex[*decimal.Decimal], layBelow bool) TileSection {
	t := TileSection{
		Code:       sec.Code,
		Alive:      sexPair(sec.AliveEnd),
		DeathsWeek: sexPair(s.week[sec.SectionID]),
		GToday:     httpx.SexPair[*num.Num]{F: numPtr(sec.G.F), M: numPtr(sec.G.M)},
		Lamps:      []string{},
	}
	if w := s.weighed.Sections[sec.SectionID].F; w != nil {
		guide := s.guideWeight(w.Date, dbq.SexF)
		t.Weight = &TileWeight{F: num.Of(w.AvgWeightG), At: w.Date, Guide: numPtr(guide)}
		if guide == nil {
			t.Weight.Reason = ReasonNoGuide
		}
	}
	if s.missing && sectionMissing(sec) {
		t.Lamps = append(t.Lamps, LampMissing)
	}
	if layBelow || feedOff(sec.G.F, feedGuide.F, s.settings.ToleranceFeedPct) ||
		feedOff(sec.G.M, feedGuide.M, s.settings.ToleranceFeedPct) {
		t.Lamps = append(t.Lamps, LampAbatere)
	}
	return t
}

// sectionMissing reports a §5.11 gap of the section on the day: a sex with
// birds at the start and the end of the day (the emptying day is not
// missing) without a mortality row (Zilnic) or without an allocation (Hrană).
func sectionMissing(sec calc.SectionDay) bool {
	gap := func(start, end int, mortality, fed bool) bool { return start > 0 && end > 0 && (!mortality || !fed) }
	return gap(sec.AliveStart.F, sec.AliveEnd.F, sec.MortalityEntered.F, sec.Fed.F) ||
		gap(sec.AliveStart.M, sec.AliveEnd.M, sec.MortalityEntered.M, sec.Fed.M)
}

// feedOff reports a typed g more than tolerance % off the guide (§5.7, D19).
func feedOff(g, guide *decimal.Decimal, tolerance decimal.Decimal) bool {
	st, ok := calc.StatusOf(calc.DeviationPct(g, guide), tolerance)
	return ok && st == calc.Abatere
}

func status(st calc.Status, ok bool) *calc.Status {
	if !ok {
		return nil
	}
	return &st
}

func numPtr(d *decimal.Decimal) *num.Num {
	if d == nil {
		return nil
	}
	n := num.Of(*d)
	return &n
}

func sexPair(p calc.PerSex[int]) httpx.SexPair[int] { return httpx.SexPair[int]{F: p.F, M: p.M} }
