// Package messages holds every Romanian string the server returns, as typed
// functions (SPEC §7.0b). Strings are verbatim from SPEC §6–§8, with ș and ț
// in their comma-below forms (U+0219, U+021B), never the cedilla forms.
//
// A string marked "(proposed)" has no board yet; SPEC.md marks it the same way.
package messages

import (
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/fmtx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Sign-in (SPEC §6.16).

// LoginRequired is the 400 for an empty login.
func LoginRequired() string { return "Scrie loginul." }

// PasswordRequired is the 400 for an empty password.
func PasswordRequired() string { return "Scrie parola." }

// LoginFailed is the 401 for an unknown login or a wrong password (the same
// sentence for both, so logins can't be probed).
func LoginFailed() string { return "Login sau parolă greșită." }

// LoginRateLimited is the 429 after too many failed attempts.
func LoginRateLimited() string { return "Prea multe încercări. Încearcă din nou peste 5 minute." }

// AccountDisabled is the 403 for a deactivated account (after a correct password).
func AccountDisabled() string { return "Contul tău este dezactivat. Vorbește cu administratorul." }

// Envelope codes (SPEC §7.0, §7.0a).

// PageForbidden is the 403 FORBIDDEN of a route the caller's role may not use
// (auth.RequireRole). A screen's own sentence (e.g. "Nu ai acces la Serii.")
// wins where SPEC gives one. (proposed)
func PageForbidden() string { return "Nu ai acces la această pagină." }

// FarmForbidden is the 403 FARM_FORBIDDEN for a farm outside the caller's farms.
func FarmForbidden() string { return "Nu ai acces la această fermă." }

// NotFound is the 404 NOT_FOUND of an /api path no route serves. (proposed)
func NotFound() string { return "Adresa cerută nu există." }

// MethodNotAllowed is the 405 METHOD_NOT_ALLOWED of an /api route called with
// a method it doesn't serve. (proposed)
func MethodNotAllowed() string { return "Operația nu este permisă la această adresă." }

// Internal is the 500 INTERNAL for anything unexpected. (proposed)
func Internal() string { return "A apărut o eroare. Încearcă din nou." }

// PayloadTooLarge is the 413 PAYLOAD_TOO_LARGE for a body over the route's limit. (proposed)
func PayloadTooLarge() string { return "Datele trimise sunt prea mari." }

// Unauthenticated is the 401 UNAUTHENTICATED for a missing, expired or revoked session. (proposed)
func Unauthenticated() string { return "Sesiunea a expirat. Intră din nou în cont." }

// Forbidden is the 403 FORBIDDEN for an Origin/CSRF/Content-Type refusal (§11.4). (proposed)
func Forbidden() string {
	return "Cererea a fost refuzată. Reîncarcă pagina și încearcă din nou."
}

// InvalidData is the generic VALIDATION issue message (malformed JSON, unknown
// field, wrong type, required) where §6 has no specific sentence. (proposed)
func InvalidData() string { return "Datele trimise nu sunt valide." }

// Unavailable is the 503 UNAVAILABLE when a deadline runs out. (proposed)
func Unavailable() string { return "Serviciul nu răspunde acum. Încearcă din nou." }

// History migration (SPEC §10.3). These texts are stored in the migrated rows.

// RegisterDifference is the note of every migrated `adjustment` row (glossary
// "diferență registru", D63).
func RegisterDifference() string { return "diferență registru" }

// FeedStockMove is the note of both rows of a migrated feed stock move
// between the ♀ and the ♂ stock (feed_stock_adjustment). (proposed)
func FeedStockMove() string { return "mutare ♀ → ♂" }

// TransferDeparture is the note of a migrated transfer entry whose birds left
// the source on an earlier day than they arrived; day is dd.mm.yyyy.
func TransferDeparture(day string) string { return "plecare " + day }

// Audit (SPEC §4.10).

// AuditKeyHensDead is the `changes[].key` of a corrected ♀ mortality value of
// a section ("H5.2 · găini moarte", the §4.10 example).
func AuditKeyHensDead(section string) string { return section + " · găini moarte" }

// Periods (SPEC §5.2, D52).

// PeriodCurrentCycle is the label of period `cycle` on the open cycle.
func PeriodCurrentCycle() string { return "Ciclul curent" }

// PeriodWholeCycle is the label of period `cycle` on a past cycle (D52).
func PeriodWholeCycle() string { return "Tot ciclul" }

// PeriodWeeks is the label of period `weeks:n`: "4 săpt.".
func PeriodWeeks(n int) string { return strconv.Itoa(n) + " săpt." }

// PeriodInterval is the label of period `interval:from,to`, the §11.3 range:
// "01.11–15.11.2026". (proposed)
func PeriodInterval(from, to timex.Day) string { return fmtx.Range(from, to) }

// IntervalInFuture is the 400 for an interval that ends after today:
// "Până la 20.11.2026 e în viitor. Alege o dată până la 18.11.2026."
func IntervalInFuture(to, today timex.Day) string {
	return "Până la " + fmtx.Day(to) + " e în viitor. Alege o dată până la " + fmtx.Day(today) + "."
}

// PeriodTable (SPEC §7.4).

// ScopeFarmTotal is the scope label of the farm view: "Total fermă".
func ScopeFarmTotal() string { return "Total fermă" }

// WeekRow is a week row's label: "SAPT 58 · 10.11–16.11" for a whole week,
// "SAPT 59 · 17.11–18.11 · 2 din 7 zile" for a partial one, and
// "SAPT 59 · 17.11 · 1 din 7 zile" for a one-day row (proposed).
func WeekRow(sapt int, from, to timex.Day) string {
	days := to.DaysSince(from) + 1
	span := fmtx.RangeNoYear(from, to)
	if days == 1 {
		span = fmtx.DayMonth(from)
	}
	label := fmtx.WeekLabel(sapt) + " · " + span
	if days < 7 {
		label += " · " + strconv.Itoa(days) + " din 7 zile"
	}
	return label
}

// TotalsLabel is the foot row's label over the loaded weeks: "Total SAPT 56–59"
// ("Total SAPT 58" for one week, proposed).
func TotalsLabel(saptFrom, saptTo int) string { return "Total " + fmtx.WeekRange(saptFrom, saptTo) }

// LoadMore is the load-more row (D190, D192): next = the weeks the next press
// loads, then the loaded span and how many of the cycle's weeks are loaded:
// "Încarcă încă 4 săptămâni · SAPT 56–59 · 4 din 40". One week: "Încarcă încă
// o săptămână · …" (proposed); 20 or more: "Încarcă încă 20 de săptămâni · …"
// (proposed).
func LoadMore(next, loadedFrom, loadedTo, loadedWeeks, totalWeeks int) string {
	return "Încarcă încă " + weeksWord(next) + " · " + fmtx.WeekRange(loadedFrom, loadedTo) + " · " +
		strconv.Itoa(loadedWeeks) + " din " + strconv.Itoa(totalWeeks)
}

// weeksWord is n weeks in Romanian: "o săptămână", "4 săptămâni", and "de"
// before the noun when the number ends in 00 or 20–99 ("20 de săptămâni",
// "101 săptămâni").
func weeksWord(n int) string { return countNoun(n, "o săptămână", "săptămâni") }

// countNoun is n with its plural noun, or one for n = 1, with "de" before the
// noun when the number ends in 00 or 20–99: countNoun(8, "1 rând", "rânduri")
// = "8 rânduri", countNoun(20, …) = "20 de rânduri", countNoun(101, …) =
// "101 rânduri".
func countNoun(n int, one, many string) string {
	if n == 1 {
		return one
	}
	if r := n % 100; n >= 20 && (r == 0 || r >= 20) {
		return strconv.Itoa(n) + " de " + many
	}
	return strconv.Itoa(n) + " " + many
}

// CycleStartNote replaces the load-more row once the cycle's first week is
// loaded: "Începutul ciclului · SAPT 20".
func CycleStartNote(sapt int) string { return "Începutul ciclului · " + fmtx.WeekLabel(sapt) }

// EmptyNext is the next step of a period with no data: "Alege ciclul curent".
func EmptyNext() string { return "Alege ciclul curent" }

// NoteDeathsIncludeCulls is the Rezumat table note (§7.4; M2 Q6 d).
func NoteDeathsIncludeCulls() string { return "Morți include sacrificările, ca în registru." }

// weekdays are the day-row weekday labels of §5.1, Sunday first like time.Weekday.
var weekdays = [7]string{"Du", "Lu", "Ma", "Mi", "Jo", "Vi", "Sâ"}

// Weekday is a day row's weekday label: Lu Ma Mi Jo Vi Sâ Du.
func Weekday(d time.Weekday) string { return weekdays[d] }

// PeriodTable group and leaf labels (§7.4). Growth Rezumat labels are
// (proposed): §7.4 names the groups but gives no example.

// GroupAlive is the "vii" group: "În viață".
func GroupAlive() string { return "În viață" }

// GroupDeaths is the "morti" group (mortality + culls): "Morți".
func GroupDeaths() string { return "Morți" }

// GroupLay is the "ouat" group: "Ouat".
func GroupLay() string { return "Ouat" }

// GroupFeedKg is the "hranaKg" group: "Hrană kg".
func GroupFeedKg() string { return "Hrană kg" }

// GroupSale is the "vanzare" group: "Vânzare".
func GroupSale() string { return "Vânzare" }

// GroupFeedG is the growth Rezumat group g/pasăre ♀ · ♂ (efectiv · ghid). (proposed)
func GroupFeedG() string { return "g/pasăre" }

// GroupWeightF is the growth Rezumat group of the ♀ weighings. (proposed)
func GroupWeightF() string { return "Greutate ♀" }

// LeafHens is the ♀ leaf.
func LeafHens() string { return "♀" }

// LeafRoosters is the ♂ leaf.
func LeafRoosters() string { return "♂" }

// LeafEggs is the egg count leaf: "ouă".
func LeafEggs() string { return "ouă" }

// LeafLayPct is the lay % leaf: "% ouat".
func LeafLayPct() string { return "% ouat" }

// LeafGuide is a guide leaf: "ghid".
func LeafGuide() string { return "ghid" }

// LeafHatching is the hatching-egg count leaf: "incub.".
func LeafHatching() string { return "incub." }

// LeafHatchingPct is the hatching-egg % leaf: "% incub.".
func LeafHatchingPct() string { return "% incub." }

// LeafFloor is the floor-egg count leaf: "jos".
func LeafFloor() string { return "jos" }

// LeafFloorPct is the floor-egg % leaf: "% jos".
func LeafFloorPct() string { return "% jos" }

// LeafActual is an actual-value leaf next to its guide: "efectiv". (proposed)
func LeafActual() string { return "efectiv" }

// LeafWeightMean is the mean weighing leaf: "medie g". (proposed)
func LeafWeightMean() string { return "medie g" }

// LeafVsGuide is a difference-from-guide leaf: "față de ghid". (proposed)
func LeafVsGuide() string { return "față de ghid" }

// SexLeaf prefixes a leaf word with the sex sign: SexLeaf("F", LeafActual())
// = "♀ efectiv". It panics on another sex value (a programming error). (proposed)
func SexLeaf[S ~string](sex S, word string) string { return sexSign(sex) + " " + word }

// sexSign is the sex sign "♀" or "♂"; it panics on another sex value.
func sexSign[S ~string](sex S) string {
	switch sex {
	case "F":
		return LeafHens()
	case "M":
		return LeafRoosters()
	default:
		panic("messages: unknown sex " + string(sex))
	}
}

// Farm page (SPEC §7.3, M2 D5, Q4).

// CycleLabel is a CycleSelector line: "22.02.2026–azi · SAPT 20–59" for the
// open cycle (end nil), "07.03.2025–28.01.2026 · SAPT 19–65" for a closed one
// (the §11.3 range).
func CycleLabel(start timex.Day, end *timex.Day, saptFrom, saptTo int) string {
	dates := fmtx.Day(start) + fmtx.EnDash + "azi"
	if end != nil {
		dates = fmtx.Range(start, *end)
	}
	return dates + " · " + fmtx.WeekRange(saptFrom, saptTo)
}

// FarmEmpty is the empty-farm state: "Ferma e goală din 28.02.2026. Ultimul
// ciclu: Seria 2026, SAPT 1–21."
func FarmEmpty(since timex.Day, series string, saptFrom, saptTo int) string {
	return "Ferma e goală din " + fmtx.Day(since) + ". Ultimul ciclu: " + series + ", " +
		fmtx.WeekRange(saptFrom, saptTo) + "."
}

// FarmNoCycle is the state of a farm that never had a cycle. (proposed, M2 Q4 c)
func FarmNoCycle() string { return "Ferma nu are încă niciun ciclu." }

// CycleStartsHint is the line under an empty farm's state (S4 board).
func CycleStartsHint() string { return "Ciclul pornește la prima populare sau transfer." }

// Rezumat summary (SPEC §5.13, §7.6, M2 D6, Q6 c).

// ReadoutLay is the laying readout: "SAPT 59 · efectiv 40,4% · ghid 57,6% · −17,2 pct".
func ReadoutLay(sapt int, value, guide, diffPts decimal.Decimal) string {
	return fmtx.WeekLabel(sapt) + " · efectiv " + fmtx.Percent(value) + " · ghid " + fmtx.Percent(guide) +
		" · " + fmtx.Points(diffPts)
}

// ReadoutLayNoGuide is the laying readout without a guide figure: "SAPT 59 ·
// efectiv 40,4% · ghid – fără ghid" for a cycle without a guide (S4 board),
// "… · ghid – fără date în ghid" when the guide has no row (proposed).
func ReadoutLayNoGuide(sapt int, value decimal.Decimal, cycleHasGuide bool) string {
	return fmtx.WeekLabel(sapt) + " · efectiv " + fmtx.Percent(value) + " · ghid " + noGuide(cycleHasGuide)
}

// ReadoutGrowth is the growth readout, the latest ♀ weight vs the guide of
// SAPT − 1: "SAPT 21 · efectiv 2 553 g · ghid 2 370 g · +7,7%". (proposed, M2 Q6 c)
func ReadoutGrowth(sapt int, value, guide, diffPct decimal.Decimal) string {
	return fmtx.WeekLabel(sapt) + " · efectiv " + grams(value) + " · ghid " + grams(guide) + " · " +
		fmtx.PercentDiff(diffPct)
}

// ReadoutGrowthNoGuide is ReadoutGrowth without a guide figure: "SAPT 21 ·
// efectiv 2 553 g · ghid – fără ghid" / "… · ghid – fără date în ghid". (proposed)
func ReadoutGrowthNoGuide(sapt int, value decimal.Decimal, cycleHasGuide bool) string {
	return fmtx.WeekLabel(sapt) + " · efectiv " + grams(value) + " · ghid " + noGuide(cycleHasGuide)
}

func grams(g decimal.Decimal) string { return fmtx.Dec(g, 0) + " g" }

// noGuide is the "–" text of a missing guide figure (§11.3).
func noGuide(cycleHasGuide bool) string {
	if cycleHasGuide {
		return "– fără date în ghid"
	}
	return "– fără ghid"
}

// CumMortalityGuideNote is the note of the cumulative-mortality stat (§5.5).
func CumMortalityGuideNote() string { return "(ghid din intrarea în ouat)" }

// HallEmptySince names an empty hall in a stat note: "H7 goală din 14.11".
func HallEmptySince(hall string, since timex.Day) string {
	return hall + " goală din " + fmtx.DayMonth(since)
}

// EmptyHallTile is the one muted line of an empty hall among the section
// tiles: "goală din 14.11.2026 · H7.1, H7.2".
func EmptyHallTile(since timex.Day, sections []string) string {
	return "goală din " + fmtx.Day(since) + " · " + strings.Join(sections, ", ")
}

// Writes (SPEC §3.1, §3.2, §7.0; M3).

// FutureDate refuses a date after today, for every role (§3.2): "20.11.2026 e
// în viitor. Alege o dată până la 18.11.2026." A write answers it as 422
// REFUSED, a form-state GET as 400 VALIDATION on `date` (M3 D5, D8).
func FutureDate(date, today timex.Day) string {
	return fmtx.Day(date) + " e în viitor. Alege o dată până la " + fmtx.Day(today) + "."
}

// PastDayVet is the 403 PAST_DAY_VET of a vet writing a day other than today
// (Z1, H6), and the form state's lockedReason on a past day (§3.2).
func PastDayVet() string { return "Pentru zilele trecute cere managerului." }

// SeriesClosed is the 409 SERIES_CLOSED of a write dated inside a cycle of a
// closed series, and the form state's lockedReason there (§3.1): "Seria 2025
// e închisă: nu mai primește înregistrări sau corecții." (proposed)
func SeriesClosed(series string) string {
	return series + " e închisă: nu mai primește înregistrări sau corecții."
}

// Stale is the 409 STALE of a version mismatch or a clientRequestId reused by
// another user or on another route (§7.0). (proposed)
func Stale() string {
	return "Înregistrarea s-a schimbat între timp. Reîncarcă și încearcă din nou."
}

// EntryAlreadySaved is the 409 ALREADY_ENTERED of a clientRequestId whose
// stored reply was purged (§7.0): no value collides, the key's entry exists.
// (proposed: the Q3 alternative sentence)
func EntryAlreadySaved() string {
	return "Valorile acestei zile sunt deja introduse. Deschide înregistrarea ca să le modifici."
}

// Hrană zilnică (SPEC §6.3, M3 D5, D6).

// FeedAlreadyEntered is the 409 ALREADY_ENTERED of a Hrană zilnică row whose
// (section, day, sex) is live in another entry, named by the first colliding
// row; at is the owning entry's creation time: "H5.1 ♀ are deja hrana
// introdusă pe 18.11.2026 la 08:44. Apasă Editează ca să schimbi valorile."
// (proposed, M3 Q3)
func FeedAlreadyEntered[S ~string](section string, sex S, date timex.Day, at time.Time) string {
	return section + " " + sexSign(sex) + " are deja hrana introdusă pe " + fmtx.Day(date) + " la " +
		fmtx.TimeOfDay(at) + ". Apasă Editează ca să schimbi valorile."
}

// FeedGOutOfRange is H1, g outside 1–300: "H1.1 ♀: 350 g e peste limită.
// Scrie o valoare între 1 și 300 g." Below 1 it says "e sub limită"
// (proposed). g is shown as sent (up to 4 decimals).
func FeedGOutOfRange[S ~string](section string, sex S, g decimal.Decimal) string {
	side := "peste"
	if g.LessThan(decimal.NewFromInt(1)) {
		side = "sub"
	}
	return section + " " + sexSign(sex) + ": " + fmtx.DecUpTo(g, 4) + " g e " + side +
		" limită. Scrie o valoare între 1 și 300 g."
}

// FeedTypeWrongSex is H2 for a type whose for_sex is not the row's sex:
// "Cocoși nu e pentru găini. Alege alt tip." (proposed)
func FeedTypeWrongSex[S ~string](feedType string, sex S) string {
	return feedType + " nu e pentru " + fmtx.SexWord(sex) + ". Alege alt tip."
}

// FeedTypeRetired is H2 for an inactive type: "Breeder 1 e retras. Alege alt
// tip." (proposed, M3 Q4)
func FeedTypeRetired(feedType string) string { return feedType + " e retras. Alege alt tip." }

// NoBirdsAtDayStart is H3, a row for a section × sex without birds at the
// start of the day (also a hall or a retired section): "H7.1 nu are găini la
// începutul zilei 18.11.2026." (proposed)
func NoBirdsAtDayStart[S ~string](section string, sex S, date timex.Day) string {
	return section + " nu are " + fmtx.SexWord(sex) + " la începutul zilei " + fmtx.Day(date) + "."
}

// FeedGuideDeviation is the H4 warning: "H5.2 ♂: 173 g, cu 11% peste ghidul
// de 156 g. Verifică gramajul; se poate salva." ("sub ghidul" when below).
// diffPct = 100 × (g / guide − 1) at full precision; the text shows its
// absolute value as a whole percent, rounded half away from zero.
func FeedGuideDeviation[S ~string](section string, sex S, g, guide, diffPct decimal.Decimal) string {
	side := "peste"
	if diffPct.Sign() < 0 {
		side = "sub"
	}
	return section + " " + sexSign(sex) + ": " + fmtx.DecUpTo(g, 1) + " g, cu " + fmtx.Dec(diffPct.Abs(), 0) +
		"% " + side + " ghidul de " + fmtx.DecUpTo(guide, 1) + " g. Verifică gramajul; se poate salva."
}

// StockWouldGoNegative is the H5 STOCK_NEGATIVE warning with the lowest stock
// in whole kg: "Stocul de Breeder 2 la Hîrbovăț ar ajunge la −1 240 kg.
// Verifică livrările."
func StockWouldGoNegative(feedType, farm string, kg decimal.Decimal) string {
	return "Stocul de " + feedType + " la " + farm + " ar ajunge la " + fmtx.Dec(kg, 0) + " kg. Verifică livrările."
}

// FeedDailySaved is the save Toast: "Hrană zilnică salvată · 8 rânduri ·
// 2 457,9 kg · stoc Breeder 2: 31 665 kg" (kg 1 decimal, stock whole kg;
// "1 rând" and "20 de rânduri" proposed).
func FeedDailySaved(rows int, kg decimal.Decimal, stockType string, stock decimal.Decimal) string {
	return "Hrană zilnică salvată · " + rowsWord(rows) + " · " + fmtx.Dec(kg, 1) + " kg · stoc " + stockType + ": " +
		fmtx.Dec(stock, 0) + " kg"
}

// FeedDailySummary is the audit summary and journal name of a save: rows of
// the save out of the section × sex rows with birds that day: "Hrană zilnică ·
// 8 din 8 rânduri · 2 457,9 kg" (S4 board).
func FeedDailySummary(rows, of int, kg decimal.Decimal) string {
	return "Hrană zilnică · " + strconv.Itoa(rows) + " din " + rowsWord(of) + " · " + fmtx.Dec(kg, 1) + " kg"
}

func rowsWord(n int) string { return countNoun(n, "1 rând", "rânduri") }

// Stock and days left (SPEC §5.8, §7.12; M3 D7).

// StockBelowZero is the `abatere` reason of a stock below 0.
func StockBelowZero() string { return "stoc sub 0 · verifică livrările" }

// StockBelowDays is the `abatere` reason of days left under
// alert.stock_days_min: "sub pragul de 5 zile" ("1 zi", "20 de zile" proposed).
func StockBelowDays(minDays decimal.Decimal) string { return "sub pragul de " + daysWord(minDays) }

// StockNoConsumption is the `oprit` reason of a type with no consumption over
// the alert.stock_avg_days window: "– fără consum în ultimele 3 zile"
// ("… în ultima zi" for a 1-day window, proposed).
func StockNoConsumption(days int) string {
	if days == 1 {
		return "– fără consum în ultima zi"
	}
	return "– fără consum în ultimele " + countNoun(days, "1 zi", "zile")
}

// NoOpenCycle is the `oprit` reason of a farm without an open cycle.
func NoOpenCycle() string { return "fără ciclu deschis" }

// LastCycleClosed is the sub line of that farm's row: "ultimul ciclu: Seria
// 2026, închis pe 28.02.2026".
func LastCycleClosed(series string, closed timex.Day) string {
	return "ultimul ciclu: " + series + ", închis pe " + fmtx.Day(closed)
}

// daysWord is a day count: "5 zile", "1 zi", "20 de zile", "2,5 zile".
func daysWord(d decimal.Decimal) string {
	if d.IsInteger() && d.Sign() > 0 && d.LessThan(decimal.NewFromInt(1_000_000)) {
		return countNoun(int(d.IntPart()), "1 zi", "zile")
	}
	return fmtx.DecUpTo(d, 1) + " zile"
}

// Consum (SPEC §7.4, §7.12; M3 D11, Q2, Q8).

// GroupConsumKg is the Consum kg group: "Hrană · kg".
func GroupConsumKg() string { return "Hrană · kg" }

// GroupConsumG is the Consum g/bird group: "Hrană · g/pasăre". Its leaves are
// SexLeaf(sex, LeafActual() | LeafGuide() | LeafVsGuide()).
func GroupConsumG() string { return "Hrană · g/pasăre" }

// LeafState is the Consum state column: "stare".
func LeafState() string { return "stare" }

// NoteConsumG is the Consum foot note (§7.12) with tolerance.feed_pct:
// "g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷
// zile-pasăre (♀ după găini, ♂ după cocoși). Toleranță ±10%."
func NoteConsumG(tolerancePct decimal.Decimal) string {
	return "g/pasăre: zi = g din registru; săptămână, secțiune și perioadă = kg ÷ zile-pasăre (♀ după găini, " +
		"♂ după cocoși). Toleranță ±" + fmtx.DecUpTo(tolerancePct, 1) + "%."
}

// RecordedKgWeek is one week of the D163 note, newest first.
type RecordedKgWeek struct {
	Sapt  int
	Spans []RecordedKgSpan
}

// RecordedKgSpan is the halls of a week whose rows differ over the same days.
type RecordedKgSpan struct {
	Halls    []string
	From, To timex.Day
}

// NoteRecordedKg is the D163 Consum note: allocations whose register kg
// differs from g × birds, per week, hall group and days: "♀ 460
// kg/compartiment în registru (g × păsări ≈ 690 kg): SAPT 57: H5, H6
// 03.11–04.11 · H7 03.11–06.11; SAPT 56: H6 02.11." (proposed format, M3 Q8;
// kg whole).
func NoteRecordedKg[S ~string](sex S, recordedKg, computedKg decimal.Decimal, weeks []RecordedKgWeek) string {
	ws := make([]string, 0, len(weeks))
	for _, w := range weeks {
		spans := make([]string, 0, len(w.Spans))
		for _, sp := range w.Spans {
			days := fmtx.DayMonth(sp.From)
			if sp.To != sp.From {
				days = fmtx.RangeNoYear(sp.From, sp.To)
			}
			spans = append(spans, strings.Join(sp.Halls, ", ")+" "+days)
		}
		ws = append(ws, fmtx.WeekLabel(w.Sapt)+": "+strings.Join(spans, " · "))
	}
	return sexSign(sex) + " " + fmtx.Dec(recordedKg, 0) + " kg/compartiment în registru (g × păsări ≈ " +
		fmtx.Dec(computedKg, 0) + " kg): " + strings.Join(ws, "; ") + "."
}
