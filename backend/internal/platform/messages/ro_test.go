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

// TestM3Strings checks the M3 copy byte for byte. Expected strings spell out
// U+202F (thousands) and U+2212 (minus) as escapes, so a look-alike character
// fails. inSpec marks the strings SPEC.md already quotes (with plain spaces for
// U+202F); the others are the plan's (proposed) copy (M3 Q3, Q4, Q8, S4/S11
// boards), which SPEC.md gains with the M3 SPEC edits.
func TestM3Strings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	d := func(s string) timex.Day { return mustDay(t, s) }
	at := time.Date(2026, 11, 18, 6, 44, 59, 0, time.UTC) // 08:44 in Chisinau
	for _, tc := range []struct {
		name   string
		got    string
		want   string
		inSpec bool
	}{
		// Writes.
		{"FutureDate", FutureDate(d("2026-11-20"), d("2026-11-18")), "20.11.2026 e în viitor. Alege o dată până la 18.11.2026.", false},
		{"PastDayVet", PastDayVet(), "Pentru zilele trecute cere managerului.", true},
		{"SeriesClosed", SeriesClosed("Seria 2025"), "Seria 2025 e închisă: nu mai primește înregistrări sau corecții.", true},
		{"Stale", Stale(), "Înregistrarea s-a schimbat între timp. Reîncarcă și încearcă din nou.", true},
		{"EntryAlreadySaved", EntryAlreadySaved(), "Valorile acestei zile sunt deja introduse. Deschide înregistrarea ca să le modifici.", false},
		{"FeedAlreadyEntered", FeedAlreadyEntered("H5.1", "F", d("2026-11-18"), at), "H5.1 ♀ are deja hrana introdusă pe 18.11.2026 la 08:44. Apasă Editează ca să schimbi valorile.", false},
		{"FeedAlreadyEntered M", FeedAlreadyEntered("H6.2", "M", d("2026-11-17"), at), "H6.2 ♂ are deja hrana introdusă pe 17.11.2026 la 08:44. Apasă Editează ca să schimbi valorile.", false},
		// H1–H5.
		{"H1 above", FeedGOutOfRange("H1.1", "F", dec("350")), "H1.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g.", true},
		{"H1 above, decimal", FeedGOutOfRange("H1.1", "M", dec("300.5")), "H1.1 ♂: 300,5 g e peste limită. Scrie o valoare între 1 și 300 g.", false},
		{"H1 below", FeedGOutOfRange("H5.2", "F", dec("0")), "H5.2 ♀: 0 g e sub limită. Scrie o valoare între 1 și 300 g.", false},
		{"H1 below, negative", FeedGOutOfRange("H5.2", "F", dec("-12.5")), "H5.2 ♀: −12,5 g e sub limită. Scrie o valoare între 1 și 300 g.", false},
		{"H1 above, thousands", FeedGOutOfRange("H5.2", "F", dec("1600")), "H5.2 ♀: 1 600 g e peste limită. Scrie o valoare între 1 și 300 g.", false},
		{"H2 wrong sex F", FeedTypeWrongSex("Cocoși", "F"), "Cocoși nu e pentru găini. Alege alt tip.", false},
		{"H2 wrong sex M", FeedTypeWrongSex("Breeder 2", "M"), "Breeder 2 nu e pentru cocoși. Alege alt tip.", false},
		{"H2 retired", FeedTypeRetired("Breeder 1"), "Breeder 1 e retras. Alege alt tip.", false},
		{"H3 găini", NoBirdsAtDayStart("H7.1", "F", d("2026-11-18")), "H7.1 nu are găini la începutul zilei 18.11.2026.", false},
		{"H3 cocoși", NoBirdsAtDayStart("H7", "M", d("2026-11-18")), "H7 nu are cocoși la începutul zilei 18.11.2026.", false},
		// 173 / 156 − 1 = 10,897… % → 11%.
		{"H4 above", FeedGuideDeviation("H5.2", "M", dec("173"), dec("156"), dec("10.8974358974")), "H5.2 ♂: 173 g, cu 11% peste ghidul de 156 g. Verifică gramajul; se poate salva.", true},
		// 139 / 157 − 1 = −11,46… % → 11% sub.
		{"H4 below", FeedGuideDeviation("H5.1", "F", dec("139"), dec("157"), dec("-11.4649681529")), "H5.1 ♀: 139 g, cu 11% sub ghidul de 157 g. Verifică gramajul; se poate salva.", false},
		{"H4 half rounds away from zero", FeedGuideDeviation("H5.1", "F", dec("140.5"), dec("156"), dec("-12.5")), "H5.1 ♀: 140,5 g, cu 13% sub ghidul de 156 g. Verifică gramajul; se poate salva.", false},
		{"H5", StockWouldGoNegative("Breeder 2", "Hîrbovăț", dec("-1240.3")), "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg. Verifică livrările.", true},
		// Toast and journal summary.
		{"FeedDailySaved", FeedDailySaved(8, dec("2457.948"), "Breeder 2", dec("31664.9")), "Hrană zilnică salvată · 8 rânduri · 2 457,9 kg · stoc Breeder 2: 31 665 kg", true},
		{"FeedDailySaved one row, negative stock", FeedDailySaved(1, dec("13.113"), "Cocoși", dec("-12.6")), "Hrană zilnică salvată · 1 rând · 13,1 kg · stoc Cocoși: −13 kg", false},
		{"FeedDailySaved 20 rows", FeedDailySaved(20, dec("5000"), "Breeder 2", dec("0")), "Hrană zilnică salvată · 20 de rânduri · 5 000,0 kg · stoc Breeder 2: 0 kg", false},
		{"FeedDailySummary", FeedDailySummary(8, 8, dec("2457.948")), "Hrană zilnică · 8 din 8 rânduri · 2 457,9 kg", false},
		{"FeedDailySummary partial", FeedDailySummary(1, 8, dec("98.4")), "Hrană zilnică · 1 din 8 rânduri · 98,4 kg", false},
		// Stock reasons.
		{"StockBelowZero", StockBelowZero(), "stoc sub 0 · verifică livrările", true},
		{"StockBelowDays", StockBelowDays(dec("5")), "sub pragul de 5 zile", true},
		{"StockBelowDays one", StockBelowDays(dec("1")), "sub pragul de 1 zi", false},
		{"StockBelowDays 20", StockBelowDays(dec("20")), "sub pragul de 20 de zile", false},
		{"StockBelowDays decimal", StockBelowDays(dec("2.5")), "sub pragul de 2,5 zile", false},
		{"StockNoConsumption", StockNoConsumption(3), "– fără consum în ultimele 3 zile", true},
		{"StockNoConsumption one", StockNoConsumption(1), "– fără consum în ultima zi", false},
		{"NoOpenCycle", NoOpenCycle(), "fără ciclu deschis", true},
		{"LastCycleClosed", LastCycleClosed("Seria 2026", d("2026-02-28")), "ultimul ciclu: Seria 2026, închis pe 28.02.2026", true},
		// Consum.
		{"GroupConsumKg", GroupConsumKg(), "Hrană · kg", false},
		{"GroupConsumG", GroupConsumG(), "Hrană · g/pasăre", false},
		{"LeafState", LeafState(), "stare", false},
		{"Consum leaf actual", SexLeaf("F", LeafActual()), "♀ efectiv", true},
		{"Consum leaf vs guide", SexLeaf("M", LeafVsGuide()), "♂ față de ghid", false},
		{"NoteConsumG", NoteConsumG(dec("10")), "g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre (♀ după găini, ♂ după cocoși). Toleranță ±10%.", true},
		{"NoteRecordedKg", NoteRecordedKg("F", dec("460"), dec("689.6"), []RecordedKgWeek{
			{Sapt: 57, Spans: []RecordedKgSpan{
				{Halls: []string{"H5", "H6"}, From: d("2026-11-03"), To: d("2026-11-04")},
				{Halls: []string{"H7"}, From: d("2026-11-03"), To: d("2026-11-06")},
			}},
			{Sapt: 56, Spans: []RecordedKgSpan{{Halls: []string{"H6"}, From: d("2026-11-02"), To: d("2026-11-02")}}},
		}), "♀ 460 kg/compartiment în registru (g × păsări ≈ 690 kg): SAPT 57: H5, H6 03.11–04.11 · H7 03.11–06.11; SAPT 56: H6 02.11.", false},
		{"EmptyHallTile (Consum, form state)", EmptyHallTile(d("2026-11-14"), []string{"H7.1", "H7.2"}), "goală din 14.11.2026 · H7.1, H7.2", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
			if !tc.inSpec {
				return
			}
			if inSpec := strings.ReplaceAll(tc.want, " ", " "); !strings.Contains(string(spec), `"`+inSpec+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", inSpec)
			}
		})
	}
}

