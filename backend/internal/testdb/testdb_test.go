package testdb_test

import (
	"context"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestMain(m *testing.M) { testdb.Main(m) }

func TestNew(t *testing.T) {
	t.Parallel()
	d := testdb.New(t)
	ctx := context.Background()

	var today timex.Day
	var user, tz string
	if err := d.App.QueryRow(ctx, "select app_today(), current_user, current_setting('timezone')").Scan(&today, &user, &tz); err != nil {
		t.Fatal(err)
	}
	if today != testdb.Today {
		t.Errorf("app_today() = %v, want 2026-11-18", today)
	}
	if user != testdb.RoleApp {
		t.Errorf("current_user = %q, want il_app", user)
	}
	if err := d.App.QueryRow(ctx, "show timezone").Scan(&tz); err != nil || tz != "UTC" {
		t.Errorf("show timezone = %q, %v", tz, err)
	}
	if got := timex.Today(d.Clock); got != testdb.Today {
		t.Errorf("clock today = %v", got)
	}
	if !d.Clock.Now().Equal(testdb.Now()) {
		t.Errorf("clock now = %v", d.Clock.Now())
	}

	// The template is fully migrated.
	latest, err := db.LatestVersion()
	if err != nil {
		t.Fatal(err)
	}
	var version int64
	if err := d.App.QueryRow(ctx, "select max(version_id) from goose_db_version").Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != latest {
		t.Errorf("goose version = %d, want %d", version, latest)
	}
}

func TestDatabasesAreIsolated(t *testing.T) {
	t.Parallel()
	a, b := testdb.New(t), testdb.New(t)
	if a.Name == b.Name {
		t.Fatalf("same name %s", a.Name)
	}
	ctx := context.Background()
	if _, err := a.App.Exec(ctx, "insert into login_attempt (login, at, success) values ('ion', now(), false)"); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := b.App.QueryRow(ctx, "select count(*) from login_attempt").Scan(&n); err != nil || n != 0 {
		t.Fatalf("other database sees %d rows (%v)", n, err)
	}
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	dsn := testdb.Empty(t)
	conn, err := db.OpenSQL(dsn, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var tables int
	if err := conn.QueryRow("select count(*) from pg_tables where schemaname = 'public'").Scan(&tables); err != nil || tables != 0 {
		t.Fatalf("empty database has %d tables (%v)", tables, err)
	}
}
