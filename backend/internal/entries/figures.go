package entries

import (
	"cmp"
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The figures a delete, an undo or a restore names and checks (M5 ARC-2):
// the bird move and the counts before it, §6.1 after the write, the feed
// stock after it.

// rowFigures are what removing (sign −1) or restoring (sign +1) a set of
// child rows does, as the messages, the audit row and the live events name
// it. The bird counts before are read before the write; the counts after
// are before + the move (every row is dated on the entry date ≤ today, so
// it moves every count from that day on by the same amount).
type rowFigures struct {
	birds       birdMove
	sections    []string // every section the rows name: bird sections, then egg units
	recordTypes []string // the S17 chips (Q5)
	halls       []string // egg units with an egg row
	eggs        int      // Σ cat1 + cat2 + cat3 + incubabile of those rows

	farm   calc.PerSex[int]               // the farm at the end of today, before
	today  map[uuid.UUID]calc.PerSex[int] // per section at the end of today, before
	atDate map[uuid.UUID]calc.PerSex[int] // per section at the end of the entry date, before (dry run)

	feedKg    map[uuid.UUID]decimal.Decimal // a feed save's kg per type (g × birds at the start of the day)
	feedTotal decimal.Decimal
	feedOf    int // the section × sex rows with birds that day ("{n} din {m} rânduri")
}

// birdMove is the bird rows of a removal or a restore.
type birdMove struct {
	pairs    []ledger.SectionSex       // in row order, once each
	codes    map[uuid.UUID]string      // section codes
	delta    map[ledger.SectionSex]int // the move of each pair's count
	move     calc.PerSex[int]          // Σ delta per sex: the farm's move
	qty      calc.PerSex[int]          // Σ stored signed qty per sex ("♀ −5 · ♂ 0")
	sexes    calc.PerSex[bool]         // a row of that sex
	sections []string                  // codes, in row order, once each
}

// readRowFigures reads the figures of rows: the farm and section counts
// for bird rows (the entry date too in a dry run, for the preview's alive
// rows), the kg of feed rows. sign is −1 for a removal, +1 for a restore.
func readRowFigures(ctx context.Context, tx *ChangeTx, rows Snapshot, sign int) (rowFigures, error) {
	e := tx.Entry
	f := rowFigures{birds: birdMoveOf(rows.BirdEvents, sign), recordTypes: []string{}}
	f.sections = slices.Clone(f.birds.sections)
	types := map[dbq.BirdEventType]bool{}
	for _, b := range rows.BirdEvents {
		types[b.Type] = true
	}
	for _, t := range []struct {
		typ   dbq.BirdEventType
		chips string
	}{{dbq.BirdEventTypeMortality, "mortalitate"}, {dbq.BirdEventTypeCull, "sacrificare"}, {dbq.BirdEventTypeSexingError, "eroare_sexare"}} {
		if types[t.typ] {
			f.recordTypes = append(f.recordTypes, t.chips)
		}
	}
	for _, g := range rows.EggProduction {
		f.halls = append(f.halls, g.Section)
		for _, n := range []*int32{g.Cat1, g.Cat2, g.Cat3, g.Hatching} {
			if n != nil {
				f.eggs += int(*n)
			}
		}
	}
	if len(rows.EggProduction) > 0 {
		f.recordTypes = append(f.recordTypes, "oua")
	}
	f.sections = append(f.sections, f.halls...)
	if len(rows.FeedAllocations) > 0 {
		f.recordTypes = append(f.recordTypes, recordTypeFeed)
		for _, r := range rows.FeedAllocations {
			if !slices.Contains(f.sections, r.Section) {
				f.sections = append(f.sections, r.Section)
			}
		}
	}
	if len(rows.FeedDeliveries) > 0 {
		f.recordTypes = append(f.recordTypes, recordTypeDelivery)
	}

	if len(f.birds.pairs) > 0 {
		day, err := calc.ReadFarmDay(ctx, tx.Tx, e.FarmID, tx.Today)
		if err != nil {
			return rowFigures{}, err
		}
		f.farm = day.AliveEnd
		if f.today, err = sectionsAliveAt(ctx, tx, tx.Today); err != nil {
			return rowFigures{}, err
		}
		f.atDate = f.today
		if tx.DryRun && e.Date != tx.Today {
			if f.atDate, err = sectionsAliveAt(ctx, tx, e.Date); err != nil {
				return rowFigures{}, err
			}
		}
	}
	if len(rows.FeedAllocations) > 0 {
		slots, err := loadFeedSlots(ctx, tx.Q, e.FarmID, e.Date)
		if err != nil {
			return rowFigures{}, err
		}
		for _, s := range slots {
			if fed(s) {
				f.feedOf++
			}
		}
		f.feedKg = map[uuid.UUID]decimal.Decimal{}
		for _, r := range rows.FeedAllocations {
			kg := decimal.Zero
			switch {
			case r.KgRecorded != nil:
				kg = r.KgRecorded.Decimal()
			case r.G != nil:
				kg = calc.RowFeedKg(r.G.Decimal(), slots[slotKey{r.SectionID, r.Sex}].AliveStart)
			}
			f.feedKg[r.FeedTypeID] = f.feedKg[r.FeedTypeID].Add(kg)
			f.feedTotal = f.feedTotal.Add(kg)
		}
	}
	return f, nil
}

func sectionsAliveAt(ctx context.Context, tx *ChangeTx, day timex.Day) (map[uuid.UUID]calc.PerSex[int], error) {
	rows, err := calc.SectionsAliveEnd(ctx, tx.Tx, tx.Entry.FarmID, day)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]calc.PerSex[int], len(rows))
	for _, r := range rows {
		out[r.SectionID] = r.AliveEnd
	}
	return out, nil
}

