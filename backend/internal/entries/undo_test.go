package entries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func (f *dailyFixture) undo(p auth.Principal, e uuid.UUID) (Result, error) {
	f.t.Helper()
	return f.svc.UndoEntry(f.t.Context(), p, e)
}

// mustPut is a PUT of e by p that must succeed.
func (f *dailyFixture) mustPut(p auth.Principal, e uuid.UUID, v DailyValues, reason string) {
	f.t.Helper()
	if _, err := f.put(p, e, v, reason, false); err != nil {
		f.t.Fatal(err)
	}
}

// T6 criterion 4, Q8: the corrector undoes his correction of today while it
// is still the entry's latest action: the row back in place, `corrected`
// back, audit `undo` without a Motiv, entry.updated + counts.changed, the
// Toast naming the value as the undo sets it. Someone else gets 403; a
// second undo (the update is no longer the latest action) 409 STALE.
func TestUndoCorrection(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	row := f.birdRows(e, "H5.2", dbq.SexF)[0]
	v, _ := f.storedValues(e)
	f.mustPut(f.ion, e, f.setMortality(v, "H5.2", dbq.SexF, dnum(3)), "Numărătoare refăcută.")

	_, err := f.undo(f.vet, e)
	if body := wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden); body.Message != messages.PageForbidden() {
		t.Errorf("message %q", body.Message)
	}
	_, err = f.undo(f.viewer, e)
	wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)

	res, err := f.undo(f.ion, e)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if res.Status != http.StatusOK || resp.Toast != "Corecție anulată · H5.2 ♀ 3 → 1" || resp.Alive == nil || resp.Alive.F != 1799 {
		t.Errorf("%d %s", res.Status, res.Body)
	}
	if got := f.birdRows(e, "H5.2", dbq.SexF); len(got) != 1 || got[0].ID != row.ID || got[0].Qty != -1 {
		t.Errorf("rows %+v", got)
	}
	if corrected, _ := f.entryState(e); corrected {
		t.Error("corrected stays true")
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionUndo || a.Reason != nil || !a.SameDayAuthor || *a.Summary != "Mortalitate · 1 secțiune" ||
		!slices.Equal(a.RecordTypes, []string{"mortalitate"}) {
		t.Errorf("audit %+v", a)
	}
	if got := changeJSON(t, f.auditChanges(e)); got != `[{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":3,"new":1}]` {
		t.Errorf("changes %s", got)
	}
	if got := f.liveTypes(); !slices.Equal(got, []string{"entry.updated", "counts.changed", "entry.updated", "counts.changed"}) {
		t.Errorf("live = %v", got)
	}
	_, err = f.undo(f.ion, e)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)
}

// T6 criterion 4: an undone edit puts back exactly the rows before it, by
// id: a value it changed in place, a value it cleared (its row un-deleted),
// a value it added (soft-deleted).
func TestUndoEditRows(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1}, delRow{"H5.2", "F", "mortality", -2})
	h51, h52 := f.birdRows(e, "H5.1", dbq.SexF)[0], f.birdRows(e, "H5.2", dbq.SexF)[0]
	v, _ := f.storedValues(e)
	f.mustPut(f.vet, e, f.setMortality(f.setMortality(f.setMortality(v, "H5.2", dbq.SexF, nil), "H5.1", dbq.SexF, dnum(3)), "H6.1", dbq.SexF, dnum(1)), "")

	res, err := f.undo(f.vet, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție anulată · 3 valori" || resp.Alive == nil || resp.Alive.F != 1797 {
		t.Errorf("%s", res.Body)
	}
	if got := f.birdRows(e, "H5.1", dbq.SexF); len(got) != 1 || got[0].ID != h51.ID || got[0].Qty != -1 {
		t.Errorf("H5.1 %+v", got)
	}
	if got := f.birdRows(e, "H5.2", dbq.SexF); len(got) != 1 || got[0].ID != h52.ID || got[0].Deleted || got[0].Qty != -2 {
		t.Errorf("H5.2 %+v", got)
	}
	if got := f.birdRows(e, "H6.1", dbq.SexF); len(got) != 1 || !got[0].Deleted {
		t.Errorf("H6.1 %+v", got)
	}
	if got, _ := f.storedValues(e); len(got.Sections) != 2 {
		t.Errorf("values %+v", got)
	}
}

