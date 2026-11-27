// Package farms serves the farm side of the shell and of S4 (SPEC §12.3): the
// bootstrap (§7.2), the farm page (§7.3) and the Rezumat summary (§7.6). The
// builders take a dbq.DBTX and the day and instant explicitly (ARC-6), so the
// goldens and the history verification run them inside a transaction; the
// Service is a thin wrapper that adds the store, the clock and the access
// check, and runs each builder in one read-only snapshot (db.Store.InReadTx).
package farms

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Service is the farms module's service.
type Service struct {
	store *db.Store
	clock timex.Clock
}

// NewService builds the farms service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock}
}

// Page is the farm page of farmID for p (§7.3). cycleID nil shows the
// default cycle. Errors are envelope errors (M2 D4): 403 FARM_FORBIDDEN for a
// farm outside the caller's farms (an unknown id included, for anyone but the
// owner), 404 NOT_FOUND for an owner's unknown farm and for a cycle of
// another farm.
func (s *Service) Page(ctx context.Context, p auth.Principal, farmID uuid.UUID, cycleID *uuid.UUID) (PageResponse, error) {
	if !p.CanSeeFarm(farmID) {
		return PageResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	var res PageResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildPage(ctx, tx, Input{
			Principal: p, FarmID: farmID, CycleID: cycleID, Today: timex.DayOf(now), Now: now,
		})
		return err
	})
	if errors.Is(err, ErrFarmNotFound) || errors.Is(err, ErrCycleNotFound) {
		return PageResponse{}, httpx.NotFound()
	}
	return res, err
}

// Bootstrap is the shell bootstrap for p (§7.2): every role may read it.
func (s *Service) Bootstrap(ctx context.Context, p auth.Principal) (BootstrapResponse, error) {
	now := s.clock.Now()
	var res BootstrapResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildBootstrap(ctx, tx, Input{Principal: p, Today: timex.DayOf(now), Now: now})
		return err
	})
	return res, err
}

// Summary is the Rezumat summary of farmID for p (§7.6): the cycle cycleID
// (nil: the default one) over period (zero: the whole cycle). Errors are
// those of Page, plus 400 VALIDATION on `period` for an interval that ends
// after today or a period that doesn't apply.
func (s *Service) Summary(ctx context.Context, p auth.Principal, farmID uuid.UUID, cycleID *uuid.UUID, period calc.Period) (SummaryResponse, error) {
	if !p.CanSeeFarm(farmID) {
		return SummaryResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	var res SummaryResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildSummary(ctx, tx, Input{
			Principal: p, FarmID: farmID, CycleID: cycleID, Period: period, Today: today, Now: now,
		})
		return err
	})
	switch {
	case errors.Is(err, ErrFarmNotFound) || errors.Is(err, ErrCycleNotFound):
		return SummaryResponse{}, httpx.NotFound()
	case errors.Is(err, calc.ErrIntervalFuture):
		return SummaryResponse{}, httpx.Validation(httpx.Issue{
			Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.IntervalInFuture(period.To, today),
		})
	case errors.Is(err, calc.ErrPeriod):
		return SummaryResponse{}, httpx.Validation(httpx.Issue{
			Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.InvalidData(),
		})
	}
	return res, err
}
