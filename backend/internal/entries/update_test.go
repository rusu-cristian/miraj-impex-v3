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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// at0815nov17 is 17.11.2026 08:15 in Chișinău: an entry of an earlier day.
var at0815nov17 = at0815.AddDate(0, 0, -1)

// storedValues is entry e's live rows as its create shape (what GET
// /api/entries/{id} returns as `values`) and its version.
func (f *dailyFixture) storedValues(e uuid.UUID) (DailyValues, string) {
	f.t.Helper()
	var day timex.Day
	var at time.Time
	if err := f.d.App.QueryRow(f.t.Context(), `select date, coalesce(updated_at, created_at) from entry where id = $1`, e).
		Scan(&day, &at); err != nil {
		f.t.Fatal(err)
	}
	s, err := LoadStored(f.t.Context(), f.d.App, e, dbq.EntryKindDaily, f.farmA, day)
	if err != nil {
		f.t.Fatal(err)
	}
	return s.Daily(), version(at)
}

// put sends a Zilnic PUT of e by p with v, the entry's current version and
// reason ("" = none).
func (f *dailyFixture) put(p auth.Principal, e uuid.UUID, v DailyValues, reason string, dryRun bool) (Result, error) {
	f.t.Helper()
	_, ver := f.storedValues(e)
	req := DailyUpdateRequest{IfVersion: ver, Values: v, DryRun: dryRun}
	if reason != "" {
		req.Reason = &reason
	}
	return f.svc.UpdateDaily(f.t.Context(), p, e, req)
}

// setMortality is v with section code's sex mortality set to n (nil clears
// the sex).
func (f *dailyFixture) setMortality(v DailyValues, code string, sex dbq.Sex, n *num.Num) DailyValues {
	out := v
	out.Sections = slices.Clone(v.Sections)
	for i, s := range out.Sections {
		if s.SectionID != f.sec[code] {
			continue
		}
		var x *DailySex
		if n != nil {
			x = &DailySex{Mortality: n}
			if cur := s.sex(sex); cur != nil {
				c := *cur
				c.Mortality = n
				x = &c
			}
		}
		out.Sections[i].setSex(sex, x)
		return out
	}
	out.Sections = append(out.Sections, DailySection{SectionID: f.sec[code]})
	out.Sections[len(out.Sections)-1].setSex(sex, &DailySex{Mortality: n})
	return out
}

// birdRow is one bird_event row of entry e (live or not) of a section, sex
// and type.
type birdRow struct {
	ID      uuid.UUID
	Qty     int
	Deleted bool
}

