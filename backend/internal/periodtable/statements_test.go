package periodtable

import (
	"context"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// statements is a pgx QueryTracer that records every statement a pool runs,
// by its "-- name: X" header ("?" without one); connection setup (SET ROLE,
// app.today) runs before the tracer sees queries of the test.
type statements struct {
	mu    sync.Mutex
	names []string
}

func (s *statements) TraceQueryStart(ctx context.Context, _ *pgx.Conn, d pgx.TraceQueryStartData) context.Context {
	name := "?"
	if rest, ok := strings.CutPrefix(d.SQL, "-- name: "); ok {
		name = strings.Fields(rest)[0]
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
func tracedPool(t *testing.T, f *farmT) (*pgxpool.Pool, *statements) {
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

// TestBuildStatements: a page is the cycle read and one day query (the
// record types are a CTE of it), plus the guide and threshold reads when the
// cycle has a guide (§7.4 Performance, T3 criterion 3).
func TestBuildStatements(t *testing.T) {
	f := newFarmT(t)
	pool, st := tracedPool(t, f)
	run := func(t *testing.T, q TableQuery, want ...string) {
		t.Helper()
		q.View, q.FarmID = "farm", &f.farm
		if _, err := Build(context.Background(), pool, Input{Query: q, Today: testdb.Today, Now: testdb.Now()}); err != nil {
			t.Fatal(err)
		}
		if got := st.take(); !slices.Equal(got, want) {
			t.Errorf("statements %q, want %q", got, want)
		}
	}
	t.Run("page 1", func(t *testing.T) {
		run(t, TableQuery{}, "CalcFarmCycles", "PeriodTableDays", "CalcGuideWeeks", "CalcSettings")
	})
	t.Run("load more with cycleId", func(t *testing.T) {
		run(t, TableQuery{CycleID: &f.cycle, Before: intp(24)}, "CalcCycleByID", "PeriodTableDays", "CalcGuideWeeks", "CalcSettings")
	})
	t.Run("no guide", func(t *testing.T) {
		f.exec(`update cycle set guide_id = null where id = $1`, f.cycle)
		run(t, TableQuery{}, "CalcFarmCycles", "PeriodTableDays")
	})
	t.Run("empty period", func(t *testing.T) {
		run(t, TableQuery{Period: Period{Kind: "interval", From: testdb.Today.AddDays(-400), To: testdb.Today.AddDays(-390)}}, "CalcFarmCycles")
	})
}

// TestReadDaysHallScope: the hall scope gives one row per egg unit and day,
// and its sums add up to the farm's.
func TestReadDaysHallScope(t *testing.T) {
	f := newFarmT(t)
	ctx := context.Background()
	q := DaysQuery{FarmID: f.farm, Scope: ScopeHall, From: day(time.November, 17), To: day(time.November, 18), WithChips: true}
	halls, err := ReadDays(ctx, f.d.App, q)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, r := range halls.Rows {
		got = append(got, r.Code+" "+r.Date.String()+" eggs "+itoa(r.Eggs)+" hen-days "+itoa(r.HenDays)+" hens end "+itoa(r.AliveEnd.F))
	}
	diff(t, "hall rows", []string{
		"H1 2026-11-17 eggs 156 hen-days 198 hens end 196",
		"H1 2026-11-18 eggs 156 hen-days 196 hens end 196",
		// H2 is emptied on 18.11 with no egg entry: its 50 hens count (M2 Q1).
		"H2 2026-11-17 eggs 32 hen-days 50 hens end 50",
		"H2 2026-11-18 eggs 0 hen-days 50 hens end 0",
	}, got)
	if halls.FarmType != "laying" || len(halls.Chips) == 0 {
		t.Errorf("farm type %q, chips %v", halls.FarmType, halls.Chips)
	}

	q.Scope, q.WithChips = ScopeFarm, false
	farm, err := ReadDays(ctx, f.d.App, q)
	if err != nil {
		t.Fatal(err)
	}
	if len(farm.Rows) != 2 || farm.Rows[1].HenDays != 246 || farm.Rows[1].Eggs != 156 || farm.Rows[1].FedBirdDays.F != 196 {
		t.Errorf("farm 18.11 = %+v", farm.Rows)
	}
	// Without WithChips the chips CTE isn't run: [] and the same rows.
	if farm.Chips == nil || len(farm.Chips) != 0 || farm.FarmType != "laying" {
		t.Errorf("farm without chips: type %q, chips %v", farm.FarmType, farm.Chips)
	}

	if _, err := ReadDays(ctx, f.d.App, DaysQuery{FarmID: f.farm, Scope: "section; drop table farm", From: q.From, To: q.To}); err == nil {
		t.Error("a scope outside the whitelist was accepted")
	}
	unknown, err := ReadDays(ctx, f.d.App, DaysQuery{FarmID: f.cycle, Scope: ScopeFarm, From: q.From, To: q.To})
	if err != nil || len(unknown.Rows) != 0 || unknown.Chips == nil {
		t.Errorf("unknown farm = %+v, %v", unknown, err)
	}
}
