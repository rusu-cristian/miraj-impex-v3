package entries

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/forms/daily (§6.2, §7.8, plan M5 Q2 proposed).
// Numbers are at full precision (the client rounds, §5.15: mean7 is shown
// with one decimal). Saved days are not re-checked: no warnings (D77).

// DailyFormResponse is the Zilnic form state of a farm on a day.
type DailyFormResponse struct {
	Farm    FormFarm   `json:"farm"`
	Date    timex.Day  `json:"date"`
	SAPT    *int       `json:"sapt"`    // null without a cycle on the date
	AgeDays *int       `json:"ageDays"` // null without a cycle on the date
	Cycle   *FormCycle `json:"cycle"`   // the cycle that contains the date
	// LockedReason is why the caller can't save this day: a vet on a past day
	// ("Pentru zilele trecute cere managerului."), or a closed series (§3.1).
	LockedReason *string `json:"lockedReason"`
	// Empty is the empty-farm state, present only when no cycle contains the
	// date: "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT
	// 1–21." with the hint (as the farm page, M2 Q4).
	Empty *DailyFormEmpty `json:"empty,omitzero"`
	// Halls are the form's groups in tree order: each hall with its shown
	// compartments and its eggs; a section without parent and children is its
	// own group (hall null). [] for a farm without birds.
	Halls      []DailyFormHall `json:"halls"`
	EmptyHalls []FormEmptyHall `json:"emptyHalls"`
	// Existing are the entries already holding Zilnic values of the day,
	// oldest first, with what the caller may do with them (§3.2).
	Existing []DailyFormEntry `json:"existing"`
	AsOf     timex.LocalTime  `json:"asOf"`
}

// DailyFormEmpty is the empty-farm state. Since is omitted for a farm whose
// cycles all start after the date (or that never had one).
type DailyFormEmpty struct {
	Since *timex.Day `json:"since,omitzero"`
	Label string     `json:"label"`
	Hint  string     `json:"hint"`
}

// DailyFormHall is one group of the form: a hall and its compartments, or
// a section without parent and children alone (Hall null).
type DailyFormHall struct {
	Hall     *DailyFormHallRef  `json:"hall"`
	Sections []DailyFormSection `json:"sections"`
	// Eggs is the group's egg unit, null at a growth farm and for a unit
	// without saved eggs that has no hens at the start of the day or is
	// retired.
	Eggs *DailyFormEggs `json:"eggs"`
}

// DailyFormHallRef names a hall.
type DailyFormHallRef struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
}

// DailyFormSection is one compartment of the form: an active section with
// birds at the start of the day or an inflow that day, or any section holding
// a saved value of the day.
type DailyFormSection struct {
	SectionID uuid.UUID `json:"sectionId"`
	Code      string    `json:"code"`
	// Sexes are the sexes present (§5.3: alive at the start of the day or an
	// inflow that day) and any sex holding a saved value, ♀ before ♂.
	Sexes      []dbq.Sex              `json:"sexes"`
	AliveStart httpx.SexPair[int]     `json:"aliveStart"`
	Yesterday  httpx.SexPair[int]     `json:"yesterday"` // deaths (mortality + culls) of the day before
	Mean7      httpx.SexPair[num.Num] `json:"mean7"`     // deaths of the 7 days before ÷ 7, full precision
	// Saved are the section's live values of the day per sex (as the create
	// shape: mortality and cull as positive counts, the sexing error signed);
	// null when the section holds none, a sex null when that sex holds none.
	Saved *httpx.SexPair[*DailySex] `json:"saved"`
}

// DailyFormEggs is a group's egg unit.
type DailyFormEggs struct {
	SectionID uuid.UUID `json:"sectionId"`
	Code      string    `json:"code"`
	HensStart int       `json:"hensStart"` // Σ ♀ alive at the start of the day of its compartments (D156)
	// GuideLayPct is the guide's lay % ♀ for SAPT(date) (§5.6 guide_lay, Z10);
	// null without a cycle, a guide or a guide cell.
	GuideLayPct *num.Num         `json:"guideLayPct"`
	Saved       *DailyEggsValues `json:"saved"` // the unit's live eggs of the day, null when none
}

// DailyEggsValues are an egg unit's saved values (DailyEggs without the
// section): Jos, Cat. 1–3, Incubabile, Masa medie.
type DailyEggsValues struct {
	Floor      *num.Num `json:"floor"`
	Cat1       *num.Num `json:"cat1"`
	Cat2       *num.Num `json:"cat2"`
	Cat3       *num.Num `json:"cat3"`
	Hatching   *num.Num `json:"hatching"`
	AvgWeightG *num.Num `json:"avgWeightG"`
}

// DailyFormEntry is an entry that holds Zilnic values of the day.
type DailyFormEntry struct {
	EntryID  uuid.UUID       `json:"entryId"`
	Author   *string         `json:"author"` // null for a migration entry
	At       timex.LocalTime `json:"at"`
	Sections []string        `json:"sections"` // the sections of its bird values, in tree order
	Eggs     []string        `json:"eggs"`     // the egg units of its eggs, in tree order
	// Editable and Correctable are the caller's §3.2 actions; both false for
	// an entry whose kind the change routes don't serve yet (Q1: the import's
	// migration entry).
	Editable    bool `json:"editable"`
	Correctable bool `json:"correctable"`
}
