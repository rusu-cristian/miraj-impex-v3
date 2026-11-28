package entries

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// at0815 is 18.11.2026 08:15 in Chișinău (06:15 UTC).
var at0815 = time.Date(2026, time.November, 18, 8, 15, 0, 0, timex.Chisinau)

// entry inserts an entry of kind on farm and day by p at the instant at
// (p.UserID nil-able through a zero principal: a migration-like entry).
func (f *feedFixture) entry(kind dbq.EntryKind, farm uuid.UUID, day timex.Day, p auth.Principal, at time.Time) uuid.UUID {
	f.t.Helper()
	var by *uuid.UUID
	if p.UserID != uuid.Nil {
		by = &p.UserID
	}
	return f.id(`insert into entry (kind, farm_id, date, created_by, created_at) values ($1, $2, $3, $4, $5) returning id`,
		kind, farm, day, by, at)
}

// zilnic is the vet's Zilnic of 18.11 saved at 08:15: H5.1 ♀ 2 dead.
func (f *feedFixture) zilnic() uuid.UUID {
	f.t.Helper()
	e := f.entry(dbq.EntryKindDaily, f.farmA, nov18, f.vet, at0815)
	f.bird(e, "H5.1", nov18, "F", "mortality", -2)
	return e
}

// allowAll lets every change through.
func allowAll(EntryMeta, EntryActions) error { return nil }

type changeResponse struct {
	Entry httpx.EntryRef  `json:"entry"`
	AsOf  timex.LocalTime `json:"asOf"`
}

// changePlan moves the version from the locked one, runs between (if not
// nil), answers a dry run with an empty preview, else an entry.updated event
// and an audit row whose entry columns the runner fills.
func changePlan(between func(ctx context.Context, tx *ChangeTx) error) ChangePlan {
	return func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		ref, err := tx.UpdateHeader(ctx, tx.Entry.VersionAt, tx.Entry.Corrected)
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
		summary := "Mortalitate · 1 secțiune"
		return Saved{
			Response: changeResponse{Entry: ref, AsOf: timex.At(tx.Now)},
			Audit:    audit.Row{Summary: &summary, RecordTypes: []string{"mortalitate"}},
			Events: []live.Event{live.EntryUpdated(live.EntryPayload{
				EntryID: tx.Entry.ID, Kind: tx.Entry.Kind, FarmIDs: tx.Entry.FarmIDs(), At: timex.At(tx.Now), Summary: summary,
			})},
		}, nil
	}
}

func (f *feedFixture) change(p auth.Principal, c ChangeSpec, plan ChangePlan) (Result, error) {
	f.t.Helper()
	if c.Allow == nil {
		c.Allow = allowAll
	}
	if c.Action == "" {
		c.Action = dbq.AuditActionUpdate
	}
	return f.r.Change(f.t.Context(), p, c, plan)
}

// T3 criterion 1: a change answers 200 with the plan's response and writes
// one audit row (the entry's columns filled by the runner, at and user from
// the clock and the caller) and its live events, once.
func TestChangeSaves(t *testing.T) {
	f := newFeedFixture(t)
	e := f.zilnic()
	before := f.rows()
	res, err := f.change(f.ion, ChangeSpec{EntryID: e}, changePlan(nil))
	if err != nil {
		t.Fatal(err)
	}
	var resp changeResponse
	if err := json.Unmarshal(res.Body, &resp); err != nil {
		t.Fatal(err)
	}
	// 08:15 + 1 µs < 12:00: the version is the clock's instant.
	if res.Status != http.StatusOK || resp.Entry.ID != e || resp.Entry.Version != "2026-11-18T10:00:00.000000Z" {
		t.Fatalf("%d %+v", res.Status, resp)
	}
	if got := f.rows(); got != [4]int{before[0], before[1] + 1, before[2] + 1, before[3]} {
		t.Fatalf("rows entry/audit/live/replay = %v, before %v", got, before)
	}
	var (
		at                time.Time
		user, entity, eid uuid.UUID
		action            dbq.AuditAction
		kind              dbq.EntryKind
		farms             []uuid.UUID
		day               timex.Day
		updatedBy         uuid.UUID
		corrected         bool
	)
	if err := f.d.App.QueryRow(t.Context(), `select at, user_id, action, entity_id, entry_id, entry_kind, farm_ids, record_date
		from audit_log`).Scan(&at, &user, &action, &entity, &eid, &kind, &farms, &day); err != nil {
		t.Fatal(err)
	}
	if !at.Equal(testdb.Now()) || user != f.ion.UserID || action != dbq.AuditActionUpdate || entity != e || eid != e ||
		kind != dbq.EntryKindDaily || len(farms) != 1 || farms[0] != f.farmA || day != nov18 {
		t.Errorf("audit = %v %v %v %v %v %v %v %v", at, user, action, entity, eid, kind, farms, day)
	}
	if err := f.d.App.QueryRow(t.Context(), `select updated_by, corrected from entry where id = $1`, e).
		Scan(&updatedBy, &corrected); err != nil {
		t.Fatal(err)
	}
	if updatedBy != f.ion.UserID || corrected {
		t.Errorf("entry updated_by %v corrected %v", updatedBy, corrected)
	}
}

