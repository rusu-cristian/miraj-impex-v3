package reports

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/reports/feed (§7.12, M3 D9, M4 D2, D6, Q6–Q8).
// Numbers are at full precision (the client rounds, §5.15).
//
// Prices are never sent to a viewer (§3.1, §7.0a, D64, D200): the viewer's
// response types (FeedResponse, FeedPage) have no price field at all, and the
// manager's and owner's (FeedResponsePriced, FeedPagePriced) add them. The
// service picks the type from the role.

// FeedBody is a body of GET /api/reports/feed: FeedResponse, FeedResponsePriced,
// FeedPage or FeedPagePriced (Service.Feed picks one from the role and the
// cursor). The interface is sealed (its method is unexported); each body
// encodes as its own type.
type FeedBody interface{ feedBody() }

func (FeedResponse) feedBody()       {}
func (FeedResponsePriced) feedBody() {}
func (FeedPage) feedBody()           {}
func (FeedPagePriced) feedBody()     {}

// FeedResponse is Rapoarte › Hrană for a viewer: Stoc and Livrări, no
// prices and no Cost.
type FeedResponse struct {
	Stock      []StockRow      `json:"stock"`
	Deliveries DeliveriesView  `json:"deliveries"`
	AsOf       timex.LocalTime `json:"asOf"`
}

// FeedResponsePriced is Rapoarte › Hrană for a manager or the owner: Stoc,
// Livrări with prices, and Cost.
type FeedResponsePriced struct {
	Stock      []StockRow       `json:"stock"`
	Deliveries DeliveriesPriced `json:"deliveries"`
	Cost       Cost             `json:"cost"`
	AsOf       timex.LocalTime  `json:"asOf"`
}

// FeedPage is a viewer's next page of Livrări rows (a request with a
// cursor, M4 Q7): no Stoc, no summary, no Cost.
type FeedPage struct {
	Deliveries DeliveriesPage  `json:"deliveries"`
	AsOf       timex.LocalTime `json:"asOf"`
}

