// Package db opens the PostgreSQL pools, runs transactions, maps constraint
// errors and applies the goose migrations (SPEC §11.9, §12.3, §12.4).
// Generated queries live in the dbq subpackage (sqlc output, never edited).
package db

import (
	"context"
	"errors"
	"fmt"

	pgxdecimal "github.com/jackc/pgx-shopspring-decimal"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/config"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Pools are the two pools of `api serve`: the app pool (role il_app) and the
// 1-connection maintenance pool (role il_maint) for the purge (SPEC §9.1).
type Pools struct {
	App   *pgxpool.Pool
	Maint *pgxpool.Pool
}

// Close closes both pools.
func (p *Pools) Close() {
	if p.App != nil {
		p.App.Close()
	}
	if p.Maint != nil {
		p.Maint.Close()
	}
}

// Open builds the app pool from DATABASE_URL with DB_MAX_CONNS connections and
// the maintenance pool from MAINT_DATABASE_URL with one connection. It does not
// wait for the database: readiness is /readyz's job.
func Open(ctx context.Context, cfg config.Config) (*Pools, error) {
	appCfg, err := PoolConfig(cfg.DatabaseURL, cfg.DBMaxConns, nil)
	if err != nil {
		return nil, fmt.Errorf("db: DATABASE_URL: %w", err)
	}
	maintCfg, err := PoolConfig(cfg.MaintDatabaseURL, 1, nil)
	if err != nil {
		return nil, fmt.Errorf("db: MAINT_DATABASE_URL: %w", err)
	}
	app, err := pgxpool.NewWithConfig(ctx, appCfg)
	if err != nil {
		return nil, fmt.Errorf("db: app pool: %w", err)
	}
	maint, err := pgxpool.NewWithConfig(ctx, maintCfg)
	if err != nil {
		app.Close()
		return nil, fmt.Errorf("db: maintenance pool: %w", err)
	}
	return &Pools{App: app, Maint: maint}, nil
}

// PoolConfig parses dsn and sets what every pool needs: maxConns, the session
// time zone UTC, the shopspring decimal codec for numeric, and, when today is
// not nil (tests only), app.today on every connection so SQL's app_today()
// agrees with the frozen Go clock (SPEC §11.1).
//
// Errors never include the DSN (it may carry a password).
func PoolConfig(dsn string, maxConns int32, today *timex.Day) (*pgxpool.Config, error) {
	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("cannot parse the connection string")
	}
	if maxConns > 0 {
		pc.MaxConns = maxConns
	}
	pc.ConnConfig.RuntimeParams["timezone"] = "UTC"
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return prepareConn(ctx, conn, today)
	}
	return pc, nil
}

// prepareConn registers the decimal codec and, in tests, sets app.today.
func prepareConn(ctx context.Context, conn *pgx.Conn, today *timex.Day) error {
	pgxdecimal.Register(conn.TypeMap())
	if today != nil {
		if _, err := conn.Exec(ctx, "select set_config('app.today', $1, false)", today.String()); err != nil {
			return fmt.Errorf("db: set app.today: %w", err)
		}
	}
	return nil
}
