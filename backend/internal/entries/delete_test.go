package entries

import (
	"bytes"
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

// delRow is one bird row of a test Zilnic entry.
type delRow struct {
	code, sex, typ string
	qty            int
}

// delDaily inserts a daily entry of p on day, created at at, with rows
// (a cull gets the reason boala) and, when eggs > 0, an H5 egg row whose
// four categories add up to eggs.
func (f *feedFixture) delDaily(p auth.Principal, day timex.Day, at time.Time, eggs int, rows ...delRow) uuid.UUID {
	f.t.Helper()
	e := f.entry(dbq.EntryKindDaily, f.farmA, day, p, at)
	for _, r := range rows {
		var reason *string
		if r.typ == "cull" {
			s := "boala"
			reason = &s
		}
		f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
			values ($1, $2, $3, $4, $5, $6, $7, $8)`, e, f.farmA, f.sec[r.code], day, r.sex, r.typ, r.qty, reason)
	}
	if eggs > 0 {
		f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs)
			values ($1, $2, $3, $4, 3, $5, 0, 0, 0)`, e, f.farmA, f.sec["H5"], day, eggs)
	}
	return e
}

func (f *feedFixture) del(p auth.Principal, id uuid.UUID, req DeleteRequest) (Result, error) {
	f.t.Helper()
	return f.svc.DeleteEntry(f.t.Context(), p, id, req)
}

func delDecode[T any](t *testing.T, res Result) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(res.Body, &v); err != nil {
		t.Fatalf("%d %s: %v", res.Status, res.Body, err)
	}
	return v
}

// delAudit is an entry's latest audit row.
type delAudit struct {
	Action        dbq.AuditAction
	Reason        *string
	SameDayAuthor bool
	Summary       *string
	RecordTypes   []string
	OldValues     []byte
	NewValues     []byte
}

func (f *feedFixture) latestAudit(e uuid.UUID) delAudit {
	f.t.Helper()
	var a delAudit
	if err := f.d.App.QueryRow(f.t.Context(), `select action, reason, same_day_author, summary, record_types,
		old_values, new_values from audit_log where entry_id = $1 order by at desc, id desc limit 1`, e).
		Scan(&a.Action, &a.Reason, &a.SameDayAuthor, &a.Summary, &a.RecordTypes, &a.OldValues, &a.NewValues); err != nil {
		f.t.Fatal(err)
	}
	return a
}

