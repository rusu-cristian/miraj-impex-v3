package entries

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The Zilnic save's state and level-2 plan (§6.2, §12.5 loadDailyState /
// planDaily, plan M5-D13). planDaily is pure: the save runs it under the farm
// lock, and dailyHolder runs it again on the pool to name the holder of a
// unique violation.

// dailyHolder is the entry holding one live Zilnic key of the day (Z9).
type dailyHolder struct {
	EntryID   uuid.UUID
	CreatedAt *time.Time
	CreatedBy *uuid.UUID
	Source    *dbq.RecordSource
}

// holderOf is the holder of a key read as nullable columns, nil when free.
func holderOf(id *uuid.UUID, at *time.Time, by *uuid.UUID, src *dbq.RecordSource) *dailyHolder {
	if id == nil {
		return nil
	}
	return &dailyHolder{EntryID: *id, CreatedAt: at, CreatedBy: by, Source: src}
}

// message is the Z9 sentence: the holder's creation time, or, for a holder
// without an author or imported (its time is the import's instant),
// messages.EntryAlreadySaved (M3 COR-1). eggs selects the egg form.
func (h dailyHolder) message(code string, sex dbq.Sex, day timex.Day, eggs bool) string {
	migrated := h.Source != nil && *h.Source == dbq.RecordSourceMigration
	switch {
	case h.CreatedBy == nil || h.CreatedAt == nil || migrated:
		return messages.EntryAlreadySaved()
	case eggs:
		return messages.EggsAlreadyEntered(code, day, *h.CreatedAt)
	default:
		return messages.DailyAlreadyEntered(code, sex, day, *h.CreatedAt)
	}
}

// dailySexSlot is one section × sex on the day.
type dailySexSlot struct {
	AliveStart int // birds at the start of the day
	Inflow     int // placements and transfers in saved that day
	Held       map[dbq.BirdEventType]*dailyHolder
}

// present: the sex has birds at the start of the day or an inflow that day
// (§5.3 "sexes present").
func (s dailySexSlot) present() bool { return s.AliveStart > 0 || s.Inflow > 0 }

// dailySection is one section of the farm on the day.
type dailySection struct {
	ID       uuid.UUID
	Code     string
	HallID   *uuid.UUID // the parent hall; nil for a parentless section
	HallCode string
	IsHall   bool
	Retired  bool
	Sex      map[dbq.Sex]dailySexSlot
	EggsHeld *dailyHolder
}

// dailyState is the farm on the day as a Zilnic save sees it, read under
// the farm lock (DailySaveSlots).
type dailyState struct {
	FarmName        string
	Growth          bool
	ClosedSeries    string // "" when the day is in no closed series' cycle
	MortalityFactor decimal.Decimal
	Sections        map[uuid.UUID]*dailySection
	HensStart       map[uuid.UUID]int // egg unit → Σ ♀ at the start of the day of its compartments
}

