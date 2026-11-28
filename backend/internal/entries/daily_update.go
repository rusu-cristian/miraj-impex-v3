package entries

import (
	"context"
	"fmt"
	"net/http"
	"slices"

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
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Zilnic adapter of PUT /api/entries/{id} (§7.8, §6.15; plan M5-D5,
// D6, Q7, Q12): the edit or Corectează of a whole Zilnic save.

// updateDaily decodes and checks a Zilnic PUT (level 1: the body as
// DailyUpdateRequest with every member known, then validateDailyUpdate on
// the stored farm's section codes) and runs it.
func (h *Handler) updateDaily(w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	req, err := httpx.Decode[DailyUpdateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		return Result{}, err
	}
	refs, err := h.svc.DailyRefs(r.Context(), t.FarmID)
	if err != nil {
		return Result{}, err
	}
	var v httpx.Validator
	validateDailyUpdate(&v, req, t, refs)
	if err := v.Err(); err != nil {
		return Result{}, err
	}
	return h.svc.UpdateDaily(r.Context(), p, t.ID, req)
}

// UpdateDaily edits or corrects Zilnic entry id for p (§7.8, §6.15, plan
// M5-D5, D6) on the change runner; req must have passed
// validateDailyUpdate. The server decides: the author on the creation day
// edits (no Motiv, a sent one is ignored; audit `update` same_day_author;
// `corrected` unchanged); a manager on own farms or the owner corrects (Motiv
// required, 400 on `reason` naming the change, not asked of a dry run; audit
// `update` with the Motiv; `corrected` set); anyone else 403. Under the
// locks, in this order (planDailyUpdate):
//   - values name the stored farm and date (400), ifVersion is the entry's
//     version (409 STALE);
//   - the diff against the stored save (DiffDaily): none → 422 "Nu ai
//     schimbat nicio cifră.";
//   - Z3, Z4, Z5, Z7 on the changed and added values only, Z9 on added keys
//     (409 ALREADY_ENTERED), Z6 warnings on every value of the new save; any
//     but Z5 → 422 before a write;
//   - the rows: a changed value in place, a cleared one soft-deleted, a new
//     one inserted; the version moved;
//   - §6.1 from the entry date over the section × sex counts that go down
//     ("Corecția ar lăsa …" / "Salvarea ar lăsa …" → 422 with the field
//     bound); a Z5 refusal of a pair §6.1 also refuses gives way to §6.1's
//     stricter bound (the Corectează form: 4 292 → "cel mult 4 284", not
//     the day's 4 286), the others are answered with it.
//
// A dry run answers 200 CorrectionPreview (Q7) and rolls back.
func (s *Service) UpdateDaily(ctx context.Context, p auth.Principal, id uuid.UUID, req DailyUpdateRequest) (Result, error) {
	return s.runner.Change(ctx, p, ChangeSpec{
		EntryID: id, Action: dbq.AuditActionUpdate, DryRun: req.DryRun, Meta: req.meta(), Allow: allowUpdate(p),
		OnConflict: func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError, c ChangeConflict) error {
			if ce.Constraint != constraintBirdDaily && ce.Constraint != constraintEggs {
				return nil
			}
			return dailyHolderErrorFor(ctx, q, req.Values, c.Entry.ID)
		},
	}, func(ctx context.Context, tx *ChangeTx) (Saved, error) {
		return planDailyUpdate(ctx, tx, req)
	})
}

// planDailyUpdate is the Zilnic adapter's plan of a PUT (UpdateDaily).
func planDailyUpdate(ctx context.Context, tx *ChangeTx, req DailyUpdateRequest) (Saved, error) {
	e := tx.Entry
	v := req.Values
	if e.Kind != dbq.EntryKindDaily {
		return Saved{}, httpx.NotFound()
	}
	if err := checkUpdateHead(tx, v.FarmID, v.Date, req.IfVersion); err != nil {
		return Saved{}, err
	}
	if !v.hasValue() { // validateDailyUpdate guarantees a value
		return Saved{}, httpx.Validation(httpx.Issue{Path: "values.sections", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	correction := !tx.Actions.Editable
	action := messages.ActionSave
	if correction {
		action = messages.ActionCorrection
	}
	stored, err := tx.Stored(ctx)
	if err != nil {
		return Saved{}, err
	}
	st, ok, err := loadDailyState(ctx, tx.Q, e.FarmID, e.Date)
	if err != nil {
		return Saved{}, err
	}
	if !ok {
		return Saved{}, httpx.NotFound()
	}
	st.dropHolds(e.ID)
	codes := st.codes()
	nv := normalizeDaily(st, v)
	diffs := diffDaily(stored.Daily(), nv, codes)
	changes := changesOf(diffs)

	preview := func(p CorrectionPreview) (Saved, error) {
		p.Changes = changes
		return Saved{}, tx.Preview(newCorrectionPreview(p))
	}
	if len(diffs) == 0 { // §6.15
		msg := messages.NothingChanged()
		refusals := []httpx.Refusal{{Field: "values", Message: msg}}
		if tx.DryRun {
			return preview(CorrectionPreview{Refusals: refusals})
		}
		return Saved{}, httpx.Refused(msg, refusals, []httpx.Warning{})
	}
	reason := req.reason()
	if correction && reason == "" && !tx.DryRun { // §6.15: the sentence names the change
		return Saved{}, httpx.Validation(httpx.Issue{Path: "reason", Code: httpx.IssueRequired, Message: reasonMissing(diffs)})
	}

	pairs, units := changedUnits(diffs)
	plan := planDaily(st, changedOnly(nv, pairs, units)) // D6: level 2 on changed and added values only
	warnings := []httpx.Warning{}
	if nv.hasDeaths() { // Z6 on every value of the new save (D6, D173)
		deaths, err := calc.DeathsWindow(ctx, tx.Tx, e.FarmID, e.Date)
		if err != nil {
			return Saved{}, err
		}
		for _, sp := range spikesOf(deaths, st.MortalityFactor, nv) {
			if !plan.Z3[ledger.SectionSex{SectionID: sp.SectionID, Sex: sp.Sex}] {
				warnings = append(warnings, sp.Warning())
			}
		}
	}
	if plan.Collision != nil && !tx.DryRun {
		return Saved{}, httpx.AlreadyEntered(plan.Collision.EntryID, plan.Collision.Message)
	}
	// Z5 alone doesn't stop a change: §6.1 below gives the stricter bound over
	// every day from the entry date (the Corectează form, §6.1 table), and a
	// Z5 refusal stays only for a pair §6.1 lets through.
	var dayBounds []dailyRefusal
	if len(plan.Refusals) > 0 {
		if !slices.ContainsFunc(plan.Refusals, func(r dailyRefusal) bool { return r.Max == nil }) {
			dayBounds = plan.Refusals
		} else {
			refusals := make([]httpx.Refusal, len(plan.Refusals))
			for i, r := range plan.Refusals {
				refusals[i] = valuesPath(r.Refusal)
			}
			if tx.DryRun {
				return preview(CorrectionPreview{Refusals: refusals, Warnings: warnings})
			}
			return Saved{}, httpx.Refused(plan.Refusals[0].Banner, refusals, warnings)
		}
	}

	newRows := dailyRows(nv, codes)
	rows := planRows(stored.Rows, newRows)
	f := rowFigures{birds: moveOf(stored.Rows.BirdEvents, newRows.BirdEvents)}
	if err := readAlive(ctx, tx, &f, rows.birds()); err != nil {
		return Saved{}, err
	}
	if err := rows.apply(ctx, tx); err != nil {
		return Saved{}, err
	}
	ref, err := tx.UpdateHeader(ctx, e.VersionAt, e.Corrected || correction)
	if err != nil {
		return Saved{}, err
	}
	var refusals []httpx.Refusal
	banner := ""
	if down := f.birds.down(); len(down) > 0 { // §6.1 from the entry date; a count that only rises can't go below 0
		bad, err := tx.CheckCounts(ctx, down, e.Date)
		if err != nil {
			return Saved{}, err
		}
		if len(bad) > 0 {
			refusals, banner = countRefusals(plan, bad, tx.Today, action)
			for i := range refusals {
				refusals[i] = valuesPath(refusals[i])
			}
		}
	}
	for _, r := range dayBounds {
		if slices.ContainsFunc(refusals, func(x httpx.Refusal) bool { return x.SectionCode == r.SectionCode && x.Sex == r.Sex }) {
			continue
		}
		refusals = append(refusals, valuesPath(r.Refusal))
		if banner == "" {
			banner = r.Banner
		}
	}
	if tx.DryRun {
		return preview(CorrectionPreview{
			Effects: moveEffects(f, e.Date, st.FarmName, signedSums(stored.Rows.BirdEvents), signedSums(newRows.BirdEvents)),
			Alive:   f.previewAlive(e.Date, tx.Today), CheckedDays: f.checkedDays(e.Date, tx.Today),
			Refusals: refusals, Warnings: warnings,
		})
	}
	if len(refusals) > 0 {
		return Saved{}, httpx.Refused(banner, refusals, warnings)
	}

	birdCodes := rowSections(newRows.BirdEvents)
	summary := messages.DailySummary(len(birdCodes), len(newRows.EggProduction))
	var toast string
	switch {
	case !correction && len(newRows.BirdEvents) > 0: // an author edit: the form's create Toast (Q15)
		sums := signedSums(newRows.BirdEvents)
		toast = messages.DailySaved(len(birdCodes), sums.F, sums.M)
	case !correction:
		toast = messages.DailyEggsSaved(len(newRows.EggProduction), eggsTotal(newRows.EggProduction))
	default:
		toast = correctionToast(diffs, f)
	}
	types := diffChips(diffs)
	corrected := e.Corrected || correction
	at := timex.At(tx.Now)
	var motiv *string
	if correction {
		motiv = &reason
	}
	events := []live.Event{live.EntryUpdated(live.EntryPayload{
		EntryID: e.ID, Kind: e.Kind, FarmIDs: e.FarmIDs(), Dates: []timex.Day{e.Date}, Sections: diffSections(diffs),
		RecordTypes: types, Author: &tx.Principal.Name, At: at, Summary: summary, Corrected: corrected,
	})}
	return Saved{
		Response: EntryWriteResponse{Entry: ref, Warnings: warnings, Toast: toast, Alive: f.farmAfter(), AsOf: at},
		Audit: audit.Row{
			Reason: motiv, SameDayAuthor: !correction, RecordTypes: types, Summary: &summary, Changes: changes,
			OldValues: updateSnapshot{Snapshot: stored.Rows, Corrected: e.Corrected}, NewValues: dailyAuditOf(nv, codes),
		},
		Events: append(events, f.countsEvent(e)...),
	}, nil
}

// valuesPath puts a refusal's field under the PUT's "values." (planDaily and
// countRefusals name the create's paths).
func valuesPath(r httpx.Refusal) httpx.Refusal {
	if r.Field != "" {
		r.Field = "values." + r.Field
	}
	return r
}

// dropHolds forgets the keys entry self holds: a change of self never
// collides with its own values (Z9 is about another entry's).
func (st dailyState) dropHolds(self uuid.UUID) {
	for _, sec := range st.Sections {
		if sec.EggsHeld != nil && sec.EggsHeld.EntryID == self {
			sec.EggsHeld = nil
		}
		for _, slot := range sec.Sex {
			for t, h := range slot.Held {
				if h != nil && h.EntryID == self {
					delete(slot.Held, t)
				}
			}
		}
	}
}

// codes is every section of the farm as id → code.
func (st dailyState) codes() map[uuid.UUID]string {
	out := make(map[uuid.UUID]string, len(st.Sections))
	for id, sec := range st.Sections {
		out[id] = sec.Code
	}
	return out
}

// dailyHolderErrorFor is dailyHolderError for a change of entry self: the
// first value of v whose key another entry holds now (self's own keys are
// not collisions). nil when none is held any more.
func dailyHolderErrorFor(ctx context.Context, q *dbq.Queries, v DailyValues, self uuid.UUID) error {
	st, ok, err := loadDailyState(ctx, q, v.FarmID, v.Date)
	if err != nil || !ok {
		return err
	}
	st.dropHolds(self)
	if c := planDaily(st, v).Collision; c != nil {
		return httpx.AlreadyEntered(c.EntryID, c.Message)
	}
	return nil
}

// normalizeDaily is v with "Fără pierderi azi" written out as planDaily
// writes it: mortality 0 and cull 0 (no reason) for every sex present of a
// bird-holding, active section; a typed sexing error stays. A noLosses that
// can't be written out (a hall, a retired section, no sex present) stays
// set, so planDaily refuses it (Z3). v is not changed.
func normalizeDaily(st dailyState, v DailyValues) DailyValues {
	out := v
	out.Sections = make([]DailySection, len(v.Sections))
	for i, s := range v.Sections {
		c := s
		if s.NoLosses {
			sec, ok := st.Sections[s.SectionID]
			expanded := false
			for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
				if !ok || sec.IsHall || sec.Retired || !sec.Sex[sex].present() {
					continue
				}
				x := DailySex{}
				if cur := c.sex(sex); cur != nil {
					x = *cur
				}
				mortality, cull := num.FromInt(0), num.FromInt(0)
				x.Mortality, x.Cull, x.CullReason = &mortality, &cull, nil
				c.setSex(sex, &x)
				expanded = true
			}
			c.NoLosses = !expanded
		}
		out.Sections[i] = c
	}
	return out
}

func (s DailySection) sex(sex dbq.Sex) *DailySex {
	if sex == dbq.SexM {
		return s.M
	}
	return s.F
}

func (s *DailySection) setSex(sex dbq.Sex, x *DailySex) {
	if sex == dbq.SexM {
		s.M = x
	} else {
		s.F = x
	}
}

// changedOnly is v keeping only the section × sex values and egg units that
// changed (pairs, units), at their indices, so planDaily's paths stay the
// request's: an unchanged sex becomes empty, an unchanged egg unit an empty
// row. A noLosses left set (normalizeDaily) is kept.
func changedOnly(v DailyValues, pairs map[ledger.SectionSex]bool, units map[uuid.UUID]bool) DailyValues {
	out := v
	out.Sections = make([]DailySection, len(v.Sections))
	for i, s := range v.Sections {
		c := s
		for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
			if !pairs[ledger.SectionSex{SectionID: s.SectionID, Sex: sex}] {
				c.setSex(sex, nil)
			}
		}
		out.Sections[i] = c
	}
	out.Eggs = make([]DailyEggs, len(v.Eggs))
	for i, g := range v.Eggs {
		if units[g.SectionID] {
			out.Eggs[i] = g
		} else {
			out.Eggs[i] = DailyEggs{SectionID: g.SectionID}
		}
	}
	return out
}

// dailyDiff is one changed value of a Zilnic save.
type dailyDiff struct {
	SectionID uuid.UUID
	Eggs      bool // an egg value; else a bird value of Change.Sex
	Change    audit.Change
}

// The fields of a Zilnic save in document order (§4.10 changes, Q12).
var (
	dailyBirdFields = []messages.DailyField{
		messages.FieldMortality, messages.FieldCull, messages.FieldCullReason, messages.FieldSexingError,
	}
	dailyEggFields = []messages.EggField{
		messages.FieldFloor, messages.FieldCat1, messages.FieldCat2, messages.FieldCat3, messages.FieldHatching,
		messages.FieldAvgWeightG,
	}
)

// DiffDaily is the diff of two Zilnic saves of one farm and day, old → new
// (§4.10 `changes`, §7.8, plan M5-D9, Q12; golden 37): one audit.Change per
// value that differs, keyed "H5.2 · găini moarte" (messages.AuditKeyDaily /
// AuditKeyEggs) with the section's code from codes. In document order:
// sections in new's order then those only in old, ♀ before ♂, mortality,
// cull, cull reason, sexing error; then the egg units the same way, Jos,
// Cat. 1–3, Incubabile, masa medie. Counts are the typed counts
// (mortality 3), a cull reason its shown word ("boală"), an empty value
// null. noLosses is not read: write it out first (the update does). Equal
// values in another notation (2.50, 2.5) are no change.
func DiffDaily(old, new DailyValues, codes map[uuid.UUID]string) []audit.Change {
	return changesOf(diffDaily(old, new, codes))
}

func changesOf(diffs []dailyDiff) []audit.Change {
	out := make([]audit.Change, len(diffs))
	for i, d := range diffs {
		out[i] = d.Change
	}
	return out
}

func diffDaily(old, new DailyValues, codes map[uuid.UUID]string) []dailyDiff {
	var out []dailyDiff
	var sectionIDs []uuid.UUID
	for _, list := range [][]DailySection{new.Sections, old.Sections} {
		for _, s := range list {
			if !slices.Contains(sectionIDs, s.SectionID) {
				sectionIDs = append(sectionIDs, s.SectionID)
			}
		}
	}
	for _, id := range sectionIDs {
		o, n := sectionOf(old, id), sectionOf(new, id)
		code := codes[id]
		for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
			for _, field := range dailyBirdFields {
				ov, nv := birdValue(o.sex(sex), field), birdValue(n.sex(sex), field)
				if !ov.Equal(nv) {
					out = append(out, dailyDiff{SectionID: id, Change: audit.Change{
						Key: messages.AuditKeyDaily(code, sex, field), Section: code, Sex: sex, Field: string(field), Old: ov, New: nv,
					}})
				}
			}
		}
	}
	var unitIDs []uuid.UUID
	for _, list := range [][]DailyEggs{new.Eggs, old.Eggs} {
		for _, g := range list {
			if !slices.Contains(unitIDs, g.SectionID) {
				unitIDs = append(unitIDs, g.SectionID)
			}
		}
	}
	for _, id := range unitIDs {
		o, n := eggsOf(old, id), eggsOf(new, id)
		code := codes[id]
		for _, field := range dailyEggFields {
			ov, nv := audit.OptNumber(o.field(field)), audit.OptNumber(n.field(field))
			if !ov.Equal(nv) {
				out = append(out, dailyDiff{SectionID: id, Eggs: true, Change: audit.Change{
					Key: messages.AuditKeyEggs(code, field), Section: code, Field: string(field), Old: ov, New: nv,
				}})
			}
		}
	}
	return out
}

func sectionOf(v DailyValues, id uuid.UUID) DailySection {
	for _, s := range v.Sections {
		if s.SectionID == id {
			return s
		}
	}
	return DailySection{SectionID: id}
}

func eggsOf(v DailyValues, id uuid.UUID) DailyEggs {
	for _, g := range v.Eggs {
		if g.SectionID == id {
			return g
		}
	}
	return DailyEggs{SectionID: id}
}

// birdValue is one field of a sex as a Change value.
func birdValue(s *DailySex, field messages.DailyField) audit.Value {
	if s == nil {
		return audit.Null()
	}
	switch field {
	case messages.FieldMortality:
		return audit.OptNumber(s.Mortality)
	case messages.FieldCull:
		return audit.OptNumber(s.Cull)
	case messages.FieldCullReason:
		if s.CullReason == nil {
			return audit.Null()
		}
		return audit.Text(messages.CullReasonWord(*s.CullReason))
	default:
		return audit.OptNumber(s.SexingError)
	}
}

func (e DailyEggs) field(f messages.EggField) *num.Num {
	switch f {
	case messages.FieldFloor:
		return e.Floor
	case messages.FieldCat1:
		return e.Cat1
	case messages.FieldCat2:
		return e.Cat2
	case messages.FieldCat3:
		return e.Cat3
	case messages.FieldHatching:
		return e.Hatching
	default:
		return e.AvgWeightG
	}
}

// changedUnits are the section × sex pairs with a changed bird value and the
// egg units with a changed egg value.
func changedUnits(diffs []dailyDiff) (map[ledger.SectionSex]bool, map[uuid.UUID]bool) {
	pairs, units := map[ledger.SectionSex]bool{}, map[uuid.UUID]bool{}
	for _, d := range diffs {
		if d.Eggs {
			units[d.SectionID] = true
		} else {
			pairs[ledger.SectionSex{SectionID: d.SectionID, Sex: d.Change.Sex}] = true
		}
	}
	return pairs, units
}

// diffChips are the S17 chips of the changed values (Q5): mortalitate,
// sacrificare (a cull or its reason), eroare_sexare, oua.
func diffChips(diffs []dailyDiff) []string {
	has := map[string]bool{}
	for _, d := range diffs {
		switch {
		case d.Eggs:
			has[chipEggs] = true
		case d.Change.Field == string(messages.FieldMortality):
			has[chipMortality] = true
		case d.Change.Field == string(messages.FieldSexingError):
			has[chipSexing] = true
		default:
			has[chipCull] = true
		}
	}
	out := []string{}
	for _, c := range []string{chipMortality, chipCull, chipSexing, chipEggs} {
		if has[c] {
			out = append(out, c)
		}
	}
	return out
}

// diffSections are the codes of the sections and egg units with a change,
// in diff order.
func diffSections(diffs []dailyDiff) []string {
	out := []string{}
	for _, d := range diffs {
		if !slices.Contains(out, d.Change.Section) {
			out = append(out, d.Change.Section)
		}
	}
	return out
}

// diffLabel and shownValue name a changed value in the Motiv sentence and the
// Toasts.
func diffLabel(d dailyDiff) messages.Label {
	c := d.Change
	if d.Eggs {
		return messages.EggLabel(c.Section, messages.EggField(c.Field))
	}
	return messages.BirdLabel(c.Section, c.Sex, messages.DailyField(c.Field))
}

func shownValue(v audit.Value, field string) messages.Shown {
	if s, ok := v.Text(); ok {
		return messages.ShownText(s)
	}
	n, ok := v.Num()
	switch {
	case !ok:
		return messages.ShownEmpty()
	case field == string(messages.FieldAvgWeightG):
		return messages.ShownDec(n.Decimal(), 2)
	default:
		return messages.ShownInt(int(n.Decimal().IntPart()))
	}
}

// reasonMissing is the §6.15 400 on `reason`: "Motiv lipsă. Scrie de ce se
// schimbă H5.2 ♀ din 1 în 3; …", or "… 3 valori; …" for several changes.
func reasonMissing(diffs []dailyDiff) string {
	if len(diffs) == 1 {
		d := diffs[0]
		return messages.ReasonMissingChange(diffLabel(d), shownValue(d.Change.Old, d.Change.Field), shownValue(d.Change.New, d.Change.Field))
	}
	return messages.ReasonMissingChanges(len(diffs))
}

// isBirdCount reports whether a diff is a bird count (mortality, cull,
// sexing error): its Toast names the section's birds.
func (d dailyDiff) isBirdCount() bool {
	return !d.Eggs && d.Change.Field != string(messages.FieldCullReason)
}

// correctionToast is the Toast of a saved correction (§7.8, Q15): one bird
// count "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281" (the section's
// birds of that sex at the end of today), one other value "Corecție salvată
// · H5 Cat. 1 13 → 15"; several values "Corecție salvată · 3 valori · în
// viață azi ♀ 12 561 · ♂ 794" when birds moved (the farm at the end of
// today), else "Corecție salvată · 3 valori".
func correctionToast(diffs []dailyDiff, f rowFigures) string {
	if len(diffs) == 1 {
		d := diffs[0]
		l, from, to := diffLabel(d), shownValue(d.Change.Old, d.Change.Field), shownValue(d.Change.New, d.Change.Field)
		if d.isBirdCount() {
			return messages.CorrectionSavedAlive(l, from, to, d.Change.Sex, countOf(f.sectionAfter(d.SectionID), d.Change.Sex))
		}
		return messages.CorrectionSaved(l, from, to)
	}
	if after := f.farmAfter(); after != nil {
		return messages.CorrectionsSavedAlive(len(diffs), after.F, after.M)
	}
	return messages.CorrectionsSaved(len(diffs))
}

// dailyRows are the child rows a Zilnic save writes (planDaily's rows, as
// Snapshot rows without ids, with codes): per section in order, ♀ before ♂,
// mortality (−n), cull (−n, its reason), sexing error (as typed); one egg
// row per egg unit with a value. v has noLosses written out.
func dailyRows(v DailyValues, codes map[uuid.UUID]string) Snapshot {
	var out Snapshot
	for _, s := range v.Sections {
		for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
			x := s.sex(sex)
			if x == nil {
				continue
			}
			row := func(t dbq.BirdEventType, qty int32, reason *string) {
				b := SnapshotBird{SectionID: s.SectionID, Section: codes[s.SectionID], Sex: sex, Type: t, Qty: qty}
				if reason != nil {
					r := dbq.CullReason(*reason)
					b.CullReason = &r
				}
				out.BirdEvents = append(out.BirdEvents, b)
			}
			if x.Mortality != nil {
				row(dbq.BirdEventTypeMortality, -int32Of(x.Mortality), nil)
			}
			if x.Cull != nil {
				row(dbq.BirdEventTypeCull, -int32Of(x.Cull), x.CullReason)
			}
			if x.SexingError != nil {
				row(dbq.BirdEventTypeSexingError, int32Of(x.SexingError), nil)
			}
		}
	}
	for _, g := range v.Eggs {
		if !g.hasValue() {
			continue
		}
		out.EggProduction = append(out.EggProduction, SnapshotEggs{
			SectionID: g.SectionID, Section: codes[g.SectionID], Floor: int32Ptr(g.Floor), Cat1: int32Ptr(g.Cat1),
			Cat2: int32Ptr(g.Cat2), Cat3: int32Ptr(g.Cat3), Hatching: int32Ptr(g.Hatching), AvgWeightG: g.AvgWeightG,
		})
	}
	return out
}