// liveTypes are the live_event types written, oldest first.
func (f *feedFixture) liveTypes() []string {
	f.t.Helper()
	rows, err := f.d.App.Query(f.t.Context(), `select type from live_event order by id`)
	if err != nil {
		f.t.Fatal(err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			f.t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

// liveRows counts the entry's live child rows and reports whether the entry
// is deleted, with its delete reason.
func (f *feedFixture) liveRows(e uuid.UUID) (rows int, deleted bool, reason *string) {
	f.t.Helper()
	if err := f.d.App.QueryRow(f.t.Context(), `select
		(select count(*) from bird_event where entry_id = $1 and deleted_at is null)
		+ (select count(*) from egg_production where entry_id = $1 and deleted_at is null)
		+ (select count(*) from feed_allocation where entry_id = $1 and deleted_at is null)
		+ (select count(*) from feed_delivery where entry_id = $1 and deleted_at is null),
		deleted_at is not null, delete_reason from entry where id = $1`, e).Scan(&rows, &deleted, &reason); err != nil {
		f.t.Fatal(err)
	}
	return rows, deleted, reason
}

func ptr[T any](v T) *T { return &v }

// T7 criterion 1: the author deletes the whole save on the creation day
// without a Motiv (a sent one is ignored): every row and the entry
// soft-deleted at the clock's instant, the version moved, audit `delete`
// same_day_author with the snapshot, live entry.deleted + counts.changed.
func TestDeleteByAuthor(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 1500, delRow{"H5.1", "F", "mortality", -2}, delRow{"H5.2", "M", "cull", -1})
	res, err := f.del(f.vet, e, DeleteRequest{Reason: ptr("ignorat")})
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	// Farm at the end of 18.11 before: ♀ 1000−2 + 500 + 200 + 0 = 1 698, ♂ 100 + 50−1 = 149.
	if res.Status != http.StatusOK || resp.Toast != messages.MortalityDeleted(nov18, 2, ptr(1700), ptr(150)) ||
		resp.Alive == nil || *resp.Alive != (httpx.SexPair[int]{F: 1700, M: 150}) || resp.Entry.Version != "2026-11-18T10:00:00.000000Z" ||
		len(resp.Warnings) != 0 {
		t.Fatalf("%d %s", res.Status, res.Body)
	}
	if resp.Toast != "Mortalitate ștearsă · 18.11.2026 · 2 secțiuni · în viață azi ♀ 1\u202f700 · ♂ 150" {
		t.Errorf("toast %q", resp.Toast)
	}
	if n, deleted, reason := f.liveRows(e); n != 0 || !deleted || reason != nil {
		t.Errorf("live rows %d, deleted %v, reason %v", n, deleted, reason)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionDelete || a.Reason != nil || !a.SameDayAuthor || a.Summary == nil ||
		*a.Summary != "Mortalitate · 2 secțiuni · Ouă · 1 hală" ||
		!slices.Equal(a.RecordTypes, []string{"mortalitate", "sacrificare", "oua"}) || a.NewValues != nil {
		t.Errorf("audit %+v (summary %v)", a, a.Summary)
	}
	var snap removedRows
	if err := json.Unmarshal(a.OldValues, &snap); err != nil {
		t.Fatal(err)
	}
	if len(snap.BirdEvents) != 2 || len(snap.EggProduction) != 1 || snap.SectionID != nil || snap.BirdEvents[0].ID == uuid.Nil ||
		snap.BirdEvents[1].CullReason == nil || *snap.BirdEvents[1].CullReason != dbq.CullReasonBoala {
		t.Errorf("snapshot %s", a.OldValues)
	}
	if got := f.liveTypes(); !slices.Equal(got, []string{"entry.deleted", "counts.changed"}) {
		t.Errorf("live = %v", got)
	}
	// The entry is gone for every change but a restore.
	_, err = f.del(f.vet, e, DeleteRequest{})
	wantCode(t, err, http.StatusNotFound, httpx.CodeNotFound)
}

// T7 criterion 1, plan M5-D5, Q10: a manager on someone else's entry and the
// owner delete with a Motiv (missing → 400 on `reason`, not asked of a dry
// run); a vet on someone else's entry and a viewer get 403 "Nu ai acces …",
// a vet on own entry of an earlier day the S16 sentence; another farm 403
// FARM_FORBIDDEN.
func TestDeleteAccess(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -2})
	ion := f.delDaily(f.ion, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	past := f.delDaily(f.vet, nov17, at0815.AddDate(0, 0, -1), 0, delRow{"H6.1", "F", "mortality", -1})

	for _, tc := range []struct {
		name    string
		p       auth.Principal
		id      uuid.UUID
		status  int
		code    string
		message string
	}{
		{"manager without Motiv", f.ion, e, 400, httpx.CodeValidation, messages.DeleteReasonMissing()},
		{"viewer", f.viewer, e, 403, httpx.CodeForbidden, messages.PageForbidden()},
		{"vet, someone else's", f.vet, ion, 403, httpx.CodeForbidden, messages.PageForbidden()},
		{"vet, own of an earlier day", f.vet, past, 403, httpx.CodeForbidden, messages.OwnEntryPastDay()},
		{"another farm", f.outsider, e, 403, httpx.CodeFarmForbidden, messages.FarmForbidden()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.del(tc.p, tc.id, DeleteRequest{})
			body := wantCode(t, err, tc.status, tc.code)
			if body.Message != tc.message {
				t.Errorf("message %q", body.Message)
			}
			if tc.code == httpx.CodeValidation && (len(body.Issues) != 1 || body.Issues[0].Path != "reason") {
				t.Errorf("issues %+v", body.Issues)
			}
		})
	}
	if msg := messages.DeleteReasonMissing(); msg != "Motiv lipsă. Scrie de ce ștergi salvarea; fără motiv nu se șterge." {
		t.Errorf("Q9 sentence %q", msg)
	}

	// A dry run without a Motiv previews; the delete with one is saved with it.
	res, err := f.del(f.ion, e, DeleteRequest{DryRun: true})
	if err != nil || res.Status != http.StatusOK {
		t.Fatalf("dry run: %v", err)
	}
	if n, deleted, _ := f.liveRows(e); n != 1 || deleted {
		t.Fatal("the dry run wrote")
	}
	if _, err := f.del(f.ion, e, DeleteRequest{Reason: ptr("  Dublură a salvării de la 08:10.  ")}); err != nil {
		t.Fatal(err)
	}
	_, deleted, reason := f.liveRows(e)
	a := f.latestAudit(e)
	if !deleted || reason == nil || *reason != "Dublură a salvării de la 08:10." || a.Reason == nil || *a.Reason != *reason || a.SameDayAuthor {
		t.Errorf("deleted %v reason %v, audit %+v", deleted, reason, a)
	}
	if _, err := f.del(f.owner, past, DeleteRequest{Reason: ptr("Greșit.")}); err != nil {
		t.Errorf("owner: %v", err)
	}
}

// T7 criterion 1: a dry run answers the Dialog sentence as effects[0], the
// alive rows and the checked days, and writes nothing (no version, audit or
// live row).
func TestDeleteDryRun(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov17, at0815.AddDate(0, 0, -1), 0, delRow{"H5.1", "F", "mortality", -2}, delRow{"H5.1", "M", "mortality", 0})
	res, err := f.del(f.ion, e, DeleteRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	want := "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:15: 1 secțiune · ♀ −2 · ♂ 0. Găini în viață azi în " +
		"Hîrbovăț: ♀ 1\u202f698 → ♀ 1\u202f700. Ștergerea rămâne în Istoric modificări."
	wantAlive := []PreviewAlive{
		{Section: "H5.1", Sex: dbq.SexF, Date: nov17, Before: 998, After: 1000},
		{Section: "H5.1", Sex: dbq.SexF, Date: nov18, Before: 998, After: 1000},
	}
	if len(p.Effects) != 1 || p.Effects[0] != want || !slices.Equal(p.Alive, wantAlive) || p.CheckedDays != 2 ||
		len(p.Changes) != 0 || len(p.Refusals) != 0 || len(p.Warnings) != 0 {
		t.Errorf("preview %s", res.Body)
	}
	if !bytes.Contains(res.Body, []byte(`"changes":[]`)) || !bytes.Contains(res.Body, []byte(`"refusals":[]`)) {
		t.Errorf("null lists: %s", res.Body)
	}
	if n, deleted, _ := f.liveRows(e); n != 2 || deleted {
		t.Error("the dry run wrote")
	}
	if got := f.rows(); got[1] != 0 || got[2] != 0 {
		t.Errorf("audit/live rows %v", got)
	}
}

// T7 criterion 1, D98, S15: sectionId deletes one compartment's bird values
// (an egg unit's eggs); the entry stays until its last live row goes; a
// section outside the save or on another kind is 400 on `sectionId`.
func TestDeleteSection(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 1200, delRow{"H5.1", "F", "mortality", -2},
		delRow{"H5.2", "F", "mortality", -1}, delRow{"H5.2", "M", "cull", -1})
	feed := f.entry(dbq.EntryKindFeedDaily, f.farmA, nov18, f.vet, at0815)
	for _, tc := range []struct {
		id      uuid.UUID
		section string
	}{{e, "H6.1"}, {e, "H7"}, {feed, "H5.1"}} {
		_, err := f.del(f.vet, tc.id, DeleteRequest{SectionID: ptr(f.sec[tc.section])})
		if body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation); len(body.Issues) != 1 || body.Issues[0].Path != "sectionId" {
			t.Errorf("%s: %+v", tc.section, body.Issues)
		}
	}

	h52 := f.sec["H5.2"]
	res, err := f.del(f.vet, e, DeleteRequest{SectionID: &h52, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	if want := "Se șterge rândul salvat la 08:15: ♀ −1 · ♂ −1. În viață în H5.2: ♀ 499 → 500 · ♂ 49 → 50. Ștergerea " +
		"rămâne în Istoric modificări."; len(p.Effects) != 1 || p.Effects[0] != want || p.CheckedDays != 1 || len(p.Alive) != 2 {
		t.Errorf("preview %s", res.Body)
	}

	res, err = f.del(f.vet, e, DeleteRequest{SectionID: &h52})
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Mortalitate ștearsă · H5.2 · în viață ♀ 500 · ♂ 50" || resp.Alive == nil || resp.Alive.F != 1698 {
		t.Errorf("%s", res.Body)
	}
	if n, deleted, _ := f.liveRows(e); n != 2 || deleted {
		t.Errorf("after H5.2: %d rows, deleted %v", n, deleted)
	}
	a := f.latestAudit(e)
	var snap removedRows
	if err := json.Unmarshal(a.OldValues, &snap); err != nil {
		t.Fatal(err)
	}
	if *a.Summary != "Mortalitate · H5.2" || snap.SectionID == nil || *snap.SectionID != h52 || len(snap.BirdEvents) != 2 ||
		len(snap.EggProduction) != 0 {
		t.Errorf("audit %s %s", *a.Summary, a.OldValues)
	}

	res, err = f.del(f.vet, e, DeleteRequest{SectionID: ptr(f.sec["H5"])})
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Ouă șterse · 18.11.2026 · H5" || resp.Alive != nil {
		t.Errorf("%s", res.Body)
	}
	if a := f.latestAudit(e); *a.Summary != "Ouă · H5" || !slices.Equal(a.RecordTypes, []string{"oua"}) {
		t.Errorf("audit %+v", a)
	}

	// The last live row: the entry goes with it.
	if _, err := f.del(f.vet, e, DeleteRequest{SectionID: ptr(f.sec["H5.1"])}); err != nil {
		t.Fatal(err)
	}
	if n, deleted, _ := f.liveRows(e); n != 0 || !deleted {
		t.Errorf("after H5.1: %d rows, deleted %v", n, deleted)
	}
}

