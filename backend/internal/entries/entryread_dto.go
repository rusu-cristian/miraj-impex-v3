package entries

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of GET /api/entries/{id} (§7.8, M5 Q6 proposed): the whole save
// as its form (D172).

// EntryReadResponse is one saved entry with what the caller may do with it.
// Values is exactly the kind's create body without WriteMeta (DailyValues,
// FeedDailyValues, DeliveryValues), so it can be sent back as PUT `values`
// unchanged; Sections and FeedTypes name what it refers to.
type EntryReadResponse struct {
	Entry    EntryHead      `json:"entry"`
	Values   any            `json:"values"`
	Sections []EntrySection `json:"sections"`
	// FeedTypes are the types the values name, for feed_daily and delivery
	// entries only (left out for daily).
	FeedTypes []EntryFeedType  `json:"feedTypes,omitzero"`
	Actions   EntryReadActions `json:"actions"`
	AsOf      timex.LocalTime  `json:"asOf"`
}

// EntryHead is the entry's header.
type EntryHead struct {
	ID      uuid.UUID     `json:"id"`
	Kind    dbq.EntryKind `json:"kind"`
	Version string        `json:"version"` // PUT's ifVersion (§7.0, §12.5)
	Farm    EntryFarm     `json:"farm"`
	Date    timex.Day     `json:"date"`
	SAPT    *int          `json:"sapt"`    // null without a cycle on the date
	AgeDays *int          `json:"ageDays"` // null without a cycle on the date
	// Author is who saved it; null for a system or migration entry.
	Author    *EntryAuthor     `json:"author"`
	CreatedAt timex.LocalTime  `json:"createdAt"`
	Source    dbq.RecordSource `json:"source"`
	// Corrected is the last change's author and instant once a Corectează
	// marked the entry (Tag "corectat", "modificat de Ion Rusu · 18.11.2026,
	// 10:12"); null otherwise.
	Corrected *EntryCorrected `json:"corrected"`
}

// EntryFarm names the entry's farm.
type EntryFarm struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// EntryAuthor names the entry's author.
type EntryAuthor struct {
	Name string `json:"name"`
}

// EntryCorrected is who changed a corrected entry last, and when.
type EntryCorrected struct {
	By string          `json:"by"`
	At timex.LocalTime `json:"at"`
}

// EntrySection is a section the values name: its code, its hall's code (its
// own for a hall or a parentless section) and whether it is retired.
type EntrySection struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"code"`
	HallCode string    `json:"hallCode"`
	Retired  bool      `json:"retired"`
}

// EntryFeedType is a feed type the values name (null forSex = ambele sexe).
type EntryFeedType struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	ForSex *dbq.Sex  `json:"forSex"`
	Active bool      `json:"active"`
}

// EntryReadActions are the changes the caller may start on the entry (§3.2,
// D70, D71, M5 Q8): edit = PUT without a Motiv (the author, the creation
// day); correct = PUT with a Motiv (manager on own farms, owner) when edit
// isn't open; delete = DELETE with or without a Motiv; undo = the caller's
// own create or update of today is still the entry's latest action; restore
// = the caller's own delete of today (one section's values) is still its
// latest action. All false for a viewer and inside a closed series.
type EntryReadActions struct {
	Edit    bool `json:"edit"`
	Correct bool `json:"correct"`
	Delete  bool `json:"delete"`
	Undo    bool `json:"undo"`
	Restore bool `json:"restore"`
}
