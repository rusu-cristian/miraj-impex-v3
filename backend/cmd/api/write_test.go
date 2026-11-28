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
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
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

// ---------------------------------------------------------------------------
// M4: Livrare hrană and Prețuri over HTTP (plan T7)
// ---------------------------------------------------------------------------

const deliveryPath = "/api/entries/delivery"

// feedReport is the part of GET /api/reports/feed the M4 chain reads.
type feedReport struct {
	Stock      []stockItem `json:"stock"`
	Deliveries struct {
		Total    int `json:"total"`
		Unpriced struct {
			Count int             `json:"count"`
			Kg    decimal.Decimal `json:"kg"`
			Dates []string        `json:"dates"`
		} `json:"unpriced"`
		Rows []feedDelivery `json:"rows"`
	} `json:"deliveries"`
	Cost *struct {
		Cycles []costCycle `json:"cycles"`
		Total  *struct {
			Label      string           `json:"label"`
			Lei        decimal.Decimal  `json:"lei"`
			LeiPerBird *decimal.Decimal `json:"leiPerBird"`
			LeiPerEgg  *decimal.Decimal `json:"leiPerEgg"`
			ExcludedKg decimal.Decimal  `json:"excludedKg"`
		} `json:"total"`
	} `json:"cost"`
}

// feedDelivery is a Livrări row with its price (manager, owner).
type feedDelivery struct {
	ID         uuid.UUID        `json:"id"`
	Date       string           `json:"date"`
	Farm       string           `json:"farm"`
	Type       string           `json:"type"`
	Kg         decimal.Decimal  `json:"kg"`
	PricePerKg *decimal.Decimal `json:"pricePerKg"`
	TotalLei   *decimal.Decimal `json:"totalLei"`
	PricedBy   *string          `json:"pricedBy"`
	Tag        string           `json:"tag"`
	Actions    struct {
		SetPrice    bool `json:"setPrice"`
		ChangePrice bool `json:"changePrice"`
	} `json:"actions"`
}

// String is the row as "2026-11-12 Breeder 2 9000 · – · – · fără preț ·
// set=false change=false".
func (d feedDelivery) String() string {
	opt := func(v *decimal.Decimal) string {
		if v == nil {
			return "–"
		}
		return v.String()
	}
	by := "–"
	if d.PricedBy != nil {
		by = *d.PricedBy
	}
	return fmt.Sprintf("%s %s %s · %s · %s · %s · %s · set=%v change=%v", d.Date, d.Type, d.Kg, opt(d.PricePerKg), opt(d.TotalLei),
		by, d.Tag, d.Actions.SetPrice, d.Actions.ChangePrice)
}

// costCycle is a Cost row.
type costCycle struct {
	Label       string           `json:"label"`
	PricedCount int              `json:"pricedCount"`
	Lei         decimal.Decimal  `json:"lei"`
	LeiPerBird  *decimal.Decimal `json:"leiPerBird"`
	LeiPerEgg   *decimal.Decimal `json:"leiPerEgg"`
	Unpriced    *struct {
		Count int             `json:"count"`
		Kg    decimal.Decimal `json:"kg"`
		Span  string          `json:"span"`
	} `json:"unpriced"`
}

// costLines are the Cost rows and total at the board's precision (2
// decimals per bird and per egg): "Hîrbovăț · Seria 2026 · 65 · 9562437 ·
// 319.74 · 1.88 · fără preț 4 18160 17.02–12.11.2026".
func costLines(r feedReport) []string {
	unit := func(v *decimal.Decimal) string {
		if v == nil {
			return "–"
		}
		return v.StringFixed(2)
	}
	var out []string
	for _, c := range r.Cost.Cycles {
		line := fmt.Sprintf("%s · %d · %s · %s · %s", c.Label, c.PricedCount, c.Lei, unit(c.LeiPerBird), unit(c.LeiPerEgg))
		if c.Unpriced != nil {
			line += fmt.Sprintf(" · fără preț %d %s %s", c.Unpriced.Count, c.Unpriced.Kg, c.Unpriced.Span)
		}
		out = append(out, line)
	}
	if t := r.Cost.Total; t != nil {
		out = append(out, fmt.Sprintf("%s · %s · %s · %s · exclus %s", t.Label, t.Lei, unit(t.LeiPerBird), unit(t.LeiPerEgg), t.ExcludedKg))
	}
	return out
}

// golden31 is the Cost of Seria 2026 on the seed (SPEC §5.9, golden 31).
var golden31 = []string{
	"Telița · Seria 2026 · 13 · 2031768 · 64.32 · –",
	"Hîrbovăț · Seria 2026 · 65 · 9562437 · 319.74 · 1.88 · fără preț 4 18160 17.02–12.11.2026",
	"Total Seria 2026 · 11594205 · 367.02 · 2.29 · exclus 18160",
}

