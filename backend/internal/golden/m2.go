package golden

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// M2 rows (13–25, 38 and the per-hall part of 14) read the PeriodTable
// (periodtable.Build, its day query and WeekSeries) and the farm summary
// (farms.BuildSummary) as of AsOf and Now, never app_today().

// Now is the instant the golden values are stated at (§11.7): 18.11.2026
// 12:00 in Europe/Chisinau, the frozen clock of the tests.
var Now = time.Date(2026, time.November, 18, 12, 0, 0, 0, timex.Chisinau)

// GuideLay is golden 13: the ♀ lay guide of SAPT 59, 53 and 56 (1 dp), and the
// farm's lay on AsOf against the guide of SAPT 59 in points (1 dp) with its
// status.
type GuideLay struct {
	Guide59     string
	DiffToday   string
	StatusToday calc.Status
	Guide53     string
	Guide56     string
}

// HallLay is a hall's lay % on AsOf and its difference from the farm's lay
// guide of the day, in points (1 dp each).
type HallLay struct {
	Hall    string
	Pct     string
	DiffPts string
}

// EggsDay is the farm's day row on AsOf: eggs, incubabile and jos with their
// shares of the eggs (1 dp).
type EggsDay struct {
	Eggs   int
	Inc    int
	IncPct string
	Jos    int
	JosPct string
}

// HallEggs is a hall's egg entry on AsOf: Cat. 1–3, incubabile and the total
// (Cat. 1 + 2 + 3 + incubabile).
type HallEggs struct {
	Hall  string
	Cat   [3]int
	Inc   int
	Total int
}

// EggWeight is the egg-weighted mean egg weight in g (1 dp): the farm on
// AsOf, hall H7 over SAPT 58 and the farm over SAPT 58.
type EggWeight struct {
	FarmDay    string
	H7Week58   string
	FarmWeek58 string
}

// LayWeek is golden 18: a week row's label and length, eggs and hen-days, and
// its rates (1 dp): lay %, guide, lay − guide in points, % incubabile, % jos
// and % Cat. 1–3.
type LayWeek struct {
	Label     string
	DaysInRow int
	Eggs      int
	HenDays   int
	Lay       string
	Guide     string
	DiffPts   string
	IncPct    string
	JosPct    string
	CatPct    [3]string
}

// WeekRemovals is a week row's deaths (mortality + culls), sales and feed kg
// (1 dp).
type WeekRemovals struct {
	Deaths calc.PerSex[int]
	Sold   calc.PerSex[int]
	FeedKg calc.PerSex[string]
}

// Week58 is golden 20: the SAPT 58 week row.
type Week58 struct {
	Eggs     int
	HenDays  int
	Lay      string
	Guide    string
	IncPct   string
	CatPct   [3]string
	Deaths   calc.PerSex[int]
	Sold     calc.PerSex[int]
	AliveEnd calc.PerSex[int]
}

// WeekFeed is a week's feed kg and g/pasăre (kg ÷ bird-days on fed days),
// 1 dp.
type WeekFeed struct {
	Kg calc.PerSex[string]
	G  calc.PerSex[string]
}

// WeekLay is a week row's eggs, lay % and guide (1 dp).
type WeekLay struct {
	Sapt  int
	Eggs  int
	Lay   string
	Guide string
}

// Weeks5756 is golden 22: SAPT 57 and 56, and the % incubabile of SAPT 56.
type Weeks5756 struct {
	Weeks    []WeekLay
	IncPct56 string
}

// PeakLay is the week of the cycle with the highest lay % (1 dp), with its
// eggs and hen-days.
type PeakLay struct {
	Sapt    int
	Eggs    int
	HenDays int
	Lay     string
}

// Foot is golden 24: the totals row of the first page (SAPT 56–59).
type Foot struct {
	Label   string
	Eggs    int
	HenDays int
	Lay     string
	Guide   string
	Deaths  calc.PerSex[int]
	Sold    calc.PerSex[int]
	IncPct  string
	FeedKg  calc.PerSex[string]
}

