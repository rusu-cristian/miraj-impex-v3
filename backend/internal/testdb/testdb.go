// Package testdb gives integration tests a real PostgreSQL 16 (SPEC §11.8).
//
// A test package calls Main from its TestMain. Main uses the server named by
// TEST_DATABASE_URL, or starts one postgres:16 container for the package's
// test process (testcontainers). It creates the roles il_owner, il_app and
// il_maint if they are missing, applies the embedded goose migrations as
// il_owner to a template database, and runs the tests. New(t) then gives each
// test its own copy of the template, an il_app pool with app.today =
// 2026-11-18 and a clock frozen at 2026-11-18T12:00 Europe/Chisinau.
//
// TEST_DATABASE_URL must name a superuser (e.g. the compose db's postgres
// user): testdb creates databases and switches roles with SET ROLE, so it
// never needs the roles' passwords and never changes them. Every database it
// creates has a name unique to the process (il_t_<pid>_<random>_…), so test
// packages running at the same time against one server don't collide.
package testdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Image is the PostgreSQL image testcontainers starts (the compose db's tag).
const Image = "postgres:16.16-bookworm"

// The roles of SPEC §11.9.
const (
	RoleOwner = "il_owner"
	RoleApp   = "il_app"
	RoleMaint = "il_maint"
)

// Today is the frozen business day of every test (SPEC §11.7).
var Today = timex.MustDay(2026, time.November, 18)

// Now is the frozen instant of every test clock: 2026-11-18T12:00 local.
func Now() time.Time { return time.Date(2026, time.November, 18, 12, 0, 0, 0, timex.Chisinau) }

// roleLockID serialises role creation across test processes sharing a server.
const roleLockID = 7_402_118_000_001

var (
	adminDSN string // superuser DSN of the server (its own database)
	prefix   string // il_t_<pid>_<random>: unique to this process
	template string // the migrated template database
	counter  atomic.Int64
	createMu sync.Mutex // one CREATE DATABASE … TEMPLATE at a time in this process
)

// Main prepares the server and the template, runs the tests and cleans up.
// Use it as: func TestMain(m *testing.M) { testdb.Main(m) }.
func Main(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	stop, err := setup(ctx)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "testdb:", err)
		if stop != nil {
			stop()
		}
		return 1
	}
	code := m.Run()
	stop()
	return code
}

// setup returns a cleanup func even on error, when there is something to clean.
func setup(ctx context.Context) (func(), error) {
	var stops []func()
	stop := func() {
		for i := len(stops) - 1; i >= 0; i-- {
			stops[i]()
		}
	}

	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		adminDSN = dsn
	} else {
		ctr, err := tcpostgres.Run(ctx, Image, tcpostgres.BasicWaitStrategies())
		if ctr != nil {
			stops = append(stops, func() { _ = testcontainers.TerminateContainer(ctr) })
		}
		if err != nil {
			return stop, fmt.Errorf("start %s: %w", Image, err)
		}
		dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			return stop, fmt.Errorf("container DSN: %w", err)
		}
		adminDSN = dsn
	}

	suffix := make([]byte, 4)
	_, _ = rand.Read(suffix)
	prefix = fmt.Sprintf("il_t_%d_%s", os.Getpid(), hex.EncodeToString(suffix))
	template = prefix + "_tpl"

	if err := ensureRoles(ctx); err != nil {
		return stop, err
	}
	if err := createDatabase(ctx, template, ""); err != nil {
		return stop, err
	}
	stops = append(stops, func() { _ = dropDatabase(context.Background(), template) })
	if err := migrate(ctx, template); err != nil {
		return stop, err
	}
	return stop, nil
}

func adminConn(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		return nil, fmt.Errorf("connect to the test server: %w", err)
	}
	return conn, nil
}

// ensureRoles checks the admin is a superuser and creates the three roles
// (NOLOGIN: tests reach them with SET ROLE) unless they exist already.
func ensureRoles(ctx context.Context) error {
	conn, err := adminConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())

	var super bool
	if err := conn.QueryRow(ctx, "select rolsuper from pg_roles where rolname = current_user").Scan(&super); err != nil {
		return fmt.Errorf("check superuser: %w", err)
	}
	if !super {
		return errors.New("TEST_DATABASE_URL must connect as a superuser (testdb creates databases and uses SET ROLE)")
	}
	if _, err := conn.Exec(ctx, "select pg_advisory_lock($1)", int64(roleLockID)); err != nil {
		return fmt.Errorf("lock: %w", err)
	}
	defer func() { _, _ = conn.Exec(context.Background(), "select pg_advisory_unlock($1)", int64(roleLockID)) }()
	for _, role := range []string{RoleOwner, RoleApp, RoleMaint} {
		var exists bool
		if err := conn.QueryRow(ctx, "select exists (select 1 from pg_roles where rolname = $1)", role).Scan(&exists); err != nil {
			return fmt.Errorf("look up role %s: %w", role, err)
		}
		if !exists {
			if _, err := conn.Exec(ctx, "create role "+pgx.Identifier{role}.Sanitize()+" nologin"); err != nil {
				return fmt.Errorf("create role %s: %w", role, err)
			}
		}
	}
	return nil
}