// T7 criterion 1, §6.1 (Q4): removing a positive sexing error that a later
// sale relied on would leave H6.1 ♀ at −5 today: 422 with the "Ștergerea
// ar lăsa …" banner and the negative days; a dry run gets it in refusals;
// nothing is written.
func TestDeleteBelowZero(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H6.1", "F", "sexing_error", 5})
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	f.bird(sale, "H6.1", nov18, "F", "sale", -205)

	banner := "Ștergerea ar lăsa H6.1 cu −5 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0."
	_, err := f.del(f.vet, e, DeleteRequest{})
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != banner || len(body.Refusals) != 1 || body.Refusals[0].MinAliveEnd == nil || *body.Refusals[0].MinAliveEnd != -5 ||
		*body.Refusals[0].FirstNegativeDate != nov18 || body.Refusals[0].SectionCode != "H6.1" || body.Refusals[0].Sex != "F" {
		t.Errorf("refused %+v", body)
	}
	res, err := f.del(f.vet, e, DeleteRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if p := delDecode[CorrectionPreview](t, res); len(p.Refusals) != 1 || p.Refusals[0].Message != banner {
		t.Errorf("dry run %s", res.Body)
	}
	if n, deleted, _ := f.liveRows(e); n != 1 || deleted || f.count("audit_log") != 0 {
		t.Error("a refused delete wrote")
	}
}

// T7 criterion 3 (Q8): the author undoes the create of today while it is
// still the entry's latest action: soft delete, audit `undo`
// same_day_author, live entry.deleted. Someone else → 403; a later change →
// 409 STALE; after 23:59 → 409 "Nu se mai poate anula: ziua salvării s-a
// încheiat.".
func TestUndoCreate(t *testing.T) {
	f := newFeedFixture(t)
	ctx := t.Context()
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -2})
	f.auditRow(e, dbq.AuditActionCreate, f.vet, at0815)

	for _, p := range []auth.Principal{f.ion, f.viewer} {
		_, err := f.svc.UndoCreate(ctx, p, e)
		wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	}
	res, err := f.svc.UndoCreate(ctx, f.vet, e)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Zilnic anulat · 18.11.2026" || resp.Alive == nil || *resp.Alive != (httpx.SexPair[int]{F: 1700, M: 150}) {
		t.Errorf("%s", res.Body)
	}
	if n, deleted, _ := f.liveRows(e); n != 0 || !deleted {
		t.Error("not deleted")
	}
	if a := f.latestAudit(e); a.Action != dbq.AuditActionUndo || !a.SameDayAuthor || a.Reason != nil || a.OldValues == nil {
		t.Errorf("audit %+v", a)
	}
	if got := f.liveTypes(); !slices.Equal(got, []string{"entry.deleted", "counts.changed"}) {
		t.Errorf("live %v", got)
	}
	// An undo is not a delete: nothing to restore.
	_, err = f.svc.RestoreEntry(ctx, f.vet, e)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)

	// A later change by anyone.
	changed := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	f.auditRow(changed, dbq.AuditActionCreate, f.vet, at0815)
	f.auditRow(changed, dbq.AuditActionUpdate, f.ion, at0815.Add(time.Hour))
	_, err = f.svc.UndoCreate(ctx, f.vet, changed)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)

	// The next day.
	late := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H6.1", "F", "mortality", -1})
	f.auditRow(late, dbq.AuditActionCreate, f.vet, at0815)
	f.d.Clock.Set(time.Date(2026, time.November, 19, 0, 0, 30, 0, timex.Chisinau))
	defer f.d.Clock.Set(testdb.Now())
	_, err = f.svc.UndoCreate(ctx, f.vet, late)
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeStale); body.Message != "Nu se mai poate anula: ziua salvării s-a încheiat." {
		t.Errorf("message %q", body.Message)
	}
}

