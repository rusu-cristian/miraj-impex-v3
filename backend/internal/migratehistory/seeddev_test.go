package migratehistory_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/migratehistory"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

const devPassword = "dev-test-pass"

func seeded(t *testing.T) (*testdb.DB, []migratehistory.TableCount, string) {
	t.Helper()
	d := testdb.New(t)
	o := opts()
	o.SeedDev, o.DevPassword = true, devPassword
	var out bytes.Buffer
	o.SeedOut = &out
	counts, err := migratehistory.Load(t.Context(), d.App, d.Clock, o)
	if err != nil {
		t.Fatal(err)
	}
	return d, counts, out.String()
}

// strings1 runs a query whose rows are one text column each.
func strings1(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) []string {
	t.Helper()
	rows, err := pool.Query(t.Context(), sql, args...)
	if err != nil {
		t.Fatal(err)
	}
	got, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func countSQL(t *testing.T, pool *pgxpool.Pool, sql string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(t.Context(), sql).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func local(s string) time.Time {
	t, err := time.ParseInLocation("2006-01-02 15:04", s, timex.Chisinau)
	if err != nil {
		panic(err)
	}
	return t
}

func TestSeedDev(t *testing.T) {
	d, counts, out := seeded(t)
	ctx := t.Context()

	// Verify passed inside Load; it still passes on the committed data, so
	// goldens 1–12, 26 and 27 are unchanged by the seed.
	if err := migratehistory.Verify(ctx, d.App); err != nil {
		t.Fatal(err)
	}
	byTable := map[string]int64{}
	for _, c := range counts {
		byTable[c.Table] = c.Rows
	}
	if byTable["app_user"] != 4 || byTable["user_farm"] != 5 || byTable["audit_log"] != 4 || byTable["bird_event"] != 4041+5 {
		t.Errorf("counts: %v", byTable)
	}

	t.Run("users", func(t *testing.T) {
		got := strings1(t, d.App, `
			select u.login || ' ' || u.name || ' ' || u.role || ' ' ||
			       coalesce((select string_agg(f.name, ',' order by f.name) from user_farm uf join farm f on f.id = uf.farm_id where uf.user_id = u.id), '-')
			from app_user u order by u.login`)
		want := []string{
			"dr.ceban Dr. Ceban vet Hîrbovăț",
			"ion.rusu Ion Rusu manager Hîrbovăț,Telița",
			"observator Observator exemplu viewer Hîrbovăț,Telița",
			"proprietar Proprietar owner -",
		}
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("users (-want +got):\n%s", diff)
		}
		hashes := strings1(t, d.App, `select password_hash from app_user`)
		for _, h := range hashes {
			if !strings.HasPrefix(h, "$argon2id$") || strings.Contains(h, devPassword) {
				t.Errorf("password_hash %q is not an argon2id hash", h)
			}
		}
		for _, login := range []string{"proprietar", "ion.rusu", "dr.ceban", "observator"} {
			if strings.Count(out, login+"\t"+devPassword+"\t") != 1 {
				t.Errorf("SeedOut lacks %s once:\n%s", login, out)
			}
		}
	})

	t.Run("culls carved out of the deaths", func(t *testing.T) {
		got := strings1(t, d.App, `
			select s.code || ' ' || c.date || ' ' || c.sex || ' cull ' || -c.qty || ' ' || c.cull_reason ||
			       ' mortality ' || -m.qty || ' ' || e.kind || ' ' || u.login || ' ' ||
			       to_char(e.created_at at time zone 'Europe/Chisinau', 'HH24:MI')
			from bird_event c
			join section s on s.id = c.section_id
			join entry e on e.id = c.entry_id
			join app_user u on u.id = e.created_by
			join bird_event m on m.section_id = c.section_id and m.date = c.date and m.sex = c.sex and m.type = 'mortality'
			where c.type = 'cull' order by c.date`)
		want := []string{
			"H6.1 2026-04-10 F cull 2 boala mortality 2 daily dr.ceban 08:20",
			"H5.1 2026-06-14 F cull 1 schiopatare mortality 2 daily dr.ceban 08:21",
			"H6.1 2026-08-10 M cull 2 boala mortality 2 daily dr.ceban 08:22",
			"H7.1 2026-10-19 F cull 1 altul mortality 2 daily dr.ceban 08:23",
			"H5.2 2026-11-11 M cull 1 schiopatare mortality 0 daily dr.ceban 08:31",
		}
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("culls (-want +got):\n%s", diff)
		}
		// The sexing error is not seeded (user answer Q-T7-1: it would change golden 10).
		if n := countSQL(t, d.App, "select count(*) from bird_event where type = 'sexing_error' and farm_id = (select id from farm where name_key = 'hirbovat')"); n != 0 {
			t.Errorf("Hîrbovăț sexing errors = %d, want 0 (not seeded, Q-T7-1)", n)
		}
	})

	t.Run("weighings", func(t *testing.T) {
		got := strings1(t, d.App, `
			select s.code || ' ' || w.sex || ' ' || w.avg_weight_g || ' ' || e.kind || ' ' || u.login || ' ' ||
			       to_char(e.created_at at time zone 'Europe/Chisinau', 'YYYY-MM-DD HH24:MI')
			from weighing w join section s on s.id = w.section_id join entry e on e.id = w.entry_id
			join app_user u on u.id = e.created_by
			where w.farm_id = (select id from farm where name_key = 'hirbovat') order by w.sex`)
		want := []string{"H5.1 F 3960.0 weighing dr.ceban 2026-11-10 09:15", "H5.1 M 4870.0 weighing dr.ceban 2026-11-10 09:32"}
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("weighings (-want +got):\n%s", diff)
		}
	})

	t.Run("event log entries", func(t *testing.T) {
		got := strings1(t, d.App, `
			select to_char(e.created_at at time zone 'Europe/Chisinau', 'YYYY-MM-DD HH24:MI') || ' ' || e.kind || ' ' ||
			       u.login || ' ' || e.source || ' ' ||
			       (select count(*) from bird_event b where b.entry_id = e.id) || '/' ||
			       (select count(*) from egg_production x where x.entry_id = e.id) || '/' ||
			       (select count(*) from feed_allocation x where x.entry_id = e.id) || '/' ||
			       (select count(*) from feed_delivery x where x.entry_id = e.id)
			from entry e join app_user u on u.id = e.created_by
			where e.kind in ('transfer', 'sale', 'delivery', 'feed_daily')
			   or (e.kind = 'daily' and e.date >= '2026-11-16')
			order by e.created_at`)
		want := []string{
			"2026-10-22 09:20 transfer ion.rusu manual 2/0/0/0",
			"2026-11-12 10:30 delivery ion.rusu manual 0/0/0/1",
			"2026-11-12 11:00 sale ion.rusu manual 1/0/0/0",
			"2026-11-13 10:30 delivery ion.rusu manual 0/0/0/1",
			"2026-11-14 12:17 sale ion.rusu manual 2/0/0/0",
			"2026-11-14 13:34 sale ion.rusu manual 4/0/0/0",
			"2026-11-15 12:08 sale ion.rusu manual 1/0/0/0",
			"2026-11-16 08:12 daily dr.ceban manual 8/0/0/0",
			"2026-11-16 08:25 daily dr.ceban manual 0/2/0/0",
			"2026-11-16 08:41 feed_daily dr.ceban manual 0/0/8/0",
			"2026-11-16 13:25 sale ion.rusu manual 2/0/0/0",
			"2026-11-17 08:07 daily dr.ceban manual 8/0/0/0",
			"2026-11-17 08:19 daily dr.ceban manual 0/2/0/0",
			"2026-11-17 08:36 feed_daily dr.ceban manual 0/0/8/0",
			"2026-11-17 11:42 sale ion.rusu manual 2/0/0/0",
			"2026-11-18 08:15 daily dr.ceban manual 8/0/0/0",
			"2026-11-18 08:27 daily dr.ceban manual 0/2/0/0",
			"2026-11-18 08:44 feed_daily dr.ceban manual 0/0/8/0",
			"2026-11-18 13:16 sale ion.rusu manual 4/0/0/0",
		}
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("authored entries (-want +got):\n%s", diff)
		}
		// The Telița links and moves stay migration entries without an author.
		if n := countSQL(t, d.App, "select count(*) from entry where kind = 'transfer' and created_by is null and source = 'migration'"); n != 21 {
			t.Errorf("unauthored transfer entries = %d, want 21", n)
		}
	})

	t.Run("correction", func(t *testing.T) {
		var corrected bool
		var updatedBy string
		var updatedAt time.Time
		err := d.App.QueryRow(ctx, `
			select e.corrected, u.login, e.updated_at from entry e join app_user u on u.id = e.updated_by
			where e.kind = 'daily' and e.date = '2026-11-17' and e.created_at = $1`, local("2026-11-17 08:07")).Scan(&corrected, &updatedBy, &updatedAt)
		if err != nil {
			t.Fatal(err)
		}
		if !corrected || updatedBy != "ion.rusu" || !updatedAt.Equal(local("2026-11-18 10:12")) {
			t.Errorf("17.11 entry: corrected %v by %s at %s", corrected, updatedBy, updatedAt)
		}
		got := strings1(t, d.App, `
			select a.action || ' ' || u.login || ' ' || to_char(a.at at time zone 'Europe/Chisinau', 'YYYY-MM-DD HH24:MI') || ' ' ||
			       a.entity || ' ' || a.entry_kind || ' ' || a.record_date || ' ' || array_to_string(a.record_types, ',') || ' ' ||
			       (a.farm_ids = array[(select id from farm where name_key = 'hirbovat')]) || ' ' ||
			       coalesce(a.reason, '-') || ' ' || a.changes::text
			from audit_log a join app_user u on u.id = a.user_id
			join entry e on e.id = a.entry_id and e.created_at = $1
			order by a.at`, local("2026-11-17 08:07"))
		want := []string{
			"create dr.ceban 2026-11-17 08:07 entry daily 2026-11-17 mortalitate true - []",
			`update ion.rusu 2026-11-18 10:12 entry daily 2026-11-17 mortalitate true Numărătoare refăcută: 2 găini găsite moarte sub hrănitoare. ` +
				`[{"key": "H5.2 · găini moarte", "new": 3, "old": 1, "sex": "F", "field": "mortality", "section": "H5.2"}]`,
		}
		if diff := gocmp.Diff(want, got); diff != "" {
			t.Errorf("correction audit rows (-want +got):\n%s", diff)
		}
	})
}

