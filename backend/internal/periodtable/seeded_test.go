package periodtable_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded. It lives in the
// external package because seed imports periodtable (through
// migratehistory and golden), so an in-package test can't import it.
func TestMain(m *testing.M) { seed.Main(m) }

// TestReadDaysAliveMatchesCalc (ARC-7): the day query's running count and
// calc's farm-day reader are two readings of §5.3; on every day of the
// seeded Hîrbovăț Seria 2026 cycle (22.02–18.11.2026) they give the same
// birds at the start and at the end of the day.
func TestReadDaysAliveMatchesCalc(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	var farm uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from farm where name_key = 'hirbovat'`).Scan(&farm); err != nil {
		t.Fatal(err)
	}
	cycle, err := calc.CycleAt(ctx, d.App, farm, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	days, err := periodtable.ReadDays(ctx, d.App, periodtable.DaysQuery{
		FarmID: farm, Scope: periodtable.ScopeFarm, From: cycle.StartDate, To: testdb.Today,
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := testdb.Today.DaysSince(cycle.StartDate) + 1; len(days.Rows) != want || want < 2 {
		t.Fatalf("day query: %d rows, want %d (one per day of the cycle)", len(days.Rows), want)
	}
	mismatches := 0
	for _, r := range days.Rows {
		fd, err := calc.ReadFarmDay(ctx, d.App, farm, r.Date)
		if err != nil {
			t.Fatal(err)
		}
		if fd.AliveEnd != r.AliveEnd || fd.AliveStart != r.AliveStart {
			mismatches++
			t.Errorf("%s: calc start %+v end %+v, day query start %+v end %+v",
				r.Date, fd.AliveStart, fd.AliveEnd, r.AliveStart, r.AliveEnd)
		}
		if mismatches > 5 {
			t.Fatal("too many mismatches")
		}
	}
	if last := days.Rows[len(days.Rows)-1].AliveEnd; last.F != 12561 || last.M != 794 {
		t.Errorf("alive at the end of 18.11.2026 = %+v, want ♀ 12 561 · ♂ 794", last)
	}
}