// T7 criterion 4 (Q8): only the deleter restores, the same day, while the
// delete is the entry's latest action; the rows come back with the entry,
// audit `restore` and live entry.restored.
func TestRestore(t *testing.T) {
	f := newFeedFixture(t)
	ctx := t.Context()
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -2})
	_, err := f.svc.RestoreEntry(ctx, f.vet, e) // never deleted
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)
	if _, err := f.del(f.vet, e, DeleteRequest{}); err != nil {
		t.Fatal(err)
	}
	for _, p := range []auth.Principal{f.ion, f.owner} {
		_, err := f.svc.RestoreEntry(ctx, p, e)
		wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	}
	_, err = f.svc.RestoreEntry(ctx, f.viewer, e)
	wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)

	live0 := len(f.liveTypes())
	res, err := f.svc.RestoreEntry(ctx, f.vet, e)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Mortalitate restaurată · 18.11.2026 · 1 secțiune · în viață azi ♀ 1\u202f698" ||
		*resp.Alive != (httpx.SexPair[int]{F: 1698, M: 150}) || resp.Entry.Version != "2026-11-18T10:00:00.000001Z" {
		t.Errorf("%s", res.Body)
	}
	if n, deleted, reason := f.liveRows(e); n != 1 || deleted || reason != nil {
		t.Errorf("rows %d deleted %v", n, deleted)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionRestore || !a.SameDayAuthor || a.NewValues == nil || *a.Summary != "Mortalitate · 1 secțiune" {
		t.Errorf("audit %+v", a)
	}
	if got := f.liveTypes()[live0:]; !slices.Equal(got, []string{"entry.restored", "counts.changed"}) {
		t.Errorf("live %v", got)
	}
	// Restored once: the restore is now the latest action.
	_, err = f.svc.RestoreEntry(ctx, f.vet, e)
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)

	// A delete of an earlier day.
	if _, err := f.del(f.vet, e, DeleteRequest{}); err != nil {
		t.Fatal(err)
	}
	f.d.Clock.Set(time.Date(2026, time.November, 19, 9, 0, 0, 0, timex.Chisinau))
	defer f.d.Clock.Set(testdb.Now())
	_, err = f.svc.RestoreEntry(ctx, f.vet, e)
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeStale); body.Message != messages.UndoDayEnded() {
		t.Errorf("message %q", body.Message)
	}
}

