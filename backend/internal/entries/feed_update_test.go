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
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// feedEntry saves the feed form of 18.11 as p with rows (the create path)
// and returns the entry.
func (f *feedFixture) feedEntry(p auth.Principal, rows ...FeedDailyRow) uuid.UUID {
	f.t.Helper()
	res, err := f.save(p, f.request(nov18, nil, rows...))
	if err != nil {
		f.t.Fatal(err)
	}
	return delDecode[EntryWriteResponse](f.t, res).Entry.ID
}

// version is entry e's current version.
func (f *feedFixture) version(e uuid.UUID) string {
	f.t.Helper()
	var at time.Time
	if err := f.d.App.QueryRow(f.t.Context(), `select coalesce(updated_at, created_at) from entry where id = $1`, e).Scan(&at); err != nil {
		f.t.Fatal(err)
	}
	return version(at)
}

// putFeed sends a Hrană zilnică PUT of e by p with rows at e's version.
func (f *feedFixture) putFeed(p auth.Principal, e uuid.UUID, reason string, dryRun bool, rows ...FeedDailyRow) (Result, error) {
	f.t.Helper()
	req := FeedDailyUpdateRequest{
		IfVersion: f.version(e), Values: FeedDailyValues{FarmID: f.farmA, Date: nov18, Rows: rows}, DryRun: dryRun,
	}
	if reason != "" {
		req.Reason = &reason
	}
	return f.svc.UpdateFeedDaily(f.t.Context(), p, e, req)
}

// feedRow is one feed_allocation row of entry e (live or not).
type feedRow struct {
	ID      uuid.UUID
	Type    uuid.UUID
	G       *decimal.Decimal
	Deleted bool
}

