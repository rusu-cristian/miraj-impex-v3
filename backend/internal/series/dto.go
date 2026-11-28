package series

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/series/{id}/page (§7.3, M4 D7, Q2–Q5). Lists are
// never nil: the builder initialises them, so they encode as [] (§7.0a).

// PageResponse is the series page (S3).
type PageResponse struct {
	Series Info `json:"series"`
	// Alive is Σ the count at the end of today over the caller's farms whose
	// open cycle is of this series; null for a closed series.
	Alive *httpx.SexPair[int] `json:"alive"`
	// OwnFarmsOnly is true when a farm with a cycle of the series is hidden
	// from the caller ("la fermele tale").
	OwnFarmsOnly bool `json:"ownFarmsOnly"`
	// SaptToday is the series' week today; null for a closed series.
	SaptToday *int `json:"saptToday"`
	// SaptSpan is [SAPT(first start), SAPT(last end, or today while a cycle
	// is open)] over the caller's cycles of the series.
	SaptSpan [2]int `json:"saptSpan"`
	// Placement is omitted when the caller can't see the placement farm.
	Placement *Placement `json:"placement,omitzero"`
	// End is present for a closed series only.
	End   *End    `json:"end,omitzero"`
	Track Track   `json:"track"`
	Farms []Panel `json:"farms"`
	// Tabs are pasari, hrana, greutate, and productie when one of the
	// caller's cycles of the series is laying (S3 has no Rezumat).
	Tabs []string        `json:"tabs"`
	AsOf timex.LocalTime `json:"asOf"`
}

// Info is the series row: its age parameters (§5.1) as stored.
type Info struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Status             string    `json:"status"` // "active" | "closed"
	PlacementDate      timex.Day `json:"placementDate"`
	PlacementInferred  bool      `json:"placementInferred"`
	AgeAtPlacementDays int       `json:"ageAtPlacementDays"`
}

// Placement is the header's "populată 07.10.2025 la Telița ♀ 28 080 · ♂ 3 510"
// ("început dedus 07.03.2025 la Hîrbovăț"): the series' first cycle (its
// start and farm, inferred when that start is), and the birds placed in the
// series' cycles (null when there is no placement event: Seria 2025).
type Placement struct {
	Date     timex.Day           `json:"date"`
	Inferred bool                `json:"inferred"`
	Farm     string              `json:"farm"`
	Placed   *httpx.SexPair[int] `json:"placed"`
}

// End is a closed series' end: the last end of the caller's cycles of the
// series and how that cycle ended (a cycle closed by the history import
// ended by emptying, as the workbook's flocks did).
type End struct {
	Date timex.Day `json:"date"`
	By   string    `json:"by"` // "emptied" | "manual"
}

// How a series ended (End.By).
const (
	EndEmptied = "emptied"
	EndManual  = "manual"
)

// Track is the series' cycle track: one lane per farm the caller sees.
type Track struct {
	Lanes []Lane    `json:"lanes"`
	Today timex.Day `json:"today"`
}

// Lane is one farm's lane: its cycles of the series, oldest first, and the
// transfer links that leave it for another lane, by date (each link is listed
// once, on its source lane).
type Lane struct {
	Farm      farms.FarmRef         `json:"farm"`
	Cycles    []LaneCycle           `json:"cycles"`
	Transfers []farms.TrackTransfer `json:"transfers"`
}

// LaneCycle is a farm track cycle with its type (the guide's, else the
// farm's).
type LaneCycle struct {
	farms.TrackCycle
	Type auth.FarmType `json:"type"`
}

// Panel is one item of PageResponse.Farms: a FarmPanel, or a HiddenFarmPanel
// for a farm with a cycle of the series that the caller can't see. The
// interface is sealed; each item encodes as its own type.
type Panel interface{ panel() }

func (FarmPanel) panel()       {}
func (HiddenFarmPanel) panel() {}

// FarmPanel is a farm the caller sees. Cycle is its cycle of the series (the
// open one, else the latest), null when it has none ("fără date pentru
// Seria 2025"); Guide is that cycle's guide (its snapshot, B4), null without
// a guide or without a cycle; Stats are that cycle's §7.6 stats over the
// whole cycle (farms.BuildCycleStats), [] without a cycle.
type FarmPanel struct {
	Farm   PanelFarm       `json:"farm"`
	Access bool            `json:"access"` // true
	Guide  *farms.GuideRef `json:"guide"`
	Cycle  *PanelCycle     `json:"cycle"`
	Stats  []farms.Stat    `json:"stats"`
}

// PanelFarm names a farm the caller sees, with its type.
type PanelFarm struct {
	ID   uuid.UUID     `json:"id"`
	Name string        `json:"name"`
	Type auth.FarmType `json:"type"`
}

// PanelCycle is a panel's cycle. SaptSpan is [SAPT(start), SAPT(end or today)].
type PanelCycle struct {
	ID            uuid.UUID  `json:"id"`
	StartDate     timex.Day  `json:"startDate"`
	EndDate       *timex.Day `json:"endDate"`
	Open          bool       `json:"open"`
	StartInferred bool       `json:"startInferred,omitzero"`
	SaptSpan      [2]int     `json:"saptSpan"`
}

// HiddenFarmPanel is a farm with a cycle of the series that the caller can't
// see: its name only ("Telița fără acces").
type HiddenFarmPanel struct {
	Farm   HiddenFarm `json:"farm"`
	Access bool       `json:"access"` // false
}

// HiddenFarm names a farm the caller can't see.
type HiddenFarm struct {
	Name string `json:"name"`
}
