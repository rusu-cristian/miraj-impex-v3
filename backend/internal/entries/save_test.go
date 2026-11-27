package entries

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

const testRoute = "POST /api/entries/test"

// saveFixture: two farms and two managers of both, on a fresh database.
type saveFixture struct {
	t          *testing.T
	d          *testdb.DB
	r          *Runner
	farmA      uuid.UUID
	farmB      uuid.UUID
	ion, maria auth.Principal
}

func newSaveFixture(t *testing.T) *saveFixture {
	t.Helper()
	f := &saveFixture{t: t, d: testdb.New(t)}
	f.r = NewRunner(db.NewStore(f.d.App), f.d.Clock)
	f.farmA = f.id(`insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying') returning id`)
	f.farmB = f.id(`insert into farm (name, name_key, type) values ('Telița', 'telita', 'growth') returning id`)
	farms := []uuid.UUID{f.farmA, f.farmB}
	f.ion = auth.Principal{
		UserID: f.id(`insert into app_user (name, login, password_hash, role) values ('Ion Rusu', 'ion', 'x', 'manager') returning id`),
		Name:   "Ion Rusu", Role: auth.Manager, FarmIDs: farms,
	}
	f.maria = auth.Principal{
		UserID: f.id(`insert into app_user (name, login, password_hash, role) values ('Maria', 'maria', 'x', 'manager') returning id`),
		Name:   "Maria", Role: auth.Manager, FarmIDs: farms,
	}
	return f
}

func (f *saveFixture) id(sql string, args ...any) uuid.UUID {
	f.t.Helper()
	var v uuid.UUID
	if err := f.d.App.QueryRow(context.Background(), sql, args...).Scan(&v); err != nil {
		f.t.Fatal(err)
	}
	return v
}

func (f *saveFixture) count(table string) int {
	f.t.Helper()
	var n int
	if err := f.d.App.QueryRow(context.Background(), "select count(*) from "+table).Scan(&n); err != nil {
		f.t.Fatal(err)
	}
	return n
}

// rows returns the row counts of entry, audit_log, live_event and request_replay.
func (f *saveFixture) rows() [4]int {
	return [4]int{f.count("entry"), f.count("audit_log"), f.count("live_event"), f.count("request_replay")}
}

func create(farm uuid.UUID, key *uuid.UUID, dryRun bool) Create {
	return Create{
		Route: testRoute, Kind: dbq.EntryKindFeedDaily, FarmIDs: []uuid.UUID{farm},
		Meta: httpx.WriteMeta{ClientRequestID: key, DryRun: dryRun},
	}
}

