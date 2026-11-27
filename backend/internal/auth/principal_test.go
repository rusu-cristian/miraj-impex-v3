package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCanSeeFarm(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	owner := Principal{Role: Owner}
	vet := Principal{Role: Vet, FarmIDs: []uuid.UUID{a}}
	none := Principal{Role: Manager, FarmIDs: []uuid.UUID{}}
	nilFarms := Principal{Role: Viewer} // never built this way, but must not mean "all"

	for _, id := range []uuid.UUID{a, b, uuid.New()} {
		if !owner.CanSeeFarm(id) {
			t.Errorf("owner can't see %s", id)
		}
	}
	if !vet.CanSeeFarm(a) || vet.CanSeeFarm(b) {
		t.Error("vet sees exactly its listed farm")
	}
	if none.CanSeeFarm(a) || nilFarms.CanSeeFarm(a) {
		t.Error("a non-owner without farms sees none")
	}
}

func TestRequireRole(t *testing.T) {
	h := RequireRole(Manager, Owner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	tests := []struct {
		name   string
		p      *Principal
		status int
		code   string // the whole body when refused
	}{
		{"owner", &Principal{Role: Owner}, http.StatusTeapot, ""},
		{"manager", &Principal{Role: Manager}, http.StatusTeapot, ""},
		{"vet", &Principal{Role: Vet}, http.StatusForbidden, `{"error":{"code":"FORBIDDEN","message":"Nu ai acces la această pagină."}}`},
		{"viewer", &Principal{Role: Viewer}, http.StatusForbidden, `{"error":{"code":"FORBIDDEN","message":"Nu ai acces la această pagină."}}`},
		{"no principal", nil, http.StatusUnauthorized, `{"error":{"code":"UNAUTHENTICATED","message":"Sesiunea a expirat. Intră din nou în cont."}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/x", nil)
			if tt.p != nil {
				r = r.WithContext(WithPrincipal(r.Context(), *tt.p))
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Errorf("status %d, want %d", w.Code, tt.status)
			}
			if tt.code != "" && strings.TrimSuffix(w.Body.String(), "\n") != tt.code {
				t.Errorf("body %s", w.Body)
			}
		})
	}
}

func TestRoleIn(t *testing.T) {
	if !Vet.In(Vet, Owner) || Viewer.In(Vet, Owner) || Owner.In() {
		t.Error("Role.In")
	}
}
