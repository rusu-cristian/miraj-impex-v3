package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeDailyForm is the Zilnic form state (§7.8, plan M5 T4, Q2).
const routeDailyForm = "GET /api/forms/daily"

// dailyFormRoutes mounts GET /api/forms/daily behind authed (Routes calls
// it, plan M5-D1).
func (h *Handler) dailyFormRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeDailyForm, authed.ThenFunc(h.dailyForm))
}

// dailyForm is GET /api/forms/daily?farmId&date (§7.8, Q2). Both parameters
// are required; any other parameter is a 400 VALIDATION issue (as the feed
// form, parseFormQuery).
func (h *Handler) dailyForm(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.svc.DailyForm(r.Context(), p, farmID, date)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// DailyForm is the Zilnic form state of farmID on date for p (§6.2, §7.8,
// Q2), read in one snapshot (InReadTx, M5-D10). Errors are envelope errors:
// 403 FORBIDDEN for a viewer, 403 FARM_FORBIDDEN for a farm outside the
// caller's farms, 400 VALIDATION on `date` for a day after today (the §3.2
// sentence), 404 NOT_FOUND for an owner's unknown farm. A vet's past day and
// a closed series are not errors: the form comes back with its lockedReason.
func (s *Service) DailyForm(ctx context.Context, p auth.Principal, farmID uuid.UUID, date timex.Day) (DailyFormResponse, error) {
	if err := checkWriter(p, farmID); err != nil {
		return DailyFormResponse{}, err
	}
	if !p.CanSeeFarm(farmID) { // checkWriter leaves the nil farm id to a create's level 1
		return DailyFormResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	today := timex.DayOf(now)
	if date.After(today) {
		return DailyFormResponse{}, httpx.Validation(httpx.Issue{
			Path: paramDate, Code: httpx.IssueOutOfRange, Message: messages.FutureDate(date, today),
		})
	}
	var res DailyFormResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildDailyForm(ctx, tx, DailyFormInput{Principal: p, FarmID: farmID, Date: date, Today: today, Now: now})
		return err
	})
	if errors.Is(err, ErrFormFarmNotFound) {
		return DailyFormResponse{}, httpx.NotFound()
	}
	return res, err
}

// DailyFormInput is what BuildDailyForm reads: the caller (for the lock and
// the entries' actions), the farm and day of the form, and the clock's day
// and instant (ARC-6: never app_today()).
type DailyFormInput struct {
	Principal auth.Principal
	FarmID    uuid.UUID
	Date      timex.Day
	Today     timex.Day
	Now       time.Time
}