// T6 criterion 4: an update of an earlier day can't be undone ("Nu se mai
// poate anula: ziua salvării s-a încheiat."); nor an update without its
// snapshot (the imported correction) → 409 STALE.
func TestUndoUpdateRefusals(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	v, _ := f.storedValues(e)
	f.mustPut(f.ion, e, f.setMortality(v, "H5.2", dbq.SexF, dnum(3)), "motiv")
	f.d.Clock.Set(time.Date(2026, time.November, 19, 0, 0, 30, 0, timex.Chisinau))
	_, err := f.undo(f.ion, e)
	f.d.Clock.Set(testdb.Now())
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeStale); body.Message != messages.UndoDayEnded() {
		t.Errorf("message %q", body.Message)
	}

	imported := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H5.1", "F", "mortality", -1})
	f.auditRow(imported, dbq.AuditActionUpdate, f.ion, at0815)
	_, err = f.undo(f.ion, imported)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)
}

// T6 criterion 4 via undo.go: the caller's latest action a create → the
// undo of the create (T7's undoCreate).
func TestUndoDispatchesCreate(t *testing.T) {
	f := newDailyFixture(t)
	res, err := f.saveDaily(f.vet, f.values(nov18, []DailySection{{SectionID: f.sec["H5.1"], F: &DailySex{Mortality: dnum(1)}}}), httpx.WriteMeta{})
	if err != nil {
		t.Fatal(err)
	}
	e := delDecode[EntryWriteResponse](t, res).Entry.ID
	res, err = f.undo(f.vet, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != messages.DailyUndone(nov18) {
		t.Errorf("%s", res.Body)
	}
	if _, deleted, _ := f.liveRows(e); !deleted {
		t.Error("entry not deleted")
	}
}

// Q8: a value the update cleared whose key another entry took meanwhile →
// 409 ALREADY_ENTERED naming that entry; §6.1 is re-checked ("Ștergerea ar
// lăsa …").
func TestUndoUpdateConflicts(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1}, delRow{"H5.2", "F", "mortality", -2})
	v, _ := f.storedValues(e)
	f.mustPut(f.vet, e, f.setMortality(v, "H5.2", dbq.SexF, nil), "")
	holder := f.delDaily(f.ion, nov18, at0815.Add(time.Hour), 0, delRow{"H5.2", "F", "mortality", -4})
	_, err := f.undo(f.vet, e)
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.EntryID == nil || *body.EntryID != holder || body.Message != messages.DailyAlreadyEntered("H5.2", dbq.SexF, nov18, at0815.Add(time.Hour)) {
		t.Errorf("envelope %+v", body)
	}

	past := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H6.2", "M", "mortality", -3})
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	f.bird(sale, "H6.2", nov18, "M", "sale", -7) // H6.2 ♂ ends 18.11 with 0
	pv, _ := f.storedValues(past)
	f.mustPut(f.ion, past, f.setMortality(pv, "H6.2", dbq.SexM, dnum(1)), "motiv") // … with 2
	f.bird(sale, "H6.2", nov18, "M", "sale", -2)                                   // … with 0 again
	_, err = f.undo(f.ion, past)
	body = wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != "Ștergerea ar lăsa H6.2 cu −2 cocoși azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." {
		t.Errorf("message %q", body.Message)
	}
	if got := f.birdRows(past, "H6.2", dbq.SexM); got[0].Qty != -1 {
		t.Errorf("the refused undo wrote: %+v", got)
	}
}

