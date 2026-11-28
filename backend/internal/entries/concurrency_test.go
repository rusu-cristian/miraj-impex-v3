package entries_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The §11.8 concurrency tests M5 can run (plan M5 Q13, D8), on the dev seed:
// two Zilnic saves of Ion Rusu on Hîrbovăț H6.2 ♂ for 18.11.2026, at the
// same time. H6.2 ♂ starts the day with 93 cocoși and ends it with 5 (golden
// 7: 88 sold). Save A types Sacrificare 4 (boală), save B Eroare sexare −4:
// different Zilnic keys, so neither meets the other's unique index, and each
// alone passes Z5 against the 93 and §6.1 (5 − 4 = 1). Together they would
// leave 5 − 4 − 4 = −3. Test (2), opposite-direction transfers, comes with M7.
//
// The names start with TestConcurrent: the nightly CI job runs them 20 times
// under the race detector (go test -race -count=20 -run 'Concurrent').

// concurrentTimeout bounds each test: a save stuck on a lock or a barrier
// fails the test instead of hanging it.
const concurrentTimeout = 5 * time.Second

// h62Fixture is the seeded database, Ion Rusu, Hîrbovăț and its H6.2.
type h62Fixture struct {
	d    *testdb.DB
	svc  *entries.Service
	ion  auth.Principal
	refs entries.DailyRefs
	farm uuid.UUID
	h62  uuid.UUID
}

// newH62Fixture loads the fixture and checks the seed is as the tests assume
// (plan M5 Verify "The seed's 18.11 Zilnic entries"): H6.2 ♂ alive 93 at the
// start of 18.11 and 5 at its end, and no live cull or sexing-error row on
// H6.2 ♂ that day, so both keys are free.
func newH62Fixture(t *testing.T) h62Fixture {
	t.Helper()
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	f := h62Fixture{d: d, svc: entries.NewService(db.NewStore(d.App), d.Clock), farm: seededFarm(t, d.App, "hirbovat")}
	f.ion = seededUser(t, d.App, "ion.rusu")
	refs, err := f.svc.DailyRefs(ctx, f.farm)
	if err != nil {
		t.Fatal(err)
	}
	f.refs = refs
	if err := d.App.QueryRow(ctx, `select id from section where farm_id = $1 and code = 'H6.2'`, f.farm).Scan(&f.h62); err != nil {
		t.Fatal(err)
	}
	if start, end := f.aliveM(t, testdb.Today.AddDays(-1)), f.aliveM(t, testdb.Today); start != 93 || end != 5 {
		t.Fatalf("seed: H6.2 ♂ %d at the start and %d at the end of 18.11, want 93 and 5", start, end)
	}
	var held int
	if err := d.App.QueryRow(ctx, `select count(*) from bird_event where section_id = $1 and date = $2 and sex = 'M'
		and type in ('cull', 'sexing_error') and deleted_at is null`, f.h62, testdb.Today).Scan(&held); err != nil {
		t.Fatal(err)
	}
	if held != 0 {
		t.Fatalf("seed: H6.2 ♂ holds %d cull or sexing-error rows on 18.11, want 0", held)
	}
	return f
}

// aliveM is H6.2 ♂ at the end of day (calc, running sums).
func (f h62Fixture) aliveM(t *testing.T, day timex.Day) int {
	t.Helper()
	secs, err := calc.SectionsAliveEnd(t.Context(), f.d.App, f.farm, day)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range secs {
		if s.SectionID == f.h62 {
			return s.AliveEnd.M
		}
	}
	t.Fatalf("no H6.2 among the sections alive on %s", day)
	return 0
}

// requests are saves A (Sacrificare 4, boală) and B (Eroare sexare −4), both
// H6.2 ♂ on 18.11.2026, each with its own clientRequestId, checked at level
// 1 as the handler does.
func (f h62Fixture) requests(t *testing.T) [2]entries.DailyCreateRequest {
	t.Helper()
	four, minusFour, reason := num.FromInt(4), num.FromInt(-4), string(dbq.CullReasonBoala)
	sexes := [2]*entries.DailySex{{Cull: &four, CullReason: &reason}, {SexingError: &minusFour}}
	var out [2]entries.DailyCreateRequest
	for i, m := range sexes {
		key := uuid.New()
		out[i] = entries.DailyCreateRequest{
			WriteMeta: httpx.WriteMeta{ClientRequestID: &key},
			DailyValues: entries.DailyValues{FarmID: f.farm, Date: testdb.Today, Eggs: []entries.DailyEggs{},
				Sections: []entries.DailySection{{SectionID: f.h62, M: m}}},
		}
		var v httpx.Validator
		if out[i].Validate(&v, f.refs); v.Err() != nil {
			t.Fatalf("request %d: %v", i, v.Err())
		}
	}
	return out
}

type saveOutcome struct {
	res entries.Result
	err error
}

// saveBoth starts both requests at once (a start barrier, one pool
// connection each) and waits for both answers.
func (f h62Fixture) saveBoth(ctx context.Context, reqs [2]entries.DailyCreateRequest) [2]saveOutcome {
	var (
		wg    sync.WaitGroup
		outs  [2]saveOutcome
		start = make(chan struct{})
	)
	for i, req := range reqs {
		wg.Go(func() {
			<-start
			res, err := f.svc.CreateDaily(ctx, f.ion, req)
			outs[i] = saveOutcome{res, err}
		})
	}
	close(start)
	wg.Wait()
	return outs
}

