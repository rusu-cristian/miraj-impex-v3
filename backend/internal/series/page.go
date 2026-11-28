package series

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// ErrSeriesNotFound is returned for a series id that doesn't exist, and for
// a series with no cycle on the caller's farms (M4 Q4: no existence leak).
var ErrSeriesNotFound = errors.New("series: series not found")

// Input is what BuildPage takes (ARC-6): the caller, the series and the
// business day and instant, explicit so that the builder never reads
// app_today().
type Input struct {
	Principal auth.Principal
	SeriesID  uuid.UUID
	Today     timex.Day
	Now       time.Time
}

// BuildPage builds GET /api/series/{id}/page (§7.3, M4 D7, Q2–Q5) on db for
// in.Principal: the owner sees every farm, anyone else the user_farm farms
// (the role check is the route's). Run it in one snapshot (InReadTx): it reads
// the series, its farms and cycles, the transfer links, each visible farm's
// count today (active series) and each panel's cycle stats
// (farms.BuildCycleStats). ErrSeriesNotFound names an unknown series or one
// without a cycle on the caller's farms.
func BuildPage(ctx context.Context, db dbq.DBTX, in Input) (PageResponse, error) {
	q := dbq.New(db)
	s, err := q.SeriesGet(ctx, in.SeriesID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PageResponse{}, ErrSeriesNotFound
	}
	if err != nil {
		return PageResponse{}, fmt.Errorf("series: page series: %w", err)
	}
	all := in.Principal.Role == auth.Owner
	ids := in.Principal.FarmIDs
	if ids == nil {
		ids = []uuid.UUID{}
	}
	farmRows, err := q.SeriesFarms(ctx, dbq.SeriesFarmsParams{AllFarms: all, FarmIds: ids, SeriesID: s.ID})
	if err != nil {
		return PageResponse{}, fmt.Errorf("series: page farms: %w", err)
	}
	cycleRows, err := q.SeriesCycles(ctx, s.ID)
	if err != nil {
		return PageResponse{}, fmt.Errorf("series: page cycles: %w", err)
	}

	farmsByID := make(map[uuid.UUID]dbq.SeriesFarmsRow, len(farmRows))
	for _, f := range farmRows {
		farmsByID[f.ID] = f
	}
	cyclesByFarm := map[uuid.UUID][]dbq.SeriesCyclesRow{} // oldest first
	var mine []dbq.SeriesCyclesRow                        // the caller's cycles of the series, oldest first
	for _, c := range cycleRows {
		cyclesByFarm[c.FarmID] = append(cyclesByFarm[c.FarmID], c)
		if farmsByID[c.FarmID].Visible {
			mine = append(mine, c)
		}
	}
	if len(mine) == 0 {
		return PageResponse{}, ErrSeriesNotFound
	}
	transfers, err := q.SeriesTransfers(ctx, dbq.SeriesTransfersParams{AllFarms: all, FarmIds: ids, SeriesID: s.ID})
	if err != nil {
		return PageResponse{}, fmt.Errorf("series: page transfers: %w", err)
	}

	age := calc.SeriesAge{PlacementDate: s.PlacementDate, AgeAtPlacementDays: int(s.AgeAtPlacementDays)}
	active := s.Status == dbq.SeriesStatusActive
	res := PageResponse{
		Series: Info{
			ID: s.ID, Name: s.Name, Status: string(s.Status), PlacementDate: s.PlacementDate,
			PlacementInferred: s.PlacementInferred, AgeAtPlacementDays: int(s.AgeAtPlacementDays),
		},
		SaptSpan: saptSpanOf(age, mine, in.Today),
		Track:    Track{Lanes: []Lane{}, Today: in.Today},
		Farms:    make([]Panel, 0, len(farmRows)),
		Tabs:     []string{farms.TabPasari, farms.TabHrana, farms.TabGreutate},
		AsOf:     timex.At(in.Now),
	}
	if active {
		sapt := calc.SAPT(age, in.Today)
		res.SaptToday = &sapt
	} else {
		res.End = endOf(mine)
	}
	if first := cycleRows[0]; farmsByID[first.FarmID].Visible {
		res.Placement = &Placement{Date: first.StartDate, Inferred: first.StartInferred, Farm: farmsByID[first.FarmID].Name}
		// Only the birds placed at the caller's farms (M4 COR-1).
		placed, err := calc.SeriesPlacement(ctx, db, s.ID, calc.FarmFilter{All: all, IDs: ids})
		if err != nil {
			return PageResponse{}, err
		}
		if placed.F+placed.M > 0 {
			res.Placement.Placed = &httpx.SexPair[int]{F: placed.F, M: placed.M}
		}
	}
	if slices.ContainsFunc(mine, func(c dbq.SeriesCyclesRow) bool {
		return cycleType(c, farmsByID[c.FarmID].Type) == dbq.FarmTypeLaying
	}) {
		res.Tabs = append(res.Tabs, farms.TabProductie)
	}

	var alive httpx.SexPair[int]
	for _, f := range farmRows {
		if !f.Visible {
			res.OwnFarmsOnly = true
			res.Farms = append(res.Farms, HiddenFarmPanel{Farm: HiddenFarm{Name: f.Name}})
			continue
		}
		cycles := cyclesByFarm[f.ID]
		res.Track.Lanes = append(res.Track.Lanes, lane(f, cycles, transfers, s.Name, age, in.Today))
		p, err := panel(ctx, db, in, f, cycles, age)
		if err != nil {
			return PageResponse{}, err
		}
		res.Farms = append(res.Farms, p)
		if active && p.Cycle != nil && p.Cycle.Open {
			day, err := calc.ReadFarmDay(ctx, db, f.ID, in.Today)
			if err != nil {
				return PageResponse{}, err
			}
			alive.F += day.AliveEnd.F
			alive.M += day.AliveEnd.M
		}
	}
	if active {
		res.Alive = &alive
	}
	return res, nil
}

