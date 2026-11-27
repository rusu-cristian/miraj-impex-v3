package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

type sexCounts struct {
	Mortality *num.Num `json:"mortality"`
	Culls     int      `json:"culls"`
}

type section struct {
	Code string               `json:"sectionCode"`
	F    sexCounts            `json:"F"`
	Tags map[string]timex.Day `json:"tags"`
}

type dailyBody struct {
	WriteMeta
	Day      timex.Day         `json:"day"`
	Sections []section         `json:"sections"`
	Note     *string           `json:"note"`
	Extra    SexPair[*num.Num] `json:"extra"`
}

func jsonRequest(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/x", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

// assertIssue checks err is 400 VALIDATION with exactly one issue at path with code.
func assertIssue(t *testing.T, err error, path, code string) {
	t.Helper()
	var e *Error
	if !errors.As(err, &e) {
		t.Fatalf("error %v (%T) is not *httpx.Error", err, err)
	}
	if e.Status != http.StatusBadRequest || e.Code != CodeValidation || e.Message != messages.InvalidData() {
		t.Fatalf("got %d %s %q, want 400 VALIDATION %q", e.Status, e.Code, e.Message, messages.InvalidData())
	}
	if len(e.Issues) != 1 {
		t.Fatalf("issues = %+v, want one", e.Issues)
	}
	if got := e.Issues[0]; got.Path != path || got.Code != code || got.Message != messages.InvalidData() {
		t.Errorf("issue = %+v, want path %q code %q", got, path, code)
	}
}

func TestDecodeValid(t *testing.T) {
	in := `{"clientRequestId":null,"source":"manual","day":"2026-11-18",
		"sections":[{"sectionCode":"H1","F":{"mortality":3,"culls":1}}],"note":"a < b & c › d","extra":{"F":1.5,"M":null}}`
	r := jsonRequest(in)
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	got, err := Decode[dailyBody](httptest.NewRecorder(), r, MaxJSONBody)
	if err != nil {
		t.Fatal(err)
	}
	if got.Day != timex.MustDay(2026, 11, 18) || len(got.Sections) != 1 || got.Sections[0].F.Mortality.String() != "3" ||
		got.Sections[0].F.Culls != 1 || *got.Note != "a < b & c › d" || got.Extra.F.String() != "1.5" || got.Extra.M != nil {
		t.Errorf("decoded %+v", got)
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name, body, path, code string
	}{
		{"unknown field", `{"day":"2026-11-18","nmae":1}`, "nmae", IssueUnknownField},
		{"unknown nested field", `{"sections":[{"F":{}},{"F":{"mortality":1,"deaths":2}}]}`, "sections[1].F.deaths", IssueUnknownField},
		{"string for int", `{"sections":[{"F":{"culls":"1"}}]}`, "sections[0].F.culls", IssueBadType},
		{"string for num.Num", `{"sections":[{"F":{}},{"F":{"mortality":"2"}}]}`, "sections[1].F.mortality", IssueBadType},
		{"string in SexPair num.Num", `{"extra":{"F":"1"}}`, "extra.F", IssueBadType},
		{"object for list", `{"sections":{}}`, "sections", IssueBadType},
		{"array at the root", `[1]`, "", IssueBadType},
		{"impossible date", `{"day":"2026-02-30"}`, "day", IssueBadDate},
		{"number for date", `{"day":20261118}`, "day", IssueBadDate},
		{"numeric map key stays a key", `{"sections":[{"tags":{"7":"2026-13-01"}}]}`, "sections[0].tags.7", IssueBadDate},
		{"malformed", `{"day":}`, "day", IssueMalformed},
		{"truncated", `{"day":"2026-11-18"`, "", IssueMalformed},
		{"empty body", ``, "", IssueMalformed},
		{"second value", `{"day":"2026-11-18"} {}`, "", IssueMalformed},
		{"trailing garbage", `{"day":"2026-11-18"}x`, "", IssueMalformed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode[dailyBody](httptest.NewRecorder(), jsonRequest(tt.body), MaxJSONBody)
			assertIssue(t, err, tt.path, tt.code)
		})
	}
}