// T7 criterion 3: on the dev seed, through the whole server. Ion Rusu
// re-enters the 12.11.2026 delivery to Hîrbovăț (9 000 kg Breeder 2) after
// the seeded one is soft-deleted (plan Assumptions): form state → dry run →
// save → Rapoarte › Hrană (stock, the new row). Then the owner prices the
// four unpriced deliveries (S12 list → save), undoes the save, changes the
// 13.11 price with a Motiv (S11 Modifică), and reads the Cost after each
// step. Figures: T3 (golden 31), T4 (the price save), T5 (stockAfter
// 21 119,3, the Toast).
func TestDeliveryPriceChainSeeded(t *testing.T) {
	t.Parallel()
	s := newSeededStack(t)
	ctx := context.Background()
	hirb := seededFarm(t, s, "hirbovat")
	manager, owner := seededSession(t, s, "ion.rusu"), seededSession(t, s, "proprietar")
	do := func(u seededReader, method, path, body string, want int) response {
		t.Helper()
		r := s.do(t, call{method: method, path: path, body: body, cookie: u.token})
		if r.status != want {
			t.Fatalf("%s %s as %s: %s, want %d", method, path, u.login, r, want)
		}
		return r
	}
	report := func(u seededReader) feedReport {
		t.Helper()
		return decodeBody[feedReport](t, do(u, http.MethodGet, "/api/reports/feed", "", http.StatusOK))
	}
	var lastAudit, lastEvent int64
	if err := s.d.App.QueryRow(ctx, `select coalesce((select max(id) from audit_log), 0), coalesce((select max(id) from live_event), 0)`).
		Scan(&lastAudit, &lastEvent); err != nil {
		t.Fatal(err)
	}

	// The seeded 12.11 delivery, soft-deleted with its entry.
	var oldEntry uuid.UUID
	if err := s.d.App.QueryRow(ctx, `select fd.entry_id from feed_delivery fd join feed_type ft on ft.id = fd.feed_type_id
		where fd.farm_id = $1 and fd.date = '2026-11-12' and ft.name = 'Breeder 2' and fd.qty_kg = 9000 and fd.deleted_at is null`,
		hirb).Scan(&oldEntry); err != nil {
		t.Fatal(err)
	}
	s.exec(t, `update feed_delivery set deleted_at = $2 where entry_id = $1`, oldEntry, testdb.Now())
	s.exec(t, `update entry set deleted_at = $2, deleted_by = (select id from app_user where login = 'proprietar') where id = $1`,
		oldEntry, testdb.Now())

	// 1. The form state: the active types and the stock now, 9 000 kg short.
	form := decodeBody[struct {
		FeedTypes []struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
		} `json:"feedTypes"`
		Stock []stockItem `json:"stock"`
	}](t, do(manager, http.MethodGet, "/api/forms/delivery?farmId="+hirb.String(), "", http.StatusOK))
	var breeder2 uuid.UUID
	for _, ft := range form.FeedTypes {
		if ft.Name == "Breeder 2" {
			breeder2 = ft.ID
		}
	}
	if breeder2 == uuid.Nil || len(form.Stock) != 2 || form.Stock[0].Type != "Breeder 2" || form.Stock[0].Kg.StringFixed(1) != "22664.9" {
		t.Fatalf("form: Breeder 2 %s, stock %q; want the type and 31 664,9 − 9 000 = 22 664,9 kg", breeder2, stockStrings(form.Stock))
	}

	// 2. A dry run: 200, nothing written. 3. The save.
	body := func(extra string) string {
		return `{` + extra + `"farmId":"` + hirb.String() + `","date":"2026-11-12","feedTypeId":"` + breeder2.String() + `","qtyKg":9000}`
	}
	wantBody(t, "dry run", do(manager, http.MethodPost, deliveryPath, body(`"dryRun":true,`), http.StatusOK), http.StatusOK,
		`{"effects":[],"warnings":[],"refusals":[]}`)
	saved := decodeBody[struct {
		Entry      httpx.EntryRef  `json:"entry"`
		Toast      string          `json:"toast"`
		StockAfter decimal.Decimal `json:"stockAfter"`
		Warnings   []httpx.Warning `json:"warnings"`
	}](t, do(manager, http.MethodPost, deliveryPath, body(`"clientRequestId":"`+uuid.NewString()+`",`), http.StatusCreated))
	if saved.Entry.Kind != string(dbq.EntryKindDelivery) || saved.StockAfter.StringFixed(1) != "21119.3" || len(saved.Warnings) != 0 ||
		saved.Toast != "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg · fără preț" {
		t.Errorf("save = %+v; want a delivery entry, stockAfter 21119.3 and the S13 Toast", saved)
	}

	// 4. Rapoarte › Hrană (manager): the stock now is golden 27's again; the
	// new row is unpriced, the newest priced one is 13.11 "din registru";
	// the manager has no price action.
	rep := report(manager)
	if got := stockStrings(rep.Stock[:1]); !strings.HasPrefix(got[0], "Breeder 2 31664.9 kg") {
		t.Errorf("stock after the delivery = %q, want Breeder 2 31 664,9 kg", got)
	}
	if d := rep.Deliveries; d.Total != 82 || d.Unpriced.Count != 4 || d.Unpriced.Kg.String() != "18160" ||
		!slices.Equal(d.Unpriced.Dates, []string{"2026-02-17", "2026-03-14", "2026-11-12"}) {
		t.Errorf("deliveries = %d, unpriced %+v", d.Total, d.Unpriced)
	}
	var nov13 feedDelivery
	var rows []string
	for _, r := range rep.Deliveries.Rows[:2] {
		rows = append(rows, r.String())
	}
	if want := []string{
		"2026-11-13 Breeder 2 25640 · 7.7613 · 199000 · din registru ·  · set=false change=false",
		"2026-11-12 Breeder 2 9000 · – · – · – · fără preț · set=false change=false",
	}; !slices.Equal(rows, want) {
		t.Errorf("newest rows\n got %q\nwant %q", rows, want)
	} else {
		nov13 = rep.Deliveries.Rows[0]
	}
	newDelivery := rep.Deliveries.Rows[1].ID

	// 5. The S12 list (owner): the four, the new one last.
	list := decodeBody[struct {
		Items []struct {
			ID uuid.UUID `json:"id"`
		} `json:"items"`
		Count int `json:"count"`
	}](t, do(owner, http.MethodGet, "/api/deliveries?priced=false", "", http.StatusOK))
	if list.Count != 4 || len(list.Items) != 4 || list.Items[3].ID != newDelivery {
		t.Fatalf("S12 list = %+v, want 4 with the new delivery %s last", list, newDelivery)
	}

	// 6. The price save: 17.02 Breeder 2 by its total, the others per kg.
	set := decodeBody[struct {
		Count     int    `json:"count"`
		Toast     string `json:"toast"`
		UndoToken string `json:"undoToken"`
	}](t, do(owner, http.MethodPost, "/api/deliveries/prices", fmt.Sprintf(`{"items":[`+
		`{"id":"%s","totalLei":16000},{"id":"%s","pricePerKg":8.5},{"id":"%s","pricePerKg":8.25},{"id":"%s","pricePerKg":7.74}]}`,
		list.Items[0].ID, list.Items[1].ID, list.Items[2].ID, list.Items[3].ID), http.StatusOK))
	if set.Count != 4 || set.UndoToken == "" || set.Toast != "Prețuri salvate · 4 livrări · 18 160 kg intră în cost" {
		t.Errorf("price save = %+v", set)
	}
	// Every Hîrbovăț delivery of Seria 2026 now has a price: 9 562 437 +
	// 16 000 + 8 500 + 50 820 + 69 660 lei, nothing left out of the total.
	if got, want := costLines(report(owner)), []string{
		golden31[0],
		"Hîrbovăț · Seria 2026 · 69 · 9707417 · 324.59 · 1.91",
		"Total Seria 2026 · 11739185 · 371.61 · 2.31 · exclus 0",
	}; !slices.Equal(got, want) {
		t.Errorf("cost after the save\n got %q\nwant %q", got, want)
	}

	// 7. The undo: the four are "fără preț" again; the Cost is golden 31.
	undo := do(owner, http.MethodPost, "/api/deliveries/prices/undo", `{"undoToken":"`+set.UndoToken+`"}`, http.StatusOK)
	if !strings.Contains(undo.body, `"toast":"Prețuri anulate · 4 livrări înapoi fără preț"`) {
		t.Errorf("undo = %s", undo)
	}
	if got := costLines(report(owner)); !slices.Equal(got, golden31) {
		t.Errorf("cost after the undo\n got %q\nwant %q", got, golden31)
	}

	// 8. S11 Modifică: 13.11 Breeder 2 from 7,7613 to 7,74 lei/kg.
	change := do(owner, http.MethodPatch, "/api/deliveries/"+nov13.ID.String()+"/price",
		`{"pricePerKg":7.74,"reason":"Factura corectată"}`, http.StatusOK)
	for _, want := range []string{
		`"pricePerKg":7.74`, `"totalLei":198453.6`, `"pricedBy":"Proprietar"`,
		"\"toast\":\"Preț schimbat · Breeder 2 · 13.11.2026 · 7,74 lei/kg · 198 454 lei\"",
	} {
		if !strings.Contains(change.body, want) {
			t.Errorf("change body lacks %s: %s", want, change.body)
		}
	}
	// 9. The Cost: 9 562 437 − 199 000 + 198 453,6 lei at Hîrbovăț.
	rep = report(owner)
	if got, want := costLines(rep), []string{
		golden31[0],
		"Hîrbovăț · Seria 2026 · 65 · 9561890.6 · 319.72 · 1.88 · fără preț 4 18160 17.02–12.11.2026",
		"Total Seria 2026 · 11593658.6 · 367.00 · 2.29 · exclus 18160",
	}; !slices.Equal(got, want) {
		t.Errorf("cost after the change\n got %q\nwant %q", got, want)
	}
	if got := rep.Deliveries.Rows[0].String(); got != "2026-11-13 Breeder 2 25640 · 7.74 · 198453.6 · Proprietar ·  · set=false change=true" {
		t.Errorf("13.11 row after the change = %q", got)
	}

	// One audit row per write, each with its summary, and the live events
	// (none of them with a price).
	audits := queryStrings(t, s, `select action || ' · ' || summary from audit_log where id > $1 order by id`, lastAudit)
	if want := []string{
		"create · Livrare · 9 000 kg Breeder 2",
		"price_set · Prețuri · 4 livrări · 18 160 kg",
		"undo · Prețuri · 4 livrări · 18 160 kg",
		"price_change · Preț schimbat · Breeder 2 · 13.11.2026",
	}; !slices.Equal(audits, want) {
		t.Errorf("audit rows\n got %q\nwant %q", audits, want)
	}
	events := queryStrings(t, s, `select type from live_event where id > $1 order by id`, lastEvent)
	if want := []string{"entry.created", "delivery.created", "price.set", "price.changed", "price.changed"}; !slices.Equal(events, want) {
		t.Errorf("live events = %q, want %q", events, want)
	}
	if priced := queryStrings(t, s, `select type from live_event where id > $1 and payload::text ~ '(lei|price)'`, lastEvent); len(priced) != 0 {
		t.Errorf("live events with a price: %q", priced)
	}
}

