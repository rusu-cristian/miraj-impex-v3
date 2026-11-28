package entries

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// feedDailyKind is the Hrană zilnică kind's adapter (kindAdapter, ARC-1):
// feed_allocation rows, one per section × sex. Its PUT and the undo of its
// update are feed_update.go.
type feedDailyKind struct{}

// load reads the live feed rows (tree order, ♀ before ♂) of the entry.
func (feedDailyKind) load(ctx context.Context, q *dbq.Queries, s *StoredEntry) error {
	rows, err := q.EntryFeedRows(ctx, s.ID)
	if err != nil {
		return fmt.Errorf("entries: feed rows of %s: %w", s.ID, err)
	}
	s.FeedTypes = []EntryFeedType{}
	for _, r := range rows {
		s.Rows.FeedAllocations = append(s.Rows.FeedAllocations, SnapshotFeed{
			ID: r.ID, SectionID: r.SectionID, Section: r.Code, Sex: r.Sex, FeedTypeID: r.FeedTypeID,
			Type: r.FeedTypeName, G: nullNum(r.GPerBird), KgRecorded: nullNum(r.KgRecorded),
		})
		s.addSection(EntrySection{ID: r.SectionID, Code: r.Code, HallCode: r.HallCode, Retired: r.Retired})
		s.addFeedType(EntryFeedType{ID: r.FeedTypeID, Name: r.FeedTypeName, ForSex: r.FeedTypeForSex, Active: r.FeedTypeActive})
	}
	return nil
}

func (feedDailyKind) values(s StoredEntry) any { return s.Feed() }

func (feedDailyKind) update(h *Handler, w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	return h.updateFeedDaily(w, r, p, t)
}

func (feedDailyKind) undoUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	return planUndoFeedUpdate(ctx, tx, before)
}

// createUndoStale: every change of a feed save is audited on its entry.
func (feedDailyKind) createUndoStale(Snapshot) bool { return false }

// section: a feed save is deleted whole.
func (feedDailyKind) section(StoredEntry, uuid.UUID) (Snapshot, error) {
	return Snapshot{}, badSection()
}

func (feedDailyKind) removalEffect(c rowsText) string {
	return messages.FeedDailyDeleteEffect(authorOf(c.Info), c.Entry.CreatedAt, correctedBy(c.Entry, c.Info),
		len(c.Rows.FeedAllocations), c.Figures.feedTotal)
}

func (feedDailyKind) removalToast(c rowsText) string {
	if c.Undo {
		return messages.FeedDailyUndone(c.Entry.Date)
	}
	return messages.FeedDailyDeleted(c.Entry.Date, len(c.Rows.FeedAllocations), c.Stock.typeName, c.Stock.kg)
}

func (feedDailyKind) restoreToast(c rowsText) string {
	return messages.FeedDailyRestored(c.Entry.Date, len(c.Rows.FeedAllocations), c.Stock.typeName, c.Stock.kg)
}

// summary: "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg".
func (feedDailyKind) summary(c rowsText) string {
	return messages.FeedDailySummary(len(c.Rows.FeedAllocations), c.Figures.feedOf, c.Figures.feedTotal)
}

// holderMessage: the Hrană sentence of the section × sex.
func (feedDailyKind) holderMessage(h dbq.EntryRestoreHolderRow, day timex.Day) string {
	return messages.FeedAlreadyEntered(h.Code, h.Sex, day, h.CreatedAt)
}

// Feed is a feed_daily entry's rows as FeedDailyValues (g null on an
// imported kg-only row).
func (s StoredEntry) Feed() FeedDailyValues {
	v := FeedDailyValues{FarmID: s.FarmID, Date: s.Date, Rows: []FeedDailyRow{}}
	for _, r := range s.Rows.FeedAllocations {
		v.Rows = append(v.Rows, FeedDailyRow{SectionID: r.SectionID, Sex: r.Sex, FeedTypeID: r.FeedTypeID, G: r.G})
	}
	return v
}