// TestM3Templates checks that the SPEC templates behind the (proposed) M3
// sentences are quoted in SPEC.md as the functions build them.
func TestM3Templates(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tpl := range []string{
		"{date} e în viitor. Alege o dată până la {today}.",  // FutureDate, §3.2
		"{type} nu e pentru {sex_word}. Alege alt tip.",      // FeedTypeWrongSex, H2
		"{sec} nu are {sex_word} la începutul zilei {date}.", // NoBirdsAtDayStart, H3
	} {
		if !strings.Contains(string(spec), `"`+tpl+`"`) {
			t.Errorf("%q is not quoted in SPEC.md", tpl)
		}
	}
}

func TestM3Plurals(t *testing.T) {
	for n, want := range map[int]string{1: "1 rând", 2: "2 rânduri", 19: "19 rânduri", 20: "20 de rânduri", 101: "101 rânduri"} {
		if got := rowsWord(n); got != want {
			t.Errorf("rowsWord(%d) = %q, want %q", n, got, want)
		}
	}
	for s, want := range map[string]string{"0": "0 zile", "1": "1 zi", "5": "5 zile", "5.0": "5 zile", "43": "43 de zile", "1.5": "1,5 zile"} {
		if got := daysWord(dec(s)); got != want {
			t.Errorf("daysWord(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestM3UnknownSexPanics(t *testing.T) {
	for name, f := range map[string]func(){
		"FeedGOutOfRange":   func() { FeedGOutOfRange("H1.1", "X", dec("350")) },
		"NoBirdsAtDayStart": func() { NoBirdsAtDayStart("H1.1", "X", timex.MustDay(2026, 11, 18)) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("no panic on sex X")
				}
			}()
			f()
		})
	}
}

// TestStockWouldGoNegativeCOR4: a stock below 0 shows whole kg, one decimal
// when whole kg round to 0, and never reads as 0 (COR-4, M4 D9).
func TestStockWouldGoNegativeCOR4(t *testing.T) {
	const head = "Stocul de Breeder 2 la Hîrbovăț ar ajunge la "
	for _, tc := range []struct{ kg, want string }{
		{"-1240", "−1 240 kg"},
		{"-1240.3", "−1 240 kg"},
		{"-0.6", "−1 kg"},
		{"-0.5", "−1 kg"},
		{"-0.45", "−0,5 kg"},
		{"-0.3", "−0,3 kg"},
		{"-0.05", "−0,1 kg"},
		{"-0.04", "−0,1 kg"},
		{"-0.0001", "−0,1 kg"},
		{"0", "0 kg"},
	} {
		want := head + tc.want + ". Verifică livrările."
		if got := StockWouldGoNegative("Breeder 2", "Hîrbovăț", dec(tc.kg)); got != want {
			t.Errorf("%s: got %q, want %q", tc.kg, got, want)
		}
	}
	// Every negative stock from −0,0001 to −2 kg reads as a negative figure.
	for i := int64(1); i <= 20000; i++ {
		got := StockWouldGoNegative("Breeder 2", "Hîrbovăț", decimal.New(-i, -4))
		if !strings.HasPrefix(got, head+"−") || strings.Contains(got, "−0 kg") || strings.Contains(got, "−0,0 kg") {
			t.Fatalf("%s: %q", decimal.New(-i, -4), got)
		}
	}
}

