package entries

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// dailyKind is the Zilnic kind's adapter (kindAdapter, ARC-1): bird_event
// rows (mortality, cull, sexing error) and egg_production rows. Its PUT is
// daily_update.go, the undo of its update undo.go.
type dailyKind struct{}

// load reads the live bird rows (tree order) and egg rows of the entry.
func (dailyKind) load(ctx context.Context, q *dbq.Queries, s *StoredEntry) error {
	birds, err := q.EntryBirdRows(ctx, s.ID)
	if err != nil {
		return fmt.Errorf("entries: bird rows of %s: %w", s.ID, err)
	}
	eggs, err := q.EntryEggRows(ctx, s.ID)
	if err != nil {
		return fmt.Errorf("entries: egg rows of %s: %w", s.ID, err)
	}
	for _, b := range birds {
		s.Rows.BirdEvents = append(s.Rows.BirdEvents, SnapshotBird{
			ID: b.ID, SectionID: b.SectionID, Section: b.Code, Sex: b.Sex, Type: b.Type, Qty: b.Qty,
			CullReason: b.CullReason,
		})
		s.addSection(EntrySection{ID: b.SectionID, Code: b.Code, HallCode: b.HallCode, Retired: b.Retired})
	}
	for _, g := range eggs {
		s.Rows.EggProduction = append(s.Rows.EggProduction, SnapshotEggs{
			ID: g.ID, SectionID: g.SectionID, Section: g.Code, Floor: g.FloorEggs, Cat1: g.Cat1, Cat2: g.Cat2,
			Cat3: g.Cat3, Hatching: g.HatchingEggs, AvgWeightG: nullNum(g.AvgEggWeightG),
		})
		s.addSection(EntrySection{ID: g.SectionID, Code: g.Code, HallCode: g.Code, Retired: g.Retired})
	}
	return nil
}

func (dailyKind) values(s StoredEntry) any { return s.Daily() }

func (dailyKind) update(h *Handler, w http.ResponseWriter, r *http.Request, p auth.Principal, t updateTarget) (Result, error) {
	return h.updateDaily(w, r, p, t)
}

func (dailyKind) undoUpdate(ctx context.Context, tx *ChangeTx, before updateSnapshot) (Saved, error) {
	return planUndoUpdate(ctx, tx, before)
}

// createUndoStale: every change of a Zilnic save is audited on its entry.
func (dailyKind) createUndoStale(Snapshot) bool { return false }

// section is a compartment's bird rows, else an egg unit's eggs (D98).
func (dailyKind) section(stored StoredEntry, id uuid.UUID) (Snapshot, error) {
	var out Snapshot
	for _, b := range stored.Rows.BirdEvents {
		if b.SectionID == id {
			out.BirdEvents = append(out.BirdEvents, b)
		}
	}
	if len(out.BirdEvents) > 0 {
		return out, nil
	}
	for _, g := range stored.Rows.EggProduction {
		if g.SectionID == id {
			out.EggProduction = append(out.EggProduction, g)
		}
	}
	if len(out.EggProduction) == 0 {
		return Snapshot{}, badSection()
	}
	return out, nil
}

// removalEffect: a section's values (S15), a hall's eggs, or the whole save
// with the farm's birds before → after.
func (dailyKind) removalEffect(c rowsText) string {
	e, f := c.Entry, c.Figures
	if c.SectionID != nil {
		if len(c.Rows.BirdEvents) == 0 {
			return messages.HallEggsDeleteEffect(e.CreatedAt, f.eggs)
		}
		id := *c.SectionID
		before, after := f.today[id], f.sectionAfter(id)
		return messages.SectionDeleteEffect(e.CreatedAt, f.birds.codes[id], f.birds.qty.F, f.birds.qty.M,
			aliveMove(before.F, after.F), aliveMove(before.M, after.M))
	}
	after := f.farmAfter()
	d := messages.DailyDelete{
		Author: authorOf(c.Info), At: e.CreatedAt, Corrected: correctedBy(e, c.Info), Sections: len(f.birds.sections),
		F: f.birds.qty.F, M: f.birds.qty.M, Halls: len(f.halls), Eggs: f.eggs, Farm: c.Info.FarmName,
	}
	if after != nil {
		d.AliveF, d.AliveM = aliveMove(f.farm.F, after.F), aliveMove(f.farm.M, after.M)
	}
	return messages.DailyDeleteEffect(d)
}

