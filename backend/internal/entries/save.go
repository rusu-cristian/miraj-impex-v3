// Package entries serves the entry forms (SPEC §7.8, §12.3): their form
// states and creates, and later edit, correct, delete, undo, restore and
// Introdus azi. Every create runs on one save runner (Runner.Create, §12.5,
// M3 D1–D3); a form plugs in only its state load, its plan and its child
// inserts.
package entries

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/pgerr"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Result is an encoded response: 201 for a save, 200 for a dry run or a
// replay. The handler sends it with httpx.WriteRaw, byte for byte.
type Result struct {
	Status int
	Body   []byte
}

// Runner runs every entry create (Create) and every change of a saved entry
// (Change) on the §12.5 skeleton.
type Runner struct {
	store *db.Store
	clock timex.Clock

	// Test seams of §11.8 (plan M5-D8), set only from export_test.go:
	// lock takes the farm locks (ledger.LockFarms); beforeCheck, when not
	// nil, runs in Tx.CheckCounts just before the §6.1 check (a barrier that
	// holds two saves there). Production code never changes them.
	lock        func(ctx context.Context, q *dbq.Queries, ids ...uuid.UUID) error
	beforeCheck func()
}

// NewRunner builds the save runner on the app pool's store.
func NewRunner(store *db.Store, clock timex.Clock) *Runner {
	return &Runner{store: store, clock: clock, lock: ledger.LockFarms}
}

// Create describes one create request to the runner.
type Create struct {
	Route   string          // "POST /api/entries/feed-daily": a clientRequestId is never shared between routes
	Kind    dbq.EntryKind   // the kind of the entry the save inserts
	FarmIDs []uuid.UUID     // every farm the save touches; locked in ascending order before any read
	Meta    httpx.WriteMeta // clientRequestId, source, voice fields, dryRun

	// OnConflict maps a unique violation of the form's own child rows (e.g.
	// feed_allocation_uq → 409 ALREADY_ENTERED + the owning entryId). It runs
	// after the transaction has rolled back, on the pool, because a 23505
	// aborts the transaction (M3 D2). nil, or a nil return, leaves the
	// violation as it is (a 500).
	OnConflict func(ctx context.Context, q *dbq.Queries, ce *pgerr.ConstraintError) error
}

// Plan is a form's part of a save, run inside the transaction after the farm
// locks: load the state, plan (level-2 refusals and warnings), insert the
// entry (Tx.InsertEntry) and its child rows, run the post-insert checks. A dry
// run ends with Tx.Preview; a save returns what the runner writes after it.
type Plan func(ctx context.Context, tx *Tx) (Saved, error)

// Saved is what a planned save hands back to the runner, which then writes,
// in this order, the audit row, the live events, the response (encoded once)
// and the replay row, and commits.
type Saved struct {
	// Response is the 201 body. Its bytes are sent and stored for replays.
	Response any
	// Audit is the save's audit row. The runner sets At (the clock), UserID
	// (the caller), Source and VoiceTranscript.
	Audit  audit.Row
	Events []live.Event
	// Restored are the child rows a change un-deletes (an undo of an update,
	// a restore). A plan whose write fails returns them with the error, so
	// that Runner.Change hands them to ChangeSpec.OnConflict (ARC-5) to name
	// the entry that took a key meanwhile. Create ignores it.
	Restored Snapshot
}

// Tx is a create's transaction as a Plan sees it.
type Tx struct {
	db.Tx
	Principal auth.Principal
	Today     timex.Day
	Now       time.Time // the clock's instant: entry.created_at, audit_log.at, asOf
	DryRun    bool
	Source    dbq.RecordSource

	kind        dbq.EntryKind
	key         *uuid.UUID // the clientRequestId to store; nil for a dry run
	out         *Result
	beforeCheck func() // Runner.beforeCheck (test seam, M5-D8)
}

// CheckCounts is §6.1 steps 2–3 for this write (plan M5-D4, D8): after the
// write's bird rows have changed, every (section, sex) of affected whose
// count at the end of a day from `from` to today would be below 0, in
// affected order (empty = every count stays ≥ 0). It runs on the write's
// transaction, so it sees the write and, behind the farm locks, no
// concurrent one.
func (t *Tx) CheckCounts(ctx context.Context, affected []ledger.SectionSex, from timex.Day) ([]ledger.Negative, error) {
	if t.beforeCheck != nil {
		t.beforeCheck()
	}
	return ledger.CheckCounts(ctx, t.Tx, affected, from, t.Today)
}