// TestM4Strings checks the M4 copy byte for byte (U+202F thousands, U+2212
// minus, U+2013 en dash spelled out as escapes). inSpec marks the strings
// SPEC.md quotes (plain spaces for U+202F): the board copy (S11, S12, S13)
// and the plan's (proposed) copy (M4 Q10–Q12, Q15, T7); the others are
// variants (plurals, decimals, one day) of a quoted string.
func TestM4Strings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	d := func(s string) timex.Day { return mustDay(t, s) }
	for _, tc := range []struct {
		name   string
		got    string
		want   string
		inSpec bool
	}{
		// S13 Livrare hrană.
		{"DeliveryQtyRequired", DeliveryQtyRequired("Breeder 2", "Hîrbovăț"), "Scrie câte kg de Breeder 2 s-au livrat la Hîrbovăț.", true},
		{"PriceForbidden", PriceForbidden(), "Prețul îl adaugă proprietarul în Prețuri hrană.", true},
		{"DeliveryAdded", DeliveryAdded(dec("9000"), "Breeder 2", "Hîrbovăț", dec("25140")), "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg · fără preț", true},
		{"DeliveryAdded seed re-save", DeliveryAdded(dec("9000.0"), "Breeder 2", "Hîrbovăț", dec("21119.3")), "Livrare adăugată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg · fără preț", false},
		{"DeliveryAdded decimal kg, negative stock", DeliveryAdded(dec("1000.5"), "Pre-ouat", "Telița", dec("-716")), "Livrare adăugată · 1 000,5 kg Pre-ouat · stoc Telița: −716 kg · fără preț", false},
		{"DeliverySummary", DeliverySummary(dec("9000.0"), "Breeder 2"), "Livrare · 9 000 kg Breeder 2", true},
		{"DeliverySummary decimal", DeliverySummary(dec("850.5"), "Cocoși"), "Livrare · 850,5 kg Cocoși", false},
		// S12 Prețuri hrană.
		{"PricesSaved", PricesSaved(4, dec("18160.0")), "Prețuri salvate · 4 livrări · 18 160 kg intră în cost", true},
		{"PricesSaved one", PricesSaved(1, dec("9000")), "Prețuri salvate · 1 livrare · 9 000 kg intră în cost", false},
		{"PricesSaved 20", PricesSaved(20, dec("120000")), "Prețuri salvate · 20 de livrări · 120 000 kg intră în cost", false},
		{"PricesSummary", PricesSummary(4, dec("18160")), "Prețuri · 4 livrări · 18 160 kg", true},
		{"PricePerKgNotPositive", PricePerKgNotPositive("Cocoși", d("2026-02-17")), "Cocoși 17.02.2026: scrie un preț mai mare de 0 lei/kg", true},
		{"TotalLeiNotPositive", TotalLeiNotPositive("Cocoși", d("2026-02-17")), "Cocoși 17.02.2026: scrie un total mai mare de 0 lei", true},
		{"PriceBothGiven", PriceBothGiven("Cocoși", d("2026-02-17")), "Cocoși 17.02.2026: scrie prețul pe kg sau totalul, nu amândouă.", true},
		{"AuditKeyPricePerKg", AuditKeyPricePerKg("Breeder 2", d("2026-11-12")), "Breeder 2 · 12.11.2026 · lei/kg", true},
		{"AuditKeyTotalLei", AuditKeyTotalLei("Breeder 2", d("2026-11-12")), "Breeder 2 · 12.11.2026 · total lei", true},
		{"UndoDayEnded", UndoDayEnded(), "Nu se mai poate anula: ziua salvării s-a încheiat.", true},
		{"PricesUndone", PricesUndone(4), "Prețuri anulate · 4 livrări înapoi fără preț", true},
		{"PricesUndone one", PricesUndone(1), "Prețuri anulate · 1 livrare înapoi fără preț", false},
		// S11 Modifică.
		{"PriceChangeReasonRequired", PriceChangeReasonRequired("Breeder 2", d("2026-11-13"), dec("7.7613"), dec("7.74")), "Breeder 2 13.11.2026: scrie motivul ca să schimbi prețul din 7,76 în 7,74 lei/kg", true},
		{"PriceChangeReasonRequired derived", PriceChangeReasonRequired("Breeder 2", d("2026-11-13"), dec("7.7613"), dec("8")), "Breeder 2 13.11.2026: scrie motivul ca să schimbi prețul din 7,76 în 8,00 lei/kg", false},
		{"ReasonTooLong", ReasonTooLong(212, 200), "Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți salva.", true},
		{"ReasonTooLong 201", ReasonTooLong(201, 200), "Motivul are 201 caractere; cel mult 200. Scurtează-l cu 1 ca să poți salva.", false},
		{"ReasonTooLong 220", ReasonTooLong(220, 200), "Motivul are 220 de caractere; cel mult 200. Scurtează-l cu 20 ca să poți salva.", false},
		{"NothingChanged", NothingChanged(), "Nu ai schimbat nicio cifră.", true},
		// 25 640 × 7,74 = 198 453,60 → 198 454 lei (D72 board).
		{"PriceChanged", PriceChanged("Breeder 2", d("2026-11-13"), dec("7.74"), dec("198453.60")), "Preț schimbat · Breeder 2 · 13.11.2026 · 7,74 lei/kg · 198 454 lei", true},
		{"PriceChangeSummary", PriceChangeSummary("Breeder 2", d("2026-11-13")), "Preț schimbat · Breeder 2 · 13.11.2026", true},
		// Rapoarte › Hrană.
		{"Unpriced", Unpriced(), "fără preț", true},
		{"PricedFromRegister", PricedFromRegister(), "din registru", true},
		{"CostRowLabel", CostRowLabel("Telița", "Seria 2026"), "Telița · Seria 2026", true},
		{"CostSpanClosed", CostSpanClosed(d("2025-10-07"), d("2026-02-28")), "07.10.2025–28.02.2026", true},
		{"CostSpanClosed one year", CostSpanClosed(d("2026-03-01"), d("2026-11-18")), "01.03–18.11.2026", false},
		{"CostSpanOpen", CostSpanOpen(d("2026-02-22"), 59), "din 22.02.2026 · SAPT 59", true},
		{"UnpricedSpan", UnpricedSpan(d("2026-02-17"), d("2026-11-12")), "17.02–12.11.2026", true},
		{"UnpricedSpan one day", UnpricedSpan(d("2026-02-17"), d("2026-02-17")), "17.02.2026", false},
		{"CostNoEggsGrowth", CostNoEggsGrowth(), "creștere, fără ouă", true},
		{"CostNoEggs", CostNoEggs(), "fără ouă în perioadă", true},
		{"CostSeriesTotal", CostSeriesTotal("Seria 2026"), "Total Seria 2026", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got %q, want %q", tc.got, tc.want)
			}
			if !tc.inSpec {
				return
			}
			if inSpec := strings.ReplaceAll(tc.want, " ", " "); !strings.Contains(string(spec), `"`+inSpec+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", inSpec)
			}
		})
	}
}

// TestM4BoardStrings checks the S11 Modifică and S12 în lucru sentences
// against their boards' text, tags stripped.
func TestM4BoardStrings(t *testing.T) {
	for _, tc := range []struct{ board, want string }{
		{"S12-preturi-hrana--in-lucru.dc.html", PricePerKgNotPositive("Cocoși", timex.MustDay(2026, 2, 17))},
		{"S11-rapoarte-hrana--modifica-pret--eroare.dc.html", PriceChangeReasonRequired("Breeder 2", timex.MustDay(2026, 11, 13), dec("7.7613"), dec("7.74"))},
	} {
		b, err := os.ReadFile("../../../../design/canvas/project/" + tc.board)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(boardText(string(b)), tc.want) {
			t.Errorf("%q is not in %s", tc.want, tc.board)
		}
	}
}

