package live

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

var farm = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func event() Event {
	return EntryCreated(EntryPayload{
		EntryID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), Kind: dbq.EntryKindFeedDaily,
		FarmIDs: []uuid.UUID{farm}, Dates: []timex.Day{testdb.Today}, RecordTypes: []string{"hrana"},
		At: timex.At(testdb.Now()), Summary: "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg",
	})
}

// listener is a separate connection that LISTENs on live_event.
func listener(t *testing.T, d *testdb.DB) *pgx.Conn {
	t.Helper()
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testdb.DSN(d.Name))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	if _, err := conn.Exec(ctx, "listen live_event"); err != nil {
		t.Fatal(err)
	}
	return conn
}

// notification waits up to wait for a live_event notification; nil when none came.
func notification(t *testing.T, conn *pgx.Conn, wait time.Duration) *string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	n, err := conn.WaitForNotification(ctx)
	if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	if n.Channel != "live_event" {
		t.Fatalf("channel = %q", n.Channel)
	}
	return &n.Payload
}

func liveEvents(t *testing.T, d *testdb.DB) int {
	t.Helper()
	var n int
	if err := d.App.QueryRow(context.Background(), `select count(*) from live_event`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestEmitCommit: an event written by il_app is in live_event after the
// commit, and its notification (payload = the row id) reaches a LISTEN
// connection only then.
func TestEmitCommit(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	conn := listener(t, d)

	tx, err := d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := Emit(ctx, dbq.New(tx), event()); err != nil {
		t.Fatal(err)
	}
	if p := notification(t, conn, 300*time.Millisecond); p != nil {
		t.Fatalf("notification %q before the commit", *p)
	}
	if n := liveEvents(t, d); n != 0 {
		t.Fatalf("%d live_event rows visible before the commit", n)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	p := notification(t, conn, 5*time.Second)
	if p == nil {
		t.Fatal("no notification after the commit")
	}

	var (
		id      int64
		farmIDs []uuid.UUID
		typ     string
		payload []byte
	)
	err = d.App.QueryRow(ctx, `select id, farm_ids, type, payload from live_event`).Scan(&id, &farmIDs, &typ, &payload)
	if err != nil {
		t.Fatal(err)
	}
	if *p != strconv.FormatInt(id, 10) {
		t.Errorf("notification payload = %q, want the row id %d", *p, id)
	}
	if typ != TypeEntryCreated || len(farmIDs) != 1 || farmIDs[0] != farm {
		t.Errorf("row = %s %v", typ, farmIDs)
	}
	var got map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"entryId": "22222222-2222-2222-2222-222222222222", "kind": "feed_daily",
		"farmIds": []any{farm.String()}, "dates": []any{"2026-11-18"}, "sections": []any{},
		"recordTypes": []any{"hrana"}, "author": nil, "at": "2026-11-18T12:00:00+02:00",
		"summary": "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg", "corrected": false,
	}
	gb, _ := json.Marshal(got)
	wb, _ := json.Marshal(want)
	if string(gb) != string(wb) {
		t.Errorf("payload = %s\nwant      %s", gb, wb)
	}
}

// TestEmitRollback: a rolled-back event leaves no row and no notification.
func TestEmitRollback(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	conn := listener(t, d)

	tx, err := d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := Emit(ctx, dbq.New(tx), event(), event()); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if p := notification(t, conn, 500*time.Millisecond); p != nil {
		t.Fatalf("notification %q after a rollback", *p)
	}
	if n := liveEvents(t, d); n != 0 {
		t.Fatalf("%d live_event rows after a rollback", n)
	}
}

// TestEntryCreatedNilLists: nil lists are sent as [] and the event's farms
// are the payload's.
func TestEntryCreatedNilLists(t *testing.T) {
	ev := EntryCreated(EntryPayload{Kind: dbq.EntryKindFeedDaily})
	b, err := json.Marshal(ev.Payload)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"farmIds", "dates", "sections", "recordTypes"} {
		if string(got[k]) != "[]" {
			t.Errorf("%s = %s, want []", k, got[k])
		}
	}
	if ev.Type != "entry.created" || ev.FarmIDs == nil {
		t.Errorf("event = %s %v", ev.Type, ev.FarmIDs)
	}
}

// priceKeys are the key fragments a live payload must never hold (M4 D2):
// the outbox has no role filter before M11.
var priceKeys = []string{"lei", "price", "total", "cost", "pricedby"}

// noPriceKey fails when any key of the JSON value b (at any depth) holds a
// price fragment.
func noPriceKey(t *testing.T, b []byte) {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatal(err)
	}
	for _, path := range priceKeyPaths("$", v) {
		t.Errorf("%s: price key in a live payload %s", path, b)
	}
}

