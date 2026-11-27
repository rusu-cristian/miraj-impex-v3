package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// ErrRollback, returned by an InTx callback, rolls the transaction back and
// makes InTx return nil (a dry run, SPEC §7.0, §12.5).
var ErrRollback = errors.New("rollback")

// Tx is a transaction and the generated queries bound to it.
type Tx struct {
	pgx.Tx
	Q *dbq.Queries
}

// Store runs queries and transactions on the app pool.
type Store struct {
	pool *pgxpool.Pool
	q    *dbq.Queries
}

// NewStore wraps pool.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, q: dbq.New(pool)}
}

// Pool returns the underlying pool.
func (s *Store) Pool() *pgxpool.Pool { return s.pool }

// Q returns the generated queries outside a transaction (single reads and
// single-statement writes).
func (s *Store) Q() *dbq.Queries { return s.q }

// InTx runs fn in one READ COMMITTED transaction (SPEC §12.1). fn's error
// rolls back; ErrRollback rolls back and returns nil. Constraint violations
// come back as *ConstraintError (MapError).
func (s *Store) InTx(ctx context.Context, fn func(Tx) error) error {
	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(t pgx.Tx) error {
		return fn(Tx{Tx: t, Q: dbq.New(t)})
	})
	if errors.Is(err, ErrRollback) {
		return nil
	}
	return MapError(err)
}
