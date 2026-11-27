package farms

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

// Handler serves the farms routes.
type Handler struct {
	svc *Service
}

// NewHandler builds the farms handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the farms routes on mux behind authed (the session
// chain, SPEC §12.3):
//   - GET /api/farms/{id}/page and GET /api/farms/{id}/summary: viewer,
//     manager, owner (vets never reach the scope pages, §3.3); the farm
//     check is the service's (M2 D4);
//   - GET /api/bootstrap: every role (M2 D7).
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	pages := authed.Append(auth.RequireRole(auth.Viewer, auth.Manager, auth.Owner))
	mux.Handle("GET /api/farms/{id}/page", pages.ThenFunc(h.page))
	mux.Handle("GET /api/farms/{id}/summary", pages.ThenFunc(h.summary))
	mux.Handle("GET /api/bootstrap", authed.ThenFunc(h.bootstrap))
}

// Query parameters of the farm routes.
const (
	paramCycleID = "cycleId"
	paramPeriod  = "period"
)

// pageQuery is a farm route's request: the path id, ?cycleId= and, for the
// summary, ?period=.
type pageQuery struct {
	farmID  uuid.UUID
	cycleID *uuid.UUID
	period  calc.Period
}

func (h *Handler) page(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	q, err := parsePageQuery(r.PathValue("id"), r.URL.Query(), paramCycleID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Page(r.Context(), p, q.farmID, q.cycleID)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	q, err := parsePageQuery(r.PathValue("id"), r.URL.Query(), paramCycleID, paramPeriod)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Summary(r.Context(), p, q.farmID, q.cycleID, q.period)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

func (h *Handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	if err := checkParams(r.URL.Query()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Bootstrap(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parsePageQuery reads the path id and the allowed query parameters of a
// farm route. Every problem is a 400 VALIDATION issue (the §7.0a codes, M2
// D2/D4): an unknown parameter → unknown_field, a repeated one or a malformed
// uuid → bad_type; `period` is read by the PeriodTable's own rules (§5.2,
// calc.ParsePeriod), so both routes accept the same values.
func parsePageQuery(id string, values url.Values, allowed ...string) (pageQuery, error) {
	var v httpx.Validator
	var q pageQuery
	farmID, err := uuid.Parse(id)
	if v.Check(err == nil, "id", httpx.IssueBadType, messages.InvalidData()) {
		q.farmID = farmID
	}
	checkKnown(&v, values, allowed...)
	if s, ok := single(&v, values, paramCycleID); ok {
		cycleID, err := uuid.Parse(s)
		if v.Check(err == nil, paramCycleID, httpx.IssueBadType, messages.InvalidData()) {
			q.cycleID = &cycleID
		}
	}
	if slices.Contains(allowed, paramPeriod) {
		if s, ok := single(&v, values, paramPeriod); ok {
			q.period = calc.ParsePeriod(&v, paramPeriod, s)
		}
	}
	return q, v.Err()
}

// checkParams refuses any query parameter outside allowed.
func checkParams(values url.Values, allowed ...string) error {
	var v httpx.Validator
	checkKnown(&v, values, allowed...)
	return v.Err()
}

// checkKnown adds an unknown_field issue per parameter outside allowed, in
// name order (a url.Values has no order of its own).
func checkKnown(v *httpx.Validator, values url.Values, allowed ...string) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(slices.Contains(allowed, name), name, httpx.IssueUnknownField, messages.InvalidData())
	}
}

// single is the one value of parameter name, false when it is absent; a
// repeated parameter is a bad_type issue.
func single(v *httpx.Validator, values url.Values, name string) (string, bool) {
	vals, ok := values[name]
	if !ok {
		return "", false
	}
	if !v.Check(len(vals) == 1, name, httpx.IssueBadType, messages.InvalidData()) {
		return "", false
	}
	return vals[0], true
}
