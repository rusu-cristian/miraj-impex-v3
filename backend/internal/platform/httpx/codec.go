package httpx

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// MaxJSONBody is the body limit for JSON requests (SPEC §7.0a: 64 KiB).
const MaxJSONBody = 64 << 10

// decodeOptions are encoding/json's own (v1) semantics, as the v1 API uses
// them, with two changes:
//   - errors keep their JSON pointer even when a type's UnmarshalJSON method
//     returns them (num.Num, timex.Day: NOTES-verify "encoding/json in Go 1.27");
//   - unknown members are refused (DisallowUnknownFields).
var decodeOptions = jsonv2.JoinOptions(
	json.DefaultOptionsV1(),
	json.ReportErrorsWithLegacySemantics(false),
	jsonv2.RejectUnknownMembers(true),
)

var (
	dayType       = reflect.TypeFor[timex.Day]()
	localTimeType = reflect.TypeFor[timex.LocalTime]()
)

// Decode reads exactly one JSON value of type T from the request body
// (SPEC §7.0a). The Content-Type must be application/json, the body at most
// maxBytes (MaxJSONBody for JSON routes), and every member must be known.
//
// The returned error is ready for WriteError: 413 PAYLOAD_TOO_LARGE for a body
// over the limit, or 400 VALIDATION with one issue naming the JSON path
// ("sections[0].F.mortality") and a code: unknown_field, bad_type, bad_date,
// malformed (not exactly one JSON value) or bad_content_type.
func Decode[T any](w http.ResponseWriter, r *http.Request, maxBytes int64) (T, error) {
	var v T
	if !IsJSON(r.Header.Get("Content-Type")) {
		return v, Validation(Issue{Path: "", Code: IssueBadContentType, Message: messages.InvalidData()})
	}
	body := r.Body
	if body == nil {
		body = http.NoBody
	}
	data, err := io.ReadAll(http.MaxBytesReader(w, body, maxBytes))
	if err != nil {
		var mbe *http.MaxBytesError
		if errors.As(err, &mbe) {
			return v, PayloadTooLarge()
		}
		return v, fmt.Errorf("httpx: read request body: %w", err)
	}
	if err := jsonv2.Unmarshal(data, &v, decodeOptions); err != nil {
		var zero T
		return zero, decodeError(err, data)
	}
	return v, nil
}

// IsJSON reports whether a Content-Type header names application/json
// (parameters such as charset are allowed).
func IsJSON(contentType string) bool {
	mt, _, err := mime.ParseMediaType(contentType)
	return err == nil && mt == "application/json"
}

// decodeError turns a decode error into 400 VALIDATION with one issue.
func decodeError(err error, data []byte) error {
	issue := Issue{Code: IssueMalformed, Message: messages.InvalidData()}
	var se *jsonv2.SemanticError
	var syn *jsontext.SyntacticError
	switch {
	case errors.As(err, &se):
		issue.Path = jsonPath(se.JSONPointer, data)
		issue.Code = semanticCode(se)
	case errors.As(err, &syn):
		issue.Path = jsonPath(syn.JSONPointer, data)
	}
	return Validation(issue)
}

func semanticCode(se *jsonv2.SemanticError) string {
	if errors.Is(se, jsonv2.ErrUnknownName) {
		return IssueUnknownField
	}
	t := se.GoType
	var ute *json.UnmarshalTypeError
	if errors.As(se, &ute) && ute.Type != nil {
		t = ute.Type
	}
	if t == dayType || t == localTimeType {
		return IssueBadDate
	}
	return IssueBadType
}

// jsonPath renders a JSON pointer ("/sections/0/F/mortality") as the path the
// client uses ("sections[0].F.mortality"). A numeric token is an array index
// only when the document has an array there, so map keys that look like
// numbers stay keys; when the document can't be parsed (a syntax error), any
// all-digit token is taken as an index.
func jsonPath(ptr jsontext.Pointer, data []byte) string {
	var cur any
	walk := jsonv2.Unmarshal(data, &cur, jsontext.AllowDuplicateNames(true), jsontext.AllowInvalidUTF8(true)) == nil
	var b strings.Builder
	for tok := range ptr.Tokens() {
		index := false
		if walk {
			switch c := cur.(type) {
			case []any:
				index = true
				i, err := strconv.Atoi(tok)
				if err != nil || i < 0 || i >= len(c) {
					walk = false
				} else {
					cur = c[i]
				}
			case map[string]any:
				cur = c[tok]
			default:
				walk = false
			}
		} else {
			index = isDigits(tok)
		}
		switch {
		case index:
			b.WriteString("[" + tok + "]")
		case b.Len() == 0:
			b.WriteString(tok)
		default:
			b.WriteString("." + tok)
		}
	}
	return b.String()
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Marshal encodes v the way JSON writes it: encoding/json with HTML escaping
// off (so "›", "<" and "&" stay readable in toasts) and a trailing newline.
// Use it to store a response byte for byte (request_replay) and WriteRaw to
// send it again.
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// JSON writes v as the response body with status (SPEC §7.0a). An encoding
// failure is a 500 through WriteError; nothing is written before encoding
// succeeds.
func JSON(w http.ResponseWriter, status int, v any) {
	b, err := Marshal(v)
	if err != nil {
		WriteError(w, fmt.Errorf("httpx: encode response: %w", err))
		return
	}
	WriteRaw(w, status, b)
}

// WriteRaw writes an already encoded JSON body with the JSON headers:
// Content-Type application/json; charset=utf-8 and Cache-Control no-store.
func WriteRaw(w http.ResponseWriter, status int, body []byte) {
	h := w.Header()
	h.Set("Content-Type", "application/json; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
