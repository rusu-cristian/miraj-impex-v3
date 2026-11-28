// Package live is the live-update outbox of SPEC §9: every write inserts its
// live_event rows and pg_notify in its own transaction (Emit), with typed
// payload constructors. The Hub, the LISTEN loop, the SSE handler and the
// purge come with M11.
package live

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Event types (§9).
const (
	TypeEntryCreated    = "entry.created"
	TypeDeliveryCreated = "delivery.created"
	TypePriceSet        = "price.set"
	TypePriceChanged    = "price.changed"
	TypeEntryUpdated    = "entry.updated"
	TypeEntryDeleted    = "entry.deleted"
	TypeEntryRestored   = "entry.restored"
	TypeCountsChanged   = "counts.changed"
)

// Event is one live_event row before it is written.
type Event struct {
	FarmIDs []uuid.UUID // the farms it concerns; empty = global (setup, users)
	Type    string      // the SSE event name, e.g. "entry.created"
	Payload any         // encoded as the event's JSON data
}

// Emit writes evs to the outbox in order: one live_event row and one
// pg_notify('live_event', id) per event, in the caller's transaction (q bound
// to it). PostgreSQL delivers the notifications at commit and drops them on
// rollback, so nothing is announced for a save that doesn't happen.
func Emit(ctx context.Context, q *dbq.Queries, evs ...Event) error {
	for _, ev := range evs {
		payload, err := json.Marshal(ev.Payload)
		if err != nil {
			return fmt.Errorf("live: encode %s: %w", ev.Type, err)
		}
		farms := ev.FarmIDs
		if farms == nil {
			farms = []uuid.UUID{}
		}
		if _, err := q.EmitLiveEvent(ctx, dbq.EmitLiveEventParams{FarmIds: farms, Type: ev.Type, Payload: payload}); err != nil {
			return err
		}
	}
	return nil
}

// EntryPayload is the data of entry.created / updated / deleted / restored
// (§9): ids and the minimum to place the change; clients refetch.
type EntryPayload struct {
	EntryID     uuid.UUID       `json:"entryId"`
	Kind        dbq.EntryKind   `json:"kind"`
	FarmIDs     []uuid.UUID     `json:"farmIds"`
	Dates       []timex.Day     `json:"dates"`
	Sections    []string        `json:"sections"`    // section codes: ["H6.1", "H6.2"]
	RecordTypes []string        `json:"recordTypes"` // the S17 chips: ["hrana"]
	Author      *string         `json:"author"`      // null for a system or migration entry
	At          timex.LocalTime `json:"at"`
	Summary     string          `json:"summary"` // the Jurnal name of the save
	Corrected   bool            `json:"corrected"`
}

// EntryCreated is the entry.created event of a new save, for every farm in
// p.FarmIDs. Nil lists in p are sent as [].
func EntryCreated(p EntryPayload) Event { return entryEvent(TypeEntryCreated, p) }

// EntryUpdated is the entry.updated event of an edit, a correction or an
// undone update (p.Corrected = the entry's flag after it), for every farm in
// p.FarmIDs. Nil lists in p are sent as [].
func EntryUpdated(p EntryPayload) Event { return entryEvent(TypeEntryUpdated, p) }

// EntryDeleted is the entry.deleted event of a delete (whole save or one
// section's values) or an undone create, for every farm in p.FarmIDs. Nil
// lists in p are sent as [].
func EntryDeleted(p EntryPayload) Event { return entryEvent(TypeEntryDeleted, p) }

// EntryRestored is the entry.restored event of a restore, for every farm in
// p.FarmIDs. Nil lists in p are sent as [].
func EntryRestored(p EntryPayload) Event { return entryEvent(TypeEntryRestored, p) }

func entryEvent(typ string, p EntryPayload) Event {
	p.FarmIDs = nonNil(p.FarmIDs)
	p.Dates = nonNil(p.Dates)
	p.Sections = nonNil(p.Sections)
	p.RecordTypes = nonNil(p.RecordTypes)
	return Event{FarmIDs: p.FarmIDs, Type: typ, Payload: p}
}

// CountsPayload is the data of counts.changed (§9): the farm whose bird counts
// changed, the earliest changed day, the sections changed (codes) and the
// farm's birds at the end of today after the write (the header counts swap,
// D197).
type CountsPayload struct {
	FarmID   uuid.UUID          `json:"farmId"`
	FromDate timex.Day          `json:"fromDate"`
	Sections []string           `json:"sections"` // section codes: ["H5.2"]
	Alive    httpx.SexPair[int] `json:"alive"`
}

// CountsChanged is the counts.changed event of a write that moved birds on
// farmID from fromDate (one event per farm), for that farm. Nil sections are
// sent as [].
func CountsChanged(farmID uuid.UUID, fromDate timex.Day, sections []string, alive httpx.SexPair[int]) Event {
	return Event{
		FarmIDs: []uuid.UUID{farmID},
		Type:    TypeCountsChanged,
		Payload: CountsPayload{FarmID: farmID, FromDate: fromDate, Sections: nonNil(sections), Alive: alive},
	}
}

// DeliveryPayload is the data of delivery.created (§9): the new delivery's
// entry, farm, type name, kg and date. It never carries a price.
type DeliveryPayload struct {
	EntryID  uuid.UUID `json:"entryId"`
	FarmID   uuid.UUID `json:"farmId"`
	FeedType string    `json:"feedType"` // the type's name: "Breeder 2"
	Kg       num.Num   `json:"kg"`
	Date     timex.Day `json:"date"`
}

// DeliveryCreated is the delivery.created event of a new delivery, for its farm.
func DeliveryCreated(p DeliveryPayload) Event {
	return Event{FarmIDs: []uuid.UUID{p.FarmID}, Type: TypeDeliveryCreated, Payload: p}
}

// PricePayload is the data of price.set / price.changed (§9): the deliveries
// whose price was set, changed or undone, their count and Σ kg. It carries no
// lei (M4 D2): the price stays out of the outbox until M11 filters by role.
type PricePayload struct {
	DeliveryIDs []uuid.UUID `json:"deliveryIds"`
	Count       int         `json:"count"`
	Kg          num.Num     `json:"kg"`
}

// PriceSet is the price.set event of a Prețuri save, for farmIDs (the
// deliveries' farms). Count is len(deliveryIDs); nil ids are sent as [].
func PriceSet(farmIDs, deliveryIDs []uuid.UUID, kg decimal.Decimal) Event {
	return priceEvent(TypePriceSet, farmIDs, deliveryIDs, kg)
}

// PriceChanged is the price.changed event of a price change or a price undo,
// for farmIDs (the deliveries' farms). Count is len(deliveryIDs).
func PriceChanged(farmIDs, deliveryIDs []uuid.UUID, kg decimal.Decimal) Event {
	return priceEvent(TypePriceChanged, farmIDs, deliveryIDs, kg)
}

func priceEvent(typ string, farmIDs, deliveryIDs []uuid.UUID, kg decimal.Decimal) Event {
	ids := nonNil(deliveryIDs)
	return Event{
		FarmIDs: nonNil(farmIDs),
		Type:    typ,
		Payload: PricePayload{DeliveryIDs: ids, Count: len(ids), Kg: num.Of(kg)},
	}
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
