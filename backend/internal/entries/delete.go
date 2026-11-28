package entries

import (
	"context"
	"fmt"
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

// routeDeleteEntry deletes a whole save or one section's Zilnic values
// (§7.8, M5 Q9).
const routeDeleteEntry = "DELETE /api/entries/{id}"

// deleteRoutes mounts DELETE /api/entries/{id} behind authed (Routes calls
// it, plan M5-D1).
func (h *Handler) deleteRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeDeleteEntry, authed.ThenFunc(h.deleteEntry))
}

// deleteEntry is DELETE /api/entries/{id} (§7.8): the path id → decode
// (64 KiB) → Validate (level 1: the Motiv's length) → service. The answer is
// sent as the runner encoded it.
func (h *Handler) deleteEntry(w http.ResponseWriter, r *http.Request) {
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
	req, err := httpx.Decode[DeleteRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var v httpx.Validator
	req.Validate(&v)
	if err := v.Err(); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.DeleteEntry(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// DeleteEntry deletes entry id for p (§7.8, §6.15, plan M5-D5, D7, Q9) on
// the change runner: the whole save, or with req.SectionID one section's
// Zilnic values (a compartment → its mortality, cull and sexing-error rows;
// an egg unit → its eggs; D98). The author deletes on the creation day
// without a Motiv; a manager on own farms or the owner otherwise, with one
// (missing → 400 on `reason`, not asked of a dry run); anyone else gets 403.
// A dry run answers 200 CorrectionPreview with the Dialog sentence as its
// first effect. §6.1 is re-checked ("Ștergerea ar lăsa …" → 422).
func (s *Service) DeleteEntry(ctx context.Context, p auth.Principal, id uuid.UUID, req DeleteRequest) (Result, error) {
	reason := req.reason()
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionDelete, DryRun: req.DryRun,
		Allow: func(e EntryMeta, a EntryActions) error { return allowDelete(p, e, a, reason, req.DryRun) },
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		rm := entryRemoval{Action: dbq.AuditActionDelete, SectionID: req.SectionID}
		if !tx.Actions.Editable && reason != "" { // a Motiv only outside the author's window (D5)
			rm.Reason = &reason
		}
		return planRemoval(ctx, tx, rm)
	})
}

// allowDelete is who may delete (plan M5-D5, Q10): the author on the
// creation day (Editable) without a Motiv; a manager on own farms or the
// owner (Correctable) with one, unless it is a dry run; anyone else 403
// (changeForbidden).
func allowDelete(p auth.Principal, e EntryMeta, a EntryActions, reason string, dryRun bool) error {
	switch {
	case a.Editable:
		return nil
	case a.Correctable:
		if reason == "" && !dryRun {
			return httpx.Validation(httpx.Issue{Path: "reason", Code: httpx.IssueRequired, Message: messages.DeleteReasonMissing()})
		}
		return nil
	}
	return changeForbidden(p, e)
}

// changeForbidden is the 403 FORBIDDEN of a change the caller may not make
// (Q10): a vet's own entry after its creation day gets the S16 sentence
// "Poți modifica sau șterge ce ai introdus doar în aceeași zi, …"; any other
// (a vet on someone else's entry, a viewer) "Nu ai acces la această pagină.".
func changeForbidden(p auth.Principal, e EntryMeta) error {
	if p.Role == auth.Vet && e.CreatedBy != nil && *e.CreatedBy == p.UserID {
		return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.OwnEntryPastDay())
	}
	return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
}

// allowOwnAction is the access check of an undo or a restore (Q8): vet,
// manager and owner reach the plan, which checks that the entry's latest
// action is the caller's own of today and that the caller's current rights
// still allow it (checkOwnLatest, SEC-2); anyone else gets 403 FORBIDDEN.
func allowOwnAction(p auth.Principal) func(EntryMeta, EntryActions) error {
	return func(EntryMeta, EntryActions) error {
		if !p.Role.In(auth.Vet, auth.Manager, auth.Owner) {
			return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
		}
		return nil
	}
}

// checkOwnLatest is the rule of an undo or a restore (Q8, SEC-2) on the
// change's locked entry: ownLatestError with the caller's actions on it.
func checkOwnLatest(tx *ChangeTx, latest LatestAction, found bool, want dbq.AuditAction) error {
	meta := EntryMeta{FarmIDs: tx.Entry.FarmIDs(), CreatedBy: tx.Entry.CreatedBy, CreatedAt: tx.Entry.CreatedAt}
	return ownLatestError(tx.Principal, meta, tx.Actions, tx.Today, latest, found, want)
}

