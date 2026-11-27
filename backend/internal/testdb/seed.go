package testdb

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// DevPassword is the password of every user of the dev seed in tests (§11.9).
const DevPassword = "dev-seed-pass"

// A Seeder fills the seeded template: migratehistory.Load with the dev seed
// (SPEC §11.8, plan M1-D6). pool acts as il_app on the template, clock is
// frozen at Now, and extracts is ExtractsDir().
//
// testdb can't import migratehistory itself: migratehistory imports auth and
// calc, whose in-package tests import testdb, so Go would refuse the import
// cycle. Package seed (internal/testdb/seed) supplies the real Seeder.
type Seeder func(ctx context.Context, pool *pgxpool.Pool, clock timex.Clock, extracts string) error

// seeder is the Seeder MainSeeded registered; nil under Main.
var seeder Seeder

// MainSeeded is Main for test packages that call NewSeeded: seed builds the
// seeded template, lazily, on the first NewSeeded of the process. Packages
// use it through seed.Main.
func MainSeeded(m *testing.M, seed Seeder) {
	seeder = seed
	os.Exit(run(m))
}

// seedOnce builds the seeded template at most once per process; every
// NewSeeded after a failure reports the same error.
var seedOnce = sync.OnceValue(func() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	return buildSeed(ctx)
})

// NewSeeded returns a fresh copy of the seeded template (il_seed): the
// migrated template plus the imported history and the dev seed of §11.9. It
// has the same pools, clock and app.today as New and is dropped when the test
// ends. The package's TestMain must call seed.Main (or MainSeeded).
func NewSeeded(t testing.TB) *DB {
	t.Helper()
	requireMain(t)
	if seeder == nil {
		t.Fatal("testdb: NewSeeded needs the package's TestMain to call seed.Main(m) instead of testdb.Main(m)")
	}
	if err := seedOnce(); err != nil {
		t.Fatal(err)
	}
	return copyOf(t, nextName("s"), seedName)
}

// buildSeed copies the migrated template to seedName and runs the seeder on
// it as il_app. Every connection to seedName is closed before it returns, so
// the database can serve as a template.
func buildSeed(ctx context.Context) error {
	extracts, err := ExtractsDir()
	if err != nil {
		return err
	}
	createMu.Lock()
	err = createDatabase(ctx, seedName, template)
	createMu.Unlock()
	if err != nil {
		return err
	}
	pool, err := openPool(ctx, seedName, RoleApp)
	if err != nil {
		return err
	}
	err = seeder(ctx, pool, timex.FixedClock(Now()), extracts)
	pool.Close()
	if err != nil {
		return fmt.Errorf("testdb: seed %s: %w", seedName, err)
	}
	return nil
}

// ExtractsDir is design/data/extracts (plan M1-D6): $IMPEX_EXTRACTS_DIR, else
// found from this source file (go:embed can't reach outside backend/).
func ExtractsDir() (string, error) {
	if d := os.Getenv("IMPEX_EXTRACTS_DIR"); d != "" {
		return d, nil
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("testdb: can't locate the extracts: set IMPEX_EXTRACTS_DIR")
	}
	dir := filepath.Join(filepath.Dir(file), "..", "..", "..", "design", "data", "extracts")
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("testdb: extracts directory: %w (set IMPEX_EXTRACTS_DIR)", err)
	}
	return dir, nil
}
