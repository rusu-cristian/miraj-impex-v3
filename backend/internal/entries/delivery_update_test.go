package entries

import (
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// pastDelivery is Ion Rusu's delivery of 500 kg Breeder 2 on 17.11, saved
// on 17.11 at 08:15 (outside his edit window: a correction).
func (f *feedFixture) pastDelivery() uuid.UUID {
	f.t.Helper()
	e := f.entry(dbq.EntryKindDelivery, f.farmA, nov17, f.ion, at0815.AddDate(0, 0, -1))
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 500)`,
		e, f.farmA, f.breeder2, nov17)
	return e
}

// putDelivery sends a delivery PUT of e by p at e's version.
func (f *feedFixture) putDelivery(p auth.Principal, e uuid.UUID, day timex.Day, ft uuid.UUID, kg string, note *string, reason string, dryRun bool) (Result, error) {
	f.t.Helper()
	q := num.MustParse(kg)
	req := DeliveryUpdateRequest{
		IfVersion: f.version(e), DryRun: dryRun,
		Values: DeliveryUpdateValues{DeliveryValues: DeliveryValues{FarmID: f.farmA, Date: day, FeedTypeID: ft, QtyKg: &q, Note: note}},
	}
	if reason != "" {
		req.Reason = &reason
	}
	return f.svc.UpdateDelivery(f.t.Context(), p, e, req)
}

// deliveryRow is the entry's delivery row: kg, type, note and price pair.
func (f *feedFixture) deliveryRow(e uuid.UUID) (kg string, ft uuid.UUID, note *string, perKg, total *decimal.Decimal) {
	f.t.Helper()
	var q decimal.Decimal
	if err := f.d.App.QueryRow(f.t.Context(), `select qty_kg, feed_type_id, note, price_per_kg, total_lei from feed_delivery
		where entry_id = $1 and deleted_at is null`, e).Scan(&q, &ft, &note, &perKg, &total); err != nil {
		f.t.Fatal(err)
	}
	return q.String(), ft, note, perKg, total
}

// T8 criterion 2, Q11, Q15: the author edits his delivery of today: kg and
// note in place, the edit Toast with the stock at the end of the delivery
// date, audit `update` same_day_author with the Q12 changes, `corrected`
// stays false, entry.updated.
func TestUpdateDeliveryAuthorEdit(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.svc.CreateDelivery(t.Context(), f.ion, f.delivery(nov18, f.breeder2, "500", nil))
	if err != nil {
		t.Fatal(err)
	}
	e := delDecode[EntryWriteResponse](t, res).Entry.ID
	live0 := len(f.liveTypes())
	res, err = f.putDelivery(f.ion, e, nov18, f.breeder2, "400", ptr("  factura 12 "), "ignorat", false)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Livrare salvată · 400 kg Breeder 2 · stoc Hîrbovăț: 1 400 kg" ||
		len(resp.Warnings) != 0 || resp.Alive != nil {
		t.Errorf("%s", res.Body)
	}
	if kg, _, note, _, _ := f.deliveryRow(e); kg != "400" || note == nil || *note != "factura 12" {
		t.Errorf("row %s %v", kg, note)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionUpdate || a.Reason != nil || !a.SameDayAuthor || *a.Summary != "Livrare · 400 kg Breeder 2" ||
		!slices.Equal(a.RecordTypes, []string{"livrare"}) {
		t.Errorf("audit %+v", a)
	}
	if got := f.changesOf(e); got != `[{"key":"Livrare · kg","field":"qtyKg","old":500,"new":400},`+
		`{"key":"Livrare · notă","field":"note","old":null,"new":"factura 12"}]` {
		t.Errorf("changes %s", got)
	}
	if f.corrected(e) {
		t.Error("an author edit set corrected")
	}
	if got := f.liveTypes()[live0:]; !slices.Equal(got, []string{"entry.updated"}) {
		t.Errorf("live %v", got)
	}
}

// T8 criterion 2, §6.15: a correction of an earlier day's delivery needs a
// Motiv naming the change; a type changed to a retired one is refused; a
// change that takes a later day's stock below 0 is saved with a
// STOCK_NEGATIVE warning (D40).
func TestUpdateDeliveryCorrection(t *testing.T) {
	f := newFeedFixture(t)
	e := f.pastDelivery()
	_, err := f.putDelivery(f.maria, e, nov17, f.breeder2, "10", nil, "", false)
	if body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation); len(body.Issues) != 1 || body.Issues[0].Path != "reason" ||
		body.Issues[0].Message != "Motiv lipsă. Scrie de ce se schimbă livrarea din 500 în 10 kg; fără motiv corecția nu se salvează." {
		t.Errorf("%+v", body)
	}
	_, err = f.putDelivery(f.maria, e, nov17, f.breeder1, "500", nil, "Tip greșit.", false)
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); len(body.Refusals) != 1 ||
		body.Refusals[0].Field != "values.feedTypeId" || body.Refusals[0].Message != messages.FeedTypeRetired("Breeder 1") {
		t.Errorf("%+v", body)
	}
	// 510 kg of Breeder 2 on each of 17.11 and 18.11: 1 000 + 500 − 1 020 = 480 kg left today.
	for _, d := range []timex.Day{nov17, nov18} {
		fe := f.entry(dbq.EntryKindFeedDaily, f.farmA, d, f.ion, testdb.Now())
		for _, code := range []string{"H5.1", "H5.2", "H6.1"} {
			f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
				values ($1, $2, $3, $4, 'F', $5, 300)`, fe, f.farmA, f.sec[code], d, f.breeder2)
		}
	}
	res, err := f.putDelivery(f.maria, e, nov17, f.breeder2, "10", nil, "Factura spune 10 kg.", false)
	if err != nil {
		t.Fatal(err)
	}
	warning := httpx.Warning{Code: httpx.WarnStockNegative, Message: "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −10 kg. Verifică livrările."}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție salvată · Livrare 500 → 10 kg" ||
		!slices.Equal(resp.Warnings, []httpx.Warning{warning}) {
		t.Errorf("%s", res.Body)
	}
	if a := f.latestAudit(e); a.Reason == nil || *a.Reason != "Factura spune 10 kg." || a.SameDayAuthor {
		t.Errorf("audit %+v", a)
	}
	if !f.corrected(e) {
		t.Error("corrected not set")
	}
}