// boardText is a board's HTML with its tags removed (inline spans join their
// text) and runs of white space collapsed to one space.
func boardText(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func TestM4Plurals(t *testing.T) {
	for n, want := range map[int]string{1: "1 livrare", 4: "4 livrări", 19: "19 livrări", 20: "20 de livrări", 82: "82 de livrări", 101: "101 livrări"} {
		if got := deliveriesWord(n); got != want {
			t.Errorf("deliveriesWord(%d) = %q, want %q", n, got, want)
		}
	}
}

// TestM5Strings checks the M5 copy byte for byte (U+202F thousands, U+2212
// minus spelled out by the expected text). inSpec marks the strings SPEC.md
// quotes (plain spaces for U+202F): since the M5 SPEC edits (§4.10, §6.1,
// §6.2, §6.15, the plan's (proposed) copy M5 Q3–Q5, Q7–Q12, Q15 and its
// variants), every one of them.
func TestM5Strings(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	d := func(s string) timex.Day { return mustDay(t, s) }
	today := d("2026-11-18")
	at := func(hhmm string) time.Time {
		tm, err := time.ParseInLocation("2006-01-02 15:04", "2026-11-17 "+hhmm, timex.Chisinau)
		if err != nil {
			t.Fatal(err)
		}
		return tm.UTC()
	}
	ip := func(n int) *int { return &n }
	three := dec("3")
	for _, tc := range []struct {
		name   string
		got    string
		want   string
		inSpec bool
	}{
		// Z2–Z9 (§6.2).
		{"Z2", NotAnInteger("H5.1", "F"), "H5.1 ♀: scrie un număr întreg.", true},
		{"Z2 M", NotAnInteger("H6.2", "M"), "H6.2 ♂: scrie un număr întreg.", true},
		{"Z2 eggs", EggCountNotAnInteger("H5"), "H5: scrie un număr întreg.", true},
		{"Z3", NoBirdsOnDay("H7.1", "F", today), "H7.1 nu are găini pe 18.11.2026.", true},
		{"Z3 cocoși", NoBirdsOnDay("H7.2", "M", today), "H7.2 nu are cocoși pe 18.11.2026.", true},
		{"Z3 eggs on a hall without hens", NoBirdsOnDay("H7", "F", today), "H7 nu are găini pe 18.11.2026.", true},
		{"Z4", CullReasonMissing("H5.1", "F", 2, today), "H5.1 ♀: Motiv lipsă pentru 2 găini sacrificate pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul.", true},
		{"Z4 one hen", CullReasonMissing("H5.1", "F", 1, today), "H5.1 ♀: Motiv lipsă pentru 1 găină sacrificată pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul.", true},
		{"Z4 one rooster", CullReasonMissing("H5.1", "M", 1, today), "H5.1 ♂: Motiv lipsă pentru 1 cocoș sacrificat pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul.", true},
		{"Z4 20 roosters", CullReasonMissing("H5.1", "M", 20, today), "H5.1 ♂: Motiv lipsă pentru 20 de cocoși sacrificați pe 18.11.2026. Alege motivul: boală, șchiopătare sau altul.", true},
		{"Z5", DayBoundExceeded("H5.2", "M", 3230, 323, today, FieldMortality, 323), "H5.2 ♂: Mortalitate + Sacrificare + Eroare sexare = 3 230, peste cei 323 cocoși în viață pe 18.11.2026. La Mortalitate scrie cel mult 323.", true},
		{"Z5 găini, Sacrificare", DayBoundExceeded("H5.2", "F", 4300, 4283, today, FieldCull, 4280), "H5.2 ♀: Mortalitate + Sacrificare + Eroare sexare = 4 300, peste cele 4 283 găini în viață pe 18.11.2026. La Sacrificare scrie cel mult 4 280.", true},
		{"Z5 one rooster, Eroare sexare", DayBoundExceeded("H6.2", "M", 2, 1, today, FieldSexingError, 0), "H6.2 ♂: Mortalitate + Sacrificare + Eroare sexare = 2, peste cel 1 cocoș în viață pe 18.11.2026. La Eroare sexare scrie cel mult 0.", true},
		{"Z5 one hen", DayBoundExceeded("H6.2", "F", 2, 1, today, FieldMortality, 1), "H6.2 ♀: Mortalitate + Sacrificare + Eroare sexare = 2, peste cea 1 găină în viață pe 18.11.2026. La Mortalitate scrie cel mult 1.", true},
		{"Z5 banner", DayBoundBanner("H5.2", "M", 3230, 323), "refuzat · H5.2 ♂ 3 230 · cel mult 323", true},
		// 6 ÷ 7 = 0,857… → 0,9 (golden 33).
		{"Z6", MortalitySpike("H6.1", "F", 4, dec("0.857142857"), three), "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva.", true},
		{"Z6 one hen", MortalitySpike("H5.1", "F", 1, dec("0"), three), "H5.1: 1 găină moartă, de peste 3 ori media pe 7 zile (0,0). Verifică numărătoarea; poți salva.", true},
		{"Z6 one rooster", MortalitySpike("H5.1", "M", 1, dec("0.142857"), three), "H5.1: 1 cocoș mort, de peste 3 ori media pe 7 zile (0,1). Verifică numărătoarea; poți salva.", true},
		{"Z6 3 roosters", MortalitySpike("H5.1", "M", 3, dec("0.285714"), three), "H5.1: 3 cocoși morți, de peste 3 ori media pe 7 zile (0,3). Verifică numărătoarea; poți salva.", true},
		{"Z6 20 roosters", MortalitySpike("H6.2", "M", 20, dec("1"), three), "H6.2: 20 de cocoși morți, de peste 3 ori media pe 7 zile (1,0). Verifică numărătoarea; poți salva.", true},
		{"Z6 decimal factor", MortalitySpike("H6.1", "F", 4, dec("0.857142857"), dec("2.5")), "H6.1: 4 găini moarte, de peste 2,5 ori media pe 7 zile (0,9). Verifică numărătoarea; poți salva.", true},
		{"Z7", EggsOnHall("H5"), "Ouăle se scriu pe hală: H5.", true},
		{"Z7 growth farm", EggsGrowthFarm("Telița"), "Telița e fermă de creștere; ouăle se scriu doar la fermele de ouat.", true},
		{"Z8", EggWeightOutOfRange("H5", dec("130")), "H5: masa medie 130 g e în afara limitei. Scrie o valoare între 20 și 120 g.", true},
		{"Z8 decimals", EggWeightOutOfRange("H6", dec("19.50")), "H6: masa medie 19,5 g e în afara limitei. Scrie o valoare între 20 și 120 g.", true},
		{"Z9", DailyAlreadyEntered("H5.1", "F", today, at("08:15")), "H5.1 ♀ are deja valori introduse pe 18.11.2026 la 08:15. Apasă Editează ca să schimbi valorile.", true},
		{"Z9 eggs", EggsAlreadyEntered("H5", today, at("08:15")), "H5 are deja ouăle introduse pe 18.11.2026 la 08:15. Apasă Editează ca să schimbi valorile.", true},
		// Toasts and summaries (Q5).
		{"DailySaved", DailySaved(4, -3, 0), "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0", true},
		{"DailySaved one section, sexing error +", DailySaved(1, 2, -2), "Zilnic salvat · 1 secțiune · ♀ +2 · ♂ −2", true},
		{"DailySaved 20 sections", DailySaved(20, -1234, -5), "Zilnic salvat · 20 de secțiuni · ♀ −1 234 · ♂ −5", true},
		{"DailyEggsSaved", DailyEggsSaved(2, 5914), "Zilnic salvat · 2 hale · 5 914 ouă", true},
		{"DailyEggsSaved one hall", DailyEggsSaved(1, 3872), "Zilnic salvat · 1 hală · 3 872 ouă", true},
		{"DailySummary birds", DailySummary(4, 0), "Mortalitate · 4 secțiuni", true},
		{"DailySummary eggs", DailySummary(0, 2), "Ouă · 2 hale", true},
		{"DailySummary both", DailySummary(4, 2), "Mortalitate · 4 secțiuni · Ouă · 2 hale", true},
		{"DailySummary 20 halls", DailySummary(0, 20), "Ouă · 20 de hale", true},
		{"MortalitySectionSummary", MortalitySectionSummary("H5.2"), "Mortalitate · H5.2", true},
		{"EggsHallSummary", EggsHallSummary("H5"), "Ouă · H5", true},
		// §6.1 banners (Q4): first day only, first day and today, today only.
		{"banner first day only", NegativeBanner(ActionCorrection, "H6", "F", d("2026-02-16"), -4, nil, today), "Corecția ar lăsa H6 cu −4 găini pe 16.02.2026. Nicio secțiune nu poate scădea sub 0.", true},
		{"banner first day and today", NegativeBanner(ActionCorrection, "H5.2", "F", d("2026-11-17"), -6, ip(-8), today), "Corecția ar lăsa H5.2 cu −6 găini pe 17.11.2026 și cu −8 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0.", true},
		{"banner today only", NegativeBanner(ActionCorrection, "H5.2", "F", today, -8, ip(-8), today), "Corecția ar lăsa H5.2 cu −8 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0.", true},
		{"banner Salvarea", NegativeBanner(ActionSave, "H6.2", "M", today, -3, ip(-3), today), "Salvarea ar lăsa H6.2 cu −3 cocoși azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0.", true},
		{"banner Ștergerea", NegativeBanner(ActionDelete, "H6.2", "M", d("2026-11-17"), -1, nil, today), "Ștergerea ar lăsa H6.2 cu −1 cocoș pe 17.11.2026. Nicio secțiune nu poate scădea sub 0.", true},
		{"banner Restaurarea", NegativeBanner(ActionRestore, "H5.2", "F", d("2026-11-16"), -20, ip(-2), today), "Restaurarea ar lăsa H5.2 cu −20 de găini pe 16.11.2026 și cu −2 găini azi, 18.11.2026. Nicio secțiune nu poate scădea sub 0.", true},
		// §6.1 fields.
		{"removal today", RemovalBound("H5.2", "F", 4292, today, -8, today, 4284), "H5.2 ♀ 4 292: ar lăsa −8 găini azi, 18.11.2026. Scrie cel mult 4 284.", true},
		{"removal past day", RemovalBound("H6", "F", 4292, d("2026-02-16"), -4, today, 4288), "H6 ♀ 4 292: ar lăsa −4 găini pe 16.02.2026. Scrie cel mult 4 288.", true},
		{"removal cocoși (concurrency)", RemovalBound("H6.2", "M", 4, today, -3, today, 1), "H6.2 ♂ 4: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel mult 1.", true},
		{"inflow", InflowBound("H7.2", "F", 4381, today, -2, today, 4383), "H7.2 ♀ 4 381: ar lăsa −2 găini azi, 18.11.2026. Scrie cel puțin 4 383.", true},
		{"inflow past day", InflowBound("H5.2", "M", 4, d("2026-11-16"), -1, today, 5), "H5.2 ♂ 4: ar lăsa −1 cocoș pe 16.11.2026. Scrie cel puțin 5.", true},
		// Access (Q10).
		{"OwnEntryPastDay", OwnEntryPastDay(), "Poți modifica sau șterge ce ai introdus doar în aceeași zi, până la 23:59. Pentru o corecție, spune-i managerului; el o salvează cu motiv.", true},
		// Audit keys (Q12).
		{"key mortality F", AuditKeyDaily("H5.2", "F", FieldMortality), "H5.2 · găini moarte", true},
		{"key mortality M", AuditKeyDaily("H5.2", "M", FieldMortality), "H5.2 · cocoși morți", true},
		{"key cull F", AuditKeyDaily("H5.2", "F", FieldCull), "H5.2 · găini sacrificate", true},
		{"key cull M", AuditKeyDaily("H5.2", "M", FieldCull), "H5.2 · cocoși sacrificați", true},
		{"key cull reason F", AuditKeyDaily("H5.2", "F", FieldCullReason), "H5.2 · motiv sacrificare găini", true},
		{"key cull reason M", AuditKeyDaily("H5.2", "M", FieldCullReason), "H5.2 · motiv sacrificare cocoși", true},
		{"key sexing error F", AuditKeyDaily("H5.2", "F", FieldSexingError), "H5.2 · eroare sexare găini", true},
		{"key sexing error M", AuditKeyDaily("H5.2", "M", FieldSexingError), "H5.2 · eroare sexare cocoși", true},
		{"key floor", AuditKeyEggs("H5", FieldFloor), "H5 · ouă jos", true},
		{"key cat1", AuditKeyEggs("H5", FieldCat1), "H5 · Cat. 1", true},
		{"key cat2", AuditKeyEggs("H5", FieldCat2), "H5 · Cat. 2", true},
		{"key cat3", AuditKeyEggs("H5", FieldCat3), "H5 · Cat. 3", true},
		{"key hatching", AuditKeyEggs("H5", FieldHatching), "H5 · incubabile", true},
		{"key avg weight", AuditKeyEggs("H5", FieldAvgWeightG), "H5 · masa medie g", true},
		{"key feed g F", AuditKeyFeed("H5.1", "F", FieldG), "H5.1 · g găini", true},
		{"key feed g M", AuditKeyFeed("H5.1", "M", FieldG), "H5.1 · g cocoși", true},
		{"key feed type F", AuditKeyFeed("H5.1", "F", FieldFeedTypeID), "H5.1 · tip hrană găini", true},
		{"key feed type M", AuditKeyFeed("H5.1", "M", FieldFeedTypeID), "H5.1 · tip hrană cocoși", true},
		{"key delivery kg", AuditKeyDelivery(FieldQtyKg), "Livrare · kg", true},
		{"key delivery type", AuditKeyDelivery(FieldDeliveryFeedTypeID), "Livrare · tip", true},
		{"key delivery note", AuditKeyDelivery(FieldNote), "Livrare · notă", true},
		{"CullReasonWord boala", CullReasonWord("boala"), "boală", true},
		{"CullReasonWord schiopatare", CullReasonWord("schiopatare"), "șchiopătare", true},
		{"CullReasonWord altul", CullReasonWord("altul"), "altul", true},
		// Motiv missing (§6.15, Q9, Q15).
		{"Motiv missing bird", ReasonMissingChange(BirdLabel("H5.2", "F", FieldMortality), ShownInt(1), ShownInt(3)), "Motiv lipsă. Scrie de ce se schimbă H5.2 ♀ din 1 în 3; fără motiv corecția nu se salvează.", true},
		{"Motiv missing eggs", ReasonMissingChange(EggLabel("H5", FieldCat1), ShownInt(13), ShownInt(15)), "Motiv lipsă. Scrie de ce se schimbă H5 Cat. 1 din 13 în 15; fără motiv corecția nu se salvează.", true},
		{"Motiv missing feed", ReasonMissingChange(FeedLabel("H5.1", "F", FieldG), ShownDec(dec("160"), 1), ShownDec(dec("158"), 1)), "Motiv lipsă. Scrie de ce se schimbă H5.1 ♀ din 160 în 158 g; fără motiv corecția nu se salvează.", true},
		{"Motiv missing delivery", ReasonMissingChange(DeliveryLabel(FieldQtyKg), ShownDec(dec("9000"), 1), ShownDec(dec("8000"), 1)), "Motiv lipsă. Scrie de ce se schimbă livrarea din 9 000 în 8 000 kg; fără motiv corecția nu se salvează.", true},
		{"Motiv missing cull reason", ReasonMissingChange(BirdLabel("H5.2", "M", FieldCullReason), ShownCullReason("boala"), ShownCullReason("schiopatare")), "Motiv lipsă. Scrie de ce se schimbă H5.2 ♂ motiv sacrificare din boală în șchiopătare; fără motiv corecția nu se salvează.", true},
		{"Motiv missing added value", ReasonMissingChange(BirdLabel("H5.2", "F", FieldCull), ShownEmpty(), ShownInt(2)), "Motiv lipsă. Scrie de ce se schimbă H5.2 ♀ sacrificare din – în 2; fără motiv corecția nu se salvează.", true},
		{"Motiv missing several", ReasonMissingChanges(3), "Motiv lipsă. Scrie de ce se schimbă 3 valori; fără motiv corecția nu se salvează.", true},
		{"Motiv missing delete", DeleteReasonMissing(), "Motiv lipsă. Scrie de ce ștergi salvarea; fără motiv nu se șterge.", true},
		// Correction Toasts.
		{"CorrectionSavedAlive", CorrectionSavedAlive(BirdLabel("H5.2", "F", FieldMortality), ShownInt(1), ShownInt(3), "F", 4281), "Corecție salvată · H5.2 ♀ 1 → 3 · în viață ♀ 4 281", true},
		{"CorrectionSavedAlive sexing error", CorrectionSavedAlive(BirdLabel("H6.2", "M", FieldSexingError), ShownInt(-2), ShownEmpty(), "M", 95), "Corecție salvată · H6.2 ♂ eroare sexare −2 → – · în viață ♂ 95", true},
		{"CorrectionSaved eggs", CorrectionSaved(EggLabel("H5", FieldCat1), ShownInt(13), ShownInt(15)), "Corecție salvată · H5 Cat. 1 13 → 15", true},
		{"CorrectionSaved egg weight", CorrectionSaved(EggLabel("H5", FieldAvgWeightG), ShownDec(dec("67.04"), 2), ShownDec(dec("67.4"), 2)), "Corecție salvată · H5 masa medie 67,04 → 67,4 g", true},
		{"CorrectionSaved feed g", CorrectionSaved(FeedLabel("H5.1", "M", FieldG), ShownDec(dec("141"), 1), ShownDec(dec("144.5"), 1)), "Corecție salvată · H5.1 ♂ 141 → 144,5 g", true},
		{"CorrectionSaved feed type", CorrectionSaved(FeedLabel("H5.1", "F", FieldFeedTypeID), ShownText("Breeder 1"), ShownText("Breeder 2")), "Corecție salvată · H5.1 ♀ tip hrană Breeder 1 → Breeder 2", true},
		{"CorrectionSaved delivery kg", CorrectionSaved(DeliveryLabel(FieldQtyKg), ShownDec(dec("9000"), 1), ShownDec(dec("8000"), 1)), "Corecție salvată · Livrare 9 000 → 8 000 kg", true},
		{"CorrectionSaved delivery note", CorrectionSaved(DeliveryLabel(FieldNote), ShownEmpty(), ShownText("sac rupt")), "Corecție salvată · Livrare notă – → sac rupt", true},
		{"CorrectionsSavedAlive", CorrectionsSavedAlive(3, 12561, 794), "Corecție salvată · 3 valori · în viață azi ♀ 12 561 · ♂ 794", true},
		{"CorrectionsSaved", CorrectionsSaved(2), "Corecție salvată · 2 valori", true},
		{"CorrectionUndone", CorrectionUndone(BirdLabel("H5.2", "F", FieldMortality), ShownInt(3), ShownInt(1)), "Corecție anulată · H5.2 ♀ 3 → 1", true},
		{"CorrectionsUndone", CorrectionsUndone(20), "Corecție anulată · 20 de valori", true},
		{"DeliverySaved", DeliverySaved(dec("9000.0"), "Breeder 2", "Hîrbovăț", dec("21119.3")), "Livrare salvată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg", true},
		// Preview effects (§7.8, §6.15, Q7).
		{"EffectAliveFrom", EffectAliveFrom("H5.2", "F", d("2026-11-17"), -2), "H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2.", true},
		{"EffectAliveFrom rise", EffectAliveFrom("H2", "F", d("2026-11-03"), 12), "H2 · găini în viață de pe 03.11.2026 se schimbă cu +12.", true},
		{"EffectDeaths", EffectDeaths("F", -5, -7), "Mortalitate · morți ♀ −5 → −7", true},
		{"EffectFarmAliveToday", EffectFarmAliveToday("Hîrbovăț", "F", 12563, 12561), "Hîrbovăț · ♀ azi 12 563 → 12 561", true},
		{"CheckedDays", CheckedDays(2), "Verificat pe 2 zile, până azi: nicio zi sub 0.", true},
		{"CheckedDays one", CheckedDays(1), "Verificat pe 1 zi, până azi: nicio zi sub 0.", true},
		{"CheckedDays 270", CheckedDays(270), "Verificat pe 270 de zile, până azi: nicio zi sub 0.", true},
		// Delete Dialogs (§6.15, S15, S16, Q9).
		{"DailyDeleteEffect", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at("08:07"), Sections: 4, F: -5, Farm: "Hîrbovăț", AliveF: &AliveMove{12563, 12568},
		}), "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:07: 4 secțiuni · ♀ −5 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 563 → ♀ 12 568. Ștergerea rămâne în Istoric modificări.", true},
		{"DailyDeleteEffect corrected", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at("08:07"), Corrected: &Corrected{By: "Ion Rusu", On: today}, Sections: 4, F: -7,
			Farm: "Hîrbovăț", AliveF: &AliveMove{12561, 12568},
		}), "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:07 și corectată de Ion Rusu pe 18.11.2026: 4 secțiuni · ♀ −7 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 561 → ♀ 12 568. Ștergerea rămâne în Istoric modificări.", true},
		{"DailyDeleteEffect both sexes and eggs", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at("08:15"), Sections: 4, F: -3, M: -2, Halls: 2, Eggs: 5914, Farm: "Hîrbovăț",
			AliveF: &AliveMove{12561, 12564}, AliveM: &AliveMove{794, 796},
		}), "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:15: 4 secțiuni · ♀ −3 · ♂ −2 · 2 hale · 5 914 ouă. Păsări în viață azi în Hîrbovăț: ♀ 12 561 → ♀ 12 564 · ♂ 794 → ♂ 796. Ștergerea rămâne în Istoric modificări.", true},
		{"DailyDeleteEffect roosters only", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at("08:15"), Sections: 1, M: -1, Farm: "Hîrbovăț", AliveM: &AliveMove{794, 795},
		}), "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:15: 1 secțiune · ♀ 0 · ♂ −1. Cocoși în viață azi în Hîrbovăț: ♂ 794 → ♂ 795. Ștergerea rămâne în Istoric modificări.", true},
		{"DailyDeleteEffect eggs only", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at("08:27"), Halls: 1, Eggs: 3872, Farm: "Hîrbovăț",
		}), "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:27: 1 hală · 3 872 ouă. Ștergerea rămâne în Istoric modificări.", true},
		{"SectionDeleteEffect", SectionDeleteEffect(at("08:15"), "H5.2", -2, 0, &AliveMove{4281, 4283}, nil), "Se șterge rândul salvat la 08:15: ♀ −2 · ♂ 0. În viață în H5.2: ♀ 4 281 → 4 283. Ștergerea rămâne în Istoric modificări.", true},
		{"SectionDeleteEffect both sexes", SectionDeleteEffect(at("08:15"), "H6.2", -1, -2, &AliveMove{1629, 1630}, &AliveMove{5, 7}), "Se șterge rândul salvat la 08:15: ♀ −1 · ♂ −2. În viață în H6.2: ♀ 1 629 → 1 630 · ♂ 5 → 7. Ștergerea rămâne în Istoric modificări.", true},
		{"SectionDeleteEffect zero row", SectionDeleteEffect(at("08:15"), "H6.2", 0, 0, nil, nil), "Se șterge rândul salvat la 08:15: ♀ 0 · ♂ 0. Ștergerea rămâne în Istoric modificări.", true},
		{"HallEggsDeleteEffect", HallEggsDeleteEffect(at("08:27"), 3872), "Se șterge rândul salvat la 08:27: 3 872 ouă. Ștergerea rămâne în Istoric modificări.", true},
		{"FeedDailyDeleteEffect", FeedDailyDeleteEffect("Dr. Ceban", at("08:44"), nil, 8, dec("2457.948")), "Se șterge salvarea Hrană zilnică introdusă de Dr. Ceban la 08:44: 8 rânduri · 2 457,9 kg. Ștergerea rămâne în Istoric modificări.", true},
		{"FeedDailyDeleteEffect corrected", FeedDailyDeleteEffect("Dr. Ceban", at("08:44"), &Corrected{By: "Ion Rusu", On: today}, 1, dec("98.4")), "Se șterge salvarea Hrană zilnică introdusă de Dr. Ceban la 08:44 și corectată de Ion Rusu pe 18.11.2026: 1 rând · 98,4 kg. Ștergerea rămâne în Istoric modificări.", true},
		{"DeliveryDeleteEffect", DeliveryDeleteEffect("Ion Rusu", at("10:30"), nil, dec("9000"), "Breeder 2"), "Se șterge livrarea introdusă de Ion Rusu la 10:30: 9 000 kg Breeder 2. Ștergerea rămâne în Istoric modificări.", true},
		// Delete, undo, restore Toasts (Q15).
		{"MortalityDeleted", MortalityDeleted(d("2026-11-17"), 4, ip(12568), nil), "Mortalitate ștearsă · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 568", true},
		{"MortalityDeleted both sexes", MortalityDeleted(d("2026-11-17"), 1, ip(12568), ip(796)), "Mortalitate ștearsă · 17.11.2026 · 1 secțiune · în viață azi ♀ 12 568 · ♂ 796", true},
		{"MortalitySectionDeleted", MortalitySectionDeleted("H5.2", ip(4283), nil), "Mortalitate ștearsă · H5.2 · în viață ♀ 4 283", true},
		{"MortalitySectionDeleted roosters", MortalitySectionDeleted("H6.2", nil, ip(7)), "Mortalitate ștearsă · H6.2 · în viață ♂ 7", true},
		{"EggsDeleted", EggsDeleted(today, []string{"H5"}), "Ouă șterse · 18.11.2026 · H5", true},
		{"EggsDeleted two halls", EggsDeleted(today, []string{"H5", "H6"}), "Ouă șterse · 18.11.2026 · H5, H6", true},
		{"FeedDailyDeleted", FeedDailyDeleted(today, 8, "Breeder 2", dec("33968.2")), "Hrană zilnică ștearsă · 18.11.2026 · 8 rânduri · stoc Breeder 2: 33 968 kg", true},
		{"DeliveryDeleted", DeliveryDeleted(dec("9000"), "Breeder 2", "Hîrbovăț", dec("22665")), "Livrare ștearsă · 9 000 kg Breeder 2 · stoc Hîrbovăț: 22 665 kg", true},
		{"DailyUndone", DailyUndone(today), "Zilnic anulat · 18.11.2026", true},
		{"FeedDailyUndone", FeedDailyUndone(today), "Hrană zilnică anulată · 18.11.2026", true},
		{"DeliveryUndone", DeliveryUndone(dec("9000.0"), "Breeder 2"), "Livrare anulată · 9 000 kg Breeder 2", true},
		{"MortalityRestored", MortalityRestored(d("2026-11-17"), 4, ip(12561), nil), "Mortalitate restaurată · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 561", true},
		{"MortalitySectionRestored", MortalitySectionRestored("H5.2", ip(4281), nil), "Mortalitate restaurată · H5.2 · în viață ♀ 4 281", true},
		{"EggsRestored", EggsRestored(today, []string{"H5"}), "Ouă restaurate · 18.11.2026 · H5", true},
		{"FeedDailyRestored", FeedDailyRestored(today, 8, "Breeder 2", dec("31664.9")), "Hrană zilnică restaurată · 18.11.2026 · 8 rânduri · stoc Breeder 2: 31 665 kg", true},
		{"DeliveryRestored", DeliveryRestored(dec("9000"), "Breeder 2", "Hîrbovăț", dec("25140")), "Livrare restaurată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("got  %q\nwant %q", tc.got, tc.want)
			}
			if !tc.inSpec {
				return
			}
			if inSpec := strings.ReplaceAll(tc.want, " ", " "); !strings.Contains(string(spec), `"`+inSpec+`"`) {
				t.Fatalf("%q is not quoted in SPEC.md", inSpec)
			}
		})
	}
}

