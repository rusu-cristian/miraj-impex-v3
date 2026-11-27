package farms_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded. It lives in the
// external package because seed imports farms (through migratehistory and
// golden), so an in-package test can't import it.
func TestMain(m *testing.M) { seed.Main(m) }

// seria2025 is Hîrbovăț and its Seria 2025 cycle (no flock count: cycle
// 07.03.2025–28.01.2026, SAPT 19–65, register from 04.04.2025, SAPT 23).
func seria2025(t *testing.T, db dbq.DBTX) (farm uuid.UUID, cycle calc.CycleInfo) {
	t.Helper()
	if err := db.QueryRow(t.Context(), `select id from farm where name_key = 'hirbovat'`).Scan(&farm); err != nil {
		t.Fatal(err)
	}
	cycles, err := calc.FarmCycles(t.Context(), db, farm)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cycles {
		if c.SeriesName == "Seria 2025" {
			return farm, c
		}
	}
	t.Fatal("Hîrbovăț has no Seria 2025 cycle")
	return farm, cycle
}

// TestBuildSummarySeria2025: a past cycle without a flock count (COR-1).
// Its period is resolved on the register start, as the table resolves it
// (SAPT 23–65), and its alive and cumulative mortality figures are null,
// like the page's alive (M2 D5).
func TestBuildSummarySeria2025(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm, cycle := seria2025(t, d.App)
	in := func(p calc.Period) farms.Input {
		return farms.Input{FarmID: farm, CycleID: &cycle.ID, Period: p, Today: testdb.Today, Now: testdb.Now()}
	}
	summary := func(p calc.Period) (farms.SummaryResponse, string) {
		t.Helper()
		res, err := farms.BuildSummary(ctx, d.App, in(p))
		if err != nil {
			t.Fatal(err)
		}
		b, err := json.Marshal(res)
		if err != nil {
			t.Fatal(err)
		}
		return res, string(b)
	}
	table := func(p calc.Period) periodtable.TableResponse {
		t.Helper()
		res, err := periodtable.Build(ctx, d.App, periodtable.Input{
			Query: periodtable.TableQuery{View: "farm", FarmID: &farm, CycleID: &cycle.ID, Period: periodtable.Period(p)},
			Today: testdb.Today, Now: testdb.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}

	// The whole cycle, taken on its last day (28.01.2026).
	res, got := summary(calc.Period{})
	for _, want := range []string{
		`"readout":null`, // no hen-days, so no lay % to show
		`{"id":"alive","F":null,"M":null,`,
		`{"id":"cumMortality","F":null,"M":null,`,
		`"tiles":[]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("summary lacks %s:\n%s", want, got)
		}
	}
	if len(res.Stats) != 5 {
		t.Errorf("stats = %d, want alive, cumMortality, weight, feedToday, layToday", len(res.Stats))
	}

	// The readout series covers the table's weeks: SAPT 23–65, not the
	// cycle's 19–65.
	series, err := periodtable.WeekSeries(ctx, d.App, farm, cycle, calc.Period{}, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	if tp := table(calc.Period{}).Period; tp.SaptFrom != 23 || tp.SaptTo != 65 || tp.WeeksTotal != 43 {
		t.Errorf("table period = %+v, want SAPT 23–65, 43 weeks", tp)
	}
	if len(series) != 43 || series[0].Sapt != 23 || series[42].Sapt != 65 {
		t.Errorf("series = %d weeks from %v, want SAPT 23–65", len(series), series)
	}

	// Before the register's first day the period has no day in the cycle as
	// the table counts it: both are empty. From 04.04.2025 neither is.
	day := func(m time.Month, d int) timex.Day { return timex.MustDay(2025, m, d) }
	const empty = `{"readout":null,"stats":[],"tiles":[],"alerts":[],"latest":[]}`
	before := calc.Period{Kind: calc.PeriodInterval, From: day(time.March, 7), To: day(time.April, 3)}
	if _, got := summary(before); got != empty {
		t.Errorf("summary before the register = %s, want %s", got, empty)
	}
	if tb := table(before); tb.Empty == nil {
		t.Errorf("table before the register is not empty: %+v", tb.Period)
	}
	first := calc.Period{Kind: calc.PeriodInterval, From: day(time.March, 7), To: day(time.April, 4)}
	if res, _ := summary(first); len(res.Stats) == 0 {
		t.Error("summary on the register's first day is empty")
	}
	if tf := table(first); tf.Empty != nil || tf.Period.SaptFrom != 23 {
		t.Errorf("table on the register's first day = %+v", tf.Period)
	}
}
