package entries

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// formFixture is a hand-built farm for the Hrană zilnică form state, today
// 18.11.2026:
//
//	Seria 2026 placed 01.01.2026 at age 0 (active; SAPT 46 on 17–18.11); Seria 2025 placed 01.01.2025 (closed).
//	Guide "G ouat": SAPT 46 feed_g ♀ 157, ♂ row with a null feed_g.
//	Feed types: Breeder 2 (♀), Cocoși (♂), Universal (ambele), Vechi (♀, retired).
//	Hîrbovăț (laying, guide G): hall H5 (H5.1, H5.2), hall H7 (H7.1), parentless H8 (retired, still ♀ 5);
//	  cycle Seria 2026 from 01.01.2026, open.
//	  01.01.2026 placement H5.1 ♀ 100 · ♂ 10, H5.2 ♀ 50, H7.1 ♀ 30, H8 ♀ 5; 14.11.2026 sale H7.1 ♀ 30 (H7 empty).
//	  01.11.2026 delivery Breeder 2 1 000 kg, Cocoși 100 kg.
//	  16.11 feed_daily: H5.1 ♀ 150 g Breeder 2, H5.1 ♂ 140 g Cocoși, H5.2 ♀ 150 g Universal.
//	  17.11 feed_daily: H5.1 ♀ 155 g Breeder 2; migration: H5.1 ♂ kg-only 1,5 kg Cocoși.
//	  18.11 feed_daily by the vet at 08:44: H5.1 ♀ 160 g Breeder 2; by the manager at 09:00: H5.1 ♂ 145 g Cocoși.
//	Telița (growth): T1; cycle Seria 2025 01.01.2025–31.12.2025 (emptied); no bird now.
type formFixture struct {
	t                  *testing.T
	d                  *testdb.DB
	vet, mgr           auth.Principal
	hirbovat, telita   uuid.UUID
	cycle              uuid.UUID
	breeder, cocosi    uuid.UUID
	universal          uuid.UUID
	entryVet, entryMgr uuid.UUID
	sec                map[string]uuid.UUID
	series26, series25 uuid.UUID
}

func fday(m time.Month, d int) timex.Day { return timex.MustDay(2026, m, d) }

