package prices

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// reasonMax is the most characters a trimmed Motiv may have (§6.15).
const reasonMax = 200

// Ref is what the level-1 checks of a price write name about one delivery
// (§7.0b): its farm, date, kg, type name and current price (nil when
// unpriced), and whether it (or its entry) is deleted.
type Ref struct {
	FarmID  uuid.UUID
	Date    timex.Day
	Kg      decimal.Decimal
	Type    string
	PerKg   *decimal.Decimal
	Total   *decimal.Decimal
	Deleted bool
}

// Refs are the Refs of the asked deliveries by id; an unknown id has none.
type Refs map[uuid.UUID]Ref

// SetRequest is the body of POST /api/deliveries/prices (§7.13, M4 Q10): one
// item per delivery, each with exactly one of pricePerKg and totalLei (a JSON
// null is the same as absent). It has no dryRun and no clientRequestId (D3):
// both are unknown fields.
type SetRequest struct {
	Items []SetItem `json:"items"`
}

// SetItem is one delivery of a price batch.
type SetItem struct {
	ID         uuid.UUID `json:"id"`
	PricePerKg *num.Num  `json:"pricePerKg"` // lei/kg, > 0, at most 4 decimals
	TotalLei   *num.Num  `json:"totalLei"`   // lei, > 0, at most 2 decimals
}

// Validate is the level-1 check of a price batch (§6.8, §7.0b, M4 Q10), in
// document order: at least one item; each id given, once, and known; the
// price per calc.DerivePrice, the "> 0" sentences naming the delivery by type
// and date ("Cocoși 17.02.2026: scrie un preț mai mare de 0 lei/kg"), both
// prices "… scrie prețul pe kg sau totalul, nu amândouă.", anything else
// (neither, too many decimals, out of range, unknown id) the generic
// sentence. Deleted or priced deliveries are level 2 (409 STALE).
func (r SetRequest) Validate(v *httpx.Validator, refs Refs) {
	if !v.Check(len(r.Items) > 0, "items", httpx.IssueRequired, messages.InvalidData()) {
		return
	}
	seen := make(map[uuid.UUID]bool, len(r.Items))
	for i, it := range r.Items {
		path := "items[" + strconv.Itoa(i) + "]."
		switch ref, ok := refs[it.ID]; {
		case it.ID == uuid.Nil:
			v.Add(path+"id", httpx.IssueRequired, messages.InvalidData())
		case seen[it.ID]:
			v.Add(path+"id", httpx.IssueOutOfRange, messages.InvalidData())
		case !ok:
			seen[it.ID] = true
			v.Add(path+"id", httpx.IssueOutOfRange, messages.InvalidData())
		default:
			seen[it.ID] = true
			if _, err := derive(ref.Kg, it.PricePerKg, it.TotalLei); err != nil {
				v.Add(priceIssue(path, ref, err))
			}
		}
	}
}

// ChangeRequest is the body of PATCH /api/deliveries/{id}/price (§6.8, §7.13,
// M4 Q12): exactly one of pricePerKg and totalLei, and the Motiv.
type ChangeRequest struct {
	PricePerKg *num.Num `json:"pricePerKg"`
	TotalLei   *num.Num `json:"totalLei"`
	Reason     *string  `json:"reason"` // trimmed, 1–200 characters
}

// reason is the request's Motiv trimmed ("" when absent).
func (r ChangeRequest) reason() string {
	if r.Reason == nil {
		return ""
	}
	return strings.TrimSpace(*r.Reason)
}

// Validate is the level-1 check of a price change (§6.8, §6.15, M4 Q12) on
// the delivery ref: the price as in a batch; the Motiv required with the S11
// sentence "Breeder 2 13.11.2026: scrie motivul ca să schimbi prețul din
// 7,76 în 7,74 lei/kg" (the new price in lei/kg, derived when a total was
// typed; the generic sentence when the delivery has no price or the new one
// is invalid) and at most 200 characters (§6.15 sentence).
func (r ChangeRequest) Validate(v *httpx.Validator, ref Ref) {
	price, err := derive(ref.Kg, r.PricePerKg, r.TotalLei)
	if err != nil {
		v.Add(priceIssue("", ref, err))
	}
	switch reason := r.reason(); {
	case reason == "":
		msg := messages.InvalidData()
		if err == nil && ref.PerKg != nil {
			msg = messages.PriceChangeReasonRequired(ref.Type, ref.Date, *ref.PerKg, price.PerKg)
		}
		v.Add("reason", httpx.IssueRequired, msg)
	case utf8.RuneCountInString(reason) > reasonMax:
		v.Add("reason", httpx.IssueTooLong, messages.ReasonTooLong(utf8.RuneCountInString(reason), reasonMax))
	}
}

// UndoRequest is the body of POST /api/deliveries/prices/undo (§7.13, M4 D4).
type UndoRequest struct {
	UndoToken string `json:"undoToken"`
}

// Validate is the level-1 check of a price undo: the token is required.
func (r UndoRequest) Validate(v *httpx.Validator) {
	v.Check(r.UndoToken != "", "undoToken", httpx.IssueRequired, messages.InvalidData())
}

// derive is calc.DerivePrice over the request's JSON numbers.
func derive(kg decimal.Decimal, perKg, total *num.Num) (calc.Price, error) {
	dec := func(n *num.Num) *decimal.Decimal {
		if n == nil {
			return nil
		}
		d := n.Decimal()
		return &d
	}
	return calc.DerivePrice(kg, dec(perKg), dec(total))
}

