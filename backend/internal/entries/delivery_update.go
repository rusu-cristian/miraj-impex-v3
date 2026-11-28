package entries

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Livrare hrană adapter of PUT /api/entries/{id} and of the undo of its
// update (§7.8, §6.7, §6.8, §6.15; plan M5-D5, Q8, Q11, Q12): kg, type and
// note are editable; the date and farm are not (D101); the price is set only
// in Prețuri, so a priced delivery keeps its typed price and the other value
// is derived again for the new kg.

// DeliveryUpdateValues is `values` of a delivery PUT: the delivery's create
// shape (DeliveryValues, what GET /api/entries/{id} returns) and the price
// fields, declared so that a price sent here is refused as on a create (a
// manager 403 PRICE_FORBIDDEN, the owner 400 unknown_field) rather than
// silently decoded away. Any value, null too, counts as present.
type DeliveryUpdateValues struct {
	DeliveryValues
	PricePerKg json.RawMessage `json:"pricePerKg"`
	TotalLei   json.RawMessage `json:"totalLei"`
}

// DeliveryUpdateRequest is a PUT of a Livrare hrană entry.
type DeliveryUpdateRequest = EntryUpdateRequest[DeliveryUpdateValues]

// note is the values' note trimmed, nil when absent or blank (as a create).
func (v DeliveryUpdateValues) note() *string {
	if v.Note == nil {
		return nil
	}
	s := strings.TrimSpace(*v.Note)
	if s == "" {
		return nil
	}
	return &s
}

// updateDelivery decodes and checks a delivery PUT: decode (64 KiB, every
// member known) → role and price fields (checkDeliveryUpdate) → the farm and
// type names → level 1 (validateDeliveryUpdate) → service.
func (h *Handler) updateDelivery(w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	req, err := httpx.Decode[DeliveryUpdateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		return Result{}, err
	}
	if err := checkDeliveryUpdate(p, req.Values); err != nil {
		return Result{}, err
	}
	refs, err := h.svc.DeliveryRefs(r.Context(), DeliveryCreateRequest{FarmID: t.FarmID, FeedTypeID: req.Values.FeedTypeID})
	if err != nil {
		return Result{}, err
	}
	var v httpx.Validator
	validateDeliveryUpdate(&v, req, t, refs)
	if err := v.Err(); err != nil {
		return Result{}, err
	}
	return h.svc.UpdateDelivery(r.Context(), p, t.ID, req)
}

// checkDeliveryUpdate is the role and the price fields of a delivery PUT
// (§6.7, Q11, as checkDelivery on a create): manager and owner only (403
// FORBIDDEN); values.pricePerKg or values.totalLei present → a manager 403
// PRICE_FORBIDDEN, the owner 400 unknown_field.
func checkDeliveryUpdate(p auth.Principal, v DeliveryUpdateValues) error {
	if err := checkDeliveryWriter(p, uuid.Nil); err != nil {
		return err
	}
	var issues httpx.Validator
	for _, f := range []struct {
		path string
		raw  []byte
	}{{"values.pricePerKg", v.PricePerKg}, {"values.totalLei", v.TotalLei}} {
		if f.raw == nil {
			continue
		}
		if p.Role != auth.Owner {
			return httpx.PriceForbidden()
		}
		issues.Add(f.path, httpx.IssueUnknownField, messages.InvalidData())
	}
	return issues.Err()
}

// validateDeliveryUpdate is the level-1 check of a delivery PUT (§7.0b):
// validateHead, the values as a delivery create checks them (type known, kg >
// 0 with the S13 sentence, one decimal, the note's length) with every path
// under "values.", validateTail.
func validateDeliveryUpdate(v *httpx.Validator, r DeliveryUpdateRequest, t updateTarget, refs DeliveryRefs) {
	r.validateHead(v, t, r.Values.FarmID, r.Values.Date)
	var values httpx.Validator
	DeliveryCreateRequest{
		FarmID: r.Values.FarmID, Date: r.Values.Date, FeedTypeID: r.Values.FeedTypeID, QtyKg: r.Values.QtyKg, Note: r.Values.Note,
	}.Validate(&values, refs)
	addIssuesUnder(v, "values.", values)
	r.validateTail(v)
}

