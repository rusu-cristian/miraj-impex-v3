package entries

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Hrană zilnică adapter of PUT /api/entries/{id} and of the undo of its
// update (§7.8, §6.3, §6.15; plan M5-D5, D6, Q8, Q11, Q12): the edit or
// Corectează of a whole feed save, its rows matched by (section, sex).

// FeedDailyUpdateRequest is a PUT of a Hrană zilnică entry.
type FeedDailyUpdateRequest = EntryUpdateRequest[FeedDailyValues]

// updateFeedDaily decodes and checks a Hrană zilnică PUT (level 1: the body
// as FeedDailyUpdateRequest with every member known, then
// validateFeedUpdate on the stored farm's sections and the named types) and
// runs it.
func (h *Handler) updateFeedDaily(w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	req, err := httpx.Decode[FeedDailyUpdateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		return Result{}, err
	}
	refs, err := h.svc.FeedRefs(r.Context(), FeedDailyCreateRequest{FarmID: t.FarmID, Rows: req.Values.Rows})
	if err != nil {
		return Result{}, err
	}
	var v httpx.Validator
	validateFeedUpdate(&v, req, t, refs)
	if err := v.Err(); err != nil {
		return Result{}, err
	}
	return h.svc.UpdateFeedDaily(r.Context(), p, t.ID, req)
}

// validateFeedUpdate is the level-1 check of a Hrană zilnică PUT (§7.0b):
// validateHead, the values as a feed create checks them (H1, types and
// sections known, each (section, sex) once) with every path under
// "values.", validateTail.
func validateFeedUpdate(v *httpx.Validator, r FeedDailyUpdateRequest, t updateTarget, refs FeedRefs) {
	r.validateHead(v, t, r.Values.FarmID, r.Values.Date)
	var values httpx.Validator
	FeedDailyCreateRequest{FarmID: r.Values.FarmID, Date: r.Values.Date, Rows: r.Values.Rows}.Validate(&values, refs)
	addIssuesUnder(v, "values.", values)
	r.validateTail(v)
}

// addIssuesUnder adds every issue of from to v with its path under prefix
// (a create's level-1 check run on PUT `values`).
func addIssuesUnder(v *httpx.Validator, prefix string, from httpx.Validator) {
	for _, is := range from.Issues() {
		v.Add(prefix+is.Path, is.Code, is.Message)
	}
}

// UpdateFeedDaily edits or corrects Hrană zilnică entry id for p (§7.8,
// §6.3, §6.15, plan M5-D5, D6, Q11) on the change runner; req must have
// passed validateFeedUpdate. The server decides edit (the author on the
// creation day: no Motiv, audit `update` same_day_author, the create's
// Toast) or correction (a manager on own farms, the owner: Motiv required,
// `corrected` set, "Corecție salvată · …"); anyone else 403. Under the locks
// (planFeedUpdate):
//   - values name the stored farm and date (400), ifVersion is the version
//     (409 STALE);
//   - the rows matched by (section, sex) with the stored ones: a g or type
//     changed in place, a row gone soft-deleted, a new row inserted; no
//     change → 422 "Nu ai schimbat nicio cifră.";
//   - H3 and H2 on the changed and added rows only, a key another entry
//     holds → 409 ALREADY_ENTERED (D6); H4 on every row of the new save;
//   - after the writes, H5 over [date, today] per type of the new save
//     (warnings, saved anyway, D40).
//
// A dry run answers 200 CorrectionPreview (Q7) and rolls back.
func (s *Service) UpdateFeedDaily(ctx context.Context, p auth.Principal, id uuid.UUID, req FeedDailyUpdateRequest) (Result, error) {
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionUpdate, DryRun: req.DryRun, Meta: req.meta(), Allow: allowUpdate(p),
		OnConflict: func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError, c ChangeConflict) error {
			if ce.Constraint != constraintFeedAllocation {
				return nil
			}
			return feedHolderFor(ctx, q, req.Values, c.Entry.ID)
		},
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		return planFeedUpdate(ctx, tx, req)
	})
}

