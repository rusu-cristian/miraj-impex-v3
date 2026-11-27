package entries

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// TestActions is the §3.2 table: who may edit (no Motiv) or correct (Motiv)
// an entry, on 18.11.2026.
func TestActions(t *testing.T) {
	var (
		farm, other           = uuid.New(), uuid.New()
		vetID, mgrID, ownerID = uuid.New(), uuid.New(), uuid.New()
		viewerID, someone     = uuid.New(), uuid.New()
		today                 = timex.MustDay(2026, time.November, 18)
		at0844                = time.Date(2026, time.November, 18, 8, 44, 0, 0, timex.Chisinau)
		yesterday2350         = time.Date(2026, time.November, 17, 23, 50, 0, 0, timex.Chisinau)
		// 22:30 UTC on 17.11 is already 00:30 on 18.11 in Chișinău: the local day counts.
		utcLate = time.Date(2026, time.November, 17, 22, 30, 0, 0, time.UTC)
	)
	vet := auth.Principal{UserID: vetID, Role: auth.Vet, FarmIDs: []uuid.UUID{farm}}
	mgr := auth.Principal{UserID: mgrID, Role: auth.Manager, FarmIDs: []uuid.UUID{farm}}
	owner := auth.Principal{UserID: ownerID, Role: auth.Owner}
	viewer := auth.Principal{UserID: viewerID, Role: auth.Viewer, FarmIDs: []uuid.UUID{farm}}
	entry := func(by uuid.UUID, at time.Time, farms ...uuid.UUID) EntryMeta {
		return EntryMeta{FarmIDs: farms, CreatedBy: &by, CreatedAt: at}
	}
	migration := EntryMeta{FarmIDs: []uuid.UUID{farm}, CreatedAt: at0844}
	closed := entry(mgrID, at0844, farm)
	closed.SeriesClosed = true

	tests := []struct {
		name string
		p    auth.Principal
		e    EntryMeta
		want EntryActions
	}{
		{"vet own entry, creation day", vet, entry(vetID, at0844, farm), EntryActions{Editable: true}},
		{"vet own entry, local day of a UTC-evening instant", vet, entry(vetID, utcLate, farm), EntryActions{Editable: true}},
		{"vet own entry, next day", vet, entry(vetID, yesterday2350, farm), EntryActions{}},
		{"vet, someone else's entry today", vet, entry(someone, at0844, farm), EntryActions{}},
		{"vet own entry on a farm no longer his", vet, entry(vetID, at0844, other), EntryActions{}},
		{"manager own entry, creation day", mgr, entry(mgrID, at0844, farm), EntryActions{Editable: true, Correctable: true}},
		{"manager own entry, past day", mgr, entry(mgrID, yesterday2350, farm), EntryActions{Correctable: true}},
		{"manager, someone else's entry today", mgr, entry(someone, at0844, farm), EntryActions{Correctable: true}},
		{"manager, another farm", mgr, entry(someone, at0844, other), EntryActions{}},
		{"manager, transfer to a farm not his", mgr, entry(mgrID, at0844, farm, other), EntryActions{}},
		{"owner, someone else's entry", owner, entry(someone, yesterday2350, other), EntryActions{Correctable: true}},
		{"owner own entry today", owner, entry(ownerID, at0844, farm), EntryActions{Editable: true, Correctable: true}},
		{"viewer", viewer, entry(viewerID, at0844, farm), EntryActions{}},
		{"migration entry, manager", mgr, migration, EntryActions{Correctable: true}},
		{"migration entry, vet", vet, migration, EntryActions{}},
		{"closed series, owner", owner, closed, EntryActions{}},
		{"no farm", owner, EntryMeta{CreatedBy: &ownerID, CreatedAt: at0844}, EntryActions{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Actions(tc.p, tc.e, today); got != tc.want {
				t.Errorf("Actions = %+v, want %+v", got, tc.want)
			}
		})
	}
}
