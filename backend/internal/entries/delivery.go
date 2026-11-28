package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/audit"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/live"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// routeDelivery is the Livrare hrană create; the runner never shares a
// clientRequestId between routes.
const routeDelivery = "POST /api/entries/delivery"

// recordTypeDelivery is the S17 chip of a delivery save.
const recordTypeDelivery = "livrare"

// createDelivery is POST /api/entries/delivery (§6.7, §7.8, M4 D5): decode
// (64 KiB) → role and farm → price fields → the farm and type names →
// Validate (level 1) → service. The answer is sent byte for byte as the
// runner encoded it (the bytes a replay sends).
func (h *Handler) createDelivery(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	req, err := httpx.Decode[DeliveryCreateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := checkDelivery(p, req); err != nil { // before any read: no name of another farm leaks
		httpx.WriteError(w, err)
		return
	}
	refs, err := h.svc.DeliveryRefs(r.Context(), req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	var v httpx.Validator
	req.Validate(&v, refs)
	if err := v.Err(); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.CreateDelivery(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}

// checkDeliveryWriter is the access check of the Livrare hrană routes (§3.3,
// §6.7): manager and owner may enter deliveries, anyone else gets 403
// FORBIDDEN "Nu ai acces la această pagină."; a farm outside the caller's
// farms is 403 FARM_FORBIDDEN. A nil farmID is a level-1 issue of the body,
// so only the role is checked then.
func checkDeliveryWriter(p auth.Principal, farmID uuid.UUID) error {
	if !p.Role.In(auth.Manager, auth.Owner) {
		return httpx.NewError(http.StatusForbidden, httpx.CodeForbidden, messages.PageForbidden())
	}
	if farmID != uuid.Nil && !p.CanSeeFarm(farmID) {
		return httpx.FarmForbidden()
	}
	return nil
}

// checkDelivery is checkDeliveryWriter, then the price fields (§6.7, M4 Q9):
// any pricePerKg or totalLei member, null included, is 403 PRICE_FORBIDDEN
// for a manager and 400 VALIDATION unknown_field for the owner (a delivery
// is created without a price; only Prețuri sets one).
func checkDelivery(p auth.Principal, req DeliveryCreateRequest) error {
	if err := checkDeliveryWriter(p, req.FarmID); err != nil {
		return err
	}
	var v httpx.Validator
	for _, f := range []struct {
		path string
		raw  []byte
	}{{"pricePerKg", req.PricePerKg}, {"totalLei", req.TotalLei}} {
		if f.raw == nil { // absent; a JSON null is "null"
			continue
		}
		if p.Role != auth.Owner {
			return httpx.PriceForbidden()
		}
		v.Add(f.path, httpx.IssueUnknownField, messages.InvalidData())
	}
	return v.Err()
}

// DeliveryRefs reads what DeliveryCreateRequest.Validate names: the farm's
// name and the type's. An unknown farm is 404 NOT_FOUND (M2 D4: only an
// owner gets this far with one). A request without farmId reads nothing.
func (s *Service) DeliveryRefs(ctx context.Context, req DeliveryCreateRequest) (DeliveryRefs, error) {
	if req.FarmID == uuid.Nil {
		return DeliveryRefs{}, nil
	}
	row, err := s.store.Q().EntriesDeliveryRefs(ctx, dbq.EntriesDeliveryRefsParams{FarmID: req.FarmID, FeedTypeID: req.FeedTypeID})
	if errors.Is(err, pgx.ErrNoRows) {
		return DeliveryRefs{}, httpx.NotFound()
	}
	if err != nil {
		return DeliveryRefs{}, fmt.Errorf("entries: delivery refs: %w", err)
	}
	return DeliveryRefs{FarmName: row.FarmName, FeedTypeName: row.FeedTypeName}, nil
}

// CreateDelivery saves one Livrare hrană (§6.7, §7.8, M4 D5) on the save
// runner. req must have passed Validate (level 1). Inside the transaction,
// after the farm lock and in this order:
//   - a future date → 422 REFUSED (§3.2);
//   - a day inside a cycle of a closed series → 409 SERIES_CLOSED (§3.1);
//   - a retired type → 422 REFUSED "Breeder 1 e retras. Alege alt tip.";
//   - a dry run answers 200 {effects: [], warnings: [], refusals: []} here,
//     before any write (a delivery has no effect or warning to preview);
//   - the entry and its feed_delivery row (no price, Q9), then the stock of
//     the type at the end of the delivery date (Q13) for the Toast.
//
// Two deliveries of one type and day are legal: there is no uniqueness key,
// so no OnConflict. A manager may enter past days without a Motiv (§3.2).
func (s *Service) CreateDelivery(ctx context.Context, p auth.Principal, req DeliveryCreateRequest) (Result, error) {
	if err := checkDelivery(p, req); err != nil {
		return Result{}, err
	}
	return s.runner.Create(ctx, p, Create{
		Route: routeDelivery, Kind: dbq.EntryKindDelivery, FarmIDs: []uuid.UUID{req.FarmID}, Meta: req.WriteMeta,
	}, func(ctx context.Context, tx *Tx) (Saved, error) {
		return planDelivery(ctx, tx, req)
	})
}

// planDelivery is the form's part of the save (§12.5 last note).
func planDelivery(ctx context.Context, tx *Tx, req DeliveryCreateRequest) (Saved, error) {
	refuse := func(r httpx.Refusal) (Saved, error) {
		if tx.DryRun {
			return Saved{}, tx.Preview(NewDryRun(nil, nil, []httpx.Refusal{r}))
		}
		return Saved{}, httpx.RefusedWith([]httpx.Refusal{r}, nil)
	}
	if req.QtyKg == nil { // Validate guarantees it; never insert a delivery without kg
		return Saved{}, httpx.Validation(httpx.Issue{Path: "qtyKg", Code: httpx.IssueRequired, Message: messages.InvalidData()})
	}
	if req.Date.After(tx.Today) { // §3.2, every role
		return refuse(httpx.Refusal{Field: "date", Message: messages.FutureDate(req.Date, tx.Today)})
	}
	dc, err := tx.Q.EntriesDeliverySaveContext(ctx, dbq.EntriesDeliverySaveContextParams{
		FarmID: req.FarmID, Day: req.Date, FeedTypeID: req.FeedTypeID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Saved{}, httpx.NotFound()
	}
	if err != nil {
		return Saved{}, fmt.Errorf("entries: delivery context: %w", err)
	}
	if dc.ClosedSeries != "" {
		return Saved{}, httpx.SeriesClosed(dc.ClosedSeries)
	}
	if dc.FeedTypeName == nil || dc.FeedTypeActive == nil { // Validate checked it exists; types are never deleted
		return Saved{}, httpx.Validation(httpx.Issue{Path: "feedTypeId", Code: httpx.IssueOutOfRange, Message: messages.InvalidData()})
	}
	typeName := *dc.FeedTypeName
	if !*dc.FeedTypeActive {
		return refuse(httpx.Refusal{Field: "feedTypeId", Message: messages.FeedTypeRetired(typeName)})
	}
	if tx.DryRun {
		return Saved{}, tx.Preview(NewDryRun(nil, nil, nil))
	}

	qty := req.QtyKg.Decimal()
	note := req.note()
	ins, err := tx.InsertEntry(ctx, Header{FarmID: req.FarmID, Date: req.Date})
	if err != nil {
		return Saved{}, err
	}
	if err := tx.Q.EntriesInsertFeedDelivery(ctx, dbq.EntriesInsertFeedDeliveryParams{
		EntryID: ins.ID, FarmID: req.FarmID, FeedTypeID: req.FeedTypeID, Date: req.Date, QtyKg: qty, Note: note,
	}); err != nil {
		return Saved{}, err
	}
	// The stock of the type at the end of the delivery date, after the insert
	// (§6.7, Q13): the one point of a one-day series.
	series, err := calc.StockSeries(ctx, tx.Tx, req.FarmID, []uuid.UUID{req.FeedTypeID}, req.Date, req.Date)
	if err != nil {
		return Saved{}, err
	}
	pts := series[req.FeedTypeID]
	if len(pts) == 0 {
		return Saved{}, fmt.Errorf("entries: no stock point for feed type %s on %s", req.FeedTypeID, req.Date)
	}
	stockAfter := pts[len(pts)-1].Kg

	summary := messages.DeliverySummary(qty, typeName)
	at := timex.At(tx.Now)
	return Saved{
		Response: DeliveryResponse{
			Entry: ins.Ref(), Toast: messages.DeliveryAdded(qty, typeName, dc.FarmName, stockAfter),
			StockAfter: num.Of(stockAfter), Warnings: []httpx.Warning{}, AsOf: at,
		},
		Audit: audit.Row{
			Action: dbq.AuditActionCreate, Entity: "entry", EntityID: ins.ID, EntryID: &ins.ID, EntryKind: &ins.Kind,
			RecordTypes: []string{recordTypeDelivery}, FarmIDs: []uuid.UUID{req.FarmID}, RecordDate: &req.Date,
			Summary: &summary,
			NewValues: deliveryAuditValues{
				FeedTypeID: req.FeedTypeID, Type: typeName, Date: req.Date, Kg: num.Of(qty), Note: note,
			},
		},
		Events: []live.Event{
			live.EntryCreated(live.EntryPayload{
				EntryID: ins.ID, Kind: ins.Kind, FarmIDs: []uuid.UUID{req.FarmID}, Dates: []timex.Day{req.Date},
				RecordTypes: []string{recordTypeDelivery}, Author: &tx.Principal.Name, At: at, Summary: summary,
			}),
			live.DeliveryCreated(live.DeliveryPayload{
				EntryID: ins.ID, FarmID: req.FarmID, FeedType: typeName, Kg: num.Of(qty), Date: req.Date,
			}),
		},
	}, nil
}
