package golden_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// m1Rows are the goldens SPEC §12.6 puts in M1.
var m1Rows = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 26, 27}

// TestGolden runs every implemented row of §11.7 against the dev seed
// (§11.9) and skips the others with the milestone that implements them.
func TestGolden(t *testing.T) {
	env := golden.Env{DB: seededDB(t).App}
	for _, r := range golden.Rows() {
		t.Run(fmt.Sprintf("%02d %s", r.ID, r.Name), func(t *testing.T) {
			if r.Got == nil {
				t.Skipf("golden %d is implemented in %s", r.ID, r.Milestone)
			}
			err := r.Check(t.Context(), env)
			if mm, ok := errors.AsType[*golden.Mismatch](err); ok {
				t.Fatalf("golden %d %s (-want +got):\n%s", mm.ID, mm.Name, cmp.Diff(mm.Want, mm.Got))
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// seededDB is the §11.7 fixture: a copy of the template loaded by
// migratehistory.Load with the dev seed. That template arrives with
// testdb.NewSeeded (M1 T8); until then there is nothing to run the goldens
// against, and TestGoldenRowsOnHandBuiltLedger exercises the rows instead.
func seededDB(t *testing.T) *testdb.DB {
	t.Helper()
	return testdb.NewSeeded(t)
}

// checkRegistry lists what is wrong with a registry: ids not exactly 1–38
// once each, a row without a name or a valid milestone, an M1 row without Got,
// or a Got without a Want.
func checkRegistry(rows []golden.Row) []string {
	var problems []string
	seen := map[int]int{}
	milestone := regexp.MustCompile(`^M([1-9]|1[0-2])$`)
	for _, r := range rows {
		seen[r.ID]++
		if r.ID < 1 || r.ID > 38 {
			problems = append(problems, fmt.Sprintf("id %d is outside 1–38", r.ID))
		}
		if strings.TrimSpace(r.Name) == "" {
			problems = append(problems, fmt.Sprintf("golden %d has no name", r.ID))
		}
		if !milestone.MatchString(r.Milestone) {
			problems = append(problems, fmt.Sprintf("golden %d has milestone %q", r.ID, r.Milestone))
		}
		if r.Milestone == "M1" && r.Got == nil {
			problems = append(problems, fmt.Sprintf("golden %d is M1 but has no Got", r.ID))
		}
		if r.Got != nil && r.Want == nil {
			problems = append(problems, fmt.Sprintf("golden %d has Got but no Want", r.ID))
		}
	}
	for id := 1; id <= 38; id++ {
		switch n := seen[id]; {
		case n == 0:
			problems = append(problems, fmt.Sprintf("id %d is missing", id))
		case n > 1:
			problems = append(problems, fmt.Sprintf("id %d appears %d times", id, n))
		}
	}
	return problems
}

func TestRegistry(t *testing.T) {
	rows := golden.Rows()
	if problems := checkRegistry(rows); len(problems) > 0 {
		t.Fatalf("registry problems:\n%s", strings.Join(problems, "\n"))
	}
	var m1, implemented []int
	for i, r := range rows {
		if r.ID != i+1 {
			t.Errorf("row %d has id %d: rows are in id order", i, r.ID)
		}
		if r.Milestone == "M1" {
			m1 = append(m1, r.ID)
		}
		if r.Got != nil {
			implemented = append(implemented, r.ID)
		}
	}
	if diff := cmp.Diff(m1Rows, m1); diff != "" {
		t.Errorf("M1 rows differ from SPEC §12.6 (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m1Rows, implemented); diff != "" {
		t.Errorf("implemented rows at M1 (-want +got):\n%s", diff)
	}
}

// TestCheckRegistryCatchesProblems shows the meta-test fails for each broken
// registry it guards against.
func TestCheckRegistryCatchesProblems(t *testing.T) {
	without := func(id int) []golden.Row {
		return slices.DeleteFunc(golden.Rows(), func(r golden.Row) bool { return r.ID == id })
	}
	tests := []struct {
		name string
		rows []golden.Row
		want string
	}{
		{"missing id", without(5), "id 5 is missing"},
		{"duplicate id", append(golden.Rows(), golden.Rows()[2]), "id 3 appears 2 times"},
		{"id out of range", append(golden.Rows(), golden.Row{ID: 39, Name: "x", Milestone: "M2"}), "id 39 is outside 1–38"},
		{"M1 row without Got", func() []golden.Row {
			rows := golden.Rows()
			rows[11].Got = nil
			return rows
		}(), "golden 12 is M1 but has no Got"},
		{"bad milestone", func() []golden.Row {
			rows := golden.Rows()
			rows[20].Milestone = "M13"
			return rows
		}(), `golden 21 has milestone "M13"`},
		{"Got without Want", func() []golden.Row {
			rows := golden.Rows()
			rows[0].Want = nil
			return rows
		}(), "golden 1 has Got but no Want"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems := checkRegistry(tc.rows)
			if !slices.Contains(problems, tc.want) {
				t.Errorf("problems %q do not include %q", problems, tc.want)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	ctx := context.Background()
	errRead := errors.New("read failed")
	row := func(want, got any, err error) golden.Row {
		return golden.Row{ID: 99, Name: "synthetic", Milestone: "M1", Want: want,
			Got: func(context.Context, golden.Env) (any, error) { return got, err }}
	}

	t.Run("equal values pass", func(t *testing.T) {
		if err := row(calc.PerSex[int]{F: 1, M: 2}, calc.PerSex[int]{F: 1, M: 2}, nil).Check(ctx, golden.Env{}); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("a different value is a Mismatch naming id, name, got and want", func(t *testing.T) {
		err := row(golden.LayDay{Eggs: 1, HensStart: 2, Pct: "50.0"}, golden.LayDay{Eggs: 1, HensStart: 3, Pct: "33.3"}, nil).
			Check(ctx, golden.Env{})
		mm, ok := errors.AsType[*golden.Mismatch](err)
		if !ok {
			t.Fatalf("err = %v, want a *golden.Mismatch", err)
		}
		if mm.ID != 99 || mm.Name != "synthetic" {
			t.Errorf("mismatch id/name = %d %q", mm.ID, mm.Name)
		}
		want := "golden 99 synthetic: got {Eggs:1 HensStart:3 Pct:33.3} want {Eggs:1 HensStart:2 Pct:50.0}"
		if err.Error() != want {
			t.Errorf("message:\n got %s\nwant %s", err, want)
		}
	})
	t.Run("a value of another type is a Mismatch", func(t *testing.T) {
		if _, ok := errors.AsType[*golden.Mismatch](row(1, int64(1), nil).Check(ctx, golden.Env{})); !ok {
			t.Fatal("int vs int64 compared equal")
		}
	})
	t.Run("a read error is returned with the row", func(t *testing.T) {
		err := row(1, nil, errRead).Check(ctx, golden.Env{})
		if !errors.Is(err, errRead) || !strings.HasPrefix(err.Error(), "golden 99 synthetic: ") {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("a row without Got is not implemented", func(t *testing.T) {
		err := golden.Row{ID: 13, Name: "guide lay", Milestone: "M2"}.Check(ctx, golden.Env{})
		if !errors.Is(err, golden.ErrNotImplemented) || !strings.Contains(err.Error(), "M2") {
			t.Fatalf("err = %v", err)
		}
	})
}

// ledger is a hand-built miniature Hîrbovăț: the golden farm and Seria 2026
// with their real names and dates, and a few birds, eggs and feed rows whose
// figures are worked out by hand below. It proves every M1 row reads the
// right farm, cycle and days and rounds at its precision, before the full
// import exists. Seria 2026: placed 07.10.2025 at age 0; cycle from
// 22.02.2026, open.
//
//	22.02.2026 placement H5.1 ♀ 60 · ♂ 8, H5.2 ♀ 20                   base ♀ 80 · ♂ 8
//	01.03.2026 adjustment H5.1 ♀ +7 (not in the base); mortality H5.1 ♂ 1
//	01.06.2026 cull H5.2 ♀ 1
//	11.11.2026 sale H5.2 ♀ 1 and adjustment H5.2 ♀ +1 (no net change: a sale before the 12–18.11 range)
//	12.11.2026 sale H5.2 ♀ 5
//	17.11.2026 mortality H5.1 ♀ 1
//	18.11.2026 mortality H5.1 ♀ 2; sale H5.2 ♀ 10, H5.1 ♂ 2
//
//	          H5.1 ♀ · ♂   H5.2 ♀   farm ♀ · ♂
//	start 18.11  66 · 7       14       80 · 7
//	end 18.11    64 · 5        4       68 · 5      (H6.1 never used: 0 · 0)
//
//	cycle deaths ♀ 1 + 1 + 2 = 4 → 4 ÷ 80 = 5,00 %; ♂ 1 ÷ 8 = 12,50 %
//	eggs 18.11 H5: 1 + 2 + 3 + 35 = 41 ÷ 80 hens = 51,25 % → 51,3 (half away from zero)
//	feed 18.11: ♀ H5.1 160 g × 66 = 10,56 kg + H5.2 155 g × 14 = 2,17 kg = 12,73 → 12,7;
//	            ♂ H5.1 150 g × 7 = 1,05 → 1,1; total 13,78 → 13,8
//	sold 18.11 by section: H5.1 0 · 2, H5.2 10 · 0 (H6.1 sold nothing: not listed; the 12.11 sale is another day)
//	sold in the cycle ♀ 1 + 5 + 10 = 16 · ♂ 2; sold 12–18.11 ♀ 15 · ♂ 2
//	stock 18.11: Breeder 2 100 − 12,73 = 87,27 → 87,3; Cocoși 10 − 1,05 = 8,95 → 9,0
type ledger struct {
	t    *testing.T
	d    *testdb.DB
	farm uuid.UUID
	sec  map[string]uuid.UUID
}

func (l *ledger) id(sql string, args ...any) uuid.UUID {
	l.t.Helper()
	var id uuid.UUID
	if err := l.d.App.QueryRow(context.Background(), sql+" returning id", args...).Scan(&id); err != nil {
		l.t.Fatalf("%s: %v", sql, err)
	}
	return id
}

func (l *ledger) section(code string, parent string, sort int) {
	l.t.Helper()
	var p *uuid.UUID
	if parent != "" {
		id := l.sec[parent]
		p = &id
	}
	l.sec[code] = l.id(`insert into section (farm_id, code, parent_id, sort_order) values ($1, $2, $3, $4)`, l.farm, code, p, sort)
}

func (l *ledger) entry(kind string, d timex.Day, series *uuid.UUID) uuid.UUID {
	l.t.Helper()
	return l.id(`insert into entry (kind, farm_id, date, series_id, source) values ($1, $2, $3, $4, 'manual')`,
		kind, l.farm, d, series)
}

func (l *ledger) bird(entry uuid.UUID, code string, d timex.Day, sex, typ string, qty int) {
	l.t.Helper()
	var reason *string
	if typ == "cull" {
		r := "boala"
		reason = &r
	}
	l.id(`insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty, cull_reason)
		values ($1, $2, $3, $4, $5, $6, $7, $8)`, entry, l.farm, l.sec[code], d, sex, typ, qty, reason)
}

func day(y int, m time.Month, d int) timex.Day { return timex.MustDay(y, m, d) }

func newLedger(t *testing.T) *ledger {
	t.Helper()
	l := &ledger{t: t, d: testdb.New(t), sec: map[string]uuid.UUID{}}
	series := l.id(`insert into series (name, name_key, year, placement_date, age_at_placement_days)
		values ('Seria 2026', 'seria2026', 2026, $1, 0)`, day(2025, time.October, 7))
	l.farm = l.id(`insert into farm (name, name_key, type) values ('Hîrbovăț', 'hirbovat', 'laying')`)
	other := l.id(`insert into farm (name, name_key, type) values ('Telița', 'telita', 'growth')`)
	l.id(`insert into cycle (series_id, farm_id, start_date, started_by) values ($1, $2, $3, 'auto')`,
		series, l.farm, day(2026, time.February, 22))
	// A Telița cycle over the same days: the rows must pick Hîrbovăț's.
	l.id(`insert into cycle (series_id, farm_id, start_date, end_date, started_by, ended_by) values ($1, $2, $3, $4, 'auto', 'emptied')`,
		series, other, day(2025, time.October, 7), day(2026, time.February, 28))

	l.section("H5", "", 1)
	l.section("H6", "", 2)
	l.section("H5.2", "H5", 2)
	l.section("H5.1", "H5", 1)
	l.section("H6.1", "H6", 1)

	e := l.entry("placement", day(2026, time.February, 22), &series)
	l.bird(e, "H5.1", day(2026, time.February, 22), "F", "placement", 60)
	l.bird(e, "H5.1", day(2026, time.February, 22), "M", "placement", 8)
	l.bird(e, "H5.2", day(2026, time.February, 22), "F", "placement", 20)
	e = l.entry("migration", day(2026, time.March, 1), nil)
	l.bird(e, "H5.1", day(2026, time.March, 1), "F", "adjustment", 7)
	e = l.entry("daily", day(2026, time.March, 1), nil)
	l.bird(e, "H5.1", day(2026, time.March, 1), "M", "mortality", -1)
	e = l.entry("daily", day(2026, time.June, 1), nil)
	l.bird(e, "H5.2", day(2026, time.June, 1), "F", "cull", -1)
	e = l.entry("sale", day(2026, time.November, 11), nil)
	l.bird(e, "H5.2", day(2026, time.November, 11), "F", "sale", -1)
	e = l.entry("migration", day(2026, time.November, 11), nil)
	l.bird(e, "H5.2", day(2026, time.November, 11), "F", "adjustment", 1)
	e = l.entry("sale", day(2026, time.November, 12), nil)
	l.bird(e, "H5.2", day(2026, time.November, 12), "F", "sale", -5)
	e = l.entry("daily", day(2026, time.November, 17), nil)
	l.bird(e, "H5.1", day(2026, time.November, 17), "F", "mortality", -1)

	today := golden.AsOf
	e = l.entry("daily", today, nil)
	l.bird(e, "H5.1", today, "F", "mortality", -2)
	l.id(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs)
		values ($1, $2, $3, $4, 0, 1, 2, 3, 35)`, e, l.farm, l.sec["H5"], today)
	e = l.entry("sale", today, nil)
	l.bird(e, "H5.2", today, "F", "sale", -10)
	l.bird(e, "H5.1", today, "M", "sale", -2)

	breeder := l.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder2', 'F')`)
	cocosi := l.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M')`)
	e = l.entry("delivery", day(2026, time.November, 1), nil)
	delivery := `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`
	l.id(delivery, e, l.farm, breeder, day(2026, time.November, 1), 100)
	l.id(delivery, e, l.farm, cocosi, day(2026, time.November, 1), 10)
	e = l.entry("feed_daily", today, nil)
	alloc := `insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, $5, $6, $7)`
	l.id(alloc, e, l.farm, l.sec["H5.1"], today, "F", breeder, 160)
	l.id(alloc, e, l.farm, l.sec["H5.2"], today, "F", breeder, 155)
	l.id(alloc, e, l.farm, l.sec["H5.1"], today, "M", cocosi, 150)
	return l
}

// TestGoldenRowsOnHandBuiltLedger runs every implemented row against the
// miniature ledger. Rows 1–4 depend only on the series and the cycle, so they
// must already give the real golden values; the others must give the
// hand-worked figures, and Check must report them as mismatches.
func TestGoldenRowsOnHandBuiltLedger(t *testing.T) {
	env := golden.Env{DB: newLedger(t).d.App}
	ps := func(f, m int) calc.PerSex[int] { return calc.PerSex[int]{F: f, M: m} }
	want := map[int]any{
		5: ps(68, 5),
		6: ps(80, 7),
		7: []golden.SectionCount{{Code: "H5.1", Count: ps(64, 5)}, {Code: "H5.2", Count: ps(4, 0)}, {Code: "H6.1", Count: ps(0, 0)}},
		8: golden.DeathsSold{Deaths: ps(2, 0), Sold: ps(10, 2),
			SoldBySection: []golden.SectionCount{{Code: "H5.1", Count: ps(0, 2)}, {Code: "H5.2", Count: ps(10, 0)}}},
		9:  ps(80, 8),
		10: golden.CumMortality{Deaths: ps(4, 1), Pct: calc.PerSex[string]{F: "5.00", M: "12.50"}},
		11: golden.SoldSpan{Cycle: ps(16, 2), Range: ps(15, 2)},
		12: golden.LayDay{Eggs: 41, HensStart: 80, Pct: "51.3"},
		26: golden.FeedDay{Kg: calc.PerSex[string]{F: "12.7", M: "1.1"}, Total: "13.8"},
		27: []golden.StockKg{{Name: "Breeder 2", Kg: "87.3"}, {Name: "Cocoși", Kg: "9.0"}},
	}
	for _, r := range golden.Rows() {
		if r.Got == nil {
			continue
		}
		t.Run(fmt.Sprintf("%02d %s", r.ID, r.Name), func(t *testing.T) {
			ctx := t.Context()
			if r.ID <= 4 {
				if err := r.Check(ctx, env); err != nil {
					t.Fatal(err)
				}
				return
			}
			w, ok := want[r.ID]
			if !ok {
				t.Fatalf("golden %d has no hand-worked value in this test", r.ID)
			}
			got, err := r.Got(ctx, env)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(w, got); diff != "" {
				t.Errorf("(-want +got):\n%s", diff)
			}
			mm, ok := errors.AsType[*golden.Mismatch](r.Check(ctx, env))
			if !ok || mm.ID != r.ID || !cmp.Equal(mm.Got, got) {
				t.Errorf("Check on the miniature ledger = %v, want a mismatch carrying %+v", mm, got)
			}
		})
	}
}

// TestGoldenRowsWithoutTheFarm shows a database without the golden farm is an
// error naming the row, not a silent zero.
func TestGoldenRowsWithoutTheFarm(t *testing.T) {
	env := golden.Env{DB: testdb.New(t).App}
	for _, r := range golden.Rows() {
		if r.Got == nil {
			continue
		}
		err := r.Check(t.Context(), env)
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("golden %d ", r.ID)) || !strings.Contains(err.Error(), `"hirbovat"`) {
			t.Errorf("golden %d on an empty database: err = %v", r.ID, err)
		}
	}
}
