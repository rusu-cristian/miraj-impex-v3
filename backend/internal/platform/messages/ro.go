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
func weeksWord(n int) string {
	if n == 1 {
		return "o săptămână"
	}
	if r := n % 100; n >= 20 && (r == 0 || r >= 20) {
		return strconv.Itoa(n) + " de săptămâni"
	}
	return strconv.Itoa(n) + " săptămâni"
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
func SexLeaf[S ~string](sex S, word string) string {
	switch sex {
	case "F":
		return LeafHens() + " " + word
	case "M":
		return LeafRoosters() + " " + word
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
