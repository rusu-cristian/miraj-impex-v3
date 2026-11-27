package migratehistory

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
)

// Check is one verification of the imported data: Run reads what it checks
// and returns it with the value it must equal.
type Check struct {
	ID   string
	Name string
	Run  func(ctx context.Context, db dbq.DBTX) (got, want any, err error)
}

// Failure is a check that didn't pass: its values differ, or Err says why it
// couldn't be read.
type Failure struct {
	ID, Name  string
	Got, Want any
	Err       error
}

func (f Failure) String() string {
	if f.Err != nil {
		return fmt.Sprintf("%s %s: %v", f.ID, f.Name, f.Err)
	}
	return fmt.Sprintf("%s %s: got %+v want %+v", f.ID, f.Name, f.Got, f.Want)
}

// VerifyError lists every check that failed, one per line as "id name got
// want".
type VerifyError struct {
	Failures []Failure
}

func (e *VerifyError) Error() string {
	lines := make([]string, 0, len(e.Failures)+1)
	lines = append(lines, fmt.Sprintf("migratehistory: verification failed (%d check(s)); nothing was imported:", len(e.Failures)))
	for _, f := range e.Failures {
		lines = append(lines, "  "+f.String())
	}
	return strings.Join(lines, "\n")
}

// Checks are the verifications of M1-D4, in order:
//   - every golden value of §11.7 that has an implementation (the registry of
//     internal/golden; later milestones add rows there and Verify runs them);
//   - the §10.3 checks that no golden covers: Hîrbovăț deliveries 69 ·
//     9 562 437 lei (4 unpriced) and eggs Σ 5 073 155 to 18.11.2026 (the
//     arrivals ♀ 26 883 · ♂ 3 024 and the stock 31 664,9 / 1 781 kg are
//     goldens 9 and 27);
//   - alive_end ≥ 0 on every row of v_section_sex_day (stands in for
//     ledger.CheckCounts until M5).
//
// Values are read as of golden.AsOf with explicit dates; nothing here sets
// app.today.
func Checks() []Check {
	var checks []Check
	for _, r := range golden.Rows() {
		if r.Got == nil {
			continue
		}
		checks = append(checks, Check{
			ID:   fmt.Sprintf("golden %d", r.ID),
			Name: r.Name,
			Run: func(ctx context.Context, db dbq.DBTX) (any, any, error) {
				err := r.Check(ctx, golden.Env{DB: db})
				if mm, ok := errors.AsType[*golden.Mismatch](err); ok {
					return mm.Got, mm.Want, nil
				}
				if err != nil {
					return nil, r.Want, err
				}
				return r.Want, r.Want, nil
			},
		})
	}
	return append(checks,
		Check{ID: "10.3", Name: "Hîrbovăț deliveries", Run: checkDeliveries},
		Check{ID: "10.3", Name: "Hîrbovăț eggs Σ to 18.11.2026", Run: checkEggs},
		Check{ID: "D4", Name: "alive_end ≥ 0 in v_section_sex_day", Run: checkNeverNegative},
	)
}

// Verify runs Checks (and extra) against db, the import transaction or a
// pool, and returns a *VerifyError naming every failure.
func Verify(ctx context.Context, db dbq.DBTX, extra ...Check) error {
	var failures []Failure
	for _, c := range append(Checks(), extra...) {
		got, want, err := c.Run(ctx, db)
		if err != nil || !reflect.DeepEqual(got, want) {
			failures = append(failures, Failure{ID: c.ID, Name: c.Name, Got: got, Want: want, Err: err})
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	if len(failures) > 0 {
		return &VerifyError{Failures: failures}
	}
	return nil
}

// Deliveries is the §10.3 delivery check's value.
type Deliveries struct {
	Count    int
	TotalLei string // 2 dp
	Unpriced int
}

func checkDeliveries(ctx context.Context, db dbq.DBTX) (any, any, error) {
	want := Deliveries{Count: 69, TotalLei: "9562437.00", Unpriced: 4}
	farm, err := golden.FarmID(ctx, db)
	if err != nil {
		return nil, want, err
	}
	d, err := dbq.New(db).VerifyFarmDeliveries(ctx, farm)
	if err != nil {
		return nil, want, err
	}
	return Deliveries{Count: int(d.Deliveries), TotalLei: d.TotalLei.StringFixed(2), Unpriced: int(d.Unpriced)}, want, nil
}

// checkEggs sums the eggs of the cycle open on golden.AsOf (Seria 2026 from
// 22.02.2026; the Seria 2025 eggs are not part of the figure).
func checkEggs(ctx context.Context, db dbq.DBTX) (any, any, error) {
	want := int64(5073155)
	farm, err := golden.FarmID(ctx, db)
	if err != nil {
		return nil, want, err
	}
	c, err := calc.CycleAt(ctx, db, farm, golden.AsOf)
	if err != nil {
		return nil, want, err
	}
	eggs, err := dbq.New(db).VerifyEggTotal(ctx, dbq.VerifyEggTotalParams{FarmID: farm, FromDay: c.StartDate, ToDay: golden.AsOf})
	return eggs, want, err
}

func checkNeverNegative(ctx context.Context, db dbq.DBTX) (any, any, error) {
	want := []string{}
	rows, err := dbq.New(db).VerifyNegativeAlive(ctx)
	if err != nil {
		return nil, want, err
	}
	got := []string{}
	for _, r := range rows {
		got = append(got, fmt.Sprintf("%s %s %s: %d", r.Code, r.Sex, r.Date, r.AliveEnd))
	}
	return got, want, nil
}