// LoadMore is golden 25: the load-more label of the first page and the
// cycle's weeks.
type LoadMore struct {
	Label      string
	SaptFrom   int
	SaptTo     int
	WeeksTotal int
}

// EmptyHall is a hall emptied in the cycle: the summary's empty-hall tile and
// the feed stat's note.
type EmptyHall struct {
	Hall  string
	Since string
	Tile  string
	Note  string
}

// openCycle returns the golden farm and its cycle on AsOf with its guide.
func openCycle(ctx context.Context, db dbq.DBTX) (uuid.UUID, calc.CycleInfo, error) {
	farm, err := FarmID(ctx, db)
	if err != nil {
		return uuid.Nil, calc.CycleInfo{}, err
	}
	cycles, err := calc.FarmCycles(ctx, db, farm)
	if err != nil {
		return uuid.Nil, calc.CycleInfo{}, err
	}
	for _, c := range cycles {
		if !c.StartDate.After(AsOf) && (c.EndDate == nil || !c.EndDate.Before(AsOf)) {
			return farm, c, nil
		}
	}
	return uuid.Nil, calc.CycleInfo{}, fmt.Errorf("no cycle of %q on %s: %w", FarmKey, AsOf, calc.ErrNoCycle)
}

// firstPage is the farm's Rezumat as the S4 page opens it on AsOf: the open
// cycle, the newest 4 weeks with their days.
func firstPage(ctx context.Context, db dbq.DBTX) (periodtable.TableResponse, error) {
	farm, err := FarmID(ctx, db)
	if err != nil {
		return periodtable.TableResponse{}, err
	}
	return periodtable.Build(ctx, db, periodtable.Input{
		Query: periodtable.TableQuery{View: "farm", FarmID: &farm},
		Today: AsOf, Now: Now,
	})
}

// week returns week k of t.
func week(t periodtable.TableResponse, k int) (periodtable.WeekRow, error) {
	for _, w := range t.Weeks {
		if w.Sapt == k {
			return w, nil
		}
	}
	return periodtable.WeekRow{}, fmt.Errorf("the table has no SAPT %d", k)
}

// weekSeries returns the farm's weeks over its cycle (WeekSeries) by SAPT.
func weekSeries(ctx context.Context, db dbq.DBTX) (map[int]periodtable.SeriesPoint, error) {
	farm, c, err := openCycle(ctx, db)
	if err != nil {
		return nil, err
	}
	points, err := periodtable.WeekSeries(ctx, db, farm, c, calc.Period{Kind: calc.PeriodCycle}, AsOf)
	if err != nil {
		return nil, err
	}
	out := make(map[int]periodtable.SeriesPoint, len(points))
	for _, p := range points {
		out[p.Sapt] = p
	}
	return out, nil
}

// point is week k of the series.
func point(s map[int]periodtable.SeriesPoint, k int) (periodtable.SeriesPoint, error) {
	p, ok := s[k]
	if !ok {
		return p, fmt.Errorf("the week series has no SAPT %d", k)
	}
	return p, nil
}

// sameLay checks that the rollup the hen-days are read from gives the table
// cell's lay % (golden 20/24: one rollup for both).
func sameLay(r calc.Rollup, c cells) error {
	cell := c.dec(periodtable.KeyLayPct)
	lay := r.LayPct()
	if c.err != nil {
		return c.err
	}
	if cell == nil || lay == nil || !cell.Equal(*lay) {
		return fmt.Errorf("lay %% of the rollup (%v) is not the table's (%v)", lay, cell)
	}
	return nil
}

func summary(ctx context.Context, db dbq.DBTX) (farms.SummaryResponse, error) {
	farm, err := FarmID(ctx, db)
	if err != nil {
		return farms.SummaryResponse{}, err
	}
	return farms.BuildSummary(ctx, db, farms.Input{FarmID: farm, Today: AsOf, Now: Now})
}

