package reports

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Cursor is the keyset position of a Livrări page (M4 D6): the last row of
// the previous page in the row order (date desc, farm name_key, type
// name_key, id). On the wire it is opaque: base64url of a JSON array of the
// four values. It is not signed; a forged one only moves the page within the
// caller's own farms.
type Cursor struct {
	Date    timex.Day
	FarmKey string
	TypeKey string
	ID      uuid.UUID
}

// ErrCursor is returned by ParseCursor for a value that is not a cursor.
var ErrCursor = errors.New("reports: not a valid cursor")

// maxCursorLen bounds a cursor read from a query string.
const maxCursorLen = 512

// String encodes c for nextCursor.
func (c Cursor) String() string {
	b, err := json.Marshal([4]string{c.Date.String(), c.FarmKey, c.TypeKey, c.ID.String()})
	if err != nil { // four strings always encode
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ParseCursor decodes a cursor written by Cursor.String: base64url (no
// padding) of a JSON array of exactly four strings, a day, two keys and a
// uuid. Anything else is ErrCursor.
func ParseCursor(s string) (Cursor, error) {
	if s == "" || len(s) > maxCursorLen {
		return Cursor{}, ErrCursor
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, ErrCursor
	}
	var parts []string
	if err := json.Unmarshal(b, &parts); err != nil || len(parts) != 4 {
		return Cursor{}, ErrCursor
	}
	day, err := timex.ParseDay(parts[0])
	if err != nil {
		return Cursor{}, ErrCursor
	}
	id, err := uuid.Parse(parts[3])
	if err != nil {
		return Cursor{}, ErrCursor
	}
	return Cursor{Date: day, FarmKey: parts[1], TypeKey: parts[2], ID: id}, nil
}

// page is one page of Livrări rows as read, before it is shaped for a role.
type page struct {
	rows []dbq.ReportsDeliveryRowsRow
	next *string // the cursor of the next page; nil on the last one
}

// readRows reads up to limit rows of the ranges after the cursor (nil: from
// the newest), one more to know whether a next page exists.
func readRows(ctx context.Context, db dbq.DBTX, ranges []calc.FarmRange, after *Cursor, limit int) (page, error) {
	p := dbq.ReportsDeliveryRowsParams{RowLimit: int32(limit + 1)}
	p.FarmIds, p.HasFrom, p.FromDays, p.ToDays = rangeArrays(ranges)
	if after != nil {
		p.CursorDate, p.CursorFarmKey, p.CursorTypeKey, p.CursorID = &after.Date, &after.FarmKey, &after.TypeKey, &after.ID
	}
	rows, err := dbq.New(db).ReportsDeliveryRows(ctx, p)
	if err != nil {
		return page{rows: []dbq.ReportsDeliveryRowsRow{}}, fmt.Errorf("reports: delivery rows: %w", err)
	}
	pg := page{rows: rows}
	if len(rows) > limit {
		pg.rows = rows[:limit]
		last := rows[limit-1]
		next := Cursor{Date: last.Date, FarmKey: last.FarmKey, TypeKey: last.TypeKey, ID: last.ID}.String()
		pg.next = &next
	}
	return pg, nil
}

// rangeArrays spreads the ranges into the parallel arrays of the queries; a
// range without a lower bound has has_from false (its from day is ignored).
func rangeArrays(ranges []calc.FarmRange) (ids []uuid.UUID, hasFrom []bool, from, to []timex.Day) {
	ids, hasFrom = make([]uuid.UUID, len(ranges)), make([]bool, len(ranges))
	from, to = make([]timex.Day, len(ranges)), make([]timex.Day, len(ranges))
	for i, r := range ranges {
		ids[i], to[i], from[i] = r.FarmID, r.To, r.To
		if r.From != nil {
			hasFrom[i], from[i] = true, *r.From
		}
	}
	return ids, hasFrom, from, to
}

// view is the page's rows without prices (viewer, D2).
func (pg page) view() []DeliveryRow {
	out := make([]DeliveryRow, 0, len(pg.rows))
	for _, r := range pg.rows {
		out = append(out, deliveryRow(r))
	}
	return out
}

func deliveryRow(r dbq.ReportsDeliveryRowsRow) DeliveryRow {
	var forSex *string
	if r.ForSex != nil {
		s := string(*r.ForSex)
		forSex = &s
	}
	return DeliveryRow{ID: r.ID, Date: r.Date, Farm: r.FarmName, Type: r.TypeName, ForSex: forSex, Kg: num.Of(r.QtyKg)}
}

// priced is the page's rows with prices and the caller's actions (Q8): the
// owner sets an unpriced row's price and changes a priced one, unless the
// delivery is inside a cycle of a closed series at its farm (the row's
// series_closed, closed_series_on); a manager has neither.
func (pg page) priced(p auth.Principal) []DeliveryRowPriced {
	out := make([]DeliveryRowPriced, 0, len(pg.rows))
	for _, r := range pg.rows {
		row := DeliveryRowPriced{DeliveryRow: deliveryRow(r)}
		isPriced := r.TotalLei.Valid && r.PricePerKg.Valid
		if isPriced {
			perKg, total := num.Of(r.PricePerKg.Decimal), num.Of(r.TotalLei.Decimal)
			row.PricePerKg, row.TotalLei = &perKg, &total
			switch {
			case r.PriceFromRegister:
				by := messages.PricedFromRegister()
				row.PricedBy = &by
			case r.PricedByName != nil:
				by := *r.PricedByName
				row.PricedBy = &by
			}
			if r.PricedAt != nil && !r.PriceFromRegister {
				at := timex.At(*r.PricedAt)
				row.PricedAt = &at
			}
		} else {
			row.Tag = messages.Unpriced()
		}
		if p.Role == auth.Owner {
			open := !r.SeriesClosed
			row.Actions = DeliveryActions{SetPrice: open && !isPriced, ChangePrice: open && isPriced}
		}
		out = append(out, row)
	}
	return out
}

// summarize totals the farms' deliveries: the count, the per-farm counts
// (farms with at least one, by name) and the unpriced count, kg and distinct
// days (oldest first).
func summarize(sc scope, dels []calc.FarmDeliveries) (int, []FarmCount, UnpricedSummary) {
	total, byFarm := 0, []FarmCount{}
	kg, dates, count := decimal.Zero, []timex.Day{}, 0
	for i, d := range dels {
		total += d.Count
		if d.Count > 0 {
			byFarm = append(byFarm, FarmCount{Farm: sc.farms[i].Name, Count: d.Count})
		}
		count += d.Unpriced.Count
		kg = kg.Add(d.Unpriced.Kg)
		for _, day := range d.Unpriced.Dates {
			if !slices.Contains(dates, day) {
				dates = append(dates, day)
			}
		}
	}
	slices.SortFunc(dates, timex.Day.Compare)
	return total, byFarm, UnpricedSummary{Count: count, Kg: num.Of(kg), Dates: dates}
}

// readLastBefore reads the newest delivery dated before the ranges (only the
// ranges with a lower bound), nil when there is none.
func readLastBefore(ctx context.Context, db dbq.DBTX, ranges []calc.FarmRange) (*LastDelivery, error) {
	var p dbq.ReportsLastDeliveryBeforeParams
	p.FarmIds, p.FromDays = []uuid.UUID{}, []timex.Day{}
	for _, r := range ranges {
		if r.From != nil {
			p.FarmIds, p.FromDays = append(p.FarmIds, r.FarmID), append(p.FromDays, *r.From)
		}
	}
	if len(p.FarmIds) == 0 {
		return nil, nil
	}
	rows, err := dbq.New(db).ReportsLastDeliveryBefore(ctx, p)
	if err != nil {
		return nil, fmt.Errorf("reports: last delivery: %w", err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	r := rows[0]
	return &LastDelivery{Date: r.Date, Farm: r.FarmName, Type: r.TypeName, Kg: num.Of(r.QtyKg)}, nil
}
