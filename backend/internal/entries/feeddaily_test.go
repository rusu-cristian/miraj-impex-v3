package entries

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// feedFixture is a laying farm (the save fixture's farm A, "Hîrbovăț") on
// Seria T, placed 01.11.2026 at age 0, so 17.11 and 18.11 are SAPT 3; its
// guide has feed_g ♀ 157 · ♂ 156 there. Birds at the start of 18.11:
//
//	H5.1 ♀ 1 000 · ♂ 100   H5.2 ♀ 500 · ♂ 50   H6.1 ♀ 200 (no ♂)
//	H7.1 ♀ 0 (sold out 14.11; H7 is its hall)   H8 retired, no birds
//
// Stock: Breeder 2 (♀) 1 000 kg and Cocoși (♂) 100 kg delivered 01.11;
// Breeder 1 (♀) is retired. Farm B ("Telița") sits in the closed Seria
// 2025 from 01.10 to 10.11.2026.
type feedFixture struct {
	*saveFixture
	svc                          *Service
	sec                          map[string]uuid.UUID
	breeder2, cocosi, breeder1   uuid.UUID
	vet, viewer, owner, outsider auth.Principal
}

var (
	nov1  = timex.MustDay(2026, time.November, 1)
	nov14 = timex.MustDay(2026, time.November, 14)
	nov17 = timex.MustDay(2026, time.November, 17)
	nov18 = timex.MustDay(2026, time.November, 18)
	nov19 = timex.MustDay(2026, time.November, 19)
)

func newFeedFixture(t *testing.T) *feedFixture {
	t.Helper()
	f := &feedFixture{saveFixture: newSaveFixture(t), sec: map[string]uuid.UUID{}}
	f.svc = NewService(db.NewStore(f.d.App), f.d.Clock)
	A := f.farmA

	guide := f.id(`insert into guide (name, name_key, type) values ('Ross T', 'ross t', 'laying') returning id`)
	for week := 1; week <= 5; week++ {
		f.exec(`insert into guide_week (guide_id, week, sex, feed_g) values ($1, $2, 'F', 157), ($1, $2, 'M', 156)`, guide, week)
	}
	series := f.id(`insert into series (name, name_key, year, placement_date) values ('Seria T', 'seria t', 2030, $1) returning id`, nov1)
	f.exec(`insert into cycle (series_id, farm_id, guide_id, start_date, started_by) values ($1, $2, $3, $4, 'auto')`,
		series, A, guide, nov1)
	closed := f.id(`insert into series (name, name_key, year, placement_date, status, closed_at)
		values ('Seria 2025', 'seria 2025', 2025, $1, 'closed', $2) returning id`, timex.MustDay(2026, time.October, 1), testdb.Now())
	f.exec(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by)
		values ($1, $2, $3, $4, 'auto', 'emptied')`, closed, f.farmB, timex.MustDay(2026, time.October, 1), timex.MustDay(2026, time.November, 10))
	f.sec["T1"] = f.id(`insert into section (farm_id, code) values ($1, 'T1') returning id`, f.farmB)

	for i, code := range []string{"H5", "H6", "H7"} {
		f.sec[code] = f.id(`insert into section (farm_id, code, sort_order) values ($1, $2, $3) returning id`, A, code, i+1)
	}
	for _, code := range []string{"H5.1", "H5.2", "H6.1", "H7.1"} {
		f.sec[code] = f.id(`insert into section (farm_id, code, parent_id) values ($1, $2, $3) returning id`,
			A, code, f.sec[code[:2]])
	}
	f.sec["H8"] = f.id(`insert into section (farm_id, code, retired_at) values ($1, 'H8', $2) returning id`, A, testdb.Now())

	placement := f.id(`insert into entry (kind, farm_id, date, series_id) values ('placement', $1, $2, $3) returning id`, A, nov1, series)
	for _, b := range []struct {
		code, sex string
		qty       int
	}{{"H5.1", "F", 1000}, {"H5.1", "M", 100}, {"H5.2", "F", 500}, {"H5.2", "M", 50}, {"H6.1", "F", 200}, {"H7.1", "F", 300}} {
		f.bird(placement, b.code, nov1, b.sex, "placement", b.qty)
	}
	sale := f.id(`insert into entry (kind, farm_id, date) values ('sale', $1, $2) returning id`, A, nov14)
	f.bird(sale, "H7.1", nov14, "F", "sale", -300)

	f.breeder2 = f.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F') returning id`)
	f.cocosi = f.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M') returning id`)
	f.breeder1 = f.id(`insert into feed_type (name, name_key, for_sex, active) values ('Breeder 1', 'breeder 1', 'F', false) returning id`)
	f.deliver(f.breeder2, nov1, "1000")
	f.deliver(f.cocosi, nov1, "100")

	user := func(name, login, role string, farms ...uuid.UUID) auth.Principal {
		id := f.id(`insert into app_user (name, login, password_hash, role) values ($1, $2, 'x', $3) returning id`, name, login, role)
		if farms == nil && role != "owner" {
			farms = []uuid.UUID{}
		}
		return auth.Principal{UserID: id, Name: name, Role: auth.Role(role), FarmIDs: farms}
	}
	f.vet = user("Dr. Ceban", "ceban", "vet", A, f.farmB)
	f.viewer = user("Ana", "ana", "viewer", A)
	f.owner = user("Proprietar", "owner", "owner")
	f.outsider = user("Vasile", "vasile", "manager", f.farmB)
	return f
}

