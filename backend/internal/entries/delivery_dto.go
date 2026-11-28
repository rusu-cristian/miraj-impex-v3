package entries

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// DeliveryCreateRequest is the body of POST /api/entries/delivery (§6.7,
// §7.8, M4 D5): one feed delivery of one type to one farm on one day.
type DeliveryCreateRequest struct {
	httpx.WriteMeta           // clientRequestId, source, voiceTranscript, voiceToken, dryRun
	FarmID          uuid.UUID `json:"farmId"`
	Date            timex.Day `json:"date"`
	FeedTypeID      uuid.UUID `json:"feedTypeId"`
	QtyKg           *num.Num  `json:"qtyKg"` // > 0, at most one decimal (numeric(12,1))
	Note            *string   `json:"note"`  // trimmed; "" = none; at most 200 characters
	// PricePerKg and TotalLei are declared so that a price in the body is
	// never an unknown-field 400: any value, null too, is "present" (a nil
	// RawMessage = absent; NOTES-verify M4). A manager gets 403
	// PRICE_FORBIDDEN, the owner 400 unknown_field (Q9): a delivery is created
	// without a price, which only Prețuri sets.
	PricePerKg json.RawMessage `json:"pricePerKg"`
	TotalLei   json.RawMessage `json:"totalLei"`
}

// DeliveryRefs are what the level-1 checks of a delivery name (§7.0b): the
// farm's name and the type's (nil when the type doesn't exist).
type DeliveryRefs struct {
	FarmName     string
	FeedTypeName *string
}

// Bounds of a delivery (feed_delivery.qty_kg numeric(12,1) > 0, note ≤ 200).
var deliveryQtyMax = decimal.New(1, 11) // exclusive: numeric(12,1) holds up to 99 999 999 999,9

const deliveryNoteMax = 200

// note is the request's note trimmed, nil when absent or blank.
func (r DeliveryCreateRequest) note() *string {
	if r.Note == nil {
		return nil
	}
	s := strings.TrimSpace(*r.Note)
	if s == "" {
		return nil
	}
	return &s
}

// Validate is the level-1 check of a delivery (§7.0b, M4 D5) in document
// order: the WriteMeta fields; farmId, date and feedTypeId required (an
// unknown type is out_of_range); qtyKg required and > 0 with the S13
// sentence naming the type and the farm ("Scrie câte kg de Breeder 2 s-au
// livrat la Hîrbovăț."), at most one decimal and inside numeric(12,1) (the
// generic sentence, M3 Q9); the trimmed note at most 200 characters. The
// price fields are checked before (checkDelivery), not here.
func (r DeliveryCreateRequest) Validate(v *httpx.Validator, refs DeliveryRefs) {
	ValidateWriteMeta(v, r.WriteMeta)
	v.Check(r.FarmID != uuid.Nil, "farmId", httpx.IssueRequired, messages.InvalidData())
	v.Check(!r.Date.IsZero(), "date", httpx.IssueRequired, messages.InvalidData())
	switch {
	case r.FeedTypeID == uuid.Nil:
		v.Add("feedTypeId", httpx.IssueRequired, messages.InvalidData())
	case r.FarmID != uuid.Nil && refs.FeedTypeName == nil:
		v.Add("feedTypeId", httpx.IssueOutOfRange, messages.InvalidData())
	}
	qtyMsg := messages.InvalidData() // the sentence needs both names
	if refs.FeedTypeName != nil && refs.FarmName != "" {
		qtyMsg = messages.DeliveryQtyRequired(*refs.FeedTypeName, refs.FarmName)
	}
	switch qty := r.QtyKg; {
	case qty == nil:
		v.Add("qtyKg", httpx.IssueRequired, qtyMsg)
	case qty.Decimal().Sign() <= 0:
		v.Add("qtyKg", httpx.IssueOutOfRange, qtyMsg)
	case !qty.Decimal().Equal(qty.Decimal().Round(1)), qty.Decimal().Round(1).Cmp(deliveryQtyMax) >= 0:
		v.Add("qtyKg", httpx.IssueOutOfRange, messages.InvalidData()) // numeric(12,1) would round or overflow
	}
	if n := r.note(); n != nil {
		v.Check(utf8.RuneCountInString(*n) <= deliveryNoteMax, "note", httpx.IssueTooLong, messages.InvalidData())
	}
}

// DeliveryResponse is the 201 body of a delivery save (§6.7, §7.8, M4 D5):
// the entry, the Toast, the stock of the type at the end of the delivery
// date after the save (Q13, full precision), no warnings, asOf.
type DeliveryResponse struct {
	Entry      httpx.EntryRef  `json:"entry"`
	Toast      string          `json:"toast"`
	StockAfter num.Num         `json:"stockAfter"`
	Warnings   []httpx.Warning `json:"warnings"`
	AsOf       timex.LocalTime `json:"asOf"`
}

// deliveryAuditValues is a delivery save's audit new_values (M4 D5).
type deliveryAuditValues struct {
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	Type       string    `json:"type"`
	Date       timex.Day `json:"date"`
	Kg         num.Num   `json:"kg"`
	Note       *string   `json:"note"`
}
