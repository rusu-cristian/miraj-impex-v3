package entries

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeDaily is the Zilnic create; the runner never shares a clientRequestId
// between routes.
const routeDaily = "POST /api/entries/daily"

// The unique keys of a live Zilnic value (§4.7): a bird value per (section,
// day, sex, type), the eggs per (egg unit, day).
const (
	constraintBirdDaily = "bird_event_daily_uq"
	constraintEggs      = "egg_production_uq"
)

// The S17 chips of a Zilnic save (§4.10, plan Q5).
const (
	chipMortality = "mortalitate"
	chipCull      = "sacrificare"
	chipSexing    = "eroare_sexare"
	chipEggs      = "oua"
)

// dailySaveRoutes registers the Zilnic save (plan M5-D1: handler.go's Routes
// mounts it, T9).
func (h *Handler) dailySaveRoutes(mux httpx.Router, authed httpx.Chain) {
	mux.Handle(routeDaily, authed.ThenFunc(h.createDaily))
}

// createDaily is POST /api/entries/daily (§7.8, §12.5, plan M5-D13): decode
// (64 KiB) → role (viewer 403 FORBIDDEN) → farm (FARM_FORBIDDEN; an owner's
// unknown farm 404) → the farm's section codes → Validate (level 1) →
// service. The answer is sent byte for byte as the runner encoded it.
func (h *Handler) createDaily(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	req, err := httpx.Decode[DailyCreateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := checkWriter(p, req.FarmID); err != nil { // before any read: no section code of another farm leaks
		httpx.WriteError(w, err)
		return
	}
	refs, err := h.svc.DailyRefs(r.Context(), req.FarmID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var v httpx.Validator
	req.Validate(&v, refs)
	if err := v.Err(); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.CreateDaily(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// DailyRefs reads what DailyValues.ValidateAt needs: the codes of the
// farm's sections. An unknown farm is 404 NOT_FOUND (only an owner gets this
// far with one; anyone else got FARM_FORBIDDEN from checkWriter). A nil
// farmID reads nothing.
func (s *Service) DailyRefs(ctx context.Context, farmID uuid.UUID) (DailyRefs, error) {
	refs := DailyRefs{Sections: map[uuid.UUID]string{}}
	if farmID == uuid.Nil {
		return refs, nil
	}
	sections, err := s.store.Q().EntriesFarmSections(ctx, farmID)
	if err != nil {
		return refs, err
	}
	if len(sections) == 0 {
		return refs, httpx.NotFound()
	}
	for _, sec := range sections {
		if sec.SectionID != nil && sec.Code != nil {
			refs.Sections[*sec.SectionID] = *sec.Code
		}
	}
	return refs, nil
}

// CreateDaily saves one Zilnic form (§6.2, §7.8, plan M5-D13) on the save
// runner. req must have passed Validate (level 1). Inside the transaction,
// after the farm lock and in this order:
//   - a future date → 422 REFUSED (§3.2); a vet on another day than today →
//     403 PAST_DAY_VET (Z1); a day inside a cycle of a closed series → 409
//     SERIES_CLOSED (§3.1);
//   - the level-2 plan (planDaily: Z3, Z9, Z4, Z5, Z7 in document order) and
//     the Z6 warnings; a key held by another entry → 409 ALREADY_ENTERED
//     with the holder; any other refusal → 422 REFUSED with all of them;
//   - the entry and its rows; §6.1 over every day from the date to today
//     (Tx.CheckCounts, Q4) → 422 REFUSED "Salvarea ar lăsa …" with the field
//     bounds;
//   - the farm's birds at the end of today (a running sum, calc.ReadFarmDay).
//
// A dry run answers 200 {effects: [], warnings, refusals} with the refusals
// in the body, or the warnings of the real save after its writes, rolled
// back. A save writes one audit row and the live events entry.created and,
// when it has bird rows, counts.changed.
func (s *Service) CreateDaily(ctx context.Context, p auth.Principal, req DailyCreateRequest) (Result, error) {
	if err := checkWriter(p, req.FarmID); err != nil {
		return Result{}, err
	}
	return s.runner.Create(ctx, p, Create{
		Route: routeDaily, Kind: dbq.EntryKindDaily, FarmIDs: []uuid.UUID{req.FarmID}, Meta: req.WriteMeta,
		OnConflict: func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError) error {
			if ce.Constraint != constraintBirdDaily && ce.Constraint != constraintEggs {
				return nil
			}
			return dailyHolderError(ctx, q, req.DailyValues)
		},
	}, func(ctx context.Context, tx *Tx) (Saved, error) {
		return planDailySave(ctx, tx, req.DailyValues)
	})
}

// planDailySave is the form's part of the save (§12.5 last note).
func planDailySave(ctx context.Context, tx *Tx, v DailyValues) (Saved, error) {
	refuse := func(refusals []httpx.Refusal, message string, warnings []httpx.Warning) (Saved, error) {
		if tx.DryRun {
			return Saved{}, tx.Preview(NewDryRun(nil, warnings, refusals))
		}
		if warnings == nil {
			warnings = []httpx.Warning{}
		}
		return Saved{}, httpx.Refused(message, refusals, warnings)
	}
	if !v.hasValue() { // Validate guarantees a value
		return Saved{}, httpx.Validation(httpx.Issue{Path: "sections", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	if v.Date.After(tx.Today) { // §3.2, every role
		msg := messages.FutureDate(v.Date, tx.Today)
		return refuse([]httpx.Refusal{{Field: "date", Message: msg}}, msg, nil)
	}
	if tx.Principal.Role == auth.Vet && v.Date != tx.Today { // Z1
		return Saved{}, httpx.PastDayVet()
	}
	st, ok, err := loadDailyState(ctx, tx.Q, v.FarmID, v.Date)
	if err != nil {
		return Saved{}, err
	}
	if !ok {
		return Saved{}, httpx.NotFound()
	}
	if st.ClosedSeries != "" {
		return Saved{}, httpx.SeriesClosed(st.ClosedSeries)
	}

	plan := planDaily(st, v)
	warnings := []httpx.Warning{}
	if v.hasDeaths() { // Z6 on the 7 days before the date: the save's own rows never count
		deaths, err := calc.DeathsWindow(ctx, tx.Tx, v.FarmID, v.Date)
		if err != nil {
			return Saved{}, err
		}
		for _, sp := range spikesOf(deaths, st.MortalityFactor, v) {
			if !plan.Z3[ledger.SectionSex{SectionID: sp.SectionID, Sex: sp.Sex}] {
				warnings = append(warnings, sp.Warning())
			}
		}
	}
	if plan.Collision != nil && !tx.DryRun {
		return Saved{}, httpx.AlreadyEntered(plan.Collision.EntryID, plan.Collision.Message)
	}
	if len(plan.Refusals) > 0 {
		refusals := make([]httpx.Refusal, len(plan.Refusals))
		for i, r := range plan.Refusals {
			refusals[i] = r.Refusal
		}
		return refuse(refusals, plan.Refusals[0].Banner, warnings)
	}

	ins, err := tx.InsertEntry(ctx, Header{FarmID: v.FarmID, Date: v.Date})
	if err != nil {
		return Saved{}, err
	}
	if len(plan.Birds) > 0 {
		rows, err := jsonRows(plan.Birds)
		if err != nil {
			return Saved{}, err
		}
		if err := tx.Q.DailyInsertBirdEvents(ctx, dbq.DailyInsertBirdEventsParams{
			EntryID: ins.ID, FarmID: v.FarmID, Date: v.Date, Rows: rows,
		}); err != nil {
			return Saved{}, err
		}
	}
	if len(plan.Eggs) > 0 {
		rows, err := jsonRows(plan.Eggs)
		if err != nil {
			return Saved{}, err
		}
		if err := tx.Q.DailyInsertEggProduction(ctx, dbq.DailyInsertEggProductionParams{
			EntryID: ins.ID, FarmID: v.FarmID, Date: v.Date, Rows: rows,
		}); err != nil {
			return Saved{}, err
		}
	}
	bad, err := tx.CheckCounts(ctx, plan.Affected, v.Date) // §6.1 steps 1–3 after the rows; none for eggs only
	if err != nil {
		return Saved{}, err
	}
	if len(bad) > 0 {
		refusals, banner := countRefusals(plan, bad, tx.Today, messages.ActionSave)
		return refuse(refusals, banner, warnings)
	}
	if tx.DryRun {
		return Saved{}, tx.Preview(NewDryRun(nil, warnings, nil))
	}

	day, err := calc.ReadFarmDay(ctx, tx.Tx, v.FarmID, tx.Today) // the farm's birds at the end of today, after the save
	if err != nil {
		return Saved{}, err
	}
	alive := httpx.SexPair[int]{F: day.AliveEnd.F, M: day.AliveEnd.M}
	summary := messages.DailySummary(len(plan.BirdCodes), len(plan.EggCodes))
	toast := messages.DailySaved(len(plan.BirdCodes), plan.SumF, plan.SumM)
	if len(plan.Birds) == 0 {
		toast = messages.DailyEggsSaved(len(plan.EggCodes), plan.EggsTotal)
	}
	types := plan.recordTypes()
	at := timex.At(tx.Now)
	events := []live.Event{live.EntryCreated(live.EntryPayload{
		EntryID: ins.ID, Kind: ins.Kind, FarmIDs: []uuid.UUID{v.FarmID}, Dates: []timex.Day{v.Date},
		Sections: append(append([]string{}, plan.BirdCodes...), plan.EggCodes...), RecordTypes: types,
		Author: &tx.Principal.Name, At: at, Summary: summary,
	})}
	if len(plan.Birds) > 0 {
		events = append(events, live.CountsChanged(v.FarmID, v.Date, plan.BirdCodes, alive))
	}
	return Saved{
		Response: EntryWriteResponse{Entry: ins.Ref(), Warnings: warnings, Toast: toast, Alive: &alive, AsOf: at},
		Audit: audit.Row{
			Action: dbq.AuditActionCreate, Entity: "entry", EntityID: ins.ID, EntryID: &ins.ID, EntryKind: &ins.Kind,
			RecordTypes: types, FarmIDs: []uuid.UUID{v.FarmID}, RecordDate: &v.Date, Summary: &summary,
			NewValues: plan.Audit,
		},
		Events: events,
	}, nil
}

// recordTypes are the S17 chips of the save (plan Q5): mortalitate for any
// mortality value (0 included), sacrificare, eroare_sexare, oua.
func (p dailyPlan) recordTypes() []string {
	has := map[dbq.BirdEventType]bool{}
	for _, b := range p.Birds {
		has[b.Type] = true
	}
	out := []string{}
	for _, c := range []struct {
		ok   bool
		chip string
	}{
		{has[dbq.BirdEventTypeMortality], chipMortality},
		{has[dbq.BirdEventTypeCull], chipCull},
		{has[dbq.BirdEventTypeSexingError], chipSexing},
		{len(p.Eggs) > 0, chipEggs},
	} {
		if c.ok {
			out = append(out, c.chip)
		}
	}
	return out
}

// dailyHolderError answers a bird_event_daily_uq / egg_production_uq
// violation after the transaction (M3 D2): 409 ALREADY_ENTERED naming the
// first value of v whose key is now held, with its entry. nil when none is
// held any more (the violation stays a 500).
func dailyHolderError(ctx context.Context, q *dbq.Queries, v DailyValues) error {
	st, ok, err := loadDailyState(ctx, q, v.FarmID, v.Date)
	if err != nil || !ok {
		return err
	}
	if c := planDaily(st, v).Collision; c != nil {
		return httpx.AlreadyEntered(c.EntryID, c.Message)
	}
	return nil
}
