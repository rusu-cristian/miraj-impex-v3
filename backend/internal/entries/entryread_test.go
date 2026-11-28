package entries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// getEntry sends GET path as p through entryReadRoutes.
func (f *feedFixture) getEntry(p auth.Principal, path string) *httptest.ResponseRecorder {
	f.t.Helper()
	mux := http.NewServeMux()
	NewHandler(f.svc).entryReadRoutes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// auditRow writes an audit row of action on entry e by p at the instant at.
func (f *feedFixture) auditRow(e uuid.UUID, action dbq.AuditAction, p auth.Principal, at time.Time) {
	f.t.Helper()
	f.exec(`insert into audit_log (at, user_id, action, entity, entity_id, entry_id) values ($1, $2, $3, 'entry', $4, $4)`,
		at, p.UserID, action, e)
}

// auditRowWith is auditRow with old_values (a snapshot an undo or a restore
// reads back).
func (f *feedFixture) auditRowWith(e uuid.UUID, action dbq.AuditAction, p auth.Principal, at time.Time, old any) {
	f.t.Helper()
	b, err := json.Marshal(old)
	if err != nil {
		f.t.Fatal(err)
	}
	f.exec(`insert into audit_log (at, user_id, action, entity, entity_id, entry_id, old_values)
		values ($1, $2, $3, 'entry', $4, $4, $5)`, at, p.UserID, action, e, b)
}

// T3 criterion 5 on hand-built data: the route's answers (400 for a
// malformed id or a query parameter, 404 for an unknown, deleted or not
// served entry, 403 FARM_FORBIDDEN for another farm), the head and the
// actions per caller.
func TestEntryReadRoute(t *testing.T) {
	f := newFeedFixture(t)
	e := f.zilnic() // the vet's, 18.11 08:15
	f.auditRow(e, dbq.AuditActionCreate, f.vet, at0815)
	deleted := f.entry(dbq.EntryKindDaily, f.farmA, nov17, f.ion, at0815)
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, deleted, testdb.Now(), f.ion.UserID)
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)

	envelope := func(w *httptest.ResponseRecorder) httpx.APIError {
		t.Helper()
		var b httpx.ErrorBody
		if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
			t.Fatalf("%d %s: %v", w.Code, w.Body, err)
		}
		return b.Error
	}
	for _, tc := range []struct {
		name, path string
		p          auth.Principal
		status     int
		code       string
		issue      string
	}{
		{"malformed id", "/api/entries/abc", f.ion, 400, httpx.CodeValidation, "id"},
		{"query parameter", "/api/entries/" + e.String() + "?farmId=x", f.ion, 400, httpx.CodeValidation, "farmId"},
		{"unknown", "/api/entries/" + uuid.NewString(), f.ion, 404, httpx.CodeNotFound, ""},
		{"deleted", "/api/entries/" + deleted.String(), f.ion, 404, httpx.CodeNotFound, ""},
		{"not served", "/api/entries/" + sale.String(), f.owner, 404, httpx.CodeNotFound, ""},
		{"another farm", "/api/entries/" + e.String(), f.outsider, 403, httpx.CodeFarmForbidden, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := f.getEntry(tc.p, tc.path)
			got := envelope(w)
			if w.Code != tc.status || got.Code != tc.code {
				t.Fatalf("got %d %s, want %d %s", w.Code, got.Code, tc.status, tc.code)
			}
			if tc.issue != "" && (len(got.Issues) == 0 || got.Issues[0].Path != tc.issue || got.Issues[0].Code == "") {
				t.Errorf("issues = %+v", got.Issues)
			}
		})
	}

	read := func(p auth.Principal) EntryReadResponse {
		t.Helper()
		w := f.getEntry(p, "/api/entries/"+e.String())
		if w.Code != http.StatusOK || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%d %s (Cache-Control %q)", w.Code, w.Body, w.Header().Get("Cache-Control"))
		}
		var res EntryReadResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if _, ok := res.Values.(map[string]any)["sections"]; !ok {
			t.Errorf("values = %v", res.Values)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
			t.Fatal(err)
		}
		if _, ok := raw["feedTypes"]; ok {
			t.Error("a daily entry carries feedTypes")
		}
		return res
	}
	res := read(f.vet)
	h := res.Entry
	if h.ID != e || h.Kind != dbq.EntryKindDaily || h.Version != "2026-11-18T06:15:00.000000Z" ||
		h.Farm != (EntryFarm{ID: f.farmA, Name: "Hîrbovăț"}) || h.Date != nov18 || h.SAPT == nil || *h.SAPT != 3 ||
		h.AgeDays == nil || *h.AgeDays != 17 || h.Author == nil || h.Author.Name != "Dr. Ceban" ||
		h.CreatedAt.String() != "2026-11-18T08:15:00+02:00" || h.Source != dbq.RecordSourceManual || h.Corrected != nil {
		t.Errorf("head = %+v", h)
	}
	if len(res.Sections) != 1 || res.Sections[0] != (EntrySection{ID: f.sec["H5.1"], Code: "H5.1", HallCode: "H5"}) ||
		res.AsOf.String() != "2026-11-18T12:00:00+02:00" {
		t.Errorf("sections %+v asOf %s", res.Sections, res.AsOf)
	}
	for _, tc := range []struct {
		name string
		p    auth.Principal
		want EntryReadActions
	}{
		{"author, the creation day: edit, delete, undo of the create", f.vet, EntryReadActions{Edit: true, Delete: true, Undo: true}},
		{"manager: correct with a Motiv", f.ion, EntryReadActions{Correct: true, Delete: true}},
		{"owner", f.owner, EntryReadActions{Correct: true, Delete: true}},
		{"viewer: read only", f.viewer, EntryReadActions{}},
	} {
		if got := read(tc.p).Actions; got != tc.want {
			t.Errorf("%s: actions %+v, want %+v", tc.name, got, tc.want)
		}
	}

	// Ion corrects it, then deletes one section's values: corrected by him;
	// his delete of today is restorable by him only. An audit row without
	// an old_values snapshot (as the imported correction) offers no undo or
	// restore: POST would answer 409 STALE (criterion 11).
	stored, err := LoadStored(t.Context(), f.d.App, e, dbq.EntryKindDaily, f.farmA, nov18)
	if err != nil {
		t.Fatal(err)
	}
	f.exec(`update entry set corrected = true, updated_by = $2, updated_at = $3 where id = $1`, e, f.ion.UserID, testdb.Now())
	f.auditRow(e, dbq.AuditActionUpdate, f.ion, testdb.Now())
	if got := read(f.ion).Actions; got != (EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("after a correction without a snapshot: %+v", got)
	}
	f.auditRowWith(e, dbq.AuditActionUpdate, f.ion, testdb.Now(), updateSnapshot{Snapshot: stored.Rows})
	if got := read(f.ion); got.Actions != (EntryReadActions{Correct: true, Delete: true, Undo: true}) ||
		got.Entry.Corrected == nil || got.Entry.Corrected.By != "Ion Rusu" || got.Entry.Corrected.At.String() != "2026-11-18T12:00:00+02:00" ||
		got.Entry.Version != "2026-11-18T10:00:00.000000Z" {
		t.Errorf("after the correction: %+v %+v", got.Actions, got.Entry)
	}
	f.auditRow(e, dbq.AuditActionDelete, f.ion, testdb.Now()) // same instant: the later row is the latest
	if got := read(f.ion).Actions; got != (EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("after a delete without a snapshot: %+v", got)
	}
	section := f.sec["H5.1"]
	f.auditRowWith(e, dbq.AuditActionDelete, f.ion, testdb.Now(), removedRows{Snapshot: stored.Rows, SectionID: &section})
	if got := read(f.ion).Actions; got != (EntryReadActions{Correct: true, Delete: true, Restore: true}) {
		t.Errorf("after the delete: %+v", got)
	}
	if got := read(f.vet).Actions; got != (EntryReadActions{Edit: true, Delete: true}) {
		t.Errorf("the vet after Ion's delete: %+v", got)
	}

	// Yesterday's action is no longer undoable.
	f.d.Clock.Advance(24 * time.Hour)
	defer f.d.Clock.Set(testdb.Now())
	if got := read(f.ion).Actions; got != (EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("the next day: %+v", got)
	}
}