// FeedPagePriced is a manager's or the owner's next page of Livrări rows.
type FeedPagePriced struct {
	Deliveries DeliveriesPagePriced `json:"deliveries"`
	AsOf       timex.LocalTime      `json:"asOf"`
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

// DeliveriesView is the viewer's Livrări: the deliveries of every farm's
// range (calc.DeliveryRange), the unpriced ones as a count only (D64), and
// the first page of rows, newest first.
type DeliveriesView struct {
	Total    int           `json:"total"`
	ByFarm   []FarmCount   `json:"byFarm"`
	Unpriced UnpricedCount `json:"unpriced"`
	Rows     []DeliveryRow `json:"rows"`
	// NextCursor reads the next page (null on the last one).
	NextCursor *string `json:"nextCursor"`
	// LastBefore is the newest delivery before the range, set only when
	// Total is 0 ("Ultima livrare: 10.02.2026 · 21 640 kg Pre-ouat").
	LastBefore *LastDelivery `json:"lastBefore"`
}

// DeliveriesPriced is the manager's and the owner's Livrări: as
// DeliveriesView, with the unpriced kg and days, and priced rows.
type DeliveriesPriced struct {
	Total      int                 `json:"total"`
	ByFarm     []FarmCount         `json:"byFarm"`
	Unpriced   UnpricedSummary     `json:"unpriced"`
	Rows       []DeliveryRowPriced `json:"rows"`
	NextCursor *string             `json:"nextCursor"`
	LastBefore *LastDelivery       `json:"lastBefore"`
}

// DeliveriesPage is a viewer's next page of rows.
type DeliveriesPage struct {
	Rows       []DeliveryRow `json:"rows"`
	NextCursor *string       `json:"nextCursor"`
}

// DeliveriesPagePriced is a manager's or the owner's next page of rows.
type DeliveriesPagePriced struct {
	Rows       []DeliveryRowPriced `json:"rows"`
	NextCursor *string             `json:"nextCursor"`
}

// FarmCount is one farm's delivery count ("Hîrbovăț 69 · Telița 13"), for
// the farms with at least one delivery in their range, by name.
type FarmCount struct {
	Farm  string `json:"farm"`
	Count int    `json:"count"`
}

// UnpricedCount is the viewer's "4 livrări fără preț": the count only.
type UnpricedCount struct {
	Count int `json:"count"`
}

// UnpricedSummary is the manager's and the owner's unpriced deliveries: the
// count, the kg and their distinct days, oldest first ("fără preț: 17.02,
// 14.03, 12.11.2026 · 18 160 kg").
type UnpricedSummary struct {
	Count int         `json:"count"`
	Kg    num.Num     `json:"kg"`
	Dates []timex.Day `json:"dates"`
}

// DeliveryRow is one Livrări row as the viewer gets it.
type DeliveryRow struct {
	ID     uuid.UUID `json:"id"`
	Date   timex.Day `json:"date"`
	Farm   string    `json:"farm"`
	Type   string    `json:"type"`
	ForSex *string   `json:"forSex"` // null = ambele sexe
	Kg     num.Num   `json:"kg"`
}

// DeliveryRowPriced is one Livrări row with its price (manager, owner).
type DeliveryRowPriced struct {
	DeliveryRow
	PricePerKg *num.Num `json:"pricePerKg"` // null = fără preț
	TotalLei   *num.Num `json:"totalLei"`
	// PricedBy is "din registru" for a migrated price, else the name of the
	// user who set it; null without a price.
	PricedBy *string `json:"pricedBy"`
	// PricedAt is when the price was set; null without one and for a
	// migrated price.
	PricedAt *timex.LocalTime `json:"pricedAt"`
	Tag      string           `json:"tag,omitzero"` // "fără preț" on an unpriced row
	Actions  DeliveryActions  `json:"actions"`
}

// DeliveryActions are the row's price actions: the owner sets the price of
// an unpriced row and changes a priced one, except inside a cycle of a
// closed series; a manager has neither.
type DeliveryActions struct {
	SetPrice    bool `json:"setPrice"`
	ChangePrice bool `json:"changePrice"`
}

// LastDelivery is the newest delivery before an empty range.
type LastDelivery struct {
	Date timex.Day `json:"date"`
	Farm string    `json:"farm"`
	Type string    `json:"type"`
	Kg   num.Num   `json:"kg"`
}

// Cost is the CostBlock (§5.9, M4 Q6): one row per farm's shown cycle with
// priced deliveries in its range, ordered by cycle start, and the series
// total when the rows cover every farm with a cycle of that one series.
type Cost struct {
	AsOf   timex.Day   `json:"asOf"`
	Cycles []CostCycle `json:"cycles"`
	Total  *CostTotal  `json:"total,omitzero"`
}

// CostCycle is one cycle's feed cost: the priced deliveries' count, kg and
// lei, lei per bird (÷ the birds at the cycle start, Birds) and per egg (÷
// the cycle's eggs in the range, Eggs; null with EggReason for a growth cycle
// or no eggs), and its unpriced deliveries (null when every one has a price).
type CostCycle struct {
	Label       string             `json:"label"` // "Telița · Seria 2026"
	Span        string             `json:"span"`  // "07.10.2025–28.02.2026" | "din 22.02.2026 · SAPT 59"
	PricedCount int                `json:"pricedCount"`
	Kg          num.Num            `json:"kg"`
	Lei         num.Num            `json:"lei"`
	Birds       httpx.SexPair[int] `json:"birds"`
	Eggs        *int               `json:"eggs"` // null for a growth cycle
	LeiPerBird  *num.Num           `json:"leiPerBird"`
	LeiPerEgg   *num.Num           `json:"leiPerEgg"`
	EggReason   string             `json:"eggReason,omitzero"`
	Unpriced    *CostUnpriced      `json:"unpriced"`
}

// CostUnpriced is a cost row's deliveries without a price, left out of the
// cost: "18 160 kg · 4 livrări, 17.02–12.11.2026 · neincluse în cost".
type CostUnpriced struct {
	Count int     `json:"count"`
	Kg    num.Num `json:"kg"`
	Span  string  `json:"span"` // "17.02–12.11.2026"
}

// CostTotal is the series total: Σ lei ÷ the series placement (Birds) and ÷
// Σ eggs of the rows (Eggs); ExcludedKg is the rows' unpriced kg.
type CostTotal struct {
	Label      string             `json:"label"` // "Total Seria 2026"
	Kg         num.Num            `json:"kg"`
	Lei        num.Num            `json:"lei"`
	Birds      httpx.SexPair[int] `json:"birds"`
	Eggs       int                `json:"eggs"`
	LeiPerBird *num.Num           `json:"leiPerBird"`
	LeiPerEgg  *num.Num           `json:"leiPerEgg"`
	EggReason  string             `json:"eggReason,omitzero"`
	ExcludedKg num.Num            `json:"excludedKg"`
}