// ownLatestError is the rule of an undo or a restore (Q8, SEC-2), shared by
// POST …/undo and …/restore and the entry read's `undo` / `restore` flags:
// the entry's latest audit row is an action of kind want, made by p on the
// local day today, and p's current rights on the entry (a, §3.2) still
// allow that action (originalAllowed). Another action came later (or there
// is none) → 409 STALE; someone else's action → 403 FORBIDDEN; an action of
// an earlier day → 409 STALE "Nu se mai poate anula: ziua salvării s-a
// încheiat."; rights lost since (a same-day demotion) → 403 FORBIDDEN
// (changeForbidden).
func ownLatestError(p auth.Principal, meta EntryMeta, a EntryActions, today timex.Day, latest LatestAction, found bool, want dbq.AuditAction) error {
	switch {
	case !found || latest.Action != want:
		return httpx.Stale()
	case latest.UserID == nil || *latest.UserID != p.UserID:
		return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
	case !latest.OwnToday(p, today):
		return httpx.NewError(http.StatusConflict, httpx.CodeStale, messages.UndoDayEnded())
	case !originalAllowed(a, latest):
		return changeForbidden(p, meta)
	}
	return nil
}

// originalAllowed reports whether actions a still permit the action latest
// records (SEC-2): an author's action (a create, or an edit or delete by the
// author in the creation-day window: same_day_author) needs Editable; a
// correction or a delete with a Motiv needs Correctable.
func originalAllowed(a EntryActions, latest LatestAction) bool {
	if latest.Action == dbq.AuditActionCreate || latest.SameDayAuthor {
		return a.Editable
	}
	return a.Correctable
}

// UndoCreate undoes p's own create of entry id (Q8, §7.8 "undo the create
// (= delete, logged undo)"): the create must still be the entry's latest
// action and of today; the whole save is soft-deleted, the audit row is
// `undo` (same_day_author) and the live event entry.deleted. POST
// /api/entries/{id}/undo (undo.go) dispatches here, or calls undoCreate
// inside its own change, when the caller's latest action is a create.
func (s *Service) UndoCreate(ctx context.Context, p auth.Principal, id uuid.UUID) (Result, error) {
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionUndo, Allow: allowOwnAction(p),
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		latest, found, err := tx.Latest(ctx)
		if err != nil {
			return Saved{}, err
		}
		return undoCreate(ctx, tx, latest, found)
	})
}

// undoCreate is the undo of a create inside a change (tx.Action = undo):
// checkOwnLatest, then the whole save removed as by a delete (a change the
// entry's audit doesn't record since, as a delivery priced in Prețuri, →
// 409 STALE: planRemoval).
func undoCreate(ctx context.Context, tx *ChangeTx, latest LatestAction, found bool) (Saved, error) {
	if err := checkOwnLatest(tx, latest, found, dbq.AuditActionCreate); err != nil {
		return Saved{}, err
	}
	return planRemoval(ctx, tx, entryRemoval{Action: dbq.AuditActionUndo})
}

// entryRemoval is one delete or undo of a create.
type entryRemoval struct {
	Action    dbq.AuditAction // delete | undo
	Reason    *string         // the Motiv of a delete outside the author's window
	SectionID *uuid.UUID      // a per-section delete (daily entries only)
}

