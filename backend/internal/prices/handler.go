package prices

import (
	"net/http"
	"net/url"
	"slices"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Handler serves the prices routes.
type Handler struct {
	svc *Service
}

// NewHandler builds the prices handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers the Prețuri routes on mux behind authed (the session
// chain, SPEC §12.3). All are the owner's (§3.3, M4 Q10), checked before the
// body is read: the S12 list refuses every other role with 403 FORBIDDEN;
// the writes refuse a manager with 403 PRICE_FORBIDDEN ("○ sees, cannot
// set") and the vet and viewer with 403 FORBIDDEN.
func (h *Handler) Routes(mux httpx.Router, authed httpx.Chain) {
	list := authed.Append(auth.RequireRole(auth.Owner))
	writes := authed.Append(requirePriceWriter)
	mux.Handle("GET /api/deliveries", list.ThenFunc(h.unpriced))
	mux.Handle("POST /api/deliveries/prices", writes.ThenFunc(h.set))
	mux.Handle("POST /api/deliveries/prices/undo", writes.ThenFunc(h.undo))
	mux.Handle("PATCH /api/deliveries/{id}/price", writes.ThenFunc(h.change))
}

// requirePriceWriter lets only the owner through to a price write (checkOwner).
func requirePriceWriter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := auth.PrincipalFrom(r.Context())
		if !ok {
			httpx.WriteError(w, httpx.Unauthenticated())
			return
		}
		if err := checkOwner(p, true); err != nil {
			httpx.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// unpriced is GET /api/deliveries?priced=false (§7.13).
func (h *Handler) unpriced(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	if err := parseUnpricedQuery(r.URL.Query()); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Unpriced(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// parseUnpricedQuery checks the S12 list's query: exactly priced=false (the
// only list §7.13 has). priced absent → required; another value (or more than
// one) → out_of_range; any other parameter → unknown_field (in name order).
func parseUnpricedQuery(values url.Values) error {
	var v httpx.Validator
	switch priced, ok := values["priced"]; {
	case !ok:
		v.Add("priced", httpx.IssueRequired, messages.InvalidData())
	case len(priced) != 1 || priced[0] != "false":
		v.Add("priced", httpx.IssueOutOfRange, messages.InvalidData())
	}
	names := make([]string, 0, len(values))
	for name := range values {
		if name != "priced" {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	for _, name := range names {
		v.Add(name, httpx.IssueUnknownField, messages.InvalidData())
	}
	return v.Err()
}

// set is POST /api/deliveries/prices (§7.13, M4 Q10): decode → the asked
// deliveries' refs (names for the messages) → Validate → service.
func (h *Handler) set(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFrom(r.Context()) // requirePriceWriter checked it
	req, err := httpx.Decode[SetRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	ids := make([]uuid.UUID, 0, len(req.Items))
	for _, it := range req.Items {
		if it.ID != uuid.Nil {
			ids = append(ids, it.ID)
		}
	}
	refs, err := h.svc.Refs(r.Context(), ids)
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
	res, err := h.svc.Set(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// undo is POST /api/deliveries/prices/undo (§7.13, M4 D4, Q11).
func (h *Handler) undo(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFrom(r.Context()) // requirePriceWriter checked it
	req, err := httpx.Decode[UndoRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.Validate(req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Undo(r.Context(), p, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}

// change is PATCH /api/deliveries/{id}/price (§7.13, M4 Q12): the path id
// (malformed → 400 bad_type on `id`) → decode → the delivery's ref (unknown
// or deleted → 404) → Validate → service.
func (h *Handler) change(w http.ResponseWriter, r *http.Request) {
	p, _ := auth.PrincipalFrom(r.Context()) // requirePriceWriter checked it
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, httpx.Validation(httpx.Issue{Path: "id", Code: httpx.IssueBadType, Message: messages.InvalidData()}))
		return
	}
	req, err := httpx.Decode[ChangeRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	refs, err := h.svc.Refs(r.Context(), []uuid.UUID{id})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	ref, ok := refs[id]
	if !ok || ref.Deleted {
		httpx.WriteError(w, httpx.NotFound())
		return
	}
	var v httpx.Validator
	req.Validate(&v, ref)
	if err := v.Err(); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Change(r.Context(), p, id, req)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, res)
}
