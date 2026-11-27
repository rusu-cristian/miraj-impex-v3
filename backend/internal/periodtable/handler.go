package periodtable

import (
	"net/http"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// Handler serves GET /api/table.
type Handler struct {
	svc *Service
}

// NewHandler builds the PeriodTable handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers GET /api/table on mux behind authed (the session chain,
// SPEC §12.3) for viewer, manager and owner: vets never reach the scope
// tables (§3.3, M2 D4). The farm check is the service's, before Build.
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	tables := authed.Append(auth.RequireRole(auth.Viewer, auth.Manager, auth.Owner))
	mux.Handle("GET /api/table", tables.ThenFunc(h.table))
}

func (h *Handler) table(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	q, err := ParseQuery(r.URL.Query())
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.Validate(q); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Table(r.Context(), p, q)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}
