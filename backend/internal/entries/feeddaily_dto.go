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

// FeedDailyCreateRequest is the body of POST /api/entries/feed-daily (§7.8):
// the confirmed Hrană zilnică rows of one farm and day (an unconfirmed
// suggestion is never sent, D13).
type FeedDailyCreateRequest struct {
	httpx.WriteMeta                // clientRequestId, source, voiceTranscript, voiceToken, dryRun
	FarmID          uuid.UUID      `json:"farmId"`
	Date            timex.Day      `json:"date"`
	Rows            []FeedDailyRow `json:"rows"`
}

// FeedDailyRow is one section × sex of the form.
type FeedDailyRow struct {
	SectionID  uuid.UUID `json:"sectionId"`
	Sex        dbq.Sex   `json:"sex"`
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	G          *num.Num  `json:"g"` // g/pasăre/zi, 1–300, at most one decimal (H1, Q9)
}

// FeedRefs are what the level-1 checks of a feed save need from the
// database (§7.0b): the codes of the farm's sections (halls and retired ones
// too: those are H3, a level-2 refusal) and the feed types that exist among
// the request's.
type FeedRefs struct {
	Sections  map[uuid.UUID]string // section id → code
	FeedTypes map[uuid.UUID]bool   // feed type id → exists
}

// Bounds of g/pasăre/zi (H1, §6.3).
var (
	feedGMin = decimal.NewFromInt(1)
	feedGMax = decimal.NewFromInt(300)
)

// Validate is the level-1 check of a feed save (§7.0b, M3 D5): the
// WriteMeta fields; farmId and date required; at least one row; per row, in
// document order: sectionId (required, a section of this farm), sex (F or M),
// feedTypeId (required, an existing type), g (required, H1 1–300 with the
// section's sentence, at most one decimal per Q9), and each (sectionId, sex)
// once. The section and type checks need refs; they are skipped when farmId
// is missing (its own issue says so).
func (r FeedDailyCreateRequest) Validate(v *httpx.Validator, refs FeedRefs) {
	ValidateWriteMeta(v, r.WriteMeta)
	v.Check(r.FarmID != uuid.Nil, "farmId", httpx.IssueRequired, messages.InvalidData())
	v.Check(!r.Date.IsZero(), "date", httpx.IssueRequired, messages.InvalidData())
	if !v.Check(len(r.Rows) > 0, "rows", httpx.IssueRequired, messages.InvalidData()) {
		return
	}
	known := r.FarmID != uuid.Nil
	type key struct {
		section uuid.UUID
		sex     dbq.Sex
	}
	seen := make(map[key]bool, len(r.Rows))
	for i, row := range r.Rows {
		path := "rows[" + strconv.Itoa(i) + "]"
		code, isSection := refs.Sections[row.SectionID]
		switch {
		case row.SectionID == uuid.Nil:
			v.Add(path+".sectionId", httpx.IssueRequired, messages.InvalidData())
		case known && !isSection:
			v.Add(path+".sectionId", httpx.IssueOutOfRange, messages.InvalidData())
		}
		sexOK := row.Sex == dbq.SexF || row.Sex == dbq.SexM
		switch {
		case row.Sex == "":
			v.Add(path+".sex", httpx.IssueRequired, messages.InvalidData())
		case !sexOK:
			v.Add(path+".sex", httpx.IssueOutOfRange, messages.InvalidData())
		}
		switch {
		case row.FeedTypeID == uuid.Nil:
			v.Add(path+".feedTypeId", httpx.IssueRequired, messages.InvalidData())
		case known && !refs.FeedTypes[row.FeedTypeID]:
			v.Add(path+".feedTypeId", httpx.IssueOutOfRange, messages.InvalidData())
		}
		if row.G == nil {
			v.Add(path+".g", httpx.IssueRequired, messages.InvalidData())
		} else if g := row.G.Decimal(); g.LessThan(feedGMin) || g.GreaterThan(feedGMax) {
			msg := messages.InvalidData() // the H1 sentence names the section and sex
			if isSection && sexOK {
				msg = messages.FeedGOutOfRange(code, row.Sex, g)
			}
			v.Add(path+".g", httpx.IssueOutOfRange, msg)
		} else if !g.Equal(g.Round(1)) { // numeric(6,1) would round it silently (Q9)
			v.Add(path+".g", httpx.IssueOutOfRange, messages.InvalidData())
		}
		if row.SectionID != uuid.Nil && sexOK {
			k := key{row.SectionID, row.Sex}
			v.Check(!seen[k], path, httpx.IssueOutOfRange, messages.InvalidData())
			seen[k] = true
		}
	}
}

// FeedDailyResponse is the 201 body of a feed save (§7.8, M3 D6): the
// entry, the total kg, the kg and the stock at the end of today after the
// save per feed type (Q5), the Toast, the H4/H5 warnings and asOf (§7.0).
type FeedDailyResponse struct {
	Entry    httpx.EntryRef  `json:"entry"`
	KgTotal  num.Num         `json:"kgTotal"`
	ByType   []FeedTypeKg    `json:"byType"`
	Toast    string          `json:"toast"`
	Warnings []httpx.Warning `json:"warnings"`
	AsOf     timex.LocalTime `json:"asOf"`
}

// FeedTypeKg is one feed type of a feed save, ordered by type name.
type FeedTypeKg struct {
	Type       string  `json:"type"`
	Kg         num.Num `json:"kg"`         // full precision
	StockAfter num.Num `json:"stockAfter"` // stock at the end of today after the save (Q5), full precision
}

// feedAuditRow is one row of a feed save's audit new_values (M3 D6).
type feedAuditRow struct {
	Section    string    `json:"section"`
	Sex        dbq.Sex   `json:"sex"`
	FeedTypeID uuid.UUID `json:"feedTypeId"`
	G          num.Num   `json:"g"`
	Kg         num.Num   `json:"kg"`
}
