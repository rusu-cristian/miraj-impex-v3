package farms

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
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

var (
	// ErrFarmNotFound is returned for a farm id that doesn't exist.
	ErrFarmNotFound = errors.New("farms: farm not found")
	// ErrCycleNotFound is returned for a cycle id that is not a cycle of the
	// farm (another farm's cycle included).
	ErrCycleNotFound = errors.New("farms: cycle not found at the farm")
)

// Input is what the farm builders take (ARC-6): the caller, the farm, the
// asked cycle (nil = the default), the period (summary only) and the business
// day and instant, explicit so that a builder never reads app_today().
type Input struct {
	Principal auth.Principal
	FarmID    uuid.UUID
	CycleID   *uuid.UUID
	Period    calc.Period
	Today     timex.Day
	Now       time.Time
}

// BuildPage builds GET /api/farms/{id}/page (§7.3, M2 D5, Q4) on db. It does
// not check the caller's access to the farm (the service does, M2 D4); the
// principal only decides the actions. ErrFarmNotFound and ErrCycleNotFound
// name a missing farm or cycle.
func BuildPage(ctx context.Context, db dbq.DBTX, in Input) (PageResponse, error) {
	q := dbq.New(db)
	farm, err := q.FarmsGetFarm(ctx, in.FarmID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PageResponse{}, ErrFarmNotFound
	}
	if err != nil {
		return PageResponse{}, fmt.Errorf("farms: page farm: %w", err)
	}
	cycles, err := calc.FarmCycles(ctx, db, in.FarmID)
	if err != nil {
		return PageResponse{}, err
	}
	shown, err := pickCycle(cycles, in.CycleID)
	if err != nil {
		return PageResponse{}, err
	}
	transfers, err := q.FarmsTransfers(ctx, in.FarmID)
	if err != nil {
		return PageResponse{}, fmt.Errorf("farms: page transfers: %w", err)
	}

	res := PageResponse{
		Farm:   PageFarm{ID: farm.ID, Name: farm.Name, Type: auth.FarmType(farm.Type), Guide: guideRef(farm.GuideID, farm.GuideName)},
		Cycles: make([]CycleItem, 0, len(cycles)),
		Track:  track(farm.ID, farm.Name, cycles, transfers, in.Today),
		Tabs:   []string{},
		AsOf:   timex.At(in.Now),
	}
	for _, c := range cycles {
		from, to := saptSpan(c, in.Today)
		res.Cycles = append(res.Cycles, CycleItem{
			ID: c.ID, Series: c.SeriesName, Label: messages.CycleLabel(c.StartDate, c.EndDate, from, to),
			Open: c.EndDate == nil, StartInferred: c.StartInferred,
		})
	}
	res.Empty = emptyState(cycles)

	// Today's count decides Vânzare; only a manager or owner can get it.
	canManage := in.Principal.Role.In(auth.Manager, auth.Owner)
	var aliveToday *calc.PerSex[int]
	if shown != nil {
		last := lastDay(shown.Cycle, in.Today)
		res.Cycle = pageCycle(*shown, last)
		res.Farm.Guide = nil // the cycle's snapshot wins, "fără ghid" included
		if shown.Guide != nil {
			res.Farm.Guide = &GuideRef{ID: shown.Guide.ID, Name: shown.Guide.Name}
		}
		res.Tabs = tabsFor(shown.Type(farm.Type))
		if shown.HasBirdEvents {
			day, err := calc.ReadFarmDay(ctx, db, in.FarmID, last)
			if err != nil {
				return PageResponse{}, err
			}
			res.Alive = &httpx.SexPair[int]{F: day.AliveEnd.F, M: day.AliveEnd.M}
			if last == in.Today {
				aliveToday = &day.AliveEnd
			}
		}
	}
	if canManage && aliveToday == nil {
		day, err := calc.ReadFarmDay(ctx, db, in.FarmID, in.Today)
		if err != nil {
			return PageResponse{}, err
		}
		aliveToday = &day.AliveEnd
	}
	birdsToday := aliveToday != nil && aliveToday.F+aliveToday.M > 0
	res.Actions = actionsFor(in.Principal.Role, birdsToday, res.Empty == nil)
	return res, nil
}