func int32Of(n *num.Num) int32 { return int32(n.Decimal().IntPart()) }

func int32Ptr(n *num.Num) *int32 {
	if n == nil {
		return nil
	}
	v := int32Of(n)
	return &v
}

// rowSections are the codes of the sections with a bird row, in row order.
func rowSections(rows []SnapshotBird) []string {
	out := []string{}
	for _, b := range rows {
		if !slices.Contains(out, b.Section) {
			out = append(out, b.Section)
		}
	}
	return out
}

// eggsTotal is Σ Cat. 1–3 + Incubabile of egg rows (D18).
func eggsTotal(rows []SnapshotEggs) int {
	n := 0
	for _, g := range rows {
		for _, x := range []*int32{g.Cat1, g.Cat2, g.Cat3, g.Hatching} {
			if x != nil {
				n += int(*x)
			}
		}
	}
	return n
}

// dailyAuditOf is a Zilnic save as its audit new_values (Q5: the create shape
// with codes), from v with noLosses written out.
func dailyAuditOf(v DailyValues, codes map[uuid.UUID]string) dailyAuditValues {
	out := dailyAuditValues{FarmID: v.FarmID, Date: v.Date, Sections: []dailyAuditSection{}, Eggs: []dailyAuditEggs{}}
	for _, s := range v.Sections {
		a := dailyAuditSection{SectionID: s.SectionID, Section: codes[s.SectionID]}
		if typed(s.F, true) {
			a.F = s.F
		}
		if typed(s.M, true) {
			a.M = s.M
		}
		if a.F != nil || a.M != nil {
			out.Sections = append(out.Sections, a)
		}
	}
	for _, g := range v.Eggs {
		if g.hasValue() {
			out.Eggs = append(out.Eggs, dailyAuditEggs{Section: codes[g.SectionID], DailyEggs: g})
		}
	}
	return out
}