func (dailyKind) removalToast(c rowsText) string {
	e, f := c.Entry, c.Figures
	if c.Undo {
		return messages.DailyUndone(e.Date)
	}
	if len(c.Rows.BirdEvents) == 0 {
		return messages.EggsDeleted(e.Date, f.halls)
	}
	aliveF, aliveM := f.namedAlive(c.SectionID)
	if c.SectionID != nil {
		return messages.MortalitySectionDeleted(f.birds.codes[*c.SectionID], aliveF, aliveM)
	}
	return messages.MortalityDeleted(e.Date, len(f.birds.sections), aliveF, aliveM)
}

func (dailyKind) restoreToast(c rowsText) string {
	e, f := c.Entry, c.Figures
	if len(c.Rows.BirdEvents) == 0 {
		return messages.EggsRestored(e.Date, f.halls)
	}
	aliveF, aliveM := f.namedAlive(c.SectionID)
	if c.SectionID != nil {
		return messages.MortalitySectionRestored(f.birds.codes[*c.SectionID], aliveF, aliveM)
	}
	return messages.MortalityRestored(e.Date, len(f.birds.sections), aliveF, aliveM)
}

// summary: per section "Mortalitate · H5.2" / "Ouă · H5", a whole save
// "Mortalitate · 4 secțiuni" (· "Ouă · 2 hale").
func (dailyKind) summary(c rowsText) string {
	f := c.Figures
	if c.SectionID != nil {
		if len(c.Rows.BirdEvents) > 0 {
			return messages.MortalitySectionSummary(f.birds.sections[0])
		}
		return messages.EggsHallSummary(f.halls[0])
	}
	return messages.DailySummary(len(f.birds.sections), len(f.halls))
}

// holderMessage: the Z9 sentence of the bird value or of the eggs.
func (dailyKind) holderMessage(h dbq.EntryRestoreHolderRow, day timex.Day) string {
	if h.Tbl == "eggs" {
		return messages.EggsAlreadyEntered(h.Code, day, h.CreatedAt)
	}
	return messages.DailyAlreadyEntered(h.Code, h.Sex, day, h.CreatedAt)
}

// badSection is the 400 of a per-section delete whose sectionId names no
// live row of the entry, or of an entry kind without sections (D98).
func badSection() error {
	return httpx.Validation(httpx.Issue{Path: "sectionId", Code: httpx.IssueOutOfRange, Message: messages.InvalidData()})
}

// Daily is a daily entry's rows as DailyValues: one section per section with
// a bird value (in row order: tree order), a sex only when it has a value,
// mortality and cull as positive counts (−qty), the sexing error as
// recorded, the cull reason as stored (boala | schiopatare | altul); one egg
// row per egg unit. noLosses is never set: a "Fără pierderi azi" save is
// stored, and read back, as its 0 values.
func (s StoredEntry) Daily() DailyValues {
	v := DailyValues{FarmID: s.FarmID, Date: s.Date, Sections: []DailySection{}, Eggs: []DailyEggs{}}
	for _, b := range s.Rows.BirdEvents {
		i := slices.IndexFunc(v.Sections, func(d DailySection) bool { return d.SectionID == b.SectionID })
		if i < 0 {
			v.Sections = append(v.Sections, DailySection{SectionID: b.SectionID})
			i = len(v.Sections) - 1
		}
		sex := &v.Sections[i].F
		if b.Sex == dbq.SexM {
			sex = &v.Sections[i].M
		}
		if *sex == nil {
			*sex = &DailySex{}
		}
		switch b.Type {
		case dbq.BirdEventTypeMortality:
			(*sex).Mortality = intNum(-int64(b.Qty))
		case dbq.BirdEventTypeCull:
			(*sex).Cull = intNum(-int64(b.Qty))
			if b.CullReason != nil {
				r := string(*b.CullReason)
				(*sex).CullReason = &r
			}
		case dbq.BirdEventTypeSexingError:
			(*sex).SexingError = intNum(int64(b.Qty))
		}
	}
	for _, g := range s.Rows.EggProduction {
		v.Eggs = append(v.Eggs, DailyEggs{
			SectionID: g.SectionID, Floor: int32Num(g.Floor), Cat1: int32Num(g.Cat1), Cat2: int32Num(g.Cat2),
			Cat3: int32Num(g.Cat3), Hatching: int32Num(g.Hatching), AvgWeightG: g.AvgWeightG,
		})
	}
	return v
}
