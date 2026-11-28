package calc

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// DayRange is a range of days, both ends included. A nil From has no lower
// bound (a farm's first cycle takes every delivery before it, §5.9).
type DayRange struct {
	From *timex.Day
	To   timex.Day
}

// Contains reports whether d is one of the range's days.
func (r DayRange) Contains(d timex.Day) bool {
	return (r.From == nil || !d.Before(*r.From)) && !d.After(r.To)
}

// DeliveryRange is the day range of a farm's Livrări and Cost (§5.9, §7.12,
// M4 Q6) for the period p, with shown the farm's shown cycle (open, else
// latest: DefaultCycle; nil when the farm has none) and cycles every cycle of
// the farm (FarmCycles):
//   - cycle: (the end of the farm's previous cycle, shown's end or today], so
//     the pre-stocking deliveries before the first birds belong to the cycle
//     and a delivery on the previous cycle's end day belongs to that one; no
//     previous cycle → no lower bound;
//   - weeks:n: the days of the cycle's resolved weeks (ResolvePeriod);
//   - interval: its days as typed, not clipped to the cycle (an end after
//     today is ErrIntervalFuture).
//
// A farm without any cycle takes every delivery up to today for cycle and
// weeks:n (there are no weeks to count), and the interval as typed.
func DeliveryRange(cycles []CycleInfo, shown *CycleInfo, p Period, today timex.Day) (DayRange, error) {
	switch p.Kind {
	case PeriodInterval:
		if p.From.IsZero() || p.To.IsZero() || p.From.After(p.To) {
			return DayRange{}, fmt.Errorf("%w: interval %s–%s", ErrPeriod, p.From, p.To)
		}
		if p.To.After(today) {
			return DayRange{}, ErrIntervalFuture
		}
		from := p.From
		return DayRange{From: &from, To: p.To}, nil
	case PeriodCycle, PeriodWeeks:
		if p.Kind == PeriodWeeks && p.Weeks < 1 {
			return DayRange{}, fmt.Errorf("%w: weeks:%d", ErrPeriod, p.Weeks)
		}
	default:
		return DayRange{}, fmt.Errorf("%w: kind %q", ErrPeriod, p.Kind)
	}
	if shown == nil {
		return DayRange{To: today}, nil
	}
	if p.Kind == PeriodWeeks {
		r, err := ResolvePeriod(shown.Cycle, p, today)
		if err != nil {
			return DayRange{}, err
		}
		from := r.From
		return DayRange{From: &from, To: r.To}, nil
	}
	to := today
	if shown.EndDate != nil && shown.EndDate.Before(today) {
		to = *shown.EndDate
	}
	out := DayRange{To: to}
	if prev := previousEnd(cycles, shown.Cycle); prev != nil {
		from := prev.AddDays(1)
		out.From = &from
	}
	return out, nil
}

// previousEnd is the latest end day of the farm's cycles that ended before c
// started, nil when none did.
func previousEnd(cycles []CycleInfo, c Cycle) *timex.Day {
	var end *timex.Day
	for _, o := range cycles {
		if o.ID == c.ID || o.EndDate == nil || !o.EndDate.Before(c.StartDate) {
			continue
		}
		if end == nil || o.EndDate.After(*end) {
			e := *o.EndDate
			end = &e
		}
	}
	return end
}

// FarmRange is one farm's day range for ReadDeliveries.
type FarmRange struct {
	FarmID uuid.UUID
	DayRange
}

// UnpricedDelivery is one live delivery without a price ("fără preț").
type UnpricedDelivery struct {
	ID         uuid.UUID
	Date       timex.Day
	FeedTypeID uuid.UUID
	Type       string
	Kg         decimal.Decimal
}

// UnpricedDeliveries are a range's deliveries without a price (§5.9).
type UnpricedDeliveries struct {
	Count int
	Kg    decimal.Decimal
	// Dates are the distinct delivery days, oldest first.
	Dates []timex.Day
	// Items are the deliveries, oldest first, then by type name.
	Items []UnpricedDelivery
}