func (f *dailyFixture) birdRows(e uuid.UUID, code string, sex dbq.Sex) []birdRow {
	f.t.Helper()
	rows, err := f.d.App.Query(f.t.Context(), `select id, qty, deleted_at is not null from bird_event
		where entry_id = $1 and section_id = $2 and sex = $3 order by deleted_at nulls first, id`, e, f.sec[code], sex)
	if err != nil {
		f.t.Fatal(err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowToStructByPos[birdRow])
	if err != nil {
		f.t.Fatal(err)
	}
	return out
}

// entryState is an entry's corrected flag and updated_by.
func (f *dailyFixture) entryState(e uuid.UUID) (corrected bool, updatedBy *uuid.UUID) {
	f.t.Helper()
	if err := f.d.App.QueryRow(f.t.Context(), `select corrected, updated_by from entry where id = $1`, e).
		Scan(&corrected, &updatedBy); err != nil {
		f.t.Fatal(err)
	}
	return corrected, updatedBy
}

// auditChanges is the changes of an entry's latest audit row.
func (f *dailyFixture) auditChanges(e uuid.UUID) []audit.Change {
	f.t.Helper()
	var raw []byte
	if err := f.d.App.QueryRow(f.t.Context(), `select changes from audit_log where entry_id = $1 order by at desc, id desc limit 1`, e).
		Scan(&raw); err != nil {
		f.t.Fatal(err)
	}
	var out []audit.Change
	if err := json.Unmarshal(raw, &out); err != nil {
		f.t.Fatal(err)
	}
	return out
}

func changeJSON(t *testing.T, c []audit.Change) string {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// T6 criterion 1, plan M5-D5: the author edits on the creation day without
// a Motiv (a sent one is ignored): audit `update` same_day_author, no
// reason, `corrected` stays false; the row changes in place; the Toast is the
// form's create Toast; entry.updated + counts.changed.
func TestUpdateDailyAuthorEdit(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1})
	before := f.birdRows(e, "H5.1", dbq.SexF)
	v, ver := f.storedValues(e)
	res, err := f.put(f.vet, e, f.setMortality(v, "H5.1", dbq.SexF, dnum(2)), "ignorat", false)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	// The farm ends 18.11 with ♀ 1 800 − 2; 2 deaths over a 0,0 mean warn (Z6, D53).
	if res.Status != http.StatusOK || resp.Toast != "Zilnic salvat · 1 secțiune · ♀ −2 · ♂ 0" || resp.Alive == nil ||
		*resp.Alive != (httpx.SexPair[int]{F: 1798, M: 433}) || resp.Entry.Version == ver || len(resp.Warnings) != 1 {
		t.Fatalf("%d %s", res.Status, res.Body)
	}
	if after := f.birdRows(e, "H5.1", dbq.SexF); len(after) != 1 || after[0].ID != before[0].ID || after[0].Qty != -2 {
		t.Errorf("rows %+v, want the same row with −2", after)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionUpdate || a.Reason != nil || !a.SameDayAuthor || a.Summary == nil ||
		*a.Summary != "Mortalitate · 1 secțiune" || !slices.Equal(a.RecordTypes, []string{"mortalitate"}) {
		t.Errorf("audit %+v", a)
	}
	if got := changeJSON(t, f.auditChanges(e)); got != `[{"key":"H5.1 · găini moarte","section":"H5.1","sex":"F","field":"mortality","old":1,"new":2}]` {
		t.Errorf("changes %s", got)
	}
	var old updateSnapshot
	if err := json.Unmarshal(a.OldValues, &old); err != nil || len(old.BirdEvents) != 1 || old.BirdEvents[0].Qty != -1 || old.Corrected {
		t.Errorf("old_values %s (%v)", a.OldValues, err)
	}
	if corrected, by := f.entryState(e); corrected || by == nil || *by != f.vet.UserID {
		t.Errorf("corrected %v, updated_by %v", corrected, by)
	}
	if got := f.liveTypes(); !slices.Equal(got, []string{"entry.updated", "counts.changed"}) {
		t.Errorf("live = %v", got)
	}
}

// T6 criterion 1, §6.15: a manager corrects someone else's entry: a Motiv
// is required (400 on `reason` naming the change; not asked of a dry run),
// then the correction is saved with it, `corrected` set and its Toast names
// the section's birds.
func TestUpdateDailyCorrection(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	v, _ := f.storedValues(e)
	v3 := f.setMortality(v, "H5.2", dbq.SexF, dnum(3))

	_, err := f.put(f.ion, e, v3, "", false)
	body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation)
	if body.Message != "Motiv lipsă. Scrie de ce se schimbă H5.2 ♀ din 1 în 3; fără motiv corecția nu se salvează." ||
		len(body.Issues) != 1 || body.Issues[0].Path != "reason" {
		t.Errorf("envelope %+v", body)
	}
	if res, err := f.put(f.ion, e, v3, "", true); err != nil || res.Status != http.StatusOK {
		t.Fatalf("dry run without Motiv: %v %s", err, res.Body)
	}

	res, err := f.put(f.ion, e, v3, "  Numărătoare refăcută.  ", false)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 497" || resp.Alive == nil || resp.Alive.F != 1797 {
		t.Errorf("%s", res.Body)
	}
	a := f.latestAudit(e)
	if a.Action != dbq.AuditActionUpdate || a.Reason == nil || *a.Reason != "Numărătoare refăcută." || a.SameDayAuthor {
		t.Errorf("audit %+v", a)
	}
	if corrected, by := f.entryState(e); !corrected || by == nil || *by != f.ion.UserID {
		t.Errorf("corrected %v, updated_by %v", corrected, by)
	}
	// A second change reads the moved version; the old one is 409 STALE.
	_, err = f.svc.UpdateDaily(t.Context(), f.ion, e, DailyUpdateRequest{IfVersion: "2026-11-18T06:15:00.000000Z", Values: v, Reason: ptr("x")})
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeStale); body.Message != messages.Stale() {
		t.Errorf("message %q", body.Message)
	}
}

