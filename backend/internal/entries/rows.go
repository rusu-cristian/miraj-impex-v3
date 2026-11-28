package entries

import "github.com/google/uuid"

// Snapshot helpers shared by the delete, the undo, the restore and the kind
// adapters (M5 ARC-2, ARC-3).

// snapshotRow is a child row of a Snapshot, known by its id.
type snapshotRow interface{ rowID() uuid.UUID }

func (b SnapshotBird) rowID() uuid.UUID     { return b.ID }
func (g SnapshotEggs) rowID() uuid.UUID     { return g.ID }
func (r SnapshotFeed) rowID() uuid.UUID     { return r.ID }
func (d SnapshotDelivery) rowID() uuid.UUID { return d.ID }

// ids is the id of each row, in order; never nil (a uuid[] query
// parameter).
func ids[T snapshotRow](rows []T) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.rowID())
	}
	return out
}

// idDiff is the match of an entry's live rows now to the rows before, by row
// id (the undo of an update, Q8).
type idDiff[T snapshotRow] struct {
	restore []T         // rows of before not live now (the update soft-deleted them): un-delete
	update  []T         // rows of before live now with other figures: change back in place to these
	remove  []uuid.UUID // live rows before doesn't hold (the update inserted them): soft-delete
}

// diffByID matches now to before by row id: restore and update in before's
// order, remove in now's. same reports whether the live row cur still holds
// before's figures.
func diffByID[T snapshotRow](now, before []T, same func(cur, before T) bool) idDiff[T] {
	var d idDiff[T]
	live := make(map[uuid.UUID]T, len(now))
	for _, r := range now {
		live[r.rowID()] = r
	}
	wanted := make(map[uuid.UUID]bool, len(before))
	for _, b := range before {
		wanted[b.rowID()] = true
		cur, ok := live[b.rowID()]
		switch {
		case !ok:
			d.restore = append(d.restore, b)
		case !same(cur, b):
			d.update = append(d.update, b)
		}
	}
	for _, r := range now {
		if !wanted[r.rowID()] {
			d.remove = append(d.remove, r.rowID())
		}
	}
	return d
}

// snapshotRows is how many rows s holds, over its four tables.
func snapshotRows(s Snapshot) int {
	return len(s.BirdEvents) + len(s.EggProduction) + len(s.FeedAllocations) + len(s.FeedDeliveries)
}