// loadDailyState reads the farm on day; ok is false for an unknown farm.
func loadDailyState(ctx context.Context, q *dbq.Queries, farmID uuid.UUID, day timex.Day) (dailyState, bool, error) {
	rows, err := q.DailySaveSlots(ctx, dbq.DailySaveSlotsParams{FarmID: farmID, Day: day})
	if err != nil || len(rows) == 0 {
		return dailyState{}, false, err
	}
	st := dailyState{
		FarmName: rows[0].FarmName, Growth: rows[0].FarmType == dbq.FarmTypeGrowth, ClosedSeries: rows[0].ClosedSeries,
		MortalityFactor: rows[0].MortalityFactor, Sections: map[uuid.UUID]*dailySection{}, HensStart: map[uuid.UUID]int{},
	}
	for _, r := range rows {
		if r.SectionID == nil || r.Code == nil {
			continue
		}
		sec := st.Sections[*r.SectionID]
		if sec == nil {
			sec = &dailySection{
				ID: *r.SectionID, Code: *r.Code, HallID: r.HallID, IsHall: r.IsHall, Retired: r.Retired,
				Sex:      map[dbq.Sex]dailySexSlot{},
				EggsHeld: holderOf(r.EggsEntryID, r.EggsCreatedAt, r.EggsCreatedBy, r.EggsSource),
			}
			if r.HallCode != nil {
				sec.HallCode = *r.HallCode
			}
			st.Sections[sec.ID] = sec
		}
		sec.Sex[r.Sex] = dailySexSlot{
			AliveStart: int(r.AliveStart), Inflow: int(r.Inflow),
			Held: map[dbq.BirdEventType]*dailyHolder{
				dbq.BirdEventTypeMortality:   holderOf(r.MortalityEntryID, r.MortalityCreatedAt, r.MortalityCreatedBy, r.MortalitySource),
				dbq.BirdEventTypeCull:        holderOf(r.CullEntryID, r.CullCreatedAt, r.CullCreatedBy, r.CullSource),
				dbq.BirdEventTypeSexingError: holderOf(r.SexingEntryID, r.SexingCreatedAt, r.SexingCreatedBy, r.SexingSource),
			},
		}
		if r.Sex == dbq.SexF && !r.IsHall { // hens of the egg unit: the hall, or the section itself
			unit := *r.SectionID
			if r.HallID != nil {
				unit = *r.HallID
			}
			st.HensStart[unit] += int(r.AliveStart)
		}
	}
	return st, true, nil
}

// dailyBird is one bird_event row of the save (§4.8).
type dailyBird struct {
	SectionID  uuid.UUID         `json:"section_id"`
	Sex        dbq.Sex           `json:"sex"`
	Type       dbq.BirdEventType `json:"type"`
	Qty        int               `json:"qty"` // the signed delta
	CullReason *string           `json:"cull_reason"`
}

// dailyEggRow is one egg_production row of the save (§4.8).
type dailyEggRow struct {
	SectionID  uuid.UUID        `json:"section_id"`
	Floor      *int             `json:"floor"`
	Cat1       *int             `json:"cat1"`
	Cat2       *int             `json:"cat2"`
	Cat3       *int             `json:"cat3"`
	Hatching   *int             `json:"hatching"`
	AvgWeightG *decimal.Decimal `json:"avg_weight_g"`
}

// total is Σ eggs of the row the way the register counts them (D18): Cat. 1
// + Cat. 2 + Cat. 3 + Incubabile, Jos apart; an empty field counts 0.
func (e dailyEggRow) total() int {
	n := 0
	for _, v := range []*int{e.Cat1, e.Cat2, e.Cat3, e.Hatching} {
		if v != nil {
			n += *v
		}
	}
	return n
}

// dailyRefusal is a level-2 refusal and the 422's error.message when it is
// the first one: the SaveBar text (Z5's "refuzat · …"), else its message.
type dailyRefusal struct {
	httpx.Refusal
	Banner string
}

// removal is one removal field of a section × sex: its JSON name and the
// birds it removes (a sexing error only when negative, as Z5 counts it).
type removal struct {
	Field messages.DailyField
	Count int
}

// dailyPair is one section × sex the save writes bird rows for.
type dailyPair struct {
	Path     string // "sections[2].M"
	Code     string
	Sex      dbq.Sex
	Removals []removal // in the Z5 naming order: mortality, cull, sexing error
}

// dailyPlan is the level-2 plan of a Zilnic save.
type dailyPlan struct {
	Refusals  []dailyRefusal
	Collision *feedCollision // the first live key held by another entry (Z9)
	Birds     []dailyBird
	Eggs      []dailyEggRow
	Pairs     map[ledger.SectionSex]dailyPair
	Affected  []ledger.SectionSex // the pairs in document order (§6.1 step 1)
	Z3        map[ledger.SectionSex]bool
	Audit     dailyAuditValues
	// Figures of the Toast, the summary and the live events.
	BirdCodes []string // sections with a bird row, in document order
	EggCodes  []string // egg units with a row, in document order
	SumF      int      // Σ signed qty ♀
	SumM      int
	EggsTotal int
}