func newFormFixture(t *testing.T) *formFixture {
	t.Helper()
	f := &formFixture{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	f.series26 = f.id(`insert into series (name, name_key, year, placement_date) values ('Seria 2026', 'seria 2026', 2026, '2026-01-01')`)
	f.series25 = f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, '2025-01-01', 'closed', $1)`, testdb.Now())
	guide := f.id(`insert into guide (name, name_key, type) values ('G ouat', 'g ouat', 'laying')`)
	f.exec(`insert into guide_week (guide_id, week, sex, feed_g) values ($1, 46, 'F', 157), ($1, 46, 'M', null)`, guide)
	f.hirbovat = f.id(`insert into farm (name, name_key, type, guide_id) values ('Hîrbovăț', 'hirbovat', 'laying', $1)`, guide)
	f.telita = f.id(`insert into farm (name, name_key, type) values ('Telița', 'telita', 'growth')`)
	f.vet = f.user("Ana", "ana", auth.Vet, f.hirbovat, f.telita)
	f.mgr = f.user("Ion", "ion", auth.Manager, f.hirbovat, f.telita)

	f.breeder = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')`)
	f.cocosi = f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M')`)
	f.universal = f.id(`insert into feed_type (name, name_key) values ('Universal', 'universal')`)
	f.id(`insert into feed_type (name, name_key, for_sex, active) values ('Vechi', 'vechi', 'F', false)`)

	h5 := f.section(f.hirbovat, "H5", nil, 1)
	h7 := f.section(f.hirbovat, "H7", nil, 2)
	f.section(f.hirbovat, "H5.2", &h5, 2) // before H5.1: the order comes from sort_order
	f.section(f.hirbovat, "H5.1", &h5, 1)
	f.section(f.hirbovat, "H7.1", &h7, 1)
	f.section(f.hirbovat, "H8", nil, 3)
	f.section(f.telita, "T1", nil, 1)

	f.cycle = f.id(`insert into cycle (series_id, farm_id, guide_id, start_date, started_by)
		values ($1, $2, $3, '2026-01-01', 'auto')`, f.series26, f.hirbovat, guide)
	f.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, '2025-01-01', '2025-12-31', 'auto', 'emptied')`, f.series25, f.telita)

	H := f.hirbovat
	e := f.entry("placement", H, fday(time.January, 1), &f.mgr.UserID, testdb.Now())
	for _, b := range []struct {
		code, sex string
		qty       int
	}{{"H5.1", "F", 100}, {"H5.1", "M", 10}, {"H5.2", "F", 50}, {"H7.1", "F", 30}, {"H8", "F", 5}} {
		f.bird(e, H, b.code, fday(time.January, 1), b.sex, "placement", b.qty)
	}
	f.exec(`update section set retired_at = $1 where id = $2`, testdb.Now(), f.sec["H8"])
	e = f.entry("sale", H, fday(time.November, 14), &f.mgr.UserID, testdb.Now())
	f.bird(e, H, "H7.1", fday(time.November, 14), "F", "sale", -30)

	e = f.entry("delivery", H, fday(time.November, 1), &f.mgr.UserID, testdb.Now())
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, '2026-11-01', 1000), ($1, $2, $4, '2026-11-01', 100)`,
		e, H, f.breeder, f.cocosi)

	e = f.entry("feed_daily", H, fday(time.November, 16), &f.mgr.UserID, testdb.Now())
	f.alloc(e, "H5.1", fday(time.November, 16), "F", f.breeder, "150", "")
	f.alloc(e, "H5.1", fday(time.November, 16), "M", f.cocosi, "140", "")
	f.alloc(e, "H5.2", fday(time.November, 16), "F", f.universal, "150", "")
	e = f.entry("feed_daily", H, fday(time.November, 17), &f.mgr.UserID, testdb.Now())
	f.alloc(e, "H5.1", fday(time.November, 17), "F", f.breeder, "155", "")
	e = f.entry("migration", H, fday(time.November, 17), nil, time.Date(2026, time.November, 1, 10, 0, 0, 0, timex.Chisinau))
	f.alloc(e, "H5.1", fday(time.November, 17), "M", f.cocosi, "", "1.5")

	at := func(h, m int) time.Time { return time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau) }
	f.entryVet = f.entry("feed_daily", H, fday(time.November, 18), &f.vet.UserID, at(8, 44))
	f.alloc(f.entryVet, "H5.1", fday(time.November, 18), "F", f.breeder, "160", "")
	f.entryMgr = f.entry("feed_daily", H, fday(time.November, 18), &f.mgr.UserID, at(9, 0))
	f.alloc(f.entryMgr, "H5.1", fday(time.November, 18), "M", f.cocosi, "145", "")
	return f
}

func (f *formFixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var id uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (f *formFixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatalf("%s: %v", sql, err)
	}
}

func (f *formFixture) user(name, login string, role auth.Role, farms ...uuid.UUID) auth.Principal {
	id := f.id(`insert into app_user (name, login, password_hash, role) values ($1, $2, 'x', $3)`, name, login, string(role))
	return auth.Principal{UserID: id, Name: name, Role: role, FarmIDs: farms}
}

func (f *formFixture) section(farm uuid.UUID, code string, parent *uuid.UUID, sort int) uuid.UUID {
	f.t.Helper()
	id := f.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, farm, code, parent, sort)
	f.sec[code] = id
	return id
}

func (f *formFixture) entry(kind string, farm uuid.UUID, d timex.Day, by *uuid.UUID, at time.Time) uuid.UUID {
	f.t.Helper()
	var series *uuid.UUID
	if kind == "placement" {
		series = &f.series26
	}
	source := "manual"
	if kind == "migration" {
		source = "migration"
	}
	return f.id(`insert into entry (kind, farm_id, date, series_id, created_by, created_at, source)
		values ($1, $2, $3, $4, $5, $6, $7)`, kind, farm, d, series, by, at, source)
}

