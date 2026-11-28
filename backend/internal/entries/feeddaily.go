package entries

import (
	"context"
	"slices"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeFeedDaily is the Hrană zilnică create; the runner never shares a
// clientRequestId between routes.
const routeFeedDaily = "POST /api/entries/feed-daily"

// constraintFeedAllocation is the (section, day, sex) key of a live
// Hrană zilnică row.
const constraintFeedAllocation = "feed_allocation_uq"

// recordTypeFeed is the S17 chip of a feed save.
const recordTypeFeed = "hrana"

// CreateFeedDaily saves one Hrană zilnică form (§6.3, §7.8, M3 D5/D6) on the
// save runner. req must have passed Validate (level 1). Inside the
// transaction, after the farm lock and in this order:
//   - a future date → 422 REFUSED (§3.2); a vet on another day than today →
//     403 PAST_DAY_VET (H6); a day inside a cycle of a closed series → 409
//     SERIES_CLOSED (§3.1);
//   - per row, in document order: H3 (no birds at the start of the day, a
//     hall, a retired section) and H2 (inactive type, or for_sex ≠ the row's
//     sex) refusals; a (section, day, sex) already live → 409
//     ALREADY_ENTERED naming the first colliding row and its entry; else 422
//     REFUSED with every refusal; H4 warnings;
//   - the entry and its rows; H5 warnings and the stock at the end of today
//     per type (Q5) from one stock series over [date, today].
//
// A dry run answers 200 {effects: [], warnings, refusals}: the refusals in
// the body (ALREADY_ENTERED among them), or the warnings of the real save
// after its writes, rolled back.
func (s *Service) CreateFeedDaily(ctx context.Context, p auth.Principal, req FeedDailyCreateRequest) (Result, error) {
	if err := checkWriter(p, req.FarmID); err != nil {
		return Result{}, err
	}
	return s.runner.Create(ctx, p, Create{
		Route: routeFeedDaily, Kind: dbq.EntryKindFeedDaily, FarmIDs: []uuid.UUID{req.FarmID}, Meta: req.WriteMeta,
		OnConflict: func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError) error {
			if ce.Constraint != constraintFeedAllocation {
				return nil
			}
			return feedHolder(ctx, q, req)
		},
	}, func(ctx context.Context, tx *Tx) (Saved, error) {
		return planFeedDaily(ctx, tx, req)
	})
}

// feedType is a feed type of the save.
type feedType struct {
	ID     uuid.UUID
	Name   string
	ForSex *dbq.Sex
	Active bool
}

// feedContext is the farm context of a feed save (EntriesFeedSaveContext).
type feedContext struct {
	FarmName     string
	ClosedSeries string                          // closed_series_on(farm, date); "" for none (M5 D12)
	GuideG       map[dbq.Sex]decimal.NullDecimal // feed_g of SAPT(date); invalid = no H4
	TolerancePct decimal.Decimal
	Types        map[uuid.UUID]feedType
	TypeOrder    []uuid.UUID // the types by name_key
}

// slotKey is one section × sex.
type slotKey struct {
	Section uuid.UUID
	Sex     dbq.Sex
}

// loadFeedContext reads the save's farm context; ok is false for an unknown
// farm.
func loadFeedContext(ctx context.Context, q *dbq.Queries, req FeedDailyCreateRequest) (feedContext, bool, error) {
	typeIDs := make([]uuid.UUID, 0, len(req.Rows))
	for _, r := range req.Rows {
		typeIDs = append(typeIDs, r.FeedTypeID)
	}
	rows, err := q.EntriesFeedSaveContext(ctx, dbq.EntriesFeedSaveContextParams{
		FarmID: req.FarmID, Day: req.Date, FeedTypeIds: typeIDs,
	})
	if err != nil || len(rows) == 0 {
		return feedContext{}, false, err
	}
	first := rows[0]
	fc := feedContext{
		FarmName: first.FarmName, ClosedSeries: first.ClosedSeries,
		GuideG:       map[dbq.Sex]decimal.NullDecimal{dbq.SexF: first.GuideFeedGF, dbq.SexM: first.GuideFeedGM},
		TolerancePct: first.ToleranceFeedPct,
		Types:        make(map[uuid.UUID]feedType, len(rows)),
	}
	for _, r := range rows {
		if r.FeedTypeID == nil || r.FeedTypeName == nil || r.FeedTypeActive == nil {
			continue
		}
		fc.Types[*r.FeedTypeID] = feedType{ID: *r.FeedTypeID, Name: *r.FeedTypeName, ForSex: r.FeedTypeForSex, Active: *r.FeedTypeActive}
		fc.TypeOrder = append(fc.TypeOrder, *r.FeedTypeID)
	}
	return fc, true, nil
}

