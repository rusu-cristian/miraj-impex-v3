package ledger

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// SectionSex is one bird-holding section and one sex: the unit §6.1 checks.
type SectionSex struct {
	SectionID uuid.UUID
	Sex       dbq.Sex
}

// Negative is one (section, sex) whose birds alive at the end of a day would
// be below 0 on or after the checked day (§6.1 step 3).
type Negative struct {
	SectionID     uuid.UUID
	Code          string // the section's code, "H5.2"
	Sex           dbq.Sex
	FirstDate     timex.Day // the first negative day on or after from
	FirstAliveEnd int       // the count at the end of FirstDate (< 0)
	LastDate      timex.Day // the last negative day: today while still negative
	MinAliveEnd   int       // the lowest count over the negative days (< 0)
	TodayAliveEnd *int      // the count at the end of today when today is negative, else nil
}

// CheckCounts is §6.1 steps 2–3 (plan M5-D4): after a write has changed the
// bird rows of affected, it returns every pair whose count at the end of a day
// is below 0 on a day from from to max(its last event, today), in affected
// order (a pair listed twice counts once). An empty result means the write
// keeps every count ≥ 0.
//
// Run it inside the write's transaction, after the farm locks (LockFarms) and
// after the change, so it sees the change and no concurrent one. It reads one
// statement over the live event deltas and never app_today(): today is the
// caller's (ARC-6), so it holds under a frozen clock and inside the
// migration's import transaction alike.
func CheckCounts(ctx context.Context, db dbq.DBTX, affected []SectionSex, from, today timex.Day) ([]Negative, error) {
	out := make([]Negative, 0)
	if len(affected) == 0 {
		return out, nil
	}
	seen := make(map[SectionSex]bool, len(affected))
	ids := make([]uuid.UUID, 0, len(affected))
	sexes := make([]string, 0, len(affected))
	for _, a := range affected {
		if seen[a] {
			continue
		}
		seen[a] = true
		ids = append(ids, a.SectionID)
		sexes = append(sexes, string(a.Sex))
	}
	rows, err := dbq.New(db).NegativeCounts(ctx, dbq.NegativeCountsParams{
		SectionIds: ids, Sexes: sexes, FromDay: from, Today: today,
	})
	if err != nil {
		return nil, fmt.Errorf("ledger: check counts from %s: %w", from, err)
	}
	for _, r := range rows {
		n := Negative{
			SectionID:     r.SectionID,
			Code:          r.Code,
			Sex:           r.Sex,
			FirstDate:     r.FirstNegativeDate,
			FirstAliveEnd: int(r.FirstAliveEnd),
			LastDate:      r.LastNegativeDate,
			MinAliveEnd:   int(r.MinAliveEnd),
		}
		if r.NegativeToday {
			v := int(r.TodayAliveEnd)
			n.TodayAliveEnd = &v
		}
		out = append(out, n)
	}
	return out, nil
}