func (f *formFixture) bird(entry, farm uuid.UUID, code string, d timex.Day, sex, typ string, qty int) {
	f.t.Helper()
	f.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty) values ($1, $2, $3, $4, $5, $6, $7)`,
		entry, farm, f.sec[code], d, sex, typ, qty)
}

// alloc inserts a live allocation; g or kg "" is null.
func (f *formFixture) alloc(entry uuid.UUID, code string, d timex.Day, sex string, typ uuid.UUID, g, kg string) {
	f.t.Helper()
	null := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	f.id(`insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird, kg_recorded)
		values ($1, $2, $3, $4, $5, $6, $7::numeric, $8::numeric)`, entry, f.hirbovat, f.sec[code], d, sex, typ, null(g), null(kg))
}

func (f *formFixture) build(p auth.Principal, farm uuid.UUID, d timex.Day) FeedFormResponse {
	f.t.Helper()
	res, err := BuildFeedForm(f.t.Context(), f.d.App, FeedFormInput{Principal: p, FarmID: farm, Date: d, Today: testdb.Today, Now: testdb.Now()})
	if err != nil {
		f.t.Fatal(err)
	}
	return res
}

func numStr(n *num.Num) string {
	if n == nil {
		return "null"
	}
	return n.String()
}

func idStr(id *uuid.UUID) string {
	if id == nil {
		return "null"
	}
	return id.String()
}

// T5 criterion 1: the Q1 shape per D8 on hand-built data.
func TestBuildFeedForm(t *testing.T) {
	f := newFormFixture(t)
	res := f.build(f.mgr, f.hirbovat, testdb.Today)

	if res.Farm != (FormFarm{ID: f.hirbovat, Name: "Hîrbovăț", Type: dbq.FarmTypeLaying}) || res.Date != testdb.Today {
		t.Errorf("farm, date = %+v, %s", res.Farm, res.Date)
	}
	if res.SAPT == nil || *res.SAPT != 46 || res.AgeDays == nil || *res.AgeDays != 321 {
		t.Errorf("sapt, ageDays = %v, %v; want 46, 321", res.SAPT, res.AgeDays)
	}
	if res.Cycle == nil || *res.Cycle != (FormCycle{ID: f.cycle, Series: "Seria 2026", StartDate: fday(time.January, 1)}) {
		t.Errorf("cycle = %+v", res.Cycle)
	}
	if res.Guide == nil || numStr(res.Guide.F) != "157" || res.Guide.M != nil {
		t.Errorf("guide = %+v, want {157, null}", res.Guide)
	}
	if res.LockedReason != nil {
		t.Errorf("lockedReason = %q, want null", *res.LockedReason)
	}

	// Rows: only sexes with birds at the start (no H5.2 ♂, no emptied H7.1,
	// no retired H8), tree order, ♀ before ♂.
	type row struct{ code, hall, sex, birds, guide, sugG, sugType, existing string }
	exist := func(e *FeedRowExisting) string {
		if e == nil {
			return "null"
		}
		return e.EntryID.String() + " " + numStr(e.G) + " " + e.FeedTypeID.String() + " " + e.Kg.String()
	}
	var got []row
	for _, r := range res.Rows {
		got = append(got, row{
			r.Code, r.HallCode, string(r.Sex), strconv.Itoa(r.BirdsStart), numStr(r.GuideG),
			numStr(r.Suggestion.G), idStr(r.Suggestion.FeedTypeID), exist(r.Existing),
		})
	}
	want := []row{
		// suggestion g from yesterday (155); type = last used; existing = the vet's 160 g × 100 ÷ 1000 = 16 kg.
		{"H5.1", "H5", "F", "100", "157", "155", f.breeder.String(), f.entryVet.String() + " 160 " + f.breeder.String() + " 16"},
		// yesterday was kg-only: g null; type Cocoși from it; no guide cell.
		{"H5.1", "H5", "M", "10", "null", "null", f.cocosi.String(), f.entryMgr.String() + " 145 " + f.cocosi.String() + " 1.45"},
		// nothing yesterday: g null; type = last used on an earlier day (16.11).
		{"H5.2", "H5", "F", "50", "157", "null", f.universal.String(), "null"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("rows\n got %v\nwant %v", got, want)
	}
	if res.Rows[0].SectionID != f.sec["H5.1"] || res.Rows[2].SectionID != f.sec["H5.2"] {
		t.Errorf("row section ids = %s, %s", res.Rows[0].SectionID, res.Rows[2].SectionID)
	}

	if want := []FormEmptyHall{{Hall: "H7", EmptySince: fday(time.November, 14), Label: "goală din 14.11.2026 · H7.1"}}; !slices.Equal(res.EmptyHalls, want) {
		t.Errorf("emptyHalls = %+v, want %+v", res.EmptyHalls, want)
	}

	// The manager may correct both entries and edit only his own (same day).
	ana, ion := "Ana", "Ion"
	at := func(h, m int) timex.LocalTime {
		return timex.At(time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau))
	}
	wantEntries := []FormEntry{
		{EntryID: f.entryVet, Author: &ana, At: at(8, 44), Rows: 1, Editable: false, Correctable: true},
		{EntryID: f.entryMgr, Author: &ion, At: at(9, 0), Rows: 1, Editable: true, Correctable: true},
	}
	if !entriesEqual(res.Existing, wantEntries) {
		t.Errorf("existing = %+v,\nwant %+v", res.Existing, wantEntries)
	}

	// Feed types: the active ones, in name order.
	var types []string
	for _, ft := range res.FeedTypes {
		types = append(types, ft.Name+":"+strPtr(ft.ForSex))
	}
	if want := []string{"Breeder 2:F", "Cocoși:M", "Universal:null"}; !slices.Equal(types, want) {
		t.Errorf("feedTypes = %v, want %v", types, want)
	}

	// Stock as of today on the saved data. H5.2 ♀ and the retired H8 ♀ have
	// birds and no allocation today, so the window is 15–17.11:
	//   Breeder 2: 1000 − (15 + 15,5 + 16) = 953,5; avg (15 + 15,5) ÷ 3;
	//   Cocoși: 100 − (1,4 + 1,5 + 1,45) = 95,65; avg (1,4 + 1,5) ÷ 3;
	//   Universal: 0 − 7,5 = −7,5 → stoc sub 0, 0 days, runs out on 17.11.
	type stock struct{ typ, sex, kg, avg, days, runOut, lamp, reason string }
	var gotStock []stock
	for _, s := range res.Stock {
		if s.Window != [2]timex.Day{fday(time.November, 15), fday(time.November, 17)} {
			t.Errorf("%s window = %v", s.Type, s.Window)
		}
		runOut := "null"
		if s.RunOut != nil {
			runOut = s.RunOut.String()
		}
		gotStock = append(gotStock, stock{s.Type, strPtr(s.ForSex), wire(&s.Kg), wire(&s.AvgDaily), wire(s.DaysLeft), runOut, string(s.Lamp), s.Reason})
	}
	wantStock := []stock{
		{"Breeder 2", "F", "953.5", "10.1667", "93.7869", "2027-02-18", "in_norma", ""},
		{"Cocoși", "M", "95.65", "0.9667", "98.9483", "2027-02-23", "in_norma", ""},
		{"Universal", "null", "-7.5", "2.5", "0", "2026-11-17", "abatere", "stoc sub 0 · verifică livrările"},
	}
	if !slices.Equal(gotStock, wantStock) {
		t.Errorf("stock\n got %v\nwant %v", gotStock, wantStock)
	}
	if res.Stock[0].FeedTypeID != f.breeder {
		t.Errorf("stock[0].feedTypeId = %s, want Breeder 2", res.Stock[0].FeedTypeID)
	}
	if res.AsOf != timex.At(testdb.Now()) {
		t.Errorf("asOf = %v", res.AsOf)
	}
}

// wire is n as the JSON number the client gets (≤ 4 decimals), "null" for nil.
func wire(n *num.Num) string {
	b, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func strPtr(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}

func entriesEqual(a, b []FormEntry) bool {
	return slices.EqualFunc(a, b, func(x, y FormEntry) bool {
		return x.EntryID == y.EntryID && strPtr(x.Author) == strPtr(y.Author) && x.At.Equal(y.At.Time) &&
			x.Rows == y.Rows && x.Editable == y.Editable && x.Correctable == y.Correctable
	})
}

// T5 criterion 1: the entries' actions follow the caller (§3.2), and a vet
// on a past day gets the lockedReason; today has none.
func TestBuildFeedFormVet(t *testing.T) {
	f := newFormFixture(t)
	res := f.build(f.vet, f.hirbovat, testdb.Today)
	if res.LockedReason != nil {
		t.Errorf("vet today: lockedReason = %q, want null", *res.LockedReason)
	}
	ana, ion := "Ana", "Ion"
	at := func(h, m int) timex.LocalTime {
		return timex.At(time.Date(2026, time.November, 18, h, m, 0, 0, timex.Chisinau))
	}
	want := []FormEntry{
		{EntryID: f.entryVet, Author: &ana, At: at(8, 44), Rows: 1, Editable: true},
		{EntryID: f.entryMgr, Author: &ion, At: at(9, 0), Rows: 1},
	}
	if !entriesEqual(res.Existing, want) {
		t.Errorf("existing = %+v,\nwant %+v", res.Existing, want)
	}

	past := f.build(f.vet, f.hirbovat, fday(time.November, 17))
	if past.LockedReason == nil || *past.LockedReason != "Pentru zilele trecute cere managerului." {
		t.Errorf("vet past day: lockedReason = %v", past.LockedReason)
	}
	// 17.11: H5.1 ♂ is the migration's kg-only row: existing g null, kg 1,5; its entry has no author.
	if e := past.Rows[1].Existing; e == nil || e.G != nil || e.Kg.String() != "1.5" {
		t.Errorf("17.11 H5.1 ♂ existing = %+v, want kg-only 1.5", e)
	}
	if n := len(past.Existing); n != 2 || past.Existing[0].Author != nil || past.Existing[0].Editable || past.Existing[0].Correctable {
		t.Errorf("17.11 existing = %+v, want the (older) migration entry without author or actions, then the feed_daily one", past.Existing)
	}
	if mgr := f.build(f.mgr, f.hirbovat, fday(time.November, 17)); mgr.LockedReason != nil {
		t.Errorf("manager past day: lockedReason = %q, want null", *mgr.LockedReason)
	}
}

// T5 criterion 1: a day in a cycle of a closed series is read-only for
// everyone; a farm without birds has rows [] and, without an open cycle, no
// stock.
func TestBuildFeedFormClosedSeriesAndEmptyFarm(t *testing.T) {
	f := newFormFixture(t)
	closed := f.build(f.mgr, f.telita, timex.MustDay(2025, time.June, 1))
	if closed.LockedReason == nil || *closed.LockedReason != "Seria 2025 e închisă: nu mai primește înregistrări sau corecții." {
		t.Errorf("closed series: lockedReason = %v", closed.LockedReason)
	}
	if closed.Cycle == nil || closed.Cycle.Series != "Seria 2025" || closed.Guide != nil {
		t.Errorf("closed series: cycle %+v guide %+v, want Seria 2025 without a guide", closed.Cycle, closed.Guide)
	}

	empty := f.build(f.mgr, f.telita, testdb.Today)
	if empty.Cycle != nil || empty.SAPT != nil || empty.AgeDays != nil || empty.Guide != nil || empty.LockedReason != nil {
		t.Errorf("no cycle today: %+v", empty)
	}
	b, err := json.Marshal(empty)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"rows":[]`, `"emptyHalls":[]`, `"existing":[]`, `"stock":[]`, `"cycle":null`, `"guide":null`, `"sapt":null`} {
		if !strings.Contains(string(b), key) {
			t.Errorf("JSON %s has no %s", b, key)
		}
	}
	if len(empty.FeedTypes) != 3 {
		t.Errorf("feedTypes = %+v, want the 3 active ones", empty.FeedTypes)
	}
	if _, err := BuildFeedForm(t.Context(), f.d.App, FeedFormInput{FarmID: uuid.New(), Date: testdb.Today, Today: testdb.Today}); err != ErrFormFarmNotFound {
		t.Errorf("unknown farm: err = %v, want ErrFormFarmNotFound", err)
	}
}

