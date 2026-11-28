package entries

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeUndoEntry undoes the caller's own latest action on an entry, made
// today (§7.8, M5 Q8).
const routeUndoEntry = "POST /api/entries/{id}/undo"

// undoRoutes mounts POST /api/entries/{id}/undo behind authed (Routes calls
// it, plan M5-D1).
func (h *Handler) undoRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeUndoEntry, authed.ThenFunc(h.undoEntry))
}

// undoEntry is POST /api/entries/{id}/undo {} (§7.8): the path id → decode
// (64 KiB, no member accepted) → service.
func (h *Handler) undoEntry(w http.ResponseWriter, r *http.Request) {
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
	if _, err := httpx.Decode[UndoRequest](w, r, httpx.MaxJSONBody); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.UndoEntry(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// UndoEntry undoes p's own latest action on entry id (§7.8, Q8) on the change
// runner, dispatched on the entry's latest audit row:
//   - a create → the whole save soft-deleted (undoCreate, delete.go);
//   - an update (an edit or a correction) → the child rows put back as that
//     row's old_values hold them (rows it changed back in place, rows it
//     soft-deleted un-deleted, rows it inserted soft-deleted) and `corrected`
//     back to its flag before; audit `undo` without a Motiv (the reason stays
//     on the update row), live entry.updated (+ counts.changed when birds
//     moved). §6.1 is re-checked ("Ștergerea ar lăsa …" → 422); a key another
//     entry took meanwhile → 409 ALREADY_ENTERED. A feed save's rows come
//     back the same way, a delivery's type, kg and note in place (its typed
//     price kept, planUndoDeliveryUpdate); both with the H5 warnings.
//
// The action must still be the entry's latest and made today by p:
// another action came later → 409 STALE; someone else's → 403 FORBIDDEN; of
// an earlier day → 409 "Nu se mai poate anula: ziua salvării s-a încheiat.".
// A viewer gets 403 FORBIDDEN.
func (s *Service) UndoEntry(ctx context.Context, p auth.Principal, id uuid.UUID) (Result, error) {
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionUndo, Allow: allowOwnAction(p), OnConflict: restoreHolder,
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		latest, found, err := tx.Latest(ctx)
		if err != nil {
			return Saved{}, err
		}
		if found && latest.Action == dbq.AuditActionCreate {
			return undoCreate(ctx, tx, latest, found)
		}
		if err := checkOwnLatest(tx, latest, found, dbq.AuditActionUpdate); err != nil {
			return Saved{}, err
		}
		before, ok := latest.updateBefore()
		if !ok {
			return Saved{}, httpx.Stale() // an update without a snapshot (the imported correction) can't be undone
		}
		return tx.adapter.undoUpdate(ctx, tx, before)
	})
}