// T3 criterion 2 (plan M5-D3): two changes at the same frozen instant give
// two versions 1 µs apart; a client still holding the first gets 409 STALE
// with the §7.0 sentence and nothing is written.
func TestUpdateHeaderVersions(t *testing.T) {
	f := newFeedFixture(t)
	e := f.entry(dbq.EntryKindDaily, f.farmA, nov18, f.ion, testdb.Now()) // created at the frozen 12:00
	update := func(ifVersion string) (Result, error) {
		v, err := ParseVersion(ifVersion)
		if err != nil {
			t.Fatal(err)
		}
		return f.change(f.ion, ChangeSpec{EntryID: e}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
			ref, err := tx.UpdateHeader(ctx, v, true)
			if err != nil {
				return Saved{}, err
			}
			return Saved{Response: changeResponse{Entry: ref}}, nil
		})
	}
	versionOf := func(res Result) string {
		var r changeResponse
		if err := json.Unmarshal(res.Body, &r); err != nil {
			t.Fatal(err)
		}
		return r.Entry.Version
	}
	created := "2026-11-18T10:00:00.000000Z"
	first, err := update(created)
	if err != nil {
		t.Fatal(err)
	}
	second, err := update(versionOf(first))
	if err != nil {
		t.Fatal(err)
	}
	if v1, v2 := versionOf(first), versionOf(second); v1 != "2026-11-18T10:00:00.000001Z" || v2 != "2026-11-18T10:00:00.000002Z" {
		t.Fatalf("versions %s, %s", v1, v2)
	}
	audits := f.count("audit_log")
	_, err = update(versionOf(first))
	body := wantCode(t, err, http.StatusConflict, httpx.CodeStale)
	if body.Message != messages.Stale() || body.Message != "Înregistrarea s-a schimbat între timp. Reîncarcă și încearcă din nou." {
		t.Errorf("message = %q", body.Message)
	}
	if f.count("audit_log") != audits {
		t.Error("a stale change wrote an audit row")
	}
	var corrected bool
	var v time.Time
	if err := f.d.App.QueryRow(t.Context(), `select corrected, updated_at from entry where id = $1`, e).Scan(&corrected, &v); err != nil {
		t.Fatal(err)
	}
	if !corrected || version(v) != versionOf(second) {
		t.Errorf("entry corrected %v version %s", corrected, version(v))
	}
	if _, err := ParseVersion("18.11.2026"); err == nil {
		t.Error("a malformed version parsed")
	}
}

