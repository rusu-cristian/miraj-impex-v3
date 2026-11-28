package ledger_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded.
func TestMain(m *testing.M) { seed.Main(m) }

func nov(d int) timex.Day { return timex.MustDay(2026, time.November, d) }

// firstSeededDay is the first day of the checked history (plan M5-D4).
var firstSeededDay = timex.MustDay(2026, time.February, 22)

type hirbovat struct {
	d     *testdb.DB
	farm  uuid.UUID
	sec   map[string]uuid.UUID // bird-holding section code → id
	pairs []ledger.SectionSex  // every bird-holding section × sex, by code
}

func newHirbovat(t *testing.T) hirbovat {
	t.Helper()
	h := hirbovat{d: testdb.NewSeeded(t), sec: map[string]uuid.UUID{}}
	ctx := t.Context()
	if err := h.d.App.QueryRow(ctx, `select id from farm where name_key = 'hirbovat'`).Scan(&h.farm); err != nil {
		t.Fatal(err)
	}
	rows, err := h.d.App.Query(ctx, `select s.id, s.code from section s
		where s.farm_id = $1 and not exists (select 1 from section c where c.parent_id = s.id)
		order by s.code`, h.farm)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			t.Fatal(err)
		}
		h.sec[code] = id
		h.pairs = append(h.pairs, ledger.SectionSex{SectionID: id, Sex: dbq.SexF}, ledger.SectionSex{SectionID: id, Sex: dbq.SexM})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(h.pairs) == 0 {
		t.Fatal("Hîrbovăț has no bird-holding section in the seed")
	}
	return h
}

