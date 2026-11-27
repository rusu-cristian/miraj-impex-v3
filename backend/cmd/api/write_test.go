package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The M3 write path over HTTP (plan T8): the Hrană zilnică save through the
// whole server (socket, middleware, session, handler, save runner, SQL), its
// contract (decoder, body limit, Origin, null lists, Cache-Control) and its
// concurrency, then the reads that show its effect (form state, Rapoarte ›
// Hrană, Rezumat, Consum). Every response passes stack.do's checks:
// Cache-Control: no-store on every /api answer (201, a 200 replay, every
// error) and no null list.

const feedDailyPath = "/api/entries/feed-daily"

// feedRow is one row of a Hrană zilnică save body.
type feedRow struct {
	SectionID  uuid.UUID   `json:"sectionId"`
	Sex        string      `json:"sex"`
	FeedTypeID uuid.UUID   `json:"feedTypeId"`
	G          json.Number `json:"g"` // a JSON number (decimal.Decimal would encode a string)
}

// feedBody is a save body: dryRun and the key only when set.
func feedBody(t testing.TB, farm uuid.UUID, rows []feedRow, key *uuid.UUID, dryRun bool) string {
	t.Helper()
	b := map[string]any{"farmId": farm, "date": testdb.Today, "rows": rows}
	if key != nil {
		b["clientRequestId"] = key
	}
	if dryRun {
		b["dryRun"] = true
	}
	out, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

// deleteFeedDay soft-deletes every live Hrană zilnică entry of farm on
// 18.11.2026 with its rows, as the owner, so the day can be saved again.
func deleteFeedDay(t testing.TB, s *stack, farm uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	var owner uuid.UUID
	if err := s.d.App.QueryRow(ctx, `select id from app_user where login = 'proprietar'`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	s.exec(t, `update feed_allocation set deleted_at = $3 where farm_id = $1 and date = $2 and deleted_at is null`,
		farm, testdb.Today, testdb.Now())
	s.exec(t, `update entry set deleted_at = $3, deleted_by = $4
		where farm_id = $1 and kind = 'feed_daily' and date = $2 and deleted_at is null`, farm, testdb.Today, testdb.Now(), owner)
}

// liveFeedEntries counts farm's live Hrană zilnică entries on 18.11.2026.
func liveFeedEntries(t testing.TB, s *stack, farm uuid.UUID) int {
	t.Helper()
	var n int
	if err := s.d.App.QueryRow(context.Background(), `select count(*) from entry
		where farm_id = $1 and kind = 'feed_daily' and date = $2 and deleted_at is null`, farm, testdb.Today).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// seededFeedRows are the live Hrană zilnică rows of farm on 18.11.2026 in
// tree order (Hîrbovăț on the seed: the 8 rows of §7.8's example).
func seededFeedRows(t testing.TB, s *stack, farm uuid.UUID) []feedRow {
	t.Helper()
	rs, err := s.d.App.Query(context.Background(), `select fa.section_id, fa.sex::text, fa.feed_type_id, fa.g_per_bird::text
		from feed_allocation fa join section s on s.id = fa.section_id
		where fa.farm_id = $1 and fa.date = $2 and fa.deleted_at is null order by s.code, fa.sex`, farm, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	var rows []feedRow
	for rs.Next() {
		var r feedRow
		var g string
		if err := rs.Scan(&r.SectionID, &r.Sex, &r.FeedTypeID, &g); err != nil {
			t.Fatal(err)
		}
		r.G = json.Number(g)
		rows = append(rows, r)
	}
	if err := rs.Err(); err != nil || len(rows) != 8 {
		t.Fatalf("seeded 18.11 rows: %d, %v; want 8", len(rows), err)
	}
	return rows
}

// seededFeedDay is a save body of farm's seeded 18.11.2026 rows, without a
// key (each save is new).
func seededFeedDay(t testing.TB, s *stack, farm uuid.UUID) string {
	t.Helper()
	return feedBody(t, farm, seededFeedRows(t, s, farm), nil, false)
}

// formState is the part of GET /api/forms/feed-daily the chain reads.
type formState struct {
	Rows []struct {
		SectionID  uuid.UUID `json:"sectionId"`
		Code       string    `json:"code"`
		Sex        string    `json:"sex"`
		BirdsStart int       `json:"birdsStart"`
		Suggestion struct {
			FeedTypeID *uuid.UUID `json:"feedTypeId"`
		} `json:"suggestion"`
		Existing *struct {
			EntryID uuid.UUID `json:"entryId"`
		} `json:"existing"`
	} `json:"rows"`
	Existing []struct {
		EntryID     uuid.UUID `json:"entryId"`
		Author      *string   `json:"author"`
		Editable    bool      `json:"editable"`
		Correctable bool      `json:"correctable"`
	} `json:"existing"`
	Stock []stockItem `json:"stock"`
}

// stockItem is a stock row of the form state or of reports/feed.
type stockItem struct {
	Farm     string           `json:"farm"`
	Type     string           `json:"type"`
	Kg       decimal.Decimal  `json:"kg"`
	AvgDaily decimal.Decimal  `json:"avgDaily"`
	DaysLeft *decimal.Decimal `json:"daysLeft"`
	Window   []string         `json:"window"`
	RunOut   *string          `json:"runOut"`
	Lamp     string           `json:"lamp"`
	Reason   string           `json:"reason"`
	Sub      string           `json:"sub"`
}

// String is the row at the board's precision: "Breeder 2 31664.9 kg · 2392.2
// kg/zi · 13.2 zile · 2026-12-01 · in_norma".
func (s stockItem) String() string {
	if s.Type == "" {
		return strings.Join([]string{s.Farm, s.Lamp, s.Reason, s.Sub}, " · ")
	}
	days, runOut := "null", "null"
	if s.DaysLeft != nil {
		days = s.DaysLeft.StringFixed(1)
	}
	if s.RunOut != nil {
		runOut = *s.RunOut
	}
	return fmt.Sprintf("%s %s kg · %s kg/zi · %s zile · %s · %s %v", s.Type, s.Kg.StringFixed(1),
		s.AvgDaily.StringFixed(1), days, runOut, s.Lamp, s.Window)
}

func stockStrings(items []stockItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.String())
	}
	return out
}

func decodeBody[T any](t testing.TB, r response) T {
	t.Helper()
	var v T
	if err := json.Unmarshal([]byte(r.body), &v); err != nil {
		t.Fatalf("body %q: %v", r.body, err)
	}
	return v
}

// T8 criterion 3: on the dev seed, the manager of Hîrbovăț re-enters
// 18.11.2026 after the seeded Hrană zilnică entry is soft-deleted (plan
// Assumptions: ♀ 160 g everywhere, ♂ 144/143/142/141): form state → dry run
// → save → replay → a second save of the same rows → form state again →
// Rapoarte › Hrană → Rezumat → Consum. The figures are §7.8's example and
// goldens 26–28 (T4, T5, T6).
func TestFeedDailyChainSeeded(t *testing.T) {
	t.Parallel()
	s := newSeededStack(t)
	hirb := seededFarm(t, s, "hirbovat")
	manager := seededSession(t, s, "ion.rusu")
	deleteFeedDay(t, s, hirb)
	get := func(path string) response {
		t.Helper()
		r := s.do(t, call{method: http.MethodGet, path: path, cookie: manager.token})
		if r.status != http.StatusOK {
			t.Fatalf("GET %s: %s, want 200", path, r)
		}
		return r
	}
	post := func(body string) response {
		t.Helper()
		return s.do(t, call{method: http.MethodPost, path: feedDailyPath, body: body, cookie: manager.token})
	}
	formPath := "/api/forms/feed-daily?farmId=" + hirb.String() + "&date=2026-11-18"
	window := []string{"2026-11-16", "2026-11-18"}

	// 1. The form state without the day's entry: 8 rows, none saved, the
	// stock before the save ("din 33 968", "din 1 936") over 15–17.11.
	form := decodeBody[formState](t, get(formPath))
	if len(form.Rows) != 8 || len(form.Existing) != 0 {
		t.Fatalf("form: %d rows, existing %+v; want 8 rows, none", len(form.Rows), form.Existing)
	}
	if got := stockStrings(form.Stock); len(got) != 2 ||
		!strings.HasPrefix(got[0], "Breeder 2 33968.1 kg") || !strings.HasPrefix(got[1], "Cocoși 1935.7 kg") ||
		!slices.Equal(form.Stock[0].Window, []string{"2026-11-15", "2026-11-17"}) {
		t.Errorf("form stock before the save = %q", got)
	}
	gM := []string{"144", "143", "142", "141"} // ♂ rows in tree order
	var rows []feedRow
	for _, r := range form.Rows {
		if r.Existing != nil || r.Suggestion.FeedTypeID == nil {
			t.Fatalf("row %s %s: existing %v, suggested type %v; want none and the last used type", r.Code, r.Sex, r.Existing, r.Suggestion.FeedTypeID)
		}
		g := json.Number("160")
		if r.Sex == string(dbq.SexM) {
			g, gM = json.Number(gM[0]), gM[1:]
		}
		rows = append(rows, feedRow{SectionID: r.SectionID, Sex: r.Sex, FeedTypeID: *r.Suggestion.FeedTypeID, G: g})
	}

	// 2. A dry run: 200 with nothing to report, nothing written.
	key := uuid.New()
	dry := post(feedBody(t, hirb, rows, &key, true))
	wantBody(t, "dry run", dry, http.StatusOK, `{"effects":[],"warnings":[],"refusals":[]}`)
	if n := liveFeedEntries(t, s, hirb); n != 0 {
		t.Fatalf("after the dry run: %d live entries, want 0", n)
	}

	// 3. The save: §7.8's example response.
	body := feedBody(t, hirb, rows, &key, false)
	first := post(body)
	if first.status != http.StatusCreated {
		t.Fatalf("save: %s, want 201", first)
	}
	for _, want := range []string{
		`"kgTotal":2457.9`,
		`"byType":[{"type":"Breeder 2","kg":2303.2,"stockAfter":31664.9},{"type":"Cocoși","kg":154.7,"stockAfter":1781}]`,
		"\"toast\":\"Hrană zilnică salvată · 8 rânduri · 2\u202f457,9 kg · stoc Breeder 2: 31\u202f665 kg\"",
		`"warnings":[]`,
		`"asOf":"2026-11-18T12:00:00+02:00"`,
	} {
		if !strings.Contains(first.body, want) {
			t.Errorf("save body lacks %s:\n%s", want, first.body)
		}
	}
	saved := decodeBody[struct {
		Entry httpx.EntryRef `json:"entry"`
	}](t, first)
	if saved.Entry.Kind != string(dbq.EntryKindFeedDaily) {
		t.Errorf("entry %+v, want kind feed_daily", saved.Entry)
	}

	// 4. The same request again: 200, the same bytes, still one entry.
	again := post(body)
	if again.status != http.StatusOK || again.body != first.body {
		t.Errorf("replay: %s, want 200 with the first body %s", again, first.body)
	}
	// 5. The same rows under another key: 409 ALREADY_ENTERED naming the
	// first colliding row and the entry that holds it (Q3).
	dup := post(feedBody(t, hirb, rows, new(uuid.New()), false))
	wantBody(t, "second save of the day", dup, http.StatusConflict, `{"error":{"code":"ALREADY_ENTERED",`+
		`"message":"H5.1 ♀ are deja hrana introdusă pe 18.11.2026 la 12:00. Apasă Editează ca să schimbi valorile.",`+
		`"entryId":"`+saved.Entry.ID.String()+`"}}`)
	if n := liveFeedEntries(t, s, hirb); n != 1 {
		t.Errorf("live entries = %d, want 1", n)
	}

	// 6. The form state after the save: every row saved in that entry, which
	// its author may edit the same day; the stock now (Q5) over 16–18.11.
	form = decodeBody[formState](t, get(formPath))
	for _, r := range form.Rows {
		if r.Existing == nil || r.Existing.EntryID != saved.Entry.ID {
			t.Errorf("row %s %s: existing %v, want entry %s", r.Code, r.Sex, r.Existing, saved.Entry.ID)
		}
	}
	if len(form.Existing) != 1 || form.Existing[0].EntryID != saved.Entry.ID || form.Existing[0].Author == nil ||
		*form.Existing[0].Author != "Ion Rusu" || !form.Existing[0].Editable || !form.Existing[0].Correctable {
		t.Errorf("existing = %+v, want the saved entry by Ion Rusu, editable and correctable", form.Existing)
	}
	wantStock := []string{
		fmt.Sprintf("Breeder 2 31664.9 kg · 2392.2 kg/zi · 13.2 zile · 2026-12-01 · in_norma %v", window),
		fmt.Sprintf("Cocoși 1781.0 kg · 170.9 kg/zi · 10.4 zile · 2026-11-28 · in_norma %v", window),
	}
	if got := stockStrings(form.Stock); !slices.Equal(got, wantStock) {
		t.Errorf("form stock after the save\n got %q\nwant %q", got, wantStock)
	}

	// 7. Rapoarte › Hrană, every farm of the manager: the same figures, then
	// Telița without an open cycle (S11).
	report := decodeBody[struct {
		Stock []stockItem `json:"stock"`
	}](t, get("/api/reports/feed"))
	wantReport := []string{
		wantStock[0], wantStock[1],
		"Telița · oprit · fără ciclu deschis · ultimul ciclu: Seria 2026, închis pe 28.02.2026",
	}
	if got := stockStrings(report.Stock); !slices.Equal(got, wantReport) {
		t.Errorf("reports/feed stock\n got %q\nwant %q", got, wantReport)
	}
	for _, it := range report.Stock[:2] {
		if it.Farm != "Hîrbovăț" {
			t.Errorf("reports/feed row %s: farm %q, want Hîrbovăț", it.Type, it.Farm)
		}
	}

	// 8. Rezumat: feedDaysLeft with the same days and run-out days (Q7).
	summary := decodeBody[struct {
		Stats []struct {
			ID    string `json:"id"`
			Items []struct {
				Type   string          `json:"type"`
				Days   decimal.Decimal `json:"days"`
				RunOut string          `json:"runOut"`
				Kg     decimal.Decimal `json:"kg"`
			} `json:"items"`
		} `json:"stats"`
	}](t, get("/api/farms/"+hirb.String()+"/summary"))
	var items []string
	for _, st := range summary.Stats {
		if st.ID == "feedDaysLeft" {
			for _, it := range st.Items {
				items = append(items, fmt.Sprintf("%s %s %s %s", it.Type, it.Days.StringFixed(1), it.RunOut, it.Kg.StringFixed(1)))
			}
		}
	}
	if want := []string{"Breeder 2 13.2 2026-12-01 31664.9", "Cocoși 10.4 2026-11-28 1781.0"}; !slices.Equal(items, want) {
		t.Errorf("summary feedDaysLeft = %q, want %q", items, want)
	}

	// 9. Consum: the 18.11 day row carries the saved kg (golden 26).
	table := decodeBody[struct {
		Weeks []struct {
			Days []struct {
				Date  string                      `json:"date"`
				Cells map[string]*decimal.Decimal `json:"cells"`
			} `json:"days"`
		} `json:"weeks"`
	}](t, get("/api/table?view=report-feed&farmId="+hirb.String()))
	var day string
	for _, w := range table.Weeks {
		for _, d := range w.Days {
			if d.Date == "2026-11-18" && d.Cells["feed.kg.F"] != nil && d.Cells["feed.kg.M"] != nil {
				day = d.Cells["feed.kg.F"].StringFixed(1) + " · " + d.Cells["feed.kg.M"].StringFixed(1)
			}
		}
	}
	if day != "2303.2 · 154.7" {
		t.Errorf("Consum 18.11 kg = %q, want 2303.2 · 154.7", day)
	}
}

// Two identical requests at once (one key, a retry racing the original)
// save once: one 201 and one 200 with the same bytes. Two saves of the same
// rows under different keys: one 201 and one 409 ALREADY_ENTERED naming the
// winner's entry. Never a 500 (§7.0 idempotency, M3 D2).
func TestFeedDailyConcurrentSeeded(t *testing.T) {
	t.Parallel()
	s := newSeededStack(t)
	hirb := seededFarm(t, s, "hirbovat")
	vet := seededSession(t, s, "dr.ceban")

	rows := seededFeedRows(t, s, hirb)

	race := func(bodies ...string) []response {
		t.Helper()
		out := make([]response, len(bodies))
		var start, done sync.WaitGroup
		start.Add(1)
		for i, b := range bodies {
			done.Go(func() {
				start.Wait()
				out[i] = s.do(t, call{method: http.MethodPost, path: feedDailyPath, body: b, cookie: vet.token})
			})
		}
		start.Done()
		done.Wait()
		slices.SortFunc(out, func(a, b response) int { return a.status - b.status })
		return out
	}

	deleteFeedDay(t, s, hirb)
	key := uuid.New()
	same := feedBody(t, hirb, rows, &key, false)
	got := race(same, same)
	if got[0].status != http.StatusOK || got[1].status != http.StatusCreated || got[0].body != got[1].body {
		t.Errorf("same key twice at once: %s / %s, want 200 and 201 with the same body", got[0], got[1])
	}
	if n := liveFeedEntries(t, s, hirb); n != 1 {
		t.Errorf("same key: %d live entries, want 1", n)
	}

	deleteFeedDay(t, s, hirb)
	got = race(feedBody(t, hirb, rows, new(uuid.New()), false), feedBody(t, hirb, rows, new(uuid.New()), false))
	if got[0].status != http.StatusCreated || got[1].status != http.StatusConflict {
		t.Fatalf("two keys at once: %s / %s, want 201 and 409", got[0], got[1])
	}
	winner := decodeBody[struct {
		Entry httpx.EntryRef `json:"entry"`
	}](t, got[0])
	loser := decodeBody[struct {
		Error struct {
			Code    string    `json:"code"`
			EntryID uuid.UUID `json:"entryId"`
		} `json:"error"`
	}](t, got[1])
	if loser.Error.Code != httpx.CodeAlreadyEntered || loser.Error.EntryID != winner.Entry.ID {
		t.Errorf("loser %s, want ALREADY_ENTERED naming the winner %s", got[1], winner.Entry.ID)
	}
	if n := liveFeedEntries(t, s, hirb); n != 1 {
		t.Errorf("two keys: %d live entries, want 1", n)
	}
}

// The HTTP contract of POST /api/entries/feed-daily (§7.0a, §11.4) on a
// hand-built farm: an unknown member deep in the body is named by its path,
// a body over 64 KiB is 413, and a cross-site request is refused before the
// handler (nothing written). A valid request from the app's origin passes
// all three and reaches the service (422: the section has no birds).
func TestFeedDailyContract(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	farm := s.farm(t, "Hîrbovăț", "hirbovat")
	token := s.session(t, s.user(t, "proprietar", "Proprietar", dbq.UserRoleOwner, true), sessionOpen)
	ctx := context.Background()
	var section, feedType uuid.UUID
	if err := s.d.App.QueryRow(ctx, `insert into section (farm_id, code) values ($1, 'H1') returning id`, farm).Scan(&section); err != nil {
		t.Fatal(err)
	}
	if err := s.d.App.QueryRow(ctx, `insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')
		returning id`).Scan(&feedType); err != nil {
		t.Fatal(err)
	}
	row := `{"sectionId":"` + section.String() + `","sex":"F","feedTypeId":"` + feedType.String() + `",`
	body := func(rowTail string) string {
		return `{"farmId":"` + farm.String() + `","date":"2026-11-18","rows":[` + row + rowTail + `}]}`
	}
	post := func(b string, header map[string]string) response {
		return s.do(t, call{method: http.MethodPost, path: feedDailyPath, body: b, cookie: token, header: header})
	}

	wantBody(t, "unknown member in a row", post(body(`"grams":160`), nil), http.StatusBadRequest,
		`{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",`+
			`"issues":[{"path":"rows[0].grams","code":"unknown_field","message":"Datele trimise nu sunt valide."}]}}`)
	wantBody(t, "65 KiB body", post(body(`"g":160,"pad":"`+strings.Repeat("a", 65<<10)+`"`), nil),
		http.StatusRequestEntityTooLarge, bodyTooLarge)
	for _, h := range []map[string]string{
		{"Origin": "https://evil.example"},
		{"Origin": "", "Sec-Fetch-Site": "cross-site"},
		{"Origin": ""},
		{"Content-Type": "text/plain"},
	} {
		wantBody(t, fmt.Sprintf("POST with %v", h), post(body(`"g":160`), h), http.StatusForbidden, bodyForbidden)
	}
	for _, h := range []map[string]string{nil, {"Origin": "", "Sec-Fetch-Site": "same-origin"}} {
		r := post(body(`"g":160`), h)
		if r.status != http.StatusUnprocessableEntity || errorCode(t, r) != httpx.CodeRefused {
			t.Errorf("POST with %v: %s, want 422 REFUSED (H3)", h, r)
		}
	}
	var n int
	if err := s.d.App.QueryRow(ctx, `select (select count(*) from entry) + (select count(*) from audit_log)
		+ (select count(*) from live_event)`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rows written: %d (%v), want 0", n, err)
	}
}
