// Package series serves the series page and its cycle track (SPEC §7.3,
// §12.3 "series page + cycle track", M4 D1, D7). BuildPage takes a dbq.DBTX
// and the day and instant explicitly (ARC-6), so a golden or the history
// verification can run it inside a transaction; the Service is a thin
// wrapper that adds the store and the clock and runs the builder in one
// read-only snapshot (db.Store.InReadTx, M4 D10).
package series

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Service is the series module's service.
type Service struct {
	store *db.Store
	clock timex.Clock
}

// NewService builds the series service on the app pool's store.
func NewService(store *db.Store, clock timex.Clock) *Service {
	return &Service{store: store, clock: clock}
}

// Page is the series page of seriesID for p (§7.3). An unknown series, or
// one without a cycle on p's farms, is 404 NOT_FOUND for every role (M4 Q4).
func (s *Service) Page(ctx context.Context, p auth.Principal, seriesID uuid.UUID) (PageResponse, error) {
	now := s.clock.Now()
	var res PageResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildPage(ctx, tx, Input{Principal: p, SeriesID: seriesID, Today: timex.DayOf(now), Now: now})
		return err
	})
	if errors.Is(err, ErrSeriesNotFound) {
		return PageResponse{}, httpx.NotFound()
	}
	return res, err
}
