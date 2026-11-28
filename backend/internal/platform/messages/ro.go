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
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
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
// Verifică livrările." A stock below 0 never reads as 0 (COR-4, M4 D9): when
// whole kg round to 0 it shows one decimal ("−0,3 kg"), and at least −0,1.
func StockWouldGoNegative(feedType, farm string, kg decimal.Decimal) string {
	return "Stocul de " + feedType + " la " + farm + " ar ajunge la " + negativeKg(kg) + " kg. Verifică livrările."
}

// negativeKg is kg in whole kg, or with one decimal (at least −0,1) for a
// negative kg that whole-kg rounding would show as 0.
func negativeKg(kg decimal.Decimal) string {
	if kg.Sign() >= 0 || num.Round(kg, 0).Sign() != 0 {
		return fmtx.Dec(kg, 0)
	}
	if tenth := decimal.New(-1, -1); num.Round(kg, 1).GreaterThan(tenth) {
		kg = tenth
	}
	return fmtx.Dec(kg, 1)
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

// Livrare hrană (SPEC §6.7, S13; M4 D5, Q9, Q13, Q15, Q16).

// DeliveryQtyRequired is the 400 of a missing or non-positive qtyKg: "Scrie
// câte kg de Breeder 2 s-au livrat la Hîrbovăț."
func DeliveryQtyRequired(feedType, farm string) string {
	return "Scrie câte kg de " + feedType + " s-au livrat la " + farm + "."
}

// PriceForbidden is the 403 PRICE_FORBIDDEN of a manager's request with price
// fields, or a manager on a price write (§6.7, M4 Q10).
func PriceForbidden() string { return "Prețul îl adaugă proprietarul în Prețuri hrană." }

// DeliveryAdded is the save Toast, kg as typed (up to 1 decimal) and the stock
// of the type at the end of the delivery date in whole kg: "Livrare adăugată ·
// 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg · fără preț".
func DeliveryAdded(kg decimal.Decimal, feedType, farm string, stock decimal.Decimal) string {
	return "Livrare adăugată · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType + " · stoc " + farm + ": " +
		fmtx.Dec(stock, 0) + " kg · " + Unpriced()
}

// DeliverySummary is the audit summary and journal name of a delivery:
// "Livrare · 9 000 kg Breeder 2". (proposed, M4 Q15)
func DeliverySummary(kg decimal.Decimal, feedType string) string {
	return "Livrare · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType
}

// deliveriesWord is a delivery count: "1 livrare", "4 livrări", "20 de livrări".
func deliveriesWord(n int) string { return countNoun(n, "1 livrare", "livrări") }

// Prețuri hrană (SPEC §6.8, §7.13, S12; M4 Q10, Q11, Q15).

// PricesSaved is the S12 save Toast: "Prețuri salvate · 4 livrări · 18 160 kg
// intră în cost" (kg up to 1 decimal).
func PricesSaved(count int, kg decimal.Decimal) string {
	return "Prețuri salvate · " + deliveriesWord(count) + " · " + fmtx.DecUpTo(kg, 1) + " kg intră în cost"
}

// PricesSummary is the audit summary of a price batch: "Prețuri · 4 livrări ·
// 18 160 kg". (proposed, M4 Q10)
func PricesSummary(count int, kg decimal.Decimal) string {
	return "Prețuri · " + deliveriesWord(count) + " · " + fmtx.DecUpTo(kg, 1) + " kg"
}

// PricePerKgNotPositive is the 400 of a pricePerKg ≤ 0, naming the delivery by
// type and date: "Cocoși 17.02.2026: scrie un preț mai mare de 0 lei/kg".
func PricePerKgNotPositive(feedType string, date timex.Day) string {
	return feedType + " " + fmtx.Day(date) + ": scrie un preț mai mare de 0 lei/kg"
}

// TotalLeiNotPositive is the 400 of a totalLei ≤ 0: "Cocoși 17.02.2026: scrie
// un total mai mare de 0 lei". (proposed, M4 Q10)
func TotalLeiNotPositive(feedType string, date timex.Day) string {
	return feedType + " " + fmtx.Day(date) + ": scrie un total mai mare de 0 lei"
}

// PriceBothGiven is the 400 of an item with both pricePerKg and totalLei:
// "Cocoși 17.02.2026: scrie prețul pe kg sau totalul, nu amândouă." (proposed,
// M4 Q10)
func PriceBothGiven(feedType string, date timex.Day) string {
	return feedType + " " + fmtx.Day(date) + ": scrie prețul pe kg sau totalul, nu amândouă."
}

// AuditKeyPricePerKg is the `changes[].key` of a delivery's lei/kg: "Breeder 2
// · 12.11.2026 · lei/kg". (proposed, M4 Q10)
func AuditKeyPricePerKg(feedType string, date timex.Day) string {
	return feedType + " · " + fmtx.Day(date) + " · lei/kg"
}

// AuditKeyTotalLei is the `changes[].key` of a delivery's total: "Breeder 2 ·
// 12.11.2026 · total lei". (proposed, M4 Q10)
func AuditKeyTotalLei(feedType string, date timex.Day) string {
	return feedType + " · " + fmtx.Day(date) + " · total lei"
}

// UndoDayEnded is the 409 STALE of a price undo token used after 23:59 of the
// save day: "Nu se mai poate anula: ziua salvării s-a încheiat." (proposed,
// M4 Q11)
func UndoDayEnded() string { return "Nu se mai poate anula: ziua salvării s-a încheiat." }

// PricesUndone is the price undo Toast: "Prețuri anulate · 4 livrări înapoi
// fără preț". (proposed, M4 Q11)
func PricesUndone(count int) string {
	return "Prețuri anulate · " + deliveriesWord(count) + " înapoi " + Unpriced()
}

// Price change (SPEC §6.8, §6.15, S11 Modifică; M4 Q12).

// PriceChangeReasonRequired is the 400 of a price change without a Motiv, the
// prices in lei/kg with 2 decimals: "Breeder 2 13.11.2026: scrie motivul ca să
// schimbi prețul din 7,76 în 7,74 lei/kg".
func PriceChangeReasonRequired(feedType string, date timex.Day, oldPerKg, newPerKg decimal.Decimal) string {
	return feedType + " " + fmtx.Day(date) + ": scrie motivul ca să schimbi prețul din " + fmtx.Dec(oldPerKg, 2) +
		" în " + fmtx.Dec(newPerKg, 2) + " lei/kg"
}

// ReasonTooLong is the §6.15 400 of a trimmed Motiv over limit characters:
// "Motivul are 212 caractere; cel mult 200. Scurtează-l cu 12 ca să poți
// salva." ("220 de caractere" proposed.)
func ReasonTooLong(length, limit int) string {
	return "Motivul are " + countNoun(length, "1 caracter", "caractere") + "; cel mult " + strconv.Itoa(limit) +
		". Scurtează-l cu " + strconv.Itoa(length-limit) + " ca să poți salva."
}

// NothingChanged is the 422 of a correction or price change that changes no
// figure (§6.15). (proposed)
func NothingChanged() string { return "Nu ai schimbat nicio cifră." }

// PriceChanged is the price change Toast, lei/kg with 2 decimals and the total
// in whole lei: "Preț schimbat · Breeder 2 · 13.11.2026 · 7,74 lei/kg · 198 454
// lei". (proposed, M4 Q12)
func PriceChanged(feedType string, date timex.Day, perKg, totalLei decimal.Decimal) string {
	return "Preț schimbat · " + feedType + " · " + fmtx.Day(date) + " · " + fmtx.Dec(perKg, 2) + " lei/kg · " +
		fmtx.Dec(totalLei, 0) + " lei"
}

// PriceChangeSummary is the audit summary of a price change, the delivery by
// type and date: "Preț schimbat · Breeder 2 · 13.11.2026". The old and new
// prices and the Motiv are in the row's changes and reason, so the summary,
// like the batch's, names no price. (proposed, M4 T7)
func PriceChangeSummary(feedType string, date timex.Day) string {
	return "Preț schimbat · " + feedType + " · " + fmtx.Day(date)
}

// Rapoarte › Hrană: Livrări and Cost (SPEC §5.9, §7.12, S11; M4 D6, Q6, Q8).

// Unpriced is the tag of a delivery without a price: "fără preț".
func Unpriced() string { return "fără preț" }

// PricedFromRegister is `pricedBy` of a migrated price: "din registru" (D163).
func PricedFromRegister() string { return "din registru" }

// CostRowLabel is a Cost row's label: "Telița · Seria 2026".
func CostRowLabel(farm, series string) string { return farm + " · " + series }

// CostSpanClosed is the span of a closed cycle's Cost row, the §11.3 range:
// "07.10.2025–28.02.2026".
func CostSpanClosed(from, to timex.Day) string { return fmtx.Range(from, to) }

// CostSpanOpen is the span of the open cycle's Cost row: "din 22.02.2026 ·
// SAPT 59".
func CostSpanOpen(start timex.Day, saptToday int) string {
	return "din " + fmtx.Day(start) + " · " + fmtx.WeekLabel(saptToday)
}

// UnpricedSpan is the first–last date of a cycle's unpriced deliveries:
// "17.02–12.11.2026"; one day alone is "17.02.2026" (proposed).
func UnpricedSpan(first, last timex.Day) string {
	if first == last {
		return fmtx.Day(first)
	}
	return fmtx.Range(first, last)
}

// CostNoEggsGrowth is the `eggReason` of a growth cycle's lei/ou: "creștere,
// fără ouă".
func CostNoEggsGrowth() string { return "creștere, fără ouă" }

// CostNoEggs is the `eggReason` of a laying cycle (or a series total) without
// eggs in the period yet: "fără ouă în perioadă". (proposed, M4 T3)
func CostNoEggs() string { return "fără ouă în perioadă" }

// CostSeriesTotal is the Cost total row's label: "Total Seria 2026".
func CostSeriesTotal(series string) string { return "Total " + series }

// Zilnic (SPEC §6.2, S6; M5 Q3, Q5, Q15).

// DailyField is a bird field of a Zilnic section × sex, named as in the JSON
// body (§7.8 DailySex).
type DailyField string

// The Zilnic bird fields.
const (
	FieldMortality   DailyField = "mortality"
	FieldCull        DailyField = "cull"
	FieldCullReason  DailyField = "cullReason"
	FieldSexingError DailyField = "sexingError"
)

// EggField is a field of a Zilnic egg row, named as in the JSON body (§7.8
// DailyEggs).
type EggField string

// The Zilnic egg fields.
const (
	FieldFloor      EggField = "floor"
	FieldCat1       EggField = "cat1"
	FieldCat2       EggField = "cat2"
	FieldCat3       EggField = "cat3"
	FieldHatching   EggField = "hatching"
	FieldAvgWeightG EggField = "avgWeightG"
)

// FeedField is a field of a Hrană zilnică row, named as in the JSON body.
type FeedField string

// The Hrană zilnică fields.
const (
	FieldG          FeedField = "g"
	FieldFeedTypeID FeedField = "feedTypeId"
)

// DeliveryField is an editable field of a Livrare hrană, named as in the JSON
// body (M5 Q11).
type DeliveryField string

// The editable Livrare hrană fields.
const (
	FieldQtyKg              DeliveryField = "qtyKg"
	FieldDeliveryFeedTypeID DeliveryField = "feedTypeId"
	FieldNote               DeliveryField = "note"
)

// NotAnInteger is Z2, a count that is not an integer ≥ 0 (a sexing error
// that is not an integer): "H5.1 ♀: scrie un număr întreg." (proposed)
func NotAnInteger[S ~string](section string, sex S) string {
	return section + " " + sexSign(sex) + ": scrie un număr întreg."
}

// EggCountNotAnInteger is Z8 for an egg count that is not an integer ≥ 0:
// "H5: scrie un număr întreg." (proposed, the Z2 form)
func EggCountNotAnInteger(hall string) string { return hall + ": scrie un număr întreg." }

// NoBirdsOnDay is Z3, a value for a section × sex without birds that day (also
// a hall or a retired section), and eggs on a unit without hens at the start
// of the day: "H7.1 nu are găini pe 18.11.2026." (proposed)
func NoBirdsOnDay[S ~string](section string, sex S, date timex.Day) string {
	return section + " nu are " + fmtx.SexWord(sex) + " pe " + fmtx.Day(date) + "."
}

// CullReasonMissing is Z4, a cull without its reason: "H5.1 ♀: Motiv lipsă
// pentru 2 găini sacrificate pe 18.11.2026. Alege motivul: boală, șchiopătare
// sau altul." ("1 găină sacrificată", "1 cocoș sacrificat", "20 de cocoși
// sacrificați" proposed.)
func CullReasonMissing[S ~string](section string, sex S, culls int, date timex.Day) string {
	return section + " " + sexSign(sex) + ": Motiv lipsă pentru " + birdsWith(culls, sex, "sacrificată", "sacrificate", "sacrificat", "sacrificați") +
		" pe " + fmtx.Day(date) + ". Alege motivul: boală, șchiopătare sau altul."
}

// DayBoundExceeded is Z5, the §6.1 per-day bound: total = Mortalitate +
// Sacrificare + Eroare sexare (removals), available = the birds alive at the
// start of the day plus that day's inflows, field = the field the bound names
// and most its largest value: "H5.2 ♂: Mortalitate + Sacrificare + Eroare
// sexare = 3 230, peste cei 323 cocoși în viață pe 18.11.2026. La Mortalitate
// scrie cel mult 323." The bird count has no "de" (as SPEC writes "cei 323
// cocoși"); one bird is "cea 1 găină" / "cel 1 cocoș" (proposed).
func DayBoundExceeded[S ~string](section string, sex S, total, available int, date timex.Day, field DailyField, most int) string {
	return section + " " + sexSign(sex) + ": Mortalitate + Sacrificare + Eroare sexare = " + fmtx.Int(int64(total)) +
		", peste " + aliveArticle(available, sex) + " în viață pe " + fmtx.Day(date) + ". La " + dailyFieldName(field) +
		" scrie cel mult " + fmtx.Int(int64(most)) + "."
}

// DayBoundBanner is the Z5 SaveBar text: "refuzat · H5.2 ♂ 3 230 · cel mult 323".
func DayBoundBanner[S ~string](section string, sex S, total, most int) string {
	return "refuzat · " + section + " " + sexSign(sex) + " " + fmtx.Int(int64(total)) + " · cel mult " + fmtx.Int(int64(most))
}

// MortalitySpike is the Z6 MORTALITY_SPIKE warning: deaths typed, mean7 the
// 7-day mean (shown with one decimal) and factor alert.mortality_factor:
// "H6.1: 4 găini moarte, de peste 3 ori media pe 7 zile (0,9). Verifică
// numărătoarea; poți salva." ("1 găină moartă", "1 cocoș mort", "20 de cocoși
// morți" proposed.)
func MortalitySpike[S ~string](section string, sex S, deaths int, mean7, factor decimal.Decimal) string {
	return section + ": " + birdsWith(deaths, sex, "moartă", "moarte", "mort", "morți") + ", de peste " +
		fmtx.DecUpTo(factor, 1) + " ori media pe 7 zile (" + fmtx.Dec(mean7, 1) + "). Verifică numărătoarea; poți salva."
}

// EggsOnHall is Z7, eggs on a section that isn't its egg unit: "Ouăle se scriu
// pe hală: H5." (proposed)
func EggsOnHall(hall string) string { return "Ouăle se scriu pe hală: " + hall + "." }

// EggsGrowthFarm is Z7 at a growth farm: "Telița e fermă de creștere; ouăle se
// scriu doar la fermele de ouat." (proposed)
func EggsGrowthFarm(farm string) string {
	return farm + " e fermă de creștere; ouăle se scriu doar la fermele de ouat."
}

// EggWeightOutOfRange is Z8, Masa medie outside 20–120 g, shown as sent (up to
// 2 decimals): "H5: masa medie 130 g e în afara limitei. Scrie o valoare între
// 20 și 120 g." (proposed)
func EggWeightOutOfRange(hall string, g decimal.Decimal) string {
	return hall + ": masa medie " + fmtx.DecUpTo(g, 2) + " g e în afara limitei. Scrie o valoare între 20 și 120 g."
}

// DailyAlreadyEntered is Z9 for a bird value whose key is live in another
// entry, created at at: "H5.1 ♀ are deja valori introduse pe 18.11.2026 la
// 08:15. Apasă Editează ca să schimbi valorile." (proposed)
func DailyAlreadyEntered[S ~string](section string, sex S, date timex.Day, at time.Time) string {
	return section + " " + sexSign(sex) + " are deja valori introduse pe " + fmtx.Day(date) + " la " +
		fmtx.TimeOfDay(at) + ". Apasă Editează ca să schimbi valorile."
}

// EggsAlreadyEntered is Z9 for an egg row: "H5 are deja ouăle introduse pe
// 18.11.2026 la 08:15. Apasă Editează ca să schimbi valorile." (proposed)
func EggsAlreadyEntered(hall string, date timex.Day, at time.Time) string {
	return hall + " are deja ouăle introduse pe " + fmtx.Day(date) + " la " + fmtx.TimeOfDay(at) +
		". Apasă Editează ca să schimbi valorile."
}

// DailySaved is the Zilnic save Toast with bird values: sections with a bird
// value and Σ signed qty per sex: "Zilnic salvat · 4 secțiuni · ♀ −3 · ♂ 0".
func DailySaved(sections, f, m int) string {
	return "Zilnic salvat · " + sectionsWord(sections) + " · " + LeafHens() + " " + signedInt(f) + " · " +
		LeafRoosters() + " " + signedInt(m)
}

// DailyEggsSaved is the Toast of an egg-only Zilnic save: "Zilnic salvat · 2
// hale · 5 914 ouă". (proposed)
func DailyEggsSaved(halls, eggs int) string {
	return "Zilnic salvat · " + hallsWord(halls) + " · " + fmtx.Int(int64(eggs)) + " ouă"
}

// DailySummary is the audit summary and Jurnal name of a Zilnic save:
// "Mortalitate · 4 secțiuni" (D178), "Ouă · 2 hale" (proposed) or both,
// "Mortalitate · 4 secțiuni · Ouă · 2 hale" (proposed). A zero count leaves
// its part out.
func DailySummary(sections, halls int) string {
	var parts []string
	if sections > 0 {
		parts = append(parts, "Mortalitate · "+sectionsWord(sections))
	}
	if halls > 0 {
		parts = append(parts, "Ouă · "+hallsWord(halls))
	}
	return strings.Join(parts, " · ")
}

// MortalitySectionSummary is the audit summary of a per-section delete of a
// section's Zilnic bird values: "Mortalitate · H5.2". (proposed)
func MortalitySectionSummary(section string) string { return "Mortalitate · " + section }

// EggsHallSummary is the audit summary of a per-hall delete of its eggs: "Ouă
// · H5". (proposed)
func EggsHallSummary(hall string) string { return "Ouă · " + hall }

// Counts never below 0 (SPEC §6.1; M5 D4, Q4).

// CountAction is the write whose §6.1 check failed; it names the banner's
// subject.
type CountAction int

// The §6.1 banner subjects (Q4).
const (
	ActionSave       CountAction = iota // a create or a same-day author edit: "Salvarea" (proposed)
	ActionCorrection                    // a correction: "Corecția"
	ActionDelete                        // a delete or an undo: "Ștergerea" (proposed)
	ActionRestore                       // a restore: "Restaurarea" (proposed)
)

func (a CountAction) subject() string {
	switch a {
	case ActionSave:
		return "Salvarea"
	case ActionCorrection:
		return "Corecția"
	case ActionDelete:
		return "Ștergerea"
	case ActionRestore:
		return "Restaurarea"
	default:
		panic("messages: unknown count action " + strconv.Itoa(int(a)))
	}
}

// NegativeBanner is the §6.1 banner (error.message) of the first section × sex
// a write would take below 0: first = its first negative day and
// firstAliveEnd the birds at its end; todayAliveEnd = the birds at the end of
// today while still negative today, else nil.
//   - negative on earlier days only: "Corecția ar lăsa H6 cu −4 găini pe
//     16.02.2026. Nicio secțiune nu poate scădea sub 0."
//   - from an earlier day to today: "Corecția ar lăsa H5.2 cu −6 găini pe
//     17.11.2026 și cu −8 găini azi, 18.11.2026. Nicio secțiune nu poate
//     scădea sub 0."
//   - today only: "Corecția ar lăsa H5.2 cu −8 găini azi, 18.11.2026. Nicio
//     secțiune nu poate scădea sub 0."
func NegativeBanner[S ~string](a CountAction, section string, sex S, first timex.Day, firstAliveEnd int,
	todayAliveEnd *int, today timex.Day) string {
	var when string
	switch {
	case first == today:
		n := firstAliveEnd
		if todayAliveEnd != nil {
			n = *todayAliveEnd
		}
		when = "cu " + birds(n, sex) + " " + todayText(today)
	case todayAliveEnd == nil:
		when = "cu " + birds(firstAliveEnd, sex) + " pe " + fmtx.Day(first)
	default:
		when = "cu " + birds(firstAliveEnd, sex) + " pe " + fmtx.Day(first) + " și cu " + birds(*todayAliveEnd, sex) +
			" " + todayText(today)
	}
	return a.subject() + " ar lăsa " + section + " " + when + ". Nicio secțiune nu poate scădea sub 0."
}

// RemovalBound is the §6.1 field message of a removal raised too far: typed =
// the typed value, last = the last negative day, minAliveEnd the lowest end of
// day and most = typed + minAliveEnd: "H5.2 ♀ 4 292: ar lăsa −8 găini azi,
// 18.11.2026. Scrie cel mult 4 284." ("… pe 16.02.2026. …" when last is not
// today.)
func RemovalBound[S ~string](section string, sex S, typed int, last timex.Day, minAliveEnd int, today timex.Day, most int) string {
	return boundHead(section, sex, typed, last, minAliveEnd, today) + " Scrie cel mult " + fmtx.Int(int64(most)) + "."
}

// InflowBound is the §6.1 field message of an inflow lowered too far, least =
// typed − minAliveEnd: "H5.2 ♂ 4: ar lăsa −3 cocoși azi, 18.11.2026. Scrie cel
// puțin 7." (proposed, the Corectează removal form)
func InflowBound[S ~string](section string, sex S, typed int, last timex.Day, minAliveEnd int, today timex.Day, least int) string {
	return boundHead(section, sex, typed, last, minAliveEnd, today) + " Scrie cel puțin " + fmtx.Int(int64(least)) + "."
}

func boundHead[S ~string](section string, sex S, typed int, last timex.Day, minAliveEnd int, today timex.Day) string {
	day := "pe " + fmtx.Day(last)
	if last == today {
		day = todayText(today)
	}
	return section + " " + sexSign(sex) + " " + fmtx.Int(int64(typed)) + ": ar lăsa " + birds(minAliveEnd, sex) + " " + day + "."
}

func todayText(today timex.Day) string { return "azi, " + fmtx.Day(today) }

// Edit, Corectează, delete, undo, restore (SPEC §6.15, §7.8, S15, S16; M5 Q7–Q9, Q12, Q15).

// OwnEntryPastDay is the 403 FORBIDDEN of a vet changing or deleting their own
// entry after its creation day (S16 board).
func OwnEntryPastDay() string {
	return "Poți modifica sau șterge ce ai introdus doar în aceeași zi, până la 23:59. Pentru o corecție, spune-i " +
		"managerului; el o salvează cu motiv."
}

// AuditKeyDaily is the `changes[].key` of a Zilnic bird value: "H5.2 · găini
// moarte" (§4.10), "H5.2 · cocoși morți", "H5.2 · găini sacrificate", "H5.2 ·
// cocoși sacrificați", "H5.2 · motiv sacrificare găini", "H5.2 · eroare sexare
// cocoși". (proposed, M5 Q12, except the §4.10 example)
func AuditKeyDaily[S ~string](section string, sex S, field DailyField) string {
	female := sexIsF(sex)
	var what string
	switch field {
	case FieldMortality:
		what = pick(female, "găini moarte", "cocoși morți")
	case FieldCull:
		what = pick(female, "găini sacrificate", "cocoși sacrificați")
	case FieldCullReason:
		what = "motiv sacrificare " + fmtx.SexWord(sex)
	case FieldSexingError:
		what = "eroare sexare " + fmtx.SexWord(sex)
	default:
		panic("messages: unknown daily field " + string(field))
	}
	return section + " · " + what
}

// AuditKeyEggs is the `changes[].key` of a Zilnic egg value: "H5 · ouă jos",
// "H5 · Cat. 1", "H5 · incubabile", "H5 · masa medie g". (proposed, M5 Q12)
func AuditKeyEggs(hall string, field EggField) string {
	w := eggFieldWord(field)
	if field == FieldAvgWeightG {
		w += " g"
	}
	return hall + " · " + w
}

// AuditKeyFeed is the `changes[].key` of a Hrană zilnică value: "H5.1 · g
// găini", "H5.1 · tip hrană cocoși". (proposed, M5 Q12)
func AuditKeyFeed[S ~string](section string, sex S, field FeedField) string {
	switch field {
	case FieldG:
		return section + " · g " + fmtx.SexWord(sex)
	case FieldFeedTypeID:
		return section + " · tip hrană " + fmtx.SexWord(sex)
	default:
		panic("messages: unknown feed field " + string(field))
	}
}

// AuditKeyDelivery is the `changes[].key` of a Livrare hrană value: "Livrare ·
// kg", "Livrare · tip", "Livrare · notă". (proposed, M5 Q12)
func AuditKeyDelivery(field DeliveryField) string {
	switch field {
	case FieldQtyKg:
		return "Livrare · kg"
	case FieldDeliveryFeedTypeID:
		return "Livrare · tip"
	case FieldNote:
		return "Livrare · notă"
	default:
		panic("messages: unknown delivery field " + string(field))
	}
}

// Shown is a value as a sentence shows it (§11.3); build it with ShownInt,
// ShownDec, ShownText, ShownCullReason or ShownEmpty.
type Shown struct{ s string }

// ShownInt is a count: "4 292", "−2".
func ShownInt(n int) Shown { return Shown{fmtx.Int(int64(n))} }

// ShownDec is a decimal with at most places decimals: "160", "67,04".
func ShownDec(d decimal.Decimal, places int32) Shown { return Shown{fmtx.DecUpTo(d, places)} }

// ShownText is a name as stored: a feed type ("Breeder 2") or a note.
func ShownText(s string) Shown { return Shown{s} }

// ShownEmpty is an empty value (null): "–".
func ShownEmpty() Shown { return Shown{"–"} }

// ShownCullReason is a cull reason's word: boala → "boală", schiopatare →
// "șchiopătare", altul → "altul". It panics on another value.
func ShownCullReason[R ~string](reason R) Shown { return Shown{CullReasonWord(reason)} }

// CullReasonWord is a cull reason as shown and as audit `changes` text values
// carry it (M5 Q12): "boală", "șchiopătare", "altul". It panics on another
// value (reasons are checked at the boundary).
func CullReasonWord[R ~string](reason R) string {
	switch reason {
	case "boala":
		return "boală"
	case "schiopatare":
		return "șchiopătare"
	case "altul":
		return "altul"
	default:
		panic("messages: unknown cull reason " + string(reason))
	}
}

// Label names one value of a save in the correction Toasts and the Motiv
// sentence; build it with BirdLabel, EggLabel, FeedLabel or DeliveryLabel.
type Label struct {
	subject string // in a Toast: "H5.2 ♀", "H5 Cat. 1", "Livrare"
	motiv   string // in "Motiv lipsă. Scrie de ce se schimbă {motiv} din …": "H5.2 ♀", "livrarea"
	unit    string // after the new value: "g", "kg"; "" for counts and text
}

// BirdLabel is a Zilnic bird value: "H5.2 ♀" (mortality, §6.15), "H5.2 ♀
// sacrificare", "H5.2 ♀ motiv sacrificare", "H5.2 ♀ eroare sexare". (proposed
// except mortality)
func BirdLabel[S ~string](section string, sex S, field DailyField) Label {
	s := section + " " + sexSign(sex)
	switch field {
	case FieldMortality:
	case FieldCull:
		s += " sacrificare"
	case FieldCullReason:
		s += " motiv sacrificare"
	case FieldSexingError:
		s += " eroare sexare"
	default:
		panic("messages: unknown daily field " + string(field))
	}
	return Label{subject: s, motiv: s}
}

// EggLabel is a Zilnic egg value: "H5 Cat. 1", "H5 ouă jos", "H5 masa medie"
// (in g). (proposed)
func EggLabel(hall string, field EggField) Label {
	s := hall + " " + eggFieldWord(field)
	l := Label{subject: s, motiv: s}
	if field == FieldAvgWeightG {
		l.unit = "g"
	}
	return l
}

// FeedLabel is a Hrană zilnică value: "H5.1 ♀" (g) or "H5.1 ♀ tip hrană".
// (proposed)
func FeedLabel[S ~string](section string, sex S, field FeedField) Label {
	s := section + " " + sexSign(sex)
	switch field {
	case FieldG:
		return Label{subject: s, motiv: s, unit: "g"}
	case FieldFeedTypeID:
		s += " tip hrană"
		return Label{subject: s, motiv: s}
	default:
		panic("messages: unknown feed field " + string(field))
	}
}

// DeliveryLabel is a Livrare hrană value: kg ("Livrare" in a Toast, "livrarea"
// in the Motiv sentence), type, note. (proposed)
func DeliveryLabel(field DeliveryField) Label {
	switch field {
	case FieldQtyKg:
		return Label{subject: "Livrare", motiv: "livrarea", unit: "kg"}
	case FieldDeliveryFeedTypeID:
		return Label{subject: "Livrare tip", motiv: "tipul livrării"}
	case FieldNote:
		return Label{subject: "Livrare notă", motiv: "nota livrării"}
	default:
		panic("messages: unknown delivery field " + string(field))
	}
}

// ReasonMissingChange is the §6.15 400 on `reason` of a correction of one
// value: "Motiv lipsă. Scrie de ce se schimbă H5.2 ♀ din 1 în 3; fără motiv
// corecția nu se salvează." (eggs "… H5 Cat. 1 din 13 în 15; …", feed "… H5.1
// ♀ din 160 în 158 g; …", delivery "… livrarea din 9 000 în 8 000 kg; …"
// proposed).
func ReasonMissingChange(l Label, from, to Shown) string {
	return "Motiv lipsă. Scrie de ce se schimbă " + l.motiv + " din " + from.s + " în " + to.s + l.withUnit() +
		"; fără motiv corecția nu se salvează."
}

// ReasonMissingChanges is the §6.15 400 on `reason` of a correction of several
// values: "Motiv lipsă. Scrie de ce se schimbă 3 valori; fără motiv corecția
// nu se salvează." (proposed)
func ReasonMissingChanges(n int) string {
	return "Motiv lipsă. Scrie de ce se schimbă " + valuesWord(n) + "; fără motiv corecția nu se salvează."
}

// DeleteReasonMissing is the 400 on `reason` of a delete that needs a Motiv:
// "Motiv lipsă. Scrie de ce ștergi salvarea; fără motiv nu se șterge."
// (proposed, M5 Q9)
func DeleteReasonMissing() string {
	return "Motiv lipsă. Scrie de ce ștergi salvarea; fără motiv nu se șterge."
}

func (l Label) withUnit() string {
	if l.unit == "" {
		return ""
	}
	return " " + l.unit
}

func (l Label) change(from, to Shown) string {
	return l.subject + " " + from.s + " → " + to.s + l.withUnit()
}

// CorrectionSavedAlive is the Toast of a correction of one bird value, alive =
// the section's birds of sex at the end of today: "Corecție salvată · H5.2 ♀
// 1 → 3 · în viață ♀ 4 281".
func CorrectionSavedAlive[S ~string](l Label, from, to Shown, sex S, alive int) string {
	return "Corecție salvată · " + l.change(from, to) + " · în viață " + sexSign(sex) + " " + fmtx.Int(int64(alive))
}

// CorrectionSaved is the Toast of a correction of one value that moves no
// bird: "Corecție salvată · H5 Cat. 1 13 → 15". (proposed)
func CorrectionSaved(l Label, from, to Shown) string {
	return "Corecție salvată · " + l.change(from, to)
}

// CorrectionsSavedAlive is the Toast of a correction of several values that
// moves birds, with the farm alive at the end of today: "Corecție salvată · 3
// valori · în viață azi ♀ 12 561 · ♂ 794". (proposed)
func CorrectionsSavedAlive(n, aliveF, aliveM int) string {
	return "Corecție salvată · " + valuesWord(n) + " · în viață azi " + LeafHens() + " " + fmtx.Int(int64(aliveF)) +
		" · " + LeafRoosters() + " " + fmtx.Int(int64(aliveM))
}

// CorrectionsSaved is the Toast of a correction of several values that moves
// no bird: "Corecție salvată · 3 valori". (proposed)
func CorrectionsSaved(n int) string { return "Corecție salvată · " + valuesWord(n) }

// CorrectionUndone is the Toast of an undone correction of one value, from →
// to as the undo sets them: "Corecție anulată · H5.2 ♀ 3 → 1". (proposed)
func CorrectionUndone(l Label, from, to Shown) string {
	return "Corecție anulată · " + l.change(from, to)
}

// CorrectionsUndone is the Toast of an undone correction of several values:
// "Corecție anulată · 3 valori". (proposed)
func CorrectionsUndone(n int) string { return "Corecție anulată · " + valuesWord(n) }

// DeliverySaved is the Toast of a same-day author edit of a delivery, kg as
// typed and the type's stock at the end of the delivery date: "Livrare
// salvată · 9 000 kg Breeder 2 · stoc Hîrbovăț: 21 119 kg". (proposed)
func DeliverySaved(kg decimal.Decimal, feedType, farm string, stock decimal.Decimal) string {
	return "Livrare salvată · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType + " · stoc " + farm + ": " + fmtx.Dec(stock, 0) + " kg"
}

// EffectAliveFrom is a correction preview line per changed section × sex
// (§7.8): "H5.2 · găini în viață de pe 17.11.2026 se schimbă cu −2." ("+12"
// for a rise, §6.15).
func EffectAliveFrom[S ~string](section string, sex S, from timex.Day, delta int) string {
	return section + " · " + fmtx.SexWord(sex) + " în viață de pe " + fmtx.Day(from) + " se schimbă cu " + signedInt(delta) + "."
}

// EffectDeaths is a correction preview line per changed sex, the save's Σ
// signed bird qty before and after: "Mortalitate · morți ♀ −5 → −7" (§7.8).
func EffectDeaths[S ~string](sex S, before, after int) string {
	return "Mortalitate · morți " + sexSign(sex) + " " + signedInt(before) + " → " + signedInt(after)
}

// EffectFarmAliveToday is a correction preview line per changed sex, the
// farm's birds at the end of today: "Hîrbovăț · ♀ azi 12 563 → 12 561" (§7.8).
func EffectFarmAliveToday[S ~string](farm string, sex S, before, after int) string {
	return farm + " · " + sexSign(sex) + " azi " + fmtx.Int(int64(before)) + " → " + fmtx.Int(int64(after))
}

// CheckedDays is the S16 line under the preview: days from the entry date to
// today, none below 0: "Verificat pe 2 zile, până azi: nicio zi sub 0." ("1
// zi", "20 de zile" proposed.)
func CheckedDays(days int) string {
	return "Verificat pe " + countNoun(days, "1 zi", "zile") + ", până azi: nicio zi sub 0."
}

// AliveMove is a bird count before and after a change.
type AliveMove struct{ Before, After int }

// Corrected names who corrected a save and on which day, for the delete
// Dialog: "… și corectată de Ion Rusu pe 18.11.2026".
type Corrected struct {
	By string
	On timex.Day
}

// DailyDelete is what the Zilnic whole-save delete Dialog says.
type DailyDelete struct {
	Author    string     // "Dr. Ceban"
	At        time.Time  // the save's creation instant
	Corrected *Corrected // nil = never corrected
	Sections  int        // sections with a bird value (0 = egg-only save)
	F, M      int        // the save's Σ signed bird qty per sex: −5, 0
	Halls     int        // halls with an egg row
	Eggs      int        // Σ eggs of those rows
	Farm      string
	AliveF    *AliveMove // the farm's ♀ today; nil = unchanged, not named
	AliveM    *AliveMove // the farm's ♂ today; nil = unchanged, not named
}

// DailyDeleteEffect is the whole-save delete Dialog (dry-run effects[0]):
// "Se șterge salvarea Zilnic introdusă de Dr. Ceban la 08:07: 4 secțiuni · ♀
// −5 · ♂ 0. Găini în viață azi în Hîrbovăț: ♀ 12 563 → ♀ 12 568. Ștergerea
// rămâne în Istoric modificări." (§6.15); a corrected save adds "… la 08:07 și
// corectată de Ion Rusu pe 18.11.2026: …" (S16 board). (proposed: "Păsări în
// viață azi …: ♀ … · ♂ …" when ♂ changes too, "Cocoși …" when only ♂ does,
// "2 hale · 5 914 ouă" for the eggs.)
func DailyDeleteEffect(d DailyDelete) string {
	var what []string
	if d.Sections > 0 {
		what = append(what, sectionsWord(d.Sections)+" · "+sexTotals(d.F, d.M))
	}
	if d.Halls > 0 {
		what = append(what, hallsWord(d.Halls)+" · "+fmtx.Int(int64(d.Eggs))+" ouă")
	}
	s := "Se șterge salvarea Zilnic " + introduced(d.Author, d.At, d.Corrected) + ": " + strings.Join(what, " · ") + "."
	if moves := farmMoves(d.AliveF, d.AliveM); moves != "" {
		s += " " + aliveNoun(d.AliveF, d.AliveM) + " în viață azi în " + d.Farm + ": " + moves + "."
	}
	return s + deleteStays
}

// SectionDeleteEffect is the per-section delete Dialog of a compartment's
// Zilnic bird values (S15): "Se șterge rândul salvat la 08:15: ♀ −2 · ♂ 0. În
// viață în H5.2: ♀ 4 281 → 4 283. Ștergerea rămâne în Istoric modificări."
// f, m = the row's Σ signed qty; aliveF/aliveM = the section's birds today
// (nil = unchanged, not named).
func SectionDeleteEffect(at time.Time, section string, f, m int, aliveF, aliveM *AliveMove) string {
	s := "Se șterge rândul salvat la " + fmtx.TimeOfDay(at) + ": " + sexTotals(f, m) + "."
	var moves []string
	if aliveF != nil {
		moves = append(moves, LeafHens()+" "+fmtx.Int(int64(aliveF.Before))+" → "+fmtx.Int(int64(aliveF.After)))
	}
	if aliveM != nil {
		moves = append(moves, LeafRoosters()+" "+fmtx.Int(int64(aliveM.Before))+" → "+fmtx.Int(int64(aliveM.After)))
	}
	if len(moves) > 0 {
		s += " În viață în " + section + ": " + strings.Join(moves, " · ") + "."
	}
	return s + deleteStays
}

// HallEggsDeleteEffect is the per-hall delete Dialog of its eggs: "Se șterge
// rândul salvat la 08:27: 3 872 ouă. Ștergerea rămâne în Istoric modificări."
// (proposed)
func HallEggsDeleteEffect(at time.Time, eggs int) string {
	return "Se șterge rândul salvat la " + fmtx.TimeOfDay(at) + ": " + fmtx.Int(int64(eggs)) + " ouă." + deleteStays
}

// FeedDailyDeleteEffect is the Hrană zilnică delete Dialog: "Se șterge
// salvarea Hrană zilnică introdusă de Dr. Ceban la 08:44: 8 rânduri · 2 457,9
// kg. Ștergerea rămâne în Istoric modificări." (proposed)
func FeedDailyDeleteEffect(author string, at time.Time, corrected *Corrected, rows int, kg decimal.Decimal) string {
	return "Se șterge salvarea Hrană zilnică " + introduced(author, at, corrected) + ": " + rowsWord(rows) + " · " +
		fmtx.Dec(kg, 1) + " kg." + deleteStays
}

// DeliveryDeleteEffect is the Livrare hrană delete Dialog: "Se șterge livrarea
// introdusă de Ion Rusu la 10:30: 9 000 kg Breeder 2. Ștergerea rămâne în
// Istoric modificări." (proposed)
func DeliveryDeleteEffect(author string, at time.Time, corrected *Corrected, kg decimal.Decimal, feedType string) string {
	return "Se șterge livrarea " + introduced(author, at, corrected) + ": " + fmtx.DecUpTo(kg, 1) + " kg " + feedType +
		"." + deleteStays
}

const deleteStays = " Ștergerea rămâne în Istoric modificări."

func introduced(author string, at time.Time, c *Corrected) string {
	s := "introdusă de " + author + " la " + fmtx.TimeOfDay(at)
	if c != nil {
		s += " și corectată de " + c.By + " pe " + fmtx.Day(c.On)
	}
	return s
}

// MortalityDeleted is the Toast of a whole Zilnic save delete with bird
// values, the farm's birds at the end of today (nil = unchanged, not named):
// "Mortalitate ștearsă · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 568".
func MortalityDeleted(date timex.Day, sections int, aliveF, aliveM *int) string {
	return "Mortalitate ștearsă · " + fmtx.Day(date) + " · " + sectionsWord(sections) + " · în viață azi " + aliveList(aliveF, aliveM)
}

// MortalitySectionDeleted is the Toast of a per-section delete, the section's
// birds at the end of today: "Mortalitate ștearsă · H5.2 · în viață ♀ 4 283".
func MortalitySectionDeleted(section string, aliveF, aliveM *int) string {
	return "Mortalitate ștearsă · " + section + " · în viață " + aliveList(aliveF, aliveM)
}

// EggsDeleted is the Toast of an egg-only Zilnic delete or a per-hall delete:
// "Ouă șterse · 18.11.2026 · H5". (proposed)
func EggsDeleted(date timex.Day, halls []string) string {
	return "Ouă șterse · " + fmtx.Day(date) + " · " + strings.Join(halls, ", ")
}

// FeedDailyDeleted is the Hrană zilnică delete Toast: "Hrană zilnică ștearsă ·
// 18.11.2026 · 8 rânduri · stoc Breeder 2: 33 968 kg". (proposed)
func FeedDailyDeleted(date timex.Day, rows int, stockType string, stock decimal.Decimal) string {
	return "Hrană zilnică ștearsă · " + fmtx.Day(date) + " · " + rowsWord(rows) + " · stoc " + stockType + ": " +
		fmtx.Dec(stock, 0) + " kg"
}

// DeliveryDeleted is the Livrare hrană delete Toast: "Livrare ștearsă · 9 000
// kg Breeder 2 · stoc Hîrbovăț: 22 665 kg". (proposed)
func DeliveryDeleted(kg decimal.Decimal, feedType, farm string, stock decimal.Decimal) string {
	return "Livrare ștearsă · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType + " · stoc " + farm + ": " + fmtx.Dec(stock, 0) + " kg"
}

// DailyUndone is the Toast of an undone Zilnic save: "Zilnic anulat ·
// 18.11.2026". (proposed)
func DailyUndone(date timex.Day) string { return "Zilnic anulat · " + fmtx.Day(date) }

// FeedDailyUndone is the Toast of an undone Hrană zilnică save: "Hrană zilnică
// anulată · 18.11.2026". (proposed)
func FeedDailyUndone(date timex.Day) string { return "Hrană zilnică anulată · " + fmtx.Day(date) }

// DeliveryUndone is the Toast of an undone delivery: "Livrare anulată · 9 000
// kg Breeder 2". (proposed)
func DeliveryUndone(kg decimal.Decimal, feedType string) string {
	return "Livrare anulată · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType
}

// MortalityRestored is the Toast of a restored whole Zilnic save: "Mortalitate
// restaurată · 17.11.2026 · 4 secțiuni · în viață azi ♀ 12 561". (proposed)
func MortalityRestored(date timex.Day, sections int, aliveF, aliveM *int) string {
	return "Mortalitate restaurată · " + fmtx.Day(date) + " · " + sectionsWord(sections) + " · în viață azi " +
		aliveList(aliveF, aliveM)
}

// MortalitySectionRestored is the Toast of a restored per-section delete:
// "Mortalitate restaurată · H5.2 · în viață ♀ 4 281". (proposed)
func MortalitySectionRestored(section string, aliveF, aliveM *int) string {
	return "Mortalitate restaurată · " + section + " · în viață " + aliveList(aliveF, aliveM)
}

// EggsRestored is the Toast of restored eggs: "Ouă restaurate · 18.11.2026 ·
// H5". (proposed)
func EggsRestored(date timex.Day, halls []string) string {
	return "Ouă restaurate · " + fmtx.Day(date) + " · " + strings.Join(halls, ", ")
}

// FeedDailyRestored is the Toast of a restored Hrană zilnică save: "Hrană
// zilnică restaurată · 18.11.2026 · 8 rânduri · stoc Breeder 2: 31 665 kg".
// (proposed)
func FeedDailyRestored(date timex.Day, rows int, stockType string, stock decimal.Decimal) string {
	return "Hrană zilnică restaurată · " + fmtx.Day(date) + " · " + rowsWord(rows) + " · stoc " + stockType + ": " +
		fmtx.Dec(stock, 0) + " kg"
}

// DeliveryRestored is the Toast of a restored delivery: "Livrare restaurată ·
// 9 000 kg Breeder 2 · stoc Hîrbovăț: 25 140 kg". (proposed)
func DeliveryRestored(kg decimal.Decimal, feedType, farm string, stock decimal.Decimal) string {
	return "Livrare restaurată · " + fmtx.DecUpTo(kg, 1) + " kg " + feedType + " · stoc " + farm + ": " +
		fmtx.Dec(stock, 0) + " kg"
}

// Helpers for the M5 copy.

// sectionsWord: "1 secțiune", "4 secțiuni", "20 de secțiuni".
func sectionsWord(n int) string { return countNoun(n, "1 secțiune", "secțiuni") }

// hallsWord: "1 hală", "2 hale", "20 de hale".
func hallsWord(n int) string { return countNoun(n, "1 hală", "hale") }

// valuesWord: "1 valoare", "3 valori", "20 de valori".
func valuesWord(n int) string { return countNoun(n, "1 valoare", "valori") }

// birds is a signed bird count with its noun: "−8 găini", "−1 găină", "−20 de
// cocoși", "4 283 de găini" (U+202F, U+2212).
func birds[S ~string](n int, sex S) string {
	return countGrouped(n, pick(sexIsF(sex), "găină", "cocoș"), fmtx.SexWord(sex))
}

// birdsWith is a bird count with an agreeing participle: birdsWith(4, "F",
// "moartă", "moarte", "mort", "morți") = "4 găini moarte"; 1 → "1 găină
// moartă"; 20 ♂ → "20 de cocoși morți".
func birdsWith[S ~string](n int, sex S, oneF, manyF, oneM, manyM string) string {
	if sexIsF(sex) {
		return countGrouped(n, "găină "+oneF, "găini "+manyF)
	}
	return countGrouped(n, "cocoș "+oneM, "cocoși "+manyM)
}

// aliveArticle is the Z5 "peste …" phrase: "cele 4 283 găini", "cei 323
// cocoși", "cea 1 găină", "cel 1 cocoș" (no "de", as SPEC writes it).
func aliveArticle[S ~string](n int, sex S) string {
	female := sexIsF(sex)
	if n == 1 {
		return pick(female, "cea 1 găină", "cel 1 cocoș")
	}
	return pick(female, "cele ", "cei ") + fmtx.Int(int64(n)) + " " + fmtx.SexWord(sex)
}

// countGrouped is countNoun for any signed n, thousands grouped with U+202F
// and the minus U+2212; "de" follows the absolute value's last two digits.
func countGrouped(n int, one, many string) string {
	abs := n
	if abs < 0 {
		abs = -abs
	}
	noun := many
	switch r := abs % 100; {
	case abs == 1:
		noun = one
	case abs >= 20 && (r == 0 || r >= 20):
		noun = "de " + many
	}
	return fmtx.Int(int64(n)) + " " + noun
}

// signedInt is a difference or a signed sum: "−3", "+2", "0".
func signedInt(n int) string { return fmtx.Signed(decimal.NewFromInt(int64(n)), 0) }

// sexTotals is "♀ −5 · ♂ 0".
func sexTotals(f, m int) string {
	return LeafHens() + " " + signedInt(f) + " · " + LeafRoosters() + " " + signedInt(m)
}

// aliveList names the given sexes' counts: "♀ 12 568", "♀ 12 568 · ♂ 794".
func aliveList(f, m *int) string {
	var parts []string
	if f != nil {
		parts = append(parts, LeafHens()+" "+fmtx.Int(int64(*f)))
	}
	if m != nil {
		parts = append(parts, LeafRoosters()+" "+fmtx.Int(int64(*m)))
	}
	return strings.Join(parts, " · ")
}

// farmMoves is "♀ 12 563 → ♀ 12 568" (· "♂ 794 → ♂ 796"), the sign repeated
// as on the §6.15 Dialog.
func farmMoves(f, m *AliveMove) string {
	var parts []string
	if f != nil {
		parts = append(parts, LeafHens()+" "+fmtx.Int(int64(f.Before))+" → "+LeafHens()+" "+fmtx.Int(int64(f.After)))
	}
	if m != nil {
		parts = append(parts, LeafRoosters()+" "+fmtx.Int(int64(m.Before))+" → "+LeafRoosters()+" "+fmtx.Int(int64(m.After)))
	}
	return strings.Join(parts, " · ")
}

// aliveNoun is the Dialog's subject: "Găini", "Cocoși" or "Păsări".
func aliveNoun(f, m *AliveMove) string {
	switch {
	case f != nil && m != nil:
		return "Păsări"
	case m != nil:
		return "Cocoși"
	default:
		return "Găini"
	}
}

func dailyFieldName(f DailyField) string {
	switch f {
	case FieldMortality:
		return "Mortalitate"
	case FieldCull:
		return "Sacrificare"
	case FieldSexingError:
		return "Eroare sexare"
	default:
		panic("messages: no Z5 name for field " + string(f))
	}
}

func eggFieldWord(f EggField) string {
	switch f {
	case FieldFloor:
		return "ouă jos"
	case FieldCat1:
		return "Cat. 1"
	case FieldCat2:
		return "Cat. 2"
	case FieldCat3:
		return "Cat. 3"
	case FieldHatching:
		return "incubabile"
	case FieldAvgWeightG:
		return "masa medie"
	default:
		panic("messages: unknown egg field " + string(f))
	}
}

// sexIsF reports whether sex is "F"; it panics on a value other than F or M.
func sexIsF[S ~string](sex S) bool { return sexSign(sex) == LeafHens() }

func pick(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}
