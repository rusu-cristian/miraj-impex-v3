package periodtable

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Service serves the PeriodTable on the app pool: a thin wrapper that adds
// the store, the clock and the farm check to Build (ARC-6), and runs them in
// one read-only snapshot (db.Store.InReadTx).
type Service struct {
	store *db.Store
	clock timex.Clock
}

// NewService builds the PeriodTable service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock}
}

// Table is the PeriodTable of q for p (§7.4), q having passed ParseQuery and
// Validate (so q.FarmID is set). Errors are envelope errors:
//   - 403 FARM_FORBIDDEN for a farm outside the caller's farms, an unknown id
//     included for anyone but the owner (M2 D4);
//   - 404 NOT_FOUND for an owner's unknown farm and for a cycleId that is not
//     one of the farm's cycles;
//   - 400 VALIDATION on `period` for an interval that ends after today ("Până
//     la {to} e în viitor. Alege o dată până la {today}.") and on the paging
//     parameter out of range for the period (before, count, sapt, fromSapt).
func (s *Service) Table(ctx context.Context, p auth.Principal, q TableQuery) (TableResponse, error) {
	if q.FarmID == nil {
		return TableResponse{}, httpx.Validation(httpx.Issue{
			Path: paramFarmID, Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	if !p.CanSeeFarm(*q.FarmID) {
		return TableResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	var res TableResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) error {
		if _, err := dbq.New(tx).FarmsGetFarm(ctx, *q.FarmID); errors.Is(err, pgx.ErrNoRows) {
			return httpx.NotFound()
		} else if err != nil {
			return fmt.Errorf("periodtable: farm: %w", err)
		}
		built, err := Build(ctx, tx, Input{Query: q, Today: today, Now: now})
		if err != nil {
			return buildError(err, q, today)
		}
		res = built
		return nil
	})
	if err != nil {
		return TableResponse{}, err
	}
	return res, nil
}

// buildError turns Build's errors into envelope errors; anything else stays
// as it is (a 500).
func buildError(err error, q TableQuery, today timex.Day) error {
	var pe *calc.PageError
	switch {
	case errors.Is(err, calc.ErrNoCycle):
		return httpx.NotFound()
	case errors.Is(err, calc.ErrIntervalFuture):
		return httpx.Validation(httpx.Issue{
			Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.IntervalInFuture(q.Period.To, today)})
	case errors.Is(err, calc.ErrPeriod):
		return httpx.Validation(httpx.Issue{Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.InvalidData()})
	case errors.As(err, &pe):
		return httpx.Validation(httpx.Issue{Path: pe.Param, Code: httpx.IssueOutOfRange, Message: messages.InvalidData()})
	}
	return err
}
