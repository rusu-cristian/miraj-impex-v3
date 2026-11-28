package entries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ChangeSpec describes one change of a saved entry to the runner: an edit or
// Corectează (PUT), a delete, an undo or a restore (§7.8, plan M5-D2).
type ChangeSpec struct {
	EntryID uuid.UUID
	// Action is the change's audit action: update, delete, undo or restore.
	// Only restore reaches the plan for a deleted entry; the others get 404.
	Action dbq.AuditAction
	DryRun bool // run every write and check, answer the plan's Preview, roll back (Q7)
	// Meta carries the voice fields of a PUT (source, voiceTranscript,
	// voiceToken, §7.8); its ClientRequestID and DryRun are not read (a
	// change has no idempotency key, M5-D2: a retry meets the moved version).
	Meta httpx.WriteMeta
	// Allow decides, under the locks, whether the caller may make this change
	// (plan M5-D5): e is the locked entry (SeriesClosed set), a is
	// Actions(caller, e, today) computed as if the series were open, so a
	// change allowed by role and window but dated inside a closed series
	// answers 409 SERIES_CLOSED rather than 403. It returns the envelope
	// error (403 FORBIDDEN, …) or nil. Required: a spec without it is
	// refused before any read.
	Allow func(e EntryMeta, a EntryActions) error
	// OnConflict maps a unique violation of a child row (an added value, or a
	// restored row whose key another entry took meanwhile) after the
	// transaction, on the pool (M3 D2): 409 ALREADY_ENTERED + the holder.
	// c is what the plan returned for it (ARC-5). nil, or a nil return,
	// leaves the violation as it is (a 500).
	OnConflict func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError, c ChangeConflict) error
}

// ChangeConflict is what Runner.Change hands OnConflict after a unique
// violation (ARC-5): the entry as the plan saw it, and the rows the plan
// was un-deleting (Saved.Restored, which the plan returns with the failed
// write's error).
type ChangeConflict struct {
	Entry    LockedEntry
	Restored Snapshot
}

// ChangePlan is an action's part of a change, run inside the transaction
// after the farm locks, the row lock, Allow and the closed-series check: load
// the stored rows (ChangeTx.Stored), plan (level-2 refusals and warnings),
// write the child rows and the header (ChangeTx.UpdateHeader), run
// Tx.CheckCounts. A dry run ends with Tx.Preview; a change returns what the
// runner writes after it (audit row, live events, the 200 body).
type ChangePlan func(ctx context.Context, tx *ChangeTx) (Saved, error)

// LockedEntry is the entry row of a change, read under its row lock.
type LockedEntry struct {
	ID         uuid.UUID
	Kind       dbq.EntryKind
	FarmID     uuid.UUID
	PeerFarmID *uuid.UUID
	Date       timex.Day
	Source     dbq.RecordSource
	CreatedBy  *uuid.UUID
	CreatedAt  time.Time
	VersionAt  time.Time // coalesce(updated_at, created_at); UpdateHeader moves it
	Corrected  bool      // UpdateHeader sets it
	Deleted    bool
}

// FarmIDs is the entry's farm and, for a transfer, its peer farm.
func (e LockedEntry) FarmIDs() []uuid.UUID { return entryFarms(e.FarmID, e.PeerFarmID) }

// Ref is the entry as responses name it, at its current version.
func (e LockedEntry) Ref() httpx.EntryRef {
	return httpx.EntryRef{ID: e.ID, Kind: string(e.Kind), Version: version(e.VersionAt)}
}

// ChangeTx is a change's transaction as a ChangePlan sees it: the create's
// Tx (Principal, Today, Now, DryRun, Source, CheckCounts, Preview) and the
// locked entry.
type ChangeTx struct {
	*Tx
	Entry   LockedEntry
	Actions EntryActions // the caller's §3.2 actions on the entry
	Action  dbq.AuditAction

	adapter kindAdapter // the entry's kind's (ARC-1); the runner serves no other kind
}

