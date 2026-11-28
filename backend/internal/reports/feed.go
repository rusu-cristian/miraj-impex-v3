// Package reports serves Rapoarte (SPEC §12.3): Rapoarte › Hrană's Stoc,
// Livrări and Cost (§7.12, M3 D9, M4 D2, D6). The builders take a dbq.DBTX
// and the day and instant explicitly (ARC-6), so the goldens run them inside
// a transaction; the Service is a thin wrapper that adds the store, the clock
// and the access check, and runs each builder in one read-only snapshot
// (db.Store.InReadTx).
package reports

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

var (
	// ErrFarmNotFound is returned by the builders for a FarmID that names no
	// farm.
	ErrFarmNotFound = errors.New("reports: farm not found")
	// ErrPricesHidden is returned by the priced builders for a caller who may
	// not see prices (vet, viewer: D64, D200). The service never asks; it
	// guards the builders.
	ErrPricesHidden = errors.New("reports: prices are not shown to this role")
)

// DefaultLimit and MaxLimit bound the Livrări rows of one page (M4 Q7).
const (
	DefaultLimit = 12
	MaxLimit     = 100
)

// Input is what the report builders read: the caller, the farm asked for
// (nil: every farm of the caller), the period (zero: cycle), the Livrări page
// (Cursor nil: the first page; Limit 0: DefaultLimit) and the day and instant
// of the request.
type Input struct {
	Principal auth.Principal
	FarmID    *uuid.UUID
	Period    calc.Period
	Cursor    *Cursor
	Limit     int
	Today     timex.Day
	Now       time.Time
}

func (in Input) period() calc.Period {
	if in.Period.Kind == "" {
		return calc.Period{Kind: calc.PeriodCycle}
	}
	return in.Period
}

func (in Input) limit() int {
	if in.Limit <= 0 {
		return DefaultLimit
	}
	return min(in.Limit, MaxLimit)
}

// seesPrices reports whether p may see prices: the owner and managers
// (§3.1, D200).
func seesPrices(p auth.Principal) bool { return p.Role == auth.Owner || p.Role == auth.Manager }

// BuildFeed builds GET /api/reports/feed (§7.12) without prices, as a viewer
// gets it: Stoc and Livrări (no Cost). See readFeed.
func BuildFeed(ctx context.Context, db dbq.DBTX, in Input) (FeedResponse, error) {
	f, err := readFeed(ctx, db, in, false)
	if err != nil {
		return FeedResponse{AsOf: timex.At(in.Now)}, err
	}
	d := DeliveriesView{
		Total: f.total, ByFarm: f.byFarm, Unpriced: UnpricedCount{Count: f.unpriced.Count},
		Rows: f.page.view(), NextCursor: f.page.next, LastBefore: f.lastBefore,
	}
	return FeedResponse{Stock: f.stock, Deliveries: d, AsOf: timex.At(in.Now)}, nil
}

// BuildFeedPriced builds GET /api/reports/feed for a manager or the owner:
// Stoc, Livrări with prices and Cost. ErrPricesHidden for anyone else.
func BuildFeedPriced(ctx context.Context, db dbq.DBTX, in Input) (FeedResponsePriced, error) {
	res := FeedResponsePriced{AsOf: timex.At(in.Now)}
	if !seesPrices(in.Principal) {
		return res, ErrPricesHidden
	}
	f, err := readFeed(ctx, db, in, true)
	if err != nil {
		return res, err
	}
	res.Stock = f.stock
	res.Deliveries = DeliveriesPriced{
		Total: f.total, ByFarm: f.byFarm, Unpriced: f.unpriced, Rows: f.page.priced(in.Principal),
		NextCursor: f.page.next, LastBefore: f.lastBefore,
	}
	res.Cost = f.cost
	return res, nil
}

// BuildFeedPage builds a next page of Livrări rows without prices (a request
// with in.Cursor, M4 Q7): the farms' ranges and one page of rows, nothing
// else.
func BuildFeedPage(ctx context.Context, db dbq.DBTX, in Input) (FeedPage, error) {
	pg, err := readPage(ctx, db, in)
	return FeedPage{Deliveries: DeliveriesPage{Rows: pg.view(), NextCursor: pg.next}, AsOf: timex.At(in.Now)}, err
}

// BuildFeedPagePriced builds a next page of Livrări rows with prices (manager,
// owner). ErrPricesHidden for anyone else.
func BuildFeedPagePriced(ctx context.Context, db dbq.DBTX, in Input) (FeedPagePriced, error) {
	res := FeedPagePriced{Deliveries: DeliveriesPagePriced{Rows: []DeliveryRowPriced{}}, AsOf: timex.At(in.Now)}
	if !seesPrices(in.Principal) {
		return res, ErrPricesHidden
	}
	pg, err := readPage(ctx, db, in)
	if err != nil {
		return res, err
	}
	res.Deliveries = DeliveriesPagePriced{Rows: pg.priced(in.Principal), NextCursor: pg.next}
	return res, nil
}