// planDaily is the level-2 plan of a Zilnic save of v on st (§6.2, plan Q3),
// in document order (sections, ♀ before ♂, then eggs):
//   - Z3: a value (0 included) for a section × sex without birds at the start
//     of the day and without an inflow that day, on a hall or on a retired
//     section; "Fără pierderi azi" on a section without any sex present;
//   - Z9: a key live in another entry (the first one is Collision);
//   - Z4: a cull > 0 without its reason;
//   - Z5: mortality + cull + max(0, −sexing error) > birds at the start of the
//     day + inflows saved that day; the bound names the first removal field
//     (mortality, cull, sexing error) that can bring the total back alone
//     (boundOf);
//   - Z7: eggs on a compartment or at a growth farm; eggs on a unit without
//     hens at the start of the day (or retired), in the Z3 form;
//   - Z9 for eggs.
//
// "Fără pierderi azi" writes mortality 0 and cull 0 for the sexes present;
// an empty sex or egg row writes nothing.
func planDaily(st dailyState, v DailyValues) dailyPlan {
	p := dailyPlan{
		Pairs: map[ledger.SectionSex]dailyPair{}, Z3: map[ledger.SectionSex]bool{},
		Audit: dailyAuditValues{FarmID: v.FarmID, Date: v.Date, Sections: []dailyAuditSection{}, Eggs: []dailyAuditEggs{}},
	}
	day := v.Date
	refuse := func(r httpx.Refusal) { p.Refusals = append(p.Refusals, dailyRefusal{Refusal: r, Banner: r.Message}) }
	collide := func(path, code string, sex dbq.Sex, h *dailyHolder, eggs bool) {
		msg := h.message(code, sex, day, eggs)
		r := httpx.Refusal{Field: path, SectionCode: code, Message: msg}
		if !eggs {
			r.Sex = string(sex)
		}
		refuse(r)
		if p.Collision == nil {
			p.Collision = &feedCollision{EntryID: h.EntryID, Message: msg}
		}
	}

	for i, s := range v.Sections {
		path := "sections[" + strconv.Itoa(i) + "]"
		sec, ok := st.Sections[s.SectionID]
		if !ok { // Validate guarantees a section of this farm
			refuse(httpx.Refusal{Field: path + ".sectionId", Message: messages.InvalidData()})
			continue
		}
		usable := !sec.IsHall && !sec.Retired
		if s.NoLosses && (!usable || (!sec.Sex[dbq.SexF].present() && !sec.Sex[dbq.SexM].present())) {
			refuse(httpx.Refusal{Field: path + ".noLosses", SectionCode: sec.Code, Message: messages.NoBirdsOnDay(sec.Code, dbq.SexF, day)})
		}
		audit := dailyAuditSection{SectionID: sec.ID, Section: sec.Code}
		for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
			val := s.F
			if sex == dbq.SexM {
				val = s.M
			}
			slot := sec.Sex[sex]
			noLosses := s.NoLosses && usable && slot.present()
			if !typed(val, true) && !noLosses {
				continue
			}
			key := ledger.SectionSex{SectionID: sec.ID, Sex: sex}
			sp := path + "." + string(sex)
			if !usable || !slot.present() { // Z3
				refuse(httpx.Refusal{Field: sp, SectionCode: sec.Code, Sex: string(sex), Message: messages.NoBirdsOnDay(sec.Code, sex, day)})
				p.Z3[key] = true
				continue
			}

			w := DailySex{}
			if val != nil {
				w = *val
			}
			if noLosses { // Validate refuses noLosses with a typed mortality or cull
				zero := num.FromInt(0)
				w.Mortality, w.Cull, w.CullReason = &zero, &zero, nil
			}
			pair := dailyPair{Path: sp, Code: sec.Code, Sex: sex}
			count := func(n *num.Num) int { return int(n.Decimal().IntPart()) }
			var rows []dailyBird
			if w.Mortality != nil {
				rows = append(rows, dailyBird{SectionID: sec.ID, Sex: sex, Type: dbq.BirdEventTypeMortality, Qty: -count(w.Mortality)})
				pair.Removals = append(pair.Removals, removal{messages.FieldMortality, count(w.Mortality)})
			}
			if w.Cull != nil {
				rows = append(rows, dailyBird{SectionID: sec.ID, Sex: sex, Type: dbq.BirdEventTypeCull, Qty: -count(w.Cull), CullReason: w.CullReason})
				pair.Removals = append(pair.Removals, removal{messages.FieldCull, count(w.Cull)})
			}
			if w.SexingError != nil {
				se := count(w.SexingError)
				rows = append(rows, dailyBird{SectionID: sec.ID, Sex: sex, Type: dbq.BirdEventTypeSexingError, Qty: se})
				if se < 0 {
					pair.Removals = append(pair.Removals, removal{messages.FieldSexingError, -se})
				}
			}

			for _, r := range rows { // Z9: one refusal per section × sex, naming the first key held
				if h := slot.Held[r.Type]; h != nil {
					collide(sp, sec.Code, sex, h, false)
					break
				}
			}
			if w.Cull != nil && count(w.Cull) > 0 && w.CullReason == nil { // Z4
				refuse(httpx.Refusal{
					Field: sp + ".cullReason", SectionCode: sec.Code, Sex: string(sex),
					Message: messages.CullReasonMissing(sec.Code, sex, count(w.Cull), day),
				})
			}
			total := 0
			for _, r := range pair.Removals {
				total += r.Count
			}
			if available := slot.AliveStart + slot.Inflow; total > available { // Z5
				field, most := boundOf(pair.Removals, available-total)
				p.Refusals = append(p.Refusals, dailyRefusal{
					Refusal: httpx.Refusal{
						Field: sp + "." + string(field), SectionCode: sec.Code, Sex: string(sex), Max: &most,
						Message: messages.DayBoundExceeded(sec.Code, sex, total, available, day, field, most),
					},
					Banner: messages.DayBoundBanner(sec.Code, sex, total, most),
				})
			}

			p.Birds = append(p.Birds, rows...)
			for _, r := range rows {
				if sex == dbq.SexF {
					p.SumF += r.Qty
				} else {
					p.SumM += r.Qty
				}
			}
			p.Pairs[key] = pair
			p.Affected = append(p.Affected, key)
			if !slices.Contains(p.BirdCodes, sec.Code) {
				p.BirdCodes = append(p.BirdCodes, sec.Code)
			}
			wc := w
			if sex == dbq.SexF {
				audit.F = &wc
			} else {
				audit.M = &wc
			}
		}
		if audit.F != nil || audit.M != nil {
			p.Audit.Sections = append(p.Audit.Sections, audit)
		}
	}

	growthRefused := false
	for j, e := range v.Eggs {
		if !e.hasValue() {
			continue
		}
		path := "eggs[" + strconv.Itoa(j) + "]"
		sec, ok := st.Sections[e.SectionID]
		switch {
		case !ok: // Validate guarantees a section of this farm
			refuse(httpx.Refusal{Field: path + ".sectionId", Message: messages.InvalidData()})
			continue
		case st.Growth: // Z7: one refusal for the whole egg list
			if !growthRefused {
				refuse(httpx.Refusal{Field: "eggs", Message: messages.EggsGrowthFarm(st.FarmName)})
				growthRefused = true
			}
			continue
		case sec.HallID != nil: // Z7: a compartment
			refuse(httpx.Refusal{Field: path + ".sectionId", SectionCode: sec.Code, Message: messages.EggsOnHall(sec.HallCode)})
			continue
		case sec.Retired || st.HensStart[sec.ID] <= 0: // no hens at the start of the day (Q3)
			refuse(httpx.Refusal{Field: path + ".sectionId", SectionCode: sec.Code, Message: messages.NoBirdsOnDay(sec.Code, dbq.SexF, day)})
			continue
		}
		if sec.EggsHeld != nil {
			collide(path, sec.Code, "", sec.EggsHeld, true)
		}
		row := dailyEggRow{
			SectionID: sec.ID, Floor: intOf(e.Floor), Cat1: intOf(e.Cat1), Cat2: intOf(e.Cat2), Cat3: intOf(e.Cat3),
			Hatching: intOf(e.Hatching),
		}
		if e.AvgWeightG != nil {
			g := e.AvgWeightG.Decimal()
			row.AvgWeightG = &g
		}
		p.Eggs = append(p.Eggs, row)
		p.EggCodes = append(p.EggCodes, sec.Code)
		p.EggsTotal += row.total()
		p.Audit.Eggs = append(p.Audit.Eggs, dailyAuditEggs{Section: sec.Code, DailyEggs: e})
	}
	return p
}

