package entries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// dailyFormFixture is the feed form's farm (formFixture: H5 {H5.1, H5.2},
// H7 {H7.1} emptied 14.11, retired parentless H8 with ♀ 5, Telița) plus the
// Zilnic data of the form state, today 18.11.2026:
//
//	Guide G: SAPT 46 lay % ♀ 57,6.
//	Deaths (daily entries by the manager): 10.11 H5.2 ♀ 5 (outside the 7 days); 12.11 H5.2 ♀ 4;
//	  16.11 H5.2 ♀ cull 1 (boala); 17.11 H5.2 ♀ 3, H5.1 ♂ 1.
//	18.11 placement H5.2 ♂ 5 (an inflow: ♂ shown with 0 at the start).
//	18.11 Zilnic by the vet at 08:15: H5.1 ♀ mortality 2, ♂ cull 1 (boala), ♂ sexing error −1; a soft-deleted
//	  H5.2 ♀ mortality 7.
//	18.11 Zilnic by the manager at 09:10: H5 eggs Jos 3, Cat. 1 1, Cat. 2 2, Cat. 3 empty, Incubabile 40, 60,5 g.
//	18.11 a deleted Zilnic entry: H5.2 ♀ mortality 9.
//	18.11 values in the import's migration entry (created 01.11): H5.2 ♀ sexing error 0, H7.1 ♀ mortality 0.
//	Telița: 01.01.2025 T1 ♀ 10 placed, 31.12.2025 sold (cycle Seria 2025, closed).
type dailyFormFixture struct {
	*formFixture
	zVet, zMgr, zMigration uuid.UUID
}

func newDailyFormFixture(t *testing.T) *dailyFormFixture {
	t.Helper()
	f := &dailyFormFixture{formFixture: newFormFixture(t)}
	H := f.hirbovat
	f.exec(`update guide_week set lay_pct = 57.6 where week = 46 and sex = 'F'`)

	for _, d := range []struct {
		day       int
		code, sex string
		typ       string
		qty       int
	}{
		{10, "H5.2", "F", "mortality", -5},
		{12, "H5.2", "F", "mortality", -4},
		{16, "H5.2", "F", "cull", -1},
		{17, "H5.2", "F", "mortality", -3},
		{17, "H5.1", "M", "mortality", -1},
	} {
		e := f.entry("daily", H, fday(time.November, d.day), &f.mgr.UserID, testdb.Now())
		f.zrow(e, H, d.code, fday(time.November, d.day), d.sex, d.typ, d.qty)
	}
	e := f.entry("placement", H, testdb.Today, &f.mgr.UserID, testdb.Now())
	f.bird(e, H, "H5.2", testdb.Today, "M", "placement", 5)

	at := func(h, m int) time.Time { return time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau) }
	f.zVet = f.entry("daily", H, testdb.Today, &f.vet.UserID, at(8, 15))
	f.zrow(f.zVet, H, "H5.1", testdb.Today, "F", "mortality", -2)
	f.zrow(f.zVet, H, "H5.1", testdb.Today, "M", "cull", -1)
	f.zrow(f.zVet, H, "H5.1", testdb.Today, "M", "sexing_error", -1)
	gone := f.zrow(f.zVet, H, "H5.2", testdb.Today, "F", "mortality", -7)
	f.exec(`update bird_event set deleted_at = $1 where id = $2`, testdb.Now(), gone)

	f.zMgr = f.entry("daily", H, testdb.Today, &f.mgr.UserID, at(9, 10))
	f.exec(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g)
		values ($1, $2, $3, $4, 3, 1, 2, null, 40, 60.5)`, f.zMgr, H, f.sec["H5"], testdb.Today)

	deleted := f.entry("daily", H, testdb.Today, &f.mgr.UserID, at(7, 0))
	f.zrow(deleted, H, "H5.2", testdb.Today, "F", "mortality", -9)
	f.exec(`update entry set deleted_at = $1, deleted_by = $2 where id = $3`, testdb.Now(), f.mgr.UserID, deleted)

	f.zMigration = f.entry("migration", H, testdb.Today, nil, time.Date(2026, time.November, 1, 10, 0, 0, 0, timex.Chisinau))
	f.zrow(f.zMigration, H, "H5.2", testdb.Today, "F", "sexing_error", 0)
	f.zrow(f.zMigration, H, "H7.1", testdb.Today, "F", "mortality", 0)

	T := f.telita
	e = f.entry("placement", T, timex.MustDay(2025, time.January, 1), &f.mgr.UserID, testdb.Now())
	f.bird(e, T, "T1", timex.MustDay(2025, time.January, 1), "F", "placement", 10)
	e = f.entry("sale", T, timex.MustDay(2025, time.December, 31), &f.mgr.UserID, testdb.Now())
	f.bird(e, T, "T1", timex.MustDay(2025, time.December, 31), "F", "sale", -10)
	return f
}

// zrow inserts a live Zilnic bird_event row (a cull with the reason boala).
func (f *dailyFormFixture) zrow(entry, farm uuid.UUID, code string, d timex.Day, sex, typ string, qty int) uuid.UUID {
	f.t.Helper()
	var reason *string
	if typ == "cull" {
		r := "boala"
		reason = &r
	}
	return f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
		values ($1, $2, $3, $4, $5, $6, $7, $8)`, entry, farm, f.sec[code], d, sex, typ, qty, reason)
}