// T3 criterion 1: 404 for an unknown entry, a deleted one (unless
// restoring) and a kind not served (Q1); 403 FARM_FORBIDDEN for a farm
// outside the caller's. The plan never runs.
func TestChangeRefusedBeforePlan(t *testing.T) {
	f := newFeedFixture(t)
	deleted := f.zilnic()
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, deleted, testdb.Now(), f.ion.UserID)
	f.exec(`update bird_event set deleted_at = $2 where entry_id = $1`, deleted, testdb.Now())
	sale := f.entry(dbq.EntryKindSale, f.farmA, nov18, f.ion, at0815)
	migration := f.entry(dbq.EntryKindMigration, f.farmA, nov1, auth.Principal{}, at0815)
	alive := f.zilnic()
	ran := false
	plan := func(context.Context, *ChangeTx) (Saved, error) { ran = true; return Saved{}, errors.New("ran") }
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		id     uuid.UUID
		action dbq.AuditAction
		status int
		code   string
	}{
		{"unknown", f.ion, uuid.New(), dbq.AuditActionUpdate, 404, httpx.CodeNotFound},
		{"deleted, update", f.ion, deleted, dbq.AuditActionUpdate, 404, httpx.CodeNotFound},
		{"deleted, undo", f.ion, deleted, dbq.AuditActionUndo, 404, httpx.CodeNotFound},
		{"sale", f.ion, sale, dbq.AuditActionDelete, 404, httpx.CodeNotFound},
		{"migration", f.owner, migration, dbq.AuditActionUpdate, 404, httpx.CodeNotFound},
		{"another farm", f.outsider, alive, dbq.AuditActionUpdate, 403, httpx.CodeFarmForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := f.change(tc.p, ChangeSpec{EntryID: tc.id, Action: tc.action}, plan)
			wantCode(t, err, tc.status, tc.code)
		})
	}
	if ran {
		t.Error("a plan ran")
	}

	// Restore reaches the plan for a deleted entry.
	var sawDeleted bool
	_, err := f.change(f.ion, ChangeSpec{EntryID: deleted, Action: dbq.AuditActionRestore},
		func(_ context.Context, tx *ChangeTx) (Saved, error) {
			sawDeleted = tx.Entry.Deleted
			return Saved{}, errors.New("stop")
		})
	if err == nil || err.Error() != "stop" || !sawDeleted {
		t.Errorf("restore: err %v, saw deleted %v", err, sawDeleted)
	}
}

// T3 criterion 1: Allow decides with the caller's §3.2 actions under the
// locks; its error is the answer. A spec without Allow is refused.
func TestChangeAllow(t *testing.T) {
	f := newFeedFixture(t)
	e := f.zilnic()
	var got []EntryActions
	record := func(_ EntryMeta, a EntryActions) error {
		got = append(got, a)
		return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
	}
	for _, p := range []auth.Principal{f.vet, f.ion, f.viewer} {
		_, err := f.r.Change(t.Context(), p, ChangeSpec{EntryID: e, Action: dbq.AuditActionUpdate, Allow: record}, changePlan(nil))
		wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
	}
	want := []EntryActions{{Editable: true}, {Correctable: true}, {}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("actions = %+v, want %+v", got, want)
	}
	if _, err := f.r.Change(t.Context(), f.ion, ChangeSpec{EntryID: e, Action: dbq.AuditActionUpdate}, changePlan(nil)); err == nil {
		t.Error("a change without Allow ran")
	}
	if got := f.rows(); got[1] != 0 || got[2] != 0 {
		t.Errorf("audit/live rows = %d/%d", got[1], got[2])
	}
}

// T3 criterion 1: an entry dated inside a cycle of a closed series is 409
// SERIES_CLOSED after Allow (which sees SeriesClosed and actions as if the
// series were open); a refusal by Allow comes first.
func TestChangeSeriesClosed(t *testing.T) {
	f := newFeedFixture(t)
	e := f.entry(dbq.EntryKindDaily, f.farmB, timex.MustDay(2026, time.November, 5), f.vet, at0815)
	var meta EntryMeta
	var acts EntryActions
	_, err := f.change(f.owner, ChangeSpec{EntryID: e, Allow: func(m EntryMeta, a EntryActions) error {
		meta, acts = m, a
		return nil
	}}, changePlan(nil))
	body := wantCode(t, err, http.StatusConflict, httpx.CodeSeriesClosed)
	if body.Message != messages.SeriesClosed("Seria 2025") || !meta.SeriesClosed || !acts.Correctable {
		t.Errorf("message %q, meta %+v, actions %+v", body.Message, meta, acts)
	}
	_, err = f.change(f.owner, ChangeSpec{EntryID: e, Allow: func(EntryMeta, EntryActions) error { return httpx.Forbidden() }},
		changePlan(nil))
	wantCode(t, err, http.StatusForbidden, httpx.CodeForbidden)
}