// rowWrites are the child-row writes of a change of a Zilnic save: rows
// soft-deleted by id, rows changed in place by id, rows inserted.
type rowWrites struct {
	deleteBirds, deleteEggs []uuid.UUID
	updateBirds             []birdUpdate
	updateEggs              []eggUpdate
	insertBirds             []dailyBird
	insertEggs              []dailyEggRow
}

// birdUpdate is one EntryUpdateBirdRows row.
type birdUpdate struct {
	ID         uuid.UUID       `json:"id"`
	Qty        int32           `json:"qty"`
	CullReason *dbq.CullReason `json:"cull_reason"`
}

// eggUpdate is one EntryUpdateEggRows row.
type eggUpdate struct {
	ID         uuid.UUID `json:"id"`
	Floor      *int32    `json:"floor"`
	Cat1       *int32    `json:"cat1"`
	Cat2       *int32    `json:"cat2"`
	Cat3       *int32    `json:"cat3"`
	Hatching   *int32    `json:"hatching"`
	AvgWeightG *num.Num  `json:"avg_weight_g"`
}

func eggUpdateOf(id uuid.UUID, g SnapshotEggs) eggUpdate {
	return eggUpdate{ID: id, Floor: g.Floor, Cat1: g.Cat1, Cat2: g.Cat2, Cat3: g.Cat3, Hatching: g.Hatching, AvgWeightG: g.AvgWeightG}
}