func intOf(n *num.Num) *int {
	if n == nil {
		return nil
	}
	v := int(n.Decimal().IntPart())
	return &v
}

// boundOf is the §6.1 bound of a section × sex whose removals are slack
// birds too many (slack < 0): the first removal field, in the Z5 naming order
// (mortality, cull, sexing error), whose count + slack is ≥ 0, i.e. that can
// bring the count back to 0 alone, with that bound. When none can alone, the
// first field with the bound 0 (plan M5 T5: the case SPEC's "typed + min"
// leaves below 0). removals is never empty for a pair a removal broke.
func boundOf(removals []removal, slack int) (messages.DailyField, int) {
	for _, r := range removals {
		if r.Count+slack >= 0 {
			return r.Field, r.Count + slack
		}
	}
	return removals[0].Field, 0
}

// countRefusals are the §6.1 refusals of the pairs CheckCounts found below 0
// after the write's rows (plan M5-D4, Q4), in its order, and the banner of the
// first one (error.message), whose subject is action's (a create or an author
// edit "Salvarea", a correction "Corecția"): "Salvarea ar lăsa H7.1 cu −5 găini pe
// 14.11.2026 și cu −5 găini azi, 18.11.2026. Nicio secțiune nu poate scădea
// sub 0." Each refusal names the removal field boundOf picks, with the
// Corectează field sentence ("H6.2 ♂ 4: ar lăsa −3 cocoși azi, 18.11.2026.
// Scrie cel mult 1.") and its max; a pair the save holds no removal for (it
// was below 0 before, or only a positive sexing error) gets the banner on its
// sex.
func countRefusals(p dailyPlan, bad []ledger.Negative, today timex.Day, action messages.CountAction) ([]httpx.Refusal, string) {
	out := make([]httpx.Refusal, 0, len(bad))
	banner := ""
	for _, n := range bad {
		pair := p.Pairs[ledger.SectionSex{SectionID: n.SectionID, Sex: n.Sex}]
		b := messages.NegativeBanner(action, n.Code, n.Sex, n.FirstDate, n.FirstAliveEnd, n.TodayAliveEnd, today)
		if banner == "" {
			banner = b
		}
		first, last, low := n.FirstDate, n.LastDate, n.MinAliveEnd
		r := httpx.Refusal{
			Field: pair.Path, SectionCode: n.Code, Sex: string(n.Sex), FirstNegativeDate: &first, LastNegativeDate: &last,
			MinAliveEnd: &low, Message: b,
		}
		if len(pair.Removals) > 0 {
			field, most := boundOf(pair.Removals, n.MinAliveEnd)
			typedCount := 0
			for _, rm := range pair.Removals {
				if rm.Field == field {
					typedCount = rm.Count
				}
			}
			r.Field = pair.Path + "." + string(field)
			r.Max = &most
			r.Message = messages.RemovalBound(n.Code, n.Sex, typedCount, n.LastDate, n.MinAliveEnd, today, most)
		}
		out = append(out, r)
	}
	return out, banner
}