// T3 criterion 1: a dry run runs the plan's writes, answers its Preview (200)
// and rolls every write back: version, child rows, audit, live.
func TestChangeDryRun(t *testing.T) {
	f := newFeedFixture(t)
	e := f.zilnic()
	res, err := f.change(f.ion, ChangeSpec{EntryID: e, DryRun: true}, changePlan(func(ctx context.Context, tx *ChangeTx) error {
		_, err := tx.Exec(ctx, `update bird_event set deleted_at = $2 where entry_id = $1`, tx.Entry.ID, tx.Now)
		return err
	}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != http.StatusOK || strings.TrimSpace(string(res.Body)) != `{"effects":[],"warnings":[],"refusals":[]}` {
		t.Fatalf("%d %s", res.Status, res.Body)
	}
	var updated *time.Time
	var liveRows int
	if err := f.d.App.QueryRow(t.Context(), `select e.updated_at, (select count(*) from bird_event b where b.entry_id = e.id
		and b.deleted_at is null) from entry e where e.id = $1`, e).Scan(&updated, &liveRows); err != nil {
		t.Fatal(err)
	}
	if updated != nil || liveRows != 1 {
		t.Errorf("after the dry run: updated_at %v, live rows %d", updated, liveRows)
	}
	if got := f.rows(); got[1] != 0 || got[2] != 0 {
		t.Errorf("audit/live rows = %d/%d", got[1], got[2])
	}

	// A dry run whose plan never calls Preview is an error, not a save.
	_, err = f.change(f.ion, ChangeSpec{EntryID: e, DryRun: true}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		return Saved{}, nil
	})
	if err == nil {
		t.Error("a dry run without Preview succeeded")
	}
}

// T3 criterion 1: a unique violation of a child row is mapped after the
// transaction by OnConflict, on the pool (the aborted transaction is gone):
// 409 ALREADY_ENTERED with the holder. Without OnConflict it stays a 500.
func TestChangeOnConflict(t *testing.T) {
	f := newFeedFixture(t)
	holder := f.zilnic() // H5.1 ♀ mortality 18.11
	e := f.entry(dbq.EntryKindDaily, f.farmA, nov18, f.ion, at0815)
	f.bird(e, "H5.2", nov18, "F", "mortality", -1)
	addH51 := changePlan(func(ctx context.Context, tx *ChangeTx) error {
		_, err := tx.Exec(ctx, `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
			values ($1, $2, $3, $4, 'F', 'mortality', -1)`, tx.Entry.ID, tx.Entry.FarmID, f.sec["H5.1"], tx.Entry.Date)
		return err
	})
	var constraint string
	_, err := f.change(f.ion, ChangeSpec{EntryID: e, OnConflict: func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError, _ ChangeConflict) error {
		constraint = ce.Constraint
		if _, err := q.EntryChangeHeader(ctx, holder); err != nil { // q is the pool's: the aborted tx is gone
			return err
		}
		return httpx.AlreadyEntered(holder, "deja")
	}}, addH51)
	body := wantCode(t, err, http.StatusConflict, httpx.CodeAlreadyEntered)
	if constraint != "bird_event_daily_uq" || body.EntryID == nil || *body.EntryID != holder {
		t.Errorf("constraint %q, holder %v", constraint, body.EntryID)
	}
	_, err = f.change(f.ion, ChangeSpec{EntryID: e}, addH51)
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); !ok || ce.Constraint != "bird_event_daily_uq" {
		t.Errorf("without OnConflict: %v", err)
	}
	if got := f.rows(); got[1] != 0 || got[2] != 0 {
		t.Errorf("audit/live rows = %d/%d", got[1], got[2])
	}
}

