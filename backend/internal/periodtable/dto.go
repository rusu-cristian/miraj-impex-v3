// Package periodtable builds the PeriodTable of SPEC §7.4 (GET /api/table):
// the week and day rows of a farm, section or series over a period, paged 4
// weeks at a time (D190–D192).
//
// Build takes a dbq.DBTX and an explicit Today and Now, so it runs on a pool
// or inside the history import's transaction alike (ARC-6); it never reads
// the database's own "today" or the dense day views, which stop there. The
// formulas live in calc: this package reads the day sums (sql.go) and lays
// them out.
package periodtable

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// TableQuery is the parsed query string of GET /api/table (§7.4).
type TableQuery struct {
	View      string     // farm | section | series | report-production | report-feed
	FarmID    *uuid.UUID // farmId
	SectionID *uuid.UUID // sectionId
	SeriesID  *uuid.UUID // seriesId
	CycleID   *uuid.UUID // cycleId (nil = open cycle, else the latest)
	Tab       string     // rezumat | pasari | hrana | greutate | productie ("" = per-view default)
	Period    Period     // parsed from "cycle" | "weeks:4" | "interval:2026-11-01,2026-11-15"
	Detail    string     // farm | halls | sections
	Types     []string   // chips; nil = every type of the period
	Grain     string     // day | week
	Before    *int       // oldest SAPT already loaded (exclusive)
	Count     int        // weeks per page, default 4
	Sapt      *int       // one week only
	FromSapt  *int       // load every week from here up to the newest not yet loaded
}

// Period is the `period` parameter (§5.2). Its fields match calc.Period, so
// calc.Period(p) converts it.
type Period struct {
	Kind     string    // cycle | weeks | interval
	Weeks    int       // weeks:n
	From, To timex.Day // interval
}

// TableResponse is the body of GET /api/table.
type TableResponse struct {
	Scope   Scope           `json:"scope"`
	Period  PeriodInfo      `json:"period"`
	Chips   []string        `json:"chips"`
	Groups  []Group         `json:"groups"`
	Weeks   []WeekRow       `json:"weeks"`
	Loaded  *Loaded         `json:"loaded,omitzero"` // absent when weeks is empty
	More    *More           `json:"more"`            // null at the first week of the period
	Start   *StartNote      `json:"start"`           // null until the first week is loaded
	Totals  *Totals         `json:"totals,omitzero"`
	Notes   []string        `json:"notes"`
	Empty   *EmptyState     `json:"empty,omitzero"` // { "next": "Alege ciclul curent" }
	NoGuide bool            `json:"noGuide,omitzero"`
	AsOf    timex.LocalTime `json:"asOf"`
}

// Scope names what the table covers.
type Scope struct {
	View      string     `json:"view"`
	FarmID    *uuid.UUID `json:"farmId,omitzero"`
	SectionID *uuid.UUID `json:"sectionId,omitzero"`
	CycleID   *uuid.UUID `json:"cycleId,omitzero"`
	SeriesID  *uuid.UUID `json:"seriesId,omitzero"`
	Label     string     `json:"label"`
}

// PeriodInfo is the resolved period (§5.2).
type PeriodInfo struct {
	Kind       string    `json:"kind"`
	Label      string    `json:"label"`
	SaptFrom   int       `json:"saptFrom"`
	SaptTo     int       `json:"saptTo"`
	From       timex.Day `json:"from"`
	To         timex.Day `json:"to"`
	WeeksTotal int       `json:"weeksTotal"`
}

// Group is a column group.
type Group struct {
	ID     string `json:"id"`
	Record string `json:"record"`
	Label  string `json:"label"`
	Scope  string `json:"scope"` // "farm" or a section id for detail groups
	Leaves []Leaf `json:"leaves"`
}

// Leaf is one column of a group.
type Leaf struct {
	ID    string `json:"id"` // "alive.F", "egg.layPct", …
	Label string `json:"label,omitzero"`
	Kind  string `json:"kind,omitzero"` // count | level | rate | guide | status | text
}

// Cells are a row's values by leaf key.
type Cells map[string]*num.Num // nil value = "–"; its reason is in Reasons[key]

// WeekRow is one week (SAPT) row.
type WeekRow struct {
	Sapt      int               `json:"sapt"`
	From      timex.Day         `json:"from"`
	To        timex.Day         `json:"to"`
	DaysInRow int               `json:"daysInRow"`
	Partial   bool              `json:"partial,omitzero"`
	Label     string            `json:"label"`
	Cells     Cells             `json:"cells"`
	Status    map[string]string `json:"status,omitzero"`  // key → in_norma | abatere
	Reasons   map[string]string `json:"reasons,omitzero"` // key → missing | empty_since:2026-11-14 | no_guide | …
	Days      []DayRow          `json:"days,omitzero"`    // grain=day
	Halls     []WeekRow         `json:"halls,omitzero"`   // report-production: hall week rows
}

// DayRow is one day row inside a week.
type DayRow struct {
	Date    timex.Day         `json:"date"`
	AgeDays int               `json:"ageDays"`
	Sapt    int               `json:"sapt"`
	Weekday string            `json:"weekday"`       // Lu Ma Mi Jo Vi Sâ Du
	Farm    string            `json:"farm,omitzero"` // series view text column
	Cells   Cells             `json:"cells"`
	Status  map[string]string `json:"status,omitzero"`
	Reasons map[string]string `json:"reasons,omitzero"`
}

// Loaded spans the weeks loaded so far.
type Loaded struct {
	SaptFrom int `json:"saptFrom"`
	SaptTo   int `json:"saptTo"`
	Weeks    int `json:"weeks"`
}

// More is the load-more row.
type More struct {
	Before    int    `json:"before"`
	Remaining int    `json:"remaining"`
	Label     string `json:"label"`
}

// StartNote replaces More at the first week of the cycle.
type StartNote struct {
	Label string `json:"label"`
}

// Totals is the foot row over the loaded weeks.
type Totals struct {
	Label   string            `json:"label"`
	Cells   Cells             `json:"cells"`
	Status  map[string]string `json:"status,omitzero"`
	Reasons map[string]string `json:"reasons,omitzero"`
}

// EmptyState is the state of a period with no data (§7.4 States).
type EmptyState struct {
	Next string `json:"next"`
}