// layToday is the summary's "Ouat azi" stat.
func layToday(s farms.SummaryResponse) (farms.LayTodayStat, error) {
	for _, st := range s.Stats {
		if l, ok := st.(farms.LayTodayStat); ok {
			return l, nil
		}
	}
	return farms.LayTodayStat{}, fmt.Errorf("the summary has no %s stat", farms.StatLayToday)
}

func guideLay(ctx context.Context, env Env) (any, error) {
	s, err := summary(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	lay, err := layToday(s)
	if err != nil {
		return nil, err
	}
	_, c, err := openCycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	if c.Guide == nil {
		return nil, fmt.Errorf("the cycle of %q has no guide", FarmKey)
	}
	g, err := calc.ReadGuideWeeks(ctx, env.DB, c.Guide.ID, 53, 59)
	if err != nil {
		return nil, err
	}
	guide := func(k int) string {
		if w := g[k].F; w != nil {
			return fixed(w.LayPct, 1)
		}
		return ""
	}
	out := GuideLay{Guide59: numFixed(lay.Guide, 1), DiffToday: numFixed(lay.DiffPts, 1), Guide53: guide(53), Guide56: guide(56)}
	if lay.Status != nil {
		out.StatusToday = *lay.Status
	}
	return out, nil
}

// layPerHall reads the lay % of each hall tile (layHall) and compares it with
// the lay guide of the day of the summary's "Ouat azi" stat.
func layPerHall(ctx context.Context, env Env) (any, error) {
	s, err := summary(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	lay, err := layToday(s)
	if err != nil {
		return nil, err
	}
	out := []HallLay{}
	for _, t := range s.Tiles {
		h, ok := t.(farms.HallTile[farms.LayingTileSection])
		if !ok || len(h.Sections) == 0 || h.Sections[0].LayHall == nil {
			continue
		}
		pct := h.Sections[0].LayHall.Decimal()
		var guide *decimal.Decimal
		if lay.Guide != nil {
			g := lay.Guide.Decimal()
			guide = &g
		}
		out = append(out, HallLay{Hall: h.Hall, Pct: fixed(&pct, 1), DiffPts: fixed(calc.DiffPts(&pct, guide), 1)})
	}
	return out, nil
}

// dayRow returns day d of t.
func dayRow(t periodtable.TableResponse, d timex.Day) (periodtable.DayRow, error) {
	for _, w := range t.Weeks {
		for _, r := range w.Days {
			if r.Date == d {
				return r, nil
			}
		}
	}
	return periodtable.DayRow{}, fmt.Errorf("the table has no day %s", d)
}

func eggsDay(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	d, err := dayRow(t, AsOf)
	if err != nil {
		return nil, err
	}
	c := cells{c: d.Cells}
	out := EggsDay{
		Eggs: c.int(periodtable.KeyEggs), Inc: c.int(periodtable.KeyInc), IncPct: c.fixed(periodtable.KeyIncPct, 1),
		Jos: c.int(periodtable.KeyJos), JosPct: c.fixed(periodtable.KeyJosPct, 1),
	}
	return out, c.err
}

// hallEggs reads the PeriodTable day query per hall on AsOf: the halls with
// an egg entry, by code.
func hallEggs(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	days, err := periodtable.ReadDays(ctx, env.DB, periodtable.DaysQuery{FarmID: farm, Scope: periodtable.ScopeHall, From: AsOf, To: AsOf})
	if err != nil {
		return nil, err
	}
	out := []HallEggs{}
	for _, d := range days.Rows {
		if d.EggRows > 0 {
			out = append(out, HallEggs{Hall: d.Code, Cat: [3]int{d.Cat1, d.Cat2, d.Cat3}, Inc: d.Hatching, Total: d.Eggs})
		}
	}
	slices.SortFunc(out, func(a, b HallEggs) int { return strings.Compare(a.Hall, b.Hall) })
	return out, nil
}

// eggWeight adds up the day query's rows into the calc rollup: the farm on
// AsOf, and H7 and the farm over the days of SAPT 58.
func eggWeight(ctx context.Context, env Env) (any, error) {
	farm, c, err := openCycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	from, to := calc.WeekFirstDay(c.Age, 58), calc.WeekLastDay(c.Age, 58)
	read := func(scope string, from, to timex.Day) (map[string]*calc.Rollup, error) {
		days, err := periodtable.ReadDays(ctx, env.DB, periodtable.DaysQuery{FarmID: farm, Scope: scope, From: from, To: to})
		if err != nil {
			return nil, err
		}
		out := map[string]*calc.Rollup{}
		for _, d := range days.Rows {
			if out[d.Code] == nil {
				out[d.Code] = &calc.Rollup{}
			}
			out[d.Code].Add(d.DayFigures)
		}
		return out, nil
	}
	weight := func(r map[string]*calc.Rollup, code string) string {
		if r[code] == nil {
			return ""
		}
		return fixed(r[code].EggWeight(), 1)
	}
	today, err := read(periodtable.ScopeFarm, AsOf, AsOf)
	if err != nil {
		return nil, err
	}
	halls, err := read(periodtable.ScopeHall, from, to)
	if err != nil {
		return nil, err
	}
	week, err := read(periodtable.ScopeFarm, from, to)
	if err != nil {
		return nil, err
	}
	return EggWeight{FarmDay: weight(today, ""), H7Week58: weight(halls, "H7"), FarmWeek58: weight(week, "")}, nil
}

// catPct is a rollup's % Cat. 1–3 (1 dp).
func catPct(r calc.Rollup) [3]string {
	p := r.CatPct()
	return [3]string{fixed(p[0], 1), fixed(p[1], 1), fixed(p[2], 1)}
}

// layWeek59 reads the SAPT 59 row of the first page, and its hen-days and
// Cat. shares from the same week's rollup (WeekSeries).
func layWeek59(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	w, err := week(t, 59)
	if err != nil {
		return nil, err
	}
	s, err := weekSeries(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	p, err := point(s, 59)
	if err != nil {
		return nil, err
	}
	c := cells{c: w.Cells}
	out := LayWeek{
		Label: w.Label, DaysInRow: w.DaysInRow, Eggs: c.int(periodtable.KeyEggs), HenDays: p.Rollup.HenDays,
		Lay: c.fixed(periodtable.KeyLayPct, 1), Guide: c.fixed(periodtable.KeyLayGuide, 1),
		DiffPts: fixed(calc.DiffPts(c.dec(periodtable.KeyLayPct), c.dec(periodtable.KeyLayGuide)), 1),
		IncPct:  c.fixed(periodtable.KeyIncPct, 1), JosPct: c.fixed(periodtable.KeyJosPct, 1), CatPct: catPct(p.Rollup),
	}
	if err := sameLay(p.Rollup, c); err != nil {
		return nil, err
	}
	return out, c.err
}

func weekRemovals(k int) func(context.Context, Env) (any, error) {
	return func(ctx context.Context, env Env) (any, error) {
		t, err := firstPage(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		w, err := week(t, k)
		if err != nil {
			return nil, err
		}
		c := cells{c: w.Cells}
		out := WeekRemovals{
			Deaths: calc.PerSex[int]{F: c.int(periodtable.KeyDeathsF), M: c.int(periodtable.KeyDeathsM)},
			Sold:   calc.PerSex[int]{F: c.int(periodtable.KeySoldF), M: c.int(periodtable.KeySoldM)},
			FeedKg: calc.PerSex[string]{F: c.fixed(periodtable.KeyFeedKgF, 1), M: c.fixed(periodtable.KeyFeedKgM, 1)},
		}
		return out, c.err
	}
}

// week58 reads the SAPT 58 row of the first page, and its hen-days and Cat.
// shares from the same week's rollup (WeekSeries).
func week58(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	w, err := week(t, 58)
	if err != nil {
		return nil, err
	}
	s, err := weekSeries(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	p, err := point(s, 58)
	if err != nil {
		return nil, err
	}
	c := cells{c: w.Cells}
	out := Week58{
		Eggs: c.int(periodtable.KeyEggs), HenDays: p.Rollup.HenDays,
		Lay: c.fixed(periodtable.KeyLayPct, 1), Guide: c.fixed(periodtable.KeyLayGuide, 1),
		IncPct: c.fixed(periodtable.KeyIncPct, 1), CatPct: catPct(p.Rollup),
		Deaths:   calc.PerSex[int]{F: c.int(periodtable.KeyDeathsF), M: c.int(periodtable.KeyDeathsM)},
		Sold:     calc.PerSex[int]{F: c.int(periodtable.KeySoldF), M: c.int(periodtable.KeySoldM)},
		AliveEnd: calc.PerSex[int]{F: c.int(periodtable.KeyAliveF), M: c.int(periodtable.KeyAliveM)},
	}
	if err := sameLay(p.Rollup, c); err != nil {
		return nil, err
	}
	return out, c.err
}

// weekFeed reads week k's kg from the first page and its g/pasăre from the
// same week's rollup (the laying Rezumat has no g column).
func weekFeed(k int) func(context.Context, Env) (any, error) {
	return func(ctx context.Context, env Env) (any, error) {
		t, err := firstPage(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		w, err := week(t, k)
		if err != nil {
			return nil, err
		}
		s, err := weekSeries(ctx, env.DB)
		if err != nil {
			return nil, err
		}
		p, err := point(s, k)
		if err != nil {
			return nil, err
		}
		c := cells{c: w.Cells}
		g := p.Rollup.FeedGPerBird()
		out := WeekFeed{
			Kg: calc.PerSex[string]{F: c.fixed(periodtable.KeyFeedKgF, 1), M: c.fixed(periodtable.KeyFeedKgM, 1)},
			G:  calc.PerSex[string]{F: fixed(g.F, 1), M: fixed(g.M, 1)},
		}
		return out, c.err
	}
}

func weeks5756(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	out := Weeks5756{Weeks: []WeekLay{}}
	for _, k := range []int{57, 56} {
		w, err := week(t, k)
		if err != nil {
			return nil, err
		}
		c := cells{c: w.Cells}
		out.Weeks = append(out.Weeks, WeekLay{Sapt: k, Eggs: c.int(periodtable.KeyEggs),
			Lay: c.fixed(periodtable.KeyLayPct, 1), Guide: c.fixed(periodtable.KeyLayGuide, 1)})
		if k == 56 {
			out.IncPct56 = c.fixed(periodtable.KeyIncPct, 1)
		}
		if c.err != nil {
			return nil, c.err
		}
	}
	return out, nil
}

// peakLay is the week of the cycle's series (§7.6 readout) with the highest
// lay % at full precision, the first one on a tie.
func peakLay(ctx context.Context, env Env) (any, error) {
	farm, c, err := openCycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	points, err := periodtable.WeekSeries(ctx, env.DB, farm, c, calc.Period{Kind: calc.PeriodCycle}, AsOf)
	if err != nil {
		return nil, err
	}
	var peak *periodtable.SeriesPoint
	for i, p := range points {
		if p.Lay != nil && (peak == nil || p.Lay.GreaterThan(*peak.Lay)) {
			peak = &points[i]
		}
	}
	if peak == nil {
		return PeakLay{}, nil
	}
	return PeakLay{Sapt: peak.Sapt, Eggs: peak.Rollup.Eggs, HenDays: peak.Rollup.HenDays, Lay: fixed(peak.Lay, 1)}, nil
}

// foot reads the first page's totals, and their hen-days from the rollups of
// the loaded weeks (WeekSeries) combined with Plus.
func foot(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	if t.Totals == nil || t.Loaded == nil {
		return nil, fmt.Errorf("the first page has no totals")
	}
	s, err := weekSeries(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	var r calc.Rollup
	for k := t.Loaded.SaptFrom; k <= t.Loaded.SaptTo; k++ {
		p, err := point(s, k)
		if err != nil {
			return nil, err
		}
		r = r.Plus(p.Rollup)
	}
	c := cells{c: t.Totals.Cells}
	out := Foot{
		Label: t.Totals.Label, Eggs: c.int(periodtable.KeyEggs), HenDays: r.HenDays,
		Lay: c.fixed(periodtable.KeyLayPct, 1), Guide: c.fixed(periodtable.KeyLayGuide, 1),
		Deaths: calc.PerSex[int]{F: c.int(periodtable.KeyDeathsF), M: c.int(periodtable.KeyDeathsM)},
		Sold:   calc.PerSex[int]{F: c.int(periodtable.KeySoldF), M: c.int(periodtable.KeySoldM)},
		IncPct: c.fixed(periodtable.KeyIncPct, 1),
		FeedKg: calc.PerSex[string]{F: c.fixed(periodtable.KeyFeedKgF, 1), M: c.fixed(periodtable.KeyFeedKgM, 1)},
	}
	if err := sameLay(r, c); err != nil {
		return nil, err
	}
	return out, c.err
}

func loadMore(ctx context.Context, env Env) (any, error) {
	t, err := firstPage(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	out := LoadMore{SaptFrom: t.Period.SaptFrom, SaptTo: t.Period.SaptTo, WeeksTotal: t.Period.WeeksTotal}
	if t.More != nil {
		out.Label = t.More.Label
	}
	return out, nil
}

// emptyHalls reads the summary's empty-hall tiles, each with the note of the
// "Hrană azi" stat.
func emptyHalls(ctx context.Context, env Env) (any, error) {
	s, err := summary(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	var note string
	for _, st := range s.Stats {
		if f, ok := st.(farms.FeedTodayStat); ok {
			note = f.Note
		}
	}
	out := []EmptyHall{}
	for _, t := range s.Tiles {
		if e, ok := t.(farms.EmptyHallTile); ok {
			out = append(out, EmptyHall{Hall: e.Hall, Since: e.EmptySince.String(), Tile: e.Label, Note: note})
		}
	}
	return out, nil
}

// cells reads a row's cells into plain values; the first missing key or
// null count is kept in err.
type cells struct {
	c   periodtable.Cells
	err error
}

func (c *cells) dec(key string) *decimal.Decimal {
	v, ok := c.c[key]
	if !ok {
		if c.err == nil {
			c.err = fmt.Errorf("the row has no cell %q", key)
		}
		return nil
	}
	if v == nil {
		return nil
	}
	d := v.Decimal()
	return &d
}

// int is a count cell; a null count is an error.
func (c *cells) int(key string) int {
	d := c.dec(key)
	if d == nil {
		if c.err == nil {
			c.err = fmt.Errorf("the count cell %q is null", key)
		}
		return 0
	}
	if !d.IsInteger() {
		if c.err == nil {
			c.err = fmt.Errorf("the count cell %q is %s", key, d)
		}
		return 0
	}
	return int(d.IntPart())
}

// fixed is a rate cell at places decimals; null is "".
func (c *cells) fixed(key string, places int32) string { return fixed(c.dec(key), places) }

// numFixed is fixed for a wire number; nil is "".
func numFixed(n *num.Num, places int32) string {
	if n == nil {
		return ""
	}
	d := n.Decimal()
	return fixed(&d, places)
}
