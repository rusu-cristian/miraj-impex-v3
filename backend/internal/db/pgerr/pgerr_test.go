package pgerr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapError(t *testing.T) {
	for _, k := range []Kind{Unique, Exclusion, Check, ForeignKey} {
		pg := &pgconn.PgError{Code: string(k), ConstraintName: "c_" + string(k), TableName: "t"}
		wrapped := fmt.Errorf("insert: %w", pg)
		ce, ok := AsConstraint(wrapped, k)
		if !ok || ce.Kind != k || ce.Constraint != "c_"+string(k) || ce.Table != "t" {
			t.Errorf("%s: AsConstraint = %+v, %v", k, ce, ok)
		}
		if !errors.Is(ce, pg) {
			t.Errorf("%s: does not unwrap to the PgError", k)
		}
		if _, ok := AsConstraint(wrapped, other(k)); ok {
			t.Errorf("%s: matched kind %s", k, other(k))
		}
	}

	// Other SQLSTATEs and non-PostgreSQL errors pass through unchanged.
	serialization := &pgconn.PgError{Code: "40001"}
	plain := errors.New("plain")
	for _, err := range []error{serialization, plain, nil} {
		if got := MapError(err); got != err {
			t.Errorf("MapError(%v) = %v, want it unchanged", err, got)
		}
	}
	if _, ok := AsConstraint(serialization, Unique); ok {
		t.Error("a 40001 is not a constraint violation")
	}
}

func other(k Kind) Kind {
	if k == Unique {
		return Check
	}
	return Unique
}