// planRows matches the stored rows to the new save's by key (plan M5-D6): a
// bird value by section × sex × type, the eggs by egg unit. A value whose
// figures differ is changed in place (it keeps its row and its unique index
// entry), a stored value missing from the new save is soft-deleted, a new
// one is inserted.
func planRows(old, new Snapshot) rowWrites {
	var w rowWrites
	stored := map[birdKey]SnapshotBird{}
	for _, b := range old.BirdEvents {
		stored[birdKey{b.SectionID, b.Sex, b.Type}] = b
	}
	seen := map[birdKey]bool{}
	for _, b := range new.BirdEvents {
		k := birdKey{b.SectionID, b.Sex, b.Type}
		seen[k] = true
		s, ok := stored[k]
		switch {
		case !ok:
			var reason *string
			if b.CullReason != nil {
				r := string(*b.CullReason)
				reason = &r
			}
			w.insertBirds = append(w.insertBirds, dailyBird{SectionID: b.SectionID, Sex: b.Sex, Type: b.Type, Qty: int(b.Qty), CullReason: reason})
		case s.Qty != b.Qty || !sameReason(s.CullReason, b.CullReason):
			w.updateBirds = append(w.updateBirds, birdUpdate{ID: s.ID, Qty: b.Qty, CullReason: b.CullReason})
		}
	}
	for _, b := range old.BirdEvents {
		if !seen[birdKey{b.SectionID, b.Sex, b.Type}] {
			w.deleteBirds = append(w.deleteBirds, b.ID)
		}
	}
	eggs := map[uuid.UUID]SnapshotEggs{}
	for _, g := range old.EggProduction {
		eggs[g.SectionID] = g
	}
	seenEggs := map[uuid.UUID]bool{}
	for _, g := range new.EggProduction {
		seenEggs[g.SectionID] = true
		s, ok := eggs[g.SectionID]
		switch {
		case !ok:
			row := dailyEggRow{
				SectionID: g.SectionID, Floor: intPtr(g.Floor), Cat1: intPtr(g.Cat1), Cat2: intPtr(g.Cat2),
				Cat3: intPtr(g.Cat3), Hatching: intPtr(g.Hatching),
			}
			if g.AvgWeightG != nil {
				d := g.AvgWeightG.Decimal()
				row.AvgWeightG = &d
			}
			w.insertEggs = append(w.insertEggs, row)
		case !sameEggs(s, g):
			w.updateEggs = append(w.updateEggs, eggUpdateOf(s.ID, g))
		}
	}
	for _, g := range old.EggProduction {
		if !seenEggs[g.SectionID] {
			w.deleteEggs = append(w.deleteEggs, g.ID)
		}
	}
	return w
}