// viewNegatives is §6.1 step 2 as SPEC words it: the negative days of the
// dense v_section_sex_day (which runs to app.today = 18.11.2026 on the test
// pool), per section × sex of the farm, keyed by "code sex".
func viewNegatives(t *testing.T, pool *pgxpool.Pool, farm uuid.UUID, from timex.Day) map[string]ledger.Negative {
	t.Helper()
	rows, err := pool.Query(t.Context(), `
		with bad as (
		  select d.section_id, d.sex, d.date, d.alive_end
		  from v_section_sex_day d
		  where d.farm_id = $1 and d.date >= $2 and d.alive_end < 0
		)
		select b.section_id, s.code, b.sex::text,
		       min(b.date), (array_agg(b.alive_end order by b.date))[1], max(b.date), min(b.alive_end),
		       (select d.alive_end from bad d where d.section_id = b.section_id and d.sex = b.sex and d.date = app_today())
		from bad b join section s on s.id = b.section_id
		group by b.section_id, s.code, b.sex`, farm, from)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]ledger.Negative{}
	for rows.Next() {
		var n ledger.Negative
		var sex string
		var first, minEnd int32
		var today *int32
		if err := rows.Scan(&n.SectionID, &n.Code, &sex, &n.FirstDate, &first, &n.LastDate, &minEnd, &today); err != nil {
			t.Fatal(err)
		}
		n.Sex = dbq.Sex(sex)
		n.FirstAliveEnd, n.MinAliveEnd = int(first), int(minEnd)
		if today != nil {
			v := int(*today)
			n.TodayAliveEnd = &v
		}
		out[n.Code+" "+sex] = n
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func (h hirbovat) check(t *testing.T, from timex.Day) map[string]ledger.Negative {
	t.Helper()
	got, err := ledger.CheckCounts(t.Context(), h.d.App, h.pairs, from, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]ledger.Negative{}
	pos := map[ledger.SectionSex]int{}
	for i, p := range h.pairs {
		pos[p] = i
	}
	last := -1
	for _, n := range got {
		i := pos[ledger.SectionSex{SectionID: n.SectionID, Sex: n.Sex}]
		if i <= last {
			t.Errorf("CheckCounts: %s %s is out of the affected order", n.Code, n.Sex)
		}
		last = i
		out[n.Code+" "+string(n.Sex)] = n
	}
	return out
}

// TestCheckCountsSeeded: the seed keeps every Hîrbovăț count ≥ 0 from
// 22.02.2026 to today, by CheckCounts and by the dense view alike.
func TestCheckCountsSeeded(t *testing.T) {
	h := newHirbovat(t)
	if got := h.check(t, firstSeededDay); len(got) != 0 {
		t.Errorf("CheckCounts on the seed = %v, want none", got)
	}
	if got := viewNegatives(t, h.d.App, h.farm, firstSeededDay); len(got) != 0 {
		t.Errorf("v_section_sex_day negative rows on the seed = %v, want none", got)
	}
}

// TestCheckCountsMatchesView (plan M5-D4): with negative days added to the
// seed, the delta-based CheckCounts reports exactly what the dense view's
// alive_end < 0 rows say, for every Hîrbovăț section × sex and several from
// days. The H5.2 ♀ case is §6.1's correction example: H5.2 ♀ 17.11 raised
// from 3 to 4 292 (here a 4 289 sale on top of the seeded value) leaves −6 on
// 17.11 and −8 today, so the bound is 4 292 + (−8) = 4 284.
func TestCheckCountsMatchesView(t *testing.T) {
	h := newHirbovat(t)
	ctx := t.Context()
	for _, code := range []string{"H5.2", "H6.1"} {
		if _, ok := h.sec[code]; !ok {
			t.Fatalf("the seed has no Hîrbovăț section %s", code)
		}
	}
	entry := func(kind string, d timex.Day) uuid.UUID {
		t.Helper()
		var id uuid.UUID
		source := "manual"
		if kind == "migration" {
			source = "migration"
		}
		if err := h.d.App.QueryRow(ctx, `insert into entry (kind, farm_id, date, source) values ($1, $2, $3, $4) returning id`,
			kind, h.farm, d, source).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	bird := func(e uuid.UUID, code string, d timex.Day, sex, typ string, qty int) {
		t.Helper()
		if _, err := h.d.App.Exec(ctx, `insert into bird_event (entry_id, farm_id, section_id, date, sex, type, qty)
			values ($1, $2, $3, $4, $5, $6, $7)`, e, h.farm, h.sec[code], d, sex, typ, qty); err != nil {
			t.Fatal(err)
		}
	}
	bird(entry("sale", nov(17)), "H5.2", nov(17), "F", "sale", -4289)
	jun := func(d int) timex.Day { return timex.MustDay(2026, time.June, d) }
	m := entry("migration", jun(1))
	bird(m, "H6.1", jun(1), "M", "adjustment", -100000) // a stretch in the history that recovers
	bird(m, "H6.1", jun(5), "M", "adjustment", 100000)

	got := h.check(t, nov(17))
	h52 := got["H5.2 F"]
	minus8 := -8
	want := ledger.Negative{SectionID: h.sec["H5.2"], Code: "H5.2", Sex: dbq.SexF, FirstDate: nov(17), FirstAliveEnd: -6,
		LastDate: nov(18), MinAliveEnd: -8, TodayAliveEnd: &minus8}
	if diff := cmp.Diff(want, h52, cmp.AllowUnexported(timex.Day{})); diff != "" {
		t.Errorf("H5.2 ♀ from 17.11 (-want +got):\n%s", diff)
	}
	if b := h52.RemovalMax(4292); b != 4284 {
		t.Errorf("bound for 4 292 = %d, want 4 284", b)
	}

	for _, from := range []timex.Day{firstSeededDay, jun(1), jun(3), jun(5), nov(17), nov(18)} {
		t.Run(from.String(), func(t *testing.T) {
			got, want := h.check(t, from), viewNegatives(t, h.d.App, h.farm, from)
			if _, ok := want["H5.2 F"]; !ok {
				t.Fatalf("the view has no negative H5.2 ♀ day from %s: the comparison would prove nothing", from)
			}
			if diff := cmp.Diff(want, got, cmp.AllowUnexported(timex.Day{})); diff != "" {
				t.Errorf("CheckCounts vs v_section_sex_day from %s (-view +CheckCounts):\n%s", from, diff)
			}
			if from.Before(jun(5)) && got["H6.1 M"].Code == "" {
				t.Errorf("from %s: the June H6.1 ♂ stretch is missing", from)
			}
		})
	}
}

// TestCheckCountsSeededTiming: one section × sex over the whole history stays
// well inside a save's budget (§11.6; NOTES-verify M5 asks ≤ 30 ms). Logged,
// with a loose bound so a slow CI machine doesn't fail it.
func TestCheckCountsSeededTiming(t *testing.T) {
	h := newHirbovat(t)
	pair := []ledger.SectionSex{{SectionID: h.sec["H5.2"], Sex: dbq.SexF}}
	ctx := context.Background()
	if _, err := ledger.CheckCounts(ctx, h.d.App, pair, firstSeededDay, testdb.Today); err != nil { // warm up
		t.Fatal(err)
	}
	const runs = 20
	start := time.Now()
	for range runs {
		if _, err := ledger.CheckCounts(ctx, h.d.App, pair, firstSeededDay, testdb.Today); err != nil {
			t.Fatal(err)
		}
	}
	per := time.Since(start) / runs
	t.Logf("CheckCounts H5.2 ♀ from %s: %s per call", firstSeededDay, per)
	if per > 300*time.Millisecond {
		t.Errorf("CheckCounts takes %s per call, over a whole save's budget", per)
	}
}
