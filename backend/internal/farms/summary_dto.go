package farms

import (
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/farms/{id}/summary (§7.6, §5.13, M2 D6, Q3, Q5).
// Numbers are at full precision (the client rounds, §5.15). A guide figure
// the comparison lacks is null with reason "no_guide" ("fără date în ghid";
// the page's guide tag tells a cycle without any guide); status is present
// only when there is a comparison.

// SummaryResponse is the Rezumat summary. Stats and Tiles are
// heterogeneous lists (keyed by id / by shape) of the typed items below.
// Alerts are M8 and Latest is the M6 journal: both [] until then (M2 Q3).
type SummaryResponse struct {
	Readout *Readout `json:"readout"` // null without a value to show (no cycle, no egg entry, no weighing)
	Stats   []Stat   `json:"stats"`
	Tiles   []Tile   `json:"tiles"`
	Alerts  []any    `json:"alerts"`
	Latest  []any    `json:"latest"`
}

// Stat is one item of SummaryResponse.Stats: AliveStat, CumMortalityStat,
// WeightStat, FeedTodayStat, FeedDaysLeftStat or LayTodayStat. The interface is sealed (its
// method is unexported); each item encodes as its own type.
type Stat interface{ stat() }

func (AliveStat) stat()        {}
func (CumMortalityStat) stat() {}
func (WeightStat) stat()       {}
func (FeedTodayStat) stat()    {}
func (FeedDaysLeftStat) stat() {}
func (LayTodayStat) stat()     {}

// Tile is one item of SummaryResponse.Tiles: a HallTile of TileSection
// (growth) or of LayingTileSection (laying), or an EmptyHallTile. Sealed
// like Stat.
type Tile interface{ tile() }

func (HallTile[S]) tile()   {}
func (EmptyHallTile) tile() {}

// Readout kinds (§5.13).
const (
	ReadoutKindLay    = "lay"    // laying: the latest week's lay % vs guide
	ReadoutKindWeight = "weight" // growth: the latest ♀ weighing vs the guide of SAPT − 1
)

// Readout is the key readout with its weekly series for the chart.
type Readout struct {
	Kind   string         `json:"kind"`
	Label  string         `json:"label"` // "SAPT 59 · efectiv 40,4% · ghid 57,6% · −17,2 pct"
	Value  num.Num        `json:"value"`
	Guide  *num.Num       `json:"guide"`
	Status *calc.Status   `json:"status,omitzero"`
	Series []ReadoutPoint `json:"series"`
}

// ReadoutPoint is one week of the readout series.
type ReadoutPoint struct {
	Sapt  int      `json:"sapt"`
	Value *num.Num `json:"value"`
	Guide *num.Num `json:"guide"`
}

// Stat ids, in display order.
const (
	StatAlive        = "alive"
	StatCumMortality = "cumMortality"
	StatWeight       = "weight"
	StatFeedToday    = "feedToday"
	StatFeedDaysLeft = "feedDaysLeft"
	StatLayToday     = "layToday"
)

// Reasons of a null figure (the §7.4 reason keys).
const (
	ReasonNoGuide    = "no_guide"    // "– fără date în ghid"
	ReasonNoWeighing = "no_weighing" // "– necântărit"
)

// AliveStat is "În viață" at the period's last day, with the cycle's arrivals
// (the §5.5 base, "la sosire") and sales up to that day (M2 Q5, proposed).
// Every figure is null for a cycle without a flock count (Seria 2025, like
// the page's alive, M2 D5; arrivals and sold per M3 Q10, proposed).
type AliveStat struct {
	ID       string              `json:"id"`
	F        *int                `json:"F"`
	M        *int                `json:"M"`
	Arrivals httpx.SexPair[*int] `json:"arrivals"`
	Sold     httpx.SexPair[*int] `json:"sold"`
}

// CumMortalityStat is "Mortalitate cumulată" (§5.5) at the period's last
// day: % per sex (null without a base, and for a cycle without a flock
// count), the ♀ guide and the deaths (null per sex for a cycle without a
// flock count, M3 Q10, proposed).
type CumMortalityStat struct {
	ID     string              `json:"id"`
	F      *num.Num            `json:"F"`
	M      *num.Num            `json:"M"`
	GuideF *num.Num            `json:"guideF"`
	Status *calc.Status        `json:"status,omitzero"`
	Reason string              `json:"reason,omitzero"`
	Note   string              `json:"note,omitzero"` // laying: "(ghid din intrarea în ouat)"
	Deaths httpx.SexPair[*int] `json:"deaths"`
}

// WeightStat is "Greutate": the latest weighing per sex in the cycle up to
// the period's last day; at and section are those of the ♀ weighing (the ♂
// one when there is no ♀ weighing), guide the guide of SAPT(at) − 1 for that
// sex (§5.10).
type WeightStat struct {
	ID      string       `json:"id"`
	F       *num.Num     `json:"F"`
	M       *num.Num     `json:"M"`
	Guide   *num.Num     `json:"guide"`
	Status  *calc.Status `json:"status,omitzero"`
	Reason  string       `json:"reason,omitzero"`
	At      *timex.Day   `json:"at"`
	Section *string      `json:"section"`
}

// FeedTodayStat is "Hrană azi pe pasăre" (§5.7): the bird-weighted g per sex
// on the day, vs the guide; note names the halls emptied in the cycle
// ("H7 goală din 14.11").
type FeedTodayStat struct {
	ID     string       `json:"id"`
	F      *num.Num     `json:"F"`
	M      *num.Num     `json:"M"`
	GuideF *num.Num     `json:"guideF"`
	GuideM *num.Num     `json:"guideM"`
	Status *calc.Status `json:"status,omitzero"`
	Reason string       `json:"reason,omitzero"`
	Note   string       `json:"note,omitzero"`
}

// FeedDaysLeftStat is the feed days left (§5.8, §7.6, M3 D10, Q7, proposed):
// the farm's stock rows as of today. It is in the stats only for the open
// cycle with a period that ends today, between feedToday and layToday.
type FeedDaysLeftStat struct {
	ID    string             `json:"id"`
	Items []FeedDaysLeftItem `json:"items"` // feed type name order; the client puts the lowest days first
}

// FeedDaysLeftItem is one feed type's stock at the end of today: days left
// (full precision, 0 below a stock of 0, null without consumption in the
// window), the run-out day (null with days), the kg and the lamp with its
// reason ("stoc sub 0 · verifică livrările", "sub pragul de 5 zile", "– fără
// consum în ultimele 3 zile"; absent when in_norma).
type FeedDaysLeftItem struct {
	Type   string      `json:"type"`
	ForSex *string     `json:"forSex"` // null = ambele sexe
	Days   *num.Num    `json:"days"`
	RunOut *timex.Day  `json:"runOut"`
	Kg     num.Num     `json:"kg"`
	Lamp   calc.Status `json:"lamp"`
	Reason string      `json:"reason,omitzero"`
}

// LayTodayStat is "Ouat azi" (§5.6), laying cycles only: the farm's lay % on
// the day (null without an egg entry), vs the guide, with the eggs and the
// number of halls that entered them.
type LayTodayStat struct {
	ID      string       `json:"id"`
	Value   *num.Num     `json:"value"`
	Guide   *num.Num     `json:"guide"`
	DiffPts *num.Num     `json:"diffPts"`
	Status  *calc.Status `json:"status,omitzero"`
	Reason  string       `json:"reason,omitzero"`
	Eggs    int          `json:"eggs"`
	Halls   int          `json:"halls"`
}

// Tile lamps (M2 D6).
const (
	LampMissing = "lipsa"   // a §5.11 gap today, after missing.from_hour
	LampAbatere = "abatere" // hall lay below guide − tolerance, or a section × sex g off the guide
)

// HallTile is one hall group of the section tiles (D41): S is TileSection
// for a growth cycle and LayingTileSection for a laying one.
type HallTile[S any] struct {
	Hall     string `json:"hall"`
	Sections []S    `json:"sections"`
}

// TileSection is one section tile. Weight is the section's latest ♀
// weighing in the cycle (M2 Q5, proposed), null when it has none
// ("– necântărit").
type TileSection struct {
	Code       string                  `json:"code"`
	Alive      httpx.SexPair[int]      `json:"alive"`
	DeathsWeek httpx.SexPair[int]      `json:"deathsWeek"` // mortality + culls in SAPT(day) ∩ cycle
	GToday     httpx.SexPair[*num.Num] `json:"gToday"`     // the typed g; null when not fed
	Weight     *TileWeight             `json:"weight"`
	Lamps      []string                `json:"lamps"`
}

// LayingTileSection adds the hall's lay % of the day (null without an egg
// entry) to a laying farm's tile.
type LayingTileSection struct {
	TileSection
	LayHall *num.Num `json:"layHall"`
}

// TileWeight is a tile's ♀ weighing with the guide of SAPT(at) − 1.
type TileWeight struct {
	F      num.Num   `json:"F"`
	At     timex.Day `json:"at"`
	Guide  *num.Num  `json:"guide"`
	Reason string    `json:"reason,omitzero"`
}

// EmptyHallTile is a hall emptied during the cycle: one muted line.
type EmptyHallTile struct {
	Hall       string    `json:"hall"`
	EmptySince timex.Day `json:"emptySince"`
	Label      string    `json:"label"` // "goală din 14.11.2026 · H7.1, H7.2"
}
