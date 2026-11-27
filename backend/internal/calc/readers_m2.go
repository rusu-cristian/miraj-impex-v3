package calc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ErrSetting is returned by ReadAppSettings when a setting is missing or is
// not the number its key needs.
var ErrSetting = errors.New("calc: app setting missing or not a number")

// GuideCells is one guide row (guide_week) for a week and a sex. A nil cell
// is "– fără date în ghid" (D29).
type GuideCells struct {
	FeedG           *decimal.Decimal // Hrană g/pasăre/zi
	BodyWeightG     *decimal.Decimal // weight at the end of the week (D157)
	MortalityCumPct *decimal.Decimal // laying: cumulative from the start of lay (D161)
	LayPct          *decimal.Decimal // % ouat, ♀ (D184)
	HatchingPct     *decimal.Decimal // % incubabile of all eggs (D153)
	EggWeightG      *decimal.Decimal
}

// GuideWeeks is a guide's rows by age week SAPT (D151). A week or a sex the
// guide has no row for reads as nil: g[59].F.
type GuideWeeks map[int]PerSex[*GuideCells]

// cells is week k's row for sex, nil when the guide has none.
func (g GuideWeeks) cells(k int, sex dbq.Sex) *GuideCells {
	w := g[k]
	if sex == dbq.SexF {
		return w.F
	}
	return w.M
}

// LayPct is the guide's lay % (♀) for week k, nil when it has none (§5.6).
func (g GuideWeeks) LayPct(k int) *decimal.Decimal {
	if c := g.cells(k, dbq.SexF); c != nil {
		return c.LayPct
	}
	return nil
}

// FeedG is the guide's g/pasăre for week k and sex, nil when it has none
// (§5.7).
func (g GuideWeeks) FeedG(k int, sex dbq.Sex) *decimal.Decimal {
	if c := g.cells(k, sex); c != nil {
		return c.FeedG
	}
	return nil
}

// BodyWeightG is the guide's weight at the end of week k for sex, nil when it
// has none (§5.10, D157: a weighing compares with week SAPT − 1).
func (g GuideWeeks) BodyWeightG(k int, sex dbq.Sex) *decimal.Decimal {
	if c := g.cells(k, sex); c != nil {
		return c.BodyWeightG
	}
	return nil
}

// ReadGuideWeeks reads the rows of a guide for the weeks [fromSAPT, toSAPT].
func ReadGuideWeeks(ctx context.Context, db dbq.DBTX, guideID uuid.UUID, fromSAPT, toSAPT int) (GuideWeeks, error) {
	if fromSAPT > toSAPT {
		return nil, errRange
	}
	rows, err := dbq.New(db).CalcGuideWeeks(ctx, dbq.CalcGuideWeeksParams{
		GuideID: guideID, FromWeek: int32(fromSAPT), ToWeek: int32(toSAPT),
	})
	if err != nil {
		return nil, fmt.Errorf("calc: guide weeks %d–%d: %w", fromSAPT, toSAPT, err)
	}
	out := make(GuideWeeks, len(rows))
	for _, r := range rows {
		c := &GuideCells{
			FeedG:           nullDec(r.FeedG),
			BodyWeightG:     nullDec(r.BodyWeightG),
			MortalityCumPct: nullDec(r.MortalityCumPct),
			LayPct:          nullDec(r.LayPct),
			HatchingPct:     nullDec(r.HatchingPct),
			EggWeightG:      nullDec(r.EggWeightG),
		}
		w := out[int(r.Week)]
		if r.Sex == dbq.SexF {
			w.F = c
		} else {
			w.M = c
		}
		out[int(r.Week)] = w
	}
	return out, nil
}