// feedHolderFor answers a feed_allocation_uq violation of a change of entry
// self after the transaction (M3 D2): 409 ALREADY_ENTERED naming the first
// row of v whose key another entry holds now (self's own keys are not
// collisions); nil when none is held any more.
func feedHolderFor(ctx context.Context, q *dbq.Queries, v FeedDailyValues, self uuid.UUID) error {
	slots, err := loadFeedSlots(ctx, q, v.FarmID, v.Date)
	if err != nil {
		return err
	}
	for _, row := range v.Rows {
		slot, ok := slots[slotKey{row.SectionID, row.Sex}]
		if !ok || (slot.HolderEntryID != nil && *slot.HolderEntryID == self) {
			continue
		}
		if c, held := collisionOf(slot, v.Date); held {
			return httpx.AlreadyEntered(c.EntryID, c.Message)
		}
	}
	return nil
}

// planFeedUpdate is the Hrană zilnică adapter's plan of a PUT
// (UpdateFeedDaily).
func planFeedUpdate(ctx context.Context, tx *ChangeTx, req FeedDailyUpdateRequest) (Saved, error) {
	e := tx.Entry
	v := req.Values
	if e.Kind != dbq.EntryKindFeedDaily {
		return Saved{}, httpx.NotFound()
	}
	if err := checkUpdateHead(tx, v.FarmID, v.Date, req.IfVersion); err != nil {
		return Saved{}, err
	}
	if len(v.Rows) == 0 { // validateFeedUpdate guarantees a row with g on each
		return Saved{}, httpx.Validation(httpx.Issue{Path: "values.rows", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	for i, row := range v.Rows {
		if row.G == nil {
			return Saved{}, httpx.Validation(httpx.Issue{
				Path: "values.rows[" + strconv.Itoa(i) + "].g", Code: httpx.IssueRequired, Message: messages.InvalidData(),
			})
		}
	}
	correction := !tx.Actions.Editable
	stored, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	old := stored.Rows.FeedAllocations
	// The context names every type of the new and the stored rows.
	named := slices.Clone(v.Rows)
	for _, r := range old {
		named = append(named, FeedDailyRow{FeedTypeID: r.FeedTypeID})
	}
	fc, ok, err := loadFeedContext(ctx, tx.Q, FeedDailyCreateRequest{FarmID: e.FarmID, Date: e.Date, Rows: named})
	if err != nil {
		return Saved{}, err
	}
	if !ok {
		return Saved{}, httpx.NotFound()
	}
	slots, err := loadFeedSlots(ctx, tx.Q, e.FarmID, e.Date)
	if err != nil {
		return Saved{}, err
	}

	// The rows matched by (section, sex) with the stored ones (D6).
	byKey := make(map[slotKey]SnapshotFeed, len(old))
	for _, r := range old {
		byKey[slotKey{r.SectionID, r.Sex}] = r
	}
	newRows := make([]feedRowValue, len(v.Rows))
	var w feedWrites
	touched := make([]bool, len(v.Rows)) // changed or added: the level-2 rules apply (D6)
	typeChanged := make([]bool, len(v.Rows))
	for i, row := range v.Rows {
		k := slotKey{row.SectionID, row.Sex}
		newRows[i] = feedRowValue{
			SectionID: row.SectionID, Section: slots[k].Code, Sex: row.Sex, TypeID: row.FeedTypeID,
			Type: fc.Types[row.FeedTypeID].Name, G: row.G,
		}
		s, had := byKey[k]
		switch {
		case !had:
			w.inserts = append(w.inserts, feedInsert(row))
			touched[i], typeChanged[i] = true, true
		case s.FeedTypeID != row.FeedTypeID || !sameNum(s.G, row.G):
			u := feedUpdate{ID: s.ID, FeedTypeID: row.FeedTypeID, G: row.G}
			if sameNum(s.G, row.G) { // a type change alone keeps an imported row's kg
				u.KgRecorded = s.KgRecorded
			}
			w.updates = append(w.updates, u)
			touched[i], typeChanged[i] = true, s.FeedTypeID != row.FeedTypeID
		}
	}
	seen := make(map[slotKey]bool, len(v.Rows))
	for _, row := range v.Rows {
		seen[slotKey{row.SectionID, row.Sex}] = true
	}
	for _, r := range old {
		if !seen[slotKey{r.SectionID, r.Sex}] {
			w.deletes = append(w.deletes, r.ID)
		}
	}
	diffs := diffFeed(feedRowValues(old), newRows)
	changes := fieldChanges(diffs)

	preview := func(p CorrectionPreview) (Saved, error) {
		p.Changes = changes
		return Saved{}, tx.Preview(newCorrectionPreview(p))
	}
	if len(diffs) == 0 { // §6.15
		msg := messages.NothingChanged()
		refusals := []httpx.Refusal{{Field: "values", Message: msg}}
		if tx.DryRun {
			return preview(CorrectionPreview{Refusals: refusals})
		}
		return Saved{}, httpx.Refused(msg, refusals, []httpx.Warning{})
	}
	reason := req.reason()
	if correction && reason == "" && !tx.DryRun { // §6.15: the sentence names the change
		return Saved{}, httpx.Validation(httpx.Issue{Path: "reason", Code: httpx.IssueRequired, Message: fieldsReasonMissing(diffs)})
	}

	var (
		refusals  []httpx.Refusal
		warnings  = []httpx.Warning{}
		collision *feedCollision
		kgs       = make([]decimal.Decimal, len(v.Rows))
	)
	for i, row := range v.Rows {
		path := "values.rows[" + strconv.Itoa(i) + "]"
		k := slotKey{row.SectionID, row.Sex}
		slot, known := slots[k]
		if touched[i] {
			if !known || !fed(slot) { // H3
				refusals = append(refusals, httpx.Refusal{
					Field: path + ".sectionId", SectionCode: slot.Code, Sex: string(row.Sex),
					Message: messages.NoBirdsAtDayStart(slot.Code, row.Sex, e.Date),
				})
			}
			if typeChanged[i] { // H2
				if r, bad := feedTypeRefusal(fc, row, path, slot.Code); bad {
					refusals = append(refusals, r)
				}
			}
			if slot.HolderEntryID != nil && *slot.HolderEntryID != e.ID {
				if c, held := collisionOf(slot, e.Date); known && held {
					refusals = append(refusals, httpx.Refusal{Field: path, SectionCode: slot.Code, Sex: string(row.Sex), Message: c.Message})
					if collision == nil {
						collision = &c
					}
				}
			}
		}
		g := row.G.Decimal()
		kgs[i] = calc.RowFeedKg(g, slot.AliveStart)
		if s, had := byKey[k]; had && !touched[i] && s.KgRecorded != nil { // an imported row left as it is keeps its kg
			kgs[i] = s.KgRecorded.Decimal()
		}
		if known && fed(slot) { // H4 on every row of the new save (D6, D173)
			if warn, ok := feedGuideWarning(fc, slot.Code, row.Sex, g); ok {
				warnings = append(warnings, warn)
			}
		}
	}
	if collision != nil && !tx.DryRun {
		return Saved{}, httpx.AlreadyEntered(collision.EntryID, collision.Message)
	}
	if len(refusals) > 0 {
		if tx.DryRun {
			return preview(CorrectionPreview{Refusals: refusals, Warnings: warnings})
		}
		return Saved{}, httpx.RefusedWith(refusals, warnings)
	}

	if err := w.apply(ctx, tx); err != nil {
		return Saved{}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, e.Corrected || correction)
	if err != nil {
		return Saved{}, err
	}

	// H5 and the stock after the change, per type of the new save, over
	// [date, today] (as the create, Q5).
	kgByType := map[uuid.UUID]decimal.Decimal{}
	kgTotal := decimal.Zero
	for i, row := range v.Rows {
		kgByType[row.FeedTypeID] = kgByType[row.FeedTypeID].Add(kgs[i])
		kgTotal = kgTotal.Add(kgs[i])
	}
	types := slices.DeleteFunc(slices.Clone(fc.TypeOrder), func(id uuid.UUID) bool {
		_, ok := kgByType[id]
		return !ok
	})
	series, err := calc.StockSeries(ctx, tx.Tx, e.FarmID, types, e.Date, tx.Today)
	if err != nil {
		return Saved{}, err
	}
	toastType, toastStock := uuid.Nil, decimal.Zero
	for _, id := range types {
		pts := series[id]
		if low, ok := calc.Lowest(pts); ok && low.Kg.IsNegative() {
			warnings = append(warnings, httpx.Warning{
				Code: httpx.WarnStockNegative, Message: messages.StockWouldGoNegative(fc.Types[id].Name, fc.FarmName, low.Kg),
			})
		}
		if toastType == uuid.Nil || kgByType[id].GreaterThan(kgByType[toastType]) {
			toastType, toastStock = id, decimal.Zero
			if len(pts) > 0 {
				toastStock = pts[len(pts)-1].Kg
			}
		}
	}
	if tx.DryRun {
		return preview(CorrectionPreview{Warnings: warnings})
	}

	of := 0 // the section × sex rows with birds that day
	for _, s := range slots {
		if fed(s) {
			of++
		}
	}
	summary := messages.FeedDailySummary(len(v.Rows), of, kgTotal)
	toast := fieldsCorrected(diffs)
	if !correction { // an author edit: the form's create Toast (Q15)
		toast = messages.FeedDailySaved(len(v.Rows), kgTotal, fc.Types[toastType].Name, toastStock)
	}
	snapshot := make([]feedAuditRow, len(v.Rows))
	for i, row := range v.Rows {
		snapshot[i] = feedAuditRow{Section: newRows[i].Section, Sex: row.Sex, FeedTypeID: row.FeedTypeID, G: *row.G, Kg: num.Of(kgs[i])}
	}
	corrected := e.Corrected || correction
	at := timex.At(tx.Now)
	var motiv *string
	if correction {
		motiv = &reason
	}
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: warnings, Toast: toast, AsOf: at},
		Audit: audit.Row{
			Reason: motiv, SameDayAuthor: !correction, RecordTypes: []string{recordTypeFeed}, Summary: &summary,
			Changes: changes, OldValues: updateSnapshot{Snapshot: stored.Rows, Corrected: e.Corrected}, NewValues: snapshot,
		},
		Events: []live.Event{live.EntryUpdated(live.EntryPayload{
			EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: fieldSections(diffs),
			RecordTypes: []string{recordTypeFeed}, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: corrected,
		})},
	}, nil
}

// feedTypeRefusal is H2 for a row's type (§6.3, M3 Q4): unknown, retired,
// or for the other sex; false when the type fits.
func feedTypeRefusal(fc feedContext, row FeedDailyRow, path, code string) (httpx.Refusal, bool) {
	r := httpx.Refusal{Field: path + ".feedTypeId", SectionCode: code, Sex: string(row.Sex)}
	switch ft, ok := fc.Types[row.FeedTypeID]; {
	case !ok:
		r.Message = messages.InvalidData()
	case !ft.Active:
		r.Message = messages.FeedTypeRetired(ft.Name)
	case ft.ForSex != nil && *ft.ForSex != row.Sex:
		r.Message = messages.FeedTypeWrongSex(ft.Name, row.Sex)
	default:
		return httpx.Refusal{}, false
	}
	return r, true
}

// planUndoFeedUpdate puts a Hrană zilnică entry back as before (the update's
// old_values, Q8): by row id (diffByID), a row of the snapshot that is live
// now with another g or type is changed back in place, one that is not live
// is un-deleted, a live row the snapshot doesn't hold (the update added it)
// is soft-deleted. The un-deleted rows go back in Saved.Restored with a
// failed write's error (OnConflict, ARC-5). H5 is checked on the save put
// back (warnings).
func planUndoFeedUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	e := tx.Entry
	now, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	d := diffByID(now.Rows.FeedAllocations, before.FeedAllocations, func(cur, b SnapshotFeed) bool {
		return cur.FeedTypeID == b.FeedTypeID && sameNum(cur.G, b.G) && sameNum(cur.KgRecorded, b.KgRecorded)
	})
	w := feedWrites{deletes: d.remove, restore: d.restore}
	for _, b := range d.update {
		w.updates = append(w.updates, feedUpdate{ID: b.ID, FeedTypeID: b.FeedTypeID, G: b.G, KgRecorded: b.KgRecorded})
	}
	diffs := diffFeed(feedRowValues(now.Rows.FeedAllocations), feedRowValues(before.FeedAllocations))
	back := Snapshot{FeedAllocations: before.FeedAllocations}

	info, err := tx.Q.EntryDeleteContext(ctx, e.ID)
	if err != nil {
		return Saved{}, fmt.Errorf("entries: undo context: %w", err)
	}
	f, err := readRowFigures(ctx, tx, back, +1)
	if err != nil {
		return Saved{}, err
	}
	if err := w.apply(ctx, tx); err != nil {
		return Saved{Restored: Snapshot{FeedAllocations: w.restore}}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, before.Corrected)
	if err != nil {
		return Saved{}, err
	}
	stock, err := readStockAfter(ctx, tx, back, f.feedKg, info.FarmName, false)
	if err != nil {
		return Saved{}, err
	}

	summary := messages.FeedDailySummary(len(before.FeedAllocations), f.feedOf, f.feedTotal)
	at := timex.At(tx.Now)
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: warningsOrEmpty(stock.warnings), Toast: fieldsUndone(diffs), AsOf: at},
		Audit: audit.Row{
			SameDayAuthor: true, RecordTypes: []string{recordTypeFeed}, Summary: &summary, Changes: fieldChanges(diffs),
			OldValues: updateSnapshot{Snapshot: now.Rows, Corrected: e.Corrected}, NewValues: before,
		},
		Events: []live.Event{live.EntryUpdated(live.EntryPayload{
			EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: fieldSections(diffs),
			RecordTypes: []string{recordTypeFeed}, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: before.Corrected,
		})},
	}, nil
}