// T7 criterion 4: a per-section restore whose key another entry took
// meanwhile → 409 ALREADY_ENTERED naming that entry; nothing comes back.
func TestRestoreKeyRetaken(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -2}, delRow{"H5.2", "F", "mortality", -1})
	if _, err := f.del(f.vet, e, DeleteRequest{SectionID: ptr(f.sec["H5.1"])}); err != nil {
		t.Fatal(err)
	}
	at := at0815.Add(2 * time.Hour)
	holder := f.delDaily(f.ion, nov18, at, 0, delRow{"H5.1", "F", "mortality", -3})
	_, err := f.svc.RestoreEntry(t.Context(), f.vet, e)
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.EntryID == nil || *body.EntryID != holder || body.Message != messages.DailyAlreadyEntered("H5.1", dbq.SexF, nov18, at) {
		t.Errorf("%+v", body)
	}
	if body.Message != "H5.1 ♀ are deja valori introduse pe 18.11.2026 la 10:15. Apasă Editează ca să schimbi valorile." {
		t.Errorf("message %q", body.Message)
	}
	if n, _, _ := f.liveRows(e); n != 1 {
		t.Errorf("live rows %d", n)
	}
}

// T7 criterion 4: §6.1 is re-checked on restore. Ion deletes his own save of
// 17.11 (H5.2 ♀ 2 dead), a sale then takes all 500 H5.2 hens on 18.11, and
// putting the 2 dead back would leave −2 today: 422 "Restaurarea ar lăsa …".
func TestRestoreBelowZero(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.ion, nov17, testdb.Now(), 0, delRow{"H5.2", "F", "mortality", -2})
	if _, err := f.del(f.ion, e, DeleteRequest{}); err != nil {
		t.Fatal(err)
	}
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, testdb.Now())
	f.bird(sale, "H5.2", nov18, "F", "sale", -500)
	_, err := f.svc.RestoreEntry(t.Context(), f.ion, e)
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != "Restaurarea ar lăsa H5.2 cu −2 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." ||
		len(body.Refusals) != 1 || *body.Refusals[0].FirstNegativeDate != nov18 {
		t.Errorf("%+v", body)
	}
	if _, deleted, _ := f.liveRows(e); !deleted {
		t.Error("a refused restore wrote")
	}
}

