package entries

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// Concurrent changes of one saved entry (M5 COR-3): the change runner's farm
// lock and the entry's version keep two changes from both applying to the
// state a client read. The names start with TestConcurrent: the nightly CI
// job runs them 20 times under the race detector.

// changeTimeout bounds each test: a change stuck on a lock fails the test
// instead of hanging it.
const changeTimeout = 5 * time.Second

// changeCaller names a change in its context, for the lock seam.
type changeCaller struct{}

// waitForAdvisoryWaiter returns once a backend of pool's database waits on an
// advisory lock (the farm lock, §12.1), or reports why it can't.
func waitForAdvisoryWaiter(ctx context.Context, pool *pgxpool.Pool) error {
	for {
		var waiting bool
		err := pool.QueryRow(ctx, `select exists (select 1 from pg_locks l
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
			return errors.New("no change waits on the farm lock")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// lockHold is what holdFirstLock reports: which change held the farm lock
// first (its context's changeCaller name), and a channel closed once it has
// the lock.
type lockHold struct {
	name     atomic.Value
	acquired chan struct{}
}

// holdFirstLock makes the change that takes the farm lock first (with first
// set, the first change of that name) hold it until another change waits on
// it, so the two really overlap.
func holdFirstLock(f *dailyFixture, ctx context.Context, first string) *lockHold {
	h := &lockHold{acquired: make(chan struct{})}
	var taken atomic.Bool
	f.svc.Runner().SetLock(func(c context.Context, q *dbq.Queries, ids ...uuid.UUID) error {
		if err := ledger.LockFarms(c, q, ids...); err != nil {
			return err
		}
		name, _ := c.Value(changeCaller{}).(string)
		if (first == "" || name == first) && taken.CompareAndSwap(false, true) {
			h.name.Store(name)
			close(h.acquired)
			return waitForAdvisoryWaiter(ctx, f.d.App)
		}
		return nil
	})
	return h
}

// statusOf is a change's answer as a status and an error code: 200 "" for a
// success; a test failure for an error that is not an envelope (a 500).
func statusOf(t *testing.T, name string, res Result, err error) (int, string) {
	t.Helper()
	if err == nil {
		return res.Status, ""
	}
	var he httpx.HTTPError
	if !errors.As(err, &he) {
		t.Fatalf("%s: %v (a 500)", name, err)
	}
	status, body := he.HTTPError()
	return status, body.Code
}

// TestConcurrentPutsSameVersion: two corrections of one entry sent with the
// same ifVersion at the same time: the first to take the farm lock answers
// 200, the other, once it gets the lock, meets the moved version: 409 STALE.
// Exactly one value is written.
func TestConcurrentPutsSameVersion(t *testing.T) {
	f := newDailyFixture(t)
	ctx, cancel := context.WithTimeout(t.Context(), changeTimeout)
	defer cancel()
	e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
	v, ver := f.storedValues(e)
	reason := "Numărătoare refăcută."
	held := holdFirstLock(f, ctx, "")

	var (
		wg    sync.WaitGroup
		start = make(chan struct{})
		res   [2]Result
		errs  [2]error
	)
	for i, n := range []int64{2, 3} {
		req := DailyUpdateRequest{IfVersion: ver, Values: f.setMortality(v, "H5.2", "F", dnum(n)), Reason: &reason}
		wg.Go(func() {
			<-start
			res[i], errs[i] = f.svc.UpdateDaily(context.WithValue(ctx, changeCaller{}, string(rune('A'+i))), f.ion, e, req)
		})
	}
	close(start)
	wg.Wait()
	if held.name.Load() == nil {
		t.Fatal("no change took the farm lock")
	}

	var ok, stale int
	for i := range res {
		switch status, code := statusOf(t, string(rune('A'+i)), res[i], errs[i]); {
		case status == http.StatusOK:
			ok++
		case status == http.StatusConflict && code == httpx.CodeStale:
			stale++
		default:
			t.Errorf("PUT %c: %d %s", 'A'+i, status, code)
		}
	}
	if ok != 1 || stale != 1 {
		t.Fatalf("%d answered 200 and %d 409 STALE, want one of each", ok, stale)
	}
	if rows := f.birdRows(e, "H5.2", "F"); len(rows) != 1 || (rows[0].Qty != -2 && rows[0].Qty != -3) {
		t.Errorf("H5.2 ♀ rows %+v, want one: the winner's", rows)
	}
	var updates int
	if err := f.d.App.QueryRow(t.Context(), `select count(*) from audit_log where entry_id = $1 and action = 'update'`, e).
		Scan(&updates); err != nil {
		t.Fatal(err)
	}
	if updates != 1 {
		t.Errorf("%d update audit rows, want 1", updates)
	}
}

// TestConcurrentPutAndDelete: a correction and a delete of one entry at the
// same time. A DELETE carries no version (§7.8), so it always applies; the
// PUT applies when it takes the farm lock first and meets the deleted entry
// otherwise: a clean 409 STALE, never a 404 or a 500. Either way the entry
// ends deleted. The first case forces the DELETE to win, so the 409 path
// runs on every pass; the second lets the two race.
func TestConcurrentPutAndDelete(t *testing.T) {
	for _, tc := range []struct {
		name  string
		first string // the change that must hold the farm lock first; "" = a race
	}{{"delete first", "delete"}, {"race", ""}} {
		t.Run(tc.name, func(t *testing.T) {
			f := newDailyFixture(t)
			ctx, cancel := context.WithTimeout(t.Context(), changeTimeout)
			defer cancel()
			e := f.delDaily(f.vet, nov18, at0815, 0, delRow{"H5.2", "F", "mortality", -1})
			v, ver := f.storedValues(e)
			reason := "Numărătoare refăcută."
			put := DailyUpdateRequest{IfVersion: ver, Values: f.setMortality(v, "H5.2", "F", dnum(3)), Reason: &reason}
			held := holdFirstLock(f, ctx, tc.first)

			var (
				wg             sync.WaitGroup
				putRes, delRes Result
				putErr, delErr error
			)
			wg.Go(func() {
				delRes, delErr = f.svc.DeleteEntry(context.WithValue(ctx, changeCaller{}, "delete"), f.ion, e,
					DeleteRequest{Reason: ptr("Introdusă de două ori.")})
			})
			wg.Go(func() {
				if tc.first == "delete" { // the PUT starts once the DELETE holds the lock
					select {
					case <-held.acquired:
					case <-ctx.Done():
					}
				}
				putRes, putErr = f.svc.UpdateDaily(context.WithValue(ctx, changeCaller{}, "put"), f.ion, e, put)
			})
			wg.Wait()
			if held.name.Load() == nil {
				t.Fatal("no change took the farm lock")
			}

			delStatus, delCode := statusOf(t, "DELETE", delRes, delErr)
			putStatus, putCode := statusOf(t, "PUT", putRes, putErr)
			if delStatus != http.StatusOK {
				t.Errorf("DELETE: %d %s, want 200", delStatus, delCode)
			}
			switch held.name.Load() {
			case "delete":
				if putStatus != http.StatusConflict || putCode != httpx.CodeStale {
					t.Errorf("PUT after the delete: %d %s, want 409 STALE", putStatus, putCode)
				}
			case "put":
				if putStatus != http.StatusOK {
					t.Errorf("PUT before the delete: %d %s, want 200", putStatus, putCode)
				}
			}
			if rows, deleted, _ := f.liveRows(e); rows != 0 || !deleted {
				t.Errorf("entry: %d live rows, deleted %v, want deleted", rows, deleted)
			}
		})
	}
}