// T6 criterion 1, §6.15: a PUT that changes nothing is 422 "Nu ai schimbat
// nicio cifră." (a dry run carries it in refusals); a value in another
// notation (3.0) is no change.
func TestUpdateDailyNothingChanged(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -3})
	v, _ := f.storedValues(e)
	same := f.setMortality(v, "H5.2", dbq.SexF, ptr(num.MustParse("3.0")))
	_, err := f.put(f.ion, e, same, "motiv", false)
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); body.Message != "Nu ai schimbat nicio cifră." {
		t.Errorf("message %q", body.Message)
	}
	res, err := f.put(f.ion, e, same, "", true)
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	if len(p.Refusals) != 1 || p.Refusals[0].Message != "Nu ai schimbat nicio cifră." || len(p.Changes) != 0 {
		t.Errorf("%s", res.Body)
	}
}

// T6 criterion 1, Q10: a vet on own entry of an earlier day gets the S16
// sentence; a vet on someone else's entry and a viewer "Nu ai acces …";
// another farm FARM_FORBIDDEN; a deleted or unknown entry 404.
func TestUpdateDailyAccess(t *testing.T) {
	f := newDailyFixture(t)
	past := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H5.1", "F", "mortality", -1})
	ion := f.delDaily(f.ion, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	gone := f.delDaily(f.ion, nov18, at0815, 0, delRow{"H6.1", "F", "mortality", -1})
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, gone, testdb.Now(), f.ion.UserID)
	pastV, _ := f.storedValues(past)
	ionV, _ := f.storedValues(ion)

	for _, tc := range []struct {
		name    string
		p       auth.Principal
		e       uuid.UUID
		v       DailyValues
		status  int
		code    string
		message string
	}{
		{"vet, own of an earlier day", f.vet, past, f.setMortality(pastV, "H5.1", dbq.SexF, dnum(2)), 403, httpx.CodeForbidden, messages.OwnEntryPastDay()},
		{"vet, someone else's", f.vet, ion, f.setMortality(ionV, "H5.2", dbq.SexF, dnum(2)), 403, httpx.CodeForbidden, messages.PageForbidden()},
		{"viewer", f.viewer, ion, f.setMortality(ionV, "H5.2", dbq.SexF, dnum(2)), 403, httpx.CodeForbidden, messages.PageForbidden()},
		{"another farm", f.outsider, ion, f.setMortality(ionV, "H5.2", dbq.SexF, dnum(2)), 403, httpx.CodeFarmForbidden, messages.FarmForbidden()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.put(tc.p, tc.e, tc.v, "motiv", false)
			if body := wantCode(t, err, tc.status, tc.code); body.Message != tc.message {
				t.Errorf("message %q", body.Message)
			}
		})
	}
	_, err := f.svc.UpdateDaily(t.Context(), f.ion, gone, DailyUpdateRequest{IfVersion: "2026-11-18T10:00:00Z", Values: ionV})
	wantCode(t, err, http.StatusNotFound, httpx.CodeNotFound)
	// The owner corrects on every farm.
	if _, err := f.put(f.owner, past, f.setMortality(pastV, "H5.1", dbq.SexF, dnum(2)), "motiv", false); err != nil {
		t.Errorf("owner: %v", err)
	}
}

