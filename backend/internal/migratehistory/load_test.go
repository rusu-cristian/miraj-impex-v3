package migratehistory_test

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/migratehistory"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// The DB tests live in an external test package: testdb's seeded template
// (M1 T8) imports migratehistory.
func TestMain(m *testing.M) { testdb.Main(m) }

func opts() migratehistory.Options {
	return migratehistory.Options{Extracts: migratehistory.ExtractsDir()}
}

func count(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), "select count(*) from "+pgx.Identifier{table}.Sanitize()).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

var wantCounts = []migratehistory.TableCount{
	{Table: "guide", Rows: 2}, {Table: "guide_week", Rows: 144}, {Table: "farm", Rows: 2}, {Table: "section", Rows: 15},
	{Table: "feed_type", Rows: 5}, {Table: "series", Rows: 2}, {Table: "cycle", Rows: 3}, {Table: "entry", Rows: 31},
	{Table: "bird_event", Rows: 4041}, {Table: "egg_production", Rows: 1645}, {Table: "feed_allocation", Rows: 3900},
	{Table: "feed_delivery", Rows: 82}, {Table: "feed_stock_adjustment", Rows: 3}, {Table: "weighing", Rows: 98},
	{Table: "legacy_mortality", Rows: 3549}, {Table: "audit_log", Rows: 2},
}

func TestLoad(t *testing.T) {
	d := testdb.New(t)
	ctx := t.Context()
	counts, err := migratehistory.Load(ctx, d.App, d.Clock, opts())
	if err != nil {
		t.Fatal(err)
	}
	if diff := gocmp.Diff(wantCounts, counts); diff != "" {
		t.Errorf("counts (-want +got):\n%s", diff)
	}
	for _, c := range wantCounts {
		if got := count(t, d.App, c.Table); got != int(c.Rows) {
			t.Errorf("%s holds %d rows, want %d", c.Table, got, c.Rows)
		}
	}
	// The committed data passes Verify again outside the transaction: goldens
	// 1–28 and 38 (M1–M3), the §10.3 checks and alive_end ≥ 0.
	if err := migratehistory.Verify(ctx, d.App); err != nil {
		t.Fatal(err)
	}
	if n := len(migratehistory.Checks()); n != 29+3 {
		t.Errorf("Verify runs %d checks, want 32 (14 M1, 14 M2 and 1 M3 goldens, 2 §10.3, alive ≥ 0)", n)
	}

	// M1-D2: timestamps from the clock, no author.
	var wrongTime, withAuthor int
	if err := d.App.QueryRow(ctx, `select count(*) filter (where created_at <> $1), count(*) filter (where created_by is not null) from entry`,
		d.Clock.Now()).Scan(&wrongTime, &withAuthor); err != nil {
		t.Fatal(err)
	}
	if wrongTime != 0 || withAuthor != 0 {
		t.Errorf("entries: %d with another created_at, %d with an author", wrongTime, withAuthor)
	}

	// M1-D3: one create row per farm on its migration entry, counts in new_values.
	rows, err := d.App.Query(ctx, `
		select f.name, a.action, a.entity, a.entry_kind, a.user_id is null, a.summary is null, a.at = $1,
		       a.entry_id = a.entity_id and e.kind = 'migration' and e.farm_id = f.id, a.farm_ids = array[f.id],
		       a.record_date = e.date, a.source, a.new_values
		from audit_log a join entry e on e.id = a.entry_id join farm f on f.id = a.farm_ids[1]
		order by f.name`, d.Clock.Now())
	if err != nil {
		t.Fatal(err)
	}
	type audit struct {
		Farm                                       string
		Action                                     dbq.AuditAction
		Entity                                     string
		Kind                                       dbq.EntryKind
		System, NoSummary, ClockAt, OnEntry, Scope bool
		RecordDate                                 bool
		Source                                     dbq.RecordSource
		Values                                     map[string]any
	}
	got, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (audit, error) {
		var a audit
		var raw []byte
		err := r.Scan(&a.Farm, &a.Action, &a.Entity, &a.Kind, &a.System, &a.NoSummary, &a.ClockAt, &a.OnEntry, &a.Scope,
			&a.RecordDate, &a.Source, &raw)
		if err == nil {
			err = json.Unmarshal(raw, &a.Values)
		}
		return a, err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("audit rows = %d, want 2", len(got))
	}
	for _, a := range got {
		if a.Action != dbq.AuditActionCreate || a.Entity != "entry" || a.Kind != dbq.EntryKindMigration || !a.System ||
			!a.NoSummary || !a.ClockAt || !a.OnEntry || !a.Scope || !a.RecordDate || a.Source != dbq.RecordSourceMigration {
			t.Errorf("audit row of %s: %+v", a.Farm, a)
		}
	}
	hir := got[0] // "Hîrbovăț" sorts before "Telița"
	wantHir := map[string]any{
		"farm": "Hîrbovăț",
		// bird_event: 3 198 mortality rows, 16 link arrivals (one per source ×
		// destination: 7 ♂ + 9 ♀), the 22.10 pair and 16 sale rows.
		"rows": map[string]any{
			"section": 9.0, "cycle": 2.0, "entry": 14.0, "bird_event": 3198.0 + 16 + 2 + 16, "egg_production": 1645.0,
			"feed_allocation": 3184.0, "feed_delivery": 69.0, "feed_stock_adjustment": 3.0, "legacy_mortality": 3549.0,
		},
		"entries": map[string]any{"migration": 1.0, "transfer": 6.0, "sale": 7.0},
	}
	if diff := gocmp.Diff(wantHir, hir.Values); diff != "" {
		t.Errorf("Hîrbovăț audit new_values (-want +got):\n%s", diff)
	}
}

