package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// errReplayNotFound: no stored reply for the key (never used, or purged
// after 7 days).
var errReplayNotFound = errors.New("entries: no stored reply")

// errReplay: the key is already taken (found under the lock, or a unique
// violation on entry_client_request_uq / request_replay_pkey). Create answers
// it after the transaction from the stored reply.
var errReplay = errors.New("entries: client request id already used")

// replay returns the stored response of an earlier save with the same key:
// 200 with the same bytes (§7.0). It never returns a zero Result with a nil
// error: no row → errReplayNotFound; another user or another route → 409
// STALE (a key is never shared between users or routes).
func replay(ctx context.Context, q *dbq.Queries, p auth.Principal, key uuid.UUID, route string) (Result, error) {
	row, err := q.GetReplay(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, errReplayNotFound
	}
	if err != nil {
		return Result{}, err
	}
	if row.UserID != p.UserID || row.Route != route {
		return Result{}, httpx.Stale()
	}
	if len(row.Body) == 0 {
		return Result{}, fmt.Errorf("entries: empty stored reply for route %s", route)
	}
	return Result{Status: http.StatusOK, Body: row.Body}, nil
}

// alreadyEntered answers a reused key whose stored reply is gone with the
// entry that holds it: 409 ALREADY_ENTERED + entryId (§7.0, §4.11) when p
// created that entry and it is of the route's kind; else 409 STALE, as a
// stored reply of another user or route would be (M3 SEC-1: a key never
// names another user's entry). An entry without an author (migration) is
// nobody's.
func alreadyEntered(ctx context.Context, q *dbq.Queries, p auth.Principal, key uuid.UUID, kind dbq.EntryKind) error {
	e, err := q.EntryIDByClientRequest(ctx, key)
	if err != nil {
		return fmt.Errorf("entries: entry of a taken client request id: %w", err)
	}
	if e.CreatedBy == nil || *e.CreatedBy != p.UserID || e.Kind != kind {
		return httpx.Stale()
	}
	return httpx.AlreadyEntered(e.ID, messages.EntryAlreadySaved())
}