func TestDecodeContentType(t *testing.T) {
	for _, ct := range []string{"", "text/plain", "application/x-www-form-urlencoded", "multipart/form-data; boundary=x", "application/jsonx", "application/json;;"} {
		t.Run(ct, func(t *testing.T) {
			r := jsonRequest(`{}`)
			r.Header.Set("Content-Type", ct)
			_, err := Decode[dailyBody](httptest.NewRecorder(), r, MaxJSONBody)
			assertIssue(t, err, "", IssueBadContentType)
		})
	}
}

func TestDecodeBodyLimit(t *testing.T) {
	// {"note":"xxx…"} of exactly MaxJSONBody bytes passes; one byte more is 413.
	exact := `{"note":"` + strings.Repeat("x", MaxJSONBody-len(`{"note":""}`)) + `"}`
	if _, err := Decode[dailyBody](httptest.NewRecorder(), jsonRequest(exact), MaxJSONBody); err != nil {
		t.Fatalf("body of exactly the limit: %v", err)
	}

	over := `{"note":"` + strings.Repeat("x", MaxJSONBody) + `"}`
	_, err := Decode[dailyBody](httptest.NewRecorder(), jsonRequest(over), MaxJSONBody)
	w := httptest.NewRecorder()
	WriteError(w, err)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status %d, want 413", w.Code)
	}
	if got, want := w.Body.String(), `{"error":{"code":"PAYLOAD_TOO_LARGE","message":"Datele trimise sunt prea mari."}}`+"\n"; got != want {
		t.Errorf("body %s want %s", got, want)
	}
}

func TestDecodeErrorReachesClientAsEnvelope(t *testing.T) {
	_, err := Decode[dailyBody](httptest.NewRecorder(), jsonRequest(`{"sections":[{"F":{"culls":"1"}}]}`), MaxJSONBody)
	w := httptest.NewRecorder()
	WriteError(w, err)
	want := `{"error":{"code":"VALIDATION","message":"Datele trimise nu sunt valide.",` +
		`"issues":[{"path":"sections[0].F.culls","code":"bad_type","message":"Datele trimise nu sunt valide."}]}}` + "\n"
	if w.Code != http.StatusBadRequest || w.Body.String() != want {
		t.Errorf("got %d %s", w.Code, w.Body)
	}
}

func TestJSONAndWriteRaw(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusCreated, map[string]any{"toast": "Hîrbovăț › H1 <ok> & gata", "warnings": []Warning{}})
	if w.Code != http.StatusCreated {
		t.Errorf("status %d", w.Code)
	}
	assertJSONHeaders(t, w.Header())
	if got, want := w.Body.String(), `{"toast":"Hîrbovăț › H1 <ok> & gata","warnings":[]}`+"\n"; got != want {
		t.Errorf("body %s want %s", got, want)
	}

	w = httptest.NewRecorder()
	raw := []byte(`{"stored":"<&>"}`)
	WriteRaw(w, http.StatusOK, raw)
	assertJSONHeaders(t, w.Header())
	if w.Body.String() != string(raw) {
		t.Errorf("WriteRaw changed the body: %s", w.Body)
	}
}

func TestJSONEncodeFailureIs500(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]any{"bad": func() {}})
	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `"code":"INTERNAL"`) {
		t.Errorf("got %d %s", w.Code, w.Body)
	}
}

func assertJSONHeaders(t *testing.T, h http.Header) {
	t.Helper()
	if got := h.Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type %q", got)
	}
	if got := h.Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control %q", got)
	}
}

func TestMarshalMatchesEncodingJSONApartFromEscaping(t *testing.T) {
	v := map[string]any{"a": []int{1, 2}, "n": num.MustParse("40.40245"), "s": "x"}
	got, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(v)
	if string(got) != string(want)+"\n" {
		t.Errorf("Marshal %s, encoding/json %s", got, want)
	}
}
