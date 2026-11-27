package auth

import (
	"net/http"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Handler serves the auth routes of SPEC §7.1.
type Handler struct {
	svc *Service
}

// NewHandler builds the auth handler.
func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Routes registers POST /api/auth/login on public (the only /api route
// without a session) and POST /api/auth/logout and GET /api/auth/me on
// authed, the chain that ends in the session check (RequireSession). Every
// other module takes only authed (SPEC §12.3).
func (h *Handler) Routes(mux httpx.Router, public, authed httpx.Chain) {
	mux.Handle("POST /api/auth/login", public.ThenFunc(h.login))
	mux.Handle("POST /api/auth/logout", authed.ThenFunc(h.logout))
	mux.Handle("GET /api/auth/me", authed.ThenFunc(h.me))
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.Decode[loginRequest](w, r, httpx.MaxJSONBody)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	if err := httpx.Validate(req); err != nil {
		httpx.WriteError(w, err)
		return
	}
	res, err := h.svc.Login(r.Context(), LoginInput{
		Login: req.Login, Password: req.Password, IP: httpx.ClientIPFrom(r.Context()), UserAgent: r.UserAgent(),
	})
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.SetUserID(w, res.User.ID)
	http.SetCookie(w, sessionCookie(res.Token))
	httpx.JSON(w, http.StatusOK, loginResponse{User: toUserDTO(res.User), Farms: toFarmDTOs(res.Farms), Today: res.Today})
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	p, ok := PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	if err := h.svc.Logout(r.Context(), p); err != nil {
		httpx.WriteError(w, err)
		return
	}
	http.SetCookie(w, clearCookie())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	p, ok := PrincipalFrom(r.Context())
	if !ok {
		httpx.WriteError(w, httpx.Unauthenticated())
		return
	}
	me, err := h.svc.Me(r.Context(), p)
	if err != nil {
		httpx.WriteError(w, err)
		return
	}
	httpx.JSON(w, http.StatusOK, meResponse{
		User: toUserDTO(me.User), Farms: toFarmDTOs(me.Farms),
		Today: timex.DayOf(me.Now), Now: timex.At(me.Now), StartPage: startPage,
	})
}
