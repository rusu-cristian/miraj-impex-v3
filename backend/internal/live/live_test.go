package live

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
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