// BuildCost builds the Cost part of GET /api/reports/feed alone (golden 31):
// the farms' ranges, their deliveries and the cost bases. ErrPricesHidden for
// a caller who may not see prices.
func BuildCost(ctx context.Context, db dbq.DBTX, in Input) (Cost, error) {
	if !seesPrices(in.Principal) {
		return Cost{AsOf: in.Today, Cycles: []CostCycle{}}, ErrPricesHidden
	}
	sc, err := readScope(ctx, db, in)
	if err != nil {
		return Cost{AsOf: in.Today, Cycles: []CostCycle{}}, err
	}
	ranges, err := sc.ranges(in.period(), in.Today)
	if err != nil {
		return Cost{AsOf: in.Today, Cycles: []CostCycle{}}, err
	}
	dels, err := calc.ReadDeliveries(ctx, db, ranges)
	if err != nil {
		return Cost{AsOf: in.Today, Cycles: []CostCycle{}}, err
	}
	return buildCost(ctx, db, sc, dels, in.Today)
}

// feed is what the first page of the report reads, before it is shaped for
// a role.
type feed struct {
	stock      []StockRow
	total      int
	byFarm     []FarmCount
	unpriced   UnpricedSummary
	page       page
	lastBefore *LastDelivery
	cost       Cost // read only when priced
}

// readFeed reads the first page of the report as of in.Today, in at most 8
// statements (M4 T3): the farms and their cycles; the Stoc (§5.8: the app
// settings and, per farm with an open cycle, calc.ReadStockDays); the
// deliveries of each farm's range (calc.DeliveryRange, calc.ReadDeliveries);
// one page of rows; the newest delivery before the ranges only when they hold
// none; and, when priced, the cost bases (one statement, only with a cost
// row).
//
// Stoc: per farm in name order, one TypeStock per feed type with a movement
// up to today (in type name order) for a farm with an open cycle, else a
// single FarmStock "fără ciclu deschis" (a farm without an open cycle never
// shows a stock, §5.8). in.FarmID nil reads every farm for the owner and the
// caller's farms for anyone else. It does not check the caller's access to
// in.FarmID (the service does, M2 D4); ErrFarmNotFound names an unknown
// in.FarmID.
func readFeed(ctx context.Context, db dbq.DBTX, in Input, priced bool) (feed, error) {
	f := feed{stock: []StockRow{}, byFarm: []FarmCount{}, unpriced: UnpricedSummary{Kg: num.Of(decimal.Zero), Dates: []timex.Day{}}}
	sc, err := readScope(ctx, db, in)
	if err != nil {
		return f, err
	}
	ranges, err := sc.ranges(in.period(), in.Today)
	if err != nil {
		return f, err
	}
	if f.stock, err = buildStock(ctx, db, sc, in.Today); err != nil {
		return f, err
	}
	dels, err := calc.ReadDeliveries(ctx, db, ranges)
	if err != nil {
		return f, err
	}
	f.total, f.byFarm, f.unpriced = summarize(sc, dels)
	if f.page, err = readRows(ctx, db, ranges, nil, in.limit()); err != nil {
		return f, err
	}
	if f.total == 0 {
		if f.lastBefore, err = readLastBefore(ctx, db, ranges); err != nil {
			return f, err
		}
	}
	if priced {
		if f.cost, err = buildCost(ctx, db, sc, dels, in.Today); err != nil {
			return f, err
		}
	}
	return f, nil
}

// readPage reads a next page of rows (in.Cursor set) in 2 statements: the
// farms and their cycles (for the ranges), and the rows.
func readPage(ctx context.Context, db dbq.DBTX, in Input) (page, error) {
	empty := page{rows: []dbq.ReportsDeliveryRowsRow{}}
	sc, err := readScope(ctx, db, in)
	if err != nil {
		return empty, err
	}
	ranges, err := sc.ranges(in.period(), in.Today)
	if err != nil {
		return empty, err
	}
	pg, err := readRows(ctx, db, ranges, in.Cursor, in.limit())
	if err != nil {
		return empty, err
	}
	return pg, nil
}

// scope is the report's farms, in name order, with their cycles.
type scope struct {
	farms []farmScope
	byID  map[uuid.UUID]int // index into farms
}

// farmScope is one farm of the report with every cycle of it, newest first
// (calc.FarmCycles' order).
type farmScope struct {
	ID     uuid.UUID
	Name   string
	Cycles []calc.CycleInfo
	meta   map[uuid.UUID]cycleMeta // by cycle id
}

