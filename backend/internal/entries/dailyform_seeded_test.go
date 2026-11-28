package entries_test

import (
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/entries"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/testdb"
)

// T4 criterion 2 on the dev seed: Ion Rusu's Zilnic form of Hîrbovăț on
// 18.11.2026 (§5.3 counts at the start of the day, the S6 boards' hints,
// §5.6 hens and guide).
func TestDailyFormSeeded(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	hirbovat := seededFarm(t, d.App, "hirbovat")
	ion := seededUser(t, d.App, "ion.rusu")

	res, err := svc.DailyForm(ctx, ion, hirbovat, testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	if res.SAPT == nil || *res.SAPT != 59 || res.AgeDays == nil || *res.AgeDays != 407 || res.LockedReason != nil || res.Empty != nil {
		t.Errorf("head: sapt %v, ageDays %v, lockedReason %v, empty %+v", res.SAPT, res.AgeDays, res.LockedReason, res.Empty)
	}

	seventh := func(n int64) decimal.Decimal { return decimal.NewFromInt(n).Div(decimal.NewFromInt(7)) }
	type sec struct {
		alive     httpx.SexPair[int]
		mean7F    decimal.Decimal
		yesterday int
	}
	want := map[string]sec{
		"H5.1": {httpx.SexPair[int]{F: 4264, M: 342}, seventh(4), 0},  // board 0,6
		"H5.2": {httpx.SexPair[int]{F: 4283, M: 323}, seventh(10), 3}, // 1,428…
		"H6.1": {httpx.SexPair[int]{F: 4218, M: 325}, seventh(9), 4},  // 1,285…
		"H6.2": {httpx.SexPair[int]{F: 1630, M: 93}, seventh(5), 0},
	}
	var halls, codes []string
	for _, h := range res.Halls {
		if h.Hall == nil {
			t.Fatalf("a Hîrbovăț group without a hall: %+v", h)
		}
		halls = append(halls, h.Hall.Code)
		for _, s := range h.Sections {
			codes = append(codes, s.Code)
			w, ok := want[s.Code]
			if !ok {
				continue
			}
			if s.AliveStart != w.alive || !s.Mean7.F.Decimal().Equal(w.mean7F) || s.Yesterday.F != w.yesterday ||
				!slices.Equal(s.Sexes, []dbq.Sex{dbq.SexF, dbq.SexM}) {
				t.Errorf("%s = alive %+v, mean7 ♀ %s, yesterday ♀ %d, sexes %v; want %+v, %s, %d", s.Code,
					s.AliveStart, s.Mean7.F, s.Yesterday.F, s.Sexes, w.alive, w.mean7F, w.yesterday)
			}
			if s.Saved == nil || s.Saved.F == nil || s.Saved.F.Mortality == nil {
				t.Errorf("%s saved = %+v, want Dr. Ceban's mortality", s.Code, s.Saved)
			}
		}
	}
	if !slices.Equal(halls, []string{"H5", "H6"}) || !slices.Equal(codes, []string{"H5.1", "H5.2", "H6.1", "H6.2"}) {
		t.Errorf("halls %v, sections %v; want H5, H6 with H5.1 … H6.2", halls, codes)
	}

	// H5 eggs: 8 547 hens at the start (D156), guide 57,6 % at SAPT 59, the
	// seeded eggs of 18.11 (§5.6: Cat. 1 13 · Cat. 2 417 · Cat. 3 27 +
	// Incubabile 3 415, 67,04 g).
	if len(res.Halls) == 2 {
		h5, h6 := res.Halls[0].Eggs, res.Halls[1].Eggs
		if h5 == nil || h5.Code != "H5" || h5.HensStart != 8547 || h5.GuideLayPct == nil || h5.GuideLayPct.String() != "57.6" || h5.Saved == nil {
			t.Fatalf("H5 eggs = %+v", h5)
		}
		got := []string{h5.Saved.Cat1.String(), h5.Saved.Cat2.String(), h5.Saved.Cat3.String(), h5.Saved.Hatching.String(), h5.Saved.AvgWeightG.String()}
		if !slices.Equal(got, []string{"13", "417", "27", "3415", "67.04"}) {
			t.Errorf("H5 saved eggs = %v", got)
		}
		if h6 == nil || h6.HensStart != 5848 {
			t.Errorf("H6 eggs = %+v, want 5 848 hens", h6)
		}
	}

	if want := []entries.FormEmptyHall{{Hall: "H7", EmptySince: timex.MustDay(2026, time.November, 14), Label: "goală din 14.11.2026 · H7.1, H7.2"}}; !slices.Equal(res.EmptyHalls, want) {
		t.Errorf("emptyHalls = %+v, want %+v", res.EmptyHalls, want)
	}

	// Existing: Dr. Ceban's mortality (08:15) and eggs entries of 18.11; Ion
	// Rusu may correct them, not edit them.
	var got []string
	for _, e := range res.Existing {
		if e.Author == nil || *e.Author != "Dr. Ceban" || e.Editable || !e.Correctable {
			t.Errorf("existing %+v: want Dr. Ceban's, correctable only", e)
		}
		got = append(got, e.At.Format("15:04")+" "+strconv.Itoa(len(e.Sections))+" "+strconv.Itoa(len(e.Eggs)))
	}
	if !slices.Equal(got, []string{"08:15 4 0", "08:27 0 2"}) {
		t.Errorf("existing = %v, want the 08:15 mortality of 4 sections, then the eggs of H5, H6", got)
	}
}

// T4 criterion 3 on the dev seed: Dr. Ceban on 17.11 (a past day) and the
// empty Telița.
func TestDailyFormSeededLockedAndEmpty(t *testing.T) {
	d := testdb.NewSeeded(t)
	ctx := t.Context()
	svc := entries.NewService(db.NewStore(d.App), d.Clock)
	ceban := seededUser(t, d.App, "dr.ceban")
	ion := seededUser(t, d.App, "ion.rusu")

	past, err := svc.DailyForm(ctx, ceban, seededFarm(t, d.App, "hirbovat"), timex.MustDay(2026, time.November, 17))
	if err != nil {
		t.Fatal(err)
	}
	if past.LockedReason == nil || *past.LockedReason != "Pentru zilele trecute cere managerului." {
		t.Errorf("Dr. Ceban 17.11: lockedReason = %v", past.LockedReason)
	}
	for _, e := range past.Existing {
		if e.Editable || e.Correctable {
			t.Errorf("Dr. Ceban 17.11: existing %+v has an action", e)
		}
	}

	telita, err := svc.DailyForm(ctx, ion, seededFarm(t, d.App, "telita"), testdb.Today)
	if err != nil {
		t.Fatal(err)
	}
	e := telita.Empty
	if telita.Cycle != nil || len(telita.Halls) != 0 || e == nil || e.Since == nil || e.Since.String() != "2026-02-28" ||
		e.Label != "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21." ||
		e.Hint != "Ciclul pornește la prima populare sau transfer." {
		t.Errorf("Telița 18.11: cycle %+v, halls %+v, empty %+v", telita.Cycle, telita.Halls, e)
	}
}
