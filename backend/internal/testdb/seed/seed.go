package seed

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/migratehistory"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// Main is testdb.Main plus the seeded template, built on the first
// testdb.NewSeeded of the process.
func Main(m *testing.M) { testdb.MainSeeded(m, Load) }

// Load is the testdb.Seeder: the history import with the dev seed, every
// seeded user's password testdb.DevPassword, nothing printed.
func Load(ctx context.Context, pool *pgxpool.Pool, clock timex.Clock, extracts string) error {
	_, err := migratehistory.Load(ctx, pool, clock, migratehistory.Options{
		Extracts:    extracts,
		SeedDev:     true,
		DevPassword: testdb.DevPassword,
	})
	return err
}