// loadFeedSlots reads every section × sex of the farm on the day.
func loadFeedSlots(ctx context.Context, q *dbq.Queries, farmID uuid.UUID, day timex.Day) (map[slotKey]dbq.EntriesFeedDaySlotsRow, error) {
	rows, err := q.EntriesFeedDaySlots(ctx, dbq.EntriesFeedDaySlotsParams{FarmID: farmID, Day: day})
	if err != nil {
		return nil, err
	}
	out := make(map[slotKey]dbq.EntriesFeedDaySlotsRow, len(rows))
	for _, r := range rows {
		out[slotKey{r.SectionID, r.Sex}] = r
	}
	return out, nil
}

// isFeedRow is the one rule of a Hrană zilnică row (§6.3 H3, M3 D5): a
// section that holds birds (not a hall), not retired, with birds at the start
// of the day. The form state lists exactly these rows (formRows) and the save
// accepts exactly these (fed); TestFormRowsAreSaveSlots ties the two reads.
func isFeedRow(isHall, retired bool, birdsStart int32) bool {
	return !isHall && !retired && birdsStart > 0
}

// fed reports whether a slot is a row of the form (isFeedRow).
func fed(s dbq.EntriesFeedDaySlotsRow) bool { return isFeedRow(s.IsHall, s.Retired, s.AliveStart) }

// feedCollision is the first row of a save whose key is already live.
type feedCollision struct {
	EntryID uuid.UUID
	Message string
}

// collisionOf is the ALREADY_ENTERED of a slot held by another entry, or
// false. Its sentence names the holder's creation time; a holder without an
// author (the history import) has no time worth naming (its import instant),
// so it gets messages.EntryAlreadySaved (M3 COR-1).
func collisionOf(slot dbq.EntriesFeedDaySlotsRow, day timex.Day) (feedCollision, bool) {
	if slot.HolderEntryID == nil {
		return feedCollision{}, false
	}
	migrated := slot.HolderSource != nil && *slot.HolderSource == dbq.RecordSourceMigration
	if slot.HolderCreatedBy == nil || migrated || slot.HolderCreatedAt == nil {
		return feedCollision{EntryID: *slot.HolderEntryID, Message: messages.EntryAlreadySaved()}, true
	}
	return feedCollision{
		EntryID: *slot.HolderEntryID,
		Message: messages.FeedAlreadyEntered(slot.Code, slot.Sex, day, *slot.HolderCreatedAt),
	}, true
}