// planRemoval is the plan of a delete or an undo of a create (plan M5-D7):
//
//  1. the entry's live rows; with a section only that section's (the kind's
//     adapter; else 400); an undo of a create the adapter calls stale → 409;
//  2. the figures the messages name, read before the writes;
//  3. the rows soft-deleted by id, and the entry itself for a whole save (or
//     when no live row is left), its version moved;
//  4. §6.1 over the (section, sex) pairs whose counts go down, from the
//     entry date ("Ștergerea ar lăsa …");
//  5. a dry run answers its preview; else the Toast, the audit row (old_values
//     = the removed rows with their ids, which a restore reads back) and the
//     live events entry.deleted (+ counts.changed when birds moved).
func planRemoval(ctx context.Context, tx *ChangeTx, rm entryRemoval) (Saved, error) {
	e := tx.Entry
	stored, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	if rm.Action == dbq.AuditActionUndo && tx.adapter.createUndoStale(stored.Rows) {
		return Saved{}, httpx.Stale() // COR-2: e.g. priced since; the undo would drop the price
	}
	rows := stored.Rows
	if rm.SectionID != nil {
		if rows, err = tx.adapter.section(stored, *rm.SectionID); err != nil {
			return Saved{}, err
		}
	}
	if snapshotRows(rows) == 0 { // an entry without a live row: nothing a Dialog or a Toast could name
		return Saved{}, httpx.Stale()
	}
	whole := rm.SectionID == nil || snapshotRows(rows) == snapshotRows(stored.Rows)
	info, err := tx.Q.EntryDeleteContext(ctx, e.ID)
	if err != nil {
		return Saved{}, fmt.Errorf("entries: delete context: %w", err)
	}
	f, err := readRowFigures(ctx, tx, rows, -1)
	if err != nil {
		return Saved{}, err
	}

	if n, err := tx.Q.EntrySoftDeleteRows(ctx, dbq.EntrySoftDeleteRowsParams{
		Now: tx.Now, EntryID: e.ID, BirdIds: ids(rows.BirdEvents), EggIds: ids(rows.EggProduction),
		FeedIds: ids(rows.FeedAllocations), DeliveryIds: ids(rows.FeedDeliveries),
	}); err != nil {
		return Saved{}, fmt.Errorf("entries: soft-delete rows: %w", err)
	} else if int(n) != snapshotRows(rows) {
		return Saved{}, httpx.Stale() // a row changed under the lock: cannot happen, refuse rather than half-delete
	}
	if whole {
		n, err := tx.Q.EntryDeleteHeader(ctx, dbq.EntryDeleteHeaderParams{
			Now: tx.Now, DeletedBy: tx.Principal.UserID, DeleteReason: rm.Reason, ID: e.ID,
		})
		if err != nil {
			return Saved{}, fmt.Errorf("entries: delete entry: %w", err)
		}
		if n != 1 {
			return Saved{}, httpx.Stale()
		}
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, e.Corrected)
	if err != nil {
		return Saved{}, err
	}
	refusals, err := negativeRefusals(ctx, tx, f.birds, messages.ActionDelete)
	if err != nil {
		return Saved{}, err
	}
	stock, err := readStockAfter(ctx, tx, rows, f.feedKg, info.FarmName, true)
	if err != nil {
		return Saved{}, err
	}

	text := rowsText{
		Entry: tx.Entry, Rows: rows, SectionID: rm.SectionID, Undo: rm.Action == dbq.AuditActionUndo, Figures: f,
		Stock: stock, Info: info,
	}
	if tx.DryRun {
		return Saved{}, tx.Preview(newCorrectionPreview(CorrectionPreview{
			Effects: []string{tx.adapter.removalEffect(text)}, Alive: f.previewAlive(e.Date, tx.Today),
			CheckedDays: f.checkedDays(e.Date, tx.Today), Refusals: refusals, Warnings: stock.warnings,
		}))
	}
	if len(refusals) > 0 {
		return Saved{}, httpx.RefusedWith(refusals, stock.warnings)
	}

	summary := tx.adapter.summary(text)
	at := timex.At(tx.Now)
	events := []live.Event{live.EntryDeleted(live.EntryPayload{
		EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: f.sections,
		RecordTypes: f.recordTypes, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: e.Corrected,
	})}
	return Saved{
		Response: EntryWriteResponse{
			Entry: ref, Warnings: warningsOrEmpty(stock.warnings), Toast: tx.adapter.removalToast(text),
			Alive: f.farmAfter(), AsOf: at,
		},
		Audit: audit.Row{
			Action: rm.Action, Reason: rm.Reason, SameDayAuthor: rm.Action == dbq.AuditActionUndo || tx.Actions.Editable,
			RecordTypes: f.recordTypes, Summary: &summary, OldValues: removedRows{Snapshot: rows, SectionID: rm.SectionID},
		},
		Events: append(events, f.countsEvent(e)...),
	}, nil
}

// correctedBy names who made the entry's latest correction and on which
// day, for the delete Dialog ("… și corectată de Ion Rusu pe 18.11.2026"):
// the latest update not by the author in the creation-day window and not
// undone (EntryDeleteContext, COR-1), not whoever changed the entry last;
// nil when the entry is not corrected or no such update is recorded.
func correctedBy(e LockedEntry, info dbq.EntryDeleteContextRow) *messages.Corrected {
	if !e.Corrected || info.CorrectedByName == nil || info.CorrectedAt == nil {
		return nil
	}
	return &messages.Corrected{By: *info.CorrectedByName, On: timex.DayOf(*info.CorrectedAt)}
}

func authorOf(info dbq.EntryDeleteContextRow) string {
	if info.Author == nil {
		return ""
	}
	return *info.Author
}