// T5 criterion 1: the service's access and date checks.
func TestFeedDailyFormService(t *testing.T) {
	f := newFormFixture(t)
	svc := NewService(db.NewStore(f.d.App), f.d.Clock)
	ctx := t.Context()
	viewer := auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
	vetT := auth.Principal{UserID: uuid.New(), Role: auth.Vet, FarmIDs: []uuid.UUID{f.telita}}
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}

	_, err := svc.FeedDailyForm(ctx, viewer, f.hirbovat, testdb.Today)
	if body := wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden); body.Message != "Nu ai acces la această pagină." {
		t.Errorf("viewer message = %q", body.Message)
	}
	_, err = svc.FeedDailyForm(ctx, vetT, f.hirbovat, testdb.Today)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.FeedDailyForm(ctx, vetT, uuid.Nil, testdb.Today)
	wantCode(t, err, http.StatusForbidden, httpx.CodeFarmForbidden)
	_, err = svc.FeedDailyForm(ctx, owner, uuid.New(), testdb.Today)
	wantCode(t, err, http.StatusNotFound, httpx.CodeNotFound)

	_, err = svc.FeedDailyForm(ctx, f.mgr, f.hirbovat, fday(time.November, 19))
	body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation)
	want := httpx.Issue{Path: "date", Code: httpx.IssueOutOfRange, Message: "19.11.2026 e în viitor. Alege o dată până la 18.11.2026."}
	if len(body.Issues) != 1 || body.Issues[0] != want || body.Message != want.Message {
		t.Errorf("future date = %+v, want %+v", body, want)
	}

	res, err := svc.FeedDailyForm(ctx, f.vet, f.hirbovat, testdb.Today)
	if err != nil || len(res.Rows) != 3 {
		t.Fatalf("vet today: %d rows, %v", len(res.Rows), err)
	}
}

