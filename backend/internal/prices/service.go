// Package prices serves Prețuri hrană (SPEC §6.8, §7.13, S12; M4 D1, D3,
// D4): the unpriced deliveries, a price batch with its same-day undo token,
// the undo, and the price change with a Motiv (S11 Modifică). Every route is
// the owner's (§3.3): a manager gets 403 PRICE_FORBIDDEN on the writes, any
// other role 403 FORBIDDEN.
//
// Every write runs in one transaction (M4 D3): the asked deliveries' farms
// are locked (ledger.LockFarms), then the rows (select … for update), then
// the level-2 checks, the update, the audit row and the live event. A batch's
// priced_at and its audit row's at are the same instant, so the undo can tell
// that the rows still hold that batch's prices (D4).
package prices

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/ledger"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/tokens"
)

// undoKind is the first field of a price undo token (M4 D4): the undo key's
// purpose label is shared with M10's setup undo, so the kind keeps one from
// verifying as the other.
const undoKind = "price"

// recordTypePrice is the S17 chip of a price audit row (§4.10).
const recordTypePrice = "pret"

// entityDelivery is the audit entity of a price row.
const entityDelivery = "feed_delivery"

// Service is the prices module's service.
type Service struct {
	store   *db.Store
	clock   timex.Clock
	undoKey tokens.Key
}

// NewService builds the prices service on the app pool's store. undoKey is
// tokens.DeriveKey(APP_SECRET, tokens.PurposeUndo) (M4 D4).
func NewService(store *db.Store, clock timex.Clock, undoKey tokens.Key) *Service {
	return &Service{store: store, clock: clock, undoKey: undoKey}
}

// now is the clock's instant at the database's precision (microseconds), so
// priced_at and the audit row's at compare equal after a round trip (D4).
func (s *Service) now() time.Time { return s.clock.Now().Truncate(time.Microsecond) }

// checkOwner is the role check of every price route (§3.3, M4 Q10): the
// owner passes; a manager on a write gets 403 PRICE_FORBIDDEN "Prețul îl
// adaugă proprietarul în Prețuri hrană."; anyone else (and a manager on the
// list) 403 FORBIDDEN "Nu ai acces la această pagină.".
func checkOwner(p auth.Principal, write bool) error {
	switch {
	case p.Role == auth.Owner:
		return nil
	case write && p.Role == auth.Manager:
		return httpx.PriceForbidden()
	}
	return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
}

// Unpriced is the S12 list for p (§7.13), in one read-only snapshot (D10).
func (s *Service) Unpriced(ctx context.Context, p auth.Principal) (UnpricedResponse, error) {
	if err := checkOwner(p, false); err != nil {
		return UnpricedResponse{}, err
	}
	now := s.clock.Now()
	var res UnpricedResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildUnpriced(ctx, tx, now)
		return err
	})
	return res, err
}

// BuildUnpriced is the S12 list (§7.13, M4 Q10): every live unpriced
// delivery oldest first, except those inside a cycle of a closed series;
// count, Σ kg and the farms by name. now is the answer's asOf.
func BuildUnpriced(ctx context.Context, q dbq.DBTX, now time.Time) (UnpricedResponse, error) {
	rows, err := dbq.New(q).PricesUnpriced(ctx)
	if err != nil {
		return UnpricedResponse{}, fmt.Errorf("prices: unpriced: %w", err)
	}
	res := UnpricedResponse{Items: make([]UnpricedItem, 0, len(rows)), Farms: []string{}, AsOf: timex.At(now)}
	kg := decimal.Zero
	type farm struct{ key, name string }
	var farms []farm
	for _, r := range rows {
		var forSex *string
		if r.ForSex != nil {
			s := string(*r.ForSex)
			forSex = &s
		}
		res.Items = append(res.Items, UnpricedItem{
			ID: r.ID, Date: r.Date, Farm: r.FarmName, Type: r.FeedTypeName, ForSex: forSex, Kg: num.Of(r.QtyKg),
		})
		kg = kg.Add(r.QtyKg)
		if f := (farm{r.FarmNameKey, r.FarmName}); !slices.Contains(farms, f) {
			farms = append(farms, f)
		}
	}
	slices.SortFunc(farms, func(a, b farm) int { return cmp.Compare(a.key, b.key) })
	for _, f := range farms {
		res.Farms = append(res.Farms, f.name)
	}
	res.Count, res.Kg = len(res.Items), num.Of(kg)
	return res, nil
}