// AppSettings are the numeric settings of app_setting that the derived
// figures use (D19 thresholds, §5.8, §5.11).
type AppSettings struct {
	ToleranceWeightPct decimal.Decimal // tolerance.weight_pct: weighing vs guide, |Δ| > it → abatere
	ToleranceLayPts    decimal.Decimal // tolerance.lay_pts: lay % vs guide, |Δ| > it points → abatere
	ToleranceFeedPct   decimal.Decimal // tolerance.feed_pct: g/pasăre vs guide, |Δ| > it % → abatere
	StockDaysMin       decimal.Decimal // alert.stock_days_min: days left below it → abatere
	StockAvgDays       int             // alert.stock_avg_days: days of the consumption average
	MortalityFactor    decimal.Decimal // alert.mortality_factor: Zilnic warning factor
	MissingFromHour    int             // missing.from_hour: today's gaps count from this local hour
	MissingDaysBack    int             // missing.days_back: Lipsesc date window
}

var settingKeys = []string{
	"tolerance.weight_pct", "tolerance.lay_pts", "tolerance.feed_pct",
	"alert.stock_days_min", "alert.stock_avg_days", "alert.mortality_factor",
	"missing.from_hour", "missing.days_back",
}

// ReadAppSettings reads the AppSettings. A missing key or a value that isn't
// a number (an integer for the hour and day counts) is ErrSetting.
func ReadAppSettings(ctx context.Context, db dbq.DBTX) (AppSettings, error) {
	rows, err := dbq.New(db).CalcSettings(ctx, settingKeys)
	if err != nil {
		return AppSettings{}, fmt.Errorf("calc: app settings: %w", err)
	}
	vals := make(map[string]string, len(rows))
	for _, r := range rows {
		vals[r.Key] = r.Value
	}
	var s AppSettings
	var errs []error
	parse := func(key string, whole bool) decimal.Decimal {
		d, err := decimal.NewFromString(vals[key])
		if err != nil || (whole && !d.IsInteger()) {
			errs = append(errs, fmt.Errorf("%w: %s", ErrSetting, key))
		}
		return d
	}
	dec := func(key string) decimal.Decimal { return parse(key, false) }
	integer := func(key string) int { return int(parse(key, true).IntPart()) }
	s.ToleranceWeightPct = dec("tolerance.weight_pct")
	s.ToleranceLayPts = dec("tolerance.lay_pts")
	s.ToleranceFeedPct = dec("tolerance.feed_pct")
	s.StockDaysMin = dec("alert.stock_days_min")
	s.StockAvgDays = integer("alert.stock_avg_days")
	s.MortalityFactor = dec("alert.mortality_factor")
	s.MissingFromHour = integer("missing.from_hour")
	s.MissingDaysBack = integer("missing.days_back")
	if len(errs) > 0 {
		return AppSettings{}, errors.Join(errs...)
	}
	return s, nil
}

// CycleGuide is the guide a cycle compares against (its snapshot, B4).
type CycleGuide struct {
	ID   uuid.UUID
	Name string
	Type dbq.FarmType
}

// CycleInfo is a cycle with what the farm page, the CycleSelector and the
// period resolution need (§4.6, §7.3).
type CycleInfo struct {
	Cycle
	FarmID        uuid.UUID
	SeriesName    string
	StartedBy     dbq.CycleStartKind
	StartInferred bool              // "început dedus" (D54)
	EndedBy       *dbq.CycleEndKind // nil while the cycle is open
	Guide         *CycleGuide       // nil = "fără ghid"
	// HasBirdEvents is false for a cycle without a flock count (Seria 2025 at
	// Hîrbovăț, D52/D54): no live bird event inside [start, end].
	HasBirdEvents bool
}

// Type is the type of the cycle: its guide's (a cycle's guide has the farm's
// type when the cycle opens, B1, B4), else farm, the farm's own type.
func (c CycleInfo) Type(farm dbq.FarmType) dbq.FarmType {
	if c.Guide != nil {
		return c.Guide.Type
	}
	return farm
}

// DefaultCycle is the cycle shown when none is asked for (§7.3, §7.4): the
// open one, else the latest. cycles are newest first, as FarmCycles returns
// them; false when there is none.
func DefaultCycle(cycles []CycleInfo) (CycleInfo, bool) {
	for _, c := range cycles {
		if c.EndDate == nil {
			return c, true
		}
	}
	if len(cycles) > 0 {
		return cycles[0], true
	}
	return CycleInfo{}, false
}