// T5 criterion 1: the query parameters.
func TestParseFormQuery(t *testing.T) {
	id := uuid.New()
	for _, tc := range []struct {
		name  string
		query string
		want  []httpx.Issue // nil = ok
	}{
		{"ok", "farmId=" + id.String() + "&date=2026-11-18", nil},
		{"both missing", "", []httpx.Issue{
			{Path: "farmId", Code: httpx.IssueRequired, Message: "Datele trimise nu sunt valide."},
			{Path: "date", Code: httpx.IssueRequired, Message: "Datele trimise nu sunt valide."},
		}},
		{"bad values", "farmId=x&date=2026-02-30", []httpx.Issue{
			{Path: "farmId", Code: httpx.IssueBadType, Message: "Datele trimise nu sunt valide."},
			{Path: "date", Code: httpx.IssueBadDate, Message: "Datele trimise nu sunt valide."},
		}},
		{"repeated and unknown", "farmId=" + id.String() + "&date=2026-11-18&date=2026-11-17&tab=x", []httpx.Issue{
			{Path: "date", Code: httpx.IssueBadType, Message: "Datele trimise nu sunt valide."},
			{Path: "tab", Code: httpx.IssueUnknownField, Message: "Datele trimise nu sunt valide."},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest(http.MethodGet, "/api/forms/feed-daily?"+tc.query, nil)
			farm, d, err := parseFormQuery(r.URL.Query())
			if tc.want == nil {
				if err != nil || farm != id || d != testdb.Today {
					t.Errorf("= %s, %s, %v", farm, d, err)
				}
				return
			}
			body := wantCode(t, err, http.StatusBadRequest, httpx.CodeValidation)
			if !slices.Equal(body.Issues, tc.want) {
				t.Errorf("issues = %+v,\nwant %+v", body.Issues, tc.want)
			}
		})
	}
}