// birds reports whether any bird row is written.
func (w rowWrites) birds() bool {
	return len(w.deleteBirds)+len(w.updateBirds)+len(w.insertBirds) > 0
}

// apply runs the writes on the change's entry, in this order: soft deletes
// (deleted_at = the clock's instant), changes in place, inserts. A count of
// rows changed that is not the asked one (a row gone under the lock) is 409
// STALE rather than half a change.
func (w rowWrites) apply(ctx context.Context, tx *ChangeTx) error {
	e := tx.Entry
	if n := len(w.deleteBirds) + len(w.deleteEggs); n > 0 {
		got, err := tx.Q.EntrySoftDeleteRows(ctx, dbq.EntrySoftDeleteRowsParams{
			Now: tx.Now, EntryID: e.ID, BirdIds: nonNilIDs(w.deleteBirds), EggIds: nonNilIDs(w.deleteEggs),
			FeedIds: []uuid.UUID{}, DeliveryIds: []uuid.UUID{},
		})
		if err != nil {
			return fmt.Errorf("entries: soft-delete rows: %w", err)
		}
		if int(got) != n {
			return httpx.Stale()
		}
	}
	if len(w.updateBirds) > 0 {
		rows, err := jsonRows(w.updateBirds)
		if err != nil {
			return err
		}
		got, err := tx.Q.EntryUpdateBirdRows(ctx, dbq.EntryUpdateBirdRowsParams{EntryID: e.ID, Rows: rows})
		if err != nil {
			return fmt.Errorf("entries: update bird rows: %w", err)
		}
		if int(got) != len(w.updateBirds) {
			return httpx.Stale()
		}
	}
	if len(w.updateEggs) > 0 {
		rows, err := jsonRows(w.updateEggs)
		if err != nil {
			return err
		}
		got, err := tx.Q.EntryUpdateEggRows(ctx, dbq.EntryUpdateEggRowsParams{EntryID: e.ID, Rows: rows})
		if err != nil {
			return fmt.Errorf("entries: update egg rows: %w", err)
		}
		if int(got) != len(w.updateEggs) {
			return httpx.Stale()
		}
	}
	if len(w.insertBirds) > 0 {
		rows, err := jsonRows(w.insertBirds)
		if err != nil {
			return err
		}
		if err := tx.Q.DailyInsertBirdEvents(ctx, dbq.DailyInsertBirdEventsParams{
			EntryID: e.ID, FarmID: e.FarmID, Date: e.Date, Rows: rows,
		}); err != nil {
			return err // a 23505 goes to OnConflict after the transaction
		}
	}
	if len(w.insertEggs) > 0 {
		rows, err := jsonRows(w.insertEggs)
		if err != nil {
			return err
		}
		if err := tx.Q.DailyInsertEggProduction(ctx, dbq.DailyInsertEggProductionParams{
			EntryID: e.ID, FarmID: e.FarmID, Date: e.Date, Rows: rows,
		}); err != nil {
			return err
		}
	}
	return nil
}

func nonNilIDs(ids []uuid.UUID) []uuid.UUID {
	if ids == nil {
		return []uuid.UUID{}
	}
	return ids
}

func intPtr(n *int32) *int {
	if n == nil {
		return nil
	}
	v := int(*n)
	return &v
}

func sameReason(a, b *dbq.CullReason) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func sameInt32(a, b *int32) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func sameNum(a, b *num.Num) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Decimal().Equal(b.Decimal())
}

// sameEggs reports whether two egg rows hold the same figures.
func sameEggs(a, b SnapshotEggs) bool {
	return sameInt32(a.Floor, b.Floor) && sameInt32(a.Cat1, b.Cat1) && sameInt32(a.Cat2, b.Cat2) &&
		sameInt32(a.Cat3, b.Cat3) && sameInt32(a.Hatching, b.Hatching) && sameNum(a.AvgWeightG, b.AvgWeightG)
}

// down are the moved section × sex pairs whose count goes down (the only ones
// §6.1 can refuse), in move order.
func (m birdMove) down() []ledger.SectionSex {
	var out []ledger.SectionSex
	for _, p := range m.pairs {
		if m.delta[p] < 0 {
			out = append(out, p)
		}
	}
	return out
}
