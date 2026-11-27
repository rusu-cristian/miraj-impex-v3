// Package reports serves Rapoarte (SPEC §12.3): in M3 the Stoc part of
// Rapoarte › Hrană (§7.12, M3 D9, Q6). The builders take a dbq.DBTX and the
// day and instant explicitly (ARC-6), so the goldens run them inside a
// transaction; the Service is a thin wrapper that adds the store, the clock
// and the access check, and runs each builder in one read-only snapshot
// (db.Store.InReadTx).
package reports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ErrFarmNotFound is returned by BuildFeed for a FarmID that names no farm.
var ErrFarmNotFound = errors.New("reports: farm not found")

// Input is what the report builders read: the caller, the farm asked for
// (nil: every farm of the caller) and the day and instant of the request.
type Input struct {
	Principal auth.Principal
	FarmID    *uuid.UUID
	Today     timex.Day
	Now       time.Time
}

// BuildFeed builds GET /api/reports/feed (§7.12) on db: the Stoc rows as of
// in.Today, per farm in name order. A farm with an open cycle has one
// TypeStock per feed type with a movement up to today, in type name order
// (calc.ReadStockDays, §5.8); a farm without one has a single FarmStock
// "fără ciclu deschis" and never shows a stock (§5.8). in.FarmID nil reads
// every farm for the owner and the caller's farms for anyone else. It does
// not check the caller's access to in.FarmID (the service does, M2 D4);
// ErrFarmNotFound names an unknown in.FarmID.
func BuildFeed(ctx context.Context, db dbq.DBTX, in Input) (FeedResponse, error) {
	res := FeedResponse{Stock: []StockRow{}, AsOf: timex.At(in.Now)}
	params := dbq.ReportsFeedFarmsParams{AllFarms: in.Principal.Role == auth.Owner, FarmIds: in.Principal.FarmIDs}
	if in.FarmID != nil {
		params = dbq.ReportsFeedFarmsParams{FarmIds: []uuid.UUID{*in.FarmID}}
	}
	if params.FarmIds == nil {
		params.FarmIds = []uuid.UUID{}
	}
	farms, err := dbq.New(db).ReportsFeedFarms(ctx, params)
	if err != nil {
		return res, fmt.Errorf("reports: feed farms: %w", err)
	}
	if in.FarmID != nil && len(farms) == 0 {
		return res, ErrFarmNotFound
	}

	var settings *calc.AppSettings
	for _, f := range farms {
		if !f.HasOpenCycle {
			row := FarmStock{Farm: f.Name, Lamp: calc.Oprit, Reason: messages.NoOpenCycle()}
			if f.SeriesName != nil && f.EndDate != nil {
				row.Sub = messages.LastCycleClosed(*f.SeriesName, *f.EndDate)
			}
			res.Stock = append(res.Stock, row)
			continue
		}
		if settings == nil {
			s, err := calc.ReadAppSettings(ctx, db)
			if err != nil {
				return res, err
			}
			settings = &s
		}
		stock, err := calc.ReadStockDays(ctx, db, f.ID, in.Today, *settings)
		if err != nil {
			return res, err
		}
		for _, it := range stock.Items {
			res.Stock = append(res.Stock, typeStock(f.Name, stock.Window, it, *settings))
		}
	}
	return res, nil
}

func typeStock(farm string, w calc.Window, it calc.StockItem, settings calc.AppSettings) TypeStock {
	row := TypeStock{
		Farm: farm, Type: it.Name, Kg: num.Of(it.Kg), AvgDaily: num.Of(it.AvgDaily),
		Window: [2]timex.Day{w.From, w.To}, RunOut: it.Left.RunOut,
		Lamp: it.Left.Lamp, Reason: it.Left.Reason.Text(settings), ForSex: it.ForSexText(),
	}
	if it.Left.Days != nil {
		days := num.Of(*it.Left.Days)
		row.DaysLeft = &days
	}
	return row
}
