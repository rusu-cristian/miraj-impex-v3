package calc

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

func TestCostOf(t *testing.T) {
	d := decimal.RequireFromString
	show := func(v *decimal.Decimal, places int32) string {
		if v == nil {
			return "nil"
		}
		return v.Round(places).String()
	}
	for _, tc := range []struct {
		name                  string
		lei                   string
		birds, eggs           int
		growth                bool
		perBird, perEgg       string // rounded as the test says, "nil" for none
		birdPlaces, eggPlaces int32
		reason                EggReason
	}{
		// Golden 31, Hîrbovăț: 9 562 437 ÷ 29 907 = 319,739…; ÷ 5 073 155 = 1,8849…
		{"laying cycle", "9562437", 29907, 5073155, false, "319.739", "1.8849", 3, 4, EggReasonNone},
		{"laying cycle, shown", "9562437", 29907, 5073155, false, "319.74", "1.88", 2, 2, EggReasonNone},
		// Telița: 2 031 768 ÷ 31 590 = 64,3168…, no lei per egg ("creștere, fără ouă").
		{"growth cycle", "2031768", 31590, 0, true, "64.3168", "nil", 4, 2, EggReasonGrowth},
		{"growth ignores eggs", "2031768", 31590, 1000, true, "64.32", "nil", 2, 2, EggReasonGrowth},
		// The series total: 11 594 205 ÷ 31 590 = 367,0214…; ÷ 5 073 155 = 2,2854…
		{"series total", "11594205", 31590, 5073155, false, "367.02", "2.29", 2, 2, EggReasonNone},
		{"laying without eggs yet", "1000", 100, 0, false, "10", "nil", 2, 2, EggReasonNoEggs},
		{"no birds", "1000", 0, 50, false, "nil", "20", 2, 2, EggReasonNone},
		{"no birds, growth", "1000", 0, 0, true, "nil", "nil", 2, 2, EggReasonGrowth},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := CostOf(d(tc.lei), tc.birds, tc.eggs, tc.growth)
			if got := show(c.PerBird, tc.birdPlaces); got != tc.perBird {
				t.Errorf("per bird = %s, want %s", got, tc.perBird)
			}
			if got := show(c.PerEgg, tc.eggPlaces); got != tc.perEgg {
				t.Errorf("per egg = %s, want %s", got, tc.perEgg)
			}
			if c.EggReason != tc.reason {
				t.Errorf("reason = %q, want %q", c.EggReason, tc.reason)
			}
		})
	}
}

func TestEggsBetween(t *testing.T) {
	f := newFixture(t)
	f.eggs()
	from12, from13, from14 := jan(12), jan(13), jan(14)
	for _, tc := range []struct {
		name string
		r    DayRange
		want int
	}{
		// 12.01 H5 63 (H6 has no total), 13.01 H5 50 + H6 4; the deleted 11.01 row never counts.
		{"no lower bound", DayRange{To: jan(31)}, 117},
		{"one day", DayRange{From: &from12, To: jan(12)}, 63},
		{"from a day", DayRange{From: &from13, To: jan(31)}, 54},
		{"before any egg", DayRange{To: jan(11)}, 0},
		{"after the last egg", DayRange{From: &from14, To: jan(31)}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EggsBetween(t.Context(), f.d.App, f.farmA, tc.r)
			if err != nil || got != tc.want {
				t.Errorf("got %d, %v; want %d", got, err, tc.want)
			}
		})
	}
	if got, err := EggsBetween(t.Context(), f.d.App, f.farmB, DayRange{To: jan(31)}); err != nil || got != 0 {
		t.Errorf("farm B: %d, %v; want 0", got, err)
	}
	if _, err := EggsBetween(t.Context(), f.d.App, f.farmA, DayRange{From: &from14, To: jan(13)}); !errors.Is(err, errRange) {
		t.Errorf("a range that starts after it ends: %v, want errRange", err)
	}
}

func TestSeriesPlacement(t *testing.T) {
	f := newFixture(t)
	// A second placement of the series at A (in A's cycle, from 10.01), a migrated one (no series on its entry) in
	// that cycle, a migrated one before it (no cycle of the series contains it) and a deleted one: the last two
	// never count.
	e := f.entry("placement", f.farmA, nil, jan(14))
	f.bird(e, f.farmA, "H7", jan(14), "F", "placement", 7, "")
	e = f.entry("migration", f.farmA, nil, jan(16))
	f.bird(e, f.farmA, "H7", jan(16), "M", "placement", 3, "")
	e = f.entry("migration", f.farmA, nil, jan(9))
	f.bird(e, f.farmA, "H7", jan(9), "F", "placement", 500, "")
	e = f.entry("placement", f.farmA, nil, jan(15))
	f.bird(e, f.farmA, "H7", jan(15), "M", "placement", 1000, "")
	user := f.id(`insert into app_user (name, login, password_hash, role) values ('Ion', 'ion', 'x', 'owner')`)
	f.exec(`update entry set deleted_at = $2, deleted_by = $3 where id = $1`, e, testdb.Now(), user)
	other := f.id(`insert into series (name, name_key, year, placement_date) values ('Seria U', 'seria u', 2031, $1)`, jan(20))

	for _, tc := range []struct {
		name   string
		series uuid.UUID
		only   FarmFilter
		want   PerSex[int]
	}{
		// 05.01 at B ♀ 100 · ♂ 20, 14.01 at A ♀ 7, 16.01 at A ♂ 3; transfers and adjustments are not placements.
		{"placements in any farm", f.series, AllFarms, PerSex[int]{F: 107, M: 23}},
		// A caller who sees A only never counts B's birds (M4 COR-1); both listed = every farm.
		{"only farm A", f.series, FarmFilter{IDs: []uuid.UUID{f.farmA}}, PerSex[int]{F: 7, M: 3}},
		{"only farm B", f.series, FarmFilter{IDs: []uuid.UUID{f.farmB}}, PerSex[int]{F: 100, M: 20}},
		{"both farms listed", f.series, FarmFilter{IDs: []uuid.UUID{f.farmA, f.farmB}}, PerSex[int]{F: 107, M: 23}},
		{"no farm", f.series, FarmFilter{}, PerSex[int]{}},
		{"a series without placements", other, AllFarms, PerSex[int]{}},
		{"an unknown series", uuid.New(), AllFarms, PerSex[int]{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SeriesPlacement(t.Context(), f.d.App, tc.series, tc.only)
			if err != nil || got != tc.want {
				t.Errorf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
}