func birdMoveOf(rows []SnapshotBird, sign int) birdMove {
	m := birdMove{codes: map[uuid.UUID]string{}, delta: map[ledger.SectionSex]int{}}
	for _, b := range rows {
		pair := ledger.SectionSex{SectionID: b.SectionID, Sex: b.Sex}
		if _, ok := m.delta[pair]; !ok {
			m.pairs = append(m.pairs, pair)
		}
		m.delta[pair] += sign * int(b.Qty)
		m.codes[b.SectionID] = b.Section
		if !slices.Contains(m.sections, b.Section) {
			m.sections = append(m.sections, b.Section)
		}
		if b.Sex == dbq.SexF {
			m.move.F += sign * int(b.Qty)
			m.qty.F += int(b.Qty)
			m.sexes.F = true
		} else {
			m.move.M += sign * int(b.Qty)
			m.qty.M += int(b.Qty)
			m.sexes.M = true
		}
	}
	return m
}

// named are the sexes a Toast names: those whose count moves, else those
// with a row (a save of zeros).
func (m birdMove) named() calc.PerSex[bool] {
	if m.move.F != 0 || m.move.M != 0 {
		return calc.PerSex[bool]{F: m.move.F != 0, M: m.move.M != 0}
	}
	return m.sexes
}

// farmAfter is the farm's birds at the end of today after the change, nil
// when no bird row changed (the response's `alive`).
func (f rowFigures) farmAfter() *httpx.SexPair[int] {
	if len(f.birds.pairs) == 0 {
		return nil
	}
	return &httpx.SexPair[int]{F: f.farm.F + f.birds.move.F, M: f.farm.M + f.birds.move.M}
}

// sectionAfter is section id's birds at the end of today after the change.
func (f rowFigures) sectionAfter(id uuid.UUID) calc.PerSex[int] {
	before := f.today[id]
	return calc.PerSex[int]{
		F: before.F + f.birds.delta[ledger.SectionSex{SectionID: id, Sex: dbq.SexF}],
		M: before.M + f.birds.delta[ledger.SectionSex{SectionID: id, Sex: dbq.SexM}],
	}
}

// previewAlive are the preview's alive rows (Q7): per moved pair, the end of
// the entry date and of today, before and after.
func (f rowFigures) previewAlive(date, today timex.Day) []PreviewAlive {
	out := []PreviewAlive{}
	for _, p := range f.birds.pairs {
		d := f.birds.delta[p]
		if d == 0 {
			continue
		}
		days := []timex.Day{date}
		if today != date {
			days = append(days, today)
		}
		for _, day := range days {
			counts := f.today
			if day == date {
				counts = f.atDate
			}
			before := countOf(counts[p.SectionID], p.Sex)
			out = append(out, PreviewAlive{Section: f.birds.codes[p.SectionID], Sex: p.Sex, Date: day, Before: before, After: before + d})
		}
	}
	return out
}

// checkedDays is the days §6.1 checked (entry date to today), 0 when no
// bird moved.
func (f rowFigures) checkedDays(date, today timex.Day) int {
	if len(f.birds.pairs) == 0 {
		return 0
	}
	return today.DaysSince(date) + 1
}

// countsEvent is the counts.changed event of a change that moved birds.
func (f rowFigures) countsEvent(e LockedEntry) []live.Event {
	after := f.farmAfter()
	if after == nil {
		return nil
	}
	return []live.Event{live.CountsChanged(e.FarmID, e.Date, f.birds.sections, *after)}
}

func countOf(p calc.PerSex[int], sex dbq.Sex) int {
	if sex == dbq.SexM {
		return p.M
	}
	return p.F
}