// T6 criterion 1, plan M5-D6: a cleared value's row is soft-deleted, a
// changed one keeps its row, an added one is a new row; an added value on a
// key another entry holds is 409 ALREADY_ENTERED naming that entry.
func TestUpdateDailyRows(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1}, delRow{"H5.2", "F", "mortality", -2})
	holder := f.delDaily(f.ion, nov18, at0815.Add(time.Hour), 0, delRow{"H6.2", "F", "mortality", -1})
	v, _ := f.storedValues(e)
	h51 := f.birdRows(e, "H5.1", dbq.SexF)

	taken := f.setMortality(v, "H6.2", dbq.SexF, dnum(4))
	_, err := f.put(f.ion, e, taken, "motiv", false)
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.EntryID == nil || *body.EntryID != holder || body.Message != messages.DailyAlreadyEntered("H6.2", dbq.SexF, nov18, at0815.Add(time.Hour)) {
		t.Errorf("envelope %+v", body)
	}

	next := f.setMortality(f.setMortality(f.setMortality(v, "H5.2", dbq.SexF, nil), "H5.1", dbq.SexF, dnum(3)), "H6.1", dbq.SexF, dnum(1))
	res, err := f.put(f.ion, e, next, "motiv", false)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.birdRows(e, "H5.1", dbq.SexF); len(got) != 1 || got[0].ID != h51[0].ID || got[0].Qty != -3 {
		t.Errorf("H5.1 %+v", got)
	}
	if got := f.birdRows(e, "H5.2", dbq.SexF); len(got) != 1 || !got[0].Deleted {
		t.Errorf("H5.2 %+v, want its row soft-deleted", got)
	}
	if got := f.birdRows(e, "H6.1", dbq.SexF); len(got) != 1 || got[0].Deleted || got[0].Qty != -1 {
		t.Errorf("H6.1 %+v", got)
	}
	// ♀ moves −2 + 2 − 1 = −1: 1 800 − 3 (this save) − 1 (the holder's) − 1.
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție salvată · 3 valori · în viață azi ♀ 1\u202f795 · ♂ 433" {
		t.Errorf("toast %q", resp.Toast)
	}
	want := `[{"key":"H5.1 · găini moarte","section":"H5.1","sex":"F","field":"mortality","old":1,"new":3},` +
		`{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":2,"new":null},` +
		`{"key":"H6.1 · găini moarte","section":"H6.1","sex":"F","field":"mortality","old":null,"new":1}]`
	if got := changeJSON(t, f.auditChanges(e)); got != want {
		t.Errorf("changes\n%s\nwant\n%s", got, want)
	}
}