// T7 criterion 4, plan M5-D7: rows an earlier edit superseded stay deleted
// after a restore, even when (frozen clock) their deleted_at equals the
// delete's: the restore takes exactly the delete snapshot's ids.
func TestRestoreKeepsSupersededRows(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.ion, nov18, testdb.Now(), 0, delRow{"H5.1", "F", "mortality", -1})
	old := f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, deleted_at)
		values ($1, $2, $3, $4, 'F', 'mortality', -3, $5) returning id`, e, f.farmA, f.sec["H5.2"], nov18, testdb.Now())
	f.bird(e, "H5.2", nov18, "F", "mortality", -4) // the edit's new value
	if _, err := f.del(f.ion, e, DeleteRequest{}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.RestoreEntry(t.Context(), f.ion, e); err != nil {
		t.Fatal(err)
	}
	var oldDeleted bool
	var live int
	if err := f.d.App.QueryRow(t.Context(), `select (select deleted_at is not null from bird_event where id = $1),
		(select count(*) from bird_event where entry_id = $2 and deleted_at is null)`, old, e).Scan(&oldDeleted, &live); err != nil {
		t.Fatal(err)
	}
	if !oldDeleted || live != 2 {
		t.Errorf("superseded row deleted %v, live rows %d", oldDeleted, live)
	}
}

// T7 criterion 5: a Hrană zilnică delete and restore name the type with the
// most kg and its stock at the end of today; the Dialog names rows and kg;
// no counts.changed.
func TestDeleteFeedDaily(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.save(f.ion, f.request(nov18, nil, f.allRows()...))
	if err != nil {
		t.Fatal(err)
	}
	var created struct {
		Entry httpx.EntryRef `json:"entry"`
	}
	if err := json.Unmarshal(res.Body, &created); err != nil {
		t.Fatal(err)
	}
	e := created.Entry.ID
	res, err = f.del(f.ion, e, DeleteRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if p := delDecode[CorrectionPreview](t, res); len(p.Effects) != 1 || p.Effects[0] != "Se șterge salvarea Hrană zilnică "+
		"introdusă de Ion Rusu la 12:00: 5 rânduri · 288,9 kg. Ștergerea rămâne în Istoric modificări." || p.CheckedDays != 0 {
		t.Errorf("preview %s", res.Body)
	}
	live0 := len(f.liveTypes())
	res, err = f.del(f.ion, e, DeleteRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Hrană zilnică ștearsă · 18.11.2026 · 5 rânduri · stoc Breeder 2: 1\u202f000 kg" ||
		resp.Alive != nil {
		t.Errorf("%s", res.Body)
	}
	if a := f.latestAudit(e); *a.Summary != "Hrană zilnică · 5 din 5 rânduri · 288,9 kg" || !slices.Equal(a.RecordTypes, []string{"hrana"}) {
		t.Errorf("audit %+v", a)
	}
	if got := f.liveTypes()[live0:]; !slices.Equal(got, []string{"entry.deleted"}) {
		t.Errorf("live %v", got)
	}
	res, err = f.svc.RestoreEntry(t.Context(), f.ion, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Hrană zilnică restaurată · 18.11.2026 · 5 rânduri · stoc Breeder 2: 734 kg" {
		t.Errorf("%s", res.Body)
	}
}

// T7 criterion 5, D40: deleting a delivery that a later day's consumption
// needed is saved with a STOCK_NEGATIVE warning; the Toast names the stock
// at the end of the delivery date. An undo of a delivery names kg and type.
func TestDeleteDelivery(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.svc.CreateDelivery(t.Context(), f.ion, f.delivery(nov17, f.breeder2, "500", nil))
	if err != nil {
		t.Fatal(err)
	}
	var created struct {
		Entry httpx.EntryRef `json:"entry"`
	}
	if err := json.Unmarshal(res.Body, &created); err != nil {
		t.Fatal(err)
	}
	e := created.Entry.ID
	// 510 kg of Breeder 2 on each of 17.11 and 18.11 (300 g × 1000 + 500 + 200 hens).
	for _, d := range []timex.Day{nov17, nov18} {
		fe := f.entry(dbq.EntryKindFeedDaily, f.farmA, d, f.ion, testdb.Now())
		for _, code := range []string{"H5.1", "H5.2", "H6.1"} {
			f.exec(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
				values ($1, $2, $3, $4, 'F', $5, 300)`, fe, f.farmA, f.sec[code], d, f.breeder2)
		}
	}
	res, err = f.del(f.ion, e, DeleteRequest{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	warning := httpx.Warning{Code: httpx.WarnStockNegative, Message: "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −20 kg. Verifică livrările."}
	if p := delDecode[CorrectionPreview](t, res); len(p.Effects) != 1 || p.Effects[0] != "Se șterge livrarea introdusă de Ion Rusu "+
		"la 12:00: 500 kg Breeder 2. Ștergerea rămâne în Istoric modificări." || !slices.Equal(p.Warnings, []httpx.Warning{warning}) {
		t.Errorf("preview %s", res.Body)
	}
	res, err = f.del(f.ion, e, DeleteRequest{})
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Livrare ștearsă · 500 kg Breeder 2 · stoc Hîrbovăț: 490 kg" || !slices.Equal(resp.Warnings, []httpx.Warning{warning}) {
		t.Errorf("%s", res.Body)
	}
	if a := f.latestAudit(e); *a.Summary != "Livrare · 500 kg Breeder 2" || !slices.Equal(a.RecordTypes, []string{"livrare"}) {
		t.Errorf("audit %+v", a)
	}
	res, err = f.svc.RestoreEntry(t.Context(), f.ion, e)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Livrare restaurată · 500 kg Breeder 2 · stoc Hîrbovăț: 990 kg" ||
		len(resp.Warnings) != 0 {
		t.Errorf("%s", res.Body)
	}

	// The undo of a delivery's create.
	res, err = f.svc.CreateDelivery(t.Context(), f.ion, f.delivery(nov18, f.cocosi, "40", nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(res.Body, &created); err != nil {
		t.Fatal(err)
	}
	res, err = f.svc.UndoCreate(t.Context(), f.ion, created.Entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Livrare anulată · 40 kg Cocoși" {
		t.Errorf("%s", res.Body)
	}
}

// The routes (mounted by T9 in Routes): DELETE with a JSON body and POST
// …/restore {}; a malformed id, an unknown member or a Motiv over 200
// characters is 400 before the service.
func TestDeleteRestoreRoutes(t *testing.T) {
	f := newFeedFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -2})
	mux := http.NewServeMux()
	h := NewHandler(f.svc)
	h.deleteRoutes(mux, httpx.NewChain())
	h.restoreRoutes(mux, httpx.NewChain())
	send := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), f.vet), method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}
	path := "/api/entries/" + e.String()
	for _, tc := range []struct {
		method, path, body, issue, code, message string
	}{
		{http.MethodDelete, "/api/entries/x", `{}`, "id", httpx.IssueBadType, messages.InvalidData()},
		{http.MethodDelete, path, `{"motiv": "x"}`, "motiv", httpx.IssueUnknownField, messages.InvalidData()},
		{http.MethodDelete, path, `{"reason": "` + strings.Repeat("ă", 212) + `"}`, "reason", httpx.IssueTooLong,
			"Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți salva."},
		{http.MethodPost, path + "/restore", `{"dryRun": true}`, "dryRun", httpx.IssueUnknownField, messages.InvalidData()},
	} {
		w := send(tc.method, tc.path, tc.body)
		b := envelopeOf(t, w)
		if w.Code != http.StatusBadRequest || len(b.Issues) != 1 || b.Issues[0].Path != tc.issue || b.Issues[0].Code != tc.code ||
			b.Message != tc.message {
			t.Errorf("%s %s %s: %d %+v", tc.method, tc.path, tc.body, w.Code, b)
		}
	}
	if w := send(http.MethodDelete, path, `{"reason": null, "sectionId": null, "dryRun": false}`); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "\"toast\":\"Mortalitate ștearsă · 18.11.2026 · 1 secțiune · în viață azi ♀ 1\u202f700\"") {
		t.Errorf("delete: %d %s", w.Code, w.Body)
	}
	if w := send(http.MethodPost, path+"/restore", `{}`); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"toast":"Mortalitate restaurată`) {
		t.Errorf("restore: %d %s", w.Code, w.Body)
	}
}