// queryStrings runs a one-column query as il_app.
func queryStrings(t testing.TB, s *stack, sql string, args ...any) []string {
	t.Helper()
	rows, err := s.d.App.Query(context.Background(), sql, args...)
	if err != nil {
		t.Fatal(err)
	}
	out, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// The HTTP contract of the M4 writes (§7.0a, §11.4) on a hand-built farm, the
// owner calling: an unknown member is named by its path (a price write has
// no dryRun and no clientRequestId, M4 D3), a body over 64 KiB is 413, and a
// cross-site request is refused before the handler. Nothing is written.
func TestM4WritesContract(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	farm := s.farm(t, "Hîrbovăț", "hirbovat")
	token := s.session(t, s.user(t, "proprietar", "Proprietar", dbq.UserRoleOwner, true), sessionOpen)
	var feedType uuid.UUID
	if err := s.d.App.QueryRow(context.Background(), `insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')
		returning id`).Scan(&feedType); err != nil {
		t.Fatal(err)
	}
	id := unknownFarm.String() // no delivery has it
	pad := `"pad":"` + strings.Repeat("a", 65<<10) + `"`
	delivery := func(extra string) string {
		return `{"farmId":"` + farm.String() + `","date":"2026-11-18","feedTypeId":"` + feedType.String() + `","qtyKg":9000,` + extra + `}`
	}
	for _, tc := range []struct {
		method, path       string
		unknown, unknownAt string
		big                string
		valid              string
		validStatus        int
	}{
		{http.MethodPost, deliveryPath, delivery(`"pret":8`), "pret", delivery(pad), delivery(`"dryRun":true`), http.StatusOK},
		{http.MethodPost, "/api/deliveries/prices", `{"items":[{"id":"` + id + `","pricePerKg":8,"pret":1}]}`, "items[0].pret",
			`{"items":[],` + pad + `}`, `{"items":[{"id":"` + id + `","pricePerKg":8}]}`, http.StatusBadRequest},
		{http.MethodPost, "/api/deliveries/prices/undo", `{"undoToken":"x","dryRun":true}`, "dryRun",
			`{"undoToken":"x",` + pad + `}`, `{"undoToken":"x"}`, http.StatusBadRequest},
		{http.MethodPatch, "/api/deliveries/" + id + "/price", `{"pricePerKg":7.74,"reason":"x","clientRequestId":"` + uuid.NewString() + `"}`,
			"clientRequestId", `{"reason":"x",` + pad + `}`, `{"pricePerKg":7.74,"reason":"x"}`, http.StatusNotFound},
	} {
		what := tc.method + " " + tc.path
		send := func(b string, header map[string]string) response {
			return s.do(t, call{method: tc.method, path: tc.path, body: b, cookie: token, header: header})
		}
		wantBody(t, what+" unknown member", send(tc.unknown, nil), http.StatusBadRequest,
			`{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",`+
				`"issues":[{"path":"`+tc.unknownAt+`","code":"unknown_field","message":"Datele trimise nu sunt valide."}]}}`)
		wantBody(t, what+" 65 KiB", send(tc.big, nil), http.StatusRequestEntityTooLarge, bodyTooLarge)
		for _, h := range []map[string]string{
			{"Origin": "https://evil.example"},
			{"Origin": "", "Sec-Fetch-Site": "cross-site"},
			{"Content-Type": "text/plain"},
		} {
			wantBody(t, fmt.Sprintf("%s with %v", what, h), send(tc.valid, h), http.StatusForbidden, bodyForbidden)
		}
		if r := send(tc.valid, nil); r.status != tc.validStatus {
			t.Errorf("%s from the app's origin: %s, want %d", what, r, tc.validStatus)
		}
	}
	var n int
	if err := s.d.App.QueryRow(context.Background(), `select (select count(*) from entry) + (select count(*) from feed_delivery)
		+ (select count(*) from audit_log) + (select count(*) from live_event)`).Scan(&n); err != nil || n != 0 {
		t.Errorf("rows written: %d (%v), want 0", n, err)
	}
}

// ---------------------------------------------------------------------------
// M5: Zilnic, the entry read, Corectează, undo, delete and restore over HTTP
// (plan T9)
// ---------------------------------------------------------------------------

// entryAnswer is the part of GET /api/entries/{id} the chain reads; values
// stay raw (the create shape of the kind, sent back as PUT values).
type entryAnswer struct {
	Entry struct {
		ID        uuid.UUID              `json:"id"`
		Version   string                 `json:"version"`
		Author    *struct{ Name string } `json:"author"`
		Corrected *struct {
			By string `json:"by"`
			At string `json:"at"`
		} `json:"corrected"`
	} `json:"entry"`
	Values   json.RawMessage          `json:"values"`
	Sections []entries.EntrySection   `json:"sections"`
	Actions  entries.EntryReadActions `json:"actions"`
}

// writeAnswer is the part of an EntryWriteResponse (201 create, 200 change)
// the chain reads.
type writeAnswer struct {
	Entry struct {
		ID uuid.UUID `json:"id"`
	} `json:"entry"`
	Toast    string              `json:"toast"`
	Alive    *httpx.SexPair[int] `json:"alive"`
	Warnings []httpx.Warning     `json:"warnings"`
}

// seededZilnicDay is a Zilnic save body of farm's seeded 18.11.2026 bird
// values (Dr. Ceban's 08:15 mortality, 4 sections), without a key, and a
// func that soft-deletes the day's live Zilnic bird rows and the entries
// left without a live row, so the body can be saved again.
func seededZilnicDay(t testing.TB, s *stack, farm uuid.UUID) (body string, clear func()) {
	t.Helper()
	rs, err := s.d.App.Query(context.Background(), `select b.section_id, b.sex::text, -b.qty from bird_event b
		join section s on s.id = b.section_id
		where b.farm_id = $1 and b.date = $2 and b.type = 'mortality' and b.deleted_at is null order by s.code, b.sex`,
		farm, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	var sections []entries.DailySection
	for rs.Next() {
		var (
			id   uuid.UUID
			sex  string
			dead int64
		)
		if err := rs.Scan(&id, &sex, &dead); err != nil {
			t.Fatal(err)
		}
		if len(sections) == 0 || sections[len(sections)-1].SectionID != id {
			sections = append(sections, entries.DailySection{SectionID: id})
		}
		n := num.FromInt(dead)
		sec := &sections[len(sections)-1]
		if sex == string(dbq.SexF) {
			sec.F = &entries.DailySex{Mortality: &n}
		} else {
			sec.M = &entries.DailySex{Mortality: &n}
		}
	}
	if err := rs.Err(); err != nil || len(sections) != 4 {
		t.Fatalf("seeded 18.11 mortality: %d sections, %v; want 4", len(sections), err)
	}
	body = entryBody(t, map[string]any{"farmId": farm, "date": testdb.Today, "sections": sections})
	return body, func() {
		var owner uuid.UUID
		if err := s.d.App.QueryRow(context.Background(), `select id from app_user where login = 'proprietar'`).Scan(&owner); err != nil {
			t.Fatal(err)
		}
		s.exec(t, `update bird_event set deleted_at = $3 where farm_id = $1 and date = $2 and deleted_at is null
			and type in ('mortality', 'cull', 'sexing_error')`, farm, testdb.Today, testdb.Now())
		s.exec(t, `update entry e set deleted_at = $3, deleted_by = $4
			where e.farm_id = $1 and e.kind = 'daily' and e.date = $2 and e.deleted_at is null
			  and not exists (select 1 from bird_event b where b.entry_id = e.id and b.deleted_at is null)
			  and not exists (select 1 from egg_production g where g.entry_id = e.id and g.deleted_at is null)`,
			farm, testdb.Today, testdb.Now(), owner)
	}
}

// entryBody marshals a request body.
func entryBody(t testing.TB, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// T9 criterion 3: on the dev seed, through the whole server.
//
//  1. Ion Rusu's Zilnic form state of Hîrbovăț, 18.11 (T4's figures).
//  2. Dr. Ceban reads his 18.11 mortality save (his to edit today), deletes
//     it (the author, same day: no Motiv) and saves the same values again:
//     dry run, then 201 with §6.2's Toast and the farm ♀ 12 561 · ♂ 794
//     (golden 5, T5); the same key again replays the 201 byte for byte. He
//     undoes the new save ("Zilnic anulat · 18.11.2026") and saves it once
//     more, so the 17.11 figures below meet the seed's 18.11.
//  3. On the corrected 17.11 save (T6, golden 37): the owner reads it, puts
//     H5.2 ♀ back to 1; Ion Rusu's dry run 1 → 3 answers §7.8's preview, his
//     save the Toast "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281";
//     he undoes it ("Corecție anulată · H5.2 ♀ 3 → 1").
//  4. He deletes the save with a Motiv (dry run, then ♀ 12 563 → 12 568, the
//     S16 Dialog's figures, T7) and restores it (♀ 12 563).
func TestZilnicChainSeeded(t *testing.T) {
	t.Parallel()
	s := newSeededStack(t)
	hirb := seededFarm(t, s, "hirbovat")
	vet, manager, owner := seededSession(t, s, "dr.ceban"), seededSession(t, s, "ion.rusu"), seededSession(t, s, "proprietar")
	send := func(u seededReader, method, path, body string, want int) response {
		t.Helper()
		r := s.do(t, call{method: method, path: path, body: body, cookie: u.token})
		if r.status != want {
			t.Fatalf("%s %s as %s: %s, want %d", method, path, u.login, r, want)
		}
		return r
	}
	entryPath := func(id uuid.UUID) string { return "/api/entries/" + id.String() }

	// 1. The form state.
	form := decodeBody[entries.DailyFormResponse](t, send(manager, http.MethodGet,
		"/api/forms/daily?farmId="+hirb.String()+"&date=2026-11-18", "", http.StatusOK))
	var h52 uuid.UUID
	for _, h := range form.Halls {
		for _, sec := range h.Sections {
			if sec.Code == "H5.2" {
				h52 = sec.SectionID
				if sec.AliveStart != (httpx.SexPair[int]{F: 4283, M: 323}) {
					t.Errorf("form H5.2 aliveStart %+v, want ♀ 4 283 · ♂ 323", sec.AliveStart)
				}
			}
		}
	}
	if h52 == uuid.Nil || form.SAPT == nil || *form.SAPT != 59 || len(form.Existing) != 2 {
		t.Fatalf("form: H5.2 %s, sapt %v, existing %+v", h52, form.SAPT, form.Existing)
	}
	var mortality uuid.UUID // Dr. Ceban's 08:15 save, the one with bird values
	for _, e := range form.Existing {
		if e.Editable || !e.Correctable {
			t.Errorf("existing %+v as Ion Rusu: want correctable only", e)
		}
		if len(e.Sections) == 4 {
			mortality = e.EntryID
		}
	}

	// 2. Dr. Ceban's 18.11 save: read, delete, save again.
	mine := decodeBody[entryAnswer](t, send(vet, http.MethodGet, entryPath(mortality), "", http.StatusOK))
	if !mine.Actions.Edit || !mine.Actions.Delete || mine.Actions.Correct {
		t.Errorf("Dr. Ceban's own save of today: actions %+v, want edit and delete", mine.Actions)
	}
	var values entries.DailyValues
	if err := json.Unmarshal(mine.Values, &values); err != nil || len(values.Sections) != 4 {
		t.Fatalf("values %s: %v", mine.Values, err)
	}
	gone := decodeBody[writeAnswer](t, send(vet, http.MethodDelete, entryPath(mortality), `{}`, http.StatusOK))
	if gone.Alive == nil || *gone.Alive != (httpx.SexPair[int]{F: 12564, M: 794}) {
		t.Errorf("delete of the 18.11 save: %+v, want ♀ 12 564 · ♂ 794", gone)
	}
	send(vet, http.MethodGet, entryPath(mortality), "", http.StatusNotFound)

	create := func(key uuid.UUID, dryRun bool) string {
		return entryBody(t, map[string]any{"farmId": values.FarmID, "date": values.Date, "sections": values.Sections,
			"clientRequestId": key, "dryRun": dryRun})
	}
	saveAgain := func() uuid.UUID {
		t.Helper()
		key := uuid.New()
		dry := decodeBody[entries.DryRunResponse](t, send(vet, http.MethodPost, dailyPath, create(key, true), http.StatusOK))
		if len(dry.Refusals) != 0 {
			t.Errorf("dry run refusals %+v", dry.Refusals)
		}
		first := send(vet, http.MethodPost, dailyPath, create(key, false), http.StatusCreated)
		saved := decodeBody[writeAnswer](t, first)
		if saved.Toast != "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0" || saved.Alive == nil ||
			*saved.Alive != (httpx.SexPair[int]{F: 12561, M: 794}) {
			t.Errorf("save: %s", first)
		}
		if again := send(vet, http.MethodPost, dailyPath, create(key, false), http.StatusOK); again.body != first.body {
			t.Errorf("replay body %s, want the 201's %s", again.body, first.body)
		}
		return saved.Entry.ID
	}
	created := saveAgain()
	undone := decodeBody[writeAnswer](t, send(vet, http.MethodPost, entryPath(created)+"/undo", `{}`, http.StatusOK))
	if undone.Toast != "Zilnic anulat · 18.11.2026" {
		t.Errorf("undo of the create: %+v", undone)
	}
	send(vet, http.MethodGet, entryPath(created), "", http.StatusNotFound)
	saveAgain()

	// 3. The corrected 17.11 save.
	var corrected uuid.UUID
	if err := s.d.App.QueryRow(t.Context(), `select id from entry where kind = 'daily' and date = '2026-11-17' and corrected`).
		Scan(&corrected); err != nil {
		t.Fatal(err)
	}
	read := func(u seededReader) entryAnswer {
		t.Helper()
		return decodeBody[entryAnswer](t, send(u, http.MethodGet, entryPath(corrected), "", http.StatusOK))
	}
	e := read(owner)
	if e.Entry.Author == nil || e.Entry.Author.Name != "Dr. Ceban" || e.Entry.Corrected == nil ||
		e.Entry.Corrected.By != "Ion Rusu" || e.Entry.Corrected.At != "2026-11-18T10:12:00+02:00" ||
		e.Actions != (entries.EntryReadActions{Correct: true, Delete: true}) {
		t.Errorf("17.11 as the owner: %+v", e)
	}
	put := func(u seededReader, h52F int64, reason string, dryRun bool, want int) response {
		t.Helper()
		cur := read(u)
		var v entries.DailyValues
		if err := json.Unmarshal(cur.Values, &v); err != nil {
			t.Fatal(err)
		}
		for i, sec := range v.Sections {
			if sec.SectionID == h52 {
				f := *sec.F
				n := num.FromInt(h52F)
				f.Mortality = &n
				v.Sections[i].F = &f
			}
		}
		return send(u, http.MethodPut, entryPath(corrected), entryBody(t, map[string]any{
			"ifVersion": cur.Entry.Version, "values": v, "reason": reason, "dryRun": dryRun}), want)
	}
	put(owner, 1, "Înapoi la registru.", false, http.StatusOK)
	preview := decodeBody[entries.CorrectionPreview](t, put(manager, 3, "", true, http.StatusOK))
	if len(preview.Changes) != 1 || preview.Changes[0].Key != "H5.2 · găini moarte" ||
		!slices.Equal(preview.Effects, []string{"H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.",
			"Mortalitate · morți ♀ −5 → −7", "Hîrbovăț · ♀ azi 12 563 → 12 561"}) ||
		preview.CheckedDays != 2 || len(preview.Refusals) != 0 || len(preview.Warnings) != 1 {
		t.Errorf("correction preview %+v", preview)
	}
	saved := decodeBody[writeAnswer](t, put(manager, 3, "Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare.", false, http.StatusOK))
	if saved.Toast != "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281" {
		t.Errorf("correction: %+v", saved)
	}
	// Without a Motiv the correction is refused (400 on reason).
	if r := put(manager, 4, "", false, http.StatusBadRequest); !strings.Contains(r.body, `"path":"reason"`) {
		t.Errorf("correction without a Motiv: %s", r)
	}
	undone = decodeBody[writeAnswer](t, send(manager, http.MethodPost, entryPath(corrected)+"/undo", `{}`, http.StatusOK))
	if undone.Toast != "Corecție anulată · H5.2 ♀ 3 → 1" {
		t.Errorf("undo of the correction: %+v", undone)
	}

	// 4. Delete with a Motiv, then restore.
	dry := decodeBody[entries.CorrectionPreview](t, send(manager, http.MethodDelete, entryPath(corrected), `{"dryRun":true}`, http.StatusOK))
	if len(dry.Effects) != 1 || !strings.HasSuffix(dry.Effects[0],
		"4 secțiuni · ♀ −5 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 563 → ♀ 12 568. Ștergerea rămâne în Istoric modificări.") {
		t.Errorf("delete preview %+v", dry)
	}
	gone = decodeBody[writeAnswer](t, send(manager, http.MethodDelete, entryPath(corrected), `{"reason":"Salvare dublă."}`, http.StatusOK))
	if gone.Toast != "Mortalitate ștearsă · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 568" ||
		gone.Alive == nil || *gone.Alive != (httpx.SexPair[int]{F: 12568, M: 794}) {
		t.Errorf("delete: %+v", gone)
	}
	back := decodeBody[writeAnswer](t, send(manager, http.MethodPost, entryPath(corrected)+"/restore", `{}`, http.StatusOK))
	if back.Toast != "Mortalitate restaurată · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 563" {
		t.Errorf("restore: %+v", back)
	}
	if e := read(manager); e.Actions.Restore || !e.Actions.Correct {
		t.Errorf("17.11 after the restore as Ion Rusu: actions %+v", e.Actions)
	}
}

// The HTTP contract of the M5 writes (§7.0a, §11.4) on the matrix fixture,
// the owner calling: an unknown member is named by its path (inside PUT's
// values too; undo and restore take {} only), a body over 64 KiB is 413
// (DELETE's body included), and a cross-site request is refused before the
// handler. A valid request reaches the service: the dry runs answer 200,
// undo (Telița's delivery, no action) and restore (the live entry, whose
// latest action is a create) 409 STALE. Nothing is written.
func TestM5WritesContract(t *testing.T) {
	t.Parallel()
	s := newStack(t)
	f := matrixFarms{hirb: s.farm(t, "Hîrbovăț", "hirbovat"), tel: s.farm(t, "Telița", "telita")}
	if err := s.d.App.QueryRow(t.Context(), `insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder 2', 'F')
		returning id`).Scan(&f.feedType); err != nil {
		t.Fatal(err)
	}
	m5MatrixFixture(t, s, &f)
	token := s.session(t, s.user(t, "proprietar", "Proprietar", dbq.UserRoleOwner, true), sessionOpen)
	counts := func() string {
		t.Helper()
		var n string
		if err := s.d.App.QueryRow(t.Context(), `select concat_ws(' ', (select count(*) from entry), (select count(*) from bird_event),
			(select count(*) from bird_event where deleted_at is null), (select max(coalesce(updated_at, created_at)) from entry),
			(select count(*) from audit_log), (select count(*) from live_event))`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	before := counts()

	pad := `"pad":"` + strings.Repeat("a", 65<<10) + `"`
	entry := "/api/entries/" + f.entry.String()
	daily := func(extra string) string {
		return `{"farmId":"` + f.hirb.String() + `","date":"2026-11-18","sections":[{"sectionId":"` + f.birds.String() +
			`","F":{"mortality":1` + extra + `}}],"dryRun":true}`
	}
	correction := func(extra string) string {
		return `{"ifVersion":"` + f.entryVersion + `","values":{"farmId":"` + f.hirb.String() + `","date":"2026-11-17",` +
			`"sections":[{"sectionId":"` + f.birds.String() + `","F":{"mortality":2` + extra + `}}]},"reason":"Recount","dryRun":true}`
	}
	for _, tc := range []struct {
		method, path       string
		unknown, unknownAt string
		big                string
		valid              string
		validStatus        int
	}{
		{http.MethodPost, dailyPath, daily(`,"deaths":1`), "sections[0].F.deaths", daily(`},"x":{` + pad), daily(""), http.StatusOK},
		{http.MethodPut, entry, correction(`,"deaths":1`), "values.sections[0].F.deaths", correction(`},"x":{` + pad),
			correction(""), http.StatusOK},
		{http.MethodDelete, entry, `{"dryRun":true,"motiv":"x"}`, "motiv", `{"dryRun":true,` + pad + `}`, `{"dryRun":true}`, http.StatusOK},
		{http.MethodPost, "/api/entries/" + f.telEntry.String() + "/undo", `{"dryRun":true}`, "dryRun", `{` + pad + `}`, `{}`,
			http.StatusConflict},
		{http.MethodPost, entry + "/restore", `{"reason":"x"}`, "reason", `{` + pad + `}`, `{}`, http.StatusConflict},
	} {
		what := tc.method + " " + tc.path
		send := func(b string, header map[string]string) response {
			return s.do(t, call{method: tc.method, path: tc.path, body: b, cookie: token, header: header})
		}
		wantBody(t, what+" unknown member", send(tc.unknown, nil), http.StatusBadRequest,
			`{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",`+
				`"issues":[{"path":"`+tc.unknownAt+`","code":"unknown_field","message":"Datele trimise nu sunt valide."}]}}`)
		wantBody(t, what+" 65 KiB", send(tc.big, nil), http.StatusRequestEntityTooLarge, bodyTooLarge)
		for _, h := range []map[string]string{
			{"Origin": "https://evil.example"},
			{"Origin": "", "Sec-Fetch-Site": "cross-site"},
			{"Origin": ""},
			{"Content-Type": "text/plain"},
		} {
			wantBody(t, fmt.Sprintf("%s with %v", what, h), send(tc.valid, h), http.StatusForbidden, bodyForbidden)
		}
		for _, h := range []map[string]string{nil, {"Origin": "", "Sec-Fetch-Site": "same-origin"}} {
			if r := send(tc.valid, h); r.status != tc.validStatus {
				t.Errorf("%s with %v: %s, want %d", what, h, r, tc.validStatus)
			}
		}
	}
	// The entry read takes no query parameter and a well-formed id.
	get := func(path string) response {
		return s.do(t, call{method: http.MethodGet, path: path, cookie: token})
	}
	wantBody(t, "entry read with a parameter", get(entry+"?x=1"), http.StatusBadRequest, bodyUnknownParam("x"))
	if r := get("/api/entries/nu-e-un-id"); r.status != http.StatusBadRequest || !strings.Contains(r.body, `"path":"id"`) {
		t.Errorf("entry read of a malformed id: %s, want 400 on id", r)
	}
	if after := counts(); after != before {
		t.Errorf("rows written: %s → %s, want none", before, after)
	}
}