func (f *feedFixture) feedRows(e uuid.UUID, code string, sex dbq.Sex) []feedRow {
	f.t.Helper()
	rows, err := f.d.App.Query(f.t.Context(), `select id, feed_type_id, g_per_bird, deleted_at is not null from feed_allocation
		where entry_id = $1 and section_id = $2 and sex = $3 order by deleted_at nulls first, id`, e, f.sec[code], sex)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	var out []feedRow
	for rows.Next() {
		var r feedRow
		if err := rows.Scan(&r.ID, &r.Type, &r.G, &r.Deleted); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

// sendPut sends PUT /api/entries/{e} as p through updateRoutes.
func (f *feedFixture) sendPut(p auth.Principal, e uuid.UUID, body string) *httptest.ResponseRecorder {
	f.t.Helper()
	mux := http.NewServeMux()
	NewHandler(f.svc).updateRoutes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPut,
		"/api/entries/"+e.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func (f *feedFixture) changesOf(e uuid.UUID) string {
	f.t.Helper()
	var raw []byte
	if err := f.d.App.QueryRow(f.t.Context(), `select changes from audit_log where entry_id = $1 order by at desc, id desc limit 1`, e).
		Scan(&raw); err != nil {
		f.t.Fatal(err)
	}
	var c []audit.Change
	if err := json.Unmarshal(raw, &c); err != nil {
		f.t.Fatal(err)
	}
	return changeJSON(f.t, c)
}

func (f *feedFixture) corrected(e uuid.UUID) bool {
	f.t.Helper()
	var c bool
	if err := f.d.App.QueryRow(f.t.Context(), `select corrected from entry where id = $1`, e).Scan(&c); err != nil {
		f.t.Fatal(err)
	}
	return c
}

// T8 criterion 1, plan M5-D5, D6: the author edits the save on the creation
// day without a Motiv: a changed g in place (same row), a row left out
// soft-deleted; audit `update` same_day_author with the Q12 changes; the
// create's Toast with the kg and stock after the edit; entry.updated only
// (no bird moves); `corrected` stays false.
func TestUpdateFeedDailyAuthorEdit(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.allRows()...)
	h51 := f.feedRows(e, "H5.1", dbq.SexF)
	live0 := len(f.liveTypes())
	rows := f.allRows()[:4]
	rows[0] = f.row("H5.1", dbq.SexF, f.breeder2, "158")
	res, err := f.putFeed(f.ion, e, "ignorat", false, rows...)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	// 158 + 15 + 75 + 7,5 kg; Breeder 2: 1 000 − 158 − 75 = 767 kg.
	if res.Status != http.StatusOK || resp.Toast != "Hrană zilnică salvată · 4 rânduri · 255,5 kg · stoc Breeder 2: 767 kg" ||
		resp.Alive != nil || len(resp.Warnings) != 0 {
		t.Fatalf("%d %s", res.Status, res.Body)
	}
	if got := f.feedRows(e, "H5.1", dbq.SexF); len(got) != 1 || got[0].ID != h51[0].ID || got[0].G.String() != "158" {
		t.Errorf("H5.1 ♀ rows %+v, want the same row with 158", got)
	}
	if got := f.feedRows(e, "H6.1", dbq.SexF); len(got) != 1 || !got[0].Deleted {
		t.Errorf("H6.1 ♀ rows %+v, want soft-deleted", got)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionUpdate || a.Reason != nil || !a.SameDayAuthor || *a.Summary != "Hrană zilnică · 4 din 5 rânduri · 255,5 kg" ||
		!slices.Equal(a.RecordTypes, []string{"hrana"}) {
		t.Errorf("audit %+v", a)
	}
	want := `[{"key":"H5.1 · g găini","section":"H5.1","sex":"F","field":"g","old":160,"new":158},` +
		`{"key":"H6.1 · g găini","section":"H6.1","sex":"F","field":"g","old":157,"new":null},` +
		`{"key":"H6.1 · tip hrană găini","section":"H6.1","sex":"F","field":"feedTypeId","old":"Breeder 2","new":null}]`
	if got := f.changesOf(e); got != want {
		t.Errorf("changes %s", got)
	}
	var old updateSnapshot
	if err := json.Unmarshal(a.OldValues, &old); err != nil || len(old.FeedAllocations) != 5 || old.Corrected {
		t.Errorf("old_values %s (%v)", a.OldValues, err)
	}
	if f.corrected(e) {
		t.Error("an author edit set corrected")
	}
	if got := f.liveTypes()[live0:]; !slices.Equal(got, []string{"entry.updated"}) {
		t.Errorf("live %v", got)
	}
}

// T8 criterion 1, §6.15: a manager corrects another's feed save: the Motiv
// is required (its sentence names the g change), then the correction is
// saved with it and `corrected` set; the corrector's undo puts the g back.
func TestUpdateFeedDailyCorrection(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.allRows()...)
	rows := f.allRows()
	rows[0] = f.row("H5.1", dbq.SexF, f.breeder2, "158")
	_, err := f.putFeed(f.maria, e, "", false, rows...)
	if body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation); len(body.Issues) != 1 || body.Issues[0].Path != "reason" ||
		body.Issues[0].Message != "Motiv lipsă. Scrie de ce se schimbă H5.1 ♀ din 160 în 158 g; fără motiv corecția nu se salvează." {
		t.Errorf("%+v", body)
	}
	res, err := f.putFeed(f.maria, e, "Cântar refăcut.", false, rows...)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție salvată · H5.1 ♀ 160 → 158 g" {
		t.Errorf("%s", res.Body)
	}
	if a := f.latestAudit(e); a.Reason == nil || *a.Reason != "Cântar refăcut." || a.SameDayAuthor {
		t.Errorf("audit %+v", a)
	}
	if !f.corrected(e) {
		t.Error("corrected not set")
	}

	res, err = f.svc.UndoEntry(t.Context(), f.maria, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție anulată · H5.1 ♀ 158 → 160 g" {
		t.Errorf("undo %s", res.Body)
	}
	if got := f.feedRows(e, "H5.1", dbq.SexF); len(got) != 1 || got[0].G.String() != "160" {
		t.Errorf("rows after undo %+v", got)
	}
	if f.corrected(e) {
		t.Error("corrected not put back")
	}
	if a := f.latestAudit(e); a.Action != dbq.AuditActionUndo || a.Reason != nil || !a.SameDayAuthor ||
		*a.Summary != "Hrană zilnică · 5 din 5 rânduri · 288,9 kg" {
		t.Errorf("undo audit %+v", a)
	}
}

