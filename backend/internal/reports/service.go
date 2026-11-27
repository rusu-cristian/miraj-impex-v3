package reports

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

// Service is the reports module's service.
type Service struct {
	store *db.Store
	clock timex.Clock
}

// NewService builds the reports service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock}
}

// Feed is Rapoarte › Hrană for p (§7.12): farmID nil means every farm of
// the caller. period is checked as the PeriodTable checks it (an interval
// that ends after today is 400 VALIDATION on `period`) but the Stoc part
// doesn't use it: the stock is "now" (M3 D9). Errors are envelope errors:
// 403 FARM_FORBIDDEN for a farm outside the caller's farms (an unknown id
// included, for anyone but the owner), 404 NOT_FOUND for an owner's unknown
// farm (M2 D4).
func (s *Service) Feed(ctx context.Context, p auth.Principal, farmID *uuid.UUID, period calc.Period) (FeedResponse, error) {
	if farmID != nil && !p.CanSeeFarm(*farmID) {
		return FeedResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	if period.Kind == calc.PeriodInterval && period.To.After(today) {
		return FeedResponse{}, httpx.Validation(httpx.Issue{
			Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.IntervalInFuture(period.To, today),
		})
	}
	var res FeedResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildFeed(ctx, tx, Input{Principal: p, FarmID: farmID, Today: today, Now: now})
		return err
	})
	if errors.Is(err, ErrFarmNotFound) {
		return FeedResponse{}, httpx.NotFound()
	}
	return res, err
}