func (f *dailyFormFixture) buildDaily(p auth.Principal, farm uuid.UUID, d timex.Day) DailyFormResponse {
	f.t.Helper()
	res, err := BuildDailyForm(f.t.Context(), f.d.App, DailyFormInput{Principal: p, FarmID: farm, Date: d, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

// dailySexStr is a saved sex as "m c reason s" ("null" for nil fields), or
// "null".
func dailySexStr(v *DailySex) string {
	if v == nil {
		return "null"
	}
	return numStr(v.Mortality) + " " + numStr(v.Cull) + " " + strPtr(v.CullReason) + " " + numStr(v.SexingError)
}

func savedStr(s *httpx.SexPair[*DailySex]) string {
	if s == nil {
		return "null"
	}
	return dailySexStr(s.F) + " | " + dailySexStr(s.M)
}

func sexesStr(s []dbq.Sex) string {
	out := make([]string, 0, len(s))
	for _, x := range s {
		out = append(out, string(x))
	}
	return strings.Join(out, ",")
}

func dailyEntriesEqual(a, b []DailyFormEntry) bool {
	return slices.EqualFunc(a, b, func(x, y DailyFormEntry) bool {
		return x.EntryID == y.EntryID && strPtr(x.Author) == strPtr(y.Author) && x.At.Equal(y.At.Time) &&
			slices.Equal(x.Sections, y.Sections) && slices.Equal(x.Eggs, y.Eggs) &&
			x.Editable == y.Editable && x.Correctable == y.Correctable
	})
}

// T4 criterion 1: the Q2 shape on hand-built data, for the manager today.
func TestBuildDailyForm(t *testing.T) {
	f := newDailyFormFixture(t)
	res := f.buildDaily(f.mgr, f.hirbovat, testdb.Today)

	if res.Farm != (FormFarm{ID: f.hirbovat, Name: "Hîrbovăț", Type: dbq.FarmTypeLaying}) || res.Date != testdb.Today {
		t.Errorf("farm, date = %+v, %s", res.Farm, res.Date)
	}
	if res.SAPT == nil || *res.SAPT != 46 || res.AgeDays == nil || *res.AgeDays != 321 {
		t.Errorf("sapt, ageDays = %v, %v; want 46, 321", res.SAPT, res.AgeDays)
	}
	if res.Cycle == nil || *res.Cycle != (FormCycle{ID: f.cycle, Series: "Seria 2026", StartDate: fday(time.January, 1)}) {
		t.Errorf("cycle = %+v", res.Cycle)
	}
	if res.LockedReason != nil || res.Empty != nil {
		t.Errorf("lockedReason, empty = %v, %+v; want null, omitted", res.LockedReason, res.Empty)
	}

	// Halls: H5 with both compartments and its eggs; H7 only for the
	// migration's saved 0 on the emptied H7.1 (no hens: no eggs); the retired
	// H8 (♀ 5, no saved value) is not shown.
	type sec struct{ code, sexes, alive, yesterday, mean7, saved string }
	type hall struct {
		code     string
		sections []sec
		eggs     string
	}
	var got []hall
	for _, h := range res.Halls {
		g := hall{code: "null", eggs: "null"}
		if h.Hall != nil {
			g.code = h.Hall.Code
			if h.Hall.ID != f.sec[h.Hall.Code] {
				t.Errorf("hall %s id = %s", h.Hall.Code, h.Hall.ID)
			}
		}
		for _, s := range h.Sections {
			if s.SectionID != f.sec[s.Code] {
				t.Errorf("section %s id = %s", s.Code, s.SectionID)
			}
			g.sections = append(g.sections, sec{
				s.Code, sexesStr(s.Sexes),
				fmtPair(s.AliveStart.F, s.AliveStart.M), fmtPair(s.Yesterday.F, s.Yesterday.M),
				s.Mean7.F.String() + " " + s.Mean7.M.String(), savedStr(s.Saved),
			})
		}
		if e := h.Eggs; e != nil {
			g.eggs = e.Code + " " + strconv.Itoa(e.HensStart) + " " + numStr(e.GuideLayPct)
			if s := e.Saved; s != nil {
				g.eggs += " saved " + strings.Join([]string{numStr(s.Floor), numStr(s.Cat1), numStr(s.Cat2), numStr(s.Cat3), numStr(s.Hatching), numStr(s.AvgWeightG)}, " ")
			}
			if e.SectionID != f.sec[e.Code] {
				t.Errorf("eggs %s id = %s", e.Code, e.SectionID)
			}
		}
		got = append(got, g)
	}
	sevenths := func(n int64) string { return decimal.NewFromInt(n).Div(decimal.NewFromInt(7)).String() }
	want := []hall{
		{code: "H5", sections: []sec{
			// H5.1: ♀ 100, ♂ 10 − 1 (17.11); deaths 11–17.11 ♂ 1; the vet's values.
			{"H5.1", "F,M", "100 9", "0 1", "0 " + sevenths(1), "2 null null null | null 1 boala -1"},
			// H5.2: ♀ 50 − 5 − 4 − 1 − 3 = 37; deaths 11–17.11 4 + 1 + 3 = 8 (10.11 is outside);
			// ♂ shown for the inflow of the day; the migration's sexing error 0; the soft-deleted
			// row and the deleted entry's 9 are not saved values.
			{"H5.2", "F,M", "37 0", "3 0", sevenths(8) + " 0", "null null null 0 | null"},
		}, eggs: "H5 137 57.6 saved 3 1 2 null 40 60.5"},
		{code: "H7", sections: []sec{
			{"H7.1", "F", "0 0", "0 0", "0 0", "0 null null null | null"},
		}, eggs: "null"},
	}
	if !slices.EqualFunc(got, want, func(a, b hall) bool {
		return a.code == b.code && a.eggs == b.eggs && slices.Equal(a.sections, b.sections)
	}) {
		t.Errorf("halls\n got %+v\nwant %+v", got, want)
	}

	if want := []FormEmptyHall{{Hall: "H7", EmptySince: fday(time.November, 14), Label: "goală din 14.11.2026 · H7.1"}}; !slices.Equal(res.EmptyHalls, want) {
		t.Errorf("emptyHalls = %+v, want %+v", res.EmptyHalls, want)
	}

	// Existing, oldest first: the migration entry (no author, no action: its
	// kind isn't served, Q1), the vet's (correctable), the manager's own
	// (editable the same day). The deleted entry and the placement are not.
	ana, ion := "Ana", "Ion"
	at := func(h, m int) timex.LocalTime {
		return timex.At(time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau))
	}
	wantEntries := []DailyFormEntry{
		{EntryID: f.zMigration, At: timex.At(time.Date(2026, time.November, 1, 10, 0, 0, 0, timex.Chisinau)), Sections: []string{"H5.2", "H7.1"}, Eggs: []string{}},
		{EntryID: f.zVet, Author: &ana, At: at(8, 15), Sections: []string{"H5.1"}, Eggs: []string{}, Correctable: true},
		{EntryID: f.zMgr, Author: &ion, At: at(9, 10), Sections: []string{}, Eggs: []string{"H5"}, Editable: true, Correctable: true},
	}
	if !dailyEntriesEqual(res.Existing, wantEntries) {
		t.Errorf("existing = %+v,\nwant %+v", res.Existing, wantEntries)
	}
	if res.AsOf != timex.At(testdb.Now()) {
		t.Errorf("asOf = %v", res.AsOf)
	}

	// JSON: lists never null, empty omitted, null saved / eggs present.
	b, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"lockedReason":null`, `"eggs":null`, `"eggs":[]`, `"sections":[]`, `"M":null`, `"mean7":{"F":1.1429,"M":0}`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON has no %s: %s", key, b)
		}
	}
	if strings.Contains(string(b), `"empty"`) {
		t.Errorf("JSON has empty while a cycle holds the day: %s", b)
	}
}

func fmtPair(a, b int) string { return strconv.Itoa(a) + " " + strconv.Itoa(b) }

// T4 criterion 1: the entries' actions follow the caller (§3.2), and a vet
// on a past day gets the lockedReason; today has none.
func TestBuildDailyFormVet(t *testing.T) {
	f := newDailyFormFixture(t)
	res := f.buildDaily(f.vet, f.hirbovat, testdb.Today)
	if res.LockedReason != nil {
		t.Errorf("vet today: lockedReason = %q, want null", *res.LockedReason)
	}
	got := make([][2]bool, 0, len(res.Existing))
	for _, e := range res.Existing {
		got = append(got, [2]bool{e.Editable, e.Correctable})
	}
	// migration: nothing; own entry today: edit; the manager's: nothing (D71).
	if want := [][2]bool{{false, false}, {true, false}, {false, false}}; !slices.Equal(got, want) {
		t.Errorf("vet actions = %v, want %v", got, want)
	}

	past := f.buildDaily(f.vet, f.hirbovat, fday(time.November, 17))
	if past.LockedReason == nil || *past.LockedReason != "Pentru zilele trecute cere managerului." {
		t.Errorf("vet past day: lockedReason = %v", past.LockedReason)
	}
	// 17.11: H5.2 ♀ 3 and H5.1 ♂ 1 saved by the manager; the vet may not touch them.
	if len(past.Existing) != 2 || past.Existing[0].Editable || past.Existing[0].Correctable {
		t.Errorf("17.11 existing = %+v", past.Existing)
	}
	if mgr := f.buildDaily(f.mgr, f.hirbovat, fday(time.November, 17)); mgr.LockedReason != nil {
		t.Errorf("manager past day: lockedReason = %q, want null", *mgr.LockedReason)
	}
}

// T4 criteria 1, 3: a day inside a closed series' cycle is read-only for
// everyone (§3.1 sentence) and a growth farm has no eggs; a day no cycle
// holds has the empty-farm state.
func TestBuildDailyFormClosedSeriesAndEmptyFarm(t *testing.T) {
	f := newDailyFormFixture(t)
	closed := f.buildDaily(f.mgr, f.telita, timex.MustDay(2025, time.June, 1))
	if closed.LockedReason == nil || *closed.LockedReason != "Seria 2025 e închisă: nu mai primește înregistrări sau corecții." {
		t.Errorf("closed series: lockedReason = %v", closed.LockedReason)
	}
	if closed.Cycle == nil || closed.Cycle.Series != "Seria 2025" || closed.SAPT == nil || *closed.SAPT != 22 || closed.Empty != nil {
		t.Errorf("closed series: cycle %+v, sapt %v, empty %+v", closed.Cycle, closed.SAPT, closed.Empty)
	}
	// T1 has no parent: its own group (hall null); a growth farm: eggs null.
	if len(closed.Halls) != 1 || closed.Halls[0].Hall != nil || closed.Halls[0].Eggs != nil ||
		len(closed.Halls[0].Sections) != 1 || closed.Halls[0].Sections[0].Code != "T1" ||
		sexesStr(closed.Halls[0].Sections[0].Sexes) != "F" || closed.Halls[0].Sections[0].AliveStart != (httpx.SexPair[int]{F: 10}) {
		t.Errorf("closed series halls = %+v", closed.Halls)
	}

	empty := f.buildDaily(f.mgr, f.telita, testdb.Today)
	since := timex.MustDay(2025, time.December, 31)
	want := DailyFormEmpty{Since: &since, Label: "Ferma e goală din 31.12.2025. Ultimul ciclu: Seria 2025, SAPT 1–53.",
		Hint: "Ciclul pornește la prima populare sau transfer."}
	if empty.Empty == nil || empty.Empty.Since == nil || *empty.Empty.Since != since || empty.Empty.Label != want.Label || empty.Empty.Hint != want.Hint {
		t.Errorf("empty = %+v, want %+v", empty.Empty, want)
	}
	if empty.Cycle != nil || empty.SAPT != nil || empty.AgeDays != nil || empty.LockedReason != nil {
		t.Errorf("no cycle today: %+v", empty)
	}
	b, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"halls":[]`, `"emptyHalls":[]`, `"existing":[]`, `"cycle":null`, `"sapt":null`, `"since":"2025-12-31"`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON %s has no %s", b, key)
		}
	}

	never := f.buildDaily(f.mgr, f.telita, timex.MustDay(2024, time.June, 1))
	if never.Empty == nil || never.Empty.Since != nil || never.Empty.Label != "Ferma nu are încă niciun ciclu." {
		t.Errorf("before any cycle: empty = %+v", never.Empty)
	}

	if _, err := BuildDailyForm(t.Context(), f.d.App, DailyFormInput{FarmID: uuid.New(), Date: testdb.Today, Today: testdb.Today}); err != ErrFormFarmNotFound {
		t.Errorf("unknown farm: err = %v, want ErrFormFarmNotFound", err)
	}
}