// UpdateDelivery edits or corrects delivery entry id for p (§7.8, §6.7,
// §6.15, plan M5-D5, Q11) on the change runner; req must have passed
// validateDeliveryUpdate. The author on the creation day edits (no Motiv,
// "Livrare salvată · …"); a manager on own farms or the owner corrects
// (Motiv required, `corrected` set, "Corecție salvată · …"). Under the locks
// (planDeliveryUpdate): values name the stored farm and date, ifVersion is
// the version; the diff of kg, type and note (none → 422 "Nu ai schimbat
// nicio cifră."); a type changed to a retired one → 422; a priced delivery
// keeps its typed price and the other value is derived for the new kg
// (calc.DerivePrice), the derived pair in `changes`; the row changed in
// place; a stock of the old or new type that dips below 0 on a day up to
// today → STOCK_NEGATIVE warning (saved anyway, D40). A dry run answers 200
// CorrectionPreview and rolls back.
func (s *Service) UpdateDelivery(ctx context.Context, p auth.Principal, id uuid.UUID, req DeliveryUpdateRequest) (Result, error) {
	if err := checkDeliveryUpdate(p, req.Values); err != nil {
		return Result{}, err
	}
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionUpdate, DryRun: req.DryRun, Meta: req.meta(), Allow: allowUpdate(p),
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		return planDeliveryUpdate(ctx, tx, req)
	})
}

