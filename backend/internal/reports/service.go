package reports

import (
	"context"
	"errors"

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

// Feed is Rapoarte › Hrană for p (§7.12, M4 D2, Q6–Q8): q.FarmID nil means
// every farm of the caller. A viewer gets the response without prices
// (FeedResponse, FeedPage), a manager or the owner the one with prices and
// Cost (FeedResponsePriced, FeedPagePriced), as a FeedBody; with q.Cursor only the next page
// of Livrări rows. An interval that ends after today is 400 VALIDATION on
// `period`. Errors are envelope errors: 403 FARM_FORBIDDEN for a farm outside
// the caller's farms (an unknown id included, for anyone but the owner), 404
// NOT_FOUND for an owner's unknown farm (M2 D4). Vets never reach it (the
// route's role check).
func (s *Service) Feed(ctx context.Context, p auth.Principal, q FeedQuery) (FeedBody, error) {
	if q.FarmID != nil && !p.CanSeeFarm(*q.FarmID) {
		return nil, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	if q.Period.Kind == calc.PeriodInterval && q.Period.To.After(today) {
		return nil, httpx.Validation(httpx.Issue{
			Path: paramPeriod, Code: httpx.IssueOutOfRange, Message: messages.IntervalInFuture(q.Period.To, today),
		})
	}
	in := Input{Principal: p, FarmID: q.FarmID, Period: q.Period, Cursor: q.Cursor, Limit: q.Limit, Today: today, Now: now}
	var res FeedBody
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		switch priced := seesPrices(p); {
		case q.Cursor != nil && priced:
			res, err = BuildFeedPagePriced(ctx, tx, in)
		case q.Cursor != nil:
			res, err = BuildFeedPage(ctx, tx, in)
		case priced:
			res, err = BuildFeedPriced(ctx, tx, in)
		default:
			res, err = BuildFeed(ctx, tx, in)
		}
		return err
	})
	if errors.Is(err, ErrFarmNotFound) {
		return nil, httpx.NotFound()
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}
