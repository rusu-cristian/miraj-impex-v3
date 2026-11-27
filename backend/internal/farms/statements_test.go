package farms

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// statements is a pgx QueryTracer that records every statement a pool runs,
// by its "-- name: X" header ("?" without one), like periodtable's; a day
// query that carries the chips CTE is "PeriodTableDays+chips".
type statements struct {
	mu    sync.Mutex
	names []string
}

func (s *statements) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := "?"
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
	}
	if strings.Contains(d.SQL, "c(ord, chip, typ)") {
		name += "+chips"
	}
	s.mu.Lock()
	s.names = append(s.names, name)
	s.mu.Unlock()
	return ctx
}

func (s *statements) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (s *statements) take() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.names
	s.names = nil
	return out
}

// tracedPool is a second il_app pool on f's database whose statements are recorded.
func tracedPool(t *testing.T, f *fixture) (*pgxpool.Pool, *statements) {
	t.Helper()
	var st statements
	pc, err := db.PoolConfig(testdb.DSN(f.d.Name), 1, &testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	prepare := pc.AfterConnect
	pc.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if _, err := conn.Exec(ctx, "set role "+testdb.RoleApp); err != nil {
			return err
		}
		return prepare(ctx, conn)
	}
	pc.ConnConfig.Tracer = &st
	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(context.Background()); err != nil { // connect before counting
		t.Fatal(err)
	}
	st.take()
	return pool, &st
}

// TestBuildSummaryStatements: the summary reads each figure once (ARC-3):
// one guide read for the weighings and the readout series, one farm day
// query whose last row is the stats' day, one hall day query, one removals
// read for the deaths and the sales, and no chips CTE (Build's only).
func TestBuildSummaryStatements(t *testing.T) {
	f := newFixture(t)
	f.summaryData()
	pool, st := tracedPool(t, f)
	run := func(t *testing.T, in Input, want ...string) {
		t.Helper()
		if _, err := BuildSummary(t.Context(), pool, in); err != nil {
			t.Fatal(err)
		}
		if got := st.take(); !slices.Equal(got, want) {
			t.Errorf("statements %q,\nwant %q", got, want)
		}
	}
	reads := []string{
		"CalcSettings", "CalcGuideWeeks", "PeriodTableDays", "PeriodTableDays",
		"CalcFarmBase", "CalcFarmRemovals", "CalcLatestWeighings",
		"CalcSectionFeedDay", "CalcSectionRemovals", "CalcEmptySince",
	}
	head := []string{"FarmsGetFarm", "CalcFarmCycles"}
	t.Run("laying, whole cycle", func(t *testing.T) {
		run(t, f.input(owner(), f.hirbovat, nil), append(head, reads...)...)
	})
	t.Run("laying, weeks:4", func(t *testing.T) {
		in := f.input(owner(), f.hirbovat, nil)
		in.Period = calc.Period{Kind: calc.PeriodWeeks, Weeks: 4}
		run(t, in, append(head, reads...)...)
	})
	t.Run("growth", func(t *testing.T) {
		run(t, f.input(owner(), f.telita, nil), append(head, reads...)...)
	})
	t.Run("past cycle without a flock count", func(t *testing.T) {
		// Its first register day moves the period's start (M2 Q7); no guide.
		withoutGuide := slices.DeleteFunc(slices.Clone(reads), func(n string) bool { return n == "CalcGuideWeeks" })
		run(t, f.input(owner(), f.hirbovat, &f.cycleH25), append(append(head, "PeriodTableRegisterStart"), withoutGuide...)...)
	})
	t.Run("empty period", func(t *testing.T) {
		in := f.input(owner(), f.hirbovat, nil)
		in.Period = calc.Period{Kind: calc.PeriodInterval, From: day(2025, time.January, 1), To: day(2025, time.January, 31)}
		run(t, in, head...)
	})
}
