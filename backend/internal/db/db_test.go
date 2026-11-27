package db_test

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

func insertUser(t *testing.T, q *dbq.Queries, login string) (dbq.InsertUserRow, error) {
	t.Helper()
	return q.InsertUser(context.Background(), dbq.InsertUserParams{
		Name: "Ion Rusu", Login: login, PasswordHash: "$argon2id$x", Role: dbq.UserRoleManager, Now: testdb.Now(),
	})
}

func TestUniqueViolation(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	q := dbq.New(d.App)
	if _, err := insertUser(t, q, "ion.rusu"); err != nil {
		t.Fatal(err)
	}
	_, err := insertUser(t, q, "ion.rusu")
	ce, ok := pgerr.AsConstraint(err, pgerr.Unique)
	if !ok {
		t.Fatalf("err = %v, want a unique violation", err)
	}
	if ce.Constraint != "app_user_login_uq" || ce.Table != "app_user" {
		t.Fatalf("constraint = %q on %q", ce.Constraint, ce.Table)
	}
	var pgErr *pgconn.PgError
	if !errors.As(ce, &pgErr) || pgErr.Code != "23505" {
		t.Fatalf("does not unwrap to the 23505 PgError: %v", ce)
	}
}

func TestMapErrorKinds(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	store := db.NewStore(d.App)

	// 23514: the login check constraint.
	_, err := insertUser(t, store.Q(), "Bad Login!")
	if ce, ok := pgerr.AsConstraint(err, pgerr.Check); !ok || ce.Constraint != "app_user_login_check" {
		t.Errorf("check: %v", err)
	}
	// 23503: a session for a user that doesn't exist.
	_, err = store.Q().InsertSession(ctx, dbq.InsertSessionParams{
		UserID: [16]byte{1}, TokenHash: []byte{1}, Now: testdb.Now(), ExpiresAt: testdb.Now(),
	})
	if ce, ok := pgerr.AsConstraint(err, pgerr.ForeignKey); !ok || ce.Constraint != "user_session_user_id_fkey" {
		t.Errorf("foreign key: %v", err)
	}
	// Other errors pass through unchanged.
	plain := errors.New("plain")
	if pgerr.MapError(plain) != plain || pgerr.MapError(nil) != nil {
		t.Error("MapError changed a non-PostgreSQL error")
	}
}

