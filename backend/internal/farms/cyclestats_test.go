package farms

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
)

// TestBuildCycleStats: the series page's panel figures (M4 D7) are the
// summary's stats for the same input, for an open laying cycle, a closed
// growth one and a past cycle without a flock count, with a period too; a
// farm without a cycle has none, and the summary's errors stay.
func TestBuildCycleStats(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	asJSON := func(v any) string {
		t.Helper()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	weeks4 := calc.Period{Kind: calc.PeriodWeeks, Weeks: 4}
	for _, tc := range []struct {
		name   string
		farm   uuid.UUID
		cycle  *uuid.UUID
		period calc.Period
	}{
		{"open laying cycle", f.hirbovat, nil, calc.Period{}},
		{"open laying cycle, 4 weeks", f.hirbovat, &f.cycleH26, weeks4},
		{"closed growth cycle", f.telita, &f.cycT, calc.Period{}},
		{"cycle without a flock count", f.hirbovat, &f.cycleH25, calc.Period{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := f.input(owner(), tc.farm, tc.cycle)
			in.Period = tc.period
			summary, err := BuildSummary(ctx, f.d.App, in)
			if err != nil {
				t.Fatal(err)
			}
			stats, err := BuildCycleStats(ctx, f.d.App, in)
			if err != nil {
				t.Fatal(err)
			}
			if len(stats) == 0 {
				t.Fatal("no stats")
			}
			if got, want := asJSON(stats), asJSON(summary.Stats); got != want {
				t.Errorf("stats = %s\nwant the summary's %s", got, want)
			}
		})
	}

	stats, err := BuildCycleStats(ctx, f.d.App, f.input(owner(), f.noua, nil))
	if err != nil || stats == nil || len(stats) != 0 {
		t.Errorf("a farm without a cycle: stats %v, err %v; want [] and no error", stats, err)
	}
	empty := f.input(owner(), f.telita, &f.cycT)
	empty.Period = calc.Period{Kind: calc.PeriodInterval, From: day(2025, time.September, 1), To: day(2025, time.September, 30)}
	if stats, err := BuildCycleStats(ctx, f.d.App, empty); err != nil || asJSON(stats) != "[]" {
		t.Errorf("a period with no day in the cycle: stats %v, err %v; want []", stats, err)
	}
	if _, err := BuildCycleStats(ctx, f.d.App, f.input(owner(), uuid.New(), nil)); !errors.Is(err, ErrFarmNotFound) {
		t.Errorf("unknown farm: err = %v, want ErrFarmNotFound", err)
	}
	if _, err := BuildCycleStats(ctx, f.d.App, f.input(owner(), f.hirbovat, &f.cycT)); !errors.Is(err, ErrCycleNotFound) {
		t.Errorf("another farm's cycle: err = %v, want ErrCycleNotFound", err)
	}
}
