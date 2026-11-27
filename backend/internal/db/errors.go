package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Kind is the class of a constraint violation.
type Kind string

// The SQLSTATE codes the Go code maps (SPEC §7.0a, §12.4).
const (
	Unique     Kind = "23505" // unique_violation: ALREADY_ENTERED, name clashes
	Exclusion  Kind = "23P01" // exclusion_violation: cycle_no_overlap
	Check      Kind = "23514" // check_violation (incl. cycle_guide_type_trg)
	ForeignKey Kind = "23503" // foreign_key_violation
)

// ConstraintError is a constraint violation with the constraint's name, so
// services can map it to the §7.0 envelope (e.g. app_user_login_uq).
type ConstraintError struct {
	Kind       Kind
	Constraint string
	Table      string
	Err        error // the *pgconn.PgError
}

func (e *ConstraintError) Error() string {
	return "db: constraint violation " + string(e.Kind) + " on " + e.Constraint
}

func (e *ConstraintError) Unwrap() error { return e.Err }

// MapError turns a PostgreSQL unique, exclusion, check or foreign-key
// violation into a *ConstraintError and returns any other error unchanged.
func MapError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch k := Kind(pgErr.Code); k {
	case Unique, Exclusion, Check, ForeignKey:
		return &ConstraintError{Kind: k, Constraint: pgErr.ConstraintName, Table: pgErr.TableName, Err: err}
	default:
		return err
	}
}

// AsConstraint reports whether err is a constraint violation of kind k and
// returns it.
func AsConstraint(err error, k Kind) (*ConstraintError, bool) {
	var ce *ConstraintError
	if errors.As(MapError(err), &ce) && ce.Kind == k {
		return ce, true
	}
	return nil, false
}