// planDeliveryUpdate is the delivery adapter's plan of a PUT
// (UpdateDelivery).
func planDeliveryUpdate(ctx context.Context, tx *ChangeTx, req DeliveryUpdateRequest) (Saved, error) {
	e := tx.Entry
	v := req.Values
	if e.Kind != dbq.EntryKindDelivery {
		return Saved{}, httpx.NotFound()
	}
	if err := checkUpdateHead(tx, v.FarmID, v.Date, req.IfVersion); err != nil {
		return Saved{}, err
	}
	if v.QtyKg == nil { // validateDeliveryUpdate guarantees it
		return Saved{}, httpx.Validation(httpx.Issue{Path: "values.qtyKg", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	correction := !tx.Actions.Editable
	stored, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	if len(stored.Rows.FeedDeliveries) != 1 {
		return Saved{}, httpx.Stale() // a delivery entry holds exactly one live row
	}
	cur := stored.Rows.FeedDeliveries[0]
	dc, err := tx.Q.EntriesDeliverySaveContext(ctx, dbq.EntriesDeliverySaveContextParams{
		FarmID: e.FarmID, Day: e.Date, FeedTypeID: v.FeedTypeID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Saved{}, httpx.NotFound()
	}
	if err != nil {
		return Saved{}, fmt.Errorf("entries: delivery context: %w", err)
	}
	if dc.FeedTypeName == nil || dc.FeedTypeActive == nil { // level 1 checked it exists; types are never deleted
		return Saved{}, httpx.Validation(httpx.Issue{Path: "values.feedTypeId", Code: httpx.IssueOutOfRange, Message: messages.InvalidData()})
	}
	target := cur
	target.FeedTypeID, target.Type, target.Kg, target.Note = v.FeedTypeID, *dc.FeedTypeName, *v.QtyKg, v.note()
	diffs := diffDelivery(cur, target)

	preview := func(p CorrectionPreview) (Saved, error) {
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
	var refusals []httpx.Refusal
	if cur.FeedTypeID != v.FeedTypeID && !*dc.FeedTypeActive { // a changed value only (D6)
		refusals = append(refusals, httpx.Refusal{Field: "values.feedTypeId", Message: messages.FeedTypeRetired(target.Type)})
	}
	if target.PricePerKg, target.TotalLei, err = repriced(cur, target.Kg.Decimal()); err != nil {
		refusals = append(refusals, httpx.Refusal{Field: "values.qtyKg", Message: messages.InvalidData()}) // the derived price doesn't fit its column
	}
	changes := append(fieldChanges(diffs), priceChanges(cur, target, e.Date)...)
	if len(refusals) > 0 {
		if tx.DryRun {
			return preview(CorrectionPreview{Changes: changes, Refusals: refusals})
		}
		return Saved{}, httpx.RefusedWith(refusals, nil)
	}

	if err := writeDelivery(ctx, tx, target); err != nil {
		return Saved{}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, e.Corrected || correction)
	if err != nil {
		return Saved{}, err
	}
	stock, warnings, err := deliveryStock(ctx, tx, dc.FarmName, target, cur)
	if err != nil {
		return Saved{}, err
	}
	if tx.DryRun {
		return preview(CorrectionPreview{Changes: changes, Warnings: warnings})
	}

	kg := target.Kg.Decimal()
	toast := fieldsCorrected(diffs)
	if !correction { // an author edit (Q15): "Livrare salvată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg"
		toast = messages.DeliverySaved(kg, target.Type, dc.FarmName, stock)
	}
	summary := messages.DeliverySummary(kg, target.Type)
	at := timex.At(tx.Now)
	var motiv *string
	if correction {
		motiv = &reason
	}
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: warnings, Toast: toast, AsOf: at},
		Audit: audit.Row{
			Reason: motiv, SameDayAuthor: !correction, RecordTypes: []string{recordTypeDelivery}, Summary: &summary,
			Changes: changes, OldValues: updateSnapshot{Snapshot: stored.Rows, Corrected: e.Corrected},
			NewValues: deliveryAuditValues{FeedTypeID: target.FeedTypeID, Type: target.Type, Date: e.Date, Kg: target.Kg, Note: target.Note},
		},
		Events: []live.Event{live.EntryUpdated(live.EntryPayload{
			EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date},
			RecordTypes: []string{recordTypeDelivery}, Author: &tx.Principal.Name, At: at, Summary: summary,
			Corrected: e.Corrected || correction,
		})},
	}, nil
}

// planUndoDeliveryUpdate puts a delivery back as before (the update's
// old_values, Q8): its type, kg and note. The typed price stays as the row
// has it now (a price set, changed or undone in Prețuri since is kept): when
// it is still the one before, the pair goes back exactly; else the other
// value is derived again for the kg put back.
func planUndoDeliveryUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	e := tx.Entry
	now, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	// The live row must still be the one before, by id (diffByID): a
	// delivery entry holds exactly one row, changed in place.
	d := diffByID(now.Rows.FeedDeliveries, before.FeedDeliveries, func(SnapshotDelivery, SnapshotDelivery) bool { return true })
	if len(before.FeedDeliveries) != 1 || len(d.restore) > 0 || len(d.remove) > 0 {
		return Saved{}, httpx.Stale()
	}
	cur, back := now.Rows.FeedDeliveries[0], before.FeedDeliveries[0]
	target := back
	target.PriceInput, target.PricedBy, target.PricedAt, target.PriceFromRegister = cur.PriceInput, cur.PricedBy, cur.PricedAt, cur.PriceFromRegister
	switch {
	case cur.TotalLei == nil:
		target.PricePerKg, target.TotalLei = nil, nil
	case back.TotalLei == nil || !sameTypedPrice(cur, back):
		if target.PricePerKg, target.TotalLei, err = repriced(cur, back.Kg.Decimal()); err != nil {
			return Saved{}, httpx.Stale() // the price typed since doesn't fit the kg put back
		}
	}
	diffs := diffDelivery(cur, target)
	dc, err := tx.Q.EntriesDeliverySaveContext(ctx, dbq.EntriesDeliverySaveContextParams{
		FarmID: e.FarmID, Day: e.Date, FeedTypeID: target.FeedTypeID,
	})
	if err != nil {
		return Saved{}, fmt.Errorf("entries: delivery context: %w", err)
	}
	if err := writeDelivery(ctx, tx, target); err != nil {
		return Saved{}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, before.Corrected)
	if err != nil {
		return Saved{}, err
	}
	_, warnings, err := deliveryStock(ctx, tx, dc.FarmName, target, cur)
	if err != nil {
		return Saved{}, err
	}

	summary := messages.DeliverySummary(target.Kg.Decimal(), target.Type)
	at := timex.At(tx.Now)
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: warnings, Toast: fieldsUndone(diffs), AsOf: at},
		Audit: audit.Row{
			SameDayAuthor: true, RecordTypes: []string{recordTypeDelivery}, Summary: &summary,
			Changes:   append(fieldChanges(diffs), priceChanges(cur, target, e.Date)...),
			OldValues: updateSnapshot{Snapshot: now.Rows, Corrected: e.Corrected}, NewValues: before,
		},
		Events: []live.Event{live.EntryUpdated(live.EntryPayload{
			EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date},
			RecordTypes: []string{recordTypeDelivery}, Author: &tx.Principal.Name, At: at, Summary: summary,
			Corrected: before.Corrected,
		})},
	}, nil
}