// T3 criterion 5: a feed_daily entry carries its feed types; an entry in a
// closed series offers no action (§3.1); an entry without a cycle has no
// SAPT or age.
func TestBuildEntryReadFeedAndClosed(t *testing.T) {
	f := newFeedFixture(t)
	ctx := t.Context()
	feed := f.entry(dbq.EntryKindFeedDaily, f.farmA, nov18, f.vet, at0815)
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, 'F', $5, 160)`, feed, f.farmA, f.sec["H5.1"], nov18, f.breeder2)
	in := EntryReadInput{Principal: f.owner, EntryID: feed, Today: testdb.Today, Now: testdb.Now()}
	res, err := BuildEntryRead(ctx, f.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.FeedTypes) != 1 || res.FeedTypes[0].Name != "Breeder 2" || res.Actions != (EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("feed: types %+v actions %+v", res.FeedTypes, res.Actions)
	}

	closed := f.entry(dbq.EntryKindDelivery, f.farmB, timex.MustDay(2026, time.November, 5), f.owner, at0815)
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 500)`,
		closed, f.farmB, f.breeder2, timex.MustDay(2026, time.November, 5))
	f.auditRow(closed, dbq.AuditActionCreate, f.owner, testdb.Now())
	in.EntryID = closed
	res, err = BuildEntryRead(ctx, f.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Actions != (EntryReadActions{}) || res.Entry.SAPT == nil {
		t.Errorf("closed series: actions %+v sapt %v", res.Actions, res.Entry.SAPT)
	}

	gap := f.entry(dbq.EntryKindDelivery, f.farmB, nov17, f.owner, at0815) // Telița between cycles
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, 500)`,
		gap, f.farmB, f.breeder2, nov17)
	in.EntryID = gap
	res, err = BuildEntryRead(ctx, f.d.App, in)
	if err != nil {
		t.Fatal(err)
	}
	if res.Entry.SAPT != nil || res.Entry.AgeDays != nil || res.Actions != (EntryReadActions{Edit: true, Delete: true}) {
		t.Errorf("no cycle: sapt %v age %v actions %+v", res.Entry.SAPT, res.Entry.AgeDays, res.Actions)
	}
}

// SEC-1: a priced delivery read by a vet or a viewer (or anyone) carries no
// price key: `values` is the create shape, which has none (Q6; the price
// is set in Prețuri and shown only to the owner).
func TestEntryReadPricedDeliveryHasNoPrice(t *testing.T) {
	f := newFeedFixture(t)
	e := f.entry(dbq.EntryKindDelivery, f.farmA, nov18, f.vet, at0815)
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei, price_input,
		priced_by, priced_at) values ($1, $2, $3, $4, 500, 7.74, 3870, 'per_kg', $5, $6)`,
		e, f.farmA, f.breeder2, nov18, f.owner.UserID, testdb.Now())
	for _, p := range []auth.Principal{f.vet, f.viewer, f.owner} {
		w := f.getEntry(p, "/api/entries/"+e.String())
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"qtyKg":500`) {
			t.Fatalf("%s: %d %s", p.Name, w.Code, w.Body)
		}
		body := strings.ToLower(w.Body.String())
		for _, key := range []string{"price", "lei", "7.74", "3870", "per_kg", "pricedby", "pricedat"} {
			if strings.Contains(body, key) {
				t.Errorf("%s: the body names %q: %s", p.Name, key, w.Body)
			}
		}
	}
}

// COR-1: `corrected {by, at}` (the entry read, the delete Dialog) is the
// entry's latest correction (an update not by the author in the
// creation-day window, not undone), not whoever changed the entry last: an
// author's edit after the correction, a per-section delete after it, or a
// later correction that was undone leave it on the correction.
func TestCorrectedIsTheLatestCorrection(t *testing.T) {
	f := newDailyFixture(t)
	defer f.d.Clock.Set(testdb.Now())
	at := func(h, m int) time.Time { return time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau) }
	corrected := func(e uuid.UUID) *EntryCorrected {
		t.Helper()
		res, err := f.svc.ReadEntry(t.Context(), f.owner, e)
		if err != nil {
			t.Fatal(err)
		}
		return res.Entry.Corrected
	}
	dialog := func(e uuid.UUID) string {
		t.Helper()
		res, err := f.del(f.owner, e, DeleteRequest{DryRun: true})
		if err != nil {
			t.Fatal(err)
		}
		return delDecode[CorrectionPreview](t, res).Effects[0]
	}
	wantIon := func(name string, e uuid.UUID) {
		t.Helper()
		if c := corrected(e); c == nil || c.By != "Ion Rusu" || c.At.String() != "2026-11-18T10:12:00+02:00" {
			t.Errorf("%s: corrected %+v, want Ion Rusu at 10:12", name, c)
		}
		if d := dialog(e); !strings.Contains(d, "corectată de Ion Rusu pe 18.11.2026") {
			t.Errorf("%s: Dialog %q", name, d)
		}
	}
	// Each case on its own Zilnic key (one live value per key and day).
	correct := func(e uuid.UUID, p auth.Principal, code string, n int64) {
		t.Helper()
		v, _ := f.storedValues(e)
		f.mustPut(p, e, f.setMortality(v, code, dbq.SexF, dnum(n)), "Numărătoare refăcută.")
	}

	// The author edits after Ion's correction: updated_by is the vet now.
	edited := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	f.d.Clock.Set(at(10, 12))
	correct(edited, f.ion, "H5.2", 3)
	f.d.Clock.Set(at(11, 0))
	v, _ := f.storedValues(edited)
	f.mustPut(f.vet, edited, f.setMortality(v, "H5.2", dbq.SexF, dnum(2)), "")
	f.d.Clock.Set(testdb.Now())
	if c, by := f.entryState(edited); !c || by == nil || *by != f.vet.UserID {
		t.Fatalf("after the edit: corrected %v, updated_by %v", c, by)
	}
	wantIon("author edit after the correction", edited)

	// The owner deletes one section's values after Ion's correction.
	deleted := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1}, delRow{"H6.1", "F", "mortality", -1})
	f.d.Clock.Set(at(10, 12))
	correct(deleted, f.ion, "H6.1", 3)
	f.d.Clock.Set(at(11, 0))
	section := f.sec["H5.1"]
	if _, err := f.del(f.owner, deleted, DeleteRequest{SectionID: &section, Reason: ptr("Introdusă de două ori.")}); err != nil {
		t.Fatal(err)
	}
	f.d.Clock.Set(testdb.Now())
	wantIon("delete after the correction", deleted)

	// The owner corrects again, then undoes his correction.
	undone := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H6.2", "F", "mortality", -1})
	f.d.Clock.Set(at(10, 12))
	correct(undone, f.ion, "H6.2", 3)
	f.d.Clock.Set(at(11, 0))
	correct(undone, f.owner, "H6.2", 4)
	if _, err := f.undo(f.owner, undone); err != nil {
		t.Fatal(err)
	}
	f.d.Clock.Set(testdb.Now())
	wantIon("an undone correction after it", undone)

	// The author's edit alone is no correction.
	plain := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H6.2", "M", "mortality", -1})
	v, _ = f.storedValues(plain)
	f.mustPut(f.vet, plain, f.setMortality(v, "H6.2", dbq.SexM, dnum(2)), "")
	if c := corrected(plain); c != nil {
		t.Errorf("an author edit: corrected %+v", c)
	}
}
