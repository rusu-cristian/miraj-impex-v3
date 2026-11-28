package series

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Handler serves the series routes.
type Handler struct {
	svc *Service
}

// NewHandler builds the series handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the series routes on mux behind authed (the session
// chain, SPEC §12.3): GET /api/series/{id}/page for the viewer, manager and
// owner (vets never reach the scope pages, §3.3: 403 FORBIDDEN); which farms
// the page shows is the builder's (M4 Q4, Q5).
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	pages := authed.Append(auth.RequireRole(auth.Viewer, auth.Manager, auth.Owner))
	mux.Handle("GET /api/series/{id}/page", pages.ThenFunc(h.page))
}

func (h *Handler) page(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	id, err := parsePageQuery(r.PathValue("id"), r.URL.Query())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Page(r.Context(), p, id)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parsePageQuery reads the path id of the series page, which takes no query
// parameter. Every problem is a 400 VALIDATION issue (§7.0a, as the farm
// page): a malformed uuid → bad_type on `id`, any parameter → unknown_field
// (in name order).
func parsePageQuery(id string, values url.Values) (uuid.UUID, error) {
	var v httpx.Validator
	seriesID, err := uuid.Parse(id)
	v.Check(err == nil, "id", httpx.IssueBadType, messages.InvalidData())
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		v.Check(false, name, httpx.IssueUnknownField, messages.InvalidData())
	}
	return seriesID, v.Err()
}