// T6 criterion 1, plan M5-D6: Z3, Z4, Z5 and Z7 judge only the changed and
// added values: a stored value that the rules would refuse today never
// blocks an unrelated correction, but changing it does. A Z5 bound that §6.1
// also refuses answers §6.1's stricter field sentence (the Corectează form).
func TestUpdateDailyRulesOnChangedValues(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.ion, nov18, at0815, 0, delRow{"H5.1", "F", "mortality", -1})
	// Eggs on H7, whose hens were all sold on 14.11 (stored before that rule).
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, cat1) values ($1, $2, $3, $4, 5)`, e, f.farmA, f.sec["H7"], nov18)
	v, _ := f.storedValues(e)
	if len(v.Eggs) != 1 {
		t.Fatalf("eggs %+v", v.Eggs)
	}
	if _, err := f.put(f.maria, e, f.setMortality(v, "H5.1", dbq.SexF, dnum(2)), "motiv", false); err != nil {
		t.Fatalf("unrelated correction: %v", err)
	}
	v, _ = f.storedValues(e)

	eggs := v
	eggs.Eggs = []DailyEggs{{SectionID: f.sec["H7"], Cat1: dnum(6)}}
	cull := v
	cull.Sections = append(slices.Clone(v.Sections), DailySection{SectionID: f.sec["H6.2"], F: &DailySex{Cull: dnum(2)}})
	// 11 dead of 10 roosters, 5 found by sexing: Z5 refuses, the day ends with 4.
	z5 := v
	z5.Sections = append(slices.Clone(v.Sections), DailySection{SectionID: f.sec["H6.2"], M: &DailySex{Mortality: dnum(11), SexingError: dnum(5)}})
	for _, tc := range []struct {
		name    string
		v       DailyValues
		field   string
		message string
	}{
		{"Z7 eggs without hens", eggs, "values.eggs[0].sectionId", messages.NoBirdsOnDay("H7", dbq.SexF, nov18)},
		{"Z3 no roosters", f.setMortality(v, "H6.1", dbq.SexM, dnum(1)), "values.sections[1].M", messages.NoBirdsOnDay("H6.1", dbq.SexM, nov18)},
		{"Z4 cull without reason", cull, "values.sections[1].F.cullReason", messages.CullReasonMissing("H6.2", dbq.SexF, 2, nov18)},
		{"Z5 over the day's birds, §6.1 lets it through", z5, "values.sections[1].M.mortality",
			messages.DayBoundExceeded("H6.2", dbq.SexM, 11, 10, nov18, messages.FieldMortality, 10)},
		{"Z5 and §6.1: the §6.1 bound", f.setMortality(v, "H6.2", dbq.SexM, dnum(11)), "values.sections[1].M.mortality",
			"H6.2 ♂ 11: ar lăsa −1 cocoș azi, 18.11.2026. Scrie cel mult 10."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.put(f.maria, e, tc.v, "motiv", false)
			body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
			if len(body.Refusals) != 1 || body.Refusals[0].Field != tc.field || body.Refusals[0].Message != tc.message {
				t.Errorf("refusals %+v", body.Refusals)
			}
		})
	}
}

// T6 criterion 1, §6.1, Q4: a correction that raises a past day's removal
// until a later day goes below 0 is 422 with the field bound and the
// "Corecția ar lăsa …" banner; the author's own edit names "Salvarea"; a dry
// run answers the same refusals in its preview and writes nothing.
func TestUpdateDailyCounts(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H6.2", "M", "mortality", -2})
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	f.bird(sale, "H6.2", nov18, "M", "sale", -8) // H6.2 ♂ ends 18.11 with 10 − 2 − 8 = 0
	v, ver := f.storedValues(e)
	raised := f.setMortality(v, "H6.2", dbq.SexM, dnum(5))

	field := "H6.2 ♂ 5: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel mult 2."
	_, err := f.put(f.ion, e, raised, "motiv", false)
	body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused)
	if body.Message != "Corecția ar lăsa H6.2 cu −3 cocoși azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0." ||
		len(body.Refusals) != 1 || body.Refusals[0].Field != "values.sections[0].M.mortality" || body.Refusals[0].Message != field ||
		body.Refusals[0].Max == nil || *body.Refusals[0].Max != 2 || *body.Refusals[0].MinAliveEnd != -3 ||
		*body.Refusals[0].FirstNegativeDate != nov18 {
		t.Errorf("envelope %+v", body)
	}

	res, err := f.put(f.ion, e, raised, "", true)
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	if len(p.Refusals) != 1 || p.Refusals[0].Message != field || p.CheckedDays != 2 {
		t.Errorf("preview %s", res.Body)
	}
	if _, now := f.storedValues(e); now != ver || f.birdRows(e, "H6.2", dbq.SexM)[0].Qty != -2 {
		t.Errorf("the dry run wrote")
	}

	own := f.delDaily(f.ion, nov17, at0815, 0, delRow{"H6.2", "F", "mortality", -1}) // created today: Ion's edit
	f.bird(sale, "H6.2", nov18, "F", "sale", -99)                                    // H6.2 ♀ ends 18.11 with 0
	ownV, _ := f.storedValues(own)
	_, err = f.put(f.ion, own, f.setMortality(ownV, "H6.2", dbq.SexF, dnum(2)), "", false)
	if body := wantCode(t, err, http.StatusUnprocessableEntity, httpx.CodeRefused); !strings.HasPrefix(body.Message, "Salvarea ar lăsa H6.2 cu −1 găină azi") {
		t.Errorf("message %q", body.Message)
	}
}

// T6 criterion 3, Q7: the dry run of a correction on an earlier day answers
// the changes, the effect lines, per changed section × sex the birds at the
// end of the entry date and of today before and after, the days checked and
// the Z6 warnings of the whole save; nothing is written.
func TestUpdateDailyPreview(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H5.2", "F", "mortality", -1}, delRow{"H6.1", "F", "mortality", -4})
	v, ver := f.storedValues(e)
	before := f.rows()
	res, err := f.put(f.ion, e, f.setMortality(v, "H5.2", dbq.SexF, dnum(3)), "", true)
	if err != nil {
		t.Fatal(err)
	}
	p := delDecode[CorrectionPreview](t, res)
	if got := changeJSON(t, p.Changes); got != `[{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":1,"new":3}]` {
		t.Errorf("changes %s", got)
	}
	// ♀ at the end of 18.11: 1 800 − 5.
	wantEffects := []string{
		"H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.", "Mortalitate · morți ♀ −5 → −7",
		"Hîrbovăț · ♀ azi 1\u202f795 → 1\u202f793",
	}
	wantAlive := []PreviewAlive{
		{Section: "H5.2", Sex: dbq.SexF, Date: nov17, Before: 499, After: 497},
		{Section: "H5.2", Sex: dbq.SexF, Date: nov18, Before: 499, After: 497},
	}
	if !slices.Equal(p.Effects, wantEffects) || !slices.Equal(p.Alive, wantAlive) || p.CheckedDays != 2 ||
		len(p.Refusals) != 0 || len(p.Warnings) != 2 || p.Warnings[1].Code != httpx.WarnMortalitySpike || p.Warnings[1].SectionCode != "H6.1" {
		t.Errorf("preview %s", res.Body)
	}
	if after := f.rows(); after != before {
		t.Errorf("rows %v → %v", before, after)
	}
	if _, now := f.storedValues(e); now != ver {
		t.Errorf("version moved")
	}
}