// Refs reads what the level-1 checks name about the deliveries ids (deleted
// ones too), before the transaction.
func (s *Service) Refs(ctx context.Context, ids []uuid.UUID) (Refs, error) {
	refs := make(Refs, len(ids))
	if len(ids) == 0 {
		return refs, nil
	}
	rows, err := s.store.Q().PricesDeliveryRefs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("prices: refs: %w", err)
	}
	for _, r := range rows {
		refs[r.ID] = Ref{
			FarmID: r.FarmID, Date: r.Date, Kg: r.QtyKg, Type: r.FeedTypeName,
			PerKg: decPtr(r.PricePerKg), Total: decPtr(r.TotalLei), Deleted: r.Deleted,
		}
	}
	return refs, nil
}

// lockDeliveries takes the farm locks of the deliveries ids, ascending, then
// locks their rows (M4 D3), by id.
func lockDeliveries(ctx context.Context, tx db.Tx, ids []uuid.UUID) (map[uuid.UUID]dbq.PricesLockDeliveriesRow, []uuid.UUID, error) {
	farms, err := tx.Q.PricesDeliveryFarms(ctx, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("prices: delivery farms: %w", err)
	}
	if err := ledger.LockFarms(ctx, tx.Q, farms...); err != nil {
		return nil, nil, fmt.Errorf("prices: lock farms: %w", err)
	}
	rows, err := tx.Q.PricesLockDeliveries(ctx, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("prices: lock deliveries: %w", err)
	}
	byID := make(map[uuid.UUID]dbq.PricesLockDeliveriesRow, len(rows))
	for _, r := range rows {
		byID[r.ID] = r
	}
	return byID, sortedFarms(farms), nil
}