// priceKeyPaths lists the paths of the keys in v that hold a price fragment.
func priceKeyPaths(path string, v any) []string {
	var out []string
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			if slices.ContainsFunc(priceKeys, func(frag string) bool { return strings.Contains(strings.ToLower(k), frag) }) {
				out = append(out, path+"."+k)
			}
			out = append(out, priceKeyPaths(path+"."+k, e)...)
		}
	case []any:
		for i, e := range x {
			out = append(out, priceKeyPaths(path+"["+strconv.Itoa(i)+"]", e)...)
		}
	}
	return out
}

// TestM4Payloads: delivery.created and price.set / price.changed carry the
// §9 fields only, numbers as JSON numbers, no lei.
func TestM4Payloads(t *testing.T) {
	entry := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	d1 := uuid.MustParse("33333333-3333-3333-3333-333333333331")
	d2 := uuid.MustParse("33333333-3333-3333-3333-333333333332")
	other := uuid.MustParse("11111111-1111-1111-1111-111111111112")
	for _, tc := range []struct {
		name  string
		ev    Event
		typ   string
		farms []uuid.UUID
		want  string
	}{
		{"delivery.created", DeliveryCreated(DeliveryPayload{
			EntryID: entry, FarmID: farm, FeedType: "Breeder 2", Kg: num.MustParse("9000.5"), Date: timex.MustDay(2026, 11, 12),
		}), TypeDeliveryCreated, []uuid.UUID{farm},
			`{"entryId":"22222222-2222-2222-2222-222222222222","farmId":"11111111-1111-1111-1111-111111111111","feedType":"Breeder 2","kg":9000.5,"date":"2026-11-12"}`},
		{"price.set", PriceSet([]uuid.UUID{farm, other}, []uuid.UUID{d1, d2}, decimal.RequireFromString("18160.0")),
			TypePriceSet, []uuid.UUID{farm, other},
			`{"deliveryIds":["33333333-3333-3333-3333-333333333331","33333333-3333-3333-3333-333333333332"],"count":2,"kg":18160}`},
		{"price.changed", PriceChanged([]uuid.UUID{farm}, []uuid.UUID{d1}, decimal.RequireFromString("25640")),
			TypePriceChanged, []uuid.UUID{farm},
			`{"deliveryIds":["33333333-3333-3333-3333-333333333331"],"count":1,"kg":25640}`},
		{"price.changed, nil lists", PriceChanged(nil, nil, decimal.Zero), TypePriceChanged, []uuid.UUID{},
			`{"deliveryIds":[],"count":0,"kg":0}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.ev.Payload)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tc.want {
				t.Errorf("payload = %s\nwant      %s", b, tc.want)
			}
			if tc.ev.Type != tc.typ || tc.ev.FarmIDs == nil || !slices.Equal(tc.ev.FarmIDs, tc.farms) {
				t.Errorf("event = %s %v, want %s %v", tc.ev.Type, tc.ev.FarmIDs, tc.typ, tc.farms)
			}
			noPriceKey(t, b)
		})
	}
}

// TestPriceKeyPathsCatchesLei: the walker itself finds price keys at depth.
func TestPriceKeyPathsCatchesLei(t *testing.T) {
	var v any
	if err := json.Unmarshal([]byte(`{"deliveryIds":[],"rows":[{"kg":1,"totalLei":1}],"pricePerKg":2}`), &v); err != nil {
		t.Fatal(err)
	}
	got := priceKeyPaths("$", v)
	slices.Sort(got)
	if want := []string{"$.pricePerKg", "$.rows[0].totalLei"}; !slices.Equal(got, want) {
		t.Errorf("paths = %v, want %v", got, want)
	}
}

// TestEmitM4Events: the M4 events go through the outbox like entry.created.
func TestEmitM4Events(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	evs := []Event{
		DeliveryCreated(DeliveryPayload{EntryID: uuid.New(), FarmID: farm, FeedType: "Cocoși", Kg: num.FromInt(1000), Date: testdb.Today}),
		PriceSet([]uuid.UUID{farm}, []uuid.UUID{uuid.New()}, decimal.NewFromInt(1000)),
		PriceChanged([]uuid.UUID{farm}, []uuid.UUID{uuid.New()}, decimal.NewFromInt(1000)),
	}
	if err := Emit(ctx, dbq.New(d.App), evs...); err != nil {
		t.Fatal(err)
	}
	rows, err := d.App.Query(ctx, `select type, payload::text from live_event order by id`)
	if err != nil {
		t.Fatal(err)
	}
	var types []string
	for rows.Next() {
		var typ, payload string
		if err := rows.Scan(&typ, &payload); err != nil {
			t.Fatal(err)
		}
		types = append(types, typ)
		noPriceKey(t, []byte(payload))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if want := []string{"delivery.created", "price.set", "price.changed"}; !slices.Equal(types, want) {
		t.Errorf("types = %v, want %v", types, want)
	}
}