func (f *feedFixture) exec(sql string, args ...any) {
	f.t.Helper()
	if _, err := f.d.App.Exec(context.Background(), sql, args...); err != nil {
		f.t.Fatal(err)
	}
}

func (f *feedFixture) bird(entry uuid.UUID, code string, d timex.Day, sex, typ string, qty int) {
	f.t.Helper()
	f.exec(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty) values ($1, $2, $3, $4, $5, $6, $7)`,
		entry, f.farmA, f.sec[code], d, sex, typ, qty)
}

func (f *feedFixture) deliver(ft uuid.UUID, d timex.Day, kg string) {
	f.t.Helper()
	e := f.id(`insert into entry (kind, farm_id, date) values ('delivery', $1, $2) returning id`, f.farmA, d)
	f.exec(`insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`,
		e, f.farmA, ft, d, kg)
}

func (f *feedFixture) row(code string, sex dbq.Sex, ft uuid.UUID, g string) FeedDailyRow {
	n := num.MustParse(g)
	return FeedDailyRow{SectionID: f.sec[code], Sex: sex, FeedTypeID: ft, G: &n}
}

// allRows are the five rows with birds on 18.11 (no warning: g within 10% of
// the guide): 160 + 75 + 31,4 kg Breeder 2, 15 + 7,5 kg Cocoși.
func (f *feedFixture) allRows() []FeedDailyRow {
	return []FeedDailyRow{
		f.row("H5.1", dbq.SexF, f.breeder2, "160"),
		f.row("H5.1", dbq.SexM, f.cocosi, "150"),
		f.row("H5.2", dbq.SexF, f.breeder2, "150"),
		f.row("H5.2", dbq.SexM, f.cocosi, "150"),
		f.row("H6.1", dbq.SexF, f.breeder2, "157"),
	}
}

func (f *feedFixture) request(d timex.Day, key *uuid.UUID, rows ...FeedDailyRow) FeedDailyCreateRequest {
	return FeedDailyCreateRequest{WriteMeta: httpx.WriteMeta{ClientRequestID: key}, FarmID: f.farmA, Date: d, Rows: rows}
}

func (f *feedFixture) save(p auth.Principal, req FeedDailyCreateRequest) (Result, error) {
	f.t.Helper()
	return f.svc.CreateFeedDaily(context.Background(), p, req)
}

// errBody is the envelope of err, or a failure when err isn't one.
func errBody(t *testing.T, err error) (int, httpx.APIError) {
	t.Helper()
	var he httpx.HTTPError
	if err == nil || !errors.As(err, &he) {
		t.Fatalf("err = %v, want an envelope error", err)
	}
	return he.HTTPError()
}

// T4 criterion 2: 201 with the D6 figures, one audit row, one entry.created
// event, the rows written with their g.
func TestCreateFeedDailySaves(t *testing.T) {
	ctx := context.Background()
	f := newFeedFixture(t)
	key := uuid.New()
	res, err := f.save(f.ion, f.request(nov18, &key, f.allRows()...))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusCreated {
		t.Fatalf("status = %d", res.Status)
	}
	var resp FeedDailyResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	body := string(res.Body)
	for _, want := range []string{
		`"kgTotal":288.9`,
		`"byType":[{"type":"Breeder 2","kg":266.4,"stockAfter":733.6},{"type":"Cocoși","kg":22.5,"stockAfter":77.5}]`,
		`"toast":"Hrană zilnică salvată · 5 rânduri · 288,9 kg · stoc Breeder 2: 734 kg"`,
		`"warnings":[]`,
		`"asOf":"2026-11-18T12:00:00+02:00"`,
		`"entry":{"id":"` + resp.Entry.ID.String() + `","kind":"feed_daily","version":"2026-11-18T10:00:00.000000Z"}`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body lacks %s:\n%s", want, body)
		}
	}
	if !strings.HasPrefix(body, `{"entry":`) || !strings.Contains(body, `,"kgTotal":`) {
		t.Errorf("key order: %s", body)
	}

	rows, err := f.d.App.Query(ctx, `select g_per_bird::text from feed_allocation where entry_id = $1 and kg_recorded is null
		order by g_per_bird`, resp.Entry.ID)
	if err != nil {
		t.Fatal(err)
	}
	gs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(gs, []string{"150.0", "150.0", "150.0", "157.0", "160.0"}) {
		t.Errorf("rows g = %v", gs)
	}

	var (
		action, summary, entity string
		recordTypes             []string
		farmIDs                 []uuid.UUID
		recordDate              timex.Day
		changes, newValues      string
	)
	if err := f.d.App.QueryRow(ctx, `select action::text, entity, summary, record_types, farm_ids, record_date,
		changes::text, new_values::text from audit_log where entry_id = $1`, resp.Entry.ID).
		Scan(&action, &entity, &summary, &recordTypes, &farmIDs, &recordDate, &changes, &newValues); err != nil {
		t.Fatal(err)
	}
	if action != "create" || entity != "entry" || summary != "Hrană zilnică · 5 din 5 rânduri · 288,9 kg" ||
		!slices.Equal(recordTypes, []string{"hrana"}) || !slices.Equal(farmIDs, []uuid.UUID{f.farmA}) ||
		recordDate != nov18 || changes != "[]" {
		t.Errorf("audit = %s %s %q %v %v %s %s", action, entity, summary, recordTypes, farmIDs, recordDate, changes)
	}
	var snapshot []feedAuditRow
	if err := json.Unmarshal([]byte(newValues), &snapshot); err != nil {
		t.Fatal(err)
	}
	if len(snapshot) != 5 || snapshot[0].Section != "H5.1" || snapshot[0].Sex != dbq.SexF ||
		snapshot[0].FeedTypeID != f.breeder2 || snapshot[0].G.String() != "160" || snapshot[0].Kg.String() != "160" {
		t.Errorf("new_values = %s", newValues)
	}

	var evType, payload string
	if err := f.d.App.QueryRow(ctx, `select type, payload::text from live_event`).Scan(&evType, &payload); err != nil {
		t.Fatal(err)
	}
	var ev struct {
		EntryID     uuid.UUID `json:"entryId"`
		Kind        string    `json:"kind"`
		Sections    []string  `json:"sections"`
		RecordTypes []string  `json:"recordTypes"`
		Author      string    `json:"author"`
		Summary     string    `json:"summary"`
	}
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		t.Fatal(err)
	}
	if evType != "entry.created" || ev.EntryID != resp.Entry.ID || ev.Kind != "feed_daily" ||
		!slices.Equal(ev.Sections, []string{"H5.1", "H5.2", "H6.1"}) || !slices.Equal(ev.RecordTypes, []string{"hrana"}) ||
		ev.Author != "Ion Rusu" || ev.Summary != summary {
		t.Errorf("live event %s %s", evType, payload)
	}
	if got := f.count("audit_log") + f.count("live_event") + f.count("request_replay"); got != 3 {
		t.Errorf("audit + live + replay rows = %d, want 3", got)
	}
}

// The toast names the type with the largest kg in the save, and the summary
// counts the rows with birds that day ("2 din 5").
func TestCreateFeedDailyToastType(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.save(f.ion, f.request(nov18, nil,
		f.row("H5.1", dbq.SexM, f.cocosi, "150"), f.row("H5.2", dbq.SexF, f.breeder2, "20")))
	if err != nil {
		t.Fatal(err)
	}
	// Cocoși 15 kg > Breeder 2 10 kg.
	if want := `"toast":"Hrană zilnică salvată · 2 rânduri · 25,0 kg · stoc Cocoși: 85 kg"`; !strings.Contains(string(res.Body), want) {
		t.Errorf("body = %s, want %s", res.Body, want)
	}
	var summary string
	if err := f.d.App.QueryRow(context.Background(), `select summary from audit_log`).Scan(&summary); err != nil {
		t.Fatal(err)
	}
	if summary != "Hrană zilnică · 2 din 5 rânduri · 25,0 kg" {
		t.Errorf("summary = %q", summary)
	}
}

// H2 and H3 refusals come together, in document order, in one 422; nothing
// is written. A dry run of the same body answers them in a 200.
func TestCreateFeedDailyRefusals(t *testing.T) {
	f := newFeedFixture(t)
	rows := []FeedDailyRow{
		f.row("H5.1", dbq.SexF, f.breeder2, "160"), // fine
		f.row("H7.1", dbq.SexF, f.breeder2, "160"), // H3: sold out on 14.11
		f.row("H7", dbq.SexF, f.breeder2, "160"),   // H3: a hall
		f.row("H8", dbq.SexF, f.breeder2, "160"),   // H3: retired
		f.row("H6.1", dbq.SexM, f.cocosi, "150"),   // H3: no roosters
		f.row("H5.2", dbq.SexF, f.breeder1, "150"), // H2: retired type
		f.row("H5.2", dbq.SexM, f.breeder2, "150"), // H2: wrong sex
	}
	want := []httpx.Refusal{
		{Field: "rows[1].sectionId", SectionCode: "H7.1", Sex: "F", Message: "H7.1 nu are găini la începutul zilei 18.11.2026."},
		{Field: "rows[2].sectionId", SectionCode: "H7", Sex: "F", Message: "H7 nu are găini la începutul zilei 18.11.2026."},
		{Field: "rows[3].sectionId", SectionCode: "H8", Sex: "F", Message: "H8 nu are găini la începutul zilei 18.11.2026."},
		{Field: "rows[4].sectionId", SectionCode: "H6.1", Sex: "M", Message: "H6.1 nu are cocoși la începutul zilei 18.11.2026."},
		{Field: "rows[5].feedTypeId", SectionCode: "H5.2", Sex: "F", Message: "Breeder 1 e retras. Alege alt tip."},
		{Field: "rows[6].feedTypeId", SectionCode: "H5.2", Sex: "M", Message: "Breeder 2 nu e pentru cocoși. Alege alt tip."},
	}
	before := f.rows()
	_, err := f.save(f.ion, f.request(nov18, nil, rows...))
	status, body := errBody(t, err)
	if status != http.StatusUnprocessableEntity || body.Code != httpx.CodeRefused || body.Message != want[0].Message {
		t.Fatalf("got %d %s %q", status, body.Code, body.Message)
	}
	if !slices.Equal(body.Refusals, want) {
		t.Errorf("refusals = %+v\nwant %+v", body.Refusals, want)
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v: a refused save wrote", before, f.rows())
	}

	req := f.request(nov18, nil, rows...)
	req.DryRun = true
	res, err := f.save(f.ion, req)
	if err != nil {
		t.Fatal(err)
	}
	var dry DryRunResponse
	if err := json.Unmarshal(res.Body, &dry); err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusOK || len(dry.Effects) != 0 || len(dry.Warnings) != 0 || !slices.Equal(dry.Refusals, want) ||
		!strings.Contains(string(res.Body), `"effects":[]`) {
		t.Errorf("dry run = %d %s", res.Status, res.Body)
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v: a dry run wrote", before, f.rows())
	}
}

// H4: g beyond tolerance.feed_pct of the guide is a warning; the save goes
// through. 140 g ♀ vs 157 is −10,8% → "sub"; 141,3 g is −10,0% → none.
func TestCreateFeedDailyGuideWarning(t *testing.T) {
	f := newFeedFixture(t)
	res, err := f.save(f.ion, f.request(nov18, nil,
		f.row("H5.2", dbq.SexM, f.cocosi, "173"),
		f.row("H5.1", dbq.SexF, f.breeder2, "140"),
		f.row("H5.2", dbq.SexF, f.breeder2, "141.3"),
	))
	if err != nil {
		t.Fatal(err)
	}
	var resp FeedDailyResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	want := []httpx.Warning{
		{Code: "FEED_GUIDE", SectionCode: "H5.2", Sex: "M",
			Message: "H5.2 ♂: 173 g, cu 11% peste ghidul de 156 g. Verifică gramajul; se poate salva."},
		{Code: "FEED_GUIDE", SectionCode: "H5.1", Sex: "F",
			Message: "H5.1 ♀: 140 g, cu 11% sub ghidul de 157 g. Verifică gramajul; se poate salva."},
	}
	if res.Status != http.StatusCreated || !slices.Equal(resp.Warnings, want) {
		t.Errorf("status %d warnings = %+v\nwant %+v", res.Status, resp.Warnings, want)
	}
}

// H5: a past-day save whose kg drive a later day below 0. Cocoși: 100 kg
// delivered, −95 adjusted on 18.11; saving 15 kg on 17.11 leaves 85 there
// and −10 at the end of 18.11, the lowest point, which the warning names.
func TestCreateFeedDailyStockNegative(t *testing.T) {
	f := newFeedFixture(t)
	e := f.id(`insert into entry (kind, farm_id, date, source) values ('migration', $1, $2, 'migration') returning id`, f.farmA, nov18)
	f.exec(`insert into feed_stock_adjustment (entry_id, farm_id, feed_type_id, date, qty_kg, note)
		values ($1, $2, $3, $4, -95, 'ajustare')`, e, f.farmA, f.cocosi, nov18)
	res, err := f.save(f.ion, f.request(nov17, nil, f.row("H5.1", dbq.SexM, f.cocosi, "150")))
	if err != nil {
		t.Fatal(err)
	}
	var resp FeedDailyResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	want := []httpx.Warning{{Code: "STOCK_NEGATIVE", Message: "Stocul de Cocoși la Hîrbovăț ar ajunge la −10 kg. Verifică livrările."}}
	if !slices.Equal(resp.Warnings, want) {
		t.Errorf("warnings = %+v, want %+v", resp.Warnings, want)
	}
	if len(resp.ByType) != 1 || resp.ByType[0].StockAfter.String() != "-10" { // the stock at the end of today (Q5)
		t.Errorf("byType = %+v", resp.ByType)
	}
	if !strings.Contains(resp.Toast, "stoc Cocoși: −10 kg") {
		t.Errorf("toast = %q", resp.Toast)
	}
}

// The access and day checks: each answers its own code and writes nothing.
func TestCreateFeedDailyChecks(t *testing.T) {
	f := newFeedFixture(t)
	closedDay := timex.MustDay(2026, time.November, 5)
	for _, tc := range []struct {
		name    string
		p       auth.Principal
		req     FeedDailyCreateRequest
		status  int
		code    string
		message string
	}{
		{"viewer", f.viewer, f.request(nov18, nil, f.allRows()...), 403, "FORBIDDEN", "Nu ai acces la această pagină."},
		{"another farm", f.outsider, f.request(nov18, nil, f.allRows()...), 403, "FARM_FORBIDDEN", "Nu ai acces la această fermă."},
		{"owner, unknown farm", f.owner, FeedDailyCreateRequest{FarmID: uuid.New(), Date: nov18, Rows: f.allRows()},
			404, "NOT_FOUND", "Adresa cerută nu există."},
		{"vet, past day", f.vet, f.request(nov17, nil, f.allRows()...), 403, "PAST_DAY_VET", "Pentru zilele trecute cere managerului."},
		{"future day", f.ion, f.request(nov19, nil, f.allRows()...), 422, "REFUSED",
			"19.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{"vet, future day", f.vet, f.request(nov19, nil, f.allRows()...), 422, "REFUSED",
			"19.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{"closed series", f.owner, FeedDailyCreateRequest{FarmID: f.farmB, Date: closedDay,
			Rows: []FeedDailyRow{{SectionID: f.sec["T1"], Sex: dbq.SexF, FeedTypeID: f.breeder2, G: f.allRows()[0].G}}},
			409, "SERIES_CLOSED", "Seria 2025 e închisă: nu mai primește înregistrări sau corecții."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := f.rows()
			_, err := f.save(tc.p, tc.req)
			status, body := errBody(t, err)
			if status != tc.status || body.Code != tc.code || body.Message != tc.message {
				t.Errorf("got %d %s %q, want %d %s %q", status, body.Code, body.Message, tc.status, tc.code, tc.message)
			}
			if f.rows() != before {
				t.Errorf("rows %v → %v", before, f.rows())
			}
		})
	}
	// The vet saves today.
	if res, err := f.save(f.vet, f.request(nov18, nil, f.allRows()...)); err != nil || res.Status != http.StatusCreated {
		t.Fatalf("vet today: %d %v", res.Status, err)
	}
}

// A second save of one live key → 409 ALREADY_ENTERED with the holder; in a
// dry run the collision is a refusal in the 200 body.
func TestCreateFeedDailyAlreadyEntered(t *testing.T) {
	f := newFeedFixture(t)
	first, err := f.save(f.ion, f.request(nov18, nil, f.row("H5.1", dbq.SexF, f.breeder2, "160")))
	if err != nil {
		t.Fatal(err)
	}
	var resp FeedDailyResponse
	if err := json.Unmarshal(first.Body, &resp); err != nil {
		t.Fatal(err)
	}
	const msg = "H5.1 ♀ are deja hrana introdusă pe 18.11.2026 la 12:00. Apasă Editează ca să schimbi valorile."
	second := f.request(nov18, nil, f.row("H5.2", dbq.SexF, f.breeder2, "150"), f.row("H5.1", dbq.SexF, f.breeder2, "150"))
	before := f.rows()
	_, err = f.save(f.maria, second)
	status, body := errBody(t, err)
	if status != http.StatusConflict || body.Code != "ALREADY_ENTERED" || body.Message != msg ||
		body.EntryID == nil || *body.EntryID != resp.Entry.ID {
		t.Fatalf("got %d %s %q %v, want 409 ALREADY_ENTERED %q %s", status, body.Code, body.Message, body.EntryID, msg, resp.Entry.ID)
	}
	if f.rows() != before {
		t.Errorf("rows %v → %v", before, f.rows())
	}

	second.DryRun = true
	res, err := f.save(f.maria, second)
	if err != nil {
		t.Fatal(err)
	}
	var dry DryRunResponse
	if err := json.Unmarshal(res.Body, &dry); err != nil {
		t.Fatal(err)
	}
	if want := []httpx.Refusal{{Field: "rows[1]", SectionCode: "H5.1", Sex: "F", Message: msg}}; res.Status != 200 ||
		!slices.Equal(dry.Refusals, want) {
		t.Errorf("dry run = %d %s", res.Status, res.Body)
	}
}

// A dry run of a valid body runs the writes, answers the real warnings in a
// 200 and leaves no row behind.
func TestCreateFeedDailyDryRun(t *testing.T) {
	f := newFeedFixture(t)
	before := [2]int{f.count("feed_allocation"), f.count("entry")}
	key := uuid.New()
	req := f.request(nov18, &key, f.row("H5.2", dbq.SexM, f.cocosi, "173"))
	req.DryRun = true
	res, err := f.save(f.ion, req)
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"effects":[],"warnings":[{"code":"FEED_GUIDE","sectionCode":"H5.2","sex":"M",` +
		`"message":"H5.2 ♂: 173 g, cu 11% peste ghidul de 156 g. Verifică gramajul; se poate salva."}],"refusals":[]}`
	if res.Status != http.StatusOK || strings.TrimSpace(string(res.Body)) != want {
		t.Errorf("dry run = %d %s", res.Status, res.Body)
	}
	if after := [2]int{f.count("feed_allocation"), f.count("entry")}; after != before ||
		f.count("audit_log")+f.count("live_event")+f.count("request_replay") != 0 {
		t.Errorf("a dry run wrote: %v → %v", before, after)
	}
	// The key wasn't taken: the real save with it goes through.
	req.DryRun = false
	if res, err := f.save(f.ion, req); err != nil || res.Status != http.StatusCreated {
		t.Fatalf("save after the dry run: %d %v", res.Status, err)
	}
}

// T4 criterion 4: two saves of the same (H5.1, ♀, 18.11) under different
// keys, started together and held at the farm lock: exactly one 201, one
// 409 ALREADY_ENTERED naming the winner's entry, no 500.
func TestCreateFeedDailyConcurrent(t *testing.T) {
	ctx := context.Background()
	f := newFeedFixture(t)
	var wg sync.WaitGroup
	defer wg.Wait()
	holder, err := f.d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(ctx) }()
	if err := dbq.New(holder).LockFarm(ctx, f.farmA); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	outs := make([]outcome, 2)
	for i, p := range []auth.Principal{f.ion, f.maria} {
		wg.Go(func() {
			<-start
			key := uuid.New()
			res, err := f.svc.CreateFeedDaily(ctx, p, f.request(nov18, &key, f.row("H5.1", dbq.SexF, f.breeder2, "160")))
			outs[i] = outcome{res, err}
		})
	}
	close(start)
	f.waitFor(2)
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	var winner FeedDailyResponse
	var conflict error
	for _, o := range outs {
		switch {
		case o.err == nil && o.res.Status == http.StatusCreated:
			if err := json.Unmarshal(o.res.Body, &winner); err != nil {
				t.Fatal(err)
			}
		case o.err != nil:
			conflict = o.err
		default:
			t.Fatalf("unexpected outcome %d %s", o.res.Status, o.res.Body)
		}
	}
	if winner.Entry.ID == uuid.Nil || conflict == nil {
		t.Fatalf("outcomes = %+v", outs)
	}
	status, body := errBody(t, conflict)
	if status != http.StatusConflict || body.Code != "ALREADY_ENTERED" || body.EntryID == nil || *body.EntryID != winner.Entry.ID {
		t.Errorf("loser = %d %s %v, want 409 ALREADY_ENTERED %s", status, body.Code, body.EntryID, winner.Entry.ID)
	}
	if n := f.count("feed_allocation"); n != 1 {
		t.Errorf("allocations = %d, want 1", n)
	}
}

