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

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/farms"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/golden"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/periodtable"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/reports"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// m1Rows are the goldens SPEC §12.6 puts in M1.
var m1Rows = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 26, 27}

// m2Rows are the goldens of M2: SPEC §12.6 "13–25", golden 38 (M2 D8) and
// the per-hall part of golden 14 (M2 Q2; M8 adds its "since").
var m2Rows = []int{13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 38}

// m3Rows are the goldens of M3: SPEC §12.6 "28" (M3 re-points 26 and 27,
// which stay M1's).
var m3Rows = []int{28}

// m4Rows are the goldens of M4: SPEC §12.6 "29–31" (M4 D8) and golden 35
// (the series page's placement, labelled M4 in the registry; M4 cleanup).
var m4Rows = []int{29, 30, 31, 35}

// m5Rows are the goldens of M5: SPEC §12.6 "33, 37" (plan M5-D11).
var m5Rows = []int{33, 37}

// implementedMilestones are the milestones done so far: each of their rows
// must have Got, except the deferred ones.
var implementedMilestones = map[string]bool{"M1": true, "M2": true, "M3": true, "M4": true, "M5": true}

// deferredRows are rows of an implemented milestone that may still lack Got,
// with the reason. None at M5.
var deferredRows = map[int]string{}

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
// once each, a row without a name or a valid milestone, a row of an
// implemented milestone (M1–M5) without Got (deferredRows excepted), or a Got
// without a Want.
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
		if _, deferred := deferredRows[r.ID]; implementedMilestones[r.Milestone] && !deferred && r.Got == nil {
			problems = append(problems, fmt.Sprintf("golden %d is %s but has no Got", r.ID, r.Milestone))
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
	var m1, m2, m3, m4, m5, implemented []int
	for i, r := range rows {
		if r.ID != i+1 {
			t.Errorf("row %d has id %d: rows are in id order", i, r.ID)
		}
		switch r.Milestone {
		case "M1":
			m1 = append(m1, r.ID)
		case "M2":
			m2 = append(m2, r.ID)
		case "M3":
			m3 = append(m3, r.ID)
		case "M4":
			if _, deferred := deferredRows[r.ID]; !deferred {
				m4 = append(m4, r.ID)
			}
		case "M5":
			m5 = append(m5, r.ID)
		}
		if r.Got != nil {
			implemented = append(implemented, r.ID)
		}
	}
	if diff := cmp.Diff(m1Rows, m1); diff != "" {
		t.Errorf("M1 rows differ from SPEC §12.6 (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m2Rows, m2); diff != "" {
		t.Errorf("M2 rows differ from SPEC §12.6 and the M2 plan (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m3Rows, m3); diff != "" {
		t.Errorf("M3 rows differ from SPEC §12.6 (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m4Rows, m4); diff != "" {
		t.Errorf("M4 rows differ from SPEC §12.6 (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(m5Rows, m5); diff != "" {
		t.Errorf("M5 rows differ from SPEC §12.6 (-want +got):\n%s", diff)
	}
	// Rows 1–31, 33, 35, 37 and 38 (M5): 35 goldens; 32, 34 and 36 are
	// skipped (M9, M8, M10).
	want := slices.Sorted(slices.Values(slices.Concat(m1Rows, m2Rows, m3Rows, m4Rows, m5Rows)))
	if diff := cmp.Diff(want, implemented); diff != "" || len(implemented) != 35 {
		t.Errorf("implemented rows at M5 (%d, want 35; -want +got):\n%s", len(implemented), diff)
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
		{"M2 row without Got", func() []golden.Row {
			rows := golden.Rows()
			rows[19].Got = nil
			return rows
		}(), "golden 20 is M2 but has no Got"},
		{"M3 row without Got", func() []golden.Row {
			rows := golden.Rows()
			rows[27].Got = nil
			return rows
		}(), "golden 28 is M3 but has no Got"},
		{"M4 row without Got", func() []golden.Row {
			rows := golden.Rows()
			rows[30].Got = nil
			return rows
		}(), "golden 31 is M4 but has no Got"},
		{"M5 row without Got", func() []golden.Row {
			rows := golden.Rows()
			rows[36].Got = nil
			return rows
		}(), "golden 37 is M5 but has no Got"},
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
// figures are worked out by hand below. It proves every implemented row reads the
// right farm, cycle and days and rounds at its precision, before the full
// import exists. Seria 2026: placed 07.10.2025 at age 0; cycle from
// 22.02.2026 (SAPT 20), open, with a laying guide whose ♀ lay % is 70 at
// SAPT 53, 60 at 56, 55 at 57, 50 at 58 and 45 at 59. SAPT 30 = 28.04–04.05,
// 56 = 27.10–02.11, 57 = 03.11–09.11, 58 = 10.11–16.11, 59 = 17.11–23.11.
//
//	22.02.2026 placement H5.1 ♀ 60 · ♂ 8, H5.2 ♀ 20                   base ♀ 80 · ♂ 8
//	01.03.2026 adjustment H5.1 ♀ +7 (not in the base); mortality H5.1 ♂ 1
//	01.06.2026 cull H5.2 ♀ 1
//	11.11.2026 sale H5.2 ♀ 1 and adjustment H5.2 ♀ +1 (no net change: a sale before the 12–18.11 range)
//	12.11.2026 sale H5.2 ♀ 5; mortality H5.1 ♀ 0 · ♂ 0 (entered as zero)
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
//	feed 12.11: ♀ H5.1 150 g × 67 = 10,05 kg + H5.2 160 g × 19 = 3,04 kg = 13,09; ♂ H5.1 140 g × 7 = 0,98
//	deliveries: 01.11.2026 Breeder 2 100 kg and Cocoși 10 kg, unpriced; 13.11.2026 Breeder 2 40 kg,
//	            register price 311 lei (7,7750 lei/kg)
//	stock 18.11: Breeder 2 100 + 40 − 13,09 − 12,73 = 114,18 → 114,2; Cocoși 10 − 0,98 − 1,05 = 7,97 → 8,0
//	days left 18.11 (28): every section × sex with birds today has a ration, so the window is 16–18.11;
//	            Breeder 2 12,73 ÷ 3 = 4,2433 → 4,2 kg/zi, 114,18 ÷ 4,2433 = 26,91 → 26,9 zile, 18.11 + 26 = 14.12.2026;
//	            Cocoși 1,05 ÷ 3 = 0,35 → 0,4 kg/zi, 7,97 ÷ 0,35 = 22,77 → 22,8 zile, 18.11 + 22 = 10.12.2026
//
// Telița (growth, Seria 2026 cycle 07.10.2025–28.02.2026): placement H1 ♀ 30 · ♂ 5 on 07.10.2025 and
// one Breeder 2 delivery of 50 kg on 10.02.2026, register price 400 lei. The M4 rows read:
//
//	29: Hîrbovăț's range (no earlier cycle, so up to 18.11): 3 deliveries, ♀ 140 kg · ♂ 10 kg, 311 lei;
//	    unpriced 2 · 110 kg: 01.11 100 (Breeder 2), 01.11 10 (Cocoși)
//	30: 13.11: 40 kg, 311 lei → 7,775 → 7,78 lei/kg (half away from zero); stored 7,7750
//	31: Telița · Seria 2026: 400 lei ÷ 35 birds = 11,43, growth (no eggs); Hîrbovăț · Seria 2026: 311 lei
//	    ÷ 88 = 3,53, ÷ 315 eggs (80 + 60 + 44 + 50 + 40 + 41) = 0,99, 110 kg unpriced; both farms of the
//	    series have a row, so the total: 711 lei ÷ the placement 123 (35 + 88) = 5,78, ÷ 315 = 2,26,
//	    110 kg left out
//	35: the series page's placement: the first cycle's farm Telița; placed in the series ♀ 30 + 80 · ♂ 5 + 8
//	    = 110 · 13 (Hîrbovăț's own 22.02 placement counts: it is inside its cycle of the series); Telița's
//	    cycle 07.10.2025–28.02.2026 = SAPT 1–21, as the real data (only the series and the cycle)
//
// Eggs, all at hall H5 (Cat. 1 · 2 · 3 + incubabile = total; jos apart; mean
// egg weight), and the hens of H5 at the start of the day:
//
//	01.05.2026 SAPT 30  0 · 0 · 0 + 80 = 80, jos 0              hens 87
//	28.10.2026 SAPT 56  0 · 0 · 0 + 60 = 60, jos 0              hens 86
//	05.11.2026 SAPT 57  0 · 1 · 0 + 43 = 44, jos 1              hens 86
//	12.11.2026 SAPT 58  1 · 0 · 1 + 48 = 50, jos 2, 65 g        hens 86
//	17.11.2026 SAPT 59  0 · 2 · 0 + 38 = 40, jos 1, 60 g        hens 81
//	18.11.2026 SAPT 59  1 · 2 · 3 + 35 = 41, jos 0, 62 g        hens 80
//
// so the M2 rows read (1 dp, half away from zero, from full precision):
//
//	13: guide 59 45,0, 53 70,0, 56 60,0; lay 18.11 51,25 − 45 = 6,25 → 6,3 pts > 3: abatere
//	14: H5 51,25 → 51,3 (6,3 pts); H6 never had birds: no tile
//	15: 18.11 eggs 41, incub. 35 (85,4 %), jos 0 (0,0 %)
//	16: H5 1 · 2 · 3 + 35 = 41
//	17: farm 18.11 62,0 g; SAPT 58 farm 65,0 g; no H7: ""
//	18: SAPT 59 (2 din 7 zile) eggs 81 ÷ 161 hen-days = 50,31 → 50,3; guide 45,0 → 5,3 pts;
//	    incub. 73 ÷ 81 = 90,1; jos 1 ÷ 81 = 1,2; Cat 1 ÷ 81 = 1,2 · 4 ÷ 81 = 4,9 · 3 ÷ 81 = 3,7
//	19: SAPT 59 deaths ♀ 1 + 2 = 3 · ♂ 0; sold ♀ 10 · ♂ 2; kg ♀ 12,7 · ♂ 1,1 (18.11 only)
//	20: SAPT 58 eggs 50 ÷ 86 = 58,1; guide 50,0; incub. 96,0; Cat 2,0 · 0,0 · 2,0; deaths 0 · 0;
//	    sold ♀ 1 + 5 = 6 · ♂ 0; alive end 16.11 ♀ 67 + 14 = 81 · ♂ 7
//	21: SAPT 58 kg ♀ 13,09 → 13,1 · ♂ 0,98 → 1,0; g ♀ 13 090 ÷ 86 = 152,2 · ♂ 980 ÷ 7 = 140,0
//	22: SAPT 57 44 ÷ 86 = 51,2, guide 55,0; SAPT 56 60 ÷ 86 = 69,8, guide 60,0, incub. 100,0
//	23: peak SAPT 30: 80 ÷ 87 = 91,95 → 92,0
//	24: SAPT 56–59 eggs 235 ÷ 419 hen-days = 56,1; guide (86 × 60 + 86 × 55 + 86 × 50 + 161 × 45) ÷ 419
//	    = 51,16 → 51,2; deaths ♀ 3 · ♂ 0; sold ♀ 16 · ♂ 2; incub. 224 ÷ 235 = 95,3; kg ♀ 25,82 → 25,8 · ♂ 2,03 → 2,0
//	25: depends only on the cycle (SAPT 20–59): the real golden value
//	38: no hall emptied in the cycle: []
//
// and the M5 rows read the Zilnic values of 17.11.2026 (H5.1 ♀ mortality 1):
//
//	33: H5.1 ♀ 1 death; its deaths 10–16.11 are the 12.11 zero → mean 0,0, and a 0 mean warns on any death (D53)
//	37: H5.2 has no mortality saved on 17.11: an error (TestCorrectionRowOnHandBuiltLedger adds one)
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
	guide := l.id(`insert into guide (name, name_key, type) values ('Ghid ouat', 'ghidouat', 'laying')`)
	for k, lay := range map[int]float64{53: 70, 56: 60, 57: 55, 58: 50, 59: 45} {
		if _, err := l.d.App.Exec(context.Background(),
			`insert into guide_week (guide_id, week, sex, lay_pct) values ($1, $2, 'F', $3)`, guide, k, lay); err != nil {
			t.Fatal(err)
		}
	}
	l.id(`insert into cycle (series_id, farm_id, start_date, started_by, guide_id) values ($1, $2, $3, 'auto', $4)`,
		series, l.farm, day(2026, time.February, 22), guide)
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
	e = l.entry("daily", day(2026, time.November, 12), nil)
	l.bird(e, "H5.1", day(2026, time.November, 12), "F", "mortality", 0)
	l.bird(e, "H5.1", day(2026, time.November, 12), "M", "mortality", 0)
	e = l.entry("daily", day(2026, time.November, 17), nil)
	l.bird(e, "H5.1", day(2026, time.November, 17), "F", "mortality", -1)

	eggs := func(d timex.Day, floor, cat1, cat2, cat3, hatching int, weight *float64) {
		t.Helper()
		e := l.entry("daily", d, nil)
		l.id(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, e, l.farm, l.sec["H5"], d, floor, cat1, cat2, cat3, hatching, weight)
	}
	grams := func(g float64) *float64 { return &g }
	eggs(day(2026, time.May, 1), 0, 0, 0, 0, 80, nil)
	eggs(day(2026, time.October, 28), 0, 0, 0, 0, 60, nil)
	eggs(day(2026, time.November, 5), 1, 0, 1, 0, 43, nil)
	eggs(day(2026, time.November, 12), 2, 1, 0, 1, 48, grams(65))
	eggs(day(2026, time.November, 17), 1, 0, 2, 0, 38, grams(60))

	today := golden.AsOf
	e = l.entry("daily", today, nil)
	l.bird(e, "H5.1", today, "F", "mortality", -2)
	l.id(`insert into egg_production (entry_id, farm_id, section_id, date, floor_eggs, cat1, cat2, cat3, hatching_eggs, avg_egg_weight_g)
		values ($1, $2, $3, $4, 0, 1, 2, 3, 35, 62.00)`, e, l.farm, l.sec["H5"], today)
	e = l.entry("sale", today, nil)
	l.bird(e, "H5.2", today, "F", "sale", -10)
	l.bird(e, "H5.1", today, "M", "sale", -2)

	breeder := l.id(`insert into feed_type (name, name_key, for_sex) values ('Breeder 2', 'breeder2', 'F')`)
	cocosi := l.id(`insert into feed_type (name, name_key, for_sex) values ('Cocoși', 'cocosi', 'M')`)
	e = l.entry("delivery", day(2026, time.November, 1), nil)
	delivery := `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg) values ($1, $2, $3, $4, $5)`
	l.id(delivery, e, l.farm, breeder, day(2026, time.November, 1), 100)
	l.id(delivery, e, l.farm, cocosi, day(2026, time.November, 1), 10)
	registerPriced := `insert into feed_delivery (entry_id, farm_id, feed_type_id, date, qty_kg, price_per_kg, total_lei, price_from_register)
		values ($1, $2, $3, $4, $5, $6, $7, true)`
	d13 := day(2026, time.November, 13)
	e = l.entry("delivery", d13, nil)
	l.id(registerPriced, e, l.farm, breeder, d13, 40, 7.775, 311)

	// Telița: its placement and one priced delivery inside its Seria 2026 cycle.
	h1 := l.id(`insert into section (farm_id, code, sort_order) values ($1, 'H1', 1)`, other)
	d0710 := day(2025, time.October, 7)
	e = l.id(`insert into entry (kind, farm_id, date, series_id, source) values ('placement', $1, $2, $3, 'manual')`, other, d0710, series)
	placement := `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty) values ($1, $2, $3, $4, $5, 'placement', $6)`
	l.id(placement, e, other, h1, d0710, "F", 30)
	l.id(placement, e, other, h1, d0710, "M", 5)
	d1002 := day(2026, time.February, 10)
	e = l.id(`insert into entry (kind, farm_id, date, source) values ('delivery', $1, $2, 'manual')`, other, d1002)
	l.id(registerPriced, e, other, breeder, d1002, 50, 8, 400)
	alloc := `insert into feed_allocation (entry_id, farm_id, section_id, date, sex, feed_type_id, g_per_bird)
		values ($1, $2, $3, $4, $5, $6, $7)`
	d12 := day(2026, time.November, 12)
	e = l.entry("feed_daily", d12, nil)
	l.id(alloc, e, l.farm, l.sec["H5.1"], d12, "F", breeder, 150)
	l.id(alloc, e, l.farm, l.sec["H5.2"], d12, "F", breeder, 160)
	l.id(alloc, e, l.farm, l.sec["H5.1"], d12, "M", cocosi, 140)
	e = l.entry("feed_daily", today, nil)
	l.id(alloc, e, l.farm, l.sec["H5.1"], today, "F", breeder, 160)
	l.id(alloc, e, l.farm, l.sec["H5.2"], today, "F", breeder, 155)
	l.id(alloc, e, l.farm, l.sec["H5.1"], today, "M", cocosi, 150)
	return l
}

// TestGoldenRowsOnHandBuiltLedger runs every implemented row against the
// miniature ledger. Rows 1–4 and 25 depend only on the series and the cycle,
// so they must already give the real golden values; the others must give the
// hand-worked figures, and Check must report them as mismatches.
func TestGoldenRowsOnHandBuiltLedger(t *testing.T) {
	env := golden.Env{DB: newLedger(t).d.App}
	ps := func(f, m int) calc.PerSex[int] { return calc.PerSex[int]{F: f, M: m} }
	pss := func(f, m string) calc.PerSex[string] { return calc.PerSex[string]{F: f, M: m} }
	cycleOnly := map[int]bool{1: true, 2: true, 3: true, 4: true, 25: true}
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
		27: []golden.StockKg{{Name: "Breeder 2", Kg: "114.2"}, {Name: "Cocoși", Kg: "8.0"}},
		28: golden.StockDaysLeft{From: "2026-11-16", To: "2026-11-18", Items: []golden.TypeDaysLeft{
			{Type: "Breeder 2", AvgDaily: "4.2", Days: "26.9", RunOut: "2026-12-14"},
			{Type: "Cocoși", AvgDaily: "0.4", Days: "22.8", RunOut: "2026-12-10"},
		}},

		13: golden.GuideLay{Guide59: "45.0", DiffToday: "6.3", StatusToday: calc.Abatere, Guide53: "70.0", Guide56: "60.0"},
		14: []golden.HallLay{{Hall: "H5", Pct: "51.3", DiffPts: "6.3"}},
		15: golden.EggsDay{Eggs: 41, Inc: 35, IncPct: "85.4", Jos: 0, JosPct: "0.0"},
		16: []golden.HallEggs{{Hall: "H5", Cat: [3]int{1, 2, 3}, Inc: 35, Total: 41}},
		17: golden.EggWeight{FarmDay: "62.0", H7Week58: "", FarmWeek58: "65.0"},
		18: golden.LayWeek{Label: "SAPT 59 · 17.11–18.11 · 2 din 7 zile", DaysInRow: 2, Eggs: 81, HenDays: 161,
			Lay: "50.3", Guide: "45.0", DiffPts: "5.3", IncPct: "90.1", JosPct: "1.2", CatPct: [3]string{"1.2", "4.9", "3.7"}},
		19: golden.WeekRemovals{Deaths: ps(3, 0), Sold: ps(10, 2), FeedKg: pss("12.7", "1.1")},
		20: golden.Week58{Eggs: 50, HenDays: 86, Lay: "58.1", Guide: "50.0", IncPct: "96.0", CatPct: [3]string{"2.0", "0.0", "2.0"},
			Deaths: ps(0, 0), Sold: ps(6, 0), AliveEnd: ps(81, 7)},
		21: golden.WeekFeed{Kg: pss("13.1", "1.0"), G: pss("152.2", "140.0")},
		22: golden.Weeks5756{
			Weeks:    []golden.WeekLay{{Sapt: 57, Eggs: 44, Lay: "51.2", Guide: "55.0"}, {Sapt: 56, Eggs: 60, Lay: "69.8", Guide: "60.0"}},
			IncPct56: "100.0",
		},
		23: golden.PeakLay{Sapt: 30, Eggs: 80, HenDays: 87, Lay: "92.0"},
		24: golden.Foot{Label: "Total SAPT 56–59", Eggs: 235, HenDays: 419, Lay: "56.1", Guide: "51.2",
			Deaths: ps(3, 0), Sold: ps(16, 2), IncPct: "95.3", FeedKg: pss("25.8", "2.0")},
		38: []golden.EmptyHall{},

		29: golden.Deliveries{Count: 3, KgF: "140", KgM: "10", Lei: "311", Unpriced: golden.UnpricedKg{
			Count: 2, Kg: "110", Items: []golden.DeliveryKg{{Date: "2026-11-01", Kg: "100"}, {Date: "2026-11-01", Kg: "10"}},
		}},
		30: golden.Price{Kg: "40", TotalLei: "311", PerKg: "7.78", StoredPerKg: "7.7750"},
		31: golden.Cost{
			Cycles: []golden.CycleCost{
				{Cycle: "Telița · Seria 2026", Lei: "400", Birds: 35, PerBird: "11.43"},
				{Cycle: "Hîrbovăț · Seria 2026", Lei: "311", Birds: 88, Eggs: 315, PerBird: "3.53", PerEgg: "0.99", UnpricedKg: "110"},
			},
			Total: golden.CostTotal{Lei: "711", PerBird: "5.78", PerEgg: "2.26", ExcludedKg: "110"},
		},
		35: golden.SeriesPlacement{Farm: "Telița", Placed: ps(110, 13), CycleSAPT: [2]int{1, 21}},

		33: []golden.Spike{{Section: "H5.1", Sex: "F", Deaths: 1, Mean7: "0.0",
			Message: "H5.1: 1 găină moartă, de peste 3 ori media pe 7 zile (0,0). Verifică numărătoarea; poți salva."}},
	}
	// Rows that must fail on the miniature ledger, with their error.
	wantErr := map[int]string{
		37: "golden 37 correction 17.11 H5.2: H5.2 ♀ has no mortality saved on 2026-11-17",
	}
	for _, r := range golden.Rows() {
		if r.Got == nil {
			continue
		}
		t.Run(fmt.Sprintf("%02d %s", r.ID, r.Name), func(t *testing.T) {
			ctx := t.Context()
			if cycleOnly[r.ID] {
				if err := r.Check(ctx, env); err != nil {
					t.Fatal(err)
				}
				return
			}
			if msg, ok := wantErr[r.ID]; ok {
				if err := r.Check(ctx, env); err == nil || err.Error() != msg {
					t.Errorf("Check = %v, want the error %q", err, msg)
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

// TestPriceRowRefusesAStoredPriceOffTheDerivation shows golden 30 is an
// error naming both prices, not a pass, when the stored price per kg is not
// the derivation of the delivery's total at 4 dp.
func TestPriceRowRefusesAStoredPriceOffTheDerivation(t *testing.T) {
	l := newLedger(t)
	ctx := t.Context()
	if _, err := l.d.App.Exec(ctx, `update feed_delivery set price_per_kg = 7.77 where date = $1`, day(2026, time.November, 13)); err != nil {
		t.Fatal(err)
	}
	row := golden.Rows()[29]
	err := row.Check(ctx, golden.Env{DB: l.d.App})
	want := "golden 30 13.11.2026 delivery price: the delivery of 2026-11-13 stores 7.7700 lei/kg, the derivation gives 7.7750"
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %s", err, want)
	}
}

// TestCorrectionRowOnHandBuiltLedger runs golden 37 on the miniature ledger
// with a Zilnic entry holding H5.2 ♀ mortality 3 on 17.11.2026. H5.2 ♀:
// placed 20, cull 1 (01.06), sales 1 + 5 (11.11, 12.11) against the +1
// adjustment, these 3 deaths → 11 at the end of 17.11; the 18.11 sale of 10
// → 1. The old value comes from where the entry says: the illustrative 1 for
// a value without an author (the register), the latest update audit row's
// `old` for an authored entry, and an error for an authored entry without
// one.
func TestCorrectionRowOnHandBuiltLedger(t *testing.T) {
	row := golden.Rows()[36]
	if row.ID != 37 {
		t.Fatalf("row %d, want golden 37", row.ID)
	}
	withEntry := func(t *testing.T) (*ledger, uuid.UUID) {
		l := newLedger(t)
		e := l.entry("daily", day17, nil)
		l.bird(e, "H5.2", day17, "F", "mortality", -3)
		return l, e
	}
	author := func(t *testing.T, l *ledger, entry uuid.UUID) uuid.UUID {
		user := l.id(`insert into app_user (name, login, password_hash, role) values ('Dr. Ceban', 'dr.ceban', 'x', 'vet')`)
		if _, err := l.d.App.Exec(t.Context(), `update entry set created_by = $2 where id = $1`, entry, user); err != nil {
			t.Fatal(err)
		}
		return user
	}
	want := func(old int) golden.Correction {
		return golden.Correction{Section: "H5.2", Sex: "F", AliveEnd17: 11, AliveEnd18: 1, Negative: 0,
			Change: golden.ValueChange{Key: "H5.2 · găini moarte", Old: old, New: 3}}
	}

	t.Run("a register value takes the illustrative old value 1", func(t *testing.T) {
		l, _ := withEntry(t)
		got, err := row.Got(t.Context(), golden.Env{DB: l.d.App})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(want(1), got); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("an authored entry takes the old value of its update audit row", func(t *testing.T) {
		l, e := withEntry(t)
		user := author(t, l, e)
		kind := dbq.EntryKindDaily
		if err := audit.Write(t.Context(), dbq.New(l.d.App), audit.Row{
			At: golden.Now, UserID: &user, Action: dbq.AuditActionUpdate, Entity: "entry", EntityID: e, EntryID: &e,
			EntryKind: &kind, FarmIDs: []uuid.UUID{l.farm}, RecordDate: &day17, Source: dbq.RecordSourceManual,
			Changes: []audit.Change{{Key: "H5.2 · găini moarte", Section: "H5.2", Sex: dbq.SexF, Field: "mortality",
				Old: audit.Int(2), New: audit.Int(3)}},
		}); err != nil {
			t.Fatal(err)
		}
		got, err := row.Got(t.Context(), golden.Env{DB: l.d.App})
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(want(2), got); diff != "" {
			t.Errorf("(-want +got):\n%s", diff)
		}
	})
	t.Run("an authored entry without an audit row is an error", func(t *testing.T) {
		l, e := withEntry(t)
		author(t, l, e)
		_, err := row.Got(t.Context(), golden.Env{DB: l.d.App})
		if err == nil || !strings.Contains(err.Error(), "has no audit row") {
			t.Errorf("err = %v, want the entry has no audit row", err)
		}
	})
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

// TestFeedDayAgreesWithCalc (M3 D12): golden 26 reads the Consum day row;
// calc.FeedKg, M1's reader of the same day, gives the same kg per sex at
// full precision.
func TestFeedDayAgreesWithCalc(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	res := f.table(t, "view=report-feed&farmId="+f.hirbovat.String())
	var cells map[string]string
	for _, w := range res.Weeks {
		for _, d := range w.Days {
			if d.Date == golden.AsOf {
				cells = map[string]string{}
				for _, key := range []string{periodtable.KeyFeedKgF, periodtable.KeyFeedKgM} {
					if v := d.Cells[key]; v != nil {
						cells[key] = v.Decimal().String()
					}
				}
			}
		}
	}
	if cells == nil {
		t.Fatalf("the Consum first page has no day %s", golden.AsOf)
	}
	kg, err := calc.FeedKg(ctx, f.db, f.hirbovat, golden.AsOf)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{periodtable.KeyFeedKgF: kg.F.String(), periodtable.KeyFeedKgM: kg.M.String()}
	if diff := cmp.Diff(want, cells); diff != "" {
		t.Errorf("Consum day %s vs calc.FeedKg (-calc +Consum):\n%s", golden.AsOf, diff)
	}
}

// stockFigures is one feed type's stock row as a builder shows it, at full
// precision. A builder that doesn't carry a field leaves it "".
type stockFigures struct {
	Type, Kg, AvgDaily, Window, Days, RunOut string
	Lamp                                     calc.Status
	Reason                                   string
}

func numString(n *num.Num) string {
	if n == nil {
		return "null"
	}
	return n.Decimal().String()
}

func dayString(d *timex.Day) string {
	if d == nil {
		return "null"
	}
	return d.String()
}

// TestStockAgreesAcrossBuilders (M3 D12): the Stoc report, the summary's
// feedDaysLeft and the Hrană zilnică form state show Hîrbovăț's stock as of
// 18.11 with the figures of calc.ReadStockDays, the reader golden 28 rounds,
// at full precision: kg, average, window, days left, run-out and lamp.
func TestStockAgreesAcrossBuilders(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	settings, err := calc.ReadAppSettings(ctx, f.db)
	if err != nil {
		t.Fatal(err)
	}
	sd, err := calc.ReadStockDays(ctx, f.db, f.hirbovat, golden.AsOf, settings)
	if err != nil {
		t.Fatal(err)
	}
	window := sd.Window.From.String() + "–" + sd.Window.To.String()
	var fromCalc, fromCalcNoAvg []stockFigures
	for _, it := range sd.Items {
		var days *num.Num
		if it.Left.Days != nil {
			d := num.Of(*it.Left.Days)
			days = &d
		}
		row := stockFigures{Type: it.Name, Kg: it.Kg.String(), AvgDaily: it.AvgDaily.String(), Window: window,
			Days: numString(days), RunOut: dayString(it.Left.RunOut), Lamp: it.Left.Lamp}
		fromCalc = append(fromCalc, row)
		row.AvgDaily, row.Window = "", ""
		fromCalcNoAvg = append(fromCalcNoAvg, row)
	}
	if len(fromCalc) != 2 {
		t.Fatalf("calc.ReadStockDays has %d types, want Breeder 2 and Cocoși", len(fromCalc))
	}

	t.Run("reports", func(t *testing.T) {
		res, err := reports.BuildFeed(ctx, f.db, reports.Input{Principal: owner, FarmID: &f.hirbovat, Today: golden.AsOf, Now: golden.Now})
		if err != nil {
			t.Fatal(err)
		}
		var got []stockFigures
		for _, r := range res.Stock {
			s, ok := r.(reports.TypeStock)
			if !ok {
				t.Fatalf("stock row %T, want reports.TypeStock", r)
			}
			kg, avg := s.Kg, s.AvgDaily
			got = append(got, stockFigures{Type: s.Type, Kg: numString(&kg), AvgDaily: numString(&avg),
				Window: s.Window[0].String() + "–" + s.Window[1].String(), Days: numString(s.DaysLeft),
				RunOut: dayString(s.RunOut), Lamp: s.Lamp, Reason: s.Reason})
		}
		if diff := cmp.Diff(fromCalc, got); diff != "" {
			t.Errorf("reports/feed vs calc (-calc +reports):\n%s", diff)
		}
	})
	t.Run("summary", func(t *testing.T) {
		res, err := farms.BuildSummary(ctx, f.db, f.input(owner, f.hirbovat))
		if err != nil {
			t.Fatal(err)
		}
		var got []stockFigures
		found := false
		for _, st := range res.Stats {
			s, ok := st.(farms.FeedDaysLeftStat)
			if !ok {
				continue
			}
			found = true
			for _, it := range s.Items {
				kg := it.Kg
				got = append(got, stockFigures{Type: it.Type, Kg: numString(&kg), Days: numString(it.Days),
					RunOut: dayString(it.RunOut), Lamp: it.Lamp, Reason: it.Reason})
			}
		}
		if !found {
			t.Fatalf("the summary has no %s stat", farms.StatFeedDaysLeft)
		}
		if diff := cmp.Diff(fromCalcNoAvg, got); diff != "" {
			t.Errorf("summary feedDaysLeft vs calc (-calc +summary):\n%s", diff)
		}
	})
	t.Run("form state", func(t *testing.T) {
		res, err := entries.BuildFeedForm(ctx, f.db, entries.FeedFormInput{
			Principal: owner, FarmID: f.hirbovat, Date: golden.AsOf, Today: golden.AsOf, Now: golden.Now,
		})
		if err != nil {
			t.Fatal(err)
		}
		var got []stockFigures
		for _, s := range res.Stock {
			kg, avg := s.Kg, s.AvgDaily
			got = append(got, stockFigures{Type: s.Type, Kg: numString(&kg), AvgDaily: numString(&avg),
				Window: s.Window[0].String() + "–" + s.Window[1].String(), Days: numString(s.DaysLeft),
				RunOut: dayString(s.RunOut), Lamp: s.Lamp, Reason: s.Reason})
		}
		if diff := cmp.Diff(fromCalc, got); diff != "" {
			t.Errorf("form stock vs calc (-calc +form):\n%s", diff)
		}
	})
}