// feedWrites are the child-row writes of a change of a feed save: rows
// soft-deleted and un-deleted by id, rows changed in place by id, rows
// inserted.
type feedWrites struct {
	deletes []uuid.UUID
	restore []SnapshotFeed // an undo's rows to un-delete
	updates []feedUpdate
	inserts []feedInsert
}

// feedUpdate is one EntryUpdateFeedRows row.
type feedUpdate struct {
	ID         uuid.UUID `json:"id"`
	FeedTypeID uuid.UUID `json:"feed_type_id"`
	G          *num.Num  `json:"g"`
	KgRecorded *num.Num  `json:"kg_recorded"`
}

// feedInsert is one EntryInsertFeedRows row.
type feedInsert struct {
	SectionID  uuid.UUID `json:"section_id"`
	Sex        dbq.Sex   `json:"sex"`
	FeedTypeID uuid.UUID `json:"feed_type_id"`
	G          *num.Num  `json:"g"`
}

// apply runs the writes on the change's entry, in this order: soft deletes
// (a key the update added is free again before its old row comes back),
// un-deletes (a key another entry took meanwhile → 23505, mapped by
// OnConflict), changes in place, inserts. A count of rows changed that is
// not the asked one is 409 STALE rather than half a change.
func (w feedWrites) apply(ctx context.Context, tx *ChangeTx) error {
	e := tx.Entry
	if len(w.deletes) > 0 {
		got, err := tx.Q.EntrySoftDeleteRows(ctx, dbq.EntrySoftDeleteRowsParams{
			Now: tx.Now, EntryID: e.ID, BirdIds: []uuid.UUID{}, EggIds: []uuid.UUID{}, FeedIds: w.deletes,
			DeliveryIds: []uuid.UUID{},
		})
		if err != nil {
			return fmt.Errorf("entries: soft-delete feed rows: %w", err)
		}
		if int(got) != len(w.deletes) {
			return httpx.Stale()
		}
	}
	if len(w.restore) > 0 {
		got, err := tx.Q.EntryRestoreRows(ctx, dbq.EntryRestoreRowsParams{
			EntryID: e.ID, BirdIds: []uuid.UUID{}, EggIds: []uuid.UUID{}, FeedIds: ids(w.restore), DeliveryIds: []uuid.UUID{},
		})
		if err != nil {
			return err // a 23505 goes to OnConflict after the transaction
		}
		if int(got) != len(w.restore) {
			return httpx.Stale()
		}
	}
	if len(w.updates) > 0 {
		rows, err := jsonRows(w.updates)
		if err != nil {
			return err
		}
		got, err := tx.Q.EntryUpdateFeedRows(ctx, dbq.EntryUpdateFeedRowsParams{EntryID: e.ID, Rows: rows})
		if err != nil {
			return fmt.Errorf("entries: update feed rows: %w", err)
		}
		if int(got) != len(w.updates) {
			return httpx.Stale()
		}
	}
	if len(w.inserts) > 0 {
		rows, err := jsonRows(w.inserts)
		if err != nil {
			return err
		}
		if err := tx.Q.EntryInsertFeedRows(ctx, dbq.EntryInsertFeedRowsParams{
			EntryID: e.ID, FarmID: e.FarmID, Date: e.Date, Rows: rows,
		}); err != nil {
			return err // a 23505 goes to OnConflict after the transaction
		}
	}
	return nil
}

