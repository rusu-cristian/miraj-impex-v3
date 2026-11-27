package messages

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// cedillas are the wrong Romanian letters, written as escapes so this file
// stays free of them.
const cedillas = "\u015f\u0163\u015e\u0162"

// specPath is backend/SPEC.md seen from this package's directory.
const specPath = "../../../SPEC.md"

func TestStrings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		got  string
		want string
	}{
		{"LoginRequired", LoginRequired(), "Scrie loginul."},
		{"PasswordRequired", PasswordRequired(), "Scrie parola."},
		{"LoginFailed", LoginFailed(), "Login sau parolă greșită."},
		{"LoginRateLimited", LoginRateLimited(), "Prea multe încercări. Încearcă din nou peste 5 minute."},
		{"AccountDisabled", AccountDisabled(), "Contul tău este dezactivat. Vorbește cu administratorul."},
		{"FarmForbidden", FarmForbidden(), "Nu ai acces la această fermă."},
		{"PageForbidden", PageForbidden(), "Nu ai acces la această pagină."},
		{"Internal", Internal(), "A apărut o eroare. Încearcă din nou."},
		{"PayloadTooLarge", PayloadTooLarge(), "Datele trimise sunt prea mari."},
		{"Unauthenticated", Unauthenticated(), "Sesiunea a expirat. Intră din nou în cont."},
		{"Forbidden", Forbidden(), "Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."},
		{"InvalidData", InvalidData(), "Datele trimise nu sunt valide."},
		{"Unavailable", Unavailable(), "Serviciul nu răspunde acum. Încearcă din nou."},
		{"NotFound", NotFound(), "Adresa cerută nu există."},
		{"MethodNotAllowed", MethodNotAllowed(), "Operația nu este permisă la această adresă."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
			// The expected text must be the SPEC's own text, byte for byte.
			if !strings.Contains(string(spec), `"`+tc.want+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", tc.want)
			}
		})
	}
}

// TestNoCedilla fails when ro.go or this file uses s/t with a cedilla
// (U+015F, U+0163, U+015E, U+0162) instead of the comma-below ș/ț/Ș/Ț.
func TestNoCedilla(t *testing.T) {
	for _, file := range []string{"ro.go", "ro_test.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			if strings.ContainsAny(line, cedillas) {
				t.Errorf("%s:%d uses a cedilla s/t (want U+0219/U+021B): %s", file, i+1, line)
			}
		}
	}
}

func TestMigrationStrings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := RegisterDifference(), "diferență registru"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("RegisterDifference() = %q, want %q quoted in SPEC.md", got, want)
	}
	if got, want := FeedStockMove(), "mutare ♀ → ♂"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("FeedStockMove() = %q, want %q quoted in SPEC.md", got, want)
	}
	// SPEC §10.3 gives the template `"plecare {out_date}"`.
	if !strings.Contains(string(spec), `"plecare {out_date}"`) {
		t.Error(`"plecare {out_date}" is not quoted in SPEC.md`)
	}
	if got, want := TransferDeparture("27.02.2026"), "plecare 27.02.2026"; got != want {
		t.Errorf("TransferDeparture = %q, want %q", got, want)
	}
}

func TestAuditKeys(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := AuditKeyHensDead("H5.2"), "H5.2 · găini moarte"; got != want || !strings.Contains(string(spec), `"`+want+`"`) {
		t.Errorf("AuditKeyHensDead(H5.2) = %q, want %q quoted in SPEC.md", got, want)
	}
}