// formStatements is a pgx QueryTracer that records every statement by its
// "-- name: X" header (the SQL text of anything else, e.g. "begin …").
type formStatements struct {
	mu    sync.Mutex
	names []string
}

func (s *formStatements) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := strings.ToLower(strings.TrimSpace(d.SQL))
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
	}
	s.mu.Lock()
	s.names = append(s.names, name)
	s.mu.Unlock()
	return ctx
}

func (s *formStatements) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (s *formStatements) take() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.names
	s.names = nil
	return out
}

// T5 criterion 3: every read in one InReadTx, at most 6 statements.
func TestFeedDailyFormStatements(t *testing.T) {
	f := newFormFixture(t)
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

	run := func(farm uuid.UUID, want ...string) {
		t.Helper()
		if _, err := svc.FeedDailyForm(t.Context(), f.mgr, farm, testdb.Today); err != nil {
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
	head := []string{"FormsFeedHead", "FormsFeedRows"}
	run(f.hirbovat, slices.Concat(head, []string{"CalcSettings", "CalcStockEnd", "CalcFeedDayComplete", "CalcConsumption"})...)
	run(f.telita, head...) // no open cycle: no stock read
}

// The stock rows are calc.ReadStockDays' figures (one source with the
// summary and Rapoarte › Hrană, golden 28).
func TestFeedFormStockMatchesCalc(t *testing.T) {
	f := newFormFixture(t)
	res := f.build(f.mgr, f.hirbovat, testdb.Today)
	settings, err := calc.ReadAppSettings(t.Context(), f.d.App)
	if err != nil {
		t.Fatal(err)
	}
	sd, err := calc.ReadStockDays(t.Context(), f.d.App, f.hirbovat, testdb.Today, settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(sd.Items) != len(res.Stock) {
		t.Fatalf("stock rows %d, calc %d", len(res.Stock), len(sd.Items))
	}
	for i, it := range sd.Items {
		s := res.Stock[i]
		if s.FeedTypeID != it.FeedTypeID || !s.Kg.Decimal().Equal(it.Kg) || !s.AvgDaily.Decimal().Equal(it.AvgDaily) {
			t.Errorf("row %d = %+v, calc %+v", i, s, it)
		}
	}
}

// The handler: query → service → JSON, errors through the envelope.
func TestFeedFormHandler(t *testing.T) {
	f := newFormFixture(t)
	h := NewHandler(NewService(db.NewStore(f.d.App), f.d.Clock))
	serve := func(p *auth.Principal, query string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, "/api/forms/feed-daily?"+query, nil)
		if p != nil {
			r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
		}
		w := httptest.NewRecorder()
		h.feedForm(w, r)
		return w
	}
	q := "farmId=" + f.hirbovat.String() + "&date=2026-11-18"
	w := serve(&f.mgr, q)
	var res FeedFormResponse
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &res) != nil || len(res.Rows) != 3 || len(res.Stock) != 3 {
		t.Errorf("200: %d %s", w.Code, w.Body)
	}
	if w := serve(&f.mgr, "farmId="+f.hirbovat.String()); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), `"path":"date","code":"required"`) {
		t.Errorf("no date: %d %s", w.Code, w.Body)
	}
	viewer := auth.Principal{UserID: uuid.New(), Role: auth.Viewer, FarmIDs: []uuid.UUID{f.hirbovat}}
	if w := serve(&viewer, q); w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":"FORBIDDEN"`) {
		t.Errorf("viewer: %d %s", w.Code, w.Body)
	}
	if w := serve(nil, q); w.Code != http.StatusUnauthorized {
		t.Errorf("no session: %d %s", w.Code, w.Body)
	}
}