// feedRowValue is one row of a feed save as a diff compares it.
type feedRowValue struct {
	SectionID uuid.UUID
	Section   string
	Sex       dbq.Sex
	TypeID    uuid.UUID
	Type      string
	G         *num.Num
}

func feedRowValues(rows []SnapshotFeed) []feedRowValue {
	out := make([]feedRowValue, len(rows))
	for i, r := range rows {
		out[i] = feedRowValue{SectionID: r.SectionID, Section: r.Section, Sex: r.Sex, TypeID: r.FeedTypeID, Type: r.Type, G: r.G}
	}
	return out
}

// diffFeed is the diff of two feed saves of one farm and day, old → new
// (§4.10 `changes`, Q12), rows matched by (section, sex): in new's order,
// then the rows only in old; per row the g ("H5.1 · g găini", g with one
// decimal) then the type ("H5.1 · tip hrană găini", the type's name). A row
// added has null old values, a row removed null new ones.
func diffFeed(old, new []feedRowValue) []fieldDiff {
	var out []fieldDiff
	at := func(rows []feedRowValue, r feedRowValue) (feedRowValue, bool) {
		i := slices.IndexFunc(rows, func(x feedRowValue) bool { return x.SectionID == r.SectionID && x.Sex == r.Sex })
		if i < 0 {
			return feedRowValue{}, false
		}
		return rows[i], true
	}
	pair := func(o, n feedRowValue, hasOld, hasNew bool) {
		r := n
		if !hasNew {
			r = o
		}
		g := func(x feedRowValue, has bool) audit.Value {
			if !has {
				return audit.Null()
			}
			return audit.OptNumber(x.G)
		}
		t := func(x feedRowValue, has bool) audit.Value {
			if !has {
				return audit.Null()
			}
			return audit.Text(x.Type)
		}
		if og, ng := g(o, hasOld), g(n, hasNew); !og.Equal(ng) {
			out = append(out, fieldDiff{SectionID: r.SectionID, Label: messages.FeedLabel(r.Section, r.Sex, messages.FieldG), Places: 1,
				Change: audit.Change{Key: messages.AuditKeyFeed(r.Section, r.Sex, messages.FieldG), Section: r.Section, Sex: r.Sex,
					Field: string(messages.FieldG), Old: og, New: ng}})
		}
		if hasOld != hasNew || o.TypeID != n.TypeID {
			out = append(out, fieldDiff{SectionID: r.SectionID, Label: messages.FeedLabel(r.Section, r.Sex, messages.FieldFeedTypeID),
				Change: audit.Change{Key: messages.AuditKeyFeed(r.Section, r.Sex, messages.FieldFeedTypeID), Section: r.Section,
					Sex: r.Sex, Field: string(messages.FieldFeedTypeID), Old: t(o, hasOld), New: t(n, hasNew)}})
		}
	}
	for _, n := range new {
		o, had := at(old, n)
		pair(o, n, had, true)
	}
	for _, o := range old {
		if _, kept := at(new, o); !kept {
			pair(o, feedRowValue{}, true, false)
		}
	}
	return out
}

