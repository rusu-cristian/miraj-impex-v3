package entries

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of DELETE /api/entries/{id}, POST /api/entries/{id}/restore and
// the undo of a create (§7.8, M5 Q7–Q9).

// reasonMax is the longest Motiv after trimming (§6.15).
const reasonMax = 200

// DeleteRequest is the body of DELETE /api/entries/{id} (§7.8): the Motiv
// (required unless the author deletes on the creation day), the section
// whose Zilnic values alone are deleted (S15, D98; daily entries only) and
// dryRun (the Dialog's preview, Q7).
type DeleteRequest struct {
	Reason    *string    `json:"reason"`
	SectionID *uuid.UUID `json:"sectionId"`
	DryRun    bool       `json:"dryRun"`
}

// reason is the Motiv trimmed, "" when absent.
func (r DeleteRequest) reason() string {
	if r.Reason == nil {
		return ""
	}
	return strings.TrimSpace(*r.Reason)
}

// Validate is the level-1 check (§7.0b): a Motiv of at most 200 characters
// after trimming (§6.15 sentence). Whether a Motiv is required depends on
// the caller's window, read under the locks (allowDelete).
func (r DeleteRequest) Validate(v *httpx.Validator) {
	if n := utf8.RuneCountInString(r.reason()); n > reasonMax {
		v.Add("reason", httpx.IssueTooLong, messages.ReasonTooLong(n, reasonMax))
	}
}

// RestoreRequest is the body of POST /api/entries/{id}/restore: {} (§7.8).
type RestoreRequest struct{}

// CorrectionPreview is the 200 of a correction's or a delete's dry run
// (§7.8, M5 Q7): the changes (none for a delete), the effect lines (a
// delete's first line is its Dialog sentence, Q9), per changed section × sex
// the birds alive at the end of the entry date and of today before and
// after, the days checked against §6.1 (entry date to today, 0 when no bird
// moves), and the refusals and warnings the save would meet.
type CorrectionPreview struct {
	Changes     []audit.Change  `json:"changes"`
	Effects     []string        `json:"effects"`
	Alive       []PreviewAlive  `json:"alive"`
	CheckedDays int             `json:"checkedDays"`
	Refusals    []httpx.Refusal `json:"refusals"`
	Warnings    []httpx.Warning `json:"warnings"`
}

// PreviewAlive is one section × sex at the end of one day, before and after
// the change.
type PreviewAlive struct {
	Section string    `json:"section"`
	Sex     dbq.Sex   `json:"sex"`
	Date    timex.Day `json:"date"`
	Before  int       `json:"before"`
	After   int       `json:"after"`
}

// newCorrectionPreview fills nil lists with empty ones ([] in JSON).
func newCorrectionPreview(p CorrectionPreview) CorrectionPreview {
	if p.Changes == nil {
		p.Changes = []audit.Change{}
	}
	if p.Effects == nil {
		p.Effects = []string{}
	}
	if p.Alive == nil {
		p.Alive = []PreviewAlive{}
	}
	if p.Refusals == nil {
		p.Refusals = []httpx.Refusal{}
	}
	if p.Warnings == nil {
		p.Warnings = []httpx.Warning{}
	}
	return p
}

// removedRows is a delete's or an undo's audit old_values (a restore's
// new_values): the child rows it soft-deleted, with their ids and tables
// (Snapshot, plan M5-D7), and the section of a per-section delete (S15).
// A restore un-deletes exactly these rows.
type removedRows struct {
	Snapshot
	SectionID *uuid.UUID `json:"sectionId,omitzero"`
}