// createDatabase creates name owned by il_owner, as a copy of from when from
// is not empty.
func createDatabase(ctx context.Context, name, from string) error {
	conn, err := adminConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	sql := "create database " + pgx.Identifier{name}.Sanitize() + " owner " + pgx.Identifier{RoleOwner}.Sanitize()
	if from != "" {
		sql += " template " + pgx.Identifier{from}.Sanitize()
	}
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	return nil
}

func dropDatabase(ctx context.Context, name string) error {
	conn, err := adminConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	_, err = conn.Exec(ctx, "drop database if exists "+pgx.Identifier{name}.Sanitize()+" with (force)")
	return err
}

// migrate applies the embedded migrations to name as il_owner, with the same
// goose provider as `api migrate up`.
func migrate(ctx context.Context, name string) error {
	sqlDB, err := db.OpenSQL(DSN(name), setRole(RoleOwner))
	if err != nil {
		return err
	}
	p, err := db.NewMigrator(sqlDB)
	if err != nil {
		_ = sqlDB.Close()
		return err
	}
	defer p.Close()
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrate %s: %w", name, err)
	}
	return nil
}

func setRole(role string) func(context.Context, *pgx.Conn) error {
	return func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "set role "+pgx.Identifier{role}.Sanitize())
		return err
	}
}

// DSN is the superuser DSN of the test server pointed at database name.
func DSN(name string) string {
	if strings.Contains(adminDSN, "://") {
		u, err := url.Parse(adminDSN)
		if err == nil {
			u.Path = "/" + name
			return u.String()
		}
	}
	return adminDSN + " dbname=" + name
}

func nextName(kind string) string {
	return fmt.Sprintf("%s_%s%d", prefix, kind, counter.Add(1))
}

// DB is one test's database.
type DB struct {
	Name  string
	App   *pgxpool.Pool // role il_app, app.today = 2026-11-18
	Clock *timex.Fixed  // 2026-11-18T12:00 Europe/Chisinau
	t     testing.TB
}

// New returns a fresh copy of the migrated template, dropped when the test ends.
func New(t testing.TB) *DB {
	t.Helper()
	requireMain(t)
	name := nextName("")
	createMu.Lock()
	err := createDatabase(context.Background(), name, template)
	createMu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := dropDatabase(context.Background(), name); err != nil {
			t.Errorf("testdb: drop %s: %v", name, err)
		}
	})
	d := &DB{Name: name, Clock: timex.FixedClock(Now()), t: t}
	d.App = d.Pool(RoleApp)
	return d
}

// Pool returns a pool on this database that acts as role (SET ROLE) with
// app.today = 2026-11-18. It is closed when the test ends, before the
// database is dropped.
func (d *DB) Pool(role string) *pgxpool.Pool {
	d.t.Helper()
	pc, err := db.PoolConfig(DSN(d.Name), 4, &Today)
	if err != nil {
		d.t.Fatal(err)
	}
	prepare := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if err := setRole(role)(ctx, conn); err != nil {
			return err
		}
		return prepare(ctx, conn)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		d.t.Fatal(err)
	}
	d.t.Cleanup(pool.Close)
	return pool
}

// Empty returns the superuser DSN of a new empty database (no migrations), for
// tests of `api migrate`. It is dropped when the test ends.
func Empty(t testing.TB) string {
	t.Helper()
	requireMain(t)
	name := nextName("empty")
	if err := createDatabase(context.Background(), name, ""); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := dropDatabase(context.Background(), name); err != nil {
			t.Errorf("testdb: drop %s: %v", name, err)
		}
	})
	return DSN(name)
}

func requireMain(t testing.TB) {
	if adminDSN == "" || template == "" {
		t.Fatal("testdb: call testdb.Main(m) from the package's TestMain")
	}
}
