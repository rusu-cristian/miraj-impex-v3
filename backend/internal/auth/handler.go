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

// Routes registers POST /api/auth/login (no session), POST /api/auth/logout
// and GET /api/auth/me (session required), each wrapped in mw.
func (h *Handler) Routes(mux httpx.Router, mw httpx.Chain) {
	mux.Handle("POST /api/auth/login", mw.ThenFunc(h.login))
	signedIn := mw.Append(h.svc.RequireSession)
	mux.Handle("POST /api/auth/logout", signedIn.ThenFunc(h.logout))
	mux.Handle("GET /api/auth/me", signedIn.ThenFunc(h.me))
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