// T4 criterion 1: the service's access and date checks.
func TestDailyFormService(t *testing.T) {
	f := newDailyFormFixture(t)
	svc := NewService(db.NewStore(f.d.App), f.d.Clock)
	ctx := t.Context()
	viewer := auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
	vetT := auth.Principal{UserID: uuid.New(), Role: auth.Vet, FarmIDs: []uuid.UUID{f.telita}}
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}

	_, err := svc.DailyForm(ctx, viewer, f.hirbovat, testdb.Today)
	if body := wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden); body.Message != "Nu ai acces la această pagină." {
		t.Errorf("viewer message = %q", body.Message)
	}
	_, err = svc.DailyForm(ctx, vetT, f.hirbovat, testdb.Today)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.DailyForm(ctx, vetT, uuid.Nil, testdb.Today)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.DailyForm(ctx, owner, uuid.New(), testdb.Today)
	wantCode(t, err, http.StatusNotFound, httpx.CodeNotFound)

	_, err = svc.DailyForm(ctx, f.mgr, f.hirbovat, fday(time.November, 19))
	body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation)
	want := httpx.Issue{Path: "date", Code: httpx.IssueOutOfRange, Message: "19.11.2026 e în viitor. Alege o dată până la 18.11.2026."}
	if len(body.Issues) != 1 || body.Issues[0] != want || body.Message != want.Message {
		t.Errorf("future date = %+v, want %+v", body, want)
	}

	for _, p := range []auth.Principal{f.vet, f.mgr, owner} {
		res, err := svc.DailyForm(ctx, p, f.hirbovat, testdb.Today)
		if err != nil || len(res.Halls) != 2 {
			t.Errorf("%s today: %d halls, %v", p.Role, len(res.Halls), err)
		}
	}
}