// Span is the first and last day of the unpriced deliveries, false when
// there are none ("17.02–12.11.2026", §5.9).
func (u UnpricedDeliveries) Span() (first, last timex.Day, ok bool) {
	if len(u.Dates) == 0 {
		return timex.Day{}, timex.Day{}, false
	}
	return u.Dates[0], u.Dates[len(u.Dates)-1], true
}

// PricedDeliveries are a range's deliveries with a price (§5.9 lei(cycle)).
type PricedDeliveries struct {
	Count int
	Kg    decimal.Decimal // kg_priced
	Lei   decimal.Decimal // Σ total_lei
}

// FarmDeliveries are one farm's live deliveries over its range (§5.9, §7.12).
type FarmDeliveries struct {
	FarmRange
	Count int
	// Kg is the delivered kg by the feed type's for_sex: F (găini), M
	// (cocoși); Both holds the "ambele sexe" types.
	Kg       PerSex[decimal.Decimal]
	KgBoth   decimal.Decimal
	Priced   PricedDeliveries
	Unpriced UnpricedDeliveries
}

// TotalKg is every delivered kg of the range, priced or not.
func (d FarmDeliveries) TotalKg() decimal.Decimal { return d.Kg.F.Add(d.Kg.M).Add(d.KgBoth) }

// ReadDeliveries reads the live deliveries of every asked farm over its range
// in one statement (§5.9, M4 D6), one result per range in the order asked.
// Soft-deleted deliveries and those of a deleted entry never count. A range
// that starts after it ends is an error.
func ReadDeliveries(ctx context.Context, db dbq.DBTX, ranges []FarmRange) ([]FarmDeliveries, error) {
	if len(ranges) == 0 {
		return []FarmDeliveries{}, nil
	}
	p := dbq.CalcDeliveriesParams{
		FarmIds:  make([]uuid.UUID, len(ranges)),
		HasFrom:  make([]bool, len(ranges)),
		FromDays: make([]timex.Day, len(ranges)),
		ToDays:   make([]timex.Day, len(ranges)),
	}
	for i, r := range ranges {
		p.FarmIds[i], p.ToDays[i] = r.FarmID, r.To
		p.FromDays[i] = r.To // ignored without HasFrom ([]timex.Day holds no NULL)
		if r.From != nil {
			if r.From.After(r.To) {
				return nil, errRange
			}
			p.HasFrom[i], p.FromDays[i] = true, *r.From
		}
	}
	rows, err := dbq.New(db).CalcDeliveries(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("calc: deliveries: %w", err)
	}
	if len(rows) != len(ranges) {
		return nil, fmt.Errorf("calc: deliveries: %d rows for %d ranges", len(rows), len(ranges))
	}
	out := make([]FarmDeliveries, len(rows))
	for i, r := range rows {
		n := len(r.UnpricedIds)
		if len(r.UnpricedDates) != n || len(r.UnpricedKgs) != n || len(r.UnpricedTypeIds) != n || len(r.UnpricedTypes) != n {
			return nil, fmt.Errorf("calc: deliveries: unpriced arrays of different lengths")
		}
		u := UnpricedDeliveries{Count: n, Kg: decimal.Zero, Dates: []timex.Day{}, Items: make([]UnpricedDelivery, n)}
		for j := range n {
			u.Items[j] = UnpricedDelivery{
				ID: r.UnpricedIds[j], Date: r.UnpricedDates[j], FeedTypeID: r.UnpricedTypeIds[j],
				Type: r.UnpricedTypes[j], Kg: r.UnpricedKgs[j],
			}
			u.Kg = u.Kg.Add(r.UnpricedKgs[j])
			if !slices.Contains(u.Dates, r.UnpricedDates[j]) { // oldest first: a new day is always the last
				u.Dates = append(u.Dates, r.UnpricedDates[j])
			}
		}
		out[i] = FarmDeliveries{
			FarmRange: ranges[r.Ord-1],
			Count:     int(r.Deliveries),
			Kg:        PerSex[decimal.Decimal]{F: r.KgF, M: r.KgM},
			KgBoth:    r.KgBoth,
			Priced:    PricedDeliveries{Count: int(r.PricedCount), Kg: r.PricedKg, Lei: r.PricedLei},
			Unpriced:  u,
		}
	}
	return out, nil
}