// Set prices a batch (§6.8, §7.13, M4 D3, Q10). req must have passed
// Validate (level 1). Inside the transaction, after the locks, in the order
// of the items: a deleted or already priced delivery → 409 STALE, one dated
// inside a cycle of a closed series → 409 SERIES_CLOSED, for the whole batch
// (nothing is saved). Then every row in one update (priced_by = p,
// priced_at = now), one price_set audit row, one price.set live event; the
// answer carries the undo token of the audit row, valid until 23:59 today.
func (s *Service) Set(ctx context.Context, p auth.Principal, req SetRequest) (SetResponse, error) {
	if err := checkOwner(p, true); err != nil {
		return SetResponse{}, err
	}
	if len(req.Items) == 0 { // Validate guarantees it
		return SetResponse{}, httpx.Validation(httpx.Issue{Path: "items", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	now := s.now()
	ids := make([]uuid.UUID, len(req.Items))
	for i, it := range req.Items {
		ids[i] = it.ID
	}
	var res SetResponse
	err := s.store.InTx(ctx, func(tx db.Tx) error {
		rows, farms, err := lockDeliveries(ctx, tx, ids)
		if err != nil {
			return err
		}
		params := dbq.PricesSetParams{
			PricedBy: p.UserID, PricedAt: now, Ids: ids,
			PricesPerKg: make([]decimal.Decimal, len(ids)), TotalsLei: make([]decimal.Decimal, len(ids)),
			PriceInputs: make([]string, len(ids)),
		}
		values := make([]priceValues, len(ids))
		var changes []auditChange
		res = SetResponse{Deliveries: make([]PricedDelivery, len(ids)), Count: len(ids)}
		kg := decimal.Zero
		first := timex.Day{}
		for i, it := range req.Items {
			r, ok := rows[it.ID]
			switch {
			case !ok, r.Deleted, r.TotalLei.Valid, r.PricePerKg.Valid:
				return httpx.Stale()
			case r.ClosedSeries != "":
				return httpx.SeriesClosed(r.ClosedSeries)
			}
			price, err := derive(r.QtyKg, it.PricePerKg, it.TotalLei)
			if err != nil { // the kg changed since the level-1 check
				path, code, msg := priceIssue("items["+strconv.Itoa(i)+"].", refOf(r), err)
				return httpx.Validation(httpx.Issue{Path: path, Code: code, Message: msg})
			}
			params.PricesPerKg[i], params.TotalsLei[i], params.PriceInputs[i] = price.PerKg, price.Total, string(price.Input)
			perKg, total, input := num.Of(price.PerKg), num.Of(price.Total), string(price.Input)
			values[i] = priceValues{
				ID: r.ID, Type: r.FeedTypeName, Date: r.Date, Kg: num.Of(r.QtyKg), PricePerKg: &perKg, TotalLei: &total, PriceInput: &input,
			}
			changes = append(changes, priceChanges(r.FeedTypeName, r.Date, nil, &perKg, nil, &total)...)
			res.Deliveries[i] = PricedDelivery{ID: r.ID, PricePerKg: perKg, TotalLei: total}
			kg = kg.Add(r.QtyKg)
			if first.IsZero() || r.Date.Before(first) {
				first = r.Date
			}
		}
		n, err := tx.Q.PricesSet(ctx, params)
		if err != nil {
			return fmt.Errorf("prices: set: %w", err)
		}
		if n != int64(len(ids)) { // the rows are locked and were checked unpriced
			return fmt.Errorf("prices: set updated %d of %d deliveries", n, len(ids))
		}
		summary := messages.PricesSummary(len(ids), kg)
		auditID, err := audit.WriteID(ctx, tx.Q, audit.Row{
			At: now, UserID: &p.UserID, Action: dbq.AuditActionPriceSet, Entity: entityDelivery, EntityID: ids[0],
			RecordTypes: []string{recordTypePrice}, FarmIDs: farms, RecordDate: &first, Summary: &summary,
			Changes: changes, NewValues: values,
		})
		if err != nil {
			return err
		}
		if err := live.Emit(ctx, tx.Q, live.PriceSet(farms, ids, kg)); err != nil {
			return err
		}
		token, err := tokens.Sign(s.undoKey, []string{undoKind, strconv.FormatInt(auditID, 10), p.UserID.String()},
			undoExpiry(now))
		if err != nil {
			return fmt.Errorf("prices: sign undo token: %w", err)
		}
		res.Kg, res.Toast, res.UndoToken, res.AsOf = num.Of(kg), messages.PricesSaved(len(ids), kg), token, timex.At(now)
		return nil
	})
	if err != nil {
		return SetResponse{}, err
	}
	return res, nil
}

// undoExpiry is when a price undo token saved at now stops being valid
// (M4 D4): 00:00 Europe/Chisinau of the next day, so it works until 23:59 of
// the save day, on a 23- or 25-hour day too (time.Date in the zone).
func undoExpiry(now time.Time) time.Time { return timex.DayOf(now).AddDays(1).Start() }

// badUndoToken is the 400 of a malformed, tampered or non-price undo token
// (M4 Q11).
func badUndoToken() error {
	return httpx.Validation(httpx.Issue{Path: "undoToken", Code: httpx.IssueBadType, Message: messages.InvalidData()})
}

// Undo returns a price batch to "fără preț" (§7.13, D122, M4 D4, Q11). The
// token must verify with the undo key and be of the price kind (else 400 on
// undoToken), not be past 23:59 of the save day (409 STALE "Nu se mai poate
// anula: ziua salvării s-a încheiat.") and be p's (403 FORBIDDEN). Inside the
// transaction, after the locks: every delivery of the batch must still hold
// that save's prices (priced_at = the audit row's at, priced_by = p, not
// deleted), else 409 STALE and nothing changes; a delivery inside a cycle
// of a closed series is 409 SERIES_CLOSED. Then the five price columns are
// cleared, an undo audit row (same_day_author, old_values = the prices) and a
// price.changed live event are written.
func (s *Service) Undo(ctx context.Context, p auth.Principal, req UndoRequest) (UndoResponse, error) {
	if err := checkOwner(p, true); err != nil {
		return UndoResponse{}, err
	}
	now := s.now()
	fields, err := tokens.Verify(s.undoKey, req.UndoToken, now)
	switch {
	case errors.Is(err, tokens.ErrExpired):
		return UndoResponse{}, httpx.NewError(http.StatusConflict, httpx.CodeStale, messages.UndoDayEnded())
	case err != nil, len(fields) != 3, fields[0] != undoKind:
		return UndoResponse{}, badUndoToken()
	}
	auditID, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return UndoResponse{}, badUndoToken()
	}
	userID, err := uuid.Parse(fields[2])
	if err != nil {
		return UndoResponse{}, badUndoToken()
	}
	if userID != p.UserID {
		return UndoResponse{}, httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
	}

	var res UndoResponse
	err = s.store.InTx(ctx, func(tx db.Tx) error {
		row, err := tx.Q.PricesAuditRow(ctx, auditID)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpx.Stale()
		}
		if err != nil {
			return fmt.Errorf("prices: audit row: %w", err)
		}
		if row.Action != dbq.AuditActionPriceSet || row.UserID == nil || *row.UserID != p.UserID {
			return httpx.Stale()
		}
		var batch []priceValues
		if err := json.Unmarshal(row.NewValues, &batch); err != nil {
			return fmt.Errorf("prices: audit row %d new_values: %w", auditID, err)
		}
		if len(batch) == 0 {
			return fmt.Errorf("prices: audit row %d lists no delivery", auditID)
		}
		ids := make([]uuid.UUID, len(batch))
		for i, b := range batch {
			ids[i] = b.ID
		}
		rows, farms, err := lockDeliveries(ctx, tx, ids)
		if err != nil {
			return err
		}
		old := make([]priceValues, len(ids))
		var changes []auditChange
		kg := decimal.Zero
		first := timex.Day{}
		for i, id := range ids {
			r, ok := rows[id]
			switch {
			case !ok, r.Deleted, r.PricedAt == nil, !r.PricedAt.Equal(row.At), r.PricedBy == nil, *r.PricedBy != p.UserID,
				!r.PricePerKg.Valid, !r.TotalLei.Valid:
				return httpx.Stale()
			case r.ClosedSeries != "":
				return httpx.SeriesClosed(r.ClosedSeries)
			}
			perKg, total := num.Of(r.PricePerKg.Decimal), num.Of(r.TotalLei.Decimal)
			old[i] = priceValues{
				ID: r.ID, Type: r.FeedTypeName, Date: r.Date, Kg: num.Of(r.QtyKg), PricePerKg: &perKg, TotalLei: &total, PriceInput: r.PriceInput,
			}
			changes = append(changes, priceChanges(r.FeedTypeName, r.Date, &perKg, nil, &total, nil)...)
			kg = kg.Add(r.QtyKg)
			if first.IsZero() || r.Date.Before(first) {
				first = r.Date
			}
		}
		n, err := tx.Q.PricesUndo(ctx, dbq.PricesUndoParams{Ids: ids, PricedAt: row.At, PricedBy: p.UserID})
		if err != nil {
			return fmt.Errorf("prices: undo: %w", err)
		}
		if n != int64(len(ids)) {
			return fmt.Errorf("prices: undo cleared %d of %d deliveries", n, len(ids))
		}
		summary := messages.PricesSummary(len(ids), kg)
		if err := audit.Write(ctx, tx.Q, audit.Row{
			At: now, UserID: &p.UserID, Action: dbq.AuditActionUndo, Entity: entityDelivery, EntityID: ids[0],
			RecordTypes: []string{recordTypePrice}, FarmIDs: farms, RecordDate: &first, SameDayAuthor: true,
			Summary: &summary, Changes: changes, OldValues: old,
		}); err != nil {
			return err
		}
		if err := live.Emit(ctx, tx.Q, live.PriceChanged(farms, ids, kg)); err != nil {
			return err
		}
		res = UndoResponse{Deliveries: ids, Toast: messages.PricesUndone(len(ids)), AsOf: timex.At(now)}
		return nil
	})
	if err != nil {
		return UndoResponse{}, err
	}
	return res, nil
}

// Change sets a new price on a priced delivery with a Motiv (§6.8, §6.15,
// S11 Modifică, M4 Q12). req must have passed Validate (level 1) against the
// delivery's Ref. Inside the transaction, after the locks: an unknown or
// deleted delivery → 404; an unpriced one → 409 STALE; one inside a cycle of
// a closed series → 409 SERIES_CLOSED; the same price per kg and total as
// stored → 422 REFUSED "Nu ai schimbat nicio cifră.". Then the row
// (priced_by = p, priced_at = now, no longer "din registru"), a price_change
// audit row with the Motiv and the old/new pairs, and a price.changed event.
func (s *Service) Change(ctx context.Context, p auth.Principal, id uuid.UUID, req ChangeRequest) (ChangeResponse, error) {
	if err := checkOwner(p, true); err != nil {
		return ChangeResponse{}, err
	}
	reason := req.reason()
	if reason == "" { // Validate guarantees it: a change never saves without a Motiv
		return ChangeResponse{}, httpx.Validation(httpx.Issue{Path: "reason", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	now := s.now()
	var res ChangeResponse
	err := s.store.InTx(ctx, func(tx db.Tx) error {
		rows, farms, err := lockDeliveries(ctx, tx, []uuid.UUID{id})
		if err != nil {
			return err
		}
		r, ok := rows[id]
		switch {
		case !ok, r.Deleted:
			return httpx.NotFound()
		case !r.TotalLei.Valid || !r.PricePerKg.Valid:
			return httpx.Stale()
		case r.ClosedSeries != "":
			return httpx.SeriesClosed(r.ClosedSeries)
		}
		price, err := derive(r.QtyKg, req.PricePerKg, req.TotalLei)
		if err != nil { // the kg changed since the level-1 check
			path, code, msg := priceIssue("", refOf(r), err)
			return httpx.Validation(httpx.Issue{Path: path, Code: code, Message: msg})
		}
		oldPerKg, oldTotal := r.PricePerKg.Decimal, r.TotalLei.Decimal
		if price.PerKg.Equal(oldPerKg) && price.Total.Equal(oldTotal) {
			field := "pricePerKg"
			if price.Input == calc.PriceInputTotal {
				field = "totalLei"
			}
			return httpx.RefusedWith([]httpx.Refusal{{Field: field, Message: messages.NothingChanged()}}, nil)
		}
		n, err := tx.Q.PricesChange(ctx, dbq.PricesChangeParams{
			PricePerKg: price.PerKg, TotalLei: price.Total, PriceInput: string(price.Input),
			PricedBy: p.UserID, PricedAt: now, ID: id,
		})
		if err != nil {
			return fmt.Errorf("prices: change: %w", err)
		}
		if n != 1 {
			return fmt.Errorf("prices: change updated %d deliveries", n)
		}
		oldP, oldT, newP, newT := num.Of(oldPerKg), num.Of(oldTotal), num.Of(price.PerKg), num.Of(price.Total)
		input := string(price.Input)
		var oldAt *timex.LocalTime
		if r.PricedAt != nil {
			at := timex.At(*r.PricedAt)
			oldAt = &at
		}
		summary := messages.PriceChangeSummary(r.FeedTypeName, r.Date)
		if err := audit.Write(ctx, tx.Q, audit.Row{
			At: now, UserID: &p.UserID, Action: dbq.AuditActionPriceChange, Entity: entityDelivery, EntityID: id,
			RecordTypes: []string{recordTypePrice}, FarmIDs: farms, RecordDate: &r.Date, Reason: &reason, Summary: &summary,
			Changes: priceChanges(r.FeedTypeName, r.Date, &oldP, &newP, &oldT, &newT),
			OldValues: changeOldValues{
				priceValues: priceValues{
					ID: id, Type: r.FeedTypeName, Date: r.Date, Kg: num.Of(r.QtyKg), PricePerKg: &oldP, TotalLei: &oldT,
					PriceInput: r.PriceInput,
				},
				PricedBy: r.PricedBy, PricedAt: oldAt, PriceFromRegister: r.PriceFromRegister,
			},
			NewValues: priceValues{
				ID: id, Type: r.FeedTypeName, Date: r.Date, Kg: num.Of(r.QtyKg), PricePerKg: &newP, TotalLei: &newT, PriceInput: &input,
			},
		}); err != nil {
			return err
		}
		if err := live.Emit(ctx, tx.Q, live.PriceChanged(farms, []uuid.UUID{id}, r.QtyKg)); err != nil {
			return err
		}
		res = ChangeResponse{
			Delivery: ChangedDelivery{ID: id, PricePerKg: newP, TotalLei: newT, PricedBy: p.Name, PricedAt: timex.At(now)},
			Toast:    messages.PriceChanged(r.FeedTypeName, r.Date, price.PerKg, price.Total),
			AsOf:     timex.At(now),
		}
		return nil
	})
	if err != nil {
		return ChangeResponse{}, err
	}
	return res, nil
}

// refOf is the Ref of a locked row (for a level-1 message raised late).
func refOf(r dbq.PricesLockDeliveriesRow) Ref {
	return Ref{
		FarmID: r.FarmID, Date: r.Date, Kg: r.QtyKg, Type: r.FeedTypeName,
		PerKg: decPtr(r.PricePerKg), Total: decPtr(r.TotalLei), Deleted: r.Deleted,
	}
}

func decPtr(d decimal.NullDecimal) *decimal.Decimal {
	if !d.Valid {
		return nil
	}
	return &d.Decimal
}

// sortedFarms is ids without repeats, ascending (the audit row's farm_ids).
func sortedFarms(ids []uuid.UUID) []uuid.UUID {
	out := slices.Clone(ids)
	slices.SortFunc(out, func(a, b uuid.UUID) int { return slices.Compare(a[:], b[:]) })
	return slices.Compact(out)
}