// The route: POST /api/entries/{id}/undo takes {} only; a viewer 403.
func TestUndoHandler(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	v, _ := f.storedValues(e)
	f.mustPut(f.vet, e, f.setMortality(v, "H5.2", dbq.SexF, dnum(2)), "")
	post := func(p auth.Principal, body string) *httptest.ResponseRecorder {
		mux := http.NewServeMux()
		NewHandler(f.svc).undoRoutes(mux, httpx.NewChain())
		req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPost,
			"/api/entries/"+e.String()+"/undo", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	code := func(w *httptest.ResponseRecorder) string {
		var b httpx.ErrorBody
		if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
			t.Fatalf("%d %s", w.Code, w.Body)
		}
		return b.Error.Code
	}
	if w := post(f.vet, `{"x":1}`); w.Code != http.StatusBadRequest || code(w) != httpx.CodeValidation {
		t.Errorf("%d %s", w.Code, w.Body)
	}
	if w := post(f.viewer, `{}`); w.Code != http.StatusForbidden || code(w) != httpx.CodeForbidden {
		t.Errorf("%d %s", w.Code, w.Body)
	}
	if w := post(f.vet, `{}`); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"toast":"Corecție anulată · H5.2 ♀ 2 → 1"`) {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}

// SEC-2: an undo or a restore also needs the caller's current rights to
// allow the action undone: Ion, a manager, corrects the vet's entry and
// deletes one section of another with a Motiv, then is demoted to vet the
// same day: undo and restore → 403 FORBIDDEN, and the entry read offers
// neither. As a manager again he may (the rows were left as they were).
func TestUndoRestoreNeedCurrentRights(t *testing.T) {
	f := newDailyFixture(t)
	demoted := f.ion
	demoted.Role = auth.Vet
	read := func(p auth.Principal, e uuid.UUID) EntryReadActions {
		t.Helper()
		res, err := f.svc.ReadEntry(t.Context(), p, e)
		if err != nil {
			t.Fatal(err)
		}
		return res.Actions
	}

	corrected := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	v, _ := f.storedValues(corrected)
	f.mustPut(f.ion, corrected, f.setMortality(v, "H5.2", dbq.SexF, dnum(3)), "Numărătoare refăcută.")
	if got := read(f.ion, corrected); !got.Undo {
		t.Errorf("the manager: actions %+v, want undo", got)
	}
	if got := read(demoted, corrected); got.Undo || got.Correct || got.Delete {
		t.Errorf("demoted: actions %+v, want none", got)
	}
	_, err := f.undo(demoted, corrected)
	if body := wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden); body.Message != messages.PageForbidden() {
		t.Errorf("message %q", body.Message)
	}
	if _, err := f.undo(f.ion, corrected); err != nil {
		t.Errorf("the manager's undo: %v", err)
	}

	split := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1}, delRow{"H6.1", "F", "mortality", -1})
	section := f.sec["H5.1"]
	if _, err := f.del(f.ion, split, DeleteRequest{SectionID: &section, Reason: ptr("Introdusă de două ori.")}); err != nil {
		t.Fatal(err)
	}
	if got := read(f.ion, split); !got.Restore {
		t.Errorf("the manager: actions %+v, want restore", got)
	}
	if got := read(demoted, split); got.Restore {
		t.Errorf("demoted: actions %+v, want no restore", got)
	}
	_, err = f.svc.RestoreEntry(t.Context(), demoted, split)
	wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	if _, err := f.svc.RestoreEntry(t.Context(), f.ion, split); err != nil {
		t.Errorf("the manager's restore: %v", err)
	}

	// The author's own action needs only the author's window: the vet
	// undoes his create of today.
	own := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H6.2", "F", "mortality", -1})
	f.auditRow(own, dbq.AuditActionCreate, f.vet, at0815)
	if got := read(f.vet, own); !got.Undo {
		t.Errorf("the author: actions %+v, want undo", got)
	}
	if _, err := f.undo(f.vet, own); err != nil {
		t.Errorf("the author's undo: %v", err)
	}
}