// T8 criterion 1, D6: H3 and H2 refuse added and changed rows only (a stored
// row of a type retired since never blocks a g correction); a key another
// entry holds → 409 ALREADY_ENTERED (in the body of a dry run).
func TestUpdateFeedDailyRules(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.row("H5.1", dbq.SexF, f.breeder2, "160"))
	// Breeder 1 is retired: a stored row of it (written before it retired).
	f.exec(`update feed_allocation set feed_type_id = $1 where entry_id = $2`, f.breeder1, e)

	res, err := f.putFeed(f.ion, e, "", false, f.row("H5.1", dbq.SexF, f.breeder1, "158"))
	if err != nil || delDecode[EntryWriteResponse](t, res).Toast != "Hrană zilnică salvată · 1 rând · 158,0 kg · stoc Breeder 1: −158 kg" {
		t.Fatalf("g change on a retired type: %v %s", err, res.Body)
	}

	_, err = f.putFeed(f.ion, e, "", false,
		f.row("H5.1", dbq.SexF, f.breeder1, "158"),
		f.row("H6.1", dbq.SexF, f.cocosi, "157"),
		f.row("H7.1", dbq.SexF, f.breeder2, "150"),
	)
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	got := []string{}
	for _, r := range body.Refusals {
		got = append(got, r.Field+" "+r.Message)
	}
	want := []string{
		"values.rows[1].feedTypeId " + messages.FeedTypeWrongSex("Cocoși", dbq.SexF),
		"values.rows[2].sectionId " + messages.NoBirdsAtDayStart("H7.1", dbq.SexF, nov18),
	}
	if !slices.Equal(got, want) {
		t.Errorf("refusals %q,\nwant %q", got, want)
	}
	if _, err := f.putFeed(f.ion, e, "", false, f.row("H5.1", dbq.SexF, f.breeder2, "158")); err != nil {
		t.Fatalf("a retired type changed to an active one: %v", err)
	}
	_, err = f.putFeed(f.ion, e, "", false, f.row("H5.1", dbq.SexF, f.breeder1, "158"))
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); len(body.Refusals) != 1 ||
		body.Refusals[0].Message != messages.FeedTypeRetired("Breeder 1") || body.Refusals[0].Field != "values.rows[0].feedTypeId" {
		t.Errorf("type change: %+v", body)
	}

	other := f.entry(dbq.EntryKindFeedDaily, f.farmA, nov18, f.vet, at0815)
	f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, 'F', $5, 157)`, other, f.farmA, f.sec["H6.1"], nov18, f.breeder2)
	added := []FeedDailyRow{f.row("H5.1", dbq.SexF, f.breeder2, "158"), f.row("H6.1", dbq.SexF, f.breeder2, "150")}
	_, err = f.putFeed(f.ion, e, "", false, added...)
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered); body.EntryID == nil || *body.EntryID != other ||
		body.Message != messages.FeedAlreadyEntered("H6.1", dbq.SexF, nov18, at0815) {
		t.Errorf("collision %+v", body)
	}
	res, err = f.putFeed(f.ion, e, "", true, added...)
	if err != nil {
		t.Fatal(err)
	}
	if p := delDecode[CorrectionPreview](t, res); len(p.Refusals) != 1 || p.Refusals[0].Field != "values.rows[1]" || len(p.Changes) != 2 {
		t.Errorf("dry run %s", res.Body)
	}
}

// §6.15, §7.0: nothing changed → 422 (in the refusals of a dry run); a stale
// ifVersion → 409; a vet on someone else's save and a viewer → 403.
func TestUpdateFeedDailyRefused(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.allRows()...)
	_, err := f.putFeed(f.ion, e, "", false, f.allRows()...)
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); body.Message != messages.NothingChanged() {
		t.Errorf("%+v", body)
	}
	res, err := f.putFeed(f.ion, e, "", true, f.allRows()...)
	if p := delDecode[CorrectionPreview](t, res); err != nil || len(p.Refusals) != 1 || p.Refusals[0].Message != messages.NothingChanged() {
		t.Errorf("dry run %v %s", err, res.Body)
	}
	req := FeedDailyUpdateRequest{
		IfVersion: "2026-11-18T10:00:00.000001Z", Values: FeedDailyValues{FarmID: f.farmA, Date: nov18, Rows: f.allRows()[:1]},
	}
	_, err = f.svc.UpdateFeedDaily(t.Context(), f.ion, e, req)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)
	for _, p := range []auth.Principal{f.vet, f.viewer} {
		_, err := f.putFeed(p, e, "motiv", false, f.allRows()[:1]...)
		wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	}
}

// T8 criterion 1, Q7: a dry run answers the changes and the warnings of the
// save (H4 on every row, H5 over [date, today]) and writes nothing.
func TestUpdateFeedDailyWarnings(t *testing.T) {
	f := newFeedFixture(t)
	// 300 g × 1 000 hens of Breeder 2 on 16.11 and 17.11: 600 kg before today.
	for _, d := range []timex.Day{nov17.AddDays(-1), nov17} {
		fe := f.entry(dbq.EntryKindFeedDaily, f.farmA, d, f.ion, testdb.Now())
		f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
			values ($1, $2, $3, $4, 'F', $5, 300)`, fe, f.farmA, f.sec["H5.1"], d, f.breeder2)
	}
	e := f.feedEntry(f.ion, f.allRows()...) // Breeder 2 left: 1 000 − 600 − 266,4 = 133,6 kg
	ver := f.version(e)
	rows := f.allRows()
	rows[0] = f.row("H5.1", dbq.SexF, f.breeder2, "300") // + 140 kg → −6,4 kg
	res, err := f.putFeed(f.maria, e, "", true, rows...)
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	codes := []string{}
	for _, w := range p.Warnings {
		codes = append(codes, w.Code)
	}
	if !slices.Equal(codes, []string{httpx.WarnFeedGuide, httpx.WarnStockNegative}) ||
		p.Warnings[1].Message != messages.StockWouldGoNegative("Breeder 2", "Hîrbovăț", decimal.RequireFromString("-6.4")) ||
		len(p.Changes) != 1 || len(p.Refusals) != 0 {
		t.Errorf("preview %s", res.Body)
	}
	if f.version(e) != ver || f.feedRows(e, "H5.1", dbq.SexF)[0].G.String() != "160" {
		t.Error("a dry run wrote")
	}
}

