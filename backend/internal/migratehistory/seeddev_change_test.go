package migratehistory

import (
	"encoding/json"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// TestCorrectionChangesBytes (plan M5-D9, T8 criterion 4): the seeded
// correction's `changes`, now written as audit.Change, encode byte for byte
// as the M1 seed's private type did ({key, section, sex, field, old, new},
// counts as JSON numbers), so the seeded audit JSON is unchanged.
func TestCorrectionChangesBytes(t *testing.T) {
	got, err := json.Marshal(correctionChanges("H5.2", dbq.SexF, 1, 3))
	if err != nil {
		t.Fatal(err)
	}
	const before = `[{"key":"H5.2 · găini moarte","section":"H5.2","sex":"F","field":"mortality","old":1,"new":3}]`
	if string(got) != before {
		t.Errorf("changes = %s\nwant      %s", got, before)
	}
}