// T3 criterion 1: the farm locks come before the entry's row lock, and the
// clock is read once they are held. While another transaction holds farm A's
// lock the change waits and the entry row stays unlocked (another session
// locks it with NOWAIT); the plan runs at the clock's instant after the wait.
func TestChangeLocksThenReadsThenClock(t *testing.T) {
	ctx := t.Context()
	f := newFeedFixture(t)
	e := f.zilnic()
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
		changeErr error
		today     timex.Day
		planNow   time.Time
	)
	wg.Go(func() {
		_, changeErr = r.Change(ctx, f.ion, ChangeSpec{EntryID: e, Action: dbq.AuditActionUpdate, Allow: allowAll},
			changePlan(func(_ context.Context, tx *ChangeTx) error { today, planNow = tx.Today, tx.Now; return nil }))
	})
	f.waitFor(1)
	other, err := f.d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Exec(ctx, `select 1 from entry where id = $1 for update nowait`, e); err != nil {
		t.Errorf("the entry row is locked before the farm lock is held: %v", err)
	}
	_ = other.Rollback(ctx)
	clock.Set(later)
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if changeErr != nil {
		t.Fatal(changeErr)
	}
	if want := timex.MustDay(2026, time.November, 19); today != want || !planNow.Equal(later) {
		t.Errorf("plan saw Today %s, Now %v; want %s, %v", today, planNow, want, later)
	}
}

// T3 criterion 3 (plan M5-D8): production runners lock with
// ledger.LockFarms and have no beforeCheck.
func TestRunnerSeamDefaults(t *testing.T) {
	store := db.NewStore(nil)
	for name, r := range map[string]*Runner{
		"NewRunner": NewRunner(store, timex.RealClock{}), "NewService": NewService(store, timex.RealClock{}).runner,
	} {
		if reflect.ValueOf(r.lock).Pointer() != reflect.ValueOf(ledger.LockFarms).Pointer() {
			t.Errorf("%s: lock is not ledger.LockFarms", name)
		}
		if r.beforeCheck != nil {
			t.Errorf("%s: beforeCheck is set", name)
		}
	}
}

// T3 criterion 3: both runner paths take the farm locks through the seam,
// and Tx.CheckCounts runs beforeCheck, then §6.1 on the write's
// transaction: a cull of 2 000 on H5.1 ♀ 18.11 (1 000 at the start of the
// day, 2 already dead) leaves −1 002.
func TestRunnerSeams(t *testing.T) {
	f := newFeedFixture(t)
	e := f.zilnic()
	var locked [][]uuid.UUID
	f.r.SetLock(func(ctx context.Context, q *dbq.Queries, ids ...uuid.UUID) error {
		locked = append(locked, ids)
		return ledger.LockFarms(ctx, q, ids...)
	})
	checks := 0
	f.r.SetBeforeCheck(func() { checks++ })

	var negs []ledger.Negative
	_, err := f.change(f.ion, ChangeSpec{EntryID: e}, changePlan(func(ctx context.Context, tx *ChangeTx) error {
		if _, err := tx.Exec(ctx, `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
			values ($1, $2, $3, $4, 'F', 'cull', -2000, 'boala')`, tx.Entry.ID, tx.Entry.FarmID, f.sec["H5.1"], tx.Entry.Date); err != nil {
			return err
		}
		var err error
		negs, err = tx.CheckCounts(ctx, []ledger.SectionSex{{SectionID: f.sec["H5.1"], Sex: dbq.SexF}}, tx.Entry.Date)
		if err != nil {
			return err
		}
		return errors.New("refused")
	}))
	if err == nil || err.Error() != "refused" {
		t.Fatalf("err = %v", err)
	}
	if checks != 1 || len(negs) != 1 || negs[0].Code != "H5.1" || negs[0].MinAliveEnd != -1002 || negs[0].FirstDate != nov18 {
		t.Errorf("checks %d, negatives %+v", checks, negs)
	}
	if _, err := f.r.Create(t.Context(), f.ion, create(f.farmA, nil, false), testPlan(f.farmA, nil)); err != nil {
		t.Fatal(err)
	}
	if len(locked) != 2 || len(locked[0]) != 1 || locked[0][0] != f.farmA || len(locked[1]) != 1 || locked[1][0] != f.farmA {
		t.Errorf("locked = %v", locked)
	}
}