// T4 criterion 4: every read in one InReadTx, at most 6 statements; a farm
// with nothing to show reads no deaths, eggs or hens.
func TestDailyFormStatements(t *testing.T) {
	f := newDailyFormFixture(t)
	var st formStatements
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
	pc.ConnConfig.Tracer = &st
	pool, err := pgxpool.NewWithConfig(t.Context(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(t.Context()); err != nil {
		t.Fatal(err)
	}
	st.take()
	svc := NewService(db.NewStore(pool), f.d.Clock)

	run := func(farm uuid.UUID, d timex.Day, want ...string) {
		t.Helper()
		if _, err := svc.DailyForm(t.Context(), f.mgr, farm, d); err != nil {
			t.Fatal(err)
		}
		got := st.take()
		if len(got) < 2 || !strings.HasPrefix(got[0], "begin") || !strings.Contains(got[0], "repeatable read") ||
			!strings.Contains(got[0], "read only") || got[len(got)-1] != "commit" {
			t.Fatalf("statements %q: want one repeatable read, read only transaction", got)
		}
		reads := got[1 : len(got)-1]
		if !slices.Equal(reads, want) || len(reads) > 6 {
			t.Errorf("reads %q,\nwant %q (≤ 6)", reads, want)
		}
	}
	head := []string{"DailyFormHead", "DailyFormSections", "DailyFormEntries"}
	run(f.hirbovat, testdb.Today, slices.Concat(head, []string{"DailyFormEggs", "CalcDeathsWindow", "CalcHensStart"})...)
	run(f.telita, timex.MustDay(2025, time.June, 1), slices.Concat(head, []string{"CalcDeathsWindow"})...) // growth: no eggs
	run(f.telita, testdb.Today, head...)                                                                   // nothing to show
}

// The handler through dailyFormRoutes: query → service → JSON, errors
// through the envelope.
func TestDailyFormHandler(t *testing.T) {
	f := newDailyFormFixture(t)
	mux := http.NewServeMux()
	NewHandler(NewService(db.NewStore(f.d.App), f.d.Clock)).dailyFormRoutes(mux, httpx.NewChain())
	serve := func(p *auth.Principal, query string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/forms/daily?"+query, nil)
		if p != nil {
			r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	q := "farmId=" + f.hirbovat.String() + "&date=2026-11-18"
	w := serve(&f.mgr, q)
	var res DailyFormResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &res) != nil || len(res.Halls) != 2 || len(res.Existing) != 3 {
		t.Errorf("200: %d %s", w.Code, w.Body)
	}
	for _, tc := range []struct{ name, query, want string }{
		{"no date", "farmId=" + f.hirbovat.String(), `"path":"date","code":"required"`},
		{"repeated", q + "&date=2026-11-17", `"path":"date","code":"bad_type"`},
		{"unknown", q + "&tab=x", `"path":"tab","code":"unknown_field"`},
		{"future", "farmId=" + f.hirbovat.String() + "&date=2026-11-19", `"path":"date","code":"out_of_range"`},
	} {
		if w := serve(&f.mgr, tc.query); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), tc.want) {
			t.Errorf("%s: %d %s", tc.name, w.Code, w.Body)
		}
	}
	viewer := auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
	if w := serve(&viewer, q); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":"FORBIDDEN"`) {
		t.Errorf("viewer: %d %s", w.Code, w.Body)
	}
	other := auth.Principal{UserID: uuid.New(), Role: auth.Manager, FarmIDs: []uuid.UUID{f.telita}}
	if w := serve(&other, q); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":"FARM_FORBIDDEN"`) {
		t.Errorf("other farm: %d %s", w.Code, w.Body)
	}
	if w := serve(nil, q); w.Code != http.StatusUnauthorized {
		t.Errorf("no session: %d %s", w.Code, w.Body)
	}
}
