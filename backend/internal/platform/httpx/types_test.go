package httpx

import (
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/num"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// TestTypesMatchSpec compares every field's Go type and json tag with the
// declarations of SPEC §7.0a.
func TestTypesMatchSpec(t *testing.T) {
	type field struct{ name, typ, tag string }
	want := map[reflect.Type][]field{
		reflect.TypeFor[ErrorBody](): {
			{"Error", "httpx.APIError", "error"},
		},
		reflect.TypeFor[APIError](): {
			{"Code", "string", "code"},
			{"Message", "string", "message"},
			{"Refusals", "[]httpx.Refusal", "refusals,omitzero"},
			{"Warnings", "[]httpx.Warning", "warnings,omitzero"},
			{"Issues", "[]httpx.Issue", "issues,omitzero"},
			{"EntryID", "*uuid.UUID", "entryId,omitzero"},
			{"Dialog", "*httpx.Dialog", "dialog,omitzero"},
			{"RetryAfterSeconds", "*int", "retryAfterSeconds,omitzero"},
		},
		reflect.TypeFor[Refusal](): {
			{"Field", "string", "field,omitzero"},
			{"SectionCode", "string", "sectionCode,omitzero"},
			{"Sex", "string", "sex,omitzero"},
			{"Max", "*int", "max,omitzero"},
			{"Min", "*int", "min,omitzero"},
			{"FirstNegativeDate", "*timex.Day", "firstNegativeDate,omitzero"},
			{"LastNegativeDate", "*timex.Day", "lastNegativeDate,omitzero"},
			{"MinAliveEnd", "*int", "minAliveEnd,omitzero"},
			{"Message", "string", "message"},
		},
		reflect.TypeFor[Warning](): {
			{"Code", "string", "code"},
			{"SectionCode", "string", "sectionCode,omitzero"},
			{"Sex", "string", "sex,omitzero"},
			{"Message", "string", "message"},
		},
		reflect.TypeFor[Issue](): {
			{"Path", "string", "path"},
			{"Code", "string", "code"},
			{"Message", "string", "message"},
		},
		reflect.TypeFor[SexPair[int]](): {
			{"F", "int", "F"},
			{"M", "int", "M"},
		},
		reflect.TypeFor[WriteMeta](): {
			{"ClientRequestID", "*uuid.UUID", "clientRequestId"},
			{"Source", "string", "source"},
			{"VoiceTranscript", "*string", "voiceTranscript"},
			{"VoiceToken", "*string", "voiceToken"},
			{"DryRun", "bool", "dryRun"},
		},
		reflect.TypeFor[EntryRef](): {
			{"ID", "uuid.UUID", "id"},
			{"Kind", "string", "kind"},
			{"Version", "string", "version"},
		},
	}
	for typ, fields := range want {
		if typ.NumField() != len(fields) {
			t.Errorf("%v has %d fields, SPEC §7.0a has %d", typ, typ.NumField(), len(fields))
			continue
		}
		for i, f := range fields {
			got := typ.Field(i)
			if got.Name != f.name || got.Type.String() != f.typ || got.Tag.Get("json") != f.tag {
				t.Errorf("%v field %d = %s %s `json:%q`, want %s %s `json:%q`",
					typ, i, got.Name, got.Type, got.Tag.Get("json"), f.name, f.typ, f.tag)
			}
		}
	}
}

func TestAPIErrorOmitsUnsetFields(t *testing.T) {
	b, err := Marshal(ErrorBody{Error: APIError{Code: CodeNotFound, Message: "x"}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"error":{"code":"NOT_FOUND","message":"x"}}`+"\n"; got != want {
		t.Errorf("got %s want %s", got, want)
	}

	id := uuid.MustParse("6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d")
	day := timex.MustDay(2026, 2, 28)
	max, retry := 4383, 120
	b, err = Marshal(ErrorBody{Error: APIError{
		Code: CodeRefused, Message: "m",
		Refusals:          []Refusal{{Field: "F", SectionCode: "H2", Sex: "F", Max: &max, FirstNegativeDate: &day, Message: "r"}},
		Warnings:          []Warning{{Code: "FEED_GUIDE", Message: "w"}},
		Issues:            []Issue{{Path: "a", Code: IssueRequired, Message: "i"}},
		EntryID:           &id,
		Dialog:            &Dialog{Title: "t", Body: "b"},
		RetryAfterSeconds: &retry,
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"error":{"code":"REFUSED","message":"m",` +
		`"refusals":[{"field":"F","sectionCode":"H2","sex":"F","max":4383,"firstNegativeDate":"2026-02-28","message":"r"}],` +
		`"warnings":[{"code":"FEED_GUIDE","message":"w"}],"issues":[{"path":"a","code":"required","message":"i"}],` +
		`"entryId":"6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d","dialog":{"title":"t","body":"b"},"retryAfterSeconds":120}}` + "\n"
	if string(b) != want {
		t.Errorf("got  %s\nwant %s", b, want)
	}
}

func TestWriteMetaIsPromoted(t *testing.T) {
	type create struct {
		WriteMeta
		Day timex.Day `json:"day"`
	}
	var c create
	in := `{"clientRequestId":"6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d","source":"voice","voiceTranscript":null,"voiceToken":"t","dryRun":true,"day":"2026-11-18"}`
	if err := json.Unmarshal([]byte(in), &c); err != nil {
		t.Fatal(err)
	}
	if c.ClientRequestID == nil || c.Source != "voice" || c.VoiceTranscript != nil || *c.VoiceToken != "t" || !c.DryRun {
		t.Errorf("decoded %+v", c)
	}
}

func TestOptional(t *testing.T) {
	type patch struct {
		GuideID Optional[uuid.UUID] `json:"guideId"`
		Grams   Optional[*num.Num]  `json:"grams"`
	}
	id := uuid.MustParse("6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d")
	tests := []struct {
		name, in string
		want     Optional[uuid.UUID]
	}{
		{"absent", `{}`, Optional[uuid.UUID]{}},
		{"null", `{"guideId":null}`, Optional[uuid.UUID]{Set: true, Null: true}},
		{"value", `{"guideId":"6f1c2a52-0d1e-4a3b-9c55-2f8e1a7b3c4d"}`, Optional[uuid.UUID]{Set: true, V: id}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Decode[patch](httptest.NewRecorder(), jsonRequest(tt.in), MaxJSONBody)
			if err != nil {
				t.Fatal(err)
			}
			if p.GuideID != tt.want {
				t.Errorf("got %+v want %+v", p.GuideID, tt.want)
			}
		})
	}

	t.Run("value inside keeps its path", func(t *testing.T) {
		_, err := Decode[patch](httptest.NewRecorder(), jsonRequest(`{"grams":"12"}`), MaxJSONBody)
		assertIssue(t, err, "grams", IssueBadType)
	})
	t.Run("number value", func(t *testing.T) {
		p, err := Decode[patch](httptest.NewRecorder(), jsonRequest(`{"grams":12.5}`), MaxJSONBody)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Grams.Set || p.Grams.Null || p.Grams.V == nil || p.Grams.V.String() != "12.5" {
			t.Errorf("got %+v", p.Grams)
		}
	})
}