// planFeedDaily is the form's part of the save (§12.5 last note).
func planFeedDaily(ctx context.Context, tx *Tx, req FeedDailyCreateRequest) (Saved, error) {
	refuse := func(refusals []httpx.Refusal, warnings []httpx.Warning) (Saved, error) {
		if tx.DryRun {
			return Saved{}, tx.Preview(NewDryRun(nil, warnings, refusals))
		}
		return Saved{}, httpx.RefusedWith(refusals, warnings)
	}
	if len(req.Rows) == 0 { // Validate guarantees at least one row
		return Saved{}, httpx.Validation(httpx.Issue{Path: "rows", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	if req.Date.After(tx.Today) { // §3.2, every role
		return refuse([]httpx.Refusal{{Field: "date", Message: messages.FutureDate(req.Date, tx.Today)}}, nil)
	}
	if tx.Principal.Role == auth.Vet && req.Date != tx.Today { // H6
		return Saved{}, httpx.PastDayVet()
	}
	fc, ok, err := loadFeedContext(ctx, tx.Q, req)
	if err != nil {
		return Saved{}, err
	}
	if !ok {
		return Saved{}, httpx.NotFound()
	}
	if fc.ClosedSeries != "" {
		return Saved{}, httpx.SeriesClosed(fc.ClosedSeries)
	}
	slots, err := loadFeedSlots(ctx, tx.Q, req.FarmID, req.Date)
	if err != nil {
		return Saved{}, err
	}

	var (
		refusals  []httpx.Refusal
		warnings  = []httpx.Warning{}
		collision *feedCollision
		kgs       = make([]decimal.Decimal, len(req.Rows))
	)
	for i, row := range req.Rows {
		if row.G == nil { // Validate guarantees it; never insert a row without g
			return Saved{}, httpx.Validation(httpx.Issue{
				Path: "rows[" + strconv.Itoa(i) + "].g", Code: httpx.IssueRequired, Message: messages.InvalidData(),
			})
		}
		path := "rows[" + strconv.Itoa(i) + "]"
		slot, known := slots[slotKey{row.SectionID, row.Sex}]
		if !known || !fed(slot) { // H3
			refusals = append(refusals, httpx.Refusal{
				Field: path + ".sectionId", SectionCode: slot.Code, Sex: string(row.Sex),
				Message: messages.NoBirdsAtDayStart(slot.Code, row.Sex, req.Date),
			})
		}
		switch ft, ok := fc.Types[row.FeedTypeID]; { // H2
		case !ok:
			refusals = append(refusals, httpx.Refusal{
				Field: path + ".feedTypeId", SectionCode: slot.Code, Sex: string(row.Sex), Message: messages.InvalidData(),
			})
		case !ft.Active:
			refusals = append(refusals, httpx.Refusal{
				Field: path + ".feedTypeId", SectionCode: slot.Code, Sex: string(row.Sex),
				Message: messages.FeedTypeRetired(ft.Name),
			})
		case ft.ForSex != nil && *ft.ForSex != row.Sex:
			refusals = append(refusals, httpx.Refusal{
				Field: path + ".feedTypeId", SectionCode: slot.Code, Sex: string(row.Sex),
				Message: messages.FeedTypeWrongSex(ft.Name, row.Sex),
			})
		}
		if c, held := collisionOf(slot, req.Date); known && held {
			refusals = append(refusals, httpx.Refusal{
				Field: path, SectionCode: slot.Code, Sex: string(row.Sex), Message: c.Message,
			})
			if collision == nil {
				collision = &c
			}
		}
		if !known || !fed(slot) {
			continue
		}
		g := row.G.Decimal()
		kgs[i] = calc.RowFeedKg(g, slot.AliveStart)
		if w, ok := feedGuideWarning(fc, slot.Code, row.Sex, g); ok {
			warnings = append(warnings, w)
		}
	}
	if collision != nil && !tx.DryRun {
		return Saved{}, httpx.AlreadyEntered(collision.EntryID, collision.Message)
	}
	if len(refusals) > 0 {
		return refuse(refusals, warnings)
	}

	ins, err := tx.InsertEntry(ctx, Header{FarmID: req.FarmID, Date: req.Date})
	if err != nil {
		return Saved{}, err
	}
	kgByType := make(map[uuid.UUID]decimal.Decimal, len(fc.TypeOrder))
	kgTotal := decimal.Zero
	snapshot := make([]feedAuditRow, 0, len(req.Rows))
	var sections []string
	for i, row := range req.Rows {
		slot := slots[slotKey{row.SectionID, row.Sex}]
		if err := tx.Q.EntriesInsertFeedAllocation(ctx, dbq.EntriesInsertFeedAllocationParams{
			EntryID: ins.ID, FarmID: req.FarmID, SectionID: row.SectionID, Date: req.Date, Sex: row.Sex,
			FeedTypeID: row.FeedTypeID, GPerBird: decimal.NewNullDecimal(row.G.Decimal()),
		}); err != nil {
			return Saved{}, err
		}
		kgByType[row.FeedTypeID] = kgByType[row.FeedTypeID].Add(kgs[i])
		kgTotal = kgTotal.Add(kgs[i])
		snapshot = append(snapshot, feedAuditRow{
			Section: slot.Code, Sex: row.Sex, FeedTypeID: row.FeedTypeID, G: *row.G, Kg: num.Of(kgs[i]),
		})
		if !slices.Contains(sections, slot.Code) {
			sections = append(sections, slot.Code)
		}
	}

	// H5 and the stock after the save: the series over [date, today] after
	// the inserts; its last point is the stock at the end of today (Q5).
	series, err := calc.StockSeries(ctx, tx.Tx, req.FarmID, fc.TypeOrder, req.Date, tx.Today)
	if err != nil {
		return Saved{}, err
	}
	byType := make([]FeedTypeKg, 0, len(fc.TypeOrder))
	toastType := -1
	for _, id := range fc.TypeOrder {
		ft, pts := fc.Types[id], series[id]
		stockAfter := decimal.Zero
		if len(pts) > 0 {
			stockAfter = pts[len(pts)-1].Kg
		}
		if low, ok := calc.Lowest(pts); ok && low.Kg.IsNegative() {
			warnings = append(warnings, httpx.Warning{
				Code: httpx.WarnStockNegative, Message: messages.StockWouldGoNegative(ft.Name, fc.FarmName, low.Kg),
			})
		}
		kg := kgByType[id]
		if toastType < 0 || kg.GreaterThan(byType[toastType].Kg.Decimal()) {
			toastType = len(byType)
		}
		byType = append(byType, FeedTypeKg{Type: ft.Name, Kg: num.Of(kg), StockAfter: num.Of(stockAfter)})
	}
	if tx.DryRun {
		return Saved{}, tx.Preview(NewDryRun(nil, warnings, nil))
	}

	of := 0 // the section × sex rows with birds that day
	for _, s := range slots {
		if fed(s) {
			of++
		}
	}
	summary := messages.FeedDailySummary(len(req.Rows), of, kgTotal)
	at := timex.At(tx.Now)
	return Saved{
		Response: FeedDailyResponse{
			Entry: ins.Ref(), KgTotal: num.Of(kgTotal), ByType: byType,
			Toast: messages.FeedDailySaved(len(req.Rows), kgTotal, byType[toastType].Type,
				byType[toastType].StockAfter.Decimal()),
			Warnings: warnings, AsOf: at,
		},
		Audit: audit.Row{
			Action: dbq.AuditActionCreate, Entity: "entry", EntityID: ins.ID, EntryID: &ins.ID, EntryKind: &ins.Kind,
			RecordTypes: []string{recordTypeFeed}, FarmIDs: []uuid.UUID{req.FarmID}, RecordDate: &req.Date,
			Summary: &summary, NewValues: snapshot,
		},
		Events: []live.Event{live.EntryCreated(live.EntryPayload{
			EntryID: ins.ID, Kind: ins.Kind, FarmIDs: []uuid.UUID{req.FarmID}, Dates: []timex.Day{req.Date},
			Sections: sections, RecordTypes: []string{recordTypeFeed}, Author: &tx.Principal.Name, At: at,
			Summary: summary,
		})},
	}, nil
}

// feedGuideWarning is H4: |g / guide − 1| > tolerance.feed_pct at full
// precision. No guide value (no guide, no guide row, an empty or 0 cell) →
// no warning.
func feedGuideWarning(fc feedContext, section string, sex dbq.Sex, g decimal.Decimal) (httpx.Warning, bool) {
	guide := fc.GuideG[sex]
	if !guide.Valid || guide.Decimal.Sign() <= 0 {
		return httpx.Warning{}, false
	}
	diffPct := g.Div(guide.Decimal).Sub(decimal.NewFromInt(1)).Mul(decimal.NewFromInt(100))
	if diffPct.Abs().Cmp(fc.TolerancePct) <= 0 {
		return httpx.Warning{}, false
	}
	return httpx.Warning{
		Code: httpx.WarnFeedGuide, SectionCode: section, Sex: string(sex),
		Message: messages.FeedGuideDeviation(section, sex, g, guide.Decimal, diffPct),
	}, true
}

// feedHolder answers a feed_allocation_uq violation after the transaction
// (M3 D2): 409 ALREADY_ENTERED naming the first row of req whose key is now
// held, with its entry. nil when none is held any more (the violation stays
// a 500).
func feedHolder(ctx context.Context, q *dbq.Queries, req FeedDailyCreateRequest) error {
	slots, err := loadFeedSlots(ctx, q, req.FarmID, req.Date)
	if err != nil {
		return err
	}
	for _, row := range req.Rows {
		if slot, ok := slots[slotKey{row.SectionID, row.Sex}]; ok {
			if c, held := collisionOf(slot, req.Date); held {
				return httpx.AlreadyEntered(c.EntryID, c.Message)
			}
		}
	}
	return nil
}
