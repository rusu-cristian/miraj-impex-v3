// Package ledger holds the rules every save shares (SPEC §12.1, §12.3): the
// farm locks of §6.1 step 0 now; the §6.1 count check (M5) and
// ReconcileCycles (§6.4, M7) later.
package ledger

import (
	"bytes"
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
)

// LockFarms takes the transaction-level advisory lock of every farm in ids,
// once each, in ascending id order (byte order of a UUID = its text order),
// so two saves that touch the same farms never deadlock. The locks end with
// the transaction; call it first in a save's transaction, before any read.
func LockFarms(ctx context.Context, q *dbq.Queries, ids ...uuid.UUID) error {
	ids = slices.Clone(ids)
	slices.SortFunc(ids, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
	for _, id := range slices.Compact(ids) {
		if err := q.LockFarm(ctx, id); err != nil {
			return err
		}
	}
	return nil
}