// pickCycle is the asked cycle (ErrCycleNotFound when it isn't one of
// cycles), else the open one, else the latest; nil when there is none.
// cycles are newest first.
func pickCycle(cycles []calc.CycleInfo, id *uuid.UUID) (*calc.CycleInfo, error) {
	if id != nil {
		i := slices.IndexFunc(cycles, func(c calc.CycleInfo) bool { return c.ID == *id })
		if i < 0 {
			return nil, ErrCycleNotFound
		}
		return &cycles[i], nil
	}
	if c, ok := calc.DefaultCycle(cycles); ok {
		return &c, nil
	}
	return nil, nil
}

// lastDay is the cycle's last shown day: its end, or today while it is open
// (or when it ends after today).
func lastDay(c calc.Cycle, today timex.Day) timex.Day {
	if c.EndDate != nil && c.EndDate.Before(today) {
		return *c.EndDate
	}
	return today
}

// saptSpan is the cycle's weeks up to its last shown day.
func saptSpan(c calc.CycleInfo, today timex.Day) (int, int) {
	return calc.SAPT(c.Age, c.StartDate), calc.SAPT(c.Age, lastDay(c.Cycle, today))
}

func pageCycle(c calc.CycleInfo, last timex.Day) *PageCycle {
	start, now := calc.SAPT(c.Age, c.StartDate), calc.SAPT(c.Age, last)
	return &PageCycle{
		ID: c.ID, Series: c.SeriesName, StartDate: c.StartDate, EndDate: c.EndDate,
		SaptStart: start, SaptToday: now, AgeDaysToday: calc.AgeDays(c.Age, last), Weeks: now - start + 1,
	}
}

// tabsFor lists the tabs of a cycle of type t: Producție only for laying (§7.3).
func tabsFor(t dbq.FarmType) []string {
	tabs := []string{TabRezumat, TabPasari, TabHrana, TabGreutate}
	if t == dbq.FarmTypeLaying {
		tabs = append(tabs, TabProductie)
	}
	return tabs
}

// actionsFor applies M2 D5: a manager or owner may move birds and start a
// cycle, sell while the farm has birds today and change the guide while a
// cycle is open; anyone else may do none of these.
func actionsFor(role auth.Role, birdsToday, openCycle bool) Actions {
	if !role.In(auth.Manager, auth.Owner) {
		return Actions{}
	}
	return Actions{Move: true, Sell: birdsToday, NewCycle: true, ChangeGuide: openCycle}
}

// emptyState is the empty-farm state (M2 Q4 b, c): nil while a cycle is
// open; "goală din" the latest cycle's end otherwise; the no-cycle sentence
// for a farm that never had one.
func emptyState(cycles []calc.CycleInfo) *EmptyState {
	if len(cycles) == 0 {
		return &EmptyState{Label: messages.FarmNoCycle(), Hint: messages.CycleStartsHint()}
	}
	if slices.ContainsFunc(cycles, func(c calc.CycleInfo) bool { return c.EndDate == nil }) {
		return nil
	}
	latest := cycles[0]
	since := *latest.EndDate
	from, to := calc.SAPT(latest.Age, latest.StartDate), calc.SAPT(latest.Age, since)
	return &EmptyState{
		Since: &since,
		Label: messages.FarmEmpty(since, latest.SeriesName, from, to),
		Hint:  messages.CycleStartsHint(),
	}
}

// track is the farm's lane (M2 Q4 a): its cycles oldest first and its
// transfer links.
func track(farmID uuid.UUID, farmName string, cycles []calc.CycleInfo, transfers []dbq.FarmsTransfersRow, today timex.Day) Track {
	lane := TrackLane{
		Farm:      FarmRef{ID: farmID, Name: farmName},
		Cycles:    make([]TrackCycle, 0, len(cycles)),
		Transfers: make([]TrackTransfer, 0, len(transfers)),
	}
	for _, c := range slices.Backward(cycles) {
		from, to := saptSpan(c, today)
		lane.Cycles = append(lane.Cycles, TrackCycle{
			ID: c.ID, Series: c.SeriesName, StartDate: c.StartDate, EndDate: c.EndDate,
			Open: c.EndDate == nil, StartInferred: c.StartInferred, SaptSpan: [2]int{from, to},
		})
	}
	for _, t := range transfers {
		lane.Transfers = append(lane.Transfers, TrackTransfer{
			Date: t.Date, FromFarm: t.FromFarm, ToFarm: t.ToFarm, F: int(t.F), M: int(t.M),
		})
	}
	return Track{Lanes: []TrackLane{lane}, Today: today}
}

func guideRef(id *uuid.UUID, name *string) *GuideRef {
	if id == nil || name == nil {
		return nil
	}
	return &GuideRef{ID: *id, Name: *name}
}