// Spike is one Z6 MORTALITY_SPIKE (§6.2): a section × sex whose typed
// mortality is above alert.mortality_factor × its 7-day mean.
type Spike struct {
	SectionID uuid.UUID
	Section   string // "H6.1"
	Sex       dbq.Sex
	Deaths    int             // the typed mortality
	Mean7     decimal.Decimal // full precision; the message shows one decimal
	Message   string          // "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). …"
}

// Warning is the spike as a response warning.
func (s Spike) Warning() httpx.Warning {
	return httpx.Warning{Code: httpx.WarnMortalitySpike, SectionCode: s.Section, Sex: string(s.Sex), Message: s.Message}
}

// MortalitySpikes is Z6 for the values v of a Zilnic save of farmID on day
// (§6.2, plan Q3, golden 33): per typed mortality of a bird-holding section,
// in document order (♀ before ♂), a spike when mortality > factor × mean7,
// compared at full precision, where mean7 = the section × sex's deaths
// (mortality + culls) of the 7 days before day ÷ 7 (calc.DeathsWindow) and
// factor = alert.mortality_factor; a 0 mean warns on any death (D53). It only
// reads (Verify-safe); the save runs the same spikesOf on the same reader
// with the factor from its own state read.
func MortalitySpikes(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day, v DailyValues) ([]Spike, error) {
	deaths, err := calc.DeathsWindow(ctx, db, farmID, day)
	if err != nil {
		return nil, err
	}
	settings, err := calc.ReadAppSettings(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("entries: mortality spikes: %w", err)
	}
	return spikesOf(deaths, settings.MortalityFactor, v), nil
}

