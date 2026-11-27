package farms

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
)

// BuildBootstrap builds GET /api/bootstrap (§7.2, M2 D7) on db, scoped to
// in.Principal: every farm for the owner, the user_farm farms for anyone
// else. Series are those with a cycle on those farms, their saptSpan over
// those cycles; feed types are the active ones; settings come from
// app_setting. Only Principal and Today of in are used.
func BuildBootstrap(ctx context.Context, db dbq.DBTX, in Input) (BootstrapResponse, error) {
	q := dbq.New(db)
	all := in.Principal.Role == auth.Owner
	ids := in.Principal.FarmIDs
	if ids == nil {
		ids = []uuid.UUID{}
	}

	farms, err := q.FarmsListFarms(ctx, dbq.FarmsListFarmsParams{AllFarms: all, FarmIds: ids})
	if err != nil {
		return BootstrapResponse{}, fmt.Errorf("farms: bootstrap farms: %w", err)
	}
	sections, err := q.FarmsSections(ctx, dbq.FarmsSectionsParams{AllFarms: all, FarmIds: ids})
	if err != nil {
		return BootstrapResponse{}, fmt.Errorf("farms: bootstrap sections: %w", err)
	}
	series, err := q.FarmsSeries(ctx, dbq.FarmsSeriesParams{Today: in.Today, AllFarms: all, FarmIds: ids})
	if err != nil {
		return BootstrapResponse{}, fmt.Errorf("farms: bootstrap series: %w", err)
	}
	feedTypes, err := q.FarmsActiveFeedTypes(ctx)
	if err != nil {
		return BootstrapResponse{}, fmt.Errorf("farms: bootstrap feed types: %w", err)
	}
	settings, err := calc.ReadAppSettings(ctx, db)
	if err != nil {
		return BootstrapResponse{}, err
	}

	res := BootstrapResponse{
		Farms:     make([]BootstrapFarm, 0, len(farms)),
		Series:    make([]SeriesItem, 0, len(series)),
		FeedTypes: make([]FeedTypeItem, 0, len(feedTypes)),
		Settings: Settings{
			ToleranceWeightPct: num.Of(settings.ToleranceWeightPct),
			ToleranceLayPts:    num.Of(settings.ToleranceLayPts),
			ToleranceFeedPct:   num.Of(settings.ToleranceFeedPct),
			StockDaysMin:       num.Of(settings.StockDaysMin),
		},
	}
	trees := sectionTrees(sections)
	for _, f := range farms {
		bf := BootstrapFarm{
			ID: f.ID, Name: f.Name, Type: auth.FarmType(f.Type), Guide: guideRef(f.GuideID, f.GuideName),
			Sections: trees[f.ID],
		}
		if bf.Sections == nil {
			bf.Sections = []SectionNode{}
		}
		if f.CycleID != nil && f.SeriesID != nil && f.SeriesName != nil && f.StartDate != nil &&
			f.PlacementDate != nil && f.AgeAtPlacementDays != nil {
			age := calc.SeriesAge{PlacementDate: *f.PlacementDate, AgeAtPlacementDays: int(*f.AgeAtPlacementDays)}
			bf.OpenCycle = &OpenCycle{
				ID: *f.CycleID, SeriesID: *f.SeriesID, SeriesName: *f.SeriesName, StartDate: *f.StartDate,
				SaptStart: calc.SAPT(age, *f.StartDate), SaptToday: calc.SAPT(age, in.Today),
				AgeDaysToday: calc.AgeDays(age, in.Today),
			}
		}
		res.Farms = append(res.Farms, bf)
	}
	for _, s := range series {
		age := calc.SeriesAge{PlacementDate: s.PlacementDate, AgeAtPlacementDays: int(s.AgeAtPlacementDays)}
		res.Series = append(res.Series, SeriesItem{
			ID: s.ID, Name: s.Name, Status: string(s.Status),
			SaptSpan: [2]int{calc.SAPT(age, s.FirstDay), calc.SAPT(age, s.LastDay)},
		})
	}
	for _, t := range feedTypes {
		var forSex *string
		if t.ForSex != nil {
			s := string(*t.ForSex)
			forSex = &s
		}
		res.FeedTypes = append(res.FeedTypes, FeedTypeItem{ID: t.ID, Name: t.Name, ForSex: forSex, Active: t.Active})
	}
	return res, nil
}

// sectionTrees groups sections (siblings in order) into each farm's tree:
// the top-level sections with a hall's compartments as its children. A
// compartment whose hall is retired is left out with it.
func sectionTrees(rows []dbq.FarmsSectionsRow) map[uuid.UUID][]SectionNode {
	trees := map[uuid.UUID][]SectionNode{}
	at := map[uuid.UUID]int{} // top-level section id → index in its farm's tree
	for _, s := range rows {
		if s.ParentID == nil {
			at[s.ID] = len(trees[s.FarmID])
			trees[s.FarmID] = append(trees[s.FarmID], SectionNode{
				ID: s.ID, Code: s.Code, IsHall: s.IsHall, Children: []SectionLeaf{},
			})
		}
	}
	for _, s := range rows {
		if s.ParentID == nil {
			continue
		}
		i, ok := at[*s.ParentID]
		if !ok {
			continue
		}
		tree := trees[s.FarmID]
		tree[i].Children = append(tree[i].Children, SectionLeaf{ID: s.ID, Code: s.Code})
	}
	return trees
}