// T8 criterion 2, Q11: a priced delivery keeps its typed price and the
// other value is derived for the new kg (the derived pair in `changes`,
// priced_by/at kept); the undo puts kg and the pair back exactly.
func TestUpdateDeliveryPriced(t *testing.T) {
	for _, tc := range []struct {
		input, perKg, total    string
		wantPerKg, wantTotal   string
		key, field, oldV, newV string
	}{
		{"per_kg", "7.74", "3870", "7.74", "3096", "Breeder 2 · 17.11.2026 · total lei", "totalLei", "3870", "3096"},
		{"total", "7.74", "3870", "9.675", "3870", "Breeder 2 · 17.11.2026 · lei/kg", "pricePerKg", "7.74", "9.675"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			f := newFeedFixture(t)
			e := f.pastDelivery()
			f.exec(`update feed_delivery set price_per_kg = $1, total_lei = $2, price_input = $3, priced_by = $4, priced_at = $5
				where entry_id = $6`, tc.perKg, tc.total, tc.input, f.owner.UserID, testdb.Now(), e)
			res, err := f.putDelivery(f.maria, e, nov17, f.breeder2, "400", nil, "", true)
			if err != nil {
				t.Fatal(err)
			}
			wantChanges := `[{"key":"Livrare · kg","field":"qtyKg","old":500,"new":400},` +
				`{"key":"` + tc.key + `","field":"` + tc.field + `","old":` + tc.oldV + `,"new":` + tc.newV + `}]`
			if p := delDecode[CorrectionPreview](t, res); changeJSON(t, p.Changes) != wantChanges {
				t.Errorf("preview %s", res.Body)
			}
			if _, err := f.putDelivery(f.maria, e, nov17, f.breeder2, "400", nil, "Cântărit la recepție.", false); err != nil {
				t.Fatal(err)
			}
			if got := f.changesOf(e); got != wantChanges {
				t.Errorf("changes %s", got)
			}
			if kg, _, _, perKg, total := f.deliveryRow(e); kg != "400" || perKg.String() != tc.wantPerKg || total.String() != tc.wantTotal {
				t.Errorf("row %s %s %s", kg, perKg, total)
			}
			var by *uuid.UUID
			if err := f.d.App.QueryRow(t.Context(), `select priced_by from feed_delivery where entry_id = $1`, e).Scan(&by); err != nil ||
				by == nil || *by != f.owner.UserID {
				t.Errorf("priced_by %v (%v)", by, err)
			}

			res, err = f.svc.UndoEntry(t.Context(), f.maria, e)
			if err != nil {
				t.Fatal(err)
			}
			if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție anulată · Livrare 400 → 500 kg" {
				t.Errorf("undo %s", res.Body)
			}
			if kg, _, _, perKg, total := f.deliveryRow(e); kg != "500" || perKg.String() != tc.perKg || total.String() != tc.total {
				t.Errorf("row after undo %s %s %s", kg, perKg, total)
			}
			if f.corrected(e) {
				t.Error("corrected not put back")
			}
		})
	}
}