// feedHolder is the safety net after a feed_allocation_uq violation (M3 D2):
// it names the first row of the request whose key is held, and nothing when
// none is.
func TestFeedHolder(t *testing.T) {
	f := newFeedFixture(t)
	q := f.r.store.Q()
	req := f.request(nov18, nil, f.row("H5.2", dbq.SexF, f.breeder2, "150"), f.row("H5.1", dbq.SexF, f.breeder2, "150"))
	if err := feedHolder(context.Background(), q, req); err != nil {
		t.Fatalf("no holder: %v", err)
	}
	first, err := f.save(f.ion, f.request(nov18, nil, f.row("H5.1", dbq.SexF, f.breeder2, "160")))
	if err != nil {
		t.Fatal(err)
	}
	var resp FeedDailyResponse
	if err := json.Unmarshal(first.Body, &resp); err != nil {
		t.Fatal(err)
	}
	status, body := errBody(t, feedHolder(context.Background(), q, req))
	if status != http.StatusConflict || body.EntryID == nil || *body.EntryID != resp.Entry.ID ||
		!strings.HasPrefix(body.Message, "H5.1 ♀ are deja hrana introdusă pe 18.11.2026") {
		t.Errorf("holder = %d %+v", status, body)
	}
}

// statementLog is a pgx QueryTracer that records the "-- name: X" of every
// statement ("?" without one: begin, commit).
type statementLog struct {
	mu    sync.Mutex
	names []string
}

