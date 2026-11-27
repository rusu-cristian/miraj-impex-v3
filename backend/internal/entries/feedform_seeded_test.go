package entries_test

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func seededFarm(t *testing.T, db dbq.DBTX, key string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	if err := db.QueryRow(t.Context(), `select id from farm where name_key = $1`, key).Scan(&id); err != nil {
		t.Fatalf("farm %q: %v", key, err)
	}
	return id
}

func fixed(n *num.Num, places int32) string {
	if n == nil {
		return "null"
	}
	return n.Decimal().StringFixed(places)
}

// formStock is a stock row at the precision the board and goldens 27–28 state.
type formStock struct {
	Type, Kg, Avg, Days, RunOut, Lamp string
	Window                            [2]timex.Day
}

func formStocks(res entries.FeedFormResponse) []formStock {
	var out []formStock
	for _, s := range res.Stock {
		runOut := "null"
		if s.RunOut != nil {
			runOut = s.RunOut.String()
		}
		out = append(out, formStock{s.Type, fixed(&s.Kg, 1), fixed(&s.AvgDaily, 1), fixed(s.DaysLeft, 3), runOut, string(s.Lamp), s.Window})
	}
	return out
}

// T5 criterion 2: Hîrbovăț's form state on 18.11.2026 on the dev seed, then
// with the 18.11 Hrană zilnică entry soft-deleted.
func TestBuildFeedFormSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	farm := seededFarm(t, d.App, "hirbovat")
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}
	build := func() entries.FeedFormResponse {
		t.Helper()
		res, err := entries.BuildFeedForm(ctx, d.App, entries.FeedFormInput{
			Principal: owner, FarmID: farm, Date: testdb.Today, Today: testdb.Today, Now: testdb.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	nov := func(day int) timex.Day { return timex.MustDay(2026, time.November, day) }

	res := build()
	if res.SAPT == nil || *res.SAPT != 59 || res.Cycle == nil || res.Cycle.Series != "Seria 2026" || res.LockedReason != nil {
		t.Errorf("sapt %v, cycle %+v, lockedReason %v; want SAPT 59, Seria 2026, none", res.SAPT, res.Cycle, res.LockedReason)
	}
	if res.Guide == nil || fixed(res.Guide.F, 0) != "157" || fixed(res.Guide.M, 0) != "156" {
		t.Errorf("guide = %+v, want ♀ 157 · ♂ 156", res.Guide)
	}
	// 8 rows: H5.1 ♀ 4 264 … H6.2 ♂ 93; ♀ 14 395 and ♂ 1 083 at the start of the day, every one already fed
	// (♀ 2 303,2 kg, ♂ 154,7 kg: golden 26).
	if n := len(res.Rows); n != 8 {
		t.Fatalf("%d rows, want 8: %+v", n, res.Rows)
	}
	first, last := res.Rows[0], res.Rows[7]
	if first.Code != "H5.1" || first.Sex != dbq.SexF || first.BirdsStart != 4264 || last.Code != "H6.2" || last.Sex != dbq.SexM || last.BirdsStart != 93 {
		t.Errorf("first %s %s %d, last %s %s %d; want H5.1 F 4264, H6.2 M 93",
			first.Code, first.Sex, first.BirdsStart, last.Code, last.Sex, last.BirdsStart)
	}
	birds := map[dbq.Sex]int{}
	kg := map[dbq.Sex]decimal.Decimal{}
	for _, r := range res.Rows {
		birds[r.Sex] += r.BirdsStart
		if r.Existing == nil {
			t.Errorf("%s %s: no existing allocation", r.Code, r.Sex)
			continue
		}
		kg[r.Sex] = kg[r.Sex].Add(r.Existing.Kg.Decimal())
		if want := map[dbq.Sex]string{dbq.SexF: "157", dbq.SexM: "156"}[r.Sex]; fixed(r.GuideG, 0) != want {
			t.Errorf("%s %s guideG = %s, want %s", r.Code, r.Sex, fixed(r.GuideG, 0), want)
		}
	}
	if birds[dbq.SexF] != 14395 || birds[dbq.SexM] != 1083 || kg[dbq.SexF].StringFixed(1) != "2303.2" || kg[dbq.SexM].StringFixed(1) != "154.7" {
		t.Errorf("birds %v, kg ♀ %s ♂ %s; want 14395 / 1083, 2303.2 / 154.7", birds, kg[dbq.SexF], kg[dbq.SexM])
	}
	if len(res.Existing) == 0 {
		t.Error("existing: no entry")
	}
	wantHalls := []entries.FormEmptyHall{{Hall: "H7", EmptySince: nov(14), Label: "goală din 14.11.2026 · H7.1, H7.2"}}
	if len(res.EmptyHalls) != 1 || res.EmptyHalls[0] != wantHalls[0] {
		t.Errorf("emptyHalls = %+v, want %+v", res.EmptyHalls, wantHalls)
	}
	// Goldens 27–28: 31 664,9 ÷ 2 392,2 = 13,237 → 01.12.2026; 1 781,0 ÷ 170,93 = 10,419 → 28.11.2026.
	w := [2]timex.Day{nov(16), nov(18)}
	want := []formStock{
		{"Breeder 2", "31664.9", "2392.2", "13.237", "2026-12-01", "in_norma", w},
		{"Cocoși", "1781.0", "170.9", "10.419", "2026-11-28", "in_norma", w},
	}
	if got := formStocks(res); !slices.Equal(got, want) {
		t.Errorf("stock\n got %+v\nwant %+v", got, want)
	}

	// Soft-delete the seeded 18.11 Hrană zilnică entry: nothing is fed, the
	// stock is back to before the save (board "din 33 968", "din 1 936") and
	// the window ends yesterday.
	var by uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from app_user where role = 'owner' limit 1`).Scan(&by); err != nil {
		t.Fatal(err)
	}
	tag, err := d.App.Exec(ctx, `update entry set deleted_at = $1, deleted_by = $2
		where farm_id = $3 and kind = 'feed_daily' and date = '2026-11-18' and deleted_at is null`, testdb.Now(), by, farm)
	if err != nil || tag.RowsAffected() != 1 {
		t.Fatalf("delete the 18.11 entry: %v rows, %v", tag.RowsAffected(), err)
	}
	if _, err := d.App.Exec(ctx, `update feed_allocation set deleted_at = $1
		where farm_id = $2 and date = '2026-11-18' and deleted_at is null`, testdb.Now(), farm); err != nil {
		t.Fatal(err)
	}
	res = build()
	for _, r := range res.Rows {
		if r.Existing != nil {
			t.Errorf("%s %s: existing %+v after the delete", r.Code, r.Sex, r.Existing)
		}
	}
	if len(res.Rows) != 8 || len(res.Existing) != 0 {
		t.Errorf("after the delete: %d rows, existing %+v; want 8 rows, []", len(res.Rows), res.Existing)
	}
	got := formStocks(res)
	if len(got) != 2 || got[0].Kg != "33968.1" || got[1].Kg != "1935.7" ||
		got[0].Window != [2]timex.Day{nov(15), nov(17)} || got[1].Window != got[0].Window {
		t.Errorf("stock after the delete = %+v, want Breeder 2 33968.1, Cocoși 1935.7, window 15–17.11", got)
	}
}

// M3 ARC-2: on the imported history and the dev seed, the form state lists
// exactly the rows the save accepts, on days with every kind of row: Hîrbovăț
// today, the day H7 was emptied and the day before, a day of Seria 2025,
// Telița inside its growth cycle, on its placement day and now (no bird).
func TestFormRowsAreSaveSlotsSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	owner := auth.Principal{UserID: uuid.New(), Role: auth.Owner}
	for _, tc := range []struct {
		farm string
		day  timex.Day
	}{
		{"hirbovat", testdb.Today},
		{"hirbovat", timex.MustDay(2026, time.November, 14)},
		{"hirbovat", timex.MustDay(2026, time.November, 13)},
		{"hirbovat", timex.MustDay(2025, time.June, 15)},
		{"telita", timex.MustDay(2025, time.October, 7)},
		{"telita", timex.MustDay(2026, time.January, 15)},
		{"telita", testdb.Today},
	} {
		farm := seededFarm(t, d.App, tc.farm)
		res, err := entries.BuildFeedForm(ctx, d.App, entries.FeedFormInput{
			Principal: owner, FarmID: farm, Date: tc.day, Today: testdb.Today, Now: testdb.Now(),
		})
		if err != nil {
			t.Fatal(err)
		}
		save, err := entries.SaveRowKeys(ctx, d.App, farm, tc.day)
		if err != nil {
			t.Fatal(err)
		}
		if form := entries.FormRowKeys(res); !slices.Equal(form, save) {
			t.Errorf("%s %s: form rows %v, save slots %v", tc.farm, tc.day, form, save)
		}
	}
}
