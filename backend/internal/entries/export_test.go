package entries

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// SaveRowKeys is every section × sex a Hrană zilnică save of farmID on day
// accepts (fed over EntriesFeedDaySlots), as "code sex", sorted.
func SaveRowKeys(ctx context.Context, db dbq.DBTX, farmID uuid.UUID, day timex.Day) ([]string, error) {
	slots, err := loadFeedSlots(ctx, dbq.New(db), farmID, day)
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, s := range slots {
		if fed(s) {
			out = append(out, s.Code+" "+string(s.Sex))
		}
	}
	slices.Sort(out)
	return out, nil
}

// FormRowKeys are the rows of a form state as "code sex", sorted.
func FormRowKeys(res FeedFormResponse) []string {
	out := []string{}
	for _, r := range res.Rows {
		out = append(out, r.Code+" "+string(r.Sex))
	}
	slices.Sort(out)
	return out
}

// Test seams of §11.8 (plan M5-D8). Production code never sets them.

// SetLock replaces the runner's farm lock (default ledger.LockFarms), e.g.
// with a no-op to prove that a missing lock breaks §6.1 (concurrency test 3).
func (r *Runner) SetLock(fn func(ctx context.Context, q *dbq.Queries, ids ...uuid.UUID) error) {
	r.lock = fn
}

// SetBeforeCheck sets a function Tx.CheckCounts runs just before the §6.1
// check (default nil), e.g. a barrier that holds two saves there.
func (r *Runner) SetBeforeCheck(fn func()) { r.beforeCheck = fn }

// Runner is the service's save runner, for the seams above.
func (s *Service) Runner() *Runner { return s.runner }
