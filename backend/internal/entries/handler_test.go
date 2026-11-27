package entries

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/auth"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
)

// post sends body to POST /api/entries/feed-daily as p, through Routes.
func (f *feedFixture) post(p auth.Principal, body string) *httptest.ResponseRecorder {
	f.t.Helper()
	mux := http.NewServeMux()
	NewHandler(f.svc).Routes(mux, httpx.NewChain())
	req := httptest.NewRequestWithContext(auth.WithPrincipal(context.Background(), p), http.MethodPost,
		"/api/entries/feed-daily", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func (f *feedFixture) body(farm uuid.UUID, rows ...map[string]any) string {
	f.t.Helper()
	b, err := json.Marshal(map[string]any{"farmId": farm, "date": "2026-11-18", "rows": rows})
	if err != nil {
		f.t.Fatal(err)
	}
	return string(b)
}

// The handler's order (M3 D5): decode → role → farm → codes → Validate →
// service; each step answers its own envelope, and the save's answer is
// sent byte for byte.
func TestFeedDailyHandler(t *testing.T) {
	f := newFeedFixture(t)
	h51F := map[string]any{"sectionId": f.sec["H5.1"], "sex": "F", "feedTypeId": f.breeder2, "g": 160}
	envelope := func(w *httptest.ResponseRecorder) httpx.APIError {
		t.Helper()
		var b httpx.ErrorBody
		if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
			t.Fatalf("%d %s: %v", w.Code, w.Body, err)
		}
		return b.Error
	}
	for _, tc := range []struct {
		name   string
		p      auth.Principal
		body   string
		status int
		code   string
		issue  string // first issue's path
	}{
		{"unknown field", f.ion, `{"farmId":"` + f.farmA.String() + `","rows":[{"grams":160}]}`, 400, "VALIDATION", "rows[0].grams"},
		{"string g", f.ion, `{"rows":[{"g":"160"}]}`, 400, "VALIDATION", "rows[0].g"},
		{"viewer, after decoding", f.viewer, f.body(f.farmA, h51F), 403, "FORBIDDEN", ""},
		{"another farm, before its codes are read", f.outsider, f.body(f.farmA, h51F), 403, "FARM_FORBIDDEN", ""},
		{"owner, unknown farm", f.owner, f.body(uuid.New(), h51F), 404, "NOT_FOUND", ""},
		{"a section of another farm", f.ion, f.body(f.farmA, map[string]any{
			"sectionId": f.sec["T1"], "sex": "F", "feedTypeId": f.breeder2, "g": 160}), 400, "VALIDATION", "rows[0].sectionId"},
		{"H1 names the section", f.ion, f.body(f.farmA, map[string]any{
			"sectionId": f.sec["H5.1"], "sex": "F", "feedTypeId": f.breeder2, "g": 350}), 400, "VALIDATION", "rows[0].g"},
		{"level 2 (H3)", f.ion, f.body(f.farmA, map[string]any{
			"sectionId": f.sec["H7.1"], "sex": "F", "feedTypeId": f.breeder2, "g": 160}), 422, "REFUSED", ""},
		{"over 64 KiB", f.ion, `{"note":"` + strings.Repeat("x", 65<<10) + `"}`, 413, "PAYLOAD_TOO_LARGE", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := f.post(tc.p, tc.body)
			e := envelope(w)
			if w.Code != tc.status || e.Code != tc.code {
				t.Fatalf("got %d %s %q, want %d %s", w.Code, e.Code, e.Message, tc.status, tc.code)
			}
			if tc.issue != "" && (len(e.Issues) == 0 || e.Issues[0].Path != tc.issue) {
				t.Errorf("issues = %+v, want the first on %s", e.Issues, tc.issue)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("Cache-Control = %q", w.Header().Get("Cache-Control"))
			}
		})
	}
	if w := f.post(f.ion, f.body(f.farmA, map[string]any{
		"sectionId": f.sec["H5.1"], "sex": "F", "feedTypeId": f.breeder2, "g": 350})); envelope(w).Message !=
		"H5.1 ♀: 350 g e peste limită. Scrie o valoare între 1 și 300 g." {
		t.Errorf("H1 message = %q", envelope(w).Message)
	}

	// 201, then the same key: 200 with the same bytes.
	key := uuid.New()
	b, err := json.Marshal(map[string]any{"farmId": f.farmA, "date": "2026-11-18", "clientRequestId": key, "rows": []any{h51F}})
	if err != nil {
		t.Fatal(err)
	}
	first := f.post(f.ion, string(b))
	again := f.post(f.ion, string(b))
	if first.Code != http.StatusCreated || again.Code != http.StatusOK || !bytes.Equal(first.Body.Bytes(), again.Body.Bytes()) ||
		first.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("save %d, replay %d:\n%s\n%s", first.Code, again.Code, first.Body, again.Body)
	}
}