// negativeRefusals is §6.1 after the write (plan M5-D4, Q4): the pairs whose
// count goes down, checked from the entry date; each negative pair is a
// refusal whose message is the banner of action ("Ștergerea ar lăsa H6.1 cu
// −5 găini azi, 18.11.2026. …"). A pair whose count only rises can't create
// a negative day, so it is not checked.
func negativeRefusals(ctx context.Context, tx *ChangeTx, m birdMove, action messages.CountAction) ([]httpx.Refusal, error) {
	var down []ledger.SectionSex
	for _, p := range m.pairs {
		if m.delta[p] < 0 {
			down = append(down, p)
		}
	}
	if len(down) == 0 {
		return nil, nil
	}
	negs, err := tx.CheckCounts(ctx, down, tx.Entry.Date)
	if err != nil {
		return nil, err
	}
	var out []httpx.Refusal
	for _, n := range negs {
		first, last, low := n.FirstDate, n.LastDate, n.MinAliveEnd
		out = append(out, httpx.Refusal{
			SectionCode: n.Code, Sex: string(n.Sex), FirstNegativeDate: &first, LastNegativeDate: &last, MinAliveEnd: &low,
			Message: messages.NegativeBanner(action, n.Code, n.Sex, n.FirstDate, n.FirstAliveEnd, n.TodayAliveEnd, tx.Today),
		})
	}
	return out, nil
}

// stockAfter is the feed stock a feed or delivery change leaves: the type a
// Toast names with its stock, and the STOCK_NEGATIVE warnings (H5, D40:
// saved anyway).
type stockAfter struct {
	typeName string
	kg       decimal.Decimal
	warnings []httpx.Warning
}

// readStockAfter reads, after the write, the stock of the types the rows
// name over [entry date, today] (calc.StockSeries); kgs is the rows' kg per
// type (rowFigures.feedKg); removing tells a delete or an undo (rows taken
// away) from a restore (rows put back):
//   - feed rows: the type with the most kg of the rows (the first by name on
//     a tie, as the create's Toast) at the end of today; on a restore, which
//     puts the consumption back, a warning per type whose stock dips below 0;
//   - a delivery: its type at the end of the delivery date (the create's
//     Toast, M4 Q13); on a delete or an undo, which takes the delivery away,
//     a warning when its stock dips below 0 on a day up to today.
func readStockAfter(ctx context.Context, tx *ChangeTx, rows Snapshot, kgs map[uuid.UUID]decimal.Decimal, farm string, removing bool) (stockAfter, error) {
	e := tx.Entry
	var out stockAfter
	switch {
	case len(rows.FeedAllocations) > 0:
		var ids []uuid.UUID
		names := map[uuid.UUID]string{}
		for _, r := range rows.FeedAllocations {
			if _, ok := names[r.FeedTypeID]; !ok {
				ids = append(ids, r.FeedTypeID)
				names[r.FeedTypeID] = r.Type
			}
		}
		series, err := calc.StockSeries(ctx, tx.Tx, e.FarmID, ids, e.Date, tx.Today)
		if err != nil {
			return stockAfter{}, err
		}
		slices.SortStableFunc(ids, func(a, b uuid.UUID) int { return cmp.Compare(names[a], names[b]) })
		best := ids[0]
		for _, id := range ids {
			if !removing {
				if low, ok := calc.Lowest(series[id]); ok && low.Kg.IsNegative() {
					out.warnings = append(out.warnings, httpx.Warning{
						Code: httpx.WarnStockNegative, Message: messages.StockWouldGoNegative(names[id], farm, low.Kg),
					})
				}
			}
			if kgs[id].GreaterThan(kgs[best]) {
				best = id
			}
		}
		out.typeName = names[best]
		if pts := series[best]; len(pts) > 0 {
			out.kg = pts[len(pts)-1].Kg
		}
	case len(rows.FeedDeliveries) > 0:
		d := rows.FeedDeliveries[0]
		series, err := calc.StockSeries(ctx, tx.Tx, e.FarmID, []uuid.UUID{d.FeedTypeID}, e.Date, tx.Today)
		if err != nil {
			return stockAfter{}, err
		}
		pts := series[d.FeedTypeID]
		out.typeName = d.Type
		if len(pts) > 0 {
			out.kg = pts[0].Kg
		}
		if removing {
			if low, ok := calc.Lowest(pts); ok && low.Kg.IsNegative() {
				out.warnings = append(out.warnings, httpx.Warning{
					Code: httpx.WarnStockNegative, Message: messages.StockWouldGoNegative(d.Type, farm, low.Kg),
				})
			}
		}
	}
	return out, nil
}

func warningsOrEmpty(w []httpx.Warning) []httpx.Warning {
	if w == nil {
		return []httpx.Warning{}
	}
	return w
}

// aliveMove is before → after, nil when unchanged (not named).
func aliveMove(before, after int) *messages.AliveMove {
	if before == after {
		return nil
	}
	return &messages.AliveMove{Before: before, After: after}
}

// namedAlive is the birds a Toast names after the change: the section's
// (per-section) or the farm's at the end of today, for the sexes named.
func (f rowFigures) namedAlive(sectionID *uuid.UUID) (aliveF, aliveM *int) {
	var after calc.PerSex[int]
	if sectionID != nil {
		after = f.sectionAfter(*sectionID)
	} else if a := f.farmAfter(); a != nil {
		after = calc.PerSex[int]{F: a.F, M: a.M}
	}
	named := f.birds.named()
	if named.F {
		aliveF = &after.F
	}
	if named.M {
		aliveM = &after.M
	}
	return aliveF, aliveM
}
