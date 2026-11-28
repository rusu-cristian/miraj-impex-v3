// Package audit writes the audit log (SPEC §4.10): one row per save action.
// Istoric (§7.10) and the setup undo (§7.19b) come in later milestones.
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Row is one audit_log row (§4.10), column for column. Changes, OldValues and
// NewValues are encoded as JSON; a nil (or JSON null) Changes is written as []
// and a nil OldValues/NewValues as SQL null. Nil RecordTypes and FarmIDs are written as
// empty arrays.
type Row struct {
	At              time.Time  // from the injected clock (M3 D3)
	UserID          *uuid.UUID // nil = system or migration
	Action          dbq.AuditAction
	Entity          string // entry, feed_delivery, cycle, series, farm, section, guide, feed_type, app_user
	EntityID        uuid.UUID
	EntryID         *uuid.UUID
	EntryKind       *dbq.EntryKind
	RecordTypes     []string // the S17 chips: hrana, mortalitate, …
	FarmIDs         []uuid.UUID
	RecordDate      *timex.Day
	Reason          *string // Motiv; nil for creates, same-day author edits and system rows
	SameDayAuthor   bool
	Summary         *string // the Jurnal name: "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg"
	Changes         any     // the old → new pairs Istoric shows; nil = []
	OldValues       any     // snapshot before; nil = null
	NewValues       any     // snapshot after; nil = null
	Source          dbq.RecordSource
	VoiceTranscript *string // only with a verified voiceToken (§8.9)
}

// Write inserts r. It runs inside the save's transaction (q bound to it), so
// the row exists exactly when the save commits. The app role has no UPDATE
// or DELETE on audit_log: a row is never changed afterwards.
func Write(ctx context.Context, q *dbq.Queries, r Row) error {
	p, err := params(r)
	if err != nil {
		return err
	}
	return q.InsertAuditLog(ctx, p)
}

// WriteID is Write returning the new row's id (the bigserial audit_log.id),
// for a save that names its audit row afterwards (the price undo token, M4 D4).
func WriteID(ctx context.Context, q *dbq.Queries, r Row) (int64, error) {
	p, err := params(r)
	if err != nil {
		return 0, err
	}
	// The conversion holds while InsertAuditLog and InsertAuditLogID keep
	// identical parameters (audit.sql says so on both).
	return q.InsertAuditLogID(ctx, dbq.InsertAuditLogIDParams(p))
}

// params is r as the insert's parameters, with the Row defaults applied.
func params(r Row) (dbq.InsertAuditLogParams, error) {
	changes, err := encode(r.Changes)
	if err != nil {
		return dbq.InsertAuditLogParams{}, fmt.Errorf("audit: changes: %w", err)
	}
	if changes == nil {
		changes = []byte("[]")
	}
	oldValues, err := encode(r.OldValues)
	if err != nil {
		return dbq.InsertAuditLogParams{}, fmt.Errorf("audit: old values: %w", err)
	}
	newValues, err := encode(r.NewValues)
	if err != nil {
		return dbq.InsertAuditLogParams{}, fmt.Errorf("audit: new values: %w", err)
	}
	if r.Source == "" {
		r.Source = dbq.RecordSourceManual
	}
	return dbq.InsertAuditLogParams{
		At:              r.At,
		UserID:          r.UserID,
		Action:          r.Action,
		Entity:          r.Entity,
		EntityID:        r.EntityID,
		EntryID:         r.EntryID,
		EntryKind:       r.EntryKind,
		RecordTypes:     nonNil(r.RecordTypes),
		FarmIds:         nonNil(r.FarmIDs),
		RecordDate:      r.RecordDate,
		Reason:          r.Reason,
		SameDayAuthor:   r.SameDayAuthor,
		Summary:         r.Summary,
		Changes:         changes,
		OldValues:       oldValues,
		NewValues:       newValues,
		Source:          r.Source,
		VoiceTranscript: r.VoiceTranscript,
	}, nil
}

// encode is v as JSON, or nil when v is nil or encodes as null (a nil slice,
// map or pointer).
func encode(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return nil, err
	}
	return b, nil
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
