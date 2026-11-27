package entries

import (
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// EntryMeta is what the §3.2 time windows need to know about a saved entry.
type EntryMeta struct {
	FarmIDs      []uuid.UUID // farm_id and, for a transfer, peer_farm_id: the caller needs every one (§3.1)
	CreatedBy    *uuid.UUID  // nil for a migration or system entry
	CreatedAt    time.Time
	SeriesClosed bool // dated inside a cycle of a closed series: read-only for everyone (§3.1)
}

// EntryActions are the changes the caller may make to an entry (§3.2):
// Editable = edit or delete without a Motiv (the author, on the creation
// day); Correctable = Corectează / Șterge with a Motiv (manager on own farms,
// owner).
type EntryActions struct {
	Editable    bool
	Correctable bool
}

// Actions is what p may do with entry e on day today (§3.2, D70, D71):
//   - vet: edit own entries while today is the local day of created_at;
//     never a correction, never someone else's entry;
//   - manager: edit own entries the same creation day; correct any entry on
//     own farms (someone else's entry from today too, D70 B);
//   - owner: as the manager, on every farm;
//   - viewer: nothing.
//
// An entry on a farm the caller can't see, a migration entry (no author: never
// editable) and an entry in a closed series follow from the same rules.
func Actions(p auth.Principal, e EntryMeta, today timex.Day) EntryActions {
	if e.SeriesClosed || !p.Role.In(auth.Vet, auth.Manager, auth.Owner) || len(e.FarmIDs) == 0 {
		return EntryActions{}
	}
	for _, id := range e.FarmIDs {
		if !p.CanSeeFarm(id) {
			return EntryActions{}
		}
	}
	own := e.CreatedBy != nil && *e.CreatedBy == p.UserID
	return EntryActions{
		Editable:    own && timex.DayOf(e.CreatedAt) == today,
		Correctable: p.Role.In(auth.Manager, auth.Owner),
	}
}