// priceIssue is the issue of a calc.DerivePrice error for the item at prefix
// ("items[1]." or "" for a change), naming the delivery of ref (M4 Q10).
func priceIssue(prefix string, ref Ref, err error) (path, code, message string) {
	field := func(in calc.PriceInput) string {
		if in == calc.PriceInputTotal {
			return prefix + "totalLei"
		}
		return prefix + "pricePerKg"
	}
	var pe *calc.PriceError
	switch {
	case errors.Is(err, calc.ErrPriceBoth):
		return prefix + "pricePerKg", httpx.IssueOutOfRange, messages.PriceBothGiven(ref.Type, ref.Date)
	case errors.Is(err, calc.ErrPriceNeither):
		return prefix + "pricePerKg", httpx.IssueRequired, messages.InvalidData()
	case errors.As(err, &pe) && errors.Is(pe.Err, calc.ErrPriceNotPositive) && pe.Input == calc.PriceInputTotal:
		return field(pe.Input), httpx.IssueOutOfRange, messages.TotalLeiNotPositive(ref.Type, ref.Date)
	case errors.As(err, &pe) && errors.Is(pe.Err, calc.ErrPriceNotPositive):
		return field(pe.Input), httpx.IssueOutOfRange, messages.PricePerKgNotPositive(ref.Type, ref.Date)
	case errors.As(err, &pe):
		return field(pe.Input), httpx.IssueOutOfRange, messages.InvalidData()
	}
	return prefix + "id", httpx.IssueOutOfRange, messages.InvalidData() // ErrPriceKg: the column refuses kg ≤ 0
}

// UnpricedItem is one row of the S12 list (§7.13, M4 Q10).
type UnpricedItem struct {
	ID     uuid.UUID `json:"id"`
	Date   timex.Day `json:"date"`
	Farm   string    `json:"farm"`
	Type   string    `json:"type"`
	ForSex *string   `json:"forSex"` // null = ambele sexe
	Kg     num.Num   `json:"kg"`
}

// UnpricedResponse is GET /api/deliveries?priced=false (§7.13, M4 Q10): the
// unpriced deliveries oldest first, their count and Σ kg, and the farms they
// are at (by name).
type UnpricedResponse struct {
	Items []UnpricedItem  `json:"items"`
	Count int             `json:"count"`
	Kg    num.Num         `json:"kg"`
	Farms []string        `json:"farms"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// PricedDelivery is one delivery of a price batch's answer: both prices as
// stored (the typed one and the derived one).
type PricedDelivery struct {
	ID         uuid.UUID `json:"id"`
	PricePerKg num.Num   `json:"pricePerKg"`
	TotalLei   num.Num   `json:"totalLei"`
}

// SetResponse is the 200 of a price batch (§6.8, §7.13, M4 Q10).
type SetResponse struct {
	Deliveries []PricedDelivery `json:"deliveries"`
	Count      int              `json:"count"`
	Kg         num.Num          `json:"kg"`
	Toast      string           `json:"toast"`
	UndoToken  string           `json:"undoToken"` // valid until 23:59 of the save day (D4)
	AsOf       timex.LocalTime  `json:"asOf"`
}

// UndoResponse is the 200 of a price undo (M4 Q11): the deliveries back to
// "fără preț".
type UndoResponse struct {
	Deliveries []uuid.UUID     `json:"deliveries"`
	Toast      string          `json:"toast"`
	AsOf       timex.LocalTime `json:"asOf"`
}

// ChangedDelivery is the delivery of a price change's answer (M4 Q12).
type ChangedDelivery struct {
	ID         uuid.UUID       `json:"id"`
	PricePerKg num.Num         `json:"pricePerKg"`
	TotalLei   num.Num         `json:"totalLei"`
	PricedBy   string          `json:"pricedBy"` // the caller's name
	PricedAt   timex.LocalTime `json:"pricedAt"`
}

// ChangeResponse is the 200 of a price change (M4 Q12); no undo token (a
// change already needs a Motiv).
type ChangeResponse struct {
	Delivery ChangedDelivery `json:"delivery"`
	Toast    string          `json:"toast"`
	AsOf     timex.LocalTime `json:"asOf"`
}

// priceValues is one delivery's price in an audit row's old_values /
// new_values (M4 Q10): the price undo reads the batch's deliveries back from
// the price_set row's new_values.
type priceValues struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Date       timex.Day `json:"date"`
	Kg         num.Num   `json:"kg"`
	PricePerKg *num.Num  `json:"pricePerKg"`
	TotalLei   *num.Num  `json:"totalLei"`
	PriceInput *string   `json:"priceInput"`
}

// changeOldValues is a price change's old_values: the replaced price and who
// set it (nil pricedBy and pricedAt with priceFromRegister = "din registru").
type changeOldValues struct {
	priceValues
	PricedBy          *uuid.UUID       `json:"pricedBy"`
	PricedAt          *timex.LocalTime `json:"pricedAt"`
	PriceFromRegister bool             `json:"priceFromRegister"`
}

// auditChange is one old → new pair of a price audit row (§4.10 changes):
// key "Breeder 2 · 12.11.2026 · lei/kg", field pricePerKg or totalLei.
type auditChange struct {
	Key   string   `json:"key"`
	Field string   `json:"field"`
	Old   *num.Num `json:"old"`
	New   *num.Num `json:"new"`
}

// priceChanges are the lei/kg and total lei pairs of one delivery.
func priceChanges(typ string, date timex.Day, oldPerKg, newPerKg, oldTotal, newTotal *num.Num) []auditChange {
	return []auditChange{
		{Key: messages.AuditKeyPricePerKg(typ, date), Field: "pricePerKg", Old: oldPerKg, New: newPerKg},
		{Key: messages.AuditKeyTotalLei(typ, date), Field: "totalLei", Old: oldTotal, New: newTotal},
	}
}
