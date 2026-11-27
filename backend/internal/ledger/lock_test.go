package ledger

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

// recorder is a dbq.DBTX that records the farm id of every LockFarm it runs.
type recorder struct {
	pgx.Tx
	locked []uuid.UUID
}

func (r *recorder) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.HasPrefix(sql, "-- name: LockFarm ") {
		r.locked = append(r.locked, args[0].(uuid.UUID))
	}
	return r.Tx.Exec(ctx, sql, args...)
}

// advisoryLocks counts the advisory locks granted to the backend running tx.
func advisoryLocks(t *testing.T, ctx context.Context, tx pgx.Tx) int {
	t.Helper()
	var n int
	err := tx.QueryRow(ctx, `select count(*) from pg_locks where locktype = 'advisory' and granted and pid = pg_backend_pid()`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// TestLockFarms: LockFarms(b, a, b) takes two locks, a before b; a second
// transaction locking a waits until the first commits.
func TestLockFarms(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	a := uuid.MustParse("10000000-0000-0000-0000-000000000000")
	b := uuid.MustParse("f0000000-0000-0000-0000-000000000000")

	var ha, hb int64
	if err := d.App.QueryRow(ctx, `select hashtextextended($1::text, 0), hashtextextended($2::text, 0)`, a, b).Scan(&ha, &hb); err != nil {
		t.Fatal(err)
	}
	if ha == hb {
		t.Fatal("the test farm ids collide under hashtextextended")
	}

	tx1, err := d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx1.Rollback(ctx) }()
	rec := &recorder{Tx: tx1}
	if err := LockFarms(ctx, dbq.New(rec), b, a, b); err != nil {
		t.Fatal(err)
	}
	if want := []uuid.UUID{a, b}; !slices.Equal(rec.locked, want) {
		t.Fatalf("lock order = %v, want %v", rec.locked, want)
	}
	if n := advisoryLocks(t, ctx, tx1); n != 2 {
		t.Fatalf("advisory locks held = %d, want 2", n)
	}

	done := make(chan error, 1)
	go func() {
		tx2, err := d.App.Begin(ctx)
		if err != nil {
			done <- err
			return
		}
		defer func() { _ = tx2.Rollback(ctx) }()
		done <- LockFarms(ctx, dbq.New(tx2), a)
	}()
	select {
	case err := <-done:
		t.Fatalf("second lock on a returned before the first commit: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("second lock on a still waits 5 s after the first commit")
	}
}

// TestLockFarmsNone: no farm, no lock and no error.
func TestLockFarmsNone(t *testing.T) {
	ctx := context.Background()
	d := testdb.New(t)
	tx, err := d.App.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := LockFarms(ctx, dbq.New(tx)); err != nil {
		t.Fatal(err)
	}
	if n := advisoryLocks(t, ctx, tx); n != 0 {
		t.Fatalf("advisory locks held = %d, want 0", n)
	}
}
