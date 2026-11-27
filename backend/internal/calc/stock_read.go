package calc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// errWindowDays is returned by FeedWindow for a window of less than one day.
var errWindowDays = errors.New("calc: the feed window needs at least one day")

// Stock is the stock of one feed type at the end of a day (§5.8).
type Stock struct {
	FeedTypeID uuid.UUID
	Name       string
	ForSex     *dbq.Sex        // nil = ambele sexe
	LastMove   timex.Day       // the latest movement day on or before the asked day
	Kg         decimal.Decimal // full precision; negative when consumption outran deliveries
}

// ForSexText is the type's for_sex as the wire has it ("F" or "M"), nil for
// ambele sexe.
func (s Stock) ForSexText() *string {
	if s.ForSex == nil {
		return nil
	}
	sex := string(*s.ForSex)
	return &sex
}

// StockAt returns the farm's stock per feed type at the end of day, ordered
// by name: Σ deliveries − Σ allocation kg + Σ adjustments dated on or before
// it. Feed types with no movement up to that day are absent. An allocation
// counts its real kg on any day (no dense-view cut-off at app.today).
func StockAt(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]Stock, error) {
	rows, err := dbq.New(db).CalcStockEnd(ctx, dbq.CalcStockEndParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, fmt.Errorf("calc: stock %s: %w", day, err)
	}
	out := make([]Stock, 0, len(rows))
	for _, r := range rows {
		out = append(out, Stock{FeedTypeID: r.FeedTypeID, Name: r.Name, ForSex: r.ForSex, LastMove: r.LastMove, Kg: r.StockEnd})
	}
	return out, nil
}

// StockPoint is a feed type's stock at the end of one day.
type StockPoint struct {
	Day timex.Day
	Kg  decimal.Decimal // full precision
}

// StockSeries returns, per asked feed type, its stock at the end of from and
// then at the end of every day in (from, to] with a movement of that type,
// oldest first. Every asked type has the from point (0 when it has no
// movement up to then); a type id that appears twice is read once.
func StockSeries(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, typeIDs []uuid.UUID, from, to timex.Day) (map[uuid.UUID][]StockPoint, error) {
	if from.After(to) {
		return nil, errRange
	}
	rows, err := dbq.New(db).CalcStockSeries(ctx, dbq.CalcStockSeriesParams{
		FarmID: farmID, FeedTypeIds: typeIDs, FromDay: from, ToDay: to,
	})
	if err != nil {
		return nil, fmt.Errorf("calc: stock series %s–%s: %w", from, to, err)
	}
	out := make(map[uuid.UUID][]StockPoint, len(typeIDs))
	for _, r := range rows {
		out[r.FeedTypeID] = append(out[r.FeedTypeID], StockPoint{Day: r.Date, Kg: r.StockEnd})
	}
	return out, nil
}

// Lowest is the point with the lowest stock (the first one on a tie), false
// for an empty series. H5 compares it with 0.
func Lowest(series []StockPoint) (StockPoint, bool) {
	if len(series) == 0 {
		return StockPoint{}, false
	}
	low := series[0]
	for _, p := range series[1:] {
		if p.Kg.LessThan(low.Kg) {
			low = p
		}
	}
	return low, true
}

// FeedWindow is the window of the average daily consumption of the farm
// (§5.8): the last days complete days, ending on today when every section ×
// sex with birds today (has_birds, §5.11: alive at the start and at the end
// of the day) has a live allocation today, else on today − 1. A farm without
// birds today ends it on today. It is per farm, shared by the feed types.
func FeedWindow(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, today timex.Day, days int) (Window, error) {
	if days < 1 {
		return Window{}, errWindowDays
	}
	complete, err := dbq.New(db).CalcFeedDayComplete(ctx, dbq.CalcFeedDayCompleteParams{FarmID: farmID, Day: today})
	if err != nil {
		return Window{}, fmt.Errorf("calc: feed window %s: %w", today, err)
	}
	end := today
	if !complete {
		end = today.AddDays(-1)
	}
	return Window{From: end.AddDays(1 - days), To: end}, nil
}

// Consumption is the farm's feed consumed per feed type over [from, to], in
// kg at full precision (§5.7). Types without an allocation in the range are
// absent (read them as 0).
func Consumption(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, from, to timex.Day) (map[uuid.UUID]decimal.Decimal, error) {
	if from.After(to) {
		return nil, errRange
	}
	rows, err := dbq.New(db).CalcConsumption(ctx, dbq.CalcConsumptionParams{FarmID: farmID, FromDay: from, ToDay: to})
	if err != nil {
		return nil, fmt.Errorf("calc: consumption %s–%s: %w", from, to, err)
	}
	out := make(map[uuid.UUID]decimal.Decimal, len(rows))
	for _, r := range rows {
		out[r.FeedTypeID] = r.Kg
	}
	return out, nil
}

// StockItem is one feed type's stock row (§5.8, §7.12): the stock at the end
// of today, the average daily consumption over the window and the days left.
type StockItem struct {
	Stock
	AvgDaily decimal.Decimal // Σ kg of the type over the window ÷ its days, full precision
	Left     DaysLeftResult
}

// StockDays is a farm's stock rows with their shared window.
type StockDays struct {
	Window Window
	Items  []StockItem // ordered by feed type name, as StockAt
}

// ReadStockDays reads the farm's stock rows as of today (§5.8): StockAt(today),
// the FeedWindow of s.StockAvgDays days, the Consumption over it and DaysLeft
// against s.StockDaysMin. The farm-level "fără ciclu deschis" row is the
// caller's.
func ReadStockDays(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, today timex.Day, s AppSettings) (StockDays, error) {
	stock, err := StockAt(ctx, db, farmID, today)
	if err != nil {
		return StockDays{}, err
	}
	w, err := FeedWindow(ctx, db, farmID, today, s.StockAvgDays)
	if err != nil {
		return StockDays{}, err
	}
	used, err := Consumption(ctx, db, farmID, w.From, w.To)
	if err != nil {
		return StockDays{}, err
	}
	n := decimal.NewFromInt(int64(s.StockAvgDays))
	out := StockDays{Window: w, Items: make([]StockItem, 0, len(stock))}
	for _, st := range stock {
		avg := used[st.FeedTypeID].Div(n)
		out.Items = append(out.Items, StockItem{Stock: st, AvgDaily: avg, Left: DaysLeft(st.Kg, avg, w.To, s.StockDaysMin)})
	}
	return out, nil
}