// Header is the entry row of a save, beside what the runner fills (kind,
// source, client_request_id, created_by, created_at).
type Header struct {
	FarmID     uuid.UUID
	PeerFarmID *uuid.UUID // transfer: the destination farm
	Date       timex.Day
	SeriesID   *uuid.UUID // placement
	Buyer      *string    // sale
	Note       *string
}

// Inserted is a new entry: its id and its version instant, read back from
// the row (§12.5 Version).
type Inserted struct {
	ID        uuid.UUID
	Kind      dbq.EntryKind
	VersionAt time.Time
}

// Ref is the entry as responses name it: id, kind, version.
func (e Inserted) Ref() httpx.EntryRef {
	return httpx.EntryRef{ID: e.ID, Kind: string(e.Kind), Version: version(e.VersionAt)}
}

// version formats an entry's coalesce(updated_at, created_at) as RFC 3339
// with microseconds (PostgreSQL's precision), in UTC and fixed width, so
// equal instants give equal strings (§7.0 optimistic concurrency, §12.5).
func version(t time.Time) string { return t.UTC().Format("2006-01-02T15:04:05.000000Z07:00") }

// InsertEntry inserts the save's entry: created_by the caller, created_at the
// clock's instant (M3 D3), the save's source, and the clientRequestId (none
// for a dry run, so a preview never collides with or waits on the real save
// of the same draft, §7.0).
func (t *Tx) InsertEntry(ctx context.Context, h Header) (Inserted, error) {
	row, err := t.Q.InsertEntry(ctx, dbq.InsertEntryParams{
		Kind: t.kind, FarmID: h.FarmID, PeerFarmID: h.PeerFarmID, Date: h.Date, SeriesID: h.SeriesID,
		Buyer: h.Buyer, Note: h.Note, Source: t.Source, ClientRequestID: t.key,
		CreatedBy: &t.Principal.UserID, CreatedAt: t.Now,
	})
	if err != nil {
		return Inserted{}, err
	}
	return Inserted{ID: row.ID, Kind: t.kind, VersionAt: row.VersionAt}, nil
}

// Preview ends a dry run: it encodes v as the 200 answer (refusals in the
// body, never a 422, §7.0) and returns db.ErrRollback, which the Plan returns
// so the transaction rolls back every write. Use it only when tx.DryRun.
func (t *Tx) Preview(v any) error {
	if !t.DryRun {
		return errors.New("entries: Preview outside a dry run")
	}
	return answerDryRun(t.out, v)
}

// answerDryRun encodes v as a 200 into out and asks InTx to roll back.
func answerDryRun(out *Result, v any) error {
	body, err := httpx.Marshal(v)
	if err != nil {
		return err
	}
	*out = Result{Status: http.StatusOK, Body: body}
	return db.ErrRollback
}

// DryRunResponse is the 200 answer of a dry run (§7.0): what the save would
// do, its warnings and its refusals.
type DryRunResponse struct {
	Effects  []Effect        `json:"effects"`
	Warnings []httpx.Warning `json:"warnings"`
	Refusals []httpx.Refusal `json:"refusals"`
}

// Effect is one line of a dry run's effect preview (§7.8: "Ciclu închis la
// Telița, sapt. 21"). A feed save has none.
type Effect struct {
	Farm  string `json:"farm,omitzero"`
	Line  string `json:"line"`
	Sub   string `json:"sub,omitzero"`
	Cycle string `json:"cycle,omitzero"` // close | continue | open
}

// NewDryRun is a DryRunResponse whose nil lists are empty ([] in JSON).
func NewDryRun(effects []Effect, warnings []httpx.Warning, refusals []httpx.Refusal) DryRunResponse {
	if effects == nil {
		effects = []Effect{}
	}
	if warnings == nil {
		warnings = []httpx.Warning{}
	}
	if refusals == nil {
		refusals = []httpx.Refusal{}
	}
	return DryRunResponse{Effects: effects, Warnings: warnings, Refusals: refusals}
}

// Unique constraints the runner maps itself: the idempotency key is taken.
const (
	constraintEntryKey  = "entry_client_request_uq"
	constraintReplayKey = "request_replay_pkey"
)

