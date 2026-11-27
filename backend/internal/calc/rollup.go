package calc

import (
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Rollups of §5.14: how a row (a day, a week, the loaded weeks) adds up its
// days. Every rate is Σ numerator ÷ Σ denominator over the row's days, never
// a mean of day rates; levels are taken at the row's last day; guides are
// weighted by hen-days or bird-days. Values stay at full precision; callers
// round once, at the end (§5.15).

// Status is the lamp of a guide comparison (§7.4 `status`).
type Status string

// Status values.
const (
	InNorma Status = "in_norma"
	Abatere Status = "abatere"
)

// Ratio is a rate kept as its two sums.
type Ratio struct {
	Num, Den decimal.Decimal
}

// Add adds one day's (or unit-day's) numerator and denominator.
func (r *Ratio) Add(num, den decimal.Decimal) {
	r.Num = r.Num.Add(num)
	r.Den = r.Den.Add(den)
}

// AddInt is Add for counts.
func (r *Ratio) AddInt(num, den int) {
	r.Add(decimal.NewFromInt(int64(num)), decimal.NewFromInt(int64(den)))
}

// Plus returns the two ratios' sums combined (a week into the totals).
func (r Ratio) Plus(o Ratio) Ratio {
	return Ratio{Num: r.Num.Add(o.Num), Den: r.Den.Add(o.Den)}
}

// Times returns k × Σnum ÷ Σden, or nil when Σden is 0 (a rate with no base is
// "–", never 0). k = 1000 turns kg ÷ bird-days into g/pasăre (§5.7).
func (r Ratio) Times(k decimal.Decimal) *decimal.Decimal {
	if r.Den.IsZero() {
		return nil
	}
	v := k.Mul(r.Num).Div(r.Den)
	return &v
}

// Value is Σnum ÷ Σden, or nil when Σden is 0.
func (r Ratio) Value() *decimal.Decimal { return r.Times(decimal.NewFromInt(1)) }

// Pct is 100 × Σnum ÷ Σden, or nil when Σden is 0.
func (r Ratio) Pct() *decimal.Decimal { return Pct(r.Num, r.Den) }

// Weighted is a guide over a row: Σ guide × weight ÷ Σ weight, the weight being
// the hen-days (lay, §5.6, D162) or the fed bird-days (feed, §5.7).
type Weighted struct {
	sum, weight decimal.Decimal
	gap         bool // a day with weight had no guide value
}

// Add adds one day's guide value with its weight. A day with weight ≤ 0 does
// not count. A day with weight and no guide value (nil, the guide has no row
// for that week) makes the whole row's guide nil: a partial guide is "–"
// (§7.4 States), never a guide over some of the days.
func (w *Weighted) Add(guide *decimal.Decimal, weight int) {
	if weight <= 0 {
		return
	}
	if guide == nil {
		w.gap = true
		return
	}
	wd := decimal.NewFromInt(int64(weight))
	w.sum = w.sum.Add(guide.Mul(wd))
	w.weight = w.weight.Add(wd)
}

// Plus combines two rows' guides.
func (w Weighted) Plus(o Weighted) Weighted {
	return Weighted{sum: w.sum.Add(o.sum), weight: w.weight.Add(o.weight), gap: w.gap || o.gap}
}

// Value is the weighted guide, or nil when no day had weight or one lacked a
// guide value.
func (w Weighted) Value() *decimal.Decimal {
	if w.gap || w.weight.IsZero() {
		return nil
	}
	v := w.sum.Div(w.weight)
	return &v
}

// Latest is a level (alive, stock): the value of the latest day added (§5.14:
// a week row's last day; the totals' latest loaded day).
type Latest[T any] struct {
	day timex.Day
	v   T
	set bool
}

// Add records v for day d; it wins when d is on or after the latest day so far.
func (l *Latest[T]) Add(d timex.Day, v T) {
	if !l.set || !d.Before(l.day) {
		l.day, l.v, l.set = d, v, true
	}
}

// Plus returns the later of the two levels.
func (l Latest[T]) Plus(o Latest[T]) Latest[T] {
	if o.set {
		l.Add(o.day, o.v)
	}
	return l
}

// Value returns the level and whether any day was added.
func (l Latest[T]) Value() (T, bool) { return l.v, l.set }

// Day returns the day of the level (the zero Day when none was added).
func (l Latest[T]) Day() timex.Day { return l.day }

// DiffPts is actual − guide in points (lay, §5.6), or nil when either is nil.
func DiffPts(actual, guide *decimal.Decimal) *decimal.Decimal {
	if actual == nil || guide == nil {
		return nil
	}
	d := actual.Sub(*guide)
	return &d
}

// DeviationPct is 100 × (actual ÷ guide − 1) (feed g §5.7, weight §5.10), or
// nil when either is nil or the guide is 0.
func DeviationPct(actual, guide *decimal.Decimal) *decimal.Decimal {
	if actual == nil || guide == nil || guide.IsZero() {
		return nil
	}
	d := actual.Sub(*guide).Mul(hundred).Div(*guide)
	return &d
}

// StatusOf is the lamp of a difference from the guide against a tolerance
// (app_setting tolerance.*, D19): |diff| > tolerance → Abatere, else InNorma,
// compared at full precision. It is false when diff is nil (no comparison).
func StatusOf(diff *decimal.Decimal, tolerance decimal.Decimal) (Status, bool) {
	if diff == nil {
		return "", false
	}
	if diff.Abs().GreaterThan(tolerance) {
		return Abatere, true
	}
	return InNorma, true
}

// Comparison is the outcome of one guide comparison, as StatusOf and
// CumMortalityStatus return it: OK is false when nothing was compared.
type Comparison struct {
	Status Status
	OK     bool
}

// Compared wraps a comparison's two results: Compared(StatusOf(diff, tol)).
func Compared(st Status, ok bool) Comparison { return Comparison{Status: st, OK: ok} }

// Worst combines comparisons into one lamp (a stat over both sexes, §7.6
// feedToday): Abatere when any made one is, InNorma when every one made is
// in norm, false when none was made.
func Worst(all ...Comparison) (Status, bool) {
	var out Status
	var ok bool
	for _, c := range all {
		if !c.OK {
			continue
		}
		if !ok || c.Status == Abatere {
			out, ok = c.Status, true
		}
	}
	return out, ok
}

// CumMortalityStatus is the lamp of the cumulative mortality % (§5.5): above
// the guide → Abatere, else InNorma. It is false when either is nil.
func CumMortalityStatus(actual, guide *decimal.Decimal) (Status, bool) {
	if actual == nil || guide == nil {
		return "", false
	}
	if actual.GreaterThan(*guide) {
		return Abatere, true
	}
	return InNorma, true
}

// DayFigures are one scope's raw sums for one day (§5.3–§5.7), as the
// PeriodTable day query returns them, with the guide of the day's SAPT.
type DayFigures struct {
	Date       timex.Day
	AliveStart PerSex[int]
	AliveEnd   PerSex[int]
	Mortality  PerSex[int]
	Culls      PerSex[int]
	Sold       PerSex[int]

	Eggs     int // Σ total_eggs = Cat. 1 + 2 + 3 + incubabile (Jos apart, D18)
	Hatching int // Σ hatching_eggs (incubabile)
	Floor    int // Σ floor_eggs (jos)
	Cat1     int
	Cat2     int
	Cat3     int
	// EggGrams is Σ eggs × avg_egg_weight_g and WeighedEggs Σ eggs over the
	// unit-days that have both (D176).
	EggGrams    decimal.Decimal
	WeighedEggs int
	// HenDays is Σ hens at the start of the day over the egg units whose day
	// counts for lay % (§5.6 as amended by M2 Q1: the unit has eggs that day
	// or is emptied that day). The query decides which unit-days count.
	HenDays int

	FeedKg      PerSex[decimal.Decimal]
	FedBirdDays PerSex[int] // Σ alive_start over the section × sex rows with an allocation

	GuideLay   *decimal.Decimal         // guide_week.lay_pct of SAPT(Date), ♀; nil = no row
	GuideFeedG PerSex[*decimal.Decimal] // guide_week.feed_g of SAPT(Date); nil = no row
}

// Rollup adds up DayFigures into one row: a day row (one day), a week row
// (its days) or the totals (Plus of the loaded week rows). The zero value is
// an empty row.
type Rollup struct {
	Mortality, Culls, Sold                  PerSex[int]
	Eggs, Hatching, Floor, Cat1, Cat2, Cat3 int
	HenDays                                 int
	FeedKg                                  PerSex[decimal.Decimal]
	FedBirdDays                             PerSex[int]
	alive                                   Latest[PerSex[int]]
	eggWeight                               Ratio
	guideLay                                Weighted
	guideFeedF, guideFeedM                  Weighted
}

// Add adds one day.
func (r *Rollup) Add(d DayFigures) {
	r.Mortality = addSex(r.Mortality, d.Mortality)
	r.Culls = addSex(r.Culls, d.Culls)
	r.Sold = addSex(r.Sold, d.Sold)
	r.Eggs += d.Eggs
	r.Hatching += d.Hatching
	r.Floor += d.Floor
	r.Cat1 += d.Cat1
	r.Cat2 += d.Cat2
	r.Cat3 += d.Cat3
	r.HenDays += d.HenDays
	r.FeedKg = PerSex[decimal.Decimal]{F: r.FeedKg.F.Add(d.FeedKg.F), M: r.FeedKg.M.Add(d.FeedKg.M)}
	r.FedBirdDays = addSex(r.FedBirdDays, d.FedBirdDays)
	r.alive.Add(d.Date, d.AliveEnd)
	r.eggWeight.Add(d.EggGrams, decimal.NewFromInt(int64(d.WeighedEggs)))
	r.guideLay.Add(d.GuideLay, d.HenDays)
	r.guideFeedF.Add(d.GuideFeedG.F, d.FedBirdDays.F)
	r.guideFeedM.Add(d.GuideFeedG.M, d.FedBirdDays.M)
}

// Plus combines two rows (week rows into the totals).
func (r Rollup) Plus(o Rollup) Rollup {
	return Rollup{
		Mortality:   addSex(r.Mortality, o.Mortality),
		Culls:       addSex(r.Culls, o.Culls),
		Sold:        addSex(r.Sold, o.Sold),
		Eggs:        r.Eggs + o.Eggs,
		Hatching:    r.Hatching + o.Hatching,
		Floor:       r.Floor + o.Floor,
		Cat1:        r.Cat1 + o.Cat1,
		Cat2:        r.Cat2 + o.Cat2,
		Cat3:        r.Cat3 + o.Cat3,
		HenDays:     r.HenDays + o.HenDays,
		FeedKg:      PerSex[decimal.Decimal]{F: r.FeedKg.F.Add(o.FeedKg.F), M: r.FeedKg.M.Add(o.FeedKg.M)},
		FedBirdDays: addSex(r.FedBirdDays, o.FedBirdDays),
		alive:       r.alive.Plus(o.alive),
		eggWeight:   r.eggWeight.Plus(o.eggWeight),
		guideLay:    r.guideLay.Plus(o.guideLay),
		guideFeedF:  r.guideFeedF.Plus(o.guideFeedF),
		guideFeedM:  r.guideFeedM.Plus(o.guideFeedM),
	}
}

// Deaths is the Rezumat "Morți": mortality + culls, as the register counts
// them (D165). Sales are never deaths (D155).
func (r Rollup) Deaths() PerSex[int] { return addSex(r.Mortality, r.Culls) }

// AliveEnd is "În viață": alive at the end of the row's latest day, false for
// an empty row.
func (r Rollup) AliveEnd() (PerSex[int], bool) { return r.alive.Value() }

// LayPct is 100 × Σ eggs ÷ Σ hen-days (§5.6), nil without hen-days.
func (r Rollup) LayPct() *decimal.Decimal { return intRatio(r.Eggs, r.HenDays).Pct() }

// GuideLay is the lay guide weighted by hen-days (§5.6, D162). For a day row
// show the guide of SAPT(d) itself (§5.14): it is defined even on a day
// without hen-days, where this is nil.
func (r Rollup) GuideLay() *decimal.Decimal { return r.guideLay.Value() }

// HatchingPct is 100 × Σ incubabile ÷ Σ eggs ("% incub.").
func (r Rollup) HatchingPct() *decimal.Decimal { return intRatio(r.Hatching, r.Eggs).Pct() }

// FloorPct is 100 × Σ jos ÷ Σ eggs ("% jos").
func (r Rollup) FloorPct() *decimal.Decimal { return intRatio(r.Floor, r.Eggs).Pct() }

// CatPct is 100 × Σ Cat. 1, 2, 3 ÷ Σ eggs (§5.6).
func (r Rollup) CatPct() [3]*decimal.Decimal {
	return [3]*decimal.Decimal{
		intRatio(r.Cat1, r.Eggs).Pct(),
		intRatio(r.Cat2, r.Eggs).Pct(),
		intRatio(r.Cat3, r.Eggs).Pct(),
	}
}

// EggWeight is the egg-weighted mean egg weight in g (D176), nil without
// weighed eggs.
func (r Rollup) EggWeight() *decimal.Decimal { return r.eggWeight.Value() }

// FeedGPerBird is g/pasăre = 1000 × Σ kg ÷ Σ bird-days on fed days (§5.7, D38).
func (r Rollup) FeedGPerBird() PerSex[*decimal.Decimal] {
	k := decimal.NewFromInt(1000)
	return PerSex[*decimal.Decimal]{
		F: Ratio{Num: r.FeedKg.F, Den: decimal.NewFromInt(int64(r.FedBirdDays.F))}.Times(k),
		M: Ratio{Num: r.FeedKg.M, Den: decimal.NewFromInt(int64(r.FedBirdDays.M))}.Times(k),
	}
}

// GuideFeedG is the feed guide g weighted by fed bird-days (§5.7).
func (r Rollup) GuideFeedG() PerSex[*decimal.Decimal] {
	return PerSex[*decimal.Decimal]{F: r.guideFeedF.Value(), M: r.guideFeedM.Value()}
}

func intRatio(num, den int) Ratio {
	var q Ratio
	q.AddInt(num, den)
	return q
}

func addSex(a, b PerSex[int]) PerSex[int] { return PerSex[int]{F: a.F + b.F, M: a.M + b.M} }
