package reports

import (
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/reports/feed (§7.12, M3 D9, Q6). Numbers are at
// full precision (the client rounds, §5.15).

// FeedResponse is Rapoarte › Hrană. In M3 it carries the Stoc part only;
// deliveries and cost come with M4 (Q6, proposed).
type FeedResponse struct {
	Stock []StockRow      `json:"stock"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// StockRow is one item of FeedResponse.Stock: a TypeStock per feed type of
// a farm with an open cycle, or one FarmStock for a farm without one. The
// interface is sealed (its method is unexported); each item encodes as its
// own type.
type StockRow interface{ stockRow() }

func (TypeStock) stockRow() {}
func (FarmStock) stockRow() {}

// TypeStock is one feed type's stock at the end of today (§5.8): the kg, the
// average daily consumption over the window, the days left (0 below a stock
// of 0, null without consumption) and the run-out day (null with days), with
// the lamp and its reason (absent when in_norma).
type TypeStock struct {
	Farm     string       `json:"farm"`
	Type     string       `json:"type"`
	ForSex   *string      `json:"forSex"` // null = ambele sexe
	Kg       num.Num      `json:"kg"`
	DaysLeft *num.Num     `json:"daysLeft"`
	AvgDaily num.Num      `json:"avgDaily"`
	Window   [2]timex.Day `json:"window"` // the farm's consumption window, both ends included
	RunOut   *timex.Day   `json:"runOut"`
	Lamp     calc.Status  `json:"lamp"`
	Reason   string       `json:"reason,omitzero"`
}

// FarmStock is the one row of a farm without an open cycle: lamp oprit,
// reason "fără ciclu deschis" and, when it had a cycle, the sub line
// "ultimul ciclu: Seria 2026, închis pe 28.02.2026" (S11).
type FarmStock struct {
	Farm   string      `json:"farm"`
	Lamp   calc.Status `json:"lamp"`
	Reason string      `json:"reason"`
	Sub    string      `json:"sub,omitzero"`
}