// T6 criterion 5: an egg-only correction moves no bird: its Toast names the
// value, no `alive`, no counts.changed, the chip oua.
func TestUpdateDailyEggs(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 13, delRow{"H5.1", "F", "mortality", -1})
	v, _ := f.storedValues(e)
	v.Eggs = slices.Clone(v.Eggs)
	v.Eggs[0].Cat1 = dnum(15)
	res, err := f.put(f.ion, e, v, "motiv", false)
	if err != nil {
		t.Fatal(err)
	}
	resp := delDecode[EntryWriteResponse](t, res)
	if resp.Toast != "Corecție salvată · H5 Cat. 1 13 → 15" || resp.Alive != nil {
		t.Errorf("%s", res.Body)
	}
	a := f.latestAudit(e)
	if !slices.Equal(a.RecordTypes, []string{"oua"}) || *a.Summary != "Mortalitate · 1 secțiune · Ouă · 1 hală" {
		t.Errorf("audit %+v", a)
	}
	if got := f.liveTypes(); !slices.Equal(got, []string{"entry.updated"}) {
		t.Errorf("live = %v", got)
	}
	var total *int
	if err := f.d.App.QueryRow(t.Context(), `select total_eggs from egg_production where entry_id = $1 and deleted_at is null`, e).Scan(&total); err != nil ||
		total == nil || *total != 15 {
		t.Errorf("total_eggs %v (%v)", total, err)
	}
}

// Q12: a cull reason changes in place (the enum through jsonb_to_recordset);
// the change carries the shown words, the Toast names the value without
// birds, the live entry.updated carries the corrected flag.
func TestUpdateDailyCullReason(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.1", "F", "cull", -2})
	v, _ := f.storedValues(e)
	v.Sections = slices.Clone(v.Sections)
	x := *v.Sections[0].F
	x.CullReason = ptr("altul")
	v.Sections[0].F = &x
	res, err := f.put(f.ion, e, v, "motiv", false)
	if err != nil {
		t.Fatal(err)
	}
	if resp := delDecode[EntryWriteResponse](t, res); resp.Toast != "Corecție salvată · H5.1 ♀ motiv sacrificare boală → altul" || resp.Alive != nil {
		t.Errorf("%s", res.Body)
	}
	var reason string
	if err := f.d.App.QueryRow(t.Context(), `select cull_reason from bird_event where entry_id = $1 and deleted_at is null`, e).Scan(&reason); err != nil || reason != "altul" {
		t.Errorf("cull_reason %q (%v)", reason, err)
	}
	want := `[{"key":"H5.1 · motiv sacrificare găini","section":"H5.1","sex":"F","field":"cullReason","old":"boală","new":"altul"}]`
	if got := changeJSON(t, f.auditChanges(e)); got != want {
		t.Errorf("changes %s", got)
	}
	if a := f.latestAudit(e); !slices.Equal(a.RecordTypes, []string{"sacrificare"}) {
		t.Errorf("chips %v", a.RecordTypes)
	}
	var corrected bool
	if err := f.d.App.QueryRow(t.Context(), `select (payload->>'corrected')::boolean from live_event where type = 'entry.updated'`).Scan(&corrected); err != nil || !corrected {
		t.Errorf("entry.updated corrected = %v (%v)", corrected, err)
	}
}