// waitForLockWaiter returns once a backend of this test's database waits on
// an advisory lock (the farm lock, §12.1), or reports why it can't.
func (f h62Fixture) waitForLockWaiter(ctx context.Context) error {
	for {
		var waiting bool
		err := f.d.App.QueryRow(ctx, `select exists (select 1 from pg_locks l
			where l.locktype = 'advisory' and not l.granted
			  and l.database = (select oid from pg_database where datname = current_database()))`).Scan(&waiting)
		if err != nil {
			return err
		}
		if waiting {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("no save waits on the farm lock")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// negativeDays counts the section × sex days of the farm below 0 up to
// app.today (18.11) in the dense view, the §6.1 invariant's own reading.
func (f h62Fixture) negativeDays(t *testing.T) int {
	t.Helper()
	var n int
	if err := f.d.App.QueryRow(t.Context(), `select count(*) from v_section_sex_day where farm_id = $1 and alive_end < 0`,
		f.farm).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestConcurrentZilnicSavesHoldTheCount is §11.8 (1) in its M5 form: with the
// farm lock, the save that gets it first is held at its §6.1 check until the
// other one is seen waiting on the lock, so the two really overlap; then one
// answers 201 and the other 422 with the bound "Scrie cel mult 1." (5 − 4 =
// 1 left for it), and no day of the farm is below 0 afterwards.
func TestConcurrentZilnicSavesHoldTheCount(t *testing.T) {
	f := newH62Fixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), concurrentTimeout)
	defer cancel()

	var (
		calls   atomic.Int32
		waitErr = make(chan error, 1)
	)
	f.svc.Runner().SetBeforeCheck(func() {
		if calls.Add(1) == 1 { // the first save, holding the farm lock after its insert
			waitErr <- f.waitForLockWaiter(ctx)
		}
	})
	outs := f.saveBoth(ctx, f.requests(t))
	select { // both saves have answered: the hook, if it ran, has sent
	case err := <-waitErr:
		if err != nil {
			t.Fatalf("the saves did not overlap: %v", err)
		}
	default:
		t.Fatalf("no save reached the §6.1 check: %+v", outs)
	}

	var created, refused int
	for i, o := range outs {
		if o.err == nil {
			if o.res.Status != http.StatusCreated {
				t.Errorf("save %c: %d %s", 'A'+i, o.res.Status, o.res.Body)
			}
			created++
			continue
		}
		he, ok := o.err.(httpx.HTTPError)
		if !ok {
			t.Fatalf("save %c: %v", 'A'+i, o.err)
		}
		status, body := he.HTTPError()
		if status != http.StatusUnprocessableEntity || body.Code != "REFUSED" || len(body.Refusals) != 1 {
			t.Fatalf("save %c: %d %+v", 'A'+i, status, body)
		}
		r := body.Refusals[0]
		if !strings.HasSuffix(r.Message, "Scrie cel mult 1.") || r.Max == nil || *r.Max != 1 || r.SectionCode != "H6.2" || r.Sex != "M" {
			t.Errorf("save %c refusal %+v, want H6.2 ♂ \"… Scrie cel mult 1.\" with max 1", 'A'+i, r)
		}
		refused++
	}
	if created != 1 || refused != 1 {
		t.Fatalf("%d saved and %d refused, want one of each", created, refused)
	}
	if got := f.aliveM(t, testdb.Today); got != 1 {
		t.Errorf("H6.2 ♂ at the end of 18.11 = %d, want 1 (5 − 4)", got)
	}
	if n := f.negativeDays(t); n != 0 {
		t.Errorf("%d section × sex days of Hîrbovăț below 0, want none", n)
	}
}

// TestConcurrentZilnicSavesWithoutTheLockBreakTheCount is §11.8 (3): with the
// farm lock stubbed out and both saves held at a barrier just before their
// §6.1 check (after their inserts, before either commits), each check reads
// under READ COMMITTED only its own uncommitted rows, both pass, both answer
// 201, and H6.2 ♂ ends 18.11 at 5 − 4 − 4 = −3. It proves the test of (1)
// can fail: the lock is what holds the invariant.
func TestConcurrentZilnicSavesWithoutTheLockBreakTheCount(t *testing.T) {
	f := newH62Fixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), concurrentTimeout)
	defer cancel()

	f.svc.Runner().SetLock(func(context.Context, *dbq.Queries, ...uuid.UUID) error { return nil })
	var (
		arrived  sync.WaitGroup
		released = make(chan struct{})
		late     atomic.Bool
	)
	arrived.Add(2)
	go func() { arrived.Wait(); close(released) }()
	f.svc.Runner().SetBeforeCheck(func() {
		arrived.Done()
		select {
		case <-released:
		case <-ctx.Done():
			late.Store(true)
		}
	})
	outs := f.saveBoth(ctx, f.requests(t))
	if late.Load() {
		t.Fatal("the two saves never stood at the §6.1 check together")
	}
	for i, o := range outs {
		if o.err != nil || o.res.Status != http.StatusCreated {
			t.Fatalf("save %c: %d %s %v, want 201 (no lock, nothing to see)", 'A'+i, o.res.Status, o.res.Body, o.err)
		}
	}
	if got := f.aliveM(t, testdb.Today); got != -3 {
		t.Errorf("H6.2 ♂ at the end of 18.11 = %d, want −3 (5 − 4 − 4)", got)
	}
	if n := f.negativeDays(t); n != 1 {
		t.Errorf("%d section × sex days of Hîrbovăț below 0, want 1 (H6.2 ♂ on 18.11)", n)
	}
}
