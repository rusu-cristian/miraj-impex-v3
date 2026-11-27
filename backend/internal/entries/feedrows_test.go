package entries

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// M3 ARC-2: the form state lists exactly the rows the save accepts
// (isFeedRow), on the hand-built farm: a retired section with birds (H8),
// halls (H5, H7), a section without ♂ (H5.2), a hall emptied on 14.11 (H7.1
// still has birds at the start of 14.11), the placement day (no bird at its
// start) and a farm without birds.
func TestFormRowsAreSaveSlots(t *testing.T) {
	f := newFormFixture(t)
	for _, tc := range []struct {
		name string
		farm uuid.UUID
		day  timex.Day
		want []string
	}{
		{"Hîrbovăț 18.11", f.hirbovat, fday(time.November, 18), []string{"H5.1 F", "H5.1 M", "H5.2 F"}},
		{"Hîrbovăț 14.11, H7 emptied that day", f.hirbovat, fday(time.November, 14), []string{"H5.1 F", "H5.1 M", "H5.2 F", "H7.1 F"}},
		{"Hîrbovăț on the placement day", f.hirbovat, fday(time.January, 1), []string{}},
		{"Telița, no bird", f.telita, fday(time.November, 18), []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			save, err := SaveRowKeys(t.Context(), f.d.App, tc.farm, tc.day)
			if err != nil {
				t.Fatal(err)
			}
			form := FormRowKeys(f.build(f.mgr, tc.farm, tc.day))
			if !slices.Equal(form, save) || !slices.Equal(save, tc.want) {
				t.Errorf("form rows %v, save slots %v, want both %v", form, save, tc.want)
			}
		})
	}
}