// planUndoUpdate puts a Zilnic entry back as before (the update's
// old_values): by row id, a row of the snapshot that is live now with other
// figures is changed back in place, one that is not live is un-deleted, and
// a live row the snapshot doesn't hold (the update inserted it) is
// soft-deleted. The un-deleted rows go back in Saved.Restored with a failed
// write's error (OnConflict, ARC-5).
func planUndoUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	e := tx.Entry
	now, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	w := undoWrites(now.Rows, before.Snapshot)
	codes := map[uuid.UUID]string{}
	for _, s := range [][]SnapshotBird{now.Rows.BirdEvents, before.BirdEvents} {
		for _, b := range s {
			codes[b.SectionID] = b.Section
		}
	}
	for _, s := range [][]SnapshotEggs{now.Rows.EggProduction, before.EggProduction} {
		for _, g := range s {
			codes[g.SectionID] = g.Section
		}
	}
	target := StoredEntry{Kind: e.Kind, FarmID: e.FarmID, Date: e.Date, Rows: before.Snapshot}
	diffs := diffDaily(now.Daily(), target.Daily(), codes)

	f := rowFigures{birds: moveOf(now.Rows.BirdEvents, before.BirdEvents)}
	if err := readAlive(ctx, tx, &f, w.birds()); err != nil {
		return Saved{}, err
	}
	if err := w.apply(ctx, tx); err != nil {
		return Saved{Restored: w.restore}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, before.Corrected)
	if err != nil {
		return Saved{}, err
	}
	refusals, err := negativeRefusals(ctx, tx, f.birds, messages.ActionDelete)
	if err != nil {
		return Saved{}, err
	}
	if len(refusals) > 0 {
		return Saved{}, httpx.RefusedWith(refusals, nil)
	}

	toast := messages.CorrectionsUndone(len(diffs))
	if len(diffs) == 1 {
		d := diffs[0]
		toast = messages.CorrectionUndone(diffLabel(d), shownValue(d.Change.Old, d.Change.Field), shownValue(d.Change.New, d.Change.Field))
	}
	summary := messages.DailySummary(len(rowSections(before.BirdEvents)), len(before.EggProduction))
	types := diffChips(diffs)
	at := timex.At(tx.Now)
	events := []live.Event{live.EntryUpdated(live.EntryPayload{
		EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: diffSections(diffs),
		RecordTypes: types, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: before.Corrected,
	})}
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: []httpx.Warning{}, Toast: toast, Alive: f.farmAfter(), AsOf: at},
		Audit: audit.Row{
			SameDayAuthor: true, RecordTypes: types, Summary: &summary, Changes: changesOf(diffs),
			OldValues: updateSnapshot{Snapshot: now.Rows, Corrected: e.Corrected}, NewValues: before,
		},
		Events: append(events, f.countsEvent(e)...),
	}, nil
}

// undoRows are the writes of an undone update.
type undoRows struct {
	rowWrites
	restore Snapshot // rows to un-delete (the update soft-deleted them)
}

// undoWrites matches the live rows now to the rows before by id (bird and
// egg rows, the Zilnic kind; diffByID).
func undoWrites(now, before Snapshot) undoRows {
	var w undoRows
	birds := diffByID(now.BirdEvents, before.BirdEvents, func(cur, b SnapshotBird) bool {
		return cur.Qty == b.Qty && sameReason(cur.CullReason, b.CullReason)
	})
	w.restore.BirdEvents, w.deleteBirds = birds.restore, birds.remove
	for _, b := range birds.update {
		w.updateBirds = append(w.updateBirds, birdUpdate{ID: b.ID, Qty: b.Qty, CullReason: b.CullReason})
	}
	eggs := diffByID(now.EggProduction, before.EggProduction, sameEggs)
	w.restore.EggProduction, w.deleteEggs = eggs.restore, eggs.remove
	for _, g := range eggs.update {
		w.updateEggs = append(w.updateEggs, eggUpdateOf(g.ID, g))
	}
	return w
}

func (w undoRows) birds() bool { return w.rowWrites.birds() || len(w.restore.BirdEvents) > 0 }

// apply runs the soft deletes first (a key the update inserted is free again
// before a row of the same key comes back), then the un-deletes (a key
// another entry took meanwhile → 23505, mapped by OnConflict), then the
// changes in place.
func (w undoRows) apply(ctx context.Context, tx *ChangeTx) error {
	if err := (rowWrites{deleteBirds: w.deleteBirds, deleteEggs: w.deleteEggs}).apply(ctx, tx); err != nil {
		return err
	}
	if n := snapshotRows(w.restore); n > 0 {
		got, err := tx.Q.EntryRestoreRows(ctx, dbq.EntryRestoreRowsParams{
			EntryID: tx.Entry.ID, BirdIds: ids(w.restore.BirdEvents), EggIds: ids(w.restore.EggProduction), FeedIds: []uuid.UUID{},
			DeliveryIds: []uuid.UUID{},
		})
		if err != nil {
			return err
		}
		if int(got) != n {
			return httpx.Stale()
		}
	}
	return (rowWrites{updateBirds: w.updateBirds, updateEggs: w.updateEggs}).apply(ctx, tx)
}
