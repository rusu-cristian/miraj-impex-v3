package reports

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

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
// Rapoarte (§3.3). The farm check and the role's response type (prices for
// manager and owner only, M4 D2) are the service's.
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	reports := authed.Append(auth.RequireRole(auth.Viewer, auth.Manager, auth.Owner))
	mux.Handle("GET /api/reports/feed", reports.ThenFunc(h.feed))
}

// Query parameters of GET /api/reports/feed.
const (
	paramFarmID = "farmId"
	paramPeriod = "period"
	paramCursor = "cursor"
	paramLimit  = "limit"
	farmIDAll   = "all"
)

// FeedQuery is a GET /api/reports/feed request: ?farmId=all|<uuid> (absent:
// all; FarmID nil), ?period= (zero: absent, read as cycle), ?cursor= (nil:
// the first page) and ?limit= (0: absent, DefaultLimit).
type FeedQuery struct {
	FarmID *uuid.UUID
	Period calc.Period
	Cursor *Cursor
	Limit  int
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
	res, err := h.svc.Feed(r.Context(), p, q)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseFeedQuery reads the query string of GET /api/reports/feed. Every
// problem is a 400 VALIDATION issue (the §7.0a codes, M2 D2): an unknown
// parameter → unknown_field (in name order), a repeated one, a farmId that
// is neither "all" nor a uuid, a cursor that is not one (ParseCursor) or a
// limit that is not an integer → bad_type, a limit outside 1–MaxLimit →
// out_of_range; `period` is read by the PeriodTable's own rules (§5.2,
// calc.ParsePeriod).
func parseFeedQuery(values url.Values) (FeedQuery, error) {
	var v httpx.Validator
	var q FeedQuery
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		known := name == paramFarmID || name == paramPeriod || name == paramCursor || name == paramLimit
		v.Check(known, name, httpx.IssueUnknownField, messages.InvalidData())
	}
	if vals, ok := values[paramFarmID]; ok && v.Check(len(vals) == 1, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
		if vals[0] != farmIDAll {
			id, err := uuid.Parse(vals[0])
			if v.Check(err == nil, paramFarmID, httpx.IssueBadType, messages.InvalidData()) {
				q.FarmID = &id
			}
		}
	}
	if vals, ok := values[paramPeriod]; ok && v.Check(len(vals) == 1, paramPeriod, httpx.IssueBadType, messages.InvalidData()) {
		q.Period = calc.ParsePeriod(&v, paramPeriod, vals[0])
	}
	if vals, ok := values[paramCursor]; ok && v.Check(len(vals) == 1, paramCursor, httpx.IssueBadType, messages.InvalidData()) {
		c, err := ParseCursor(vals[0])
		if v.Check(err == nil, paramCursor, httpx.IssueBadType, messages.InvalidData()) {
			q.Cursor = &c
		}
	}
	if vals, ok := values[paramLimit]; ok && v.Check(len(vals) == 1, paramLimit, httpx.IssueBadType, messages.InvalidData()) {
		n, err := strconv.Atoi(vals[0])
		if v.Check(err == nil, paramLimit, httpx.IssueBadType, messages.InvalidData()) &&
			v.Check(n >= 1 && n <= MaxLimit, paramLimit, httpx.IssueOutOfRange, messages.InvalidData()) {
			q.Limit = n
		}
	}
	return q, v.Err()
}