// CycleByID returns the cycle of the farm with that id, or ErrNoCycle when
// there is none, including when the id is another farm's cycle.
func CycleByID(ctx context.Context, db dbq.DBTX, farmID, cycleID uuid.UUID) (CycleInfo, error) {
	r, err := dbq.New(db).CalcCycleByID(ctx, dbq.CalcCycleByIDParams{CycleID: cycleID, FarmID: farmID})
	if errors.Is(err, pgx.ErrNoRows) {
		return CycleInfo{}, ErrNoCycle
	}
	if err != nil {
		return CycleInfo{}, fmt.Errorf("calc: cycle %s: %w", cycleID, err)
	}
	return cycleInfo(dbq.CalcFarmCyclesRow(r)), nil
}

// FarmCycles returns every cycle of the farm, newest first.
func FarmCycles(ctx context.Context, db dbq.DBTX, farmID uuid.UUID) ([]CycleInfo, error) {
	rows, err := dbq.New(db).CalcFarmCycles(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("calc: farm cycles: %w", err)
	}
	out := make([]CycleInfo, 0, len(rows))
	for _, r := range rows {
		out = append(out, cycleInfo(r))
	}
	return out, nil
}

func cycleInfo(r dbq.CalcFarmCyclesRow) CycleInfo {
	c := CycleInfo{
		Cycle: Cycle{
			ID:        r.ID,
			SeriesID:  r.SeriesID,
			StartDate: r.StartDate,
			EndDate:   r.EndDate,
			Age:       SeriesAge{PlacementDate: r.PlacementDate, AgeAtPlacementDays: int(r.AgeAtPlacementDays)},
		},
		FarmID:        r.FarmID,
		SeriesName:    r.SeriesName,
		StartedBy:     r.StartedBy,
		StartInferred: r.StartInferred,
		EndedBy:       r.EndedBy,
		HasBirdEvents: r.HasBirdEvents,
	}
	if r.GuideID != nil && r.GuideName != nil && r.GuideType != nil {
		c.Guide = &CycleGuide{ID: *r.GuideID, Name: *r.GuideName, Type: *r.GuideType}
	}
	return c
}

// Weighing is one section × sex weighing (§5.10) with the birds of that
// section × sex at the end of the weighing day, the weight of the
// bird-weighted mean at farm or hall scope (§5.10).
type Weighing struct {
	SectionID  uuid.UUID
	Code       string
	Date       timex.Day
	AvgWeightG decimal.Decimal
	AliveEnd   int
}

// LatestWeighings is the latest weighing per sex, at farm scope and per
// section, of a day range. A sex without a weighing reads as nil.
type LatestWeighings struct {
	// Farm is the farm's latest weighing per sex. When several sections
	// weighed that sex on that day, it is the first of them in tree order
	// (hall, then compartment).
	Farm PerSex[*Weighing]
	// Sections is each section's latest weighing per sex, by section id;
	// sections without a weighing are absent.
	Sections map[uuid.UUID]PerSex[*Weighing]
}

