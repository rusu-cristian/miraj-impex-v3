package golden

import (
	"context"
	"fmt"
	"time"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/series"
)

// M4 rows (M4 D8): golden 29 reads the farm's deliveries over its Seria 2026
// range (calc.DeliveryRange, calc.ReadDeliveries), golden 30 derives the
// 13.11.2026 delivery's price from the row Rapoarte › Hrană › Livrări shows
// (calc.DerivePrice) and golden 31 is the Cost block (reports.BuildCost), as
// of AsOf and Now, never app_today(). Golden 35 is the placement block and
// the placement farm's track cycle of the series page (series.BuildPage).

// Deliveries is golden 29: the farm's live deliveries over its shown cycle's
// range (§5.9: pre-stocking days included): the count, the kg by the feed
// type's sex and the priced lei (whole kg and lei), and the unpriced ones.
type Deliveries struct {
	Count    int
	KgF, KgM string
	Lei      string
	Unpriced UnpricedKg
}

// UnpricedKg are the deliveries without a price ("fără preț"): their count,
// Σ kg and each delivery's day and kg, oldest first, then by type name.
type UnpricedKg struct {
	Count int
	Kg    string
	Items []DeliveryKg
}

// DeliveryKg is one delivery's day and kg (whole kg).
type DeliveryKg struct {
	Date string
	Kg   string
}

// Price is golden 30: a delivery's kg and invoice total (whole) and the price
// per kg derived from them (§6.8) at the 2 decimals the S11 board shows.
// StoredPerKg is the stored price_per_kg (4 dp), which must equal the
// derivation at 4 dp.
type Price struct {
	Kg          string
	TotalLei    string
	PerKg       string
	StoredPerKg string
}

// Cost is golden 31: the Cost block of Rapoarte › Hrană for the owner, every
// farm, period=cycle (§5.9, M4 Q6). Total is zero when the block has none.
type Cost struct {
	Cycles []CycleCost
	Total  CostTotal
}

// CycleCost is one cost row: its label ("Telița · Seria 2026"), the priced lei
// (whole), the birds at the cycle start (♀ + ♂), the eggs in the range (0 for
// a growth cycle), lei per bird and per egg (2 dp; "" when there is none) and
// the unpriced kg left out ("" when none).
type CycleCost struct {
	Cycle      string
	Lei        string
	Birds      int
	Eggs       int
	PerBird    string
	PerEgg     string
	UnpricedKg string
}

// CostTotal is the series total: Σ lei (whole), lei per bird of the series
// placement and per egg (2 dp), and the unpriced kg left out (whole).
type CostTotal struct {
	Lei        string
	PerBird    string
	PerEgg     string
	ExcludedKg string
}

// ownerAt is the owner as the M4 rows read the reports: every farm, prices
// shown.
var ownerAt = auth.Principal{Name: "golden", Role: auth.Owner}

// deliveries reads golden 29: calc.ReadDeliveries over the golden farm's
// range for period=cycle (calc.DeliveryRange of its cycle on AsOf).
func deliveries(ctx context.Context, env Env) (any, error) {
	farm, c, err := cycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	cycles, err := calc.FarmCycles(ctx, env.DB, farm)
	if err != nil {
		return nil, err
	}
	var shown *calc.CycleInfo
	for i := range cycles {
		if cycles[i].ID == c.ID {
			shown = &cycles[i]
		}
	}
	if shown == nil {
		return nil, fmt.Errorf("the cycle %s on %s is not among the farm's cycles", c.ID, AsOf)
	}
	r, err := calc.DeliveryRange(cycles, shown, calc.Period{Kind: calc.PeriodCycle}, AsOf)
	if err != nil {
		return nil, err
	}
	dels, err := calc.ReadDeliveries(ctx, env.DB, []calc.FarmRange{{FarmID: farm, DayRange: r}})
	if err != nil {
		return nil, err
	}
	if len(dels) != 1 {
		return nil, fmt.Errorf("calc.ReadDeliveries returned %d farms, want 1", len(dels))
	}
	d := dels[0]
	if !d.KgBoth.IsZero() {
		return nil, fmt.Errorf("%s kg of feed for both sexes: golden 29 has none", d.KgBoth)
	}
	u := UnpricedKg{Count: d.Unpriced.Count, Kg: fixed(&d.Unpriced.Kg, 0), Items: make([]DeliveryKg, 0, len(d.Unpriced.Items))}
	for _, it := range d.Unpriced.Items {
		u.Items = append(u.Items, DeliveryKg{Date: it.Date.String(), Kg: fixed(&it.Kg, 0)})
	}
	return Deliveries{Count: d.Count, KgF: fixed(&d.Kg.F, 0), KgM: fixed(&d.Kg.M, 0), Lei: fixed(&d.Priced.Lei, 0), Unpriced: u}, nil
}

// priceDay is the day of golden 30's delivery.
var priceDay = day(2026, time.November, 13)

