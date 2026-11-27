package testdb_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"slices"
	"sync"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// TestNewSeeded checks a seeded copy has New's pool, clock and app.today and
// holds the import plus the dev seed (§11.8, §11.9).
func TestNewSeeded(t *testing.T) {
	t.Parallel()
	d := testdb.NewSeeded(t)
	ctx := context.Background()

	var today timex.Day
	var user string
	if err := d.App.QueryRow(ctx, "select app_today(), current_user").Scan(&today, &user); err != nil {
		t.Fatal(err)
	}
	if user != testdb.RoleApp {
		t.Errorf("current_user = %q, want il_app", user)
	}
	if today != testdb.Today {
		t.Errorf("app_today() = %v, want 2026-11-18", today)
	}
	if !d.Clock.Now().Equal(testdb.Now()) {
		t.Errorf("clock now = %v", d.Clock.Now())
	}

	rows, err := d.App.Query(ctx, "select name from farm")
	if err != nil {
		t.Fatal(err)
	}
	var farms []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		farms = append(farms, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	slices.Sort(farms)
	if want := []string{"Hîrbovăț", "Telița"}; !slices.Equal(farms, want) {
		t.Errorf("farms = %q, want %q", farms, want)
	}

	var minAlive, users int
	if err := d.App.QueryRow(ctx, "select min(alive_end) from v_section_sex_day").Scan(&minAlive); err != nil {
		t.Fatal(err)
	}
	if minAlive != 0 {
		t.Errorf("min(alive_end) over v_section_sex_day = %d, want 0", minAlive)
	}
	if err := d.App.QueryRow(ctx, "select count(*) from app_user").Scan(&users); err != nil {
		t.Fatal(err)
	}
	if users != 4 {
		t.Errorf("app_user has %d rows, want the 4 dev users", users)
	}
}

// TestSeededNamesAreUniqueToTheProcess checks seeded copies are named like
// New's (il_t_<pid>_<random>_…) and never reuse a name.
func TestSeededNamesAreUniqueToTheProcess(t *testing.T) {
	t.Parallel()
	plain, a, b := testdb.New(t), testdb.NewSeeded(t), testdb.NewSeeded(t)
	ours := regexp.MustCompile(fmt.Sprintf(`^(il_t_%d_[0-9a-f]{8})_`, os.Getpid()))
	m := ours.FindStringSubmatch(plain.Name)
	if m == nil {
		t.Fatalf("New's name %q is not il_t_<pid>_<random>_…", plain.Name)
	}
	for _, d := range []*testdb.DB{a, b} {
		if p := ours.FindStringSubmatch(d.Name); p == nil || p[1] != m[1] {
			t.Errorf("seeded name %q does not share New's process prefix %q", d.Name, m[1])
		}
	}
	if a.Name == b.Name || a.Name == plain.Name || b.Name == plain.Name {
		t.Errorf("names repeat: %q %q %q", plain.Name, a.Name, b.Name)
	}
}

// TestSeededTemplateIsBuiltOnce asks for many seeded copies at once: the
// seeder runs once per process.
func TestSeededTemplateIsBuiltOnce(t *testing.T) {
	t.Run("copies", func(t *testing.T) {
		for i := range 4 {
			t.Run(fmt.Sprint(i), func(t *testing.T) {
				t.Parallel()
				testdb.NewSeeded(t)
			})
		}
	})
	if n := seedBuilds.Load(); n != 1 {
		t.Errorf("seeded template built %d times, want 1", n)
	}
}

// TestSeededCopiesAreIsolated writes to two seeded copies at the same time:
// neither sees the other's row, and a later copy sees neither (the template is
// untouched).
func TestSeededCopiesAreIsolated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	copies := map[string]*testdb.DB{"Test A": testdb.NewSeeded(t), "Test B": testdb.NewSeeded(t)}

	var wg sync.WaitGroup
	errs := make(chan error, len(copies))
	start := make(chan struct{})
	for name, d := range copies {
		wg.Go(func() {
			<-start
			_, err := d.App.Exec(ctx, "insert into farm (name, name_key, type) values ($1, $2, 'laying')", name, name)
			if err != nil {
				errs <- fmt.Errorf("insert %s: %w", name, err)
			}
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}

	testFarms := func(d *testdb.DB) []string {
		t.Helper()
		rows, err := d.App.Query(ctx, "select name from farm where name like 'Test %' order by name")
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				t.Fatal(err)
			}
			names = append(names, n)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		return names
	}
	for name, d := range copies {
		if got := testFarms(d); !slices.Equal(got, []string{name}) {
			t.Errorf("copy that wrote %q sees test farms %q", name, got)
		}
	}
	if got := testFarms(testdb.NewSeeded(t)); len(got) != 0 {
		t.Errorf("a fresh seeded copy sees test farms %q: the template was written", got)
	}
}