// BuildDailyForm builds GET /api/forms/daily (§6.2, §7.8, Q2) on db for
// in.FarmID on in.Date. It reads, in this order: the head (farm, cycle on
// the date, guide lay % of SAPT(date), closed series, latest ended cycle),
// the section × sex rows with the day's saved bird values, the entries
// holding the day's values, and, when the form shows anything, the day's
// eggs (laying farm), calc.DeathsWindow and calc.HensStart (laying farm): at
// most 6 statements, to be run in one read snapshot (the service's
// InReadTx).
//
// It doesn't check the caller's role, farm or date (the service does); an
// unknown farm is ErrFormFarmNotFound.
func BuildDailyForm(ctx context.Context, db dbq.DBTX, in DailyFormInput) (DailyFormResponse, error) {
	q := dbq.New(db)
	head, err := q.DailyFormHead(ctx, dbq.DailyFormHeadParams{FarmID: in.FarmID, Day: in.Date})
	if errors.Is(err, pgx.ErrNoRows) {
		return DailyFormResponse{}, ErrFormFarmNotFound
	}
	if err != nil {
		return DailyFormResponse{}, fmt.Errorf("entries: daily form head: %w", err)
	}
	rows, err := q.DailyFormSections(ctx, dbq.DailyFormSectionsParams{FarmID: in.FarmID, Day: in.Date})
	if err != nil {
		return DailyFormResponse{}, fmt.Errorf("entries: daily form sections: %w", err)
	}
	entryRows, err := q.DailyFormEntries(ctx, dbq.DailyFormEntriesParams{FarmID: in.FarmID, Day: in.Date})
	if err != nil {
		return DailyFormResponse{}, fmt.Errorf("entries: daily form entries: %w", err)
	}

	res := DailyFormResponse{
		Farm: FormFarm{ID: head.ID, Name: head.Name, Type: head.Type},
		Date: in.Date, AsOf: timex.At(in.Now),
		Halls: []DailyFormHall{}, EmptyHalls: []FormEmptyHall{}, Existing: []DailyFormEntry{},
	}
	if head.CycleID != nil {
		age := calc.SeriesAge{PlacementDate: *head.PlacementDate, AgeAtPlacementDays: int(*head.AgeAtPlacementDays)}
		sapt, ageDays := calc.SAPT(age, in.Date), calc.AgeDays(age, in.Date)
		res.SAPT, res.AgeDays = &sapt, &ageDays
		res.Cycle = &FormCycle{ID: *head.CycleID, Series: *head.SeriesName, StartDate: *head.CycleStart}
	} else {
		res.Empty = dailyEmpty(head)
	}
	res.LockedReason = dailyLockedReason(in, head.ClosedSeries)
	res.EmptyHalls = dailyEmptyHalls(rows, res.Cycle)
	res.Existing = dailyEntries(in, head.ClosedSeries != "", entryRows)

	sections := shownSections(rows)
	laying := head.Type == dbq.FarmTypeLaying
	var eggs []dbq.DailyFormEggsRow
	if laying {
		if eggs, err = q.DailyFormEggs(ctx, dbq.DailyFormEggsParams{FarmID: in.FarmID, Day: in.Date}); err != nil {
			return DailyFormResponse{}, fmt.Errorf("entries: daily form eggs: %w", err)
		}
	}
	if len(sections) == 0 && len(eggs) == 0 {
		return res, nil // nothing to show: no deaths or hens to read
	}
	deaths, err := calc.DeathsWindow(ctx, db, in.FarmID, in.Date)
	if err != nil {
		return DailyFormResponse{}, err
	}
	var units []calc.EggUnitHens
	if laying {
		if units, err = calc.HensStart(ctx, db, in.FarmID, in.Date); err != nil {
			return DailyFormResponse{}, err
		}
	}
	res.Halls = dailyHalls(rows, sections, deaths, units, eggs, nullNum(head.GuideLayPct))
	return res, nil
}

// dailyLockedReason is why the caller can't save the day, checked as the
// save checks it: a vet on a day other than today (Z1), then a closed series
// (§3.1).
func dailyLockedReason(in DailyFormInput, closedSeries string) *string {
	var msg string
	switch {
	case in.Principal.Role == auth.Vet && in.Date != in.Today:
		msg = messages.PastDayVet()
	case closedSeries != "":
		msg = messages.SeriesClosed(closedSeries)
	default:
		return nil
	}
	return &msg
}

// dailyEmpty is the empty-farm state of a day no cycle contains (as the farm
// page, M2 Q4 b, c): "goală din" the end of the latest cycle that ended
// before the day, the no-cycle sentence when none did.
func dailyEmpty(head dbq.DailyFormHeadRow) *DailyFormEmpty {
	if head.LastEnd == nil || head.LastStart == nil || head.LastSeries == nil ||
		head.LastPlacementDate == nil || head.LastAgeAtPlacementDays == nil {
		return &DailyFormEmpty{Label: messages.FarmNoCycle(), Hint: messages.CycleStartsHint()}
	}
	since := *head.LastEnd
	age := calc.SeriesAge{PlacementDate: *head.LastPlacementDate, AgeAtPlacementDays: int(*head.LastAgeAtPlacementDays)}
	return &DailyFormEmpty{
		Since: &since,
		Label: messages.FarmEmpty(since, *head.LastSeries, calc.SAPT(age, *head.LastStart), calc.SAPT(age, since)),
		Hint:  messages.CycleStartsHint(),
	}
}

