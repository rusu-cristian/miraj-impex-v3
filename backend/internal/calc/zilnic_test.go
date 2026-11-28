package calc

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/shopspring/decimal"
)

// TestDeathsWindow reads the fixture's deaths (mortality + culls, never sales
// or deleted rows) around 12–13.01 at farm A, every bird-holding section in
// tree order.
func TestDeathsWindow(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	type row struct {
		Code      string
		Yesterday PerSex[int]
		Week      PerSex[int]
	}
	read := func(day int) []row {
		t.Helper()
		got, err := DeathsWindow(ctx, f.d.App, f.farmA, jan(day))
		if err != nil {
			t.Fatal(err)
		}
		out := make([]row, 0, len(got))
		for _, s := range got {
			if s.SectionID != f.sec[s.Code] {
				t.Errorf("%s: section id %s, want %s", s.Code, s.SectionID, f.sec[s.Code])
			}
			out = append(out, row{s.Code, s.Yesterday, s.Week})
		}
		return out
	}
	tests := []struct {
		day  int
		want []row
	}{
		// 12.01: H5.1 ♀ mortality 2; H5.2 ♀ mortality 0 + cull 1; H6.1 ♂ mortality 1; the deleted H6.1 ♀ 50 ignored.
		{13, []row{
			{"H5.1", PerSex[int]{F: 2}, PerSex[int]{F: 2}},
			{"H5.2", PerSex[int]{F: 1}, PerSex[int]{F: 1}},
			{"H6.1", PerSex[int]{M: 1}, PerSex[int]{M: 1}},
			{"H7", PerSex[int]{}, PerSex[int]{}},
		}},
		// 13.01: H5.1 ♀ mortality 1 (its sale of 10 and H6.1 ♂'s sale of 3 are not deaths).
		{14, []row{
			{"H5.1", PerSex[int]{F: 1}, PerSex[int]{F: 3}},
			{"H5.2", PerSex[int]{}, PerSex[int]{F: 1}},
			{"H6.1", PerSex[int]{}, PerSex[int]{M: 1}},
			{"H7", PerSex[int]{}, PerSex[int]{}},
		}},
		// 19.01: the window 12–18.01 still holds 12.01; 20.01: 13–19.01 no longer does.
		{19, []row{
			{"H5.1", PerSex[int]{}, PerSex[int]{F: 3}},
			{"H5.2", PerSex[int]{}, PerSex[int]{F: 1}},
			{"H6.1", PerSex[int]{}, PerSex[int]{M: 1}},
			{"H7", PerSex[int]{}, PerSex[int]{}},
		}},
		{20, []row{
			{"H5.1", PerSex[int]{}, PerSex[int]{F: 1}},
			{"H5.2", PerSex[int]{}, PerSex[int]{}},
			{"H6.1", PerSex[int]{}, PerSex[int]{}},
			{"H7", PerSex[int]{}, PerSex[int]{}},
		}},
		// 12.01 itself is not in its own window.
		{12, []row{
			{"H5.1", PerSex[int]{}, PerSex[int]{}},
			{"H5.2", PerSex[int]{}, PerSex[int]{}},
			{"H6.1", PerSex[int]{}, PerSex[int]{}},
			{"H7", PerSex[int]{}, PerSex[int]{}},
		}},
	}
	for _, tc := range tests {
		if diff := cmp.Diff(tc.want, read(tc.day)); diff != "" {
			t.Errorf("DeathsWindow %02d.01 (-want +got):\n%s", tc.day, diff)
		}
	}
}

// TestHensStart: hens at the start of the day per egg unit, a hall summing its
// compartments and a parentless section counting itself; males never count.
func TestHensStart(t *testing.T) {
	f := newFixture(t)
	ctx := t.Context()
	type row struct {
		Code string
		Hens int
	}
	read := func(farm string, day int) []row {
		t.Helper()
		id := f.farmA
		if farm == "B" {
			id = f.farmB
		}
		got, err := HensStart(ctx, f.d.App, id, jan(day))
		if err != nil {
			t.Fatal(err)
		}
		out := make([]row, 0, len(got))
		for _, u := range got {
			if u.UnitID != f.sec[u.Code] {
				t.Errorf("%s: unit id %s, want %s", u.Code, u.UnitID, f.sec[u.Code])
			}
			out = append(out, row{u.Code, u.HensStart})
		}
		return out
	}
	tests := []struct {
		farm string
		day  int
		want []row
	}{
		{"A", 10, []row{{"H5", 0}, {"H6", 0}, {"H7", 0}}}, // the 10.01 transfers in count from 11.01
		{"A", 12, []row{{"H5", 57 + 37}, {"H6", 8}, {"H7", 0}}},
		{"A", 13, []row{{"H5", 55 + 36}, {"H6", 8}, {"H7", 0}}},
		{"A", 14, []row{{"H5", 44 + 36}, {"H6", 8}, {"H7", 0}}},
		{"B", 6, []row{{"T1", 100}}},
		{"B", 11, []row{{"T1", 0}}},
	}
	for _, tc := range tests {
		if diff := cmp.Diff(tc.want, read(tc.farm, tc.day)); diff != "" {
			t.Errorf("HensStart %s %02d.01 (-want +got):\n%s", tc.farm, tc.day, diff)
		}
	}
}

// TestZilnicReadersError: a failing read is returned, never an empty list.
func TestZilnicReadersError(t *testing.T) {
	f := newFixture(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if got, err := DeathsWindow(ctx, f.d.App, f.farmA, jan(13)); err == nil || got != nil {
		t.Errorf("DeathsWindow on a cancelled context = %v, %v; want nil and an error", got, err)
	}
	if got, err := HensStart(ctx, f.d.App, f.farmA, jan(13)); err == nil || got != nil {
		t.Errorf("HensStart on a cancelled context = %v, %v; want nil and an error", got, err)
	}
}

func TestMean7(t *testing.T) {
	for _, tc := range []struct {
		week int
		want string // to 4 decimals
	}{
		{6, "0.8571"},
		{4, "0.5714"},
		{0, "0"},
		{7, "1"},
		{8, "1.1429"},
	} {
		if got := Mean7(tc.week).Round(4); !got.Equal(decimal.RequireFromString(tc.want)) {
			t.Errorf("Mean7(%d) = %s, want %s", tc.week, got, tc.want)
		}
	}
	// Z6 compares at full precision: 3 × 6/7 = 2,571… < 3 (a 3 warns), and 3 × 7/7 = 3 is not > 3.
	three := decimal.NewFromInt(3)
	if !decimal.NewFromInt(3).GreaterThan(three.Mul(Mean7(6))) {
		t.Error("3 deaths against a 6/7 mean should exceed 3 × mean7")
	}
	if decimal.NewFromInt(3).GreaterThan(three.Mul(Mean7(7))) {
		t.Error("3 deaths against a mean of exactly 1 should not exceed 3 × mean7")
	}
}
