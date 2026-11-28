package entries

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// DailyCreateRequest is the body of POST /api/entries/daily (§7.8): the
// Zilnic values of one farm and day (DailyValues, the shape PUT sends back)
// with the create's WriteMeta.
type DailyCreateRequest struct {
	httpx.WriteMeta // clientRequestId, source, voiceTranscript, voiceToken, dryRun
	DailyValues
}

// DailyRefs is what the level-1 check of a Zilnic save needs from the
// database (§7.0b): the codes of the farm's sections, halls and retired ones
// included (those are Z3 / Z7, level-2 refusals).
type DailyRefs struct {
	Sections map[uuid.UUID]string // section id → code
}

// Bounds of a Zilnic save's level-1 checks (§6.2 Z2, Z8; §4.8 columns).
var (
	eggWeightMin = decimal.NewFromInt(20)
	eggWeightMax = decimal.NewFromInt(120)
	int32Max     = decimal.NewFromInt(2147483647)
	int32Min     = decimal.NewFromInt(-2147483647) // a removal is stored as −count, so counts stay within ±(2³¹ − 1)
)

// cullReasons are the stored cull reasons (§4.8, D36).
var cullReasons = map[string]bool{
	string(dbq.CullReasonBoala): true, string(dbq.CullReasonSchiopatare): true, string(dbq.CullReasonAltul): true,
}

// Validate is the level-1 check of a Zilnic save (§7.0b, plan M5-D13): the
// WriteMeta fields, then DailyValues.ValidateAt.
func (r DailyCreateRequest) Validate(v *httpx.Validator, refs DailyRefs) {
	ValidateWriteMeta(v, r.WriteMeta)
	r.DailyValues.ValidateAt(v, refs, "")
}

// ValidateAt is the level-1 check of a Zilnic save's values (§7.0b, §6.2,
// plan M5-D13), every path prefixed with prefix ("" for POST, "values." for
// PUT). In document order: farmId and date required; at least one typed value
// or "Fără pierderi azi" (an empty section or egg row writes nothing); per
// section: sectionId (required, a section of this farm, once in the list),
// noLosses not with a typed mortality or cull of that section, per sex:
// counts integral ≥ 0 (Z2 with the section's sentence), the sexing error
// integral and signed, all within int32, cullReason in the enum and only with
// a cull value; per egg row: sectionId as above, counts integral ≥ 0 (Z8),
// avgWeightG 20–120 g (Z8 sentence) with at most 2 decimals (numeric(5,2)
// would round it silently, M3 Q9). The section checks need refs; they are
// skipped when farmId is missing (its own issue says so).
func (d DailyValues) ValidateAt(v *httpx.Validator, refs DailyRefs, prefix string) {
	v.Check(d.FarmID != uuid.Nil, prefix+"farmId", httpx.IssueRequired, messages.InvalidData())
	v.Check(!d.Date.IsZero(), prefix+"date", httpx.IssueRequired, messages.InvalidData())
	if !v.Check(d.hasValue(), prefix+"sections", httpx.IssueRequired, messages.InvalidData()) {
		return
	}
	known := d.FarmID != uuid.Nil
	section := func(path string, id uuid.UUID, seen map[uuid.UUID]bool) (string, bool) {
		code, ok := refs.Sections[id]
		switch {
		case id == uuid.Nil:
			v.Add(path+".sectionId", httpx.IssueRequired, messages.InvalidData())
		case known && !ok:
			v.Add(path+".sectionId", httpx.IssueOutOfRange, messages.InvalidData())
		case seen[id]:
			v.Add(path, httpx.IssueOutOfRange, messages.InvalidData())
		}
		seen[id] = true
		return code, ok
	}

	seen := make(map[uuid.UUID]bool, len(d.Sections))
	for i, s := range d.Sections {
		path := prefix + "sections[" + strconv.Itoa(i) + "]"
		code, ok := section(path, s.SectionID, seen)
		if s.NoLosses && (typed(s.F, false) || typed(s.M, false)) {
			v.Add(path+".noLosses", httpx.IssueOutOfRange, messages.InvalidData())
		}
		for _, sx := range []struct {
			sex dbq.Sex
			val *DailySex
		}{{dbq.SexF, s.F}, {dbq.SexM, s.M}} {
			if sx.val == nil {
				continue
			}
			p := path + "." + string(sx.sex)
			birdCount := func(field string, n *num.Num, signed bool) {
				if n == nil {
					return
				}
				msg := messages.InvalidData() // Z2 names the section and sex
				if ok {
					msg = messages.NotAnInteger(code, sx.sex)
				}
				x := n.Decimal()
				switch {
				case !x.IsInteger():
					v.Add(p+"."+field, httpx.IssueNotInteger, msg)
				case !signed && x.IsNegative():
					v.Add(p+"."+field, httpx.IssueOutOfRange, msg)
				case x.GreaterThan(int32Max) || x.LessThan(int32Min):
					v.Add(p+"."+field, httpx.IssueOutOfRange, messages.InvalidData())
				}
			}
			birdCount("mortality", sx.val.Mortality, false)
			birdCount("cull", sx.val.Cull, false)
			birdCount("sexingError", sx.val.SexingError, true)
			if r := sx.val.CullReason; r != nil {
				switch {
				case !cullReasons[*r]:
					v.Add(p+".cullReason", httpx.IssueOutOfRange, messages.InvalidData())
				case sx.val.Cull == nil:
					v.Add(p+".cullReason", httpx.IssueOutOfRange, messages.InvalidData())
				}
			}
		}
	}

	seen = make(map[uuid.UUID]bool, len(d.Eggs))
	for i, e := range d.Eggs {
		path := prefix + "eggs[" + strconv.Itoa(i) + "]"
		code, ok := section(path, e.SectionID, seen)
		for _, c := range []struct {
			field string
			n     *num.Num
		}{{"floor", e.Floor}, {"cat1", e.Cat1}, {"cat2", e.Cat2}, {"cat3", e.Cat3}, {"hatching", e.Hatching}} {
			if c.n == nil {
				continue
			}
			msg := messages.InvalidData() // Z8 names the hall
			if ok {
				msg = messages.EggCountNotAnInteger(code)
			}
			x := c.n.Decimal()
			switch {
			case !x.IsInteger():
				v.Add(path+"."+c.field, httpx.IssueNotInteger, msg)
			case x.IsNegative():
				v.Add(path+"."+c.field, httpx.IssueOutOfRange, msg)
			case x.GreaterThan(int32Max):
				v.Add(path+"."+c.field, httpx.IssueOutOfRange, messages.InvalidData())
			}
		}
		if e.AvgWeightG != nil {
			g := e.AvgWeightG.Decimal()
			switch {
			case g.LessThan(eggWeightMin) || g.GreaterThan(eggWeightMax):
				msg := messages.InvalidData()
				if ok {
					msg = messages.EggWeightOutOfRange(code, g)
				}
				v.Add(path+".avgWeightG", httpx.IssueOutOfRange, msg)
			case !g.Equal(g.Round(2)):
				v.Add(path+".avgWeightG", httpx.IssueOutOfRange, messages.InvalidData())
			}
		}
	}
}

