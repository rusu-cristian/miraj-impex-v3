package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeRestoreEntry restores the caller's own delete of today (§7.8, Q8).
const routeRestoreEntry = "POST /api/entries/{id}/restore"

// restoreRoutes mounts POST /api/entries/{id}/restore behind authed (Routes
// calls it, plan M5-D1).
func (h *Handler) restoreRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeRestoreEntry, authed.ThenFunc(h.restoreEntry))
}

// restoreEntry is POST /api/entries/{id}/restore {} (§7.8): the path id →
// decode (64 KiB, no member accepted) → service.
func (h *Handler) restoreEntry(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	id, err := parseEntryPath(r)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if _, err := httpx.Decode[RestoreRequest](w, r, httpx.MaxJSONBody); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.RestoreEntry(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// RestoreEntry undoes p's own delete of entry id (§7.8, Q8, plan M5-D7) on
// the change runner: the delete must still be the entry's latest action and
// of today (else 409 STALE, someone else's 403); exactly the rows its audit
// snapshot names come back (and the entry, after a whole-save delete), so
// rows an earlier edit superseded stay deleted. §6.1 is re-checked
// ("Restaurarea ar lăsa …" → 422); a row whose key another entry took
// meanwhile → 409 ALREADY_ENTERED naming that entry. Audit `restore`, live
// entry.restored (+ counts.changed when birds moved).
func (s *Service) RestoreEntry(ctx context.Context, p auth.Principal, id uuid.UUID) (Result, error) {
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionRestore, Allow: allowOwnAction(p), OnConflict: restoreHolder,
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		latest, found, err := tx.Latest(ctx)
		if err != nil {
			return Saved{}, err
		}
		if err := checkOwnLatest(tx, latest, found, dbq.AuditActionDelete); err != nil {
			return Saved{}, err
		}
		removed, ok := latest.deletedRows()
		if !ok {
			return Saved{}, httpx.Stale() // a delete without a snapshot (none written by M5's code) can't be restored
		}
		return planRestore(ctx, tx, removed)
	})
}

// planRestore is the restore's plan: the figures before, the entry and the
// snapshot's rows un-deleted, the version moved, §6.1 over the pairs whose
// counts go down, the stock warnings of restored feed rows (H5, saved
// anyway), the Toast, the audit row (new_values = the rows restored) and
// the live events.
func planRestore(ctx context.Context, tx *ChangeTx, removed removedRows) (Saved, error) {
	e := tx.Entry
	rows := removed.Snapshot
	info, err := tx.Q.EntryDeleteContext(ctx, e.ID)
	if err != nil {
		return Saved{}, fmt.Errorf("entries: restore context: %w", err)
	}
	f, err := readRowFigures(ctx, tx, rows, +1)
	if err != nil {
		return Saved{}, err
	}
	if e.Deleted {
		n, err := tx.Q.EntryRestoreHeader(ctx, e.ID)
		if err != nil {
			return Saved{}, fmt.Errorf("entries: restore entry: %w", err)
		}
		if n != 1 {
			return Saved{}, httpx.Stale()
		}
	}
	n, err := tx.Q.EntryRestoreRows(ctx, dbq.EntryRestoreRowsParams{
		EntryID: e.ID, BirdIds: ids(rows.BirdEvents), EggIds: ids(rows.EggProduction), FeedIds: ids(rows.FeedAllocations),
		DeliveryIds: ids(rows.FeedDeliveries),
	})
	if err != nil {
		return Saved{Restored: rows}, err // a 23505 goes to OnConflict after the transaction (ARC-5)
	}
	if int(n) != snapshotRows(rows) {
		return Saved{}, httpx.Stale() // a row of the snapshot is live already: the delete is not what it was
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, e.Corrected)
	if err != nil {
		return Saved{}, err
	}
	refusals, err := negativeRefusals(ctx, tx, f.birds, messages.ActionRestore)
	if err != nil {
		return Saved{}, err
	}
	stock, err := readStockAfter(ctx, tx, rows, f.feedKg, info.FarmName, false)
	if err != nil {
		return Saved{}, err
	}
	if len(refusals) > 0 {
		return Saved{}, httpx.RefusedWith(refusals, stock.warnings)
	}

	text := rowsText{Entry: e, Rows: rows, SectionID: removed.SectionID, Figures: f, Stock: stock, Info: info}
	summary := tx.adapter.summary(text)
	at := timex.At(tx.Now)
	events := []live.Event{live.EntryRestored(live.EntryPayload{
		EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: f.sections,
		RecordTypes: f.recordTypes, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: e.Corrected,
	})}
	return Saved{
		Response: EntryWriteResponse{
			Entry: ref, Warnings: warningsOrEmpty(stock.warnings), Toast: tx.adapter.restoreToast(text),
			Alive: f.farmAfter(), AsOf: at,
		},
		Audit: audit.Row{
			SameDayAuthor: true, RecordTypes: f.recordTypes, Summary: &summary, NewValues: removed,
		},
		Events: append(events, f.countsEvent(e)...),
	}, nil
}

// restoreHolder is the OnConflict of an undo and a restore (M3 D2, ARC-5):
// after the transaction, 409 ALREADY_ENTERED naming the first row of
// c.Restored whose key another entry holds now, with that entry and the
// kind's Z9 / Hrană sentence (kindAdapter.holderMessage; EntryAlreadySaved
// for an imported holder). nil when the plan was un-deleting no row or none
// collides any more (the violation stays a 500).
func restoreHolder(ctx context.Context, q *dbq.Queries, _ *pgerr.ConstraintError, c ChangeConflict) error {
	rows := c.Restored
	if snapshotRows(rows) == 0 {
		return nil
	}
	a, err := adapterOf(c.Entry.Kind)
	if err != nil {
		return nil
	}
	h, err := q.EntryRestoreHolder(ctx, dbq.EntryRestoreHolderParams{
		EntryID: c.Entry.ID, BirdIds: ids(rows.BirdEvents), EggIds: ids(rows.EggProduction), FeedIds: ids(rows.FeedAllocations),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("entries: restore holder: %w", err)
	}
	if h.CreatedBy == nil || h.Source == dbq.RecordSourceMigration {
		return httpx.AlreadyEntered(h.HolderEntryID, messages.EntryAlreadySaved())
	}
	return httpx.AlreadyEntered(h.HolderEntryID, a.holderMessage(h, c.Entry.Date))
}