// Q8, §6.8: a price changed in Prețuri after the correction is kept by the
// undo, which derives the other value again for the kg it puts back.
func TestUndoDeliveryUpdateKeepsNewPrice(t *testing.T) {
	f := newFeedFixture(t)
	e := f.pastDelivery()
	f.exec(`update feed_delivery set price_per_kg = 7.74, total_lei = 3870, price_input = 'per_kg', priced_by = $1, priced_at = $2
		where entry_id = $3`, f.owner.UserID, testdb.Now(), e)
	if _, err := f.putDelivery(f.maria, e, nov17, f.breeder2, "400", nil, "Cântărit la recepție.", false); err != nil {
		t.Fatal(err)
	}
	f.exec(`update feed_delivery set price_per_kg = 8, total_lei = 3200 where entry_id = $1`, e) // Prețuri › Modifică
	if _, err := f.svc.UndoEntry(t.Context(), f.maria, e); err != nil {
		t.Fatal(err)
	}
	if kg, _, _, perKg, total := f.deliveryRow(e); kg != "500" || perKg.String() != "8" || total.String() != "4000" {
		t.Errorf("row after undo %s %s %s", kg, perKg, total)
	}
}

// T8 criterion 2, Q11: price fields in `values` → a manager 403
// PRICE_FORBIDDEN, the owner 400 unknown_field (null counts too); a vet
// 403; the create's level-1 checks under "values."; nothing changed → 422;
// a stale version → 409.
func TestUpdateDeliveryHandler(t *testing.T) {
	f := newFeedFixture(t)
	e := f.pastDelivery()
	body := func(over map[string]any) string {
		vals := map[string]any{"farmId": f.farmA, "date": "2026-11-17", "feedTypeId": f.breeder2, "qtyKg": 400, "note": nil}
		for k, v := range over {
			vals[k] = v
		}
		b, err := json.Marshal(map[string]any{"ifVersion": f.version(e), "values": vals, "reason": "Cântărit la recepție."})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	for _, tc := range []struct {
		name        string
		p           auth.Principal
		body        string
		status      int
		code, issue string
		message     string
	}{
		{"manager price", f.ion, body(map[string]any{"pricePerKg": 7.74}), 403, httpx.CodePriceForbidden, "", messages.PriceForbidden()},
		{"owner price", f.owner, body(map[string]any{"totalLei": nil}), 400, httpx.CodeValidation, "values.totalLei", messages.InvalidData()},
		{"vet", f.vet, body(nil), 403, httpx.CodeForbidden, "", messages.PageForbidden()},
		{"viewer", f.viewer, body(nil), 403, httpx.CodeForbidden, "", messages.PageForbidden()},
		{"kg", f.ion, body(map[string]any{"qtyKg": 0}), 400, httpx.CodeValidation, "values.qtyKg", messages.DeliveryQtyRequired("Breeder 2", "Hîrbovăț")},
		{"unknown member", f.ion, body(map[string]any{"kg": 1}), 400, httpx.CodeValidation, "values.kg", messages.InvalidData()},
		{"date", f.ion, body(map[string]any{"date": "2026-11-18"}), 400, httpx.CodeValidation, "values.date", messages.InvalidData()},
		{"nothing changed", f.ion, body(map[string]any{"qtyKg": 500, "note": "  "}), 422, httpx.CodeRefused, "", messages.NothingChanged()},
	} {
		w := f.sendPut(tc.p, e, tc.body)
		b := envelopeOf(t, w)
		if w.Code != tc.status || b.Code != tc.code || b.Message != tc.message || (tc.issue != "" && (len(b.Issues) == 0 || b.Issues[0].Path != tc.issue)) {
			t.Errorf("%s: %d %s", tc.name, w.Code, w.Body)
		}
	}
	if w := f.sendPut(f.ion, e, body(nil)); w.Code != http.StatusOK {
		t.Errorf("%d %s", w.Code, w.Body)
	}
	if w := f.sendPut(f.ion, e, `{"ifVersion": "2026-11-17T06:15:00Z", "reason": "x", "values": {"farmId": "`+f.farmA.String()+
		`", "date": "2026-11-17", "feedTypeId": "`+f.breeder2.String()+`", "qtyKg": 300}}`); w.Code != http.StatusConflict {
		t.Errorf("stale: %d %s", w.Code, w.Body)
	}
}