// panel is the farm panel of f, a farm the caller sees, with cycles its
// cycles of the series (oldest first).
func panel(ctx context.Context, db dbq.DBTX, in Input, f dbq.SeriesFarmsRow, cycles []dbq.SeriesCyclesRow, age calc.SeriesAge) (FarmPanel, error) {
	p := FarmPanel{
		Farm:   PanelFarm{ID: f.ID, Name: f.Name, Type: auth.FarmType(f.Type)},
		Access: true,
		Stats:  []farms.Stat{},
	}
	c, ok := shownCycle(cycles)
	if !ok {
		return p, nil
	}
	p.Cycle = &PanelCycle{
		ID: c.ID, StartDate: c.StartDate, EndDate: c.EndDate, Open: c.EndDate == nil,
		StartInferred: c.StartInferred, SaptSpan: saptSpan(age, c, in.Today),
	}
	if c.GuideID != nil && c.GuideName != nil {
		p.Guide = &farms.GuideRef{ID: *c.GuideID, Name: *c.GuideName}
	}
	stats, err := farms.BuildCycleStats(ctx, db, farms.Input{
		Principal: in.Principal, FarmID: f.ID, CycleID: &c.ID, Today: in.Today, Now: in.Now,
	})
	if err != nil {
		return FarmPanel{}, fmt.Errorf("series: panel %s: %w", f.Name, err)
	}
	p.Stats = stats
	return p, nil
}

// lane is f's track lane: its cycles of the series, oldest first, and the
// links that leave f, by date.
func lane(f dbq.SeriesFarmsRow, cycles []dbq.SeriesCyclesRow, transfers []dbq.SeriesTransfersRow, seriesName string, age calc.SeriesAge, today timex.Day) Lane {
	l := Lane{
		Farm:      farms.FarmRef{ID: f.ID, Name: f.Name},
		Cycles:    make([]LaneCycle, 0, len(cycles)),
		Transfers: []farms.TrackTransfer{},
	}
	for _, c := range cycles {
		l.Cycles = append(l.Cycles, LaneCycle{
			TrackCycle: farms.TrackCycle{
				ID: c.ID, Series: seriesName, StartDate: c.StartDate, EndDate: c.EndDate,
				Open: c.EndDate == nil, StartInferred: c.StartInferred, SaptSpan: saptSpan(age, c, today),
			},
			Type: auth.FarmType(cycleType(c, f.Type)),
		})
	}
	for _, t := range transfers {
		if t.FromFarmID == f.ID {
			l.Transfers = append(l.Transfers, farms.TrackTransfer{
				Date: t.Date, FromFarm: t.FromFarm, ToFarm: t.ToFarm, F: int(t.F), M: int(t.M),
			})
		}
	}
	return l
}

// shownCycle is the farm's cycle of the series shown on its panel: the open
// one, else the latest. cycles are oldest first.
func shownCycle(cycles []dbq.SeriesCyclesRow) (dbq.SeriesCyclesRow, bool) {
	if i := slices.IndexFunc(cycles, func(c dbq.SeriesCyclesRow) bool { return c.EndDate == nil }); i >= 0 {
		return cycles[i], true
	}
	if n := len(cycles); n > 0 {
		return cycles[n-1], true
	}
	return dbq.SeriesCyclesRow{}, false
}

// cycleType is the cycle's type: its guide's, else the farm's (as
// calc.CycleInfo.Type).
func cycleType(c dbq.SeriesCyclesRow, farm dbq.FarmType) dbq.FarmType {
	if c.GuideType != nil {
		return *c.GuideType
	}
	return farm
}

// lastDay is the cycle's last shown day: its end, or today while it is open
// (or when it ends after today).
func lastDay(c dbq.SeriesCyclesRow, today timex.Day) timex.Day {
	if c.EndDate != nil && c.EndDate.Before(today) {
		return *c.EndDate
	}
	return today
}

// saptSpan is the cycle's weeks up to its last shown day.
func saptSpan(age calc.SeriesAge, c dbq.SeriesCyclesRow, today timex.Day) [2]int {
	return [2]int{calc.SAPT(age, c.StartDate), calc.SAPT(age, lastDay(c, today))}
}

// saptSpanOf is the weeks of cycles (non-empty, oldest first): from the
// first start to the latest last shown day.
func saptSpanOf(age calc.SeriesAge, cycles []dbq.SeriesCyclesRow, today timex.Day) [2]int {
	last := lastDay(cycles[0], today)
	for _, c := range cycles[1:] {
		if d := lastDay(c, today); d.After(last) {
			last = d
		}
	}
	return [2]int{calc.SAPT(age, cycles[0].StartDate), calc.SAPT(age, last)}
}

// endOf is a closed series' end over cycles (non-empty): the latest end and
// how that cycle ended; nil while one of them is open. A cycle closed by the
// history import (ended_by migration) ended by emptying, as the workbook's
// flocks did (S3 "încheiată prin golire 28.01.2026").
func endOf(cycles []dbq.SeriesCyclesRow) *End {
	var end *End
	for _, c := range cycles {
		if c.EndDate == nil || c.EndedBy == nil {
			return nil
		}
		if end != nil && !c.EndDate.After(end.Date) {
			continue
		}
		by := EndEmptied
		if *c.EndedBy == dbq.CycleEndKindManual {
			by = EndManual
		}
		end = &End{Date: *c.EndDate, By: by}
	}
	return end
}