// Q8: the undo of a correction that removed a row and added another puts
// the removed row back (un-deleted, same id) and soft-deletes the added one.
func TestUndoFeedUpdateRows(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.row("H5.1", dbq.SexF, f.breeder2, "160"), f.row("H5.2", dbq.SexF, f.breeder2, "150"))
	h52 := f.feedRows(e, "H5.2", dbq.SexF)[0]
	if _, err := f.putFeed(f.maria, e, "Rând greșit.", false,
		f.row("H5.1", dbq.SexF, f.breeder2, "160"), f.row("H6.1", dbq.SexF, f.breeder2, "157")); err != nil {
		t.Fatal(err)
	}
	res, err := f.svc.UndoEntry(t.Context(), f.maria, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție anulată · 4 valori" {
		t.Errorf("%s", res.Body)
	}
	if got := f.feedRows(e, "H5.2", dbq.SexF); len(got) != 1 || got[0].ID != h52.ID || got[0].Deleted {
		t.Errorf("H5.2 ♀ %+v, want its row back", got)
	}
	if got := f.feedRows(e, "H6.1", dbq.SexF); len(got) != 1 || !got[0].Deleted {
		t.Errorf("H6.1 ♀ %+v, want the added row soft-deleted", got)
	}
}

// The handler: a feed entry decodes its own `values` (unknown members named
// under it), the create's level-1 checks run with "values." paths, then the
// service.
func TestUpdateFeedDailyHandler(t *testing.T) {
	f := newFeedFixture(t)
	e := f.feedEntry(f.ion, f.row("H5.1", dbq.SexF, f.breeder2, "160"))
	body := func(row map[string]any) string {
		b, err := json.Marshal(map[string]any{
			"ifVersion": f.version(e), "reason": "Cântar refăcut.",
			"values": map[string]any{"farmId": f.farmA, "date": "2026-11-18", "rows": []any{row}},
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	row := func(over map[string]any) map[string]any {
		r := map[string]any{"sectionId": f.sec["H5.1"], "sex": "F", "feedTypeId": f.breeder2, "g": 158}
		for k, v := range over {
			r[k] = v
		}
		return r
	}
	for _, tc := range []struct {
		name  string
		row   map[string]any
		issue string
		msg   string
	}{
		{"unknown member", row(map[string]any{"kg": 1}), "values.rows[0].kg", messages.InvalidData()},
		{"H1", row(map[string]any{"g": 350}), "values.rows[0].g", messages.FeedGOutOfRange("H5.1", dbq.SexF, decimal.NewFromInt(350))},
		{"unknown type", row(map[string]any{"feedTypeId": uuid.New()}), "values.rows[0].feedTypeId", messages.InvalidData()},
	} {
		w := f.sendPut(f.ion, e, body(tc.row))
		b := envelopeOf(t, w)
		if w.Code != http.StatusBadRequest || len(b.Issues) == 0 || b.Issues[0].Path != tc.issue || b.Issues[0].Message != tc.msg {
			t.Errorf("%s: %d %s", tc.name, w.Code, w.Body)
		}
	}
	if w := f.sendPut(f.ion, e, body(row(nil))); w.Code != http.StatusOK {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}