// COR-2: a delivery priced in Prețuri after its create can't be undone as
// a create (the price's audit rows don't name the entry, so the create is
// still its latest action): 409 STALE, the row kept, and the entry read
// offers no undo. With the price undone in Prețuri the undo works again.
func TestUndoCreateOfPricedDelivery(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.svc.CreateDelivery(t.Context(), f.ion, f.delivery(nov18, f.breeder2, "500", nil))
	if err != nil {
		t.Fatal(err)
	}
	e := delDecode[EntryWriteResponse](t, res).Entry.ID
	f.exec(`update feed_delivery set price_per_kg = 7.74, total_lei = 3870, price_input = 'per_kg', priced_by = $1, priced_at = $2
		where entry_id = $3`, f.owner.UserID, testdb.Now(), e)
	readUndo := func() bool {
		t.Helper()
		r, err := f.svc.ReadEntry(t.Context(), f.ion, e)
		if err != nil {
			t.Fatal(err)
		}
		return r.Actions.Undo
	}
	if readUndo() {
		t.Error("the read offers the undo of a priced delivery's create")
	}
	_, err = f.svc.UndoEntry(t.Context(), f.ion, e)
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeStale); body.Message != messages.Stale() {
		t.Errorf("message %q", body.Message)
	}
	if rows, deleted, _ := f.liveRows(e); rows != 1 || deleted {
		t.Errorf("after the refused undo: %d rows, deleted %v", rows, deleted)
	}

	f.exec(`update feed_delivery set price_per_kg = null, total_lei = null, price_input = null, priced_by = null, priced_at = null
		where entry_id = $1`, e)
	if !readUndo() {
		t.Error("unpriced again: no undo offered")
	}
	if _, err := f.svc.UndoEntry(t.Context(), f.ion, e); err != nil {
		t.Fatal(err)
	}
	if _, deleted, _ := f.liveRows(e); !deleted {
		t.Error("the undone create left the entry")
	}
}
