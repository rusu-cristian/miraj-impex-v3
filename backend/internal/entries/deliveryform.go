package entries

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db/dbq"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// deliveryForm is GET /api/forms/delivery?farmId (§7.8, M4 Q14). farmId is
// required; any other parameter (a date too) is a 400 VALIDATION issue.
func (h *Handler) deliveryForm(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	farmID, err := parseDeliveryFormQuery(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.DeliveryForm(r.Context(), p, farmID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseDeliveryFormQuery reads ?farmId with parseFormQuery's codes: missing
// → required, repeated or a malformed uuid → bad_type, any other parameter
// → unknown_field (in name order).
func parseDeliveryFormQuery(values url.Values) (uuid.UUID, error) {
	var v httpx.Validator
	var farmID uuid.UUID
	if s, ok := requiredParam(&v, values, paramFarmID); ok {
		id, err := uuid.Parse(s)
		if v.Check(err == nil, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
			farmID = id
		}
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(name == paramFarmID, name, httpx.IssueUnknownField, messages.InvalidData())
	}
	return farmID, v.Err()
}

// DeliveryForm is the Livrare hrană form state of farmID for p (§7.8, M4
// Q14), read in one snapshot (InReadTx, M4 D10). Errors are envelope errors:
// 403 FORBIDDEN for a vet or a viewer, 403 FARM_FORBIDDEN for a farm outside
// the caller's farms, 404 NOT_FOUND for an owner's unknown farm.
func (s *Service) DeliveryForm(ctx context.Context, p auth.Principal, farmID uuid.UUID) (DeliveryFormResponse, error) {
	if err := checkDeliveryWriter(p, farmID); err != nil {
		return DeliveryFormResponse{}, err
	}
	if !p.CanSeeFarm(farmID) { // checkDeliveryWriter leaves the nil farm id to a create's level 1
		return DeliveryFormResponse{}, httpx.FarmForbidden()
	}
	now := s.clock.Now()
	var res DeliveryFormResponse
	err := s.store.InReadTx(ctx, func(tx dbq.DBTX) (err error) {
		res, err = BuildDeliveryForm(ctx, tx, DeliveryFormInput{FarmID: farmID, Today: timex.DayOf(now), Now: now})
		return err
	})
	if errors.Is(err, ErrFormFarmNotFound) {
		return DeliveryFormResponse{}, httpx.NotFound()
	}
	return res, err
}

// DeliveryFormInput is what BuildDeliveryForm reads: the farm and the
// clock's day and instant (ARC-6: never app_today()).
type DeliveryFormInput struct {
	FarmID uuid.UUID
	Today  timex.Day
	Now    time.Time
}

// BuildDeliveryForm builds GET /api/forms/delivery (§7.8, M4 Q14) on db for
// in.FarmID: the head (farm, open cycle, active feed types), then, for a farm
// with an open cycle, the settings and the stock as of in.Today
// (calc.ReadStockDays, as the feed form): at most 5 statements, to be run in
// one read snapshot (the service's InReadTx).
//
// It doesn't check the caller (the service does); an unknown farm is
// ErrFormFarmNotFound.
func BuildDeliveryForm(ctx context.Context, db dbq.DBTX, in DeliveryFormInput) (DeliveryFormResponse, error) {
	head, err := dbq.New(db).FormsDeliveryHead(ctx, in.FarmID)
	if errors.Is(err, pgx.ErrNoRows) {
		return DeliveryFormResponse{}, ErrFormFarmNotFound
	}
	if err != nil {
		return DeliveryFormResponse{}, fmt.Errorf("entries: delivery form head: %w", err)
	}
	res := DeliveryFormResponse{
		Farm: FormFarm{ID: head.ID, Name: head.Name, Type: head.Type}, Today: in.Today,
		FeedTypes: make([]FormFeedType, 0, len(head.TypeIds)), Stock: []FormStock{}, AsOf: timex.At(in.Now),
	}
	for i, id := range head.TypeIds {
		res.FeedTypes = append(res.FeedTypes, FormFeedType{ID: id, Name: head.TypeNames[i], ForSex: forSex(head.TypeSexes[i])})
	}
	if head.CycleID == nil {
		return res, nil
	}
	age := calc.SeriesAge{PlacementDate: *head.PlacementDate, AgeAtPlacementDays: int(*head.AgeAtPlacementDays)}
	res.Cycle = &DeliveryFormCycle{
		ID: *head.CycleID, Series: *head.SeriesName, StartDate: *head.CycleStart, SaptToday: calc.SAPT(age, in.Today),
	}
	if res.Stock, err = formStock(ctx, db, FeedFormInput{FarmID: in.FarmID, Today: in.Today}); err != nil {
		return DeliveryFormResponse{}, err
	}
	return res, nil
}
