package farms

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// server serves the farms routes, the caller being p (nil: no session
// middleware ran).
func (f *fixture) server(p *auth.Principal) http.Handler {
	mux := http.NewServeMux()
	withCaller := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if p != nil {
				r = r.WithContext(auth.WithPrincipal(r.Context(), *p))
			}
			next.ServeHTTP(w, r)
		})
	}
	svc := NewService(db.NewStore(f.d.App), f.d.Clock)
	NewHandler(svc).Routes(mux, httpx.NewChain(withCaller))
	return mux
}

type errorBody struct {
	Error struct {
		Code    string        `json:"code"`
		Message string        `json:"message"`
		Issues  []httpx.Issue `json:"issues"`
	} `json:"error"`
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestPageHandler(t *testing.T) {
	f := newFixture(t)
	hirbovat, telita := f.hirbovat.String(), f.telita.String()
	vet := principal(auth.Vet, f.hirbovat)
	viewer := principal(auth.Viewer, f.telita)
	manager := principal(auth.Manager, f.hirbovat)
	own := owner()

	for _, tc := range []struct {
		name      string
		caller    *auth.Principal
		path      string
		status    int
		code      string
		issuePath string
		issueCode string
	}{
		{"no session", nil, "/api/farms/" + hirbovat + "/page", 401, "UNAUTHENTICATED", "", ""},
		{"vet", &vet, "/api/farms/" + hirbovat + "/page", 403, "FORBIDDEN", "", ""},
		{"viewer on another farm", &viewer, "/api/farms/" + hirbovat + "/page", 403, "FARM_FORBIDDEN", "", ""},
		{"viewer on an unknown farm", &viewer, "/api/farms/" + uuid.NewString() + "/page", 403, "FARM_FORBIDDEN", "", ""},
		{"owner on an unknown farm", &own, "/api/farms/" + uuid.NewString() + "/page", 404, "NOT_FOUND", "", ""},
		{"another farm's cycle", &own, "/api/farms/" + hirbovat + "/page?cycleId=" + f.cycT.String(), 404, "NOT_FOUND", "", ""},
		{"malformed farm id", &own, "/api/farms/hirbovat/page", 400, "VALIDATION", "id", "bad_type"},
		{"malformed cycle id", &manager, "/api/farms/" + hirbovat + "/page?cycleId=x", 400, "VALIDATION", "cycleId", "bad_type"},
		{"repeated cycle id", &manager, "/api/farms/" + hirbovat + "/page?cycleId=" + f.cycleH26.String() + "&cycleId=" + f.cycleH25.String(),
			400, "VALIDATION", "cycleId", "bad_type"},
		{"unknown parameter", &manager, "/api/farms/" + hirbovat + "/page?period=cycle", 400, "VALIDATION", "period", "unknown_field"},
		{"viewer on own farm", &viewer, "/api/farms/" + telita + "/page", 200, "", "", ""},
		{"manager, past cycle", &manager, "/api/farms/" + hirbovat + "/page?cycleId=" + f.cycleH25.String(), 200, "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := get(t, f.server(tc.caller), tc.path)
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.status, rec.Body)
			}
			if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", cc)
			}
			if tc.status == 200 {
				var res PageResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
					t.Fatal(err)
				}
				if res.Cycle == nil {
					t.Errorf("no cycle shown: %s", rec.Body)
				}
				return
			}
			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Errorf("code = %s, want %s", body.Error.Code, tc.code)
			}
			if tc.issuePath != "" {
				if len(body.Error.Issues) != 1 || body.Error.Issues[0].Path != tc.issuePath || body.Error.Issues[0].Code != tc.issueCode {
					t.Errorf("issues = %+v, want one %s %s", body.Error.Issues, tc.issuePath, tc.issueCode)
				}
			}
		})
	}

	rec := get(t, f.server(&viewer), "/api/farms/"+hirbovat+"/page")
	if want := `{"error":{"code":"FARM_FORBIDDEN","message":"Nu ai acces la această fermă."}}` + "\n"; rec.Body.String() != want {
		t.Errorf("FARM_FORBIDDEN body = %s, want %s", rec.Body, want)
	}
}

func TestBootstrapHandler(t *testing.T) {
	f := newFixture(t)
	for _, role := range []auth.Role{auth.Vet, auth.Viewer, auth.Manager, auth.Owner} {
		t.Run(string(role), func(t *testing.T) {
			p := principal(role, f.hirbovat)
			if role == auth.Owner {
				p = owner()
			}
			rec := get(t, f.server(&p), "/api/bootstrap")
			if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status %d, Cache-Control %q: %s", rec.Code, rec.Header().Get("Cache-Control"), rec.Body)
			}
			var res BootstrapResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatal(err)
			}
			want := 1
			if role == auth.Owner {
				want = 3
			}
			if len(res.Farms) != want {
				t.Errorf("farms = %d, want %d", len(res.Farms), want)
			}
		})
	}

	own := owner()
	if rec := get(t, f.server(&own), "/api/bootstrap?x=1"); rec.Code != 400 {
		t.Errorf("unknown parameter: status %d, want 400: %s", rec.Code, rec.Body)
	}
	if rec := get(t, f.server(nil), "/api/bootstrap"); rec.Code != 401 {
		t.Errorf("no session: status %d, want 401", rec.Code)
	}
}
