package calc_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb/seed"
)

// TestMain serves the whole test binary: the in-package tests take
// testdb.New, the seeded ones here testdb.NewSeeded.
func TestMain(m *testing.M) { seed.Main(m) }

func nov(d int) timex.Day { return timex.MustDay(2026, time.November, d) }

// TestZilnicReadersSeeded: the Zilnic figures of the plan's Assumptions and the
// S6/S16 boards on the seed. H6.1 ♀ on 17.11: deaths 10–16.11 = 6, mean 6 ÷ 7
// = 0,857… (shown 0,9, golden 33); H5.1 ♀ on 18.11: 4 ÷ 7 = 0,571… (board
// 0,6); hens at the start of 18.11: H5 8 547, H6 5 848.
func TestZilnicReadersSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	var farm uuid.UUID
	if err := d.App.QueryRow(ctx, `select id from farm where name_key = 'hirbovat'`).Scan(&farm); err != nil {
		t.Fatal(err)
	}
	deaths := func(day timex.Day) map[string]calc.SectionDeaths {
		t.Helper()
		rows, err := calc.DeathsWindow(ctx, d.App, farm, day)
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]calc.SectionDeaths{}
		for _, r := range rows {
			out[r.Code] = r
		}
		return out
	}

	h61 := deaths(nov(17))["H6.1"]
	if h61.Week.F != 6 {
		t.Errorf("H6.1 ♀ deaths 10–16.11 = %d, want 6", h61.Week.F)
	}
	mean := calc.Mean7(h61.Week.F)
	if got := num.Round(mean, 1); !got.Equal(decimal.RequireFromString("0.9")) {
		t.Errorf("H6.1 ♀ mean7 on 17.11 shown = %s, want 0.9", got)
	}
	if want := decimal.NewFromInt(6).Div(decimal.NewFromInt(7)); !mean.Equal(want) {
		t.Errorf("H6.1 ♀ mean7 on 17.11 = %s, want 6/7 = %s", mean, want)
	}
	if !decimal.NewFromInt(4).GreaterThan(decimal.NewFromInt(3).Mul(mean)) {
		t.Error("H6.1 ♀ 17.11: 4 deaths should be over 3 × mean7 (golden 33's warning)")
	}

	h51 := deaths(nov(18))["H5.1"]
	if h51.Week.F != 4 {
		t.Errorf("H5.1 ♀ deaths 11–17.11 = %d, want 4", h51.Week.F)
	}
	if got := num.Round(calc.Mean7(h51.Week.F), 1); !got.Equal(decimal.RequireFromString("0.6")) {
		t.Errorf("H5.1 ♀ mean7 on 18.11 shown = %s, want 0.6", got)
	}

	units, err := calc.HensStart(ctx, d.App, farm, nov(18))
	if err != nil {
		t.Fatal(err)
	}
	hens := map[string]int{}
	for _, u := range units {
		hens[u.Code] = u.HensStart
	}
	for code, want := range map[string]int{"H5": 8547, "H6": 5848} {
		if hens[code] != want {
			t.Errorf("%s hens at the start of 18.11 = %d, want %d (all units: %v)", code, hens[code], want, hens)
		}
	}
}