// fieldDiff is one changed value of a feed or delivery save, with what the
// Motiv sentence and the Toasts name it by.
type fieldDiff struct {
	SectionID uuid.UUID // the row's section; nil for a delivery
	Label     messages.Label
	Places    int32 // decimals a number is shown with
	Change    audit.Change
}

// shown is v as the Motiv sentence and the Toasts show it.
func (d fieldDiff) shown(v audit.Value) messages.Shown {
	if s, ok := v.Text(); ok {
		return messages.ShownText(s)
	}
	if n, ok := v.Num(); ok {
		return messages.ShownDec(n.Decimal(), d.Places)
	}
	return messages.ShownEmpty()
}

func fieldChanges(diffs []fieldDiff) []audit.Change {
	out := make([]audit.Change, len(diffs))
	for i, d := range diffs {
		out[i] = d.Change
	}
	return out
}

// fieldSections are the codes of the sections with a change, in diff order.
func fieldSections(diffs []fieldDiff) []string {
	out := []string{}
	for _, d := range diffs {
		if d.Change.Section != "" && !slices.Contains(out, d.Change.Section) {
			out = append(out, d.Change.Section)
		}
	}
	return out
}

// fieldsReasonMissing is the §6.15 400 on `reason`: "Motiv lipsă. Scrie de
// ce se schimbă H5.1 ♀ din 160 în 158 g; …" (one value), "… 3 valori; …".
func fieldsReasonMissing(diffs []fieldDiff) string {
	if len(diffs) == 1 {
		d := diffs[0]
		return messages.ReasonMissingChange(d.Label, d.shown(d.Change.Old), d.shown(d.Change.New))
	}
	return messages.ReasonMissingChanges(len(diffs))
}

// fieldsCorrected is a correction's Toast (Q15): "Corecție salvată · H5.1 ♀
// 160 → 158 g" (one value), "Corecție salvată · 3 valori".
func fieldsCorrected(diffs []fieldDiff) string {
	if len(diffs) == 1 {
		d := diffs[0]
		return messages.CorrectionSaved(d.Label, d.shown(d.Change.Old), d.shown(d.Change.New))
	}
	return messages.CorrectionsSaved(len(diffs))
}

// fieldsUndone is an undone update's Toast (Q15), from → to as the undo sets
// them: "Corecție anulată · Livrare 8 000 → 9 000 kg", "Corecție anulată · 3
// valori".
func fieldsUndone(diffs []fieldDiff) string {
	if len(diffs) == 1 {
		d := diffs[0]
		return messages.CorrectionUndone(d.Label, d.shown(d.Change.Old), d.shown(d.Change.New))
	}
	return messages.CorrectionsUndone(len(diffs))
}