// TestM5SpecShortForm: the "only the first negative day" banner starts with
// SPEC §6.1's short form (spec E1) and ends with the banner's closing
// sentence.
func TestM5SpecShortForm(t *testing.T) {
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	const short = "Corecția ar lăsa H6 cu −4 găini pe 16.02.2026."
	if !strings.Contains(string(spec), `"`+short+`"`) {
		t.Fatalf("%q is not quoted in SPEC.md", short)
	}
	got := NegativeBanner(ActionCorrection, "H6", "F", mustDay(t, "2026-02-16"), -4, nil, mustDay(t, "2026-11-18"))
	if got != short+" Nicio secțiune nu poate scădea sub 0." {
		t.Errorf("got %q", got)
	}
}

// TestM5BoardStrings checks the M5 sentences that a board shows as one text
// (tags stripped) against the board itself.
func TestM5BoardStrings(t *testing.T) {
	today := timex.MustDay(2026, 11, 18)
	at0807 := time.Date(2026, 11, 17, 8, 7, 0, 0, timex.Chisinau)
	minus8 := -8
	for _, tc := range []struct{ board, want string }{
		{"S16-corecteaza--refuzat.dc.html", MortalitySpike("H6.1", "F", 4, dec("0.857142857"), dec("3"))},
		{"S6-zilnic--abatere.dc.html", MortalitySpike("H5.1", "F", 4, dec("0.571428571"), dec("3"))},
		{"S16-corecteaza--refuzat.dc.html", RemovalBound("H5.2", "F", 4292, today, -8, today, 4284)},
		{"S16-corecteaza--refuzat.dc.html", NegativeBanner(ActionCorrection, "H5.2", "F", timex.MustDay(2026, 11, 17), -6, &minus8, today)},
		{"S16-corecteaza--motiv-lipsa.dc.html", ReasonMissingChange(BirdLabel("H5.2", "F", FieldMortality), ShownInt(1), ShownInt(3))},
		{"S16-corecteaza--vet-blocat.dc.html", OwnEntryPastDay()},
		{"S6-zilnic--motiv-lipsa.dc.html", CullReasonMissing("H5.1", "F", 2, today)},
		{"S16-corecteaza--sterge-confirmare.dc.html", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at0807, Sections: 4, F: -5, Farm: "Hîrbovăț", AliveF: &AliveMove{12563, 12568},
		})},
		{"S16-corecteaza--desktop-sterge-confirmare.dc.html", DailyDeleteEffect(DailyDelete{
			Author: "Dr. Ceban", At: at0807, Corrected: &Corrected{By: "Ion Rusu", On: today}, Sections: 4, F: -7,
			Farm: "Hîrbovăț", AliveF: &AliveMove{12561, 12568},
		})},
		{"S16-corecteaza--motiv-lipsa.dc.html", CheckedDays(2)},
	} {
		b, err := os.ReadFile("../../../../design/canvas/project/" + tc.board)
		if err != nil {
			t.Fatal(err)
		}
		// boardText collapses white space, U+202F included (the boards write it).
		if want := strings.ReplaceAll(tc.want, "\u202f", " "); !strings.Contains(boardText(string(b)), want) {
			t.Errorf("%q is not in %s", tc.want, tc.board)
		}
	}
}

