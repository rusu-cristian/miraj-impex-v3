package golden

import (
	"context"
	"fmt"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
)

// M3 rows (M3 D12): golden 26 reads the Consum day row (periodtable.Build,
// view=report-feed), golden 27 the calc stock reader (calc.StockAt) and
// golden 28 the days left (calc.ReadStockDays), as of AsOf and Now, never
// app_today(). The Stoc report (reports.BuildFeed), the summary's
// feedDaysLeft and the Hrană zilnică form state show the same figures as
// golden 28 (tested in golden_test).

// StockDaysLeft is golden 28: the farm's consumption window (§5.8) and, per
// feed type with a movement, the average daily kg (1 dp), the days left
// (1 dp) and the run-out day.
type StockDaysLeft struct {
	From, To string
	Items    []TypeDaysLeft
}

// TypeDaysLeft is one feed type of StockDaysLeft. Days and RunOut are ""
// when the type has no consumption over the window.
type TypeDaysLeft struct {
	Type     string
	AvgDaily string
	Days     string
	RunOut   string
}

// consumFirstPage is the farm's Consum as Rapoarte › Hrană opens it on AsOf:
// the open cycle, the newest 4 weeks with their days.
func consumFirstPage(ctx context.Context, env Env) (periodtable.TableResponse, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return periodtable.TableResponse{}, err
	}
	return periodtable.Build(ctx, env.DB, periodtable.Input{
		Query: periodtable.TableQuery{View: "report-feed", FarmID: &farm},
		Today: AsOf, Now: Now,
	})
}

// feedDay reads the Consum day row of AsOf and rounds each sex and their
// total once, from full precision. A sex without a ration that day ("–") is
// an error: golden 26 has both.
func feedDay(ctx context.Context, env Env) (any, error) {
	t, err := consumFirstPage(ctx, env)
	if err != nil {
		return nil, err
	}
	d, err := dayRow(t, AsOf)
	if err != nil {
		return nil, err
	}
	c := cells{c: d.Cells}
	f, m := c.dec(periodtable.KeyFeedKgF), c.dec(periodtable.KeyFeedKgM)
	if c.err != nil {
		return nil, c.err
	}
	if f == nil || m == nil {
		return nil, fmt.Errorf("the Consum day row %s has no ration (♀ %v, ♂ %v)", AsOf, f, m)
	}
	total := f.Add(*m)
	return FeedDay{Kg: calc.PerSex[string]{F: fixed(f, 1), M: fixed(m, 1)}, Total: fixed(&total, 1)}, nil
}

// daysLeft reads the farm's stock rows as of AsOf with the app settings
// (alert.stock_avg_days, alert.stock_days_min).
func daysLeft(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	settings, err := calc.ReadAppSettings(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	s, err := calc.ReadStockDays(ctx, env.DB, farm, AsOf, settings)
	if err != nil {
		return nil, err
	}
	out := StockDaysLeft{From: s.Window.From.String(), To: s.Window.To.String(), Items: make([]TypeDaysLeft, 0, len(s.Items))}
	for _, it := range s.Items {
		item := TypeDaysLeft{Type: it.Name, AvgDaily: fixed(&it.AvgDaily, 1), Days: fixed(it.Left.Days, 1)}
		if it.Left.RunOut != nil {
			item.RunOut = it.Left.RunOut.String()
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}