// writeDelivery changes the entry's delivery row in place to d; a row gone
// under the lock is 409 STALE.
func writeDelivery(ctx context.Context, tx *ChangeTx, d SnapshotDelivery) error {
	n, err := tx.Q.EntryUpdateDeliveryRow(ctx, dbq.EntryUpdateDeliveryRowParams{
		FeedTypeID: d.FeedTypeID, QtyKg: d.Kg.Decimal(), Note: d.Note, PricePerKg: nullDecimal(d.PricePerKg),
		TotalLei: nullDecimal(d.TotalLei), ID: d.ID, EntryID: tx.Entry.ID,
	})
	if err != nil {
		return fmt.Errorf("entries: update delivery row: %w", err)
	}
	if n != 1 {
		return httpx.Stale()
	}
	return nil
}

// deliveryStock reads, after the write, the stock of d's type at the end of
// the delivery date (the edit Toast, as the create's, M4 Q13) and a
// STOCK_NEGATIVE warning per type of d and of before (when it changed) whose
// stock dips below 0 on a day of [date, today] (§6.3 H5, D40), in name
// order.
func deliveryStock(ctx context.Context, tx *ChangeTx, farm string, d, before SnapshotDelivery) (decimal.Decimal, []httpx.Warning, error) {
	e := tx.Entry
	names := map[uuid.UUID]string{d.FeedTypeID: d.Type, before.FeedTypeID: before.Type}
	ids := []uuid.UUID{d.FeedTypeID}
	if before.FeedTypeID != d.FeedTypeID {
		ids = append(ids, before.FeedTypeID)
	}
	slices.SortStableFunc(ids, func(a, b uuid.UUID) int { return cmp.Compare(names[a], names[b]) })
	series, err := calc.StockSeries(ctx, tx.Tx, e.FarmID, ids, e.Date, tx.Today)
	if err != nil {
		return decimal.Zero, nil, err
	}
	warnings := []httpx.Warning{}
	for _, id := range ids {
		if low, ok := calc.Lowest(series[id]); ok && low.Kg.IsNegative() {
			warnings = append(warnings, httpx.Warning{
				Code: httpx.WarnStockNegative, Message: messages.StockWouldGoNegative(names[id], farm, low.Kg),
			})
		}
	}
	stock := decimal.Zero
	if pts := series[d.FeedTypeID]; len(pts) > 0 {
		stock = pts[0].Kg
	}
	return stock, warnings, nil
}