// price reads golden 30: the golden farm's one delivery of 13.11.2026 as the
// owner's Livrări shows it (reports.BuildFeedPagePriced over that day), and
// the price per kg calc.DerivePrice gives from its kg and total. A stored
// price per kg that differs from the derivation at 4 dp is an error.
func price(ctx context.Context, env Env) (any, error) {
	farm, err := FarmID(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	pg, err := reports.BuildFeedPagePriced(ctx, env.DB, reports.Input{
		Principal: ownerAt, FarmID: &farm, Period: calc.Period{Kind: calc.PeriodInterval, From: priceDay, To: priceDay},
		Today: AsOf, Now: Now,
	})
	if err != nil {
		return nil, err
	}
	if n := len(pg.Deliveries.Rows); n != 1 {
		return nil, fmt.Errorf("%d deliveries on %s, want 1", n, priceDay)
	}
	row := pg.Deliveries.Rows[0]
	if row.TotalLei == nil || row.PricePerKg == nil {
		return nil, fmt.Errorf("the delivery of %s has no price", priceDay)
	}
	kg, total, perKg := row.Kg.Decimal(), row.TotalLei.Decimal(), row.PricePerKg.Decimal()
	p, err := calc.DerivePrice(kg, nil, &total)
	if err != nil {
		return nil, err
	}
	stored := fixed(&perKg, calc.PricePerKgPlaces)
	if derived := fixed(&p.PerKg, calc.PricePerKgPlaces); stored != derived {
		return nil, fmt.Errorf("the delivery of %s stores %s lei/kg, the derivation gives %s", priceDay, stored, derived)
	}
	return Price{Kg: fixed(&kg, 0), TotalLei: fixed(&p.Total, 0), PerKg: fixed(&p.PerKg, 2), StoredPerKg: stored}, nil
}

// cost reads golden 31: reports.BuildCost for the owner, every farm,
// period=cycle. It requires the golden farm, so an empty database is an
// error, not an empty block.
func cost(ctx context.Context, env Env) (any, error) {
	if _, err := FarmID(ctx, env.DB); err != nil {
		return nil, err
	}
	c, err := reports.BuildCost(ctx, env.DB, reports.Input{Principal: ownerAt, Today: AsOf, Now: Now})
	if err != nil {
		return nil, err
	}
	out := Cost{Cycles: make([]CycleCost, 0, len(c.Cycles))}
	for _, r := range c.Cycles {
		row := CycleCost{
			Cycle: r.Label, Lei: numFixed(&r.Lei, 0), Birds: r.Birds.F + r.Birds.M,
			PerBird: numFixed(r.LeiPerBird, 2), PerEgg: numFixed(r.LeiPerEgg, 2),
		}
		if r.Eggs != nil {
			row.Eggs = *r.Eggs
		}
		if r.Unpriced != nil {
			row.UnpricedKg = numFixed(&r.Unpriced.Kg, 0)
		}
		out.Cycles = append(out.Cycles, row)
	}
	if t := c.Total; t != nil {
		out.Total = CostTotal{
			Lei: numFixed(&t.Lei, 0), PerBird: numFixed(t.LeiPerBird, 2), PerEgg: numFixed(t.LeiPerEgg, 2),
			ExcludedKg: numFixed(&t.ExcludedKg, 0),
		}
	}
	return out, nil
}

// SeriesPlacement is golden 35: the series page's placement (§7.3) as the
// owner sees it: the placement farm (the farm of the series' first cycle),
// the birds placed in the series per sex, and the age weeks (SAPT) of the
// placement farm's cycle of the series on the cycle track.
type SeriesPlacement struct {
	Farm      string
	Placed    calc.PerSex[int]
	CycleSAPT [2]int
}

// seriesPlacement reads golden 35: series.BuildPage for the owner, the series
// of the golden farm's cycle on AsOf (Seria 2026), and the lane of the
// placement farm, which must hold exactly one cycle of the series.
func seriesPlacement(ctx context.Context, env Env) (any, error) {
	_, c, err := cycle(ctx, env.DB)
	if err != nil {
		return nil, err
	}
	pg, err := series.BuildPage(ctx, env.DB, series.Input{Principal: ownerAt, SeriesID: c.SeriesID, Today: AsOf, Now: Now})
	if err != nil {
		return nil, err
	}
	p := pg.Placement
	if p == nil || p.Placed == nil {
		return nil, fmt.Errorf("the series page of %s has no placed birds", pg.Series.Name)
	}
	out := SeriesPlacement{Farm: p.Farm, Placed: calc.PerSex[int]{F: p.Placed.F, M: p.Placed.M}}
	for _, l := range pg.Track.Lanes {
		if l.Farm.Name != p.Farm {
			continue
		}
		if n := len(l.Cycles); n != 1 {
			return nil, fmt.Errorf("%s has %d cycles of %s, want 1", p.Farm, n, pg.Series.Name)
		}
		out.CycleSAPT = l.Cycles[0].SaptSpan
		return out, nil
	}
	return nil, fmt.Errorf("the series page of %s has no lane for %s", pg.Series.Name, p.Farm)
}
