package reports

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/calc"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Handler serves the reports routes.
type Handler struct {
	svc *Service
}

// NewHandler builds the reports handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers GET /api/reports/feed on mux behind authed (the session
// chain, SPEC §12.3) for viewer, manager and owner: vets never reach
// Rapoarte (§3.3). The farm check is the service's (M2 D4).
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	reports := authed.Append(auth.RequireRole(auth.Viewer, auth.Manager, auth.Owner))
	mux.Handle("GET /api/reports/feed", reports.ThenFunc(h.feed))
}

// Query parameters of GET /api/reports/feed.
const (
	paramFarmID = "farmId"
	paramPeriod = "period"
	farmIDAll   = "all"
)

// feedQuery is a GET /api/reports/feed request: ?farmId=all|<uuid> (absent:
// all; farmID nil) and ?period= (zero: absent).
type feedQuery struct {
	farmID *uuid.UUID
	period calc.Period
}

func (h *Handler) feed(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	q, err := parseFeedQuery(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Feed(r.Context(), p, q.farmID, q.period)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseFeedQuery reads the query string of GET /api/reports/feed. Every
// problem is a 400 VALIDATION issue (the §7.0a codes, M2 D2): an unknown
// parameter → unknown_field (in name order), a repeated one or a farmId that
// is neither "all" nor a uuid → bad_type; `period` is read by the
// PeriodTable's own rules (§5.2, calc.ParsePeriod).
func parseFeedQuery(values url.Values) (feedQuery, error) {
	var v httpx.Validator
	var q feedQuery
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(name == paramFarmID || name == paramPeriod, name, httpx.IssueUnknownField, messages.InvalidData())
	}
	if vals, ok := values[paramFarmID]; ok && v.Check(len(vals) == 1, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
		if vals[0] != farmIDAll {
			id, err := uuid.Parse(vals[0])
			if v.Check(err == nil, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
				q.farmID = &id
			}
		}
	}
	if vals, ok := values[paramPeriod]; ok && v.Check(len(vals) == 1, paramPeriod, httpx.IssueBadType, messages.InvalidData()) {
		q.period = calc.ParsePeriod(&v, paramPeriod, vals[0])
	}
	return q, v.Err()
}