func (s *statementLog) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := "?"
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
	}
	s.mu.Lock()
	s.names = append(s.names, name)
	s.mu.Unlock()
	return ctx
}

func (s *statementLog) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (s *statementLog) take() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.names
	s.names = nil
	return out
}

// T4 criterion 5: one save with a key is at most 12 statements besides the
// per-row inserts, begin and commit included.
func TestCreateFeedDailyStatements(t *testing.T) {
	f := newFeedFixture(t)
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
	log.take()

	svc := NewService(db.NewStore(pool), f.d.Clock)
	key := uuid.New()
	if _, err := svc.CreateFeedDaily(context.Background(), f.ion, f.request(nov18, &key, f.allRows()...)); err != nil {
		t.Fatal(err)
	}
	got := log.take()
	others := slices.DeleteFunc(slices.Clone(got), func(n string) bool { return n == "EntriesInsertFeedAllocation" })
	if inserts := len(got) - len(others); inserts != 5 {
		t.Errorf("row inserts = %d, want 5", inserts)
	}
	want := []string{
		"GetReplay", "?", "LockFarm", "EntryIDByClientRequest", "EntriesFeedSaveContext", "EntriesFeedDaySlots",
		"InsertEntry", "CalcStockSeries", "InsertAuditLog", "EmitLiveEvent", "InsertReplay", "?",
	}
	if !slices.Equal(others, want) {
		t.Errorf("statements %q,\nwant %q", others, want)
	}
	if len(others) > 12 {
		t.Errorf("%d statements besides the row inserts, want at most 12", len(others))
	}
}