// cycleMeta is what the report needs about a cycle besides calc.CycleInfo.
type cycleMeta struct {
	Type        dbq.FarmType // the guide's type, else the farm's (calc.CycleInfo.Type)
	SeriesFarms int          // farms with a cycle of the series, the caller's or not
}

// shown is the farm's shown cycle: the open one, else the latest
// (calc.DefaultCycle); nil without any.
func (f farmScope) shown() *calc.CycleInfo {
	c, ok := calc.DefaultCycle(f.Cycles)
	if !ok {
		return nil
	}
	return &c
}

// readScope reads the report's farms and their cycles (one statement).
func readScope(ctx context.Context, db dbq.DBTX, in Input) (scope, error) {
	params := dbq.ReportsFeedCyclesParams{AllFarms: in.Principal.Role == auth.Owner, FarmIds: in.Principal.FarmIDs}
	if in.FarmID != nil {
		params = dbq.ReportsFeedCyclesParams{FarmIds: []uuid.UUID{*in.FarmID}}
	}
	if params.FarmIds == nil {
		params.FarmIds = []uuid.UUID{}
	}
	rows, err := dbq.New(db).ReportsFeedCycles(ctx, params)
	if err != nil {
		return scope{}, fmt.Errorf("reports: feed farms: %w", err)
	}
	sc := scope{farms: []farmScope{}, byID: map[uuid.UUID]int{}}
	for _, r := range rows {
		i, ok := sc.byID[r.FarmID]
		if !ok {
			i = len(sc.farms)
			sc.byID[r.FarmID] = i
			sc.farms = append(sc.farms, farmScope{
				ID: r.FarmID, Name: r.FarmName, Cycles: []calc.CycleInfo{}, meta: map[uuid.UUID]cycleMeta{},
			})
		}
		if r.CycleID == nil || r.SeriesID == nil || r.StartDate == nil || r.PlacementDate == nil ||
			r.SeriesName == nil || r.SeriesStatus == nil {
			continue // a farm without any cycle
		}
		f := &sc.farms[i]
		f.Cycles = append(f.Cycles, calc.CycleInfo{
			Cycle: calc.Cycle{
				ID: *r.CycleID, SeriesID: *r.SeriesID, StartDate: *r.StartDate, EndDate: r.EndDate,
				Age: calc.SeriesAge{PlacementDate: *r.PlacementDate, AgeAtPlacementDays: int(r.AgeAtPlacementDays)},
			},
			FarmID:     r.FarmID,
			SeriesName: *r.SeriesName,
		})
		typ := r.FarmType
		if r.GuideType != nil {
			typ = *r.GuideType
		}
		f.meta[*r.CycleID] = cycleMeta{Type: typ, SeriesFarms: int(r.SeriesFarms)}
	}
	if in.FarmID != nil && len(sc.farms) == 0 {
		return sc, ErrFarmNotFound
	}
	return sc, nil
}

// ranges is each farm's Livrări and Cost range for p (calc.DeliveryRange,
// M4 Q6), in the scope's farm order.
func (sc scope) ranges(p calc.Period, today timex.Day) ([]calc.FarmRange, error) {
	out := make([]calc.FarmRange, 0, len(sc.farms))
	for _, f := range sc.farms {
		r, err := calc.DeliveryRange(f.Cycles, f.shown(), p, today)
		if err != nil {
			return nil, fmt.Errorf("reports: %s range: %w", f.Name, err)
		}
		out = append(out, calc.FarmRange{FarmID: f.ID, DayRange: r})
	}
	return out, nil
}

// buildStock is the Stoc part (§5.8, M3 D9): see readFeed. The settings are
// read once, and only for a farm with an open cycle.
func buildStock(ctx context.Context, db dbq.DBTX, sc scope, today timex.Day) ([]StockRow, error) {
	out := []StockRow{}
	var settings *calc.AppSettings
	for _, f := range sc.farms {
		if len(f.Cycles) == 0 || f.Cycles[0].EndDate != nil { // the latest cycle is closed, or none
			row := FarmStock{Farm: f.Name, Lamp: calc.Oprit, Reason: messages.NoOpenCycle()}
			if len(f.Cycles) > 0 {
				row.Sub = messages.LastCycleClosed(f.Cycles[0].SeriesName, *f.Cycles[0].EndDate)
			}
			out = append(out, row)
			continue
		}
		if settings == nil {
			s, err := calc.ReadAppSettings(ctx, db)
			if err != nil {
				return nil, err
			}
			settings = &s
		}
		stock, err := calc.ReadStockDays(ctx, db, f.ID, today, *settings)
		if err != nil {
			return nil, err
		}
		for _, it := range stock.Items {
			out = append(out, typeStock(f.Name, stock.Window, it, *settings))
		}
	}
	return out, nil
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
