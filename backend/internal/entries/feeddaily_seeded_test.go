package entries_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T4 criterion 3: on the dev seed, Hîrbovăț's 18.11.2026 Hrană zilnică entry
// is soft-deleted with its rows, and the same 8 rows are posted again
// through the handler (♀ 160 g everywhere, ♂ 144/143/142/141): §7.8's
// example response, goldens 26 and 27 again through the save. A repeat with
// the same clientRequestId answers 200 with the same bytes.
func TestCreateFeedDailySeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := seededFarm(t, d.App, "hirbovat")
	var owner auth.Principal
	if err := d.App.QueryRow(ctx, `select id, name from app_user where role = 'owner' order by login limit 1`).
		Scan(&owner.UserID, &owner.Name); err != nil {
		t.Fatal(err)
	}
	owner.Role = auth.Owner

	// The seeded 18.11 entry and its rows, in tree order.
	var old uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from entry where farm_id = $1 and kind = 'feed_daily' and date = $2
		and deleted_at is null`, farm, testdb.Today).Scan(&old); err != nil {
		t.Fatal(err)
	}
	rs, err := d.App.Query(ctx, `select fa.section_id, fa.sex::text, fa.feed_type_id, fa.g_per_bird::text
		from feed_allocation fa join section s on s.id = fa.section_id
		where fa.entry_id = $1 and fa.deleted_at is null order by s.code, fa.sex`, old)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := pgx.CollectRows(rs, func(r pgx.CollectableRow) (map[string]any, error) {
		var section, feedType uuid.UUID
		var sex, g string
		if err := r.Scan(&section, &sex, &feedType, &g); err != nil {
			return nil, err
		}
		return map[string]any{"sectionId": section, "sex": sex, "feedTypeId": feedType, "g": json.Number(g)}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 8 {
		t.Fatalf("seeded 18.11 rows = %d, want 8", len(rows))
	}
	if _, err := d.App.Exec(ctx, `update feed_allocation set deleted_at = $2 where entry_id = $1`, old, testdb.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := d.App.Exec(ctx, `update entry set deleted_at = $2, deleted_by = $3 where id = $1`,
		old, testdb.Now(), owner.UserID); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	entries.NewHandler(entries.NewService(db.NewStore(d.App), d.Clock)).Routes(mux, httpx.NewChain())
	key := uuid.New()
	payload, err := json.Marshal(map[string]any{"farmId": farm, "date": testdb.Today, "clientRequestId": key, "rows": rows})
	if err != nil {
		t.Fatal(err)
	}
	post := func() *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(auth.WithPrincipal(ctx, owner), http.MethodPost, "/api/entries/feed-daily",
			bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		return w
	}

	first := post()
	if first.Code != http.StatusCreated {
		t.Fatalf("POST = %d %s", first.Code, first.Body)
	}
	body := first.Body.String()
	for _, want := range []string{
		`"kgTotal":2457.9`,
		`"byType":[{"type":"Breeder 2","kg":2303.2,"stockAfter":31664.9},{"type":"Cocoși","kg":154.7,"stockAfter":1781}]`,
		"\"toast\":\"Hrană zilnică salvată · 8 rânduri · 2\u202f457,9 kg · stoc Breeder 2: 31\u202f665 kg\"",
		`"warnings":[]`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body lacks %s:\n%s", want, body)
		}
	}
	var resp entries.FeedDailyResponse
	if err := json.Unmarshal(first.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	var (
		n                   int
		action, summary     string
		recordTypes         []string
		events, eventsTyped int
	)
	if err := d.App.QueryRow(ctx, `select count(*), min(action::text), min(summary), min(record_types::text)::text[]
		from audit_log where entry_id = $1`, resp.Entry.ID).Scan(&n, &action, &summary, &recordTypes); err != nil {
		t.Fatal(err)
	}
	if n != 1 || action != "create" || summary != "Hrană zilnică · 8 din 8 rânduri · 2\u202f457,9 kg" ||
		len(recordTypes) != 1 || recordTypes[0] != "hrana" {
		t.Errorf("audit = %d %s %q %v", n, action, summary, recordTypes)
	}
	if err := d.App.QueryRow(ctx, `select count(*), count(*) filter (where type = 'entry.created')
		from live_event where payload->>'entryId' = $1::text`, resp.Entry.ID).Scan(&events, &eventsTyped); err != nil {
		t.Fatal(err)
	}
	if events != 1 || eventsTyped != 1 {
		t.Errorf("live events = %d (%d entry.created), want 1", events, eventsTyped)
	}

	again := post()
	if again.Code != http.StatusOK || !bytes.Equal(again.Body.Bytes(), first.Body.Bytes()) {
		t.Errorf("replay = %d %s, want 200 with the first body", again.Code, again.Body)
	}
	if err := d.App.QueryRow(ctx, `select count(*) from entry where farm_id = $1 and kind = 'feed_daily' and date = $2
		and deleted_at is null`, farm, testdb.Today).Scan(&n); err != nil || n != 1 {
		t.Errorf("live 18.11 feed entries = %d (%v), want 1", n, err)
	}
}

// M3 COR-1: a Hrană zilnică save on an old day whose rows the history import
// holds (an entry without author, source migration) is 409 ALREADY_ENTERED
// with the import's entry and the timeless sentence (the import instant is
// not a time the user knows), not "… introdusă pe … la hh:mm".
func TestCreateFeedDailyMigratedHolderSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := seededFarm(t, d.App, "hirbovat")
	owner := auth.Principal{UserID: uuid.New(), Name: "Test", Role: auth.Owner}

	// The latest migrated allocation of Hîrbovăț's open cycle (Seria 2026, from 22.02.2026).
	var (
		holder, section, feedType uuid.UUID
		day                       timex.Day
		sex                       dbq.Sex
	)
	if err := d.App.QueryRow(ctx, `select n.id, fa.section_id, fa.sex, fa.feed_type_id, fa.date
		from feed_allocation fa join entry n on n.id = fa.entry_id
		where fa.farm_id = $1 and n.source = 'migration' and n.created_by is null and fa.deleted_at is null
		  and fa.date >= '2026-02-22'
		order by fa.date desc, fa.section_id, fa.sex limit 1`, farm).Scan(&holder, &section, &sex, &feedType, &day); err != nil {
		t.Fatal(err)
	}
	g := num.MustParse("150")
	_, err := entries.NewService(db.NewStore(d.App), d.Clock).CreateFeedDaily(ctx, owner, entries.FeedDailyCreateRequest{
		FarmID: farm, Date: day,
		Rows: []entries.FeedDailyRow{{SectionID: section, Sex: sex, FeedTypeID: feedType, G: &g}},
	})
	var he httpx.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("save on %s: err = %v, want 409 ALREADY_ENTERED", day, err)
	}
	status, body := he.HTTPError()
	if status != http.StatusConflict || body.Code != httpx.CodeAlreadyEntered || body.EntryID == nil ||
		*body.EntryID != holder || body.Message != messages.EntryAlreadySaved() {
		t.Errorf("save on %s = %d %+v, want 409 ALREADY_ENTERED entryId %s %q",
			day, status, body, holder, messages.EntryAlreadySaved())
	}
}
