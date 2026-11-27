package entries

import (
	"net/http"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// Handler serves the entries routes.
type Handler struct {
	svc *Service
}

// NewHandler builds the entries handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the entries routes on mux behind authed (the session
// chain, SPEC §12.3). Every route checks the role and the farm itself
// (checkWriter: vet, manager, owner); the creates do it after decoding (M3
// D5: decode → role → farm → codes → Validate → service).
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	// Form states (§7.8, feedform.go).
	mux.Handle("GET /api/forms/feed-daily", authed.ThenFunc(h.feedForm))
	// Creates.
	mux.Handle(routeFeedDaily, authed.ThenFunc(h.createFeedDaily))
}

// createFeedDaily is POST /api/entries/feed-daily (§7.8, M3 D5). The answer
// is sent byte for byte as the runner encoded it (the bytes a replay sends).
func (h *Handler) createFeedDaily(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	req, err := httpx.Decode[FeedDailyCreateRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := checkWriter(p, req.FarmID); err != nil { // before any read: no section code of another farm leaks
		httpx.WriteError(w, err)
		return
	}
	refs, err := h.svc.FeedRefs(r.Context(), req)
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
	res, err := h.svc.CreateFeedDaily(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.WriteRaw(w, res.Status, res.Body)
}