// Stored loads the entry's live child rows as its create shape and snapshot
// (its kind's adapter).
func (t *ChangeTx) Stored(ctx context.Context) (StoredEntry, error) {
	e := t.Entry
	s := StoredEntry{ID: e.ID, Kind: e.Kind, FarmID: e.FarmID, Date: e.Date, Sections: []EntrySection{}}
	if err := t.adapter.load(ctx, t.Q, &s); err != nil {
		return StoredEntry{}, err
	}
	return s, nil
}

// UpdateHeader moves the entry's version (plan M5-D3) while it still is
// ifVersion: updated_at = max(Now, version + 1 µs), updated_by = the caller,
// corrected = the flag after the change. A moved version (another change
// committed since the client read it) is 409 STALE "Înregistrarea s-a
// schimbat între timp. …" (§7.0). It returns the new ref and keeps t.Entry
// in step.
func (t *ChangeTx) UpdateHeader(ctx context.Context, ifVersion time.Time, corrected bool) (httpx.EntryRef, error) {
	v, err := t.Q.UpdateEntryHeader(ctx, dbq.UpdateEntryHeaderParams{
		Now: t.Now, UpdatedBy: t.Principal.UserID, Corrected: corrected, ID: t.Entry.ID, IfVersion: ifVersion,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return httpx.EntryRef{}, httpx.Stale()
	}
	if err != nil {
		return httpx.EntryRef{}, fmt.Errorf("entries: update entry header: %w", err)
	}
	t.Entry.VersionAt, t.Entry.Corrected = v, corrected
	return t.Entry.Ref(), nil
}

// Latest is the entry's latest audit row, false when it has none (undo and
// restore act on the caller's own latest action, Q8).
func (t *ChangeTx) Latest(ctx context.Context) (LatestAction, bool, error) {
	row, err := t.Q.EntryLatestAudit(ctx, t.Entry.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return LatestAction{}, false, nil
	}
	if err != nil {
		return LatestAction{}, false, fmt.Errorf("entries: latest audit row: %w", err)
	}
	return LatestAction{
		ID: row.ID, Action: row.Action, UserID: row.UserID, At: row.At, Reason: row.Reason,
		SameDayAuthor: row.SameDayAuthor, Changes: row.Changes, OldValues: row.OldValues, NewValues: row.NewValues,
	}, true, nil
}

// LatestAction is an entry's latest audit row (EntryLatestAudit). The entry
// read fills Action, UserID, At, SameDayAuthor and OldValues.
type LatestAction struct {
	ID            int64
	Action        dbq.AuditAction
	UserID        *uuid.UUID
	At            time.Time
	Reason        *string
	SameDayAuthor bool   // the author's action in the creation-day window (no Motiv), SEC-2
	Changes       []byte // jsonb: []audit.Change
	OldValues     []byte // jsonb or nil: a Snapshot for M5's actions
	NewValues     []byte // jsonb or nil
}

// updateBefore is the entry as an update's undo puts it back (its audit
// old_values, Q8); false when the row holds no snapshot (the imported
// correction): that update can't be undone (409 STALE).
func (l LatestAction) updateBefore() (updateSnapshot, bool) {
	var before updateSnapshot
	if len(l.OldValues) == 0 || json.Unmarshal(l.OldValues, &before) != nil || snapshotRows(before.Snapshot) == 0 {
		return updateSnapshot{}, false
	}
	return before, true
}

// deletedRows is the rows a delete took (its audit old_values, D7), which a
// restore puts back; false when the row holds no snapshot (409 STALE).
func (l LatestAction) deletedRows() (removedRows, bool) {
	var removed removedRows
	if len(l.OldValues) == 0 || json.Unmarshal(l.OldValues, &removed) != nil || snapshotRows(removed.Snapshot) == 0 {
		return removedRows{}, false
	}
	return removed, true
}

// OwnToday reports whether p made this action on the local day today: the
// only actions p may undo or restore (Q8: "the caller's own latest action on
// the entry, the same local day").
func (l LatestAction) OwnToday(p auth.Principal, today timex.Day) bool {
	return l.UserID != nil && *l.UserID == p.UserID && timex.DayOf(l.At) == today
}

// ParseVersion reads an entry version (EntryRef.Version, PUT ifVersion):
// RFC 3339 with up to nanoseconds; the stored version has microseconds, so a
// version the server wrote parses back to the same instant (plan M5-D3).
func ParseVersion(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

// Change runs one change of entry c.EntryID for p (§12.5 "the same skeleton
// serves every write", plan M5-D2), in one READ COMMITTED transaction:
//
//  1. checkEntryHead (no lock: farm_id and peer_farm_id never change): no
//     entry → 404; a farm outside p's → 403 FARM_FORBIDDEN; a kind not
//     served (Q1) → 404; deleted (unless restoring) → 404;
//  2. lock the farms (ascending), then read the clock (M3 COR-3);
//  3. select the entry for update and re-check it (deleted meanwhile unless
//     restoring, or other farms → 409 STALE);
//  4. c.Allow with the caller's actions (§3.2); a closed series → 409
//     SERIES_CLOSED (§3.1);
//  5. the plan; a dry run ends with its Preview and rolls back every write;
//  6. the audit row (the runner sets At, UserID, Source, VoiceTranscript, and
//     the entry's Entity, EntityID, EntryID, EntryKind, FarmIDs, RecordDate
//     and c.Action where the plan left them empty), the live events, the
//     response encoded once (200), commit.
//
// After the transaction, a unique violation goes to c.OnConflict with the
// plan's ChangeConflict (M3 D2, ARC-5).
// Errors are envelope errors (httpx.HTTPError) or unexpected errors (500).
func (r *Runner) Change(ctx context.Context, p auth.Principal, c ChangeSpec, plan ChangePlan) (Result, error) {
	if c.Allow == nil {
		return Result{}, errors.New("entries: a change without Allow")
	}
	var (
		out      Result
		conflict ChangeConflict
	)
	err := r.store.InTx(ctx, func(dtx db.Tx) error {
		head, err := checkEntryHead(ctx, dtx.Q, p, c.EntryID, c.Action == dbq.AuditActionRestore)
		if err != nil {
			return err
		}
		farms := head.FarmIDs()

		if err := r.lock(ctx, dtx.Q, farms...); err != nil {
			return err
		}
		// The clock is read once the locks are held (M3 COR-3), as in Create.
		now := r.clock.Now()
		today := timex.DayOf(now)
		row, err := dtx.Q.EntryChangeLock(ctx, c.EntryID)
		if err != nil {
			return fmt.Errorf("entries: lock entry: %w", err) // entries are never removed, so a row exists
		}
		e := LockedEntry{
			ID: row.ID, Kind: row.Kind, FarmID: row.FarmID, PeerFarmID: row.PeerFarmID, Date: row.Date,
			Source: row.Source, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, VersionAt: row.VersionAt,
			Corrected: row.Corrected, Deleted: row.Deleted,
		}
		if e.Kind != head.Kind || e.FarmID != head.FarmID || !samePeer(e.PeerFarmID, head.PeerFarmID) {
			return httpx.Stale()
		}
		if e.Deleted && c.Action != dbq.AuditActionRestore {
			return httpx.Stale() // deleted while this change waited on the locks (COR-3): the entry moved
		}
		meta := EntryMeta{FarmIDs: farms, CreatedBy: e.CreatedBy, CreatedAt: e.CreatedAt}
		acts := Actions(p, meta, today)
		meta.SeriesClosed = row.ClosedSeries != ""
		if err := c.Allow(meta, acts); err != nil {
			return err
		}
		if row.ClosedSeries != "" {
			return httpx.SeriesClosed(row.ClosedSeries)
		}

		src, transcript := voiceSource(p, e.Kind, c.Meta)
		t := &ChangeTx{
			Tx: &Tx{
				Tx: dtx, Principal: p, Today: today, Now: now, DryRun: c.DryRun, Source: src,
				kind: e.Kind, out: &out, beforeCheck: r.beforeCheck,
			},
			Entry: e, Actions: acts, Action: c.Action, adapter: head.adapter,
		}
		saved, err := plan(ctx, t)
		conflict = ChangeConflict{Entry: t.Entry, Restored: saved.Restored}
		if err != nil {
			return err // db.ErrRollback after Preview: a dry run's answer is in out
		}
		if t.DryRun {
			return errors.New("entries: a dry run reached the change without Preview")
		}
		a := saved.Audit
		a.At, a.UserID, a.Source, a.VoiceTranscript = now, &p.UserID, src, transcript
		fillChangeAudit(&a, c.Action, t.Entry)
		if err := audit.Write(ctx, dtx.Q, a); err != nil {
			return err
		}
		if err := live.Emit(ctx, dtx.Q, saved.Events...); err != nil {
			return err
		}
		body, err := httpx.Marshal(saved.Response)
		if err != nil {
			return err
		}
		out = Result{Status: http.StatusOK, Body: body}
		return nil
	})
	if err == nil {
		return out, nil
	}
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); ok && c.OnConflict != nil {
		if mapped := c.OnConflict(ctx, r.store.Q(), ce, conflict); mapped != nil {
			return Result{}, mapped
		}
	}
	return Result{}, err
}

// entryHead is an entry read without a lock (EntryChangeHeader): its kind,
// farms, date and deleted flag, which a change and a PUT's pre-read check
// before anything else, and its kind's adapter.
type entryHead struct {
	ID         uuid.UUID
	Kind       dbq.EntryKind
	FarmID     uuid.UUID
	PeerFarmID *uuid.UUID
	Date       timex.Day
	Deleted    bool
	adapter    kindAdapter
}

// FarmIDs is the entry's farm and, for a transfer, its peer farm.
func (h entryHead) FarmIDs() []uuid.UUID { return entryFarms(h.FarmID, h.PeerFarmID) }

// checkEntryHead reads entry id for a change by p (ARC-4: PUT's pre-read
// and Runner.Change share it), in this order: no entry → 404; a farm of it
// outside p's farms → 403 FARM_FORBIDDEN; a kind not served (Q1) → 404;
// deleted → 404 unless deletedOK (a restore).
func checkEntryHead(ctx context.Context, q *dbq.Queries, p auth.Principal, id uuid.UUID, deletedOK bool) (entryHead, error) {
	row, err := q.EntryChangeHeader(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return entryHead{}, httpx.NotFound()
	}
	if err != nil {
		return entryHead{}, fmt.Errorf("entries: change header: %w", err)
	}
	h := entryHead{ID: id, Kind: row.Kind, FarmID: row.FarmID, PeerFarmID: row.PeerFarmID, Date: row.Date, Deleted: row.Deleted}
	for _, f := range h.FarmIDs() {
		if !p.CanSeeFarm(f) {
			return entryHead{}, httpx.FarmForbidden()
		}
	}
	a, err := adapterOf(h.Kind)
	if err != nil {
		return entryHead{}, httpx.NotFound()
	}
	if h.Deleted && !deletedOK {
		return entryHead{}, httpx.NotFound()
	}
	h.adapter = a
	return h, nil
}

// fillChangeAudit sets the entry's own columns of a change's audit row where
// the plan left them empty.
func fillChangeAudit(a *audit.Row, action dbq.AuditAction, e LockedEntry) {
	if a.Action == "" {
		a.Action = action
	}
	if a.Entity == "" {
		a.Entity = "entry"
	}
	if a.EntityID == uuid.Nil {
		a.EntityID = e.ID
	}
	if a.EntryID == nil {
		a.EntryID = &e.ID
	}
	if a.EntryKind == nil {
		a.EntryKind = &e.Kind
	}
	if a.FarmIDs == nil {
		a.FarmIDs = e.FarmIDs()
	}
	if a.RecordDate == nil {
		a.RecordDate = &e.Date
	}
}

func samePeer(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