func mustDay(t *testing.T, s string) timex.Day {
	t.Helper()
	d, err := timex.ParseDay(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// TestM2Strings checks the M2 copy byte for byte. Expected strings spell out
// U+202F (thousands), U+2212 (minus) and U+2013 (en dash), so a look-alike
// character fails. Each must also be quoted in SPEC.md, which writes the
// thousands separator as a plain space.
func TestM2Strings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	d := func(s string) timex.Day { return mustDay(t, s) }
	end := d("2026-01-28")
	for _, tc := range []struct {
		name string
		got  string
		want string
	}{
		{"PeriodCurrentCycle", PeriodCurrentCycle(), "Ciclul curent"},
		{"PeriodWholeCycle", PeriodWholeCycle(), "Tot ciclul"},
		{"PeriodWeeks", PeriodWeeks(4), "4 săpt."},
		{"PeriodWeeks 13", PeriodWeeks(13), "13 săpt."},
		{"PeriodInterval", PeriodInterval(d("2026-11-01"), d("2026-11-15")), "01.11–15.11.2026"},
		{"IntervalInFuture", IntervalInFuture(d("2026-11-20"), d("2026-11-18")), "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."},
		{"ScopeFarmTotal", ScopeFarmTotal(), "Total fermă"},
		{"WeekRow partial", WeekRow(59, d("2026-11-17"), d("2026-11-18")), "SAPT 59 · 17.11–18.11 · 2 din 7 zile"},
		{"WeekRow whole", WeekRow(58, d("2026-11-10"), d("2026-11-16")), "SAPT 58 · 10.11–16.11"},
		{"WeekRow one day", WeekRow(59, d("2026-11-17"), d("2026-11-17")), "SAPT 59 · 17.11 · 1 din 7 zile"},
		{"TotalsLabel", TotalsLabel(56, 59), "Total SAPT 56–59"},
		{"TotalsLabel one week", TotalsLabel(58, 58), "Total SAPT 58"},
		{"LoadMore", LoadMore(4, 56, 59, 4, 40), "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40"},
		{"LoadMore one", LoadMore(1, 21, 59, 39, 40), "Încarcă încă o săptămână · SAPT 21–59 · 39 din 40"},
		{"LoadMore 20", LoadMore(20, 40, 59, 20, 40), "Încarcă încă 20 de săptămâni · SAPT 40–59 · 20 din 40"},
		{"CycleStartNote", CycleStartNote(20), "Începutul ciclului · SAPT 20"},
		{"EmptyNext", EmptyNext(), "Alege ciclul curent"},
		{"NoteDeathsIncludeCulls", NoteDeathsIncludeCulls(), "Morți include sacrificările, ca în registru."},
		{"GroupAlive", GroupAlive(), "În viață"},
		{"GroupDeaths", GroupDeaths(), "Morți"},
		{"GroupLay", GroupLay(), "Ouat"},
		{"GroupFeedKg", GroupFeedKg(), "Hrană kg"},
		{"GroupSale", GroupSale(), "Vânzare"},
		{"GroupFeedG", GroupFeedG(), "g/pasăre"},
		{"GroupWeightF", GroupWeightF(), "Greutate ♀"},
		{"LeafHens", LeafHens(), "♀"},
		{"LeafRoosters", LeafRoosters(), "♂"},
		{"LeafEggs", LeafEggs(), "ouă"},
		{"LeafLayPct", LeafLayPct(), "% ouat"},
		{"LeafGuide", LeafGuide(), "ghid"},
		{"LeafHatching", LeafHatching(), "incub."},
		{"LeafHatchingPct", LeafHatchingPct(), "% incub."},
		{"LeafFloor", LeafFloor(), "jos"},
		{"LeafFloorPct", LeafFloorPct(), "% jos"},
		{"LeafWeightMean", LeafWeightMean(), "medie g"},
		{"LeafVsGuide", LeafVsGuide(), "față de ghid"},
		{"SexLeaf F actual", SexLeaf("F", LeafActual()), "♀ efectiv"},
		{"SexLeaf M guide", SexLeaf("M", LeafGuide()), "♂ ghid"},
		{"CycleLabel open", CycleLabel(d("2026-02-22"), nil, 20, 59), "22.02.2026–azi · SAPT 20–59"},
		{"CycleLabel closed", CycleLabel(d("2025-03-07"), &end, 19, 65), "07.03.2025–28.01.2026 · SAPT 19–65"},
		{"FarmEmpty", FarmEmpty(d("2026-02-28"), "Seria 2026", 1, 21), "Ferma e goală din 28.02.2026. Ultimul ciclu: Seria 2026, SAPT 1–21."},
		{"FarmNoCycle", FarmNoCycle(), "Ferma nu are încă niciun ciclu."},
		{"CycleStartsHint", CycleStartsHint(), "Ciclul pornește la prima populare sau transfer."},
		{"ReadoutLay", ReadoutLay(59, dec("40.4025"), dec("57.6"), dec("-17.1975")), "SAPT 59 · efectiv 40,4% · ghid 57,6% · −17,2 pct"},
		{"ReadoutLayNoGuide no guide", ReadoutLayNoGuide(59, dec("40.4025"), false), "SAPT 59 · efectiv 40,4% · ghid – fără ghid"},
		{"ReadoutLayNoGuide no row", ReadoutLayNoGuide(59, dec("40.4025"), true), "SAPT 59 · efectiv 40,4% · ghid – fără date în ghid"},
		{"ReadoutGrowth", ReadoutGrowth(21, dec("2553"), dec("2370"), dec("7.7215")), "SAPT 21 · efectiv 2 553 g · ghid 2 370 g · +7,7%"},
		{"ReadoutGrowthNoGuide no guide", ReadoutGrowthNoGuide(21, dec("2553"), false), "SAPT 21 · efectiv 2 553 g · ghid – fără ghid"},
		{"ReadoutGrowthNoGuide no row", ReadoutGrowthNoGuide(21, dec("2553"), true), "SAPT 21 · efectiv 2 553 g · ghid – fără date în ghid"},
		{"CumMortalityGuideNote", CumMortalityGuideNote(), "(ghid din intrarea în ouat)"},
		{"HallEmptySince", HallEmptySince("H7", d("2026-11-14")), "H7 goală din 14.11"},
		{"EmptyHallTile", EmptyHallTile(d("2026-11-14"), []string{"H7.1", "H7.2"}), "goală din 14.11.2026 · H7.1, H7.2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
			if inSpec := strings.ReplaceAll(tc.want, " ", " "); !strings.Contains(string(spec), `"`+inSpec+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", inSpec)
			}
		})
	}
}

func TestM2Plurals(t *testing.T) {
	for n, want := range map[int]string{
		2: "2 săptămâni", 3: "3 săptămâni", 19: "19 săptămâni", 20: "20 de săptămâni",
		43: "43 de săptămâni", 100: "100 de săptămâni", 101: "101 săptămâni", 119: "119 săptămâni", 120: "120 de săptămâni",
	} {
		if got := weeksWord(n); got != want {
			t.Errorf("weeksWord(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestWeekdays(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for wd := time.Monday; wd <= time.Saturday; wd++ {
		got = append(got, Weekday(wd))
	}
	got = append(got, Weekday(time.Sunday))
	const want = "Lu Ma Mi Jo Vi Sâ Du"
	if strings.Join(got, " ") != want || !strings.Contains(string(spec), "`"+want+"`") {
		t.Errorf("weekdays Monday first = %q, want %q (as in SPEC §5.1)", strings.Join(got, " "), want)
	}
	// 18.11.2026 is a Wednesday: "Mi" (the §7.4 day row).
	if got := Weekday(mustDay(t, "2026-11-18").Weekday()); got != "Mi" {
		t.Errorf("18.11.2026 = %q, want Mi", got)
	}
}

func TestSexLeafPanicsOnUnknownSex(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("SexLeaf(\"X\") did not panic")
		}
	}()
	SexLeaf("X", LeafGuide())
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }
