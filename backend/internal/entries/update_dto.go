package entries

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of PUT /api/entries/{id} and POST /api/entries/{id}/undo (§7.8,
// M5 Q6–Q8).

// EntryUpdateRequest is the body of PUT /api/entries/{id} (§7.8): the
// version the client read (ifVersion), the complete new state of the save in
// the create shape of the entry's kind (V: DailyValues, …; Q6: what GET
// /api/entries/{id} returns as `values`), the Motiv (required for a
// correction, ignored for a same-day author edit), dryRun (the correction
// preview, Q7) and the voice fields of a create. It carries no
// clientRequestId (plan M5-D2: a retry meets the moved version).
type EntryUpdateRequest[V any] struct {
	IfVersion       string  `json:"ifVersion"`
	Values          V       `json:"values"`
	Reason          *string `json:"reason"`
	DryRun          bool    `json:"dryRun"`
	Source          string  `json:"source"`
	VoiceTranscript *string `json:"voiceTranscript"`
	VoiceToken      *string `json:"voiceToken"`
}

// DailyUpdateRequest is a PUT of a Zilnic entry.
type DailyUpdateRequest = EntryUpdateRequest[DailyValues]

// reason is the Motiv trimmed, "" when absent.
func (r EntryUpdateRequest[V]) reason() string {
	if r.Reason == nil {
		return ""
	}
	return strings.TrimSpace(*r.Reason)
}

// meta is the request's voice fields as the change runner reads them.
func (r EntryUpdateRequest[V]) meta() httpx.WriteMeta {
	return httpx.WriteMeta{Source: r.Source, VoiceTranscript: r.VoiceTranscript, VoiceToken: r.VoiceToken, DryRun: r.DryRun}
}

// validateHead is the first part of the level-1 check every kind's PUT
// shares (§7.0b, plan M5-D3, D6), in body order: ifVersion required and an
// RFC 3339 instant (else bad_type); values.farmId and values.date, when sent,
// equal to the stored entry's (D101: the date is not editable; "Datele
// trimise nu sunt valide."). The kind's own check of `values` follows
// (ValidateAt with the "values." prefix), then validateTail.
func (r EntryUpdateRequest[V]) validateHead(v *httpx.Validator, t updateTarget, farmID uuid.UUID, date timex.Day) {
	if v.Check(r.IfVersion != "", "ifVersion", httpx.IssueRequired, messages.InvalidData()) {
		_, err := ParseVersion(r.IfVersion)
		v.Check(err == nil, "ifVersion", httpx.IssueBadType, messages.InvalidData())
	}
	if farmID != uuid.Nil {
		v.Check(farmID == t.FarmID, "values.farmId", httpx.IssueOutOfRange, messages.InvalidData())
	}
	if !date.IsZero() {
		v.Check(date == t.Date, "values.date", httpx.IssueOutOfRange, messages.InvalidData())
	}
}

// validateTail is the rest of the shared level-1 check, after `values`: the
// Motiv's length (§6.15) and the source.
func (r EntryUpdateRequest[V]) validateTail(v *httpx.Validator) {
	if n := utf8.RuneCountInString(r.reason()); n > reasonMax {
		v.Add("reason", httpx.IssueTooLong, messages.ReasonTooLong(n, reasonMax))
	}
	ValidateWriteMeta(v, httpx.WriteMeta{Source: r.Source})
}

// validateDailyUpdate is the level-1 check of a Zilnic PUT (§7.0b):
// validateHead, the values as a Zilnic create checks them with every path
// under "values." (DailyValues.ValidateAt), validateTail. t is the stored
// entry; refs the codes of its farm's sections.
func validateDailyUpdate(v *httpx.Validator, r DailyUpdateRequest, t updateTarget, refs DailyRefs) {
	r.validateHead(v, t, r.Values.FarmID, r.Values.Date)
	r.Values.ValidateAt(v, refs, "values.")
	r.validateTail(v)
}

// UndoRequest is the body of POST /api/entries/{id}/undo: {} (§7.8).
type UndoRequest struct{}

// updateSnapshot is an update's audit old_values (an undo's too): the entry's
// live child rows before it, with their ids and tables (Snapshot, plan
// M5-D7), and the entry's corrected flag before it. An undo of the update
// puts exactly these rows and this flag back (Q8).
type updateSnapshot struct {
	Snapshot
	Corrected bool `json:"corrected"`
}