// ReadLatestWeighings reads the latest live weighings of the farm in
// [from, to]; pass the cycle start as from to keep an earlier cycle's
// weighings out.
func ReadLatestWeighings(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (LatestWeighings, error) {
	if from.After(to) {
		return LatestWeighings{}, errRange
	}
	rows, err := dbq.New(db).CalcLatestWeighings(ctx, dbq.CalcLatestWeighingsParams{FarmID: farmID, FromDay: from, ToDay: to})
	if err != nil {
		return LatestWeighings{}, fmt.Errorf("calc: latest weighings %s–%s: %w", from, to, err)
	}
	out := LatestWeighings{Sections: make(map[uuid.UUID]PerSex[*Weighing], len(rows))}
	for _, r := range rows {
		w := &Weighing{SectionID: r.SectionID, Code: r.Code, Date: r.Date, AvgWeightG: r.AvgWeightG, AliveEnd: int(r.AliveEnd)}
		s := out.Sections[r.SectionID]
		if r.Sex == dbq.SexF {
			s.F = w
			if out.Farm.F == nil { // rows come newest first, then in tree order, per sex
				out.Farm.F = w
			}
		} else {
			s.M = w
			if out.Farm.M == nil {
				out.Farm.M = w
			}
		}
		out.Sections[r.SectionID] = s
	}
	return out, nil
}

// SectionDay is one bird-holding section's figures for one day (§5.3, §5.7,
// §5.11).
type SectionDay struct {
	SectionID uuid.UUID
	Code      string
	HallID    uuid.UUID // the hall (egg unit) of the section, the section itself when it has no parent
	HallCode  string
	// AliveStart and AliveEnd are the birds at the start and end of the day.
	AliveStart PerSex[int]
	AliveEnd   PerSex[int]
	// MortalityEntered is true when the day has a live mortality row (even
	// entered as 0) for the section × sex (§5.11 Zilnic).
	MortalityEntered PerSex[bool]
	// Fed is true when the day has a live allocation for the section × sex.
	Fed PerSex[bool]
	// G is the typed g/pasăre (§5.7), nil when not fed or for a migrated
	// kg-only row.
	G PerSex[*decimal.Decimal]
	// Kg is the day's feed at full precision: the migrated kg, else
	// g × AliveStart / 1000 (DATA03); 0 when not fed.
	Kg PerSex[decimal.Decimal]
}

// SectionsOnDay returns every bird-holding section of the farm (a section
// without children) with its figures for day, in tree order (hall, then
// compartment).
func SectionsOnDay(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]SectionDay, error) {
	rows, err := dbq.New(db).CalcSectionFeedDay(ctx, dbq.CalcSectionFeedDayParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: sections on %s: %w", day, err)
	}
	out := make([]SectionDay, 0, len(rows))
	for _, r := range rows {
		out = append(out, SectionDay{
			SectionID:        r.SectionID,
			Code:             r.Code,
			HallID:           r.HallID,
			HallCode:         r.HallCode,
			AliveStart:       pair(r.AliveStartF, r.AliveStartM),
			AliveEnd:         pair(r.AliveEndF, r.AliveEndM),
			MortalityEntered: PerSex[bool]{F: r.MortalityEnteredF, M: r.MortalityEnteredM},
			Fed:              PerSex[bool]{F: r.FedF, M: r.FedM},
			G:                PerSex[*decimal.Decimal]{F: nullDec(r.GF), M: nullDec(r.GM)},
			Kg:               PerSex[decimal.Decimal]{F: r.KgF, M: r.KgM},
		})
	}
	return out, nil
}

// SectionEmpty is a section's count at the end of a day, both sexes together,
// and since when it is empty (§5.3). A hall counts its compartments.
type SectionEmpty struct {
	SectionID uuid.UUID
	Code      string
	IsHall    bool
	AliveEnd  int
	// EmptySince is the last day the count went from > 0 to 0, nil while
	// the section has birds or when it never had any.
	EmptySince *timex.Day
}

// EmptySince returns every section of the farm (halls and the sections that
// hold birds) with its count and empty_since at the end of day, in tree
// order (hall, then its compartments).
func EmptySince(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]SectionEmpty, error) {
	rows, err := dbq.New(db).CalcEmptySince(ctx, dbq.CalcEmptySinceParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: empty since %s: %w", day, err)
	}
	out := make([]SectionEmpty, 0, len(rows))
	for _, r := range rows {
		out = append(out, SectionEmpty{
			SectionID: r.SectionID, Code: r.Code, IsHall: r.IsHall, AliveEnd: int(r.AliveEnd), EmptySince: r.EmptySince,
		})
	}
	return out, nil
}

// nullDec is the value of n, or nil when it is null.
func nullDec(n decimal.NullDecimal) *decimal.Decimal {
	if !n.Valid {
		return nil
	}
	d := n.Decimal
	return &d
}
