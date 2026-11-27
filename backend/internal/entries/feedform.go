package entries

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Query parameters of the form-state routes.
const (
	paramFarmID = "farmId"
	paramDate   = "date"
)

// feedForm is GET /api/forms/feed-daily?farmId&date (§7.8, M3 D8). Both
// parameters are required; any other parameter is a 400 VALIDATION issue.
func (h *Handler) feedForm(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	farmID, date, err := parseFormQuery(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.FeedDailyForm(r.Context(), p, farmID, date)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseFormQuery reads ?farmId&date. Every problem is a 400 VALIDATION issue
// (§7.0a codes): a missing parameter → required, a repeated one or a
// malformed uuid → bad_type, a malformed day → bad_date, an unknown
// parameter → unknown_field (in name order).
func parseFormQuery(values url.Values) (uuid.UUID, timex.Day, error) {
	var v httpx.Validator
	var farmID uuid.UUID
	var date timex.Day
	if s, ok := requiredParam(&v, values, paramFarmID); ok {
		id, err := uuid.Parse(s)
		if v.Check(err == nil, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
			farmID = id
		}
	}
	if s, ok := requiredParam(&v, values, paramDate); ok {
		d, err := timex.ParseDay(s)
		if v.Check(err == nil, paramDate, httpx.IssueBadDate, messages.InvalidData()) {
			date = d
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(name == paramFarmID || name == paramDate, name, httpx.IssueUnknownField, messages.InvalidData())
	}
	return farmID, date, v.Err()
}

// requiredParam is the one value of parameter name: a missing one is a
// required issue, a repeated one a bad_type issue.
func requiredParam(v *httpx.Validator, values url.Values, name string) (string, bool) {
	vals, ok := values[name]
	if !v.Check(ok, name, httpx.IssueRequired, messages.InvalidData()) {
		return "", false
	}
	if !v.Check(len(vals) == 1, name, httpx.IssueBadType, messages.InvalidData()) {
		return "", false
	}
	return vals[0], true
}

// FeedDailyForm is the Hrană zilnică form state of farmID on date for p
// (§7.8, M3 D8), read in one snapshot (InReadTx, M3 D13). Errors are
// envelope errors: 403 FORBIDDEN for a viewer, 403 FARM_FORBIDDEN for a farm
// outside the caller's farms, 400 VALIDATION on `date` for a day after today
// (the §3.2 sentence), 404 NOT_FOUND for an owner's unknown farm. A vet's
// past day and a closed series are not errors: the form comes back with its
// lockedReason.
func (s *Service) FeedDailyForm(ctx context.Context, p auth.Principal, farmID uuid.UUID, date timex.Day) (FeedFormResponse, error) {
	if err := checkWriter(p, farmID); err != nil {
		return FeedFormResponse{}, err
	}
	if !p.CanSeeFarm(farmID) { // checkWriter leaves the nil farm id to a create's level 1
		return FeedFormResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	if date.After(today) {
		return FeedFormResponse{}, httpx.Validation(httpx.Issue{
			Path: paramDate, Code: httpx.IssueOutOfRange, Message: messages.FutureDate(date, today),
		})
	}
	var res FeedFormResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildFeedForm(ctx, tx, FeedFormInput{Principal: p, FarmID: farmID, Date: date, Today: today, Now: now})
		return err
	})
	if errors.Is(err, ErrFormFarmNotFound) {
		return FeedFormResponse{}, httpx.NotFound()
	}
	return res, err
}

// ErrFormFarmNotFound: the form's farm doesn't exist.
var ErrFormFarmNotFound = errors.New("entries: farm not found")

// FeedFormInput is what BuildFeedForm reads: the caller (for the lock and the
// entries' actions), the farm and day of the form, and the clock's day and
// instant (ARC-6: never app_today()).
type FeedFormInput struct {
	Principal auth.Principal
	FarmID    uuid.UUID
	Date      timex.Day
	Today     timex.Day
	Now       time.Time
}

// BuildFeedForm builds GET /api/forms/feed-daily (§7.8, M3 D8, Q1) on db
// for in.FarmID on in.Date. It reads, in this order: the head (farm, cycle
// on the date, guide g of SAPT(date), active feed types), the section × sex
// rows of the day, and, for a farm with an open cycle, the settings and the
// stock as of in.Today (calc.ReadStockDays, Q5): at most 6 statements, to be
// run in one read snapshot (the service's InReadTx).
//
// It doesn't check the caller's role, farm or date (the service does); an
// unknown farm is ErrFormFarmNotFound.
func BuildFeedForm(ctx context.Context, db dbq.DBTX, in FeedFormInput) (FeedFormResponse, error) {
	q := dbq.New(db)
	head, err := q.FormsFeedHead(ctx, dbq.FormsFeedHeadParams{FarmID: in.FarmID, Day: in.Date})
	if errors.Is(err, pgx.ErrNoRows) {
		return FeedFormResponse{}, ErrFormFarmNotFound
	}
	if err != nil {
		return FeedFormResponse{}, fmt.Errorf("entries: feed form head: %w", err)
	}
	rows, err := q.FormsFeedRows(ctx, dbq.FormsFeedRowsParams{FarmID: in.FarmID, Day: in.Date})
	if err != nil {
		return FeedFormResponse{}, fmt.Errorf("entries: feed form rows: %w", err)
	}
	res := FeedFormResponse{
		Farm: FormFarm{ID: head.ID, Name: head.Name, Type: head.Type},
		Date: in.Date, AsOf: timex.At(in.Now),
		Rows: []FeedFormRow{}, EmptyHalls: []FormEmptyHall{}, Existing: []FormEntry{}, Stock: []FormStock{},
		FeedTypes: feedTypes(head),
	}
	if head.CycleID != nil {
		age := calc.SeriesAge{PlacementDate: *head.PlacementDate, AgeAtPlacementDays: int(*head.AgeAtPlacementDays)}
		sapt, ageDays := calc.SAPT(age, in.Date), calc.AgeDays(age, in.Date)
		res.SAPT, res.AgeDays = &sapt, &ageDays
		res.Cycle = &FormCycle{ID: *head.CycleID, Series: *head.SeriesName, StartDate: *head.CycleStart}
		if head.GuideID != nil {
			res.Guide = &httpx.SexPair[*num.Num]{F: nullNum(head.GuideF), M: nullNum(head.GuideM)}
		}
	}
	res.LockedReason = lockedReason(in, head)
	res.Rows = formRows(rows, res.Guide)
	res.EmptyHalls = emptyHalls(rows, res.Cycle)
	res.Existing = formEntries(in, head, rows)
	if head.OpenCycle {
		if res.Stock, err = formStock(ctx, db, in); err != nil {
			return FeedFormResponse{}, err
		}
	}
	return res, nil
}

// lockedReason is why the caller can't save the day, checked as the save
// checks it (M3 D5): a vet on a day other than today, then a closed series.
func lockedReason(in FeedFormInput, head dbq.FormsFeedHeadRow) *string {
	var msg string
	switch {
	case in.Principal.Role == auth.Vet && in.Date != in.Today:
		msg = messages.PastDayVet()
	case head.SeriesClosed && head.SeriesName != nil:
		msg = messages.SeriesClosed(*head.SeriesName)
	default:
		return nil
	}
	return &msg
}

// feedTypes are the head's active feed types, in name order.
func feedTypes(head dbq.FormsFeedHeadRow) []FormFeedType {
	out := make([]FormFeedType, 0, len(head.TypeIds))
	for i, id := range head.TypeIds {
		out = append(out, FormFeedType{ID: id, Name: head.TypeNames[i], ForSex: forSex(head.TypeSexes[i])})
	}
	return out
}

// forSex is a feed type's for_sex as the wire has it: nil for ambele sexe.
func forSex(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// formRows are the active section × sex rows with birds at the start of the
// day, in the query's tree order.
func formRows(rows []dbq.FormsFeedRowsRow, guide *httpx.SexPair[*num.Num]) []FeedFormRow {
	out := []FeedFormRow{}
	for _, r := range rows {
		if !isFeedRow(false, r.Retired, r.BirdsStart) { // FormsFeedRows reads bird-holding sections only
			continue
		}
		row := FeedFormRow{
			SectionID: r.SectionID, Code: r.Code, HallCode: r.HallCode, Sex: r.Sex, BirdsStart: int(r.BirdsStart),
			Suggestion: FeedSuggestion{G: nullNum(r.YesterdayG), FeedTypeID: r.LastTypeID},
		}
		if guide != nil {
			row.GuideG = guide.F
			if r.Sex == dbq.SexM {
				row.GuideG = guide.M
			}
		}
		row.Existing = rowExisting(r)
		out = append(out, row)
	}
	return out
}

// rowExisting is the row's live allocation of the day, nil when none. Its kg
// is the migrated kg, else calc.RowFeedKg (g × birds at the start ÷ 1000, as
// the save and calc's readers).
func rowExisting(r dbq.FormsFeedRowsRow) *FeedRowExisting {
	if r.ExistingEntryID == nil || r.ExistingTypeID == nil {
		return nil
	}
	kg := decimal.Zero
	switch {
	case r.ExistingKgRecorded.Valid:
		kg = r.ExistingKgRecorded.Decimal
	case r.ExistingG.Valid:
		kg = calc.RowFeedKg(r.ExistingG.Decimal, r.BirdsStart)
	}
	return &FeedRowExisting{EntryID: *r.ExistingEntryID, G: nullNum(r.ExistingG), FeedTypeID: *r.ExistingTypeID, Kg: num.Of(kg)}
}

// emptyHalls are the halls with no bird at the end of the day that were
// emptied on or after the start of the date's cycle (as the summary's tiles,
// golden 38), in tree order, each with its bird-holding sections.
func emptyHalls(rows []dbq.FormsFeedRowsRow, cycle *FormCycle) []FormEmptyHall {
	out := []FormEmptyHall{}
	if cycle == nil {
		return out
	}
	type hall struct {
		row   dbq.FormsFeedRowsRow // its first row: the hall's code, count and empty_since
		codes []string
	}
	var halls []*hall
	byID := map[uuid.UUID]*hall{}
	seen := map[uuid.UUID]bool{}
	for _, r := range rows {
		h, ok := byID[r.HallID]
		if !ok {
			h = &hall{row: r}
			byID[r.HallID] = h
			halls = append(halls, h)
		}
		if !seen[r.SectionID] { // rows come per section × sex
			seen[r.SectionID] = true
			h.codes = append(h.codes, r.Code)
		}
	}
	for _, h := range halls {
		since := h.row.HallEmptySince
		if h.row.HallAliveEnd == 0 && since != nil && !since.Before(cycle.StartDate) {
			out = append(out, FormEmptyHall{Hall: h.row.HallCode, EmptySince: *since, Label: messages.EmptyHallTile(*since, h.codes)})
		}
	}
	return out
}

// formEntries are the entries holding the day's allocations, oldest first,
// with the caller's actions on them (§3.2, entries.Actions).
func formEntries(in FeedFormInput, head dbq.FormsFeedHeadRow, rows []dbq.FormsFeedRowsRow) []FormEntry {
	out := []FormEntry{}
	at := map[uuid.UUID]int{}
	created := map[uuid.UUID]time.Time{}
	for _, r := range rows {
		if r.ExistingEntryID == nil || r.ExistingCreatedAt == nil {
			continue
		}
		id := *r.ExistingEntryID
		if i, ok := at[id]; ok {
			out[i].Rows++
			continue
		}
		meta := EntryMeta{
			FarmIDs: []uuid.UUID{in.FarmID}, CreatedBy: r.ExistingCreatedBy, CreatedAt: *r.ExistingCreatedAt,
			SeriesClosed: head.SeriesClosed,
		}
		if r.ExistingPeerFarmID != nil {
			meta.FarmIDs = append(meta.FarmIDs, *r.ExistingPeerFarmID)
		}
		acts := Actions(in.Principal, meta, in.Today)
		at[id] = len(out)
		created[id] = *r.ExistingCreatedAt
		out = append(out, FormEntry{
			EntryID: id, Author: r.ExistingAuthor, At: timex.At(*r.ExistingCreatedAt), Rows: 1,
			Editable: acts.Editable, Correctable: acts.Correctable,
		})
	}
	slices.SortStableFunc(out, func(a, b FormEntry) int {
		if c := created[a.EntryID].Compare(created[b.EntryID]); c != 0 {
			return c
		}
		return cmp.Compare(a.EntryID.String(), b.EntryID.String())
	})
	return out
}

// formStock is the farm's stock rows as of in.Today (§5.8, M3 D7, Q5).
func formStock(ctx context.Context, db dbq.DBTX, in FeedFormInput) ([]FormStock, error) {
	settings, err := calc.ReadAppSettings(ctx, db)
	if err != nil {
		return nil, err
	}
	sd, err := calc.ReadStockDays(ctx, db, in.FarmID, in.Today, settings)
	if err != nil {
		return nil, err
	}
	out := make([]FormStock, 0, len(sd.Items))
	for _, it := range sd.Items {
		out = append(out, FormStock{
			FeedTypeID: it.FeedTypeID, Type: it.Name, ForSex: it.ForSexText(), Kg: num.Of(it.Kg), AvgDaily: num.Of(it.AvgDaily),
			Window: [2]timex.Day{sd.Window.From, sd.Window.To}, DaysLeft: decPtr(it.Left.Days), RunOut: it.Left.RunOut,
			Lamp: it.Left.Lamp, Reason: it.Left.Reason.Text(settings),
		})
	}
	return out, nil
}

// nullNum is n as a wire number, nil when it is null.
func nullNum(n decimal.NullDecimal) *num.Num {
	if !n.Valid {
		return nil
	}
	v := num.Of(n.Decimal)
	return &v
}

// decPtr is d as a wire number, nil for nil.
func decPtr(d *decimal.Decimal) *num.Num {
	if d == nil {
		return nil
	}
	v := num.Of(*d)
	return &v
}