func TestSeedDevGeneratedPasswords(t *testing.T) {
	d := testdb.New(t)
	o := opts()
	o.SeedDev = true
	var out bytes.Buffer
	o.SeedOut = &out
	if _, err := migratehistory.Load(t.Context(), d.App, d.Clock, o); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n")[1:] {
		f := strings.Split(line, "\t")
		if len(f) != 3 || len(f[1]) != 12 || seen[f[1]] {
			t.Errorf("credential line %q: want login, a fresh 12-character password, name", line)
		}
		seen[f[1]] = true
	}
	if len(seen) != 4 {
		t.Errorf("%d passwords printed, want 4:\n%s", len(seen), out.String())
	}
}

func TestSeedDevNothingPrintedOnFailure(t *testing.T) {
	d := testdb.New(t)
	o := opts()
	o.SeedDev = true
	var out bytes.Buffer
	o.SeedOut = &out
	failing := migratehistory.Check{ID: "test", Name: "fail", Run: func(context.Context, dbq.DBTX) (any, any, error) { return 1, 2, nil }}
	if _, err := migratehistory.Load(t.Context(), d.App, d.Clock, migratehistory.WithCheck(o, failing)); err == nil {
		t.Fatal("want the injected check to fail")
	}
	if out.Len() != 0 || count(t, d.App, "app_user") != 0 {
		t.Errorf("a rolled-back seed printed %q and left %d users", out.String(), count(t, d.App, "app_user"))
	}
}
