package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	migrations "github.com/rusu-cristian/miraj-impex-v3/backend/db"
)

// MigrationsFS is the embedded db/migrations directory, rooted so goose sees
// 0001_baseline.sql, 0002_roles.sql, … at its top level.
func MigrationsFS() fs.FS {
	sub, err := fs.Sub(migrations.Migrations, "migrations")
	if err != nil {
		panic("db: embedded migrations: " + err.Error()) // the embed pattern guarantees the directory
	}
	return sub
}

// OpenSQL opens a database/sql handle (pgx stdlib driver) for goose. The
// optional afterConnect runs on every new connection (testdb uses it to SET
// ROLE). Errors never include the DSN.
func OpenSQL(dsn string, afterConnect func(context.Context, *pgx.Conn) error) (*sql.DB, error) {
	cc, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db: cannot parse the connection string")
	}
	cc.RuntimeParams["timezone"] = "UTC"
	var opts []stdlib.OptionOpenDB
	if afterConnect != nil {
		opts = append(opts, stdlib.OptionAfterConnect(afterConnect))
	}
	return stdlib.OpenDB(*cc, opts...), nil
}

// NewMigrator returns a goose provider over the embedded migrations. It holds
// a PostgreSQL session advisory lock (pg_try_advisory_lock, retried every
// second for up to 5 minutes) while it migrates, so two concurrent
// `api migrate up` runs apply each file once.
func NewMigrator(sqlDB *sql.DB) (*goose.Provider, error) {
	locker, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(1, 300))
	if err != nil {
		return nil, fmt.Errorf("db: goose locker: %w", err)
	}
	p, err := goose.NewProvider(goose.DialectPostgres, sqlDB, MigrationsFS(), goose.WithSessionLocker(locker))
	if err != nil {
		return nil, fmt.Errorf("db: goose provider: %w", err)
	}
	return p, nil
}

// LatestVersion is the newest embedded migration version (/readyz compares it
// with the database's goose version).
func LatestVersion() (int64, error) {
	names, err := fs.Glob(MigrationsFS(), "*.sql")
	if err != nil {
		return 0, fmt.Errorf("db: list migrations: %w", err)
	}
	var latest int64
	for _, name := range names {
		v, err := goose.NumericComponent(name)
		if err != nil {
			return 0, fmt.Errorf("db: migration %s: %w", name, err)
		}
		latest = max(latest, v)
	}
	if latest == 0 {
		return 0, fmt.Errorf("db: no embedded migrations")
	}
	return latest, nil
}