// T6 criterion 2: DiffDaily keys and orders the changes (Q12): sections in
// the new save's order then those only in the old one, ♀ before ♂, the
// fields in form order, a cull reason as its shown word; then the egg units.
func TestDiffDaily(t *testing.T) {
	a, b, h := uuid.New(), uuid.New(), uuid.New()
	codes := map[uuid.UUID]string{a: "H5.1", b: "H5.2", h: "H5"}
	boala, altul := "boala", "altul"
	old := DailyValues{
		Sections: []DailySection{
			{SectionID: a, F: &DailySex{Mortality: dnum(1), Cull: dnum(2), CullReason: &boala}},
			{SectionID: b, M: &DailySex{SexingError: dnum(-1)}},
		},
		Eggs: []DailyEggs{{SectionID: h, Cat1: dnum(13), AvgWeightG: ptr(num.MustParse("67.04"))}},
	}
	new := DailyValues{
		Sections: []DailySection{
			{SectionID: b, F: &DailySex{Mortality: dnum(0)}},
			{SectionID: a, F: &DailySex{Mortality: dnum(1), Cull: dnum(2), CullReason: &altul}},
		},
		Eggs: []DailyEggs{{SectionID: h, Cat1: dnum(15), AvgWeightG: ptr(num.MustParse("67.040"))}},
	}
	want := `[{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":null,"new":0},` +
		`{"key":"H5.2 · eroare sexare cocoși","section":"H5.2","sex":"M","field":"sexingError","old":-1,"new":null},` +
		`{"key":"H5.1 · motiv sacrificare găini","section":"H5.1","sex":"F","field":"cullReason","old":"boală","new":"altul"},` +
		`{"key":"H5 · Cat. 1","section":"H5","field":"cat1","old":13,"new":15}]`
	if got := changeJSON(t, DiffDaily(old, new, codes)); got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
	if got := DiffDaily(old, old, codes); len(got) != 0 {
		t.Errorf("same save: %v", got)
	}
}