// TestM5Plurals: the counted nouns of the M5 copy, with "de" from 20 and
// thousands grouped.
func TestM5Plurals(t *testing.T) {
	for _, tc := range []struct{ got, want string }{
		{sectionsWord(1), "1 secțiune"}, {sectionsWord(4), "4 secțiuni"}, {sectionsWord(19), "19 secțiuni"}, {sectionsWord(20), "20 de secțiuni"},
		{hallsWord(1), "1 hală"}, {hallsWord(2), "2 hale"}, {hallsWord(20), "20 de hale"}, {hallsWord(101), "101 hale"},
		{valuesWord(1), "1 valoare"}, {valuesWord(3), "3 valori"}, {valuesWord(20), "20 de valori"},
		{birdsWith(1, "F", "moartă", "moarte", "mort", "morți"), "1 găină moartă"},
		{birdsWith(4, "F", "moartă", "moarte", "mort", "morți"), "4 găini moarte"},
		{birdsWith(1, "M", "moartă", "moarte", "mort", "morți"), "1 cocoș mort"},
		{birdsWith(20, "M", "moartă", "moarte", "mort", "morți"), "20 de cocoși morți"},
		{birdsWith(0, "M", "moartă", "moarte", "mort", "morți"), "0 cocoși morți"},
		{birds(-1, "F"), "−1 găină"}, {birds(-8, "F"), "−8 găini"}, {birds(-20, "M"), "−20 de cocoși"},
		{birds(-119, "M"), "−119 cocoși"}, {birds(-1200, "F"), "−1 200 de găini"}, {birds(-1201, "F"), "−1 201 găini"},
		{aliveArticle(4283, "F"), "cele 4 283 găini"}, {aliveArticle(323, "M"), "cei 323 cocoși"},
		{aliveArticle(1, "F"), "cea 1 găină"}, {aliveArticle(1, "M"), "cel 1 cocoș"}, {aliveArticle(20, "M"), "cei 20 cocoși"},
	} {
		if tc.got != tc.want {
			t.Errorf("got %q, want %q", tc.got, tc.want)
		}
	}
}