func TestInTx(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	store := db.NewStore(d.App)
	count := func() int {
		var n int
		if err := d.App.QueryRow(ctx, "select count(*) from login_attempt").Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	insert := func(tx db.Tx) error {
		_, err := tx.Q.InsertLoginAttempt(ctx, dbq.InsertLoginAttemptParams{Login: "ion", At: testdb.Now()})
		return err
	}

	if err := store.InTx(ctx, insert); err != nil || count() != 1 {
		t.Fatalf("commit: err %v, rows %d", err, count())
	}
	if err := store.InTx(ctx, func(tx db.Tx) error { _ = insert(tx); return db.ErrRollback }); err != nil || count() != 1 {
		t.Fatalf("ErrRollback: err %v, rows %d", err, count())
	}
	boom := errors.New("boom")
	if err := store.InTx(ctx, func(tx db.Tx) error { _ = insert(tx); return boom }); !errors.Is(err, boom) || count() != 1 {
		t.Fatalf("error: err %v, rows %d", err, count())
	}
	// A constraint error inside the transaction comes back typed.
	err := store.InTx(ctx, func(tx db.Tx) error {
		if _, err := insertUser(t, tx.Q, "dup.login"); err != nil {
			return err
		}
		_, err := insertUser(t, tx.Q, "dup.login")
		return err
	})
	if _, ok := pgerr.AsConstraint(err, pgerr.Unique); !ok {
		t.Fatalf("constraint in tx: %v", err)
	}
	var users int
	if err := d.App.QueryRow(ctx, "select count(*) from app_user").Scan(&users); err != nil || users != 0 {
		t.Fatalf("rolled-back tx left %d users", users)
	}
}

func TestInReadTx(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	store := db.NewStore(d.App)

	// One REPEATABLE READ, READ ONLY transaction: a row committed by another
	// connection after the first read stays invisible to the second.
	var iso, ro string
	var before, after int
	err := store.InReadTx(ctx, func(q dbq.DBTX) error {
		if err := q.QueryRow(ctx, "select current_setting('transaction_isolation'), current_setting('transaction_read_only')").Scan(&iso, &ro); err != nil {
			return err
		}
		if err := q.QueryRow(ctx, "select count(*) from login_attempt").Scan(&before); err != nil {
			return err
		}
		if _, err := dbq.New(d.App).InsertLoginAttempt(ctx, dbq.InsertLoginAttemptParams{Login: "ion", At: testdb.Now()}); err != nil {
			return err
		}
		return q.QueryRow(ctx, "select count(*) from login_attempt").Scan(&after)
	})
	if err != nil {
		t.Fatal(err)
	}
	if iso != "repeatable read" || ro != "on" {
		t.Errorf("transaction %q, read only %q; want repeatable read, on", iso, ro)
	}
	if before != 0 || after != 0 {
		t.Errorf("counts %d then %d inside one snapshot, want 0 and 0", before, after)
	}

	// A write is refused; fn's error comes back as it is.
	err = store.InReadTx(ctx, func(q dbq.DBTX) error {
		_, err := dbq.New(q).InsertLoginAttempt(ctx, dbq.InsertLoginAttemptParams{Login: "ion", At: testdb.Now()})
		return err
	})
	if pe, ok := errors.AsType[*pgconn.PgError](err); !ok || pe.Code != "25006" { // read_only_sql_transaction
		t.Errorf("write in a read tx: %v, want 25006", err)
	}
	boom := errors.New("boom")
	if err := store.InReadTx(ctx, func(dbq.DBTX) error { return boom }); !errors.Is(err, boom) {
		t.Errorf("fn's error = %v, want boom", err)
	}
}

func TestDecimalCodec(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	var got decimal.Decimal
	if err := d.App.QueryRow(context.Background(), "select 168122::numeric / 186088 * 100").Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got.Round(1).String() != "90.3" {
		t.Fatalf("numeric → decimal = %s", got)
	}
}

func TestAuthQueries(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()
	q := dbq.New(d.App)
	now := testdb.Now()
	ip := netip.MustParseAddr("192.0.2.7")

	for i, success := range []bool{false, false, true} {
		if _, err := q.InsertLoginAttempt(ctx, dbq.InsertLoginAttemptParams{
			Login: "ion", Ip: &ip, At: now.Add(-10*time.Minute + time.Duration(i)*time.Second), Success: success,
		}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := q.LoginFailuresSince(ctx, dbq.LoginFailuresSinceParams{Login: "ion", Since: now.Add(-15 * time.Minute)})
	if err != nil || got.Failures != 2 || !got.Latest.Equal(now.Add(-10*time.Minute+time.Second)) {
		t.Fatalf("login failures = %+v, %v", got, err)
	}
	ipGot, err := q.IPFailuresSince(ctx, dbq.IPFailuresSinceParams{Ip: ip, Since: now.Add(-15 * time.Minute)})
	if err != nil || ipGot.Failures != 2 {
		t.Fatalf("ip failures = %+v, %v", ipGot, err)
	}
	none, err := q.LoginFailuresSince(ctx, dbq.LoginFailuresSinceParams{Login: "nobody", Since: now})
	if err != nil || none.Failures != 0 || !none.Latest.Equal(now) {
		t.Fatalf("no failures = %+v, %v", none, err)
	}

	user, err := insertUser(t, q, "ion.rusu")
	if err != nil {
		t.Fatal(err)
	}
	sid, err := q.InsertSession(ctx, dbq.InsertSessionParams{
		UserID: user.ID, TokenHash: []byte("hash"), Now: now, ExpiresAt: now.Add(30 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	s, err := q.GetActiveSession(ctx, dbq.GetActiveSessionParams{TokenHash: []byte("hash"), Now: now})
	if err != nil || s.SessionID != sid || s.Role != dbq.UserRoleManager {
		t.Fatalf("active session = %+v, %v", s, err)
	}
	// Expired at `now` (D5: the comparison uses the parameter, not SQL now()).
	if _, err := q.GetActiveSession(ctx, dbq.GetActiveSessionParams{TokenHash: []byte("hash"), Now: now.Add(31 * 24 * time.Hour)}); err == nil {
		t.Fatal("expired session found")
	}
	// Slides at most once per window.
	n, err := q.SlideSession(ctx, dbq.SlideSessionParams{ID: sid, Now: now.Add(30 * time.Minute), ExpiresAt: now.Add(31 * 24 * time.Hour), SeenBefore: now.Add(-30 * time.Minute)})
	if err != nil || n != 0 {
		t.Fatalf("slide within the hour = %d, %v", n, err)
	}
	n, err = q.SlideSession(ctx, dbq.SlideSessionParams{ID: sid, Now: now.Add(2 * time.Hour), ExpiresAt: now.Add(31 * 24 * time.Hour), SeenBefore: now.Add(time.Hour)})
	if err != nil || n != 1 {
		t.Fatalf("slide after the hour = %d, %v", n, err)
	}
	if n, err := q.RevokeUserSessions(ctx, dbq.RevokeUserSessionsParams{UserID: user.ID, Now: now}); err != nil || n != 1 {
		t.Fatalf("revoke all = %d, %v", n, err)
	}
	if _, err := q.GetActiveSession(ctx, dbq.GetActiveSessionParams{TokenHash: []byte("hash"), Now: now}); err == nil {
		t.Fatal("revoked session found")
	}
	if owners, err := q.CountActiveOwners(ctx); err != nil || owners != 0 {
		t.Fatalf("owners = %d, %v", owners, err)
	}
	if farms, err := q.ListFarms(ctx); err != nil || farms == nil || len(farms) != 0 {
		t.Fatalf("farms = %#v, %v (want an empty, non-nil slice)", farms, err)
	}
}

func TestOpenPools(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	cfg := config.Config{DatabaseURL: testdb.DSN(d.Name), MaintDatabaseURL: testdb.DSN(d.Name), DBMaxConns: 3}
	pools, err := db.Open(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pools.Close()
	if pools.App.Config().MaxConns != 3 || pools.Maint.Config().MaxConns != 1 {
		t.Fatalf("MaxConns app %d maint %d", pools.App.Config().MaxConns, pools.Maint.Config().MaxConns)
	}
	var tz string
	if err := pools.App.QueryRow(context.Background(), "show timezone").Scan(&tz); err != nil || tz != "UTC" {
		t.Fatalf("timezone = %q, %v", tz, err)
	}
	// Production pools leave app.today unset: app_today() follows the real clock.
	var setting string
	if err := pools.App.QueryRow(context.Background(), "select coalesce(current_setting('app.today', true), '')").Scan(&setting); err != nil || setting != "" {
		t.Fatalf("app.today = %q, %v", setting, err)
	}
	if _, err := db.Open(context.Background(), config.Config{DatabaseURL: "postgres://%zz", MaintDatabaseURL: "x"}); err == nil {
		t.Fatal("bad DSN accepted")
	}
}

func TestLatestVersion(t *testing.T) {
	v, err := db.LatestVersion()
	if err != nil || v != 2 {
		t.Fatalf("LatestVersion = %d, %v", v, err)
	}
}