// Create runs one create for p (§12.5, M3 D2):
//
//  1. a key (not a dry run) with a stored reply → that reply (200, the same
//     bytes), or 409 STALE for another user or route;
//  2. one READ COMMITTED transaction: lock c.FarmIDs ascending; read the
//     clock (Tx.Now, Tx.Today); look the key up again under the lock (a
//     concurrent identical request that committed meanwhile wins); then
//     plan; for a dry run, plan's Preview rolls back; else audit row → live
//     events → the response encoded once → the replay row → commit;
//  3. after the transaction, on the pool: a taken key (found under the lock,
//     or entry_client_request_uq / request_replay_pkey) answers like 1; when
//     its reply was purged, 409 ALREADY_ENTERED with the entry that holds it
//     if p created that entry and it is of c.Kind, else 409 STALE; another
//     unique violation goes to c.OnConflict.
//
// Errors are envelope errors (httpx.HTTPError) or unexpected errors (500).
func (r *Runner) Create(ctx context.Context, p auth.Principal, c Create, plan Plan) (Result, error) {
	var key *uuid.UUID
	if c.Meta.ClientRequestID != nil && !c.Meta.DryRun {
		key = c.Meta.ClientRequestID
		res, err := replay(ctx, r.store.Q(), p, *key, c.Route)
		if !errors.Is(err, errReplayNotFound) {
			return res, err // a stored reply, 409 STALE or a database error
		}
	}
	src, transcript := voiceSource(p, c.Kind, c.Meta)

	var out Result
	err := r.store.InTx(ctx, func(dtx db.Tx) error {
		if err := r.lock(ctx, dtx.Q, c.FarmIDs...); err != nil {
			return err
		}
		// The clock is read once the locks are held (M3 COR-3): a save that
		// waited on them past midnight is dated, checked and audited at the
		// instant it runs, not when it arrived.
		now := r.clock.Now()
		if key != nil { // under the lock: a request with this key may have committed while we waited
			_, err := dtx.Q.EntryIDByClientRequest(ctx, *key)
			if err == nil {
				return errReplay
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}
		t := &Tx{
			Tx: dtx, Principal: p, Today: timex.DayOf(now), Now: now, DryRun: c.Meta.DryRun, Source: src,
			kind: c.Kind, key: key, out: &out, beforeCheck: r.beforeCheck,
		}
		saved, err := plan(ctx, t)
		if err != nil {
			return err // db.ErrRollback after Preview: a dry run's answer is in out
		}
		if t.DryRun {
			return errors.New("entries: a dry run reached the save without Preview")
		}
		a := saved.Audit
		a.At, a.UserID, a.Source, a.VoiceTranscript = now, &p.UserID, src, transcript
		if err := audit.Write(ctx, dtx.Q, a); err != nil {
			return err
		}
		if err := live.Emit(ctx, dtx.Q, saved.Events...); err != nil {
			return err
		}
		body, err := httpx.Marshal(saved.Response)
		if err != nil {
			return err
		}
		if key != nil {
			if err := dtx.Q.InsertReplay(ctx, dbq.InsertReplayParams{
				ClientRequestID: *key, UserID: p.UserID, Route: c.Route, Status: http.StatusCreated, Body: body,
			}); err != nil {
				return err
			}
		}
		out = Result{Status: http.StatusCreated, Body: body}
		return nil
	})
	if err == nil {
		return out, nil
	}
	if ce, ok := pgerr.AsConstraint(err, pgerr.Unique); ok {
		switch {
		case key != nil && (ce.Constraint == constraintEntryKey || ce.Constraint == constraintReplayKey):
			err = errReplay
		case c.OnConflict != nil:
			if mapped := c.OnConflict(ctx, r.store.Q(), ce); mapped != nil {
				return Result{}, mapped
			}
		}
	}
	if errors.Is(err, errReplay) { // the key is taken: a concurrent duplicate, or an older save
		res, rerr := replay(ctx, r.store.Q(), p, *key, c.Route)
		if errors.Is(rerr, errReplayNotFound) { // its stored reply was purged (7 days); entry keeps the key
			return Result{}, alreadyEntered(ctx, r.store.Q(), p, *key, c.Kind)
		}
		return res, rerr
	}
	return Result{}, err
}