// TestM5Panics: an unknown sex, field, reason or action is a programming
// error.
func TestM5Panics(t *testing.T) {
	today := timex.MustDay(2026, 11, 18)
	for name, f := range map[string]func(){
		"MortalitySpike sex":     func() { MortalitySpike("H1", "X", 1, dec("0"), dec("3")) },
		"NoBirdsOnDay sex":       func() { NoBirdsOnDay("H1", "X", today) },
		"AuditKeyDaily field":    func() { AuditKeyDaily("H1", "F", DailyField("g")) },
		"AuditKeyEggs field":     func() { AuditKeyEggs("H1", EggField("total")) },
		"AuditKeyFeed field":     func() { AuditKeyFeed("H1", "F", FeedField("kg")) },
		"AuditKeyDelivery field": func() { AuditKeyDelivery(DeliveryField("pricePerKg")) },
		"BirdLabel field":        func() { BirdLabel("H1", "F", DailyField("g")) },
		"FeedLabel field":        func() { FeedLabel("H1", "F", FeedField("kg")) },
		"DeliveryLabel field":    func() { DeliveryLabel(DeliveryField("totalLei")) },
		"DayBoundExceeded field": func() { DayBoundExceeded("H1", "F", 2, 1, today, FieldCullReason, 1) },
		"CullReasonWord":         func() { CullReasonWord("boală") },
		"NegativeBanner action":  func() { NegativeBanner(CountAction(9), "H1", "F", today, -1, nil, today) },
	} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("no panic")
				}
			}()
			f()
		})
	}
}

// TestAuditKeyHensDeadIsDaily: the M1 key function and the M5 one agree.
func TestAuditKeyHensDeadIsDaily(t *testing.T) {
	if a, b := AuditKeyHensDead("H5.2"), AuditKeyDaily("H5.2", "F", FieldMortality); a != b {
		t.Errorf("%q != %q", a, b)
	}
}
