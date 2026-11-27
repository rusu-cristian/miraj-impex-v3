package entries

import (
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// The values of WriteMeta.Source (§7.8). Empty means manual.
const (
	sourceManual = "manual"
	sourceVoice  = "voice"
)

// ValidateWriteMeta is the level-1 check of the fields every create body
// embeds (M3 D4): `source` is "manual" (the default when absent) or "voice";
// anything else is a 400 VALIDATION issue on `source`.
func ValidateWriteMeta(v *httpx.Validator, m httpx.WriteMeta) {
	switch m.Source {
	case "", sourceManual, sourceVoice:
	default:
		v.Add("source", httpx.IssueOutOfRange, messages.InvalidData())
	}
}

// voiceSource is the source and voice transcript a save is recorded with
// (§4.7, §8.9). A save counts as voice, and keeps its transcript, only with a
// voiceToken that verifies for this user, form and transcript; tokens come
// with M12, so until then every save is manual and the transcript is ignored
// (M3 D4). form is the entry kind the token must name.
func voiceSource(_ auth.Principal, _ dbq.EntryKind, _ httpx.WriteMeta) (dbq.RecordSource, *string) {
	return dbq.RecordSourceManual, nil
}
