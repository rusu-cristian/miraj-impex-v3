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

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Event types (§9).
const (
	TypeEntryCreated = "entry.created"
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
func EntryCreated(p EntryPayload) Event {
	p.FarmIDs = nonNil(p.FarmIDs)
	p.Dates = nonNil(p.Dates)
	p.Sections = nonNil(p.Sections)
	p.RecordTypes = nonNil(p.RecordTypes)
	return Event{FarmIDs: p.FarmIDs, Type: TypeEntryCreated, Payload: p}
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
