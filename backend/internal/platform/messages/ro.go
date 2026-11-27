// Package messages holds every Romanian string the server returns, as typed
// functions (SPEC §7.0b). Strings are verbatim from SPEC §6–§8, with ș and ț
// in their comma-below forms (U+0219, U+021B), never the cedilla forms.
//
// A string marked "(proposed)" has no board yet; SPEC.md marks it the same way.
package messages

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