// spikesOf is MortalitySpikes on read inputs.
func spikesOf(deaths []calc.SectionDeaths, factor decimal.Decimal, v DailyValues) []Spike {
	week := make(map[uuid.UUID]calc.SectionDeaths, len(deaths))
	for _, d := range deaths {
		week[d.SectionID] = d
	}
	out := []Spike{}
	for _, s := range v.Sections {
		d, ok := week[s.SectionID]
		if !ok { // a hall: Z3
			continue
		}
		for _, sex := range []dbq.Sex{dbq.SexF, dbq.SexM} {
			val, sum := s.F, d.Week.F
			if sex == dbq.SexM {
				val, sum = s.M, d.Week.M
			}
			if val == nil || val.Mortality == nil {
				continue
			}
			dead := val.Mortality.Decimal()
			mean := calc.Mean7(sum)
			if !dead.GreaterThan(factor.Mul(mean)) {
				continue
			}
			n := int(dead.IntPart())
			out = append(out, Spike{
				SectionID: s.SectionID, Section: d.Code, Sex: sex, Deaths: n, Mean7: mean,
				Message: messages.MortalitySpike(d.Code, sex, n, mean, factor),
			})
		}
	}
	return out
}

// hasDeaths reports whether v types a mortality above 0 (Z6 can't warn
// otherwise, so the save skips the deaths read).
func (d DailyValues) hasDeaths() bool {
	for _, s := range d.Sections {
		for _, x := range []*DailySex{s.F, s.M} {
			if x != nil && x.Mortality != nil && x.Mortality.Decimal().IsPositive() {
				return true
			}
		}
	}
	return false
}

// jsonRows encodes rows for jsonb_to_recordset (DailyInsertBirdEvents,
// DailyInsertEggProduction).
func jsonRows[T any](rows []T) ([]byte, error) {
	b, err := json.Marshal(rows)
	if err != nil {
		return nil, fmt.Errorf("entries: encode daily rows: %w", err)
	}
	return b, nil
}
