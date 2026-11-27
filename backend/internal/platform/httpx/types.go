// Package httpx is the HTTP edge shared by every module (SPEC §7.0, §7.0a,
// §7.0b, §11.4, §11.11): the JSON codec (Decode, JSON, WriteRaw), the error
// envelope and WriteError, the Validator, and the middleware (request id,
// access log, recover, client IP, CSRF/Origin, CORS, security headers) with a
// Chain to compose them.
//
// Handlers are thin: Decode → Validate → authorize → service → DTO → JSON, and
// errors leave only through WriteError.
package httpx

import (
	"encoding/json/jsontext"
	jsonv2 "encoding/json/v2"

	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// The shared wire types of SPEC §7.0a, tag for tag.

// ErrorBody is the error envelope: {"error": {...}}.
type ErrorBody struct {
	Error APIError `json:"error"`
}

// APIError is the body of the error envelope.
type APIError struct {
	Code              string     `json:"code"`                       // VALIDATION, REFUSED, FARM_FORBIDDEN, …
	Message           string     `json:"message"`                    // the Romanian sentence (Banner / SaveBar)
	Refusals          []Refusal  `json:"refusals,omitzero"`          // REFUSED: always non-nil (written as [] when empty)
	Warnings          []Warning  `json:"warnings,omitzero"`          // REFUSED: always non-nil
	Issues            []Issue    `json:"issues,omitzero"`            // VALIDATION
	EntryID           *uuid.UUID `json:"entryId,omitzero"`           // ALREADY_ENTERED
	Dialog            *Dialog    `json:"dialog,omitzero"`            // CONFIRM_REQUIRED (§6.6 title/body)
	RetryAfterSeconds *int       `json:"retryAfterSeconds,omitzero"` // RATE_LIMITED
}

// Dialog is the confirm Dialog of a 409 CONFIRM_REQUIRED (SPEC §6.6: title and body).
type Dialog struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Refusal is one §6 business-rule refusal of a 422 REFUSED.
type Refusal struct {
	Field             string     `json:"field,omitzero"` // form path ("F", "sections[2].F.mortality")
	SectionCode       string     `json:"sectionCode,omitzero"`
	Sex               string     `json:"sex,omitzero"`
	Max               *int       `json:"max,omitzero"` // §6.1 bound for a removal
	Min               *int       `json:"min,omitzero"` // §6.1 bound for a lowered inflow
	FirstNegativeDate *timex.Day `json:"firstNegativeDate,omitzero"`
	LastNegativeDate  *timex.Day `json:"lastNegativeDate,omitzero"`
	MinAliveEnd       *int       `json:"minAliveEnd,omitzero"`
	Message           string     `json:"message"`
}

// Warning never blocks a save; successful responses and REFUSED carry them.
type Warning struct {
	Code        string `json:"code"` // MORTALITY_SPIKE, FEED_GUIDE, STOCK_NEGATIVE, WEIGHT_GUIDE, …
	SectionCode string `json:"sectionCode,omitzero"`
	Sex         string `json:"sex,omitzero"`
	Message     string `json:"message"`
}

// Issue is one shape problem of a 400 VALIDATION (SPEC §7.0b).
type Issue struct {
	Path    string `json:"path"`    // JSON path: "sections[0].F.mortality"
	Code    string `json:"code"`    // not_integer, out_of_range, required, unknown_field, bad_date, bad_type, too_long
	Message string `json:"message"` // Romanian: the §6 sentence where one exists, else a short generic one
}

// SexPair holds one value per sex.
type SexPair[T any] struct {
	F T `json:"F"`
	M T `json:"M"`
}

// WriteMeta is embedded in every create body (its fields are promoted in JSON).
type WriteMeta struct {
	ClientRequestID *uuid.UUID `json:"clientRequestId"`
	Source          string     `json:"source"` // "manual" (default) | "voice"
	VoiceTranscript *string    `json:"voiceTranscript"`
	VoiceToken      *string    `json:"voiceToken"`
	DryRun          bool       `json:"dryRun"`
}

// EntryRef names a saved entry and its version.
type EntryRef struct {
	ID      uuid.UUID `json:"id"`
	Kind    string    `json:"kind"`
	Version string    `json:"version"` // coalesce(updated_at, created_at) as RFC 3339 with microseconds (opaque to the client)
}

// Optional is a PATCH field where absent means "leave as is" and null means
// "clear" (SPEC §7.0a): Set is true when the key appears, Null when its value
// is JSON null, V holds any other value.
type Optional[T any] struct {
	Set  bool
	Null bool
	V    T
}

// UnmarshalJSONFrom implements encoding/json/v2's UnmarshalerFrom. Decoding V
// on the same decoder keeps the caller's options (unknown members are still
// refused) and the JSON path of any error inside V.
func (o *Optional[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	*o = Optional[T]{Set: true}
	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}
		o.Null = true
		return nil
	}
	return jsonv2.UnmarshalDecode(dec, &o.V)
}