type testResponse struct {
	Entry httpx.EntryRef  `json:"entry"`
	Toast string          `json:"toast"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// testPlan inserts a feed_daily entry with no child rows for farm on today,
// runs between (if not nil) after the insert, and answers a dry run with an
// empty preview.
func testPlan(farm uuid.UUID, between func(ctx context.Context, tx *Tx) error) Plan {
	return func(ctx context.Context, tx *Tx) (Saved, error) {
		ins, err := tx.InsertEntry(ctx, Header{FarmID: farm, Date: tx.Today})
		if err != nil {
			return Saved{}, err
		}
		if between != nil {
			if err := between(ctx, tx); err != nil {
				return Saved{}, err
			}
		}
		if tx.DryRun {
			return Saved{}, tx.Preview(NewDryRun(nil, nil, nil))
		}
		summary := "Hrană zilnică · 0 din 0 rânduri · 0 kg"
		return Saved{
			Response: testResponse{Entry: ins.Ref(), Toast: "salvat · " + ins.ID.String(), AsOf: timex.At(tx.Now)},
			Audit: audit.Row{
				Action: dbq.AuditActionCreate, Entity: "entry", EntityID: ins.ID, EntryID: &ins.ID, EntryKind: &ins.Kind,
				RecordTypes: []string{"hrana"}, FarmIDs: []uuid.UUID{farm}, RecordDate: &tx.Today, Summary: &summary,
			},
			Events: []live.Event{live.EntryCreated(live.EntryPayload{
				EntryID: ins.ID, Kind: ins.Kind, FarmIDs: []uuid.UUID{farm}, Dates: []timex.Day{tx.Today},
				RecordTypes: []string{"hrana"}, Author: &tx.Principal.Name, At: timex.At(tx.Now), Summary: summary,
			})},
		}, nil
	}
}

func wantCode(t *testing.T, err error, status int, code string) httpx.APIError {
	t.Helper()
	var he httpx.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("err = %v, want %d %s", err, status, code)
	}
	st, body := he.HTTPError()
	if st != status || body.Code != code {
		t.Fatalf("err = %d %s %q, want %d %s", st, body.Code, body.Message, status, code)
	}
	return body
}

// TestCreateSaves: a save inserts its entry (created_at from the clock,
// created_by the caller, source manual), one audit row, one live event and
// the replay row holding the exact response bytes, and answers 201.
func TestCreateSaves(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	transcript := "pune 160 de grame"
	c := create(f.farmA, &key, false)
	c.Meta.Source, c.Meta.VoiceTranscript = "voice", &transcript // no voiceToken: recorded manual (M3 D4)

	res, err := f.r.Create(ctx, f.ion, c, testPlan(f.farmA, nil))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusCreated {
		t.Fatalf("status = %d", res.Status)
	}
	if got := f.rows(); got != [4]int{1, 1, 1, 1} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
	var resp testResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}

	var (
		createdAt time.Time
		createdBy uuid.UUID
		source    dbq.RecordSource
		crid      uuid.UUID
		version   time.Time
	)
	err = f.d.App.QueryRow(ctx, `select created_at, created_by, source, client_request_id,
		coalesce(updated_at, created_at) from entry where id = $1`, resp.Entry.ID).
		Scan(&createdAt, &createdBy, &source, &crid, &version)
	if err != nil {
		t.Fatal(err)
	}
	if !createdAt.Equal(testdb.Now()) || createdBy != f.ion.UserID || source != dbq.RecordSourceManual || crid != key {
		t.Errorf("entry = %v %v %v %v", createdAt, createdBy, source, crid)
	}
	if resp.Entry.Kind != "feed_daily" || resp.Entry.Version != "2026-11-18T10:00:00.000000Z" ||
		resp.Entry.Version != version.UTC().Format("2006-01-02T15:04:05.000000Z07:00") {
		t.Errorf("entry ref = %+v (row version %v)", resp.Entry, version)
	}

	var (
		at          time.Time
		userID      uuid.UUID
		auditSource dbq.RecordSource
		voice       *string
	)
	if err := f.d.App.QueryRow(ctx, `select at, user_id, source, voice_transcript from audit_log`).
		Scan(&at, &userID, &auditSource, &voice); err != nil {
		t.Fatal(err)
	}
	if !at.Equal(testdb.Now()) || userID != f.ion.UserID || auditSource != dbq.RecordSourceManual || voice != nil {
		t.Errorf("audit = %v %v %v %v", at, userID, auditSource, voice)
	}

	var (
		route  string
		status int16
		body   []byte
	)
	if err := f.d.App.QueryRow(ctx, `select route, status, body from request_replay where client_request_id = $1`, key).
		Scan(&route, &status, &body); err != nil {
		t.Fatal(err)
	}
	if route != testRoute || status != http.StatusCreated || !bytes.Equal(body, res.Body) {
		t.Errorf("replay = %s %d %q, want %q", route, status, body, res.Body)
	}
}

// TestCreateWithoutKey: no clientRequestId, no replay row; two saves are two entries.
func TestCreateWithoutKey(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	for range 2 {
		if res, err := f.r.Create(ctx, f.ion, create(f.farmA, nil, false), testPlan(f.farmA, nil)); err != nil || res.Status != http.StatusCreated {
			t.Fatalf("Create = %d, %v", res.Status, err)
		}
	}
	if got := f.rows(); got != [4]int{2, 2, 2, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateReplay: the same key twice is one entry; the second answer is
// 200 with the same bytes, and the plan doesn't run again.
func TestCreateReplay(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	first, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	if err != nil {
		t.Fatal(err)
	}
	f.d.Clock.Advance(time.Minute) // a replay is the stored bytes, not a new encoding
	ran := false
	second, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), func(context.Context, *Tx) (Saved, error) {
		ran = true
		return Saved{}, errors.New("plan ran on a replay")
	})
	if err != nil {
		t.Fatal(err)
	}
	if ran || first.Status != http.StatusCreated || second.Status != http.StatusOK || !bytes.Equal(first.Body, second.Body) {
		t.Fatalf("first %d %q, second %d %q (plan ran: %v)", first.Status, first.Body, second.Status, second.Body, ran)
	}
	if got := f.rows(); got != [4]int{1, 1, 1, 1} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateKeyStale: the same key from another user, or on another route,
// is 409 STALE and writes nothing.
func TestCreateKeyStale(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	if _, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil)); err != nil {
		t.Fatal(err)
	}
	_, err := f.r.Create(ctx, f.maria, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	body := wantCode(t, err, http.StatusConflict, httpx.CodeStale)
	if body.Message != messages.Stale() {
		t.Errorf("message = %q", body.Message)
	}
	other := create(f.farmA, &key, false)
	other.Route = "POST /api/entries/daily"
	_, err = f.r.Create(ctx, f.ion, other, testPlan(f.farmA, nil))
	wantCode(t, err, http.StatusConflict, httpx.CodeStale)
	if got := f.rows(); got != [4]int{1, 1, 1, 1} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateKeyPurged: a key whose replay row was purged (il_maint) is 409
// ALREADY_ENTERED with the entry that holds it.
func TestCreateKeyPurged(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	res, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	if err != nil {
		t.Fatal(err)
	}
	var first testResponse
	if err := json.Unmarshal(res.Body, &first); err != nil {
		t.Fatal(err)
	}
	if _, err := f.d.Pool(testdb.RoleMaint).Exec(ctx, `delete from request_replay`); err != nil {
		t.Fatal(err)
	}
	_, err = f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if body.EntryID == nil || *body.EntryID != first.Entry.ID || body.Message != messages.EntryAlreadySaved() {
		t.Errorf("409 = %+v, want entryId %s", body, first.Entry.ID)
	}
	if got := f.rows(); got != [4]int{1, 1, 1, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateKeyPurgedOtherUserOrRoute: after the purge, the key of Ion's
// feed_daily entry sent by Maria, or by Ion on a route of another kind, is
// 409 STALE with no entryId (M3 SEC-1), and writes nothing; Ion on his own
// route still gets ALREADY_ENTERED.
func TestCreateKeyPurgedOtherUserOrRoute(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	if _, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.d.Pool(testdb.RoleMaint).Exec(ctx, `delete from request_replay`); err != nil {
		t.Fatal(err)
	}
	other := create(f.farmA, &key, false)
	other.Route, other.Kind = "POST /api/entries/daily", dbq.EntryKindDaily
	for _, tc := range []struct {
		name string
		p    auth.Principal
		c    Create
	}{
		{"another user", f.maria, create(f.farmA, &key, false)},
		{"another user, another farm", f.maria, create(f.farmB, &key, false)},
		{"another route", f.ion, other},
	} {
		_, err := f.r.Create(ctx, tc.p, tc.c, testPlan(tc.c.FarmIDs[0], nil))
		body := wantCode(t, err, http.StatusConflict, httpx.CodeStale)
		if body.EntryID != nil || body.Message != messages.Stale() {
			t.Errorf("%s: 409 = %+v, want STALE without entryId", tc.name, body)
		}
	}
	_, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	if body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered); body.EntryID == nil {
		t.Errorf("own key: 409 = %+v, want entryId", body)
	}
	if got := f.rows(); got != [4]int{1, 1, 1, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateDryRun: a dry run runs the plan's writes, answers 200 with the
// preview and leaves no entry, audit, live or replay row; its key stays free
// for the real save.
func TestCreateDryRun(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	var keyInTx *uuid.UUID
	res, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, true), testPlan(f.farmA, func(ctx context.Context, tx *Tx) error {
		return tx.QueryRow(ctx, `select client_request_id from entry`).Scan(&keyInTx)
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusOK || string(res.Body) != `{"effects":[],"warnings":[],"refusals":[]}`+"\n" {
		t.Fatalf("dry run = %d %q", res.Status, res.Body)
	}
	if keyInTx != nil {
		t.Errorf("the preview's entry took the key %v", keyInTx)
	}
	if got := f.rows(); got != [4]int{0, 0, 0, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
	res, err = f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
	if err != nil || res.Status != http.StatusCreated {
		t.Fatalf("save after the dry run = %d, %v", res.Status, err)
	}
}

// TestCreateDryRunWithoutPreview: a plan that returns a save for a dry run
// is a programming error; nothing is written.
func TestCreateDryRunWithoutPreview(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	plan := func(ctx context.Context, tx *Tx) (Saved, error) {
		if _, err := tx.InsertEntry(ctx, Header{FarmID: f.farmA, Date: tx.Today}); err != nil {
			return Saved{}, err
		}
		return Saved{Response: struct{}{}}, nil
	}
	if _, err := f.r.Create(ctx, f.ion, create(f.farmA, nil, true), plan); err == nil {
		t.Fatal("no error")
	}
	if got := f.rows(); got != [4]int{0, 0, 0, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreatePlanError: a plan's refusal leaves nothing behind and is
// returned as it is.
func TestCreatePlanError(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	refused := httpx.RefusedWith([]httpx.Refusal{{Message: "nu"}}, nil)
	_, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, func(context.Context, *Tx) error {
		return refused
	}))
	if !errors.Is(err, refused) {
		t.Fatalf("err = %v", err)
	}
	if got := f.rows(); got != [4]int{0, 0, 0, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// waitFor polls the server until n backends of this database wait on a lock
// (advisory or a transaction lock), or fails after 5 s. It reads pg_locks:
// il_app can't see another session's wait_event in pg_stat_activity, but it
// sees every pid and database there.
func (f *saveFixture) waitFor(n int) {
	f.t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var got int
		err := f.d.App.QueryRow(context.Background(), `select count(*) from pg_locks l
			join pg_stat_activity a on a.pid = l.pid
			where not l.granted and a.datname = current_database()`).Scan(&got)
		if err != nil {
			f.t.Fatal(err)
		}
		if got >= n {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	f.t.Fatalf("fewer than %d backends wait on a lock after 5 s", n)
}

type outcome struct {
	res Result
	err error
}

// TestCreateConcurrentSameKey: two identical requests started together (a
// barrier, then both held at the farm lock) make one entry; one answers 201,
// the other 200 with the same bytes (the re-check under the lock).
func TestCreateConcurrentSameKey(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()

	var wg sync.WaitGroup
	defer wg.Wait()                   // runs last: the requests end once the holder is gone
	holder, err := f.d.App.Begin(ctx) // holds farm A's lock until both requests wait on it
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(ctx) }()
	if err := dbq.New(holder).LockFarm(ctx, f.farmA); err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	outs := make([]outcome, 2)
	for i := range outs {
		wg.Go(func() {
			<-start
			res, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, nil))
			outs[i] = outcome{res, err}
		})
	}
	close(start)
	f.waitFor(2)
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()

	checkOneWinner(t, outs)
	if got := f.rows(); got != [4]int{1, 1, 1, 1} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateClockAfterLock: the clock is read once the farm lock is held
// (M3 COR-3). A save that arrives at 23:59 on 18.11 and waits on the lock
// past midnight runs with Today 19.11, and its entry's created_at, its audit
// row and its asOf carry the instant after the wait.
func TestCreateClockAfterLock(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	clock := timex.FixedClock(time.Date(2026, time.November, 18, 23, 59, 0, 0, timex.Chisinau))
	r := NewRunner(db.NewStore(f.d.App), clock)
	later := time.Date(2026, time.November, 19, 0, 0, 30, 0, timex.Chisinau)

	holder, err := f.d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(ctx) }()
	if err := dbq.New(holder).LockFarm(ctx, f.farmA); err != nil {
		t.Fatal(err)
	}
	var (
		wg        sync.WaitGroup
		res       Result
		saveErr   error
		today     timex.Day
		planNow   time.Time
		recordNow = func(_ context.Context, tx *Tx) error { today, planNow = tx.Today, tx.Now; return nil }
	)
	wg.Go(func() {
		res, saveErr = r.Create(ctx, f.ion, create(f.farmA, nil, false), testPlan(f.farmA, recordNow))
	})
	f.waitFor(1)
	clock.Set(later)
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if saveErr != nil {
		t.Fatal(saveErr)
	}
	if want := timex.MustDay(2026, time.November, 19); today != want || !planNow.Equal(later) {
		t.Errorf("plan saw Today %s, Now %v; want %s, %v", today, planNow, want, later)
	}
	var resp testResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	var createdAt, auditAt time.Time
	var date timex.Day
	if err := f.d.App.QueryRow(ctx, `select created_at, date from entry where id = $1`, resp.Entry.ID).Scan(&createdAt, &date); err != nil {
		t.Fatal(err)
	}
	if err := f.d.App.QueryRow(ctx, `select at from audit_log where entry_id = $1`, resp.Entry.ID).Scan(&auditAt); err != nil {
		t.Fatal(err)
	}
	if !createdAt.Equal(later) || !auditAt.Equal(later) || date != today || resp.AsOf.String() != timex.At(later).String() {
		t.Errorf("created_at %v, audit at %v, date %s, asOf %v; want %v, %v, %s, %v",
			createdAt, auditAt, date, resp.AsOf, later, later, today, timex.At(later))
	}
}

// checkOneWinner: no error, statuses {201, 200}, byte-equal bodies.
func checkOneWinner(t *testing.T, outs []outcome) {
	t.Helper()
	for _, o := range outs {
		if o.err != nil {
			t.Fatal(o.err)
		}
	}
	statuses := []int{outs[0].res.Status, outs[1].res.Status}
	slices.Sort(statuses)
	if !slices.Equal(statuses, []int{http.StatusOK, http.StatusCreated}) {
		t.Fatalf("statuses = %v, want 200 and 201", statuses)
	}
	if !bytes.Equal(outs[0].res.Body, outs[1].res.Body) {
		t.Fatalf("bodies differ:\n%q\n%q", outs[0].res.Body, outs[1].res.Body)
	}
}

// TestCreateKeyConstraint: when the re-check can't see the other save (two
// farms, so no shared lock, and the first save not yet committed),
// entry_client_request_uq settles it after the transaction: one entry, 201
// and 200 with the same bytes.
func TestCreateKeyConstraint(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	key := uuid.New()
	inserted, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	outs := make([]outcome, 2)
	var wg sync.WaitGroup
	defer wg.Wait()
	defer releaseOnce.Do(func() { close(release) }) // also on a failure below
	wg.Go(func() {
		res, err := f.r.Create(ctx, f.ion, create(f.farmA, &key, false), testPlan(f.farmA, func(context.Context, *Tx) error {
			close(inserted)
			<-release
			return nil
		}))
		outs[0] = outcome{res, err}
	})
	<-inserted
	wg.Go(func() {
		res, err := f.r.Create(ctx, f.ion, create(f.farmB, &key, false), testPlan(f.farmB, nil))
		outs[1] = outcome{res, err}
	})
	f.waitFor(1) // the second insert waits on the first's uncommitted key
	releaseOnce.Do(func() { close(release) })
	wg.Wait()

	checkOneWinner(t, outs)
	if outs[0].res.Status != http.StatusCreated {
		t.Errorf("the first save answered %d", outs[0].res.Status)
	}
	if got := f.rows(); got != [4]int{1, 1, 1, 1} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestCreateOnConflict: another unique violation goes to OnConflict after
// the rollback; without it (or when it returns nil) the constraint error is
// returned as it is.
func TestCreateOnConflict(t *testing.T) {
	ctx := context.Background()
	f := newSaveFixture(t)
	dup := testPlan(f.farmA, func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, `insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying')`)
		return err
	})
	mapped := httpx.AlreadyEntered(f.farmA, "deja")
	var seen string
	c := create(f.farmA, nil, false)
	c.OnConflict = func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError) error {
		seen = ce.Constraint
		var n int // the pool is usable: the aborted transaction is gone
		if err := f.d.App.QueryRow(ctx, `select count(*) from entry`).Scan(&n); err != nil || n != 0 {
			t.Errorf("entries after the rollback = %d, %v", n, err)
		}
		return mapped
	}
	if _, err := f.r.Create(ctx, f.ion, c, dup); !errors.Is(err, mapped) || seen != "farm_name_uq" {
		t.Fatalf("err = %v, constraint %q", err, seen)
	}

	c.OnConflict = func(context.Context, *dbq.Queries, *pgerr.ConstraintError) error { return nil }
	_, err := f.r.Create(ctx, f.ion, c, dup)
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); !ok || ce.Constraint != "farm_name_uq" {
		t.Fatalf("err = %v", err)
	}
	if got := f.rows(); got != [4]int{0, 0, 0, 0} {
		t.Fatalf("rows entry/audit/live/replay = %v", got)
	}
}

// TestValidateWriteMeta: source is manual (or absent) or voice.
func TestValidateWriteMeta(t *testing.T) {
	for src, ok := range map[string]bool{"": true, "manual": true, "voice": true, "migration": false, "Voice": false} {
		var v httpx.Validator
		ValidateWriteMeta(&v, httpx.WriteMeta{Source: src})
		if v.Valid() != ok {
			t.Errorf("source %q valid = %v, want %v", src, v.Valid(), ok)
		}
		if !ok {
			is := v.Issues()
			if len(is) != 1 || is[0].Path != "source" || is[0].Code != httpx.IssueOutOfRange || is[0].Message != messages.InvalidData() {
				t.Errorf("source %q issues = %+v", src, is)
			}
		}
	}
}
