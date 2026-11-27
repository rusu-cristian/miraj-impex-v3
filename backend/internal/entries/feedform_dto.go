package entries

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/forms/feed-daily (§7.8, M3 D8, Q1 proposed).
// Numbers are at full precision (the client rounds, §5.15).

// FeedFormResponse is the Hrană zilnică form state of a farm on a day. The
// hall heads ("H5 · în viață la începutul zilei ♀ 8 547 · ♂ 665") are summed
// by the client from Rows; "Stoc după azi" is Stock[].Kg minus the draft's kg
// of the type (the dry run gives the exact figures).
type FeedFormResponse struct {
	Farm    FormFarm   `json:"farm"`
	Date    timex.Day  `json:"date"`
	SAPT    *int       `json:"sapt"`    // null without a cycle on the date
	AgeDays *int       `json:"ageDays"` // null without a cycle on the date
	Cycle   *FormCycle `json:"cycle"`   // the cycle that contains the date
	// Guide is the guide's g/pasăre of SAPT(date) per sex: null without a
	// cycle or a guide; a null cell = no guide row for that week × sex.
	Guide *httpx.SexPair[*num.Num] `json:"guide"`
	// LockedReason is why the caller can't save this day: a vet on a past day
	// ("Pentru zilele trecute cere managerului."), or a closed series.
	LockedReason *string `json:"lockedReason"`
	// Rows are the active bird-holding sections × sex with birds at the start
	// of the day, in tree order (♀ before ♂); [] for a farm without birds.
	Rows       []FeedFormRow   `json:"rows"`
	EmptyHalls []FormEmptyHall `json:"emptyHalls"`
	// Existing are the entries already holding feed rows of the day, oldest
	// first, with what the caller may do with them (§3.2).
	Existing []FormEntry `json:"existing"`
	// Stock is the farm's stock per feed type at the end of today (Q5) on the
	// saved data (§5.8); [] for a farm without an open cycle.
	Stock     []FormStock     `json:"stock"`
	FeedTypes []FormFeedType  `json:"feedTypes"` // the active types; the client filters by row sex
	AsOf      timex.LocalTime `json:"asOf"`
}

// FormFarm names the form's farm.
type FormFarm struct {
	ID   uuid.UUID    `json:"id"`
	Name string       `json:"name"`
	Type dbq.FarmType `json:"type"`
}

// FormCycle is the cycle that contains the form's date.
type FormCycle struct {
	ID        uuid.UUID `json:"id"`
	Series    string    `json:"series"`
	StartDate timex.Day `json:"startDate"`
}

// FeedFormRow is one section × sex of the form.
type FeedFormRow struct {
	SectionID  uuid.UUID      `json:"sectionId"`
	Code       string         `json:"code"`
	HallCode   string         `json:"hallCode"`
	Sex        dbq.Sex        `json:"sex"`
	BirdsStart int            `json:"birdsStart"`
	GuideG     *num.Num       `json:"guideG"` // null without a guide or a guide row
	Suggestion FeedSuggestion `json:"suggestion"`
	// Existing is the row's live allocation of the day, null when none.
	Existing *FeedRowExisting `json:"existing"`
}

// FeedSuggestion prefills a row (§6.3, spec C2): yesterday's typed g (null
// when none or kg-only) and the type last used in the row on an earlier day.
type FeedSuggestion struct {
	G          *num.Num   `json:"g"`
	FeedTypeID *uuid.UUID `json:"feedTypeId"`
}

// FeedRowExisting is a row's saved allocation of the day.
type FeedRowExisting struct {
	EntryID    uuid.UUID `json:"entryId"`
	G          *num.Num  `json:"g"` // null for a migrated kg-only row
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	Kg         num.Num   `json:"kg"` // the migrated kg, else g × birdsStart ÷ 1000
}

// FormEmptyHall is a hall emptied during the cycle, as the summary's tiles
// show it: "goală din 14.11.2026 · H7.1, H7.2".
type FormEmptyHall struct {
	Hall       string    `json:"hall"`
	EmptySince timex.Day `json:"emptySince"`
	Label      string    `json:"label"`
}

// FormEntry is an entry that holds feed rows of the day.
type FormEntry struct {
	EntryID     uuid.UUID       `json:"entryId"`
	Author      *string         `json:"author"` // null for a migration entry
	At          timex.LocalTime `json:"at"`
	Rows        int             `json:"rows"` // its section × sex rows on the day
	Editable    bool            `json:"editable"`
	Correctable bool            `json:"correctable"`
}

// FormStock is one feed type's stock row (§5.8, as §7.12's) plus its id.
type FormStock struct {
	FeedTypeID uuid.UUID    `json:"feedTypeId"`
	Type       string       `json:"type"`
	ForSex     *string      `json:"forSex"` // null = ambele sexe
	Kg         num.Num      `json:"kg"`     // the stock now
	AvgDaily   num.Num      `json:"avgDaily"`
	Window     [2]timex.Day `json:"window"` // the farm's consumption window, both ends included
	DaysLeft   *num.Num     `json:"daysLeft"`
	RunOut     *timex.Day   `json:"runOut"`
	Lamp       calc.Status  `json:"lamp"`
	Reason     string       `json:"reason,omitzero"`
}

// FormFeedType is an active feed type.
type FormFeedType struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	ForSex *string   `json:"forSex"` // null = ambele sexe
}