// hasValue reports whether the values write anything: a typed bird value, a
// "Fără pierderi azi" or a typed egg value.
func (d DailyValues) hasValue() bool {
	for _, s := range d.Sections {
		if s.NoLosses || typed(s.F, true) || typed(s.M, true) {
			return true
		}
	}
	for _, e := range d.Eggs {
		if e.hasValue() {
			return true
		}
	}
	return false
}

// typed reports whether a sex holds a removal (mortality or cull) or, with
// sexing, any bird value.
func typed(s *DailySex, sexing bool) bool {
	if s == nil {
		return false
	}
	return s.Mortality != nil || s.Cull != nil || (sexing && s.SexingError != nil)
}

// hasValue reports whether an egg row holds any value (an empty row is not
// written).
func (e DailyEggs) hasValue() bool {
	return e.Floor != nil || e.Cat1 != nil || e.Cat2 != nil || e.Cat3 != nil || e.Hatching != nil || e.AvgWeightG != nil
}

// EntryWriteResponse is the 201 body of a Zilnic save (§7.8), the shape
// SPEC shares by every entry create, edit, correction, delete, undo and
// restore: the entry, the warnings (Z6), the Toast, the farm's birds at the
// end of today after the save, asOf (§7.0).
type EntryWriteResponse struct {
	Entry    httpx.EntryRef      `json:"entry"`
	Warnings []httpx.Warning     `json:"warnings"`
	Toast    string              `json:"toast"`
	Alive    *httpx.SexPair[int] `json:"alive,omitzero"`
	AsOf     timex.LocalTime     `json:"asOf"`
}

// dailyAuditValues is a Zilnic save's audit new_values (plan M5 Q5): the
// save in the create shape as written ("Fără pierderi azi" expanded into its
// 0 values), with the section codes.
type dailyAuditValues struct {
	FarmID   uuid.UUID           `json:"farmId"`
	Date     timex.Day           `json:"date"`
	Sections []dailyAuditSection `json:"sections"`
	Eggs     []dailyAuditEggs    `json:"eggs"`
}

// dailyAuditSection is one section of dailyAuditValues.
type dailyAuditSection struct {
	SectionID uuid.UUID `json:"sectionId"`
	Section   string    `json:"section"`
	F         *DailySex `json:"F"`
	M         *DailySex `json:"M"`
}

// dailyAuditEggs is one egg row of dailyAuditValues.
type dailyAuditEggs struct {
	Section string `json:"section"`
	DailyEggs
}
