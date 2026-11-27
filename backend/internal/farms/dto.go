package farms

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of SPEC §7.2 (bootstrap) and §7.3 (farm page). Lists are never
// nil: the builders initialise them, so they encode as [] (§7.0a).

// GuideRef is a guide tag: { id, name }. A nil *GuideRef is "fără ghid".
type GuideRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// FarmRef names a farm.
type FarmRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// PageResponse is GET /api/farms/{id}/page (§7.3, M2 D5, Q4).
type PageResponse struct {
	Farm PageFarm `json:"farm"`
	// Cycle is the cycle shown: the asked one, else the open one, else the
	// latest; null when the farm never had a cycle.
	Cycle *PageCycle `json:"cycle"`
	// Alive is the farm's count at the end of the shown cycle's last day
	// (today for the open cycle); null without a cycle or for a cycle without
	// a flock count (Seria 2025 at Hîrbovăț).
	Alive *httpx.SexPair[int] `json:"alive"`
	// Cycles is the CycleSelector: every cycle of the farm, newest first.
	Cycles  []CycleItem `json:"cycles"`
	Track   Track       `json:"track"`
	Actions Actions     `json:"actions"`
	Tabs    []string    `json:"tabs"`
	// Empty is the empty-farm state, present while the farm has no open cycle.
	Empty *EmptyState     `json:"empty,omitzero"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// PageFarm is the page header's farm. Guide is the guide the shown cycle
// compares against (the cycle's snapshot, B4), the farm's own guide when no
// cycle is shown; null = "fără ghid".
type PageFarm struct {
	ID    uuid.UUID     `json:"id"`
	Name  string        `json:"name"`
	Type  auth.FarmType `json:"type"`
	Guide *GuideRef     `json:"guide"`
}

// PageCycle is the shown cycle. SaptToday and AgeDaysToday are taken on the
// cycle's last shown day: today for the open cycle, its end for a past one.
// Weeks = SaptToday − SaptStart + 1.
type PageCycle struct {
	ID           uuid.UUID  `json:"id"`
	Series       string     `json:"series"`
	StartDate    timex.Day  `json:"startDate"`
	EndDate      *timex.Day `json:"endDate"`
	SaptStart    int        `json:"saptStart"`
	SaptToday    int        `json:"saptToday"`
	AgeDaysToday int        `json:"ageDaysToday"`
	Weeks        int        `json:"weeks"`
}

// CycleItem is one CycleSelector line.
type CycleItem struct {
	ID            uuid.UUID `json:"id"`
	Series        string    `json:"series"`
	Label         string    `json:"label"` // "22.02.2026–azi · SAPT 20–59"
	Open          bool      `json:"open"`
	StartInferred bool      `json:"startInferred,omitzero"`
}

// Track is the compact farm track (M2 Q4 a): one lane, this farm's.
type Track struct {
	Lanes []TrackLane `json:"lanes"`
	Today timex.Day   `json:"today"`
}

// TrackLane is one farm's lane: its cycles, oldest first, and its transfer
// links to and from other farms, by date.
type TrackLane struct {
	Farm      FarmRef         `json:"farm"`
	Cycles    []TrackCycle    `json:"cycles"`
	Transfers []TrackTransfer `json:"transfers"`
}

// TrackCycle is one cycle on a lane. SaptSpan is [SAPT(start), SAPT(end or today)].
type TrackCycle struct {
	ID            uuid.UUID  `json:"id"`
	Series        string     `json:"series"`
	StartDate     timex.Day  `json:"startDate"`
	EndDate       *timex.Day `json:"endDate"`
	Open          bool       `json:"open"`
	StartInferred bool       `json:"startInferred,omitzero"`
	SaptSpan      [2]int     `json:"saptSpan"`
}

// TrackTransfer is a transfer link between two farms on one day, with the
// birds that arrived.
type TrackTransfer struct {
	Date     timex.Day `json:"date"`
	FromFarm string    `json:"fromFarm"`
	ToFarm   string    `json:"toFarm"`
	F        int       `json:"F"`
	M        int       `json:"M"`
}

// Actions are the page actions the caller may use (M2 D5).
type Actions struct {
	Move        bool `json:"move"`
	Sell        bool `json:"sell"`
	NewCycle    bool `json:"newCycle"`
	ChangeGuide bool `json:"changeGuide"`
}

// EmptyState is the empty-farm state (M2 Q4 b, c). Since is omitted for a
// farm that never had a cycle.
type EmptyState struct {
	Since *timex.Day `json:"since,omitzero"`
	Label string     `json:"label"` // "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21."
	Hint  string     `json:"hint"`  // "Ciclul pornește la prima populare sau transfer."
}

// The page tabs (§7.3); Producție only for laying cycles.
const (
	TabRezumat   = "rezumat"
	TabPasari    = "pasari"
	TabHrana     = "hrana"
	TabGreutate  = "greutate"
	TabProductie = "productie"
)

// BootstrapResponse is GET /api/bootstrap (§7.2, M2 D7).
type BootstrapResponse struct {
	Farms     []BootstrapFarm `json:"farms"`
	Series    []SeriesItem    `json:"series"`
	FeedTypes []FeedTypeItem  `json:"feedTypes"`
	Settings  Settings        `json:"settings"`
}

// BootstrapFarm is one of the caller's farms with its own guide (the guide of
// its next cycle), its open cycle (null when empty) and its section tree.
type BootstrapFarm struct {
	ID        uuid.UUID     `json:"id"`
	Name      string        `json:"name"`
	Type      auth.FarmType `json:"type"`
	Guide     *GuideRef     `json:"guide"`
	OpenCycle *OpenCycle    `json:"openCycle"`
	Sections  []SectionNode `json:"sections"`
}

// OpenCycle is a farm's open cycle on today.
type OpenCycle struct {
	ID           uuid.UUID `json:"id"`
	SeriesID     uuid.UUID `json:"seriesId"`
	SeriesName   string    `json:"seriesName"`
	StartDate    timex.Day `json:"startDate"`
	SaptStart    int       `json:"saptStart"`
	SaptToday    int       `json:"saptToday"`
	AgeDaysToday int       `json:"ageDaysToday"`
}

// SectionNode is a top-level active section; a hall lists its active
// compartments.
type SectionNode struct {
	ID       uuid.UUID     `json:"id"`
	Code     string        `json:"code"`
	IsHall   bool          `json:"isHall"`
	Children []SectionLeaf `json:"children"`
}

// SectionLeaf is a compartment of a hall.
type SectionLeaf struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
}

// SeriesItem is a series with a cycle on the caller's farms. SaptSpan is
// SAPT(first cycle start) … SAPT(last cycle end, or today while one is open),
// over those farms' cycles.
type SeriesItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"` // "active" | "closed"
	SaptSpan [2]int    `json:"saptSpan"`
}

// FeedTypeItem is an active feed type; ForSex null = both sexes.
type FeedTypeItem struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	ForSex *string   `json:"forSex"`
	Active bool      `json:"active"`
}

// Settings are the thresholds the client shows (app_setting, D19).
type Settings struct {
	ToleranceWeightPct num.Num `json:"toleranceWeightPct"`
	ToleranceLayPts    num.Num `json:"toleranceLayPts"`
	ToleranceFeedPct   num.Num `json:"toleranceFeedPct"`
	StockDaysMin       num.Num `json:"stockDaysMin"`
}
