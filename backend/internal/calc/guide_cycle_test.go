package calc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

func TestGuideWeeksLookups(t *testing.T) {
	g := GuideWeeks{
		58: {
			F: &GuideCells{LayPct: mustDecp("58.8"), FeedG: mustDecp("157"), BodyWeightG: mustDecp("3900")},
			M: &GuideCells{FeedG: mustDecp("156"), BodyWeightG: mustDecp("4800")},
		},
		59: {F: &GuideCells{LayPct: mustDecp("57.6")}},
	}
	str := func(d *decimal.Decimal) string {
		if d == nil {
			return "<nil>"
		}
		return d.String()
	}
	for _, tc := range []struct {
		name string
		got  *decimal.Decimal
		want string
	}{
		{"lay 58", g.LayPct(58), "58.8"},
		{"lay 59", g.LayPct(59), "57.6"},
		{"lay, no week", g.LayPct(60), "<nil>"},
		{"feed ♀", g.FeedG(58, dbq.SexF), "157"},
		{"feed ♂", g.FeedG(58, dbq.SexM), "156"},
		{"feed, cell without a value", g.FeedG(59, dbq.SexF), "<nil>"},
		{"feed, no ♂ row", g.FeedG(59, dbq.SexM), "<nil>"},
		{"weight ♀", g.BodyWeightG(58, dbq.SexF), "3900"},
		{"weight ♂", g.BodyWeightG(58, dbq.SexM), "4800"},
		{"weight, no week", g.BodyWeightG(57, dbq.SexF), "<nil>"},
		{"no guide", GuideWeeks(nil).LayPct(58), "<nil>"},
		{"no guide feed", GuideWeeks(nil).FeedG(58, dbq.SexM), "<nil>"},
	} {
		if got := str(tc.got); got != tc.want {
			t.Errorf("%s = %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestCycleTypeAndDefaultCycle(t *testing.T) {
	end := day(2026, time.January, 28)
	open := CycleInfo{Cycle: Cycle{ID: uuid.New()}, Guide: &CycleGuide{Type: dbq.FarmTypeLaying}}
	closed := CycleInfo{Cycle: Cycle{ID: uuid.New(), EndDate: &end}}

	if got := open.Type(dbq.FarmTypeGrowth); got != dbq.FarmTypeLaying {
		t.Errorf("type with a guide = %s, want the guide's", got)
	}
	if got := closed.Type(dbq.FarmTypeGrowth); got != dbq.FarmTypeGrowth {
		t.Errorf("type without a guide = %s, want the farm's", got)
	}

	older := CycleInfo{Cycle: Cycle{ID: uuid.New(), EndDate: &end}}
	for _, tc := range []struct {
		name   string
		cycles []CycleInfo
		want   uuid.UUID
		ok     bool
	}{
		{"the open one", []CycleInfo{closed, open}, open.ID, true},
		{"else the latest", []CycleInfo{closed, older}, closed.ID, true},
		{"none", nil, uuid.Nil, false},
	} {
		got, ok := DefaultCycle(tc.cycles)
		if got.ID != tc.want || ok != tc.ok {
			t.Errorf("%s: DefaultCycle = %s, %v; want %s, %v", tc.name, got.ID, ok, tc.want, tc.ok)
		}
	}
}