// putHTTP sends PUT /api/entries/{id} as p through updateRoutes.
func (f *dailyFixture) putHTTP(p auth.Principal, path, body string) *httptest.ResponseRecorder {
	f.t.Helper()
	mux := http.NewServeMux()
	NewHandler(f.svc).updateRoutes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPut, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

// The handler (§7.8, plan M5-D5, D6): role before any read, the stored entry
// (404, FARM_FORBIDDEN, a kind not served 404), the kind's decode with
// unknown members named under `values` (a feed entry decodes its own
// shape), level 1 (ifVersion,
// values.farmId / values.date equal to the stored ones, the Motiv's length,
// the Zilnic checks under "values."), then the service.
func TestUpdateHandler(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	feed := f.entry(dbq.EntryKindFeedDaily, f.farmA, nov18, f.ion, at0815)
	_, ver := f.storedValues(e)
	path := "/api/entries/" + e.String()
	body := func(over map[string]any, values map[string]any) string {
		vals := map[string]any{"farmId": f.farmA, "date": "2026-11-18",
			"sections": []any{map[string]any{"sectionId": f.sec["H5.2"], "F": map[string]any{"mortality": 3}}}}
		for k, x := range values {
			vals[k] = x
		}
		m := map[string]any{"ifVersion": ver, "values": vals, "reason": "Numărătoare refăcută."}
		for k, x := range over {
			m[k] = x
		}
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	envelope := func(w *httptest.ResponseRecorder) httpx.APIError {
		t.Helper()
		var b httpx.ErrorBody
		if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
			t.Fatalf("%d %s: %v", w.Code, w.Body, err)
		}
		return b.Error
	}
	other := uuid.New()
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		path   string
		body   string
		status int
		code   string
		issue  string // path of the first issue
	}{
		{"viewer", f.viewer, path, body(nil, nil), 403, httpx.CodeForbidden, ""},
		{"unknown entry", f.ion, "/api/entries/" + other.String(), body(nil, nil), 404, httpx.CodeNotFound, ""},
		{"kind not served", f.ion, "/api/entries/" + sale.String(), body(nil, nil), 404, httpx.CodeNotFound, ""},
		{"the kind's own values", f.ion, "/api/entries/" + feed.String(), body(nil, nil), 400, httpx.CodeValidation, "values.sections"},
		{"another farm", f.outsider, path, body(nil, nil), 403, httpx.CodeFarmForbidden, ""},
		{"malformed id", f.ion, "/api/entries/x", body(nil, nil), 400, httpx.CodeValidation, "id"},
		{"unknown member in values", f.ion, path, body(nil, map[string]any{"sections": []any{map[string]any{"sectionId": f.sec["H5.2"], "F": map[string]any{"foo": 1}}}}), 400, httpx.CodeValidation, "values.sections[0].F.foo"},
		{"clientRequestId", f.ion, path, body(map[string]any{"clientRequestId": uuid.New()}, nil), 400, httpx.CodeValidation, "clientRequestId"},
		{"ifVersion missing", f.ion, path, body(map[string]any{"ifVersion": ""}, nil), 400, httpx.CodeValidation, "ifVersion"},
		{"ifVersion malformed", f.ion, path, body(map[string]any{"ifVersion": "ieri"}, nil), 400, httpx.CodeValidation, "ifVersion"},
		{"date changed", f.ion, path, body(nil, map[string]any{"date": "2026-11-17"}), 400, httpx.CodeValidation, "values.date"},
		{"farm changed", f.ion, path, body(nil, map[string]any{"farmId": f.farmB}), 400, httpx.CodeValidation, "values.farmId"},
		{"Z2 under values", f.ion, path, body(nil, map[string]any{"sections": []any{map[string]any{"sectionId": f.sec["H5.2"], "F": map[string]any{"mortality": 1.5}}}}), 400, httpx.CodeValidation, "values.sections[0].F.mortality"},
		{"Motiv too long", f.ion, path, body(map[string]any{"reason": strings.Repeat("a", 212)}, nil), 400, httpx.CodeValidation, "reason"},
		{"stale", f.ion, path, body(map[string]any{"ifVersion": "2026-11-18T06:15:00.000001Z"}, nil), 409, httpx.CodeStale, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := f.putHTTP(tc.p, tc.path, tc.body)
			env := envelope(w)
			if w.Code != tc.status || env.Code != tc.code {
				t.Fatalf("%d %s", w.Code, w.Body)
			}
			if tc.issue != "" && (len(env.Issues) == 0 || env.Issues[0].Path != tc.issue) {
				t.Errorf("issues %+v", env.Issues)
			}
		})
	}
	if env := envelope(f.putHTTP(f.ion, path, body(map[string]any{"reason": strings.Repeat("a", 212)}, nil))); env.Message != messages.ReasonTooLong(212, 200) {
		t.Errorf("message %q", env.Message)
	}

	w := f.putHTTP(f.ion, path, body(nil, nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"toast":"Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 497"`) {
		t.Errorf("%d %s", w.Code, w.Body)
	}
}

// T6 criterion 5: a correction is at most 16 statements besides the per-row
// writes, begin and commit.
func TestUpdateDailyStatements(t *testing.T) {
	f := newDailyFixture(t)
	e := f.delDaily(f.vet, nov17, at0815nov17, 0, delRow{"H5.2", "F", "mortality", -1}, delRow{"H6.1", "F", "mortality", -4})
	v, _ := f.storedValues(e)
	var log statementLog
	pc, err := db.PoolConfig(testdb.DSN(f.d.Name), 1, &testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	prepare := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "set role "+testdb.RoleApp); err != nil {
			return err
		}
		return prepare(ctx, conn)
	}
	pc.ConnConfig.Tracer = &log
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	_, ver := f.storedValues(e)
	log.take()

	svc := NewService(db.NewStore(pool), f.d.Clock)
	reason := "motiv"
	if _, err := svc.UpdateDaily(context.Background(), f.ion, e, DailyUpdateRequest{
		IfVersion: ver, Values: f.setMortality(v, "H5.2", dbq.SexF, dnum(3)), Reason: &reason,
	}); err != nil {
		t.Fatal(err)
	}
	got := slices.DeleteFunc(log.take(), func(n string) bool { return n == "?" })
	want := []string{
		"EntryChangeHeader", "LockFarm", "EntryChangeLock", "EntryBirdRows", "EntryEggRows", "DailySaveSlots",
		"CalcDeathsWindow", "CalcFarmSexDay", "CalcSectionAliveEnd", "EntryUpdateBirdRows", "UpdateEntryHeader",
		"NegativeCounts", "InsertAuditLog", "EmitLiveEvent", "EmitLiveEvent",
	}
	perRow := 1 // EntryUpdateBirdRows
	if !slices.Equal(got, want) || len(got)-perRow > 16 {
		t.Errorf("statements %q,\nwant %q", got, want)
	}
}