// diffDelivery is the diff of a delivery's editable values, old → new (§4.10
// `changes`, Q12), in body order: "Livrare · tip" (the type's name),
// "Livrare · kg" (one decimal), "Livrare · notă" (text or null).
func diffDelivery(old, new SnapshotDelivery) []fieldDiff {
	var out []fieldDiff
	add := func(field messages.DeliveryField, places int32, o, n audit.Value) {
		if !o.Equal(n) {
			out = append(out, fieldDiff{Label: messages.DeliveryLabel(field), Places: places, Change: audit.Change{
				Key: messages.AuditKeyDelivery(field), Field: string(field), Old: o, New: n,
			}})
		}
	}
	if old.FeedTypeID != new.FeedTypeID {
		out = append(out, fieldDiff{Label: messages.DeliveryLabel(messages.FieldDeliveryFeedTypeID), Change: audit.Change{
			Key: messages.AuditKeyDelivery(messages.FieldDeliveryFeedTypeID), Field: string(messages.FieldDeliveryFeedTypeID),
			Old: audit.Text(old.Type), New: audit.Text(new.Type),
		}})
	}
	add(messages.FieldQtyKg, 1, audit.Number(old.Kg), audit.Number(new.Kg))
	add(messages.FieldNote, 0, audit.OptText(old.Note), audit.OptText(new.Note))
	return out
}

// repriced is the price pair a delivery keeps at kg (Q11, §6.8 "the other is
// derived"): nil, nil while unpriced; the stored pair when kg is unchanged;
// else the typed input (price_input) stays and the other is calc.DerivePrice's.
// A register price has no typed input: it keeps its total lei, the invoice's
// (C6), and its lei/kg is derived.
func repriced(d SnapshotDelivery, kg decimal.Decimal) (perKg, total *num.Num, err error) {
	if d.TotalLei == nil || d.PricePerKg == nil {
		return nil, nil, nil
	}
	if d.Kg.Decimal().Equal(kg) {
		return d.PricePerKg, d.TotalLei, nil
	}
	var p calc.Price
	if d.PriceInput != nil && *d.PriceInput == string(calc.PriceInputPerKg) {
		typed := d.PricePerKg.Decimal()
		p, err = calc.DerivePrice(kg, &typed, nil)
	} else {
		typed := d.TotalLei.Decimal()
		p, err = calc.DerivePrice(kg, nil, &typed)
	}
	if err != nil {
		return nil, nil, err
	}
	pk, t := num.Of(p.PerKg), num.Of(p.Total)
	return &pk, &t, nil
}

// sameTypedPrice reports whether two priced states of a delivery have the
// same typed input and value (the one repriced keeps).
func sameTypedPrice(a, b SnapshotDelivery) bool {
	if !sameString(a.PriceInput, b.PriceInput) {
		return false
	}
	if a.PriceInput != nil && *a.PriceInput == string(calc.PriceInputPerKg) {
		return sameNum(a.PricePerKg, b.PricePerKg)
	}
	return sameNum(a.TotalLei, b.TotalLei)
}

// priceChanges are the derived price values a change moved, old → new, keyed
// as M4's ("Breeder 2 · 12.11.2026 · lei/kg" / "… · total lei"; Q11, Q12).
func priceChanges(old, new SnapshotDelivery, date timex.Day) []audit.Change {
	var out []audit.Change
	if !sameNum(old.PricePerKg, new.PricePerKg) {
		out = append(out, audit.Change{
			Key: messages.AuditKeyPricePerKg(new.Type, date), Field: "pricePerKg",
			Old: audit.OptNumber(old.PricePerKg), New: audit.OptNumber(new.PricePerKg),
		})
	}
	if !sameNum(old.TotalLei, new.TotalLei) {
		out = append(out, audit.Change{
			Key: messages.AuditKeyTotalLei(new.Type, date), Field: "totalLei",
			Old: audit.OptNumber(old.TotalLei), New: audit.OptNumber(new.TotalLei),
		})
	}
	return out
}

func sameString(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// nullDecimal is n as a nullable numeric parameter.
func nullDecimal(n *num.Num) decimal.NullDecimal {
	if n == nil {
		return decimal.NullDecimal{}
	}
	return decimal.NewNullDecimal(n.Decimal())
}