// formSection is one section's two sex rows of DailyFormSections, as shown.
type formSection struct {
	rows  [2]*dbq.DailyFormSectionsRow // [0] ♀, [1] ♂ (either may be nil)
	sexes []dbq.Sex
}

// shownSections are the sections the form shows, by section id: an active
// section with birds at the start of the day or an inflow that day (the
// sexes present, §5.3), or any section holding a saved value of the day
// (that sex is shown too).
func shownSections(rows []dbq.DailyFormSectionsRow) map[uuid.UUID]*formSection {
	all := map[uuid.UUID]*formSection{}
	for i := range rows {
		r := &rows[i]
		s, ok := all[r.SectionID]
		if !ok {
			s = &formSection{}
			all[r.SectionID] = s
		}
		s.rows[sexIndex(r.Sex)] = r
	}
	out := map[uuid.UUID]*formSection{}
	for id, s := range all {
		for _, r := range s.rows {
			if r == nil {
				continue
			}
			present := !r.Retired && (r.AliveStart > 0 || r.Inflow)
			if present || savedSex(r) != nil {
				s.sexes = append(s.sexes, r.Sex)
			}
		}
		if len(s.sexes) > 0 {
			out[id] = s
		}
	}
	return out
}

func sexIndex(s dbq.Sex) int {
	if s == dbq.SexM {
		return 1
	}
	return 0
}

// savedSex is a row's saved values in the create shape (kinds.go Daily: the
// counts positive, the sexing error as recorded), nil when it holds none.
func savedSex(r *dbq.DailyFormSectionsRow) *DailySex {
	if !r.HasMortality && !r.HasCull && !r.HasSexingError {
		return nil
	}
	v := &DailySex{}
	if r.HasMortality {
		v.Mortality = intNum(int64(r.Mortality))
	}
	if r.HasCull {
		v.Cull = intNum(int64(r.Cull))
		if r.CullReason != "" {
			reason := r.CullReason
			v.CullReason = &reason
		}
	}
	if r.HasSexingError {
		v.SexingError = intNum(int64(r.SexingError))
	}
	return v
}

// dailyHalls groups the shown sections by hall in tree order (the rows'
// order), each hall with its egg unit; a hall without a shown section is
// kept when its egg unit is shown (an active unit with hens at the start of
// the day, or saved eggs). units is nil at a growth farm (no eggs).
func dailyHalls(rows []dbq.DailyFormSectionsRow, shown map[uuid.UUID]*formSection, deaths []calc.SectionDeaths,
	units []calc.EggUnitHens, eggs []dbq.DailyFormEggsRow, guideLay *num.Num) []DailyFormHall {
	byDeaths := map[uuid.UUID]calc.SectionDeaths{}
	for _, d := range deaths {
		byDeaths[d.SectionID] = d
	}
	byEggs := map[uuid.UUID]dbq.DailyFormEggsRow{}
	for _, e := range eggs {
		byEggs[e.SectionID] = e
	}
	byUnit := map[uuid.UUID]calc.EggUnitHens{}
	for _, u := range units {
		byUnit[u.UnitID] = u
	}

	out := []DailyFormHall{}
	at := map[uuid.UUID]int{} // hall id → index in out
	done := map[uuid.UUID]bool{}
	for _, r := range rows {
		if _, ok := at[r.HallID]; !ok {
			hall := DailyFormHall{Sections: []DailyFormSection{}}
			if r.ParentID != nil {
				hall.Hall = &DailyFormHallRef{ID: r.HallID, Code: r.HallCode}
			}
			if u, ok := byUnit[r.HallID]; ok {
				hall.Eggs = formEggs(u, r.HallRetired, byEggs, guideLay)
			}
			out = append(out, hall)
			at[r.HallID] = len(out) - 1
		}
		s, ok := shown[r.SectionID]
		if !ok || done[r.SectionID] {
			continue
		}
		done[r.SectionID] = true
		i := at[r.HallID]
		out[i].Sections = append(out[i].Sections, formSectionOf(r, s, byDeaths[r.SectionID]))
	}
	kept := out[:0]
	for _, h := range out {
		if len(h.Sections) > 0 || h.Eggs != nil {
			kept = append(kept, h)
		}
	}
	return kept
}

