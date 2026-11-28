package entries

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/forms/delivery (§7.8, M4 Q14 proposed). Numbers
// are at full precision (the client rounds, §5.15).

// DeliveryFormResponse is the Livrare hrană form state of a farm. The client
// derives "după livrare" = Stock[].Kg + the typed kg, and "ajunge X zile" =
// that ÷ AvgDaily; the save answers the server's stock (stockAfter).
type DeliveryFormResponse struct {
	Farm  FormFarm           `json:"farm"`
	Today timex.Day          `json:"today"`
	Cycle *DeliveryFormCycle `json:"cycle"` // the farm's open cycle; null without one
	// FeedTypes are the active types in name order (a delivery of a retired
	// type is refused).
	FeedTypes []FormFeedType `json:"feedTypes"`
	// Stock is the farm's stock per feed type at the end of today on the saved
	// data (§5.8, the feed form's rows); [] for a farm without an open cycle
	// (§5.8: its leftover stock is never shown).
	Stock []FormStock     `json:"stock"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// DeliveryFormCycle is the farm's open cycle.
type DeliveryFormCycle struct {
	ID        uuid.UUID `json:"id"`
	Series    string    `json:"series"`
	StartDate timex.Day `json:"startDate"`
	SaptToday int       `json:"saptToday"`
}