func TestLoadRefusesASecondRun(t *testing.T) {
	d := testdb.New(t)
	if _, err := migratehistory.Load(t.Context(), d.App, d.Clock, opts()); err != nil {
		t.Fatal(err)
	}
	_, err := migratehistory.Load(t.Context(), d.App, d.Clock, opts())
	if !errors.Is(err, migratehistory.ErrAlreadyMigrated) {
		t.Fatalf("second load: err = %v, want ErrAlreadyMigrated", err)
	}
	if n := count(t, d.App, "entry"); n != 31 {
		t.Errorf("entries after the refused run = %d, want 31", n)
	}
}

func TestLoadRollsBackOnAFailedCheck(t *testing.T) {
	d := testdb.New(t)
	failing := migratehistory.Check{ID: "test", Name: "injected", Run: func(ctx context.Context, db dbq.DBTX) (any, any, error) {
		// The import's rows are visible inside the transaction.
		var n int
		err := db.QueryRow(ctx, "select count(*) from entry").Scan(&n)
		return n, 0, err
	}}
	_, err := migratehistory.Load(t.Context(), d.App, d.Clock, migratehistory.WithCheck(opts(), failing))
	ve, ok := errors.AsType[*migratehistory.VerifyError](err)
	if !ok {
		t.Fatalf("err = %v, want *VerifyError", err)
	}
	if len(ve.Failures) != 1 || !strings.Contains(err.Error(), "test injected: got 31 want 0") {
		t.Errorf("failures = %v\nerror:\n%s", ve.Failures, err)
	}
	for _, table := range []string{"entry", "farm", "guide", "audit_log", "bird_event"} {
		if n := count(t, d.App, table); n != 0 {
			t.Errorf("%s holds %d rows after the rollback, want 0", table, n)
		}
	}
}

func TestLoadReplace(t *testing.T) {
	d := testdb.New(t)
	ctx := t.Context()
	if _, err := migratehistory.Load(ctx, d.App, d.Clock, opts()); err != nil {
		t.Fatal(err)
	}
	settings := count(t, d.App, "app_setting")

	// il_app may not empty audit_log: refused, nothing changes.
	replace := opts()
	replace.Replace = true
	if _, err := migratehistory.Load(ctx, d.App, d.Clock, replace); !errors.Is(err, migratehistory.ErrReplaceNeedsOwner) {
		t.Fatalf("replace as il_app: err = %v, want ErrReplaceNeedsOwner", err)
	}

	// As the owner it empties every domain table and loads again.
	owner := d.Pool(testdb.RoleOwner)
	if _, err := owner.Exec(ctx, "insert into app_user (name, login, password_hash, role, password_generated_at) values ('X', 'x.user', 'h', 'owner', now())"); err != nil {
		t.Fatal(err)
	}
	counts, err := migratehistory.Load(ctx, owner, d.Clock, replace)
	if err != nil {
		t.Fatal(err)
	}
	if diff := gocmp.Diff(wantCounts, counts); diff != "" {
		t.Errorf("counts (-want +got):\n%s", diff)
	}
	for table, want := range map[string]int{"entry": 31, "audit_log": 2, "app_user": 0, "app_setting": settings} {
		if n := count(t, d.App, table); n != want {
			t.Errorf("%s holds %d rows after --replace, want %d", table, n, want)
		}
	}
	var migrations int
	if err := owner.QueryRow(ctx, "select count(*) from goose_db_version").Scan(&migrations); err != nil || migrations == 0 {
		t.Errorf("goose_db_version emptied (%d rows, %v)", migrations, err)
	}
}

// noDB fails the test if Load ever starts a transaction.
type noDB struct{ t *testing.T }

func (n noDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	n.t.Error("Load touched the database")
	return nil, errors.New("no database")
}

func TestLoadBadExtractsTouchesNoDatabase(t *testing.T) {
	_, err := migratehistory.Load(t.Context(), noDB{t}, timex.FixedClock(testdb.Now()), migratehistory.Options{Extracts: filepath.Join(t.TempDir(), "nope")})
	if err == nil || !strings.Contains(err.Error(), "extracts directory") {
		t.Errorf("err = %v, want an extracts directory error", err)
	}
}
