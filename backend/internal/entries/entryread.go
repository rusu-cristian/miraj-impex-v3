package entries

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeEntryRead is the entry read (§7.8, M5 Q6).
const routeEntryRead = "GET /api/entries/{id}"

// entryReadRoutes mounts GET /api/entries/{id} behind authed (Routes calls
// it, plan M5-D1).
func (h *Handler) entryReadRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeEntryRead, authed.ThenFunc(h.readEntry))
}

// readEntry is GET /api/entries/{id} (§7.8, M5 Q6, Q10): the path id
// (malformed → 400 bad_type on `id`; no query parameter is accepted) →
// service.
func (h *Handler) readEntry(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.svc.ReadEntry(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseEntryPath reads the {id} of an entry route: a malformed id is a 400
// VALIDATION bad_type issue on `id`; a query parameter is an unknown_field
// issue (the entry routes take none).
func parseEntryPath(r *http.Request) (uuid.UUID, error) {
	var v httpx.Validator
	id, err := uuid.Parse(r.PathValue("id"))
	v.Check(err == nil, "id", httpx.IssueBadType, messages.InvalidData())
	for _, name := range slices.Sorted(maps.Keys(r.URL.Query())) {
		v.Add(name, httpx.IssueUnknownField, messages.InvalidData())
	}
	return id, v.Err()
}

// ReadEntry is entry id as its form for p (§7.8, M5 Q6, Q10), read in one
// snapshot (InReadTx, M5-D10) at the clock's instant.
func (s *Service) ReadEntry(ctx context.Context, p auth.Principal, id uuid.UUID) (EntryReadResponse, error) {
	now := s.clock.Now()
	var res EntryReadResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildEntryRead(ctx, tx, EntryReadInput{Principal: p, EntryID: id, Today: timex.DayOf(now), Now: now})
		return err
	})
	return res, err
}

// EntryReadInput is what BuildEntryRead reads: the caller, the entry, and
// the clock's day and instant (ARC-6: never app_today()).
type EntryReadInput struct {
	Principal auth.Principal
	EntryID   uuid.UUID
	Today     timex.Day
	Now       time.Time
}

// BuildEntryRead is the entry read on db (a read snapshot, or the import
// transaction): two statements for feed_daily and delivery, three for daily
// (the head, then the kind adapter's rows).
// Errors are envelope errors: 404 NOT_FOUND for an unknown entry; 403
// FARM_FORBIDDEN when one of its farms is not the caller's (any role that
// sees every farm of the entry may read it, a viewer too, Q10); 404 for a
// deleted entry or a kind not served yet (Q1).
func BuildEntryRead(ctx context.Context, db dbq.DBTX, in EntryReadInput) (EntryReadResponse, error) {
	head, err := dbq.New(db).EntryReadHead(ctx, in.EntryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return EntryReadResponse{}, httpx.NotFound()
	}
	if err != nil {
		return EntryReadResponse{}, fmt.Errorf("entries: entry read head: %w", err)
	}
	farms := entryFarms(head.FarmID, head.PeerFarmID)
	for _, f := range farms {
		if !in.Principal.CanSeeFarm(f) {
			return EntryReadResponse{}, httpx.FarmForbidden()
		}
	}
	kind, err := adapterOf(head.Kind)
	if head.Deleted || err != nil {
		return EntryReadResponse{}, httpx.NotFound()
	}
	stored := StoredEntry{ID: head.ID, Kind: head.Kind, FarmID: head.FarmID, Date: head.Date, Sections: []EntrySection{}}
	if err := kind.load(ctx, dbq.New(db), &stored); err != nil {
		return EntryReadResponse{}, err
	}

	h := EntryHead{
		ID: head.ID, Kind: head.Kind, Version: version(head.VersionAt),
		Farm: EntryFarm{ID: head.FarmID, Name: head.FarmName}, Date: head.Date,
		CreatedAt: timex.At(head.CreatedAt), Source: head.Source,
	}
	if head.PlacementDate != nil && head.AgeAtPlacementDays != nil {
		age := calc.SeriesAge{PlacementDate: *head.PlacementDate, AgeAtPlacementDays: int(*head.AgeAtPlacementDays)}
		sapt, days := calc.SAPT(age, head.Date), calc.AgeDays(age, head.Date)
		h.SAPT, h.AgeDays = &sapt, &days
	}
	if head.Author != nil {
		h.Author = &EntryAuthor{Name: *head.Author}
	}
	if head.Corrected && head.CorrectedAt != nil { // the latest correction (COR-1), not the latest change
		by := ""
		if head.CorrectedByName != nil {
			by = *head.CorrectedByName
		}
		h.Corrected = &EntryCorrected{By: by, At: timex.At(*head.CorrectedAt)}
	}

	meta := EntryMeta{FarmIDs: farms, CreatedBy: head.CreatedBy, CreatedAt: head.CreatedAt, SeriesClosed: head.ClosedSeries != ""}
	latest, found := LatestAction{}, head.LatestAction != nil && head.LatestAt != nil
	if found {
		latest = LatestAction{
			Action: *head.LatestAction, UserID: head.LatestUserID, At: *head.LatestAt,
			SameDayAuthor: head.LatestSameDayAuthor != nil && *head.LatestSameDayAuthor, OldValues: head.LatestOldValues,
		}
	}
	return EntryReadResponse{
		Entry: h, Values: kind.values(stored), Sections: stored.Sections, FeedTypes: stored.FeedTypes,
		Actions: readActions(in.Principal, meta, kind, stored, latest, found, in.Today), AsOf: timex.At(in.Now),
	}, nil
}

// readActions are p's actions on an entry (EntryReadActions): edit and
// correct from Actions (§3.2); undo and restore true exactly when POST
// …/undo and …/restore would act (criterion 11): the same rule on the
// entry's latest audit row (ownLatestError, with the caller's rights, SEC-2)
// and the same snapshot and kind checks (an update or a delete without
// old_values, as the imported correction, can't be undone or restored; a
// create the kind calls stale, as a delivery priced since, can't be
// undone). Inside a closed series Actions gives nothing, so all are false.
func readActions(p auth.Principal, meta EntryMeta, kind kindAdapter, stored StoredEntry, latest LatestAction, found bool, today timex.Day) EntryReadActions {
	a := Actions(p, meta, today)
	out := EntryReadActions{Edit: a.Editable, Correct: a.Correctable && !a.Editable, Delete: a.Editable || a.Correctable}
	if !found || ownLatestError(p, meta, a, today, latest, found, latest.Action) != nil {
		return out
	}
	switch latest.Action {
	case dbq.AuditActionCreate:
		out.Undo = !kind.createUndoStale(stored.Rows)
	case dbq.AuditActionUpdate:
		_, out.Undo = latest.updateBefore()
	case dbq.AuditActionDelete:
		_, out.Restore = latest.deletedRows()
	}
	return out
}

// entryFarms is an entry's farm and, for a transfer, its peer farm.
func entryFarms(farmID uuid.UUID, peer *uuid.UUID) []uuid.UUID {
	if peer == nil || *peer == farmID {
		return []uuid.UUID{farmID}
	}
	return []uuid.UUID{farmID, *peer}
}