// formSectionOf is a shown section with its counts, deaths and saved values.
func formSectionOf(r dbq.DailyFormSectionsRow, s *formSection, d calc.SectionDeaths) DailyFormSection {
	out := DailyFormSection{
		SectionID: r.SectionID, Code: r.Code, Sexes: s.sexes,
		Yesterday: httpx.SexPair[int]{F: d.Yesterday.F, M: d.Yesterday.M},
		Mean7:     httpx.SexPair[num.Num]{F: num.Of(calc.Mean7(d.Week.F)), M: num.Of(calc.Mean7(d.Week.M))},
	}
	var saved httpx.SexPair[*DailySex]
	if f := s.rows[0]; f != nil {
		out.AliveStart.F = int(f.AliveStart)
		saved.F = savedSex(f)
	}
	if m := s.rows[1]; m != nil {
		out.AliveStart.M = int(m.AliveStart)
		saved.M = savedSex(m)
	}
	if saved.F != nil || saved.M != nil {
		out.Saved = &saved
	}
	return out
}

// formEggs is an egg unit's eggs, nil when it holds no saved eggs and has
// no hens at the start of the day or is retired (the save refuses eggs
// there: Z3's egg form, plan M5 Q3).
func formEggs(u calc.EggUnitHens, retired bool, saved map[uuid.UUID]dbq.DailyFormEggsRow, guideLay *num.Num) *DailyFormEggs {
	e, ok := saved[u.UnitID]
	if (u.HensStart <= 0 || retired) && !ok {
		return nil
	}
	out := &DailyFormEggs{SectionID: u.UnitID, Code: u.Code, HensStart: u.HensStart, GuideLayPct: guideLay}
	if ok {
		out.Saved = &DailyEggsValues{
			Floor: int32Num(e.FloorEggs), Cat1: int32Num(e.Cat1), Cat2: int32Num(e.Cat2), Cat3: int32Num(e.Cat3),
			Hatching: int32Num(e.HatchingEggs), AvgWeightG: nullNum(e.AvgEggWeightG),
		}
	}
	return out
}

// dailyEmptyHalls are the halls with no bird at the end of the day that were
// emptied on or after the start of the date's cycle (as the summary's tiles
// and the feed form, golden 38), in tree order, each with its bird-holding
// sections.
func dailyEmptyHalls(rows []dbq.DailyFormSectionsRow, cycle *FormCycle) []FormEmptyHall {
	out := []FormEmptyHall{}
	if cycle == nil {
		return out
	}
	type hall struct {
		row   dbq.DailyFormSectionsRow // its first row: the hall's code, count and empty_since
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

// dailyEntries are the entries holding the day's values, in the query's
// order (oldest first), with their sections and egg units and the caller's
// actions on them (§3.2, entries.Actions). An entry of a kind the change
// routes don't serve (Q1: the import's migration entry) has no action.
func dailyEntries(in DailyFormInput, seriesClosed bool, rows []dbq.DailyFormEntriesRow) []DailyFormEntry {
	out := []DailyFormEntry{}
	at := map[uuid.UUID]int{}
	for _, r := range rows {
		i, ok := at[r.EntryID]
		if !ok {
			e := DailyFormEntry{
				EntryID: r.EntryID, Author: r.Author, At: timex.At(r.CreatedAt), Sections: []string{}, Eggs: []string{},
			}
			if Served(r.Kind) {
				meta := EntryMeta{
					FarmIDs: entryFarms(in.FarmID, r.PeerFarmID), CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt,
					SeriesClosed: seriesClosed,
				}
				acts := Actions(in.Principal, meta, in.Today)
				e.Editable, e.Correctable = acts.Editable, acts.Correctable
			}
			out = append(out, e)
			i = len(out) - 1
			at[r.EntryID] = i
		}
		if r.Egg {
			out[i].Eggs = append(out[i].Eggs, r.Code)
		} else {
			out[i].Sections = append(out[i].Sections, r.Code)
		}
	}
	return out
}
