package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/db"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/messages"
)

// Envelope codes (SPEC §7.0).
const (
	CodeValidation       = "VALIDATION"        // 400
	CodeUnauthenticated  = "UNAUTHENTICATED"   // 401
	CodeForbidden        = "FORBIDDEN"         // 403
	CodeFarmForbidden    = "FARM_FORBIDDEN"    // 403
	CodePastDayVet       = "PAST_DAY_VET"      // 403
	CodePriceForbidden   = "PRICE_FORBIDDEN"   // 403
	CodeAccountDisabled  = "ACCOUNT_DISABLED"  // 403
	CodeNotFound         = "NOT_FOUND"         // 404
	CodeAlreadyEntered   = "ALREADY_ENTERED"   // 409
	CodeConfirmRequired  = "CONFIRM_REQUIRED"  // 409
	CodeSeriesClosed     = "SERIES_CLOSED"     // 409
	CodeStale            = "STALE"             // 409
	CodePayloadTooLarge  = "PAYLOAD_TOO_LARGE" // 413
	CodeRefused          = "REFUSED"           // 422
	CodeRateLimited      = "RATE_LIMITED"      // 429
	CodeInternal         = "INTERNAL"          // 500
	CodeVoiceUnavailable = "VOICE_UNAVAILABLE" // 503
	CodeUnavailable      = "UNAVAILABLE"       // 503
)

// Issue codes (SPEC §7.0a Issue.code). IssueMalformed and IssueBadContentType
// are the decoder's codes for a body that is not one JSON value and for a
// non-JSON Content-Type.
const (
	IssueNotInteger     = "not_integer"
	IssueOutOfRange     = "out_of_range"
	IssueRequired       = "required"
	IssueUnknownField   = "unknown_field"
	IssueBadDate        = "bad_date"
	IssueBadType        = "bad_type"
	IssueTooLong        = "too_long"
	IssueMalformed      = "malformed"
	IssueBadContentType = "bad_content_type"
)

// HTTPError is an error that knows its status and envelope. *Error implements
// it; module error types (e.g. a REFUSED error carrying refusals) can too,
// without httpx importing the module.
type HTTPError interface {
	error
	HTTPError() (status int, body APIError)
}

// Error is an envelope error with its HTTP status. Err is the cause: it is
// logged for 5xx and never sent to the client.
type Error struct {
	Status int
	APIError
	Err error
}

func (e *Error) Error() string {
	s := "httpx: " + strconv.Itoa(e.Status) + " " + e.Code
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}

func (e *Error) Unwrap() error { return e.Err }

// HTTPError implements HTTPError.
func (e *Error) HTTPError() (int, APIError) { return e.Status, e.APIError }

// NewError is an envelope error with a status, code and message.
func NewError(status int, code, message string) *Error {
	return &Error{Status: status, APIError: APIError{Code: code, Message: message}}
}

// Unauthenticated is 401 UNAUTHENTICATED: no, expired or revoked session.
func Unauthenticated() *Error {
	return NewError(http.StatusUnauthorized, CodeUnauthenticated, messages.Unauthenticated())
}

// Forbidden is 403 FORBIDDEN: an Origin/CSRF/Content-Type refusal (§11.4) or
// a role the route doesn't allow.
func Forbidden() *Error {
	return NewError(http.StatusForbidden, CodeForbidden, messages.Forbidden())
}

// PayloadTooLarge is 413 PAYLOAD_TOO_LARGE: a body over the route's limit.
func PayloadTooLarge() *Error {
	return NewError(http.StatusRequestEntityTooLarge, CodePayloadTooLarge, messages.PayloadTooLarge())
}

// RateLimited is 429 RATE_LIMITED with retryAfterSeconds; WriteError also
// sets the Retry-After header.
func RateLimited(message string, retryAfterSeconds int) *Error {
	e := NewError(http.StatusTooManyRequests, CodeRateLimited, message)
	e.RetryAfterSeconds = &retryAfterSeconds
	return e
}

// Refused is 422 REFUSED; nil refusals or warnings are written as [].
func Refused(message string, refusals []Refusal, warnings []Warning) *Error {
	e := NewError(http.StatusUnprocessableEntity, CodeRefused, message)
	e.Refusals, e.Warnings = refusals, warnings
	return e
}

// Validation is 400 VALIDATION with issues; error.message is the first
// issue's message (SPEC §7.0b).
func Validation(issues ...Issue) *Error {
	msg := messages.InvalidData()
	if len(issues) > 0 && issues[0].Message != "" {
		msg = issues[0].Message
	}
	e := NewError(http.StatusBadRequest, CodeValidation, msg)
	e.Issues = issues
	return e
}

// WriteError is the only way errors leave a handler (SPEC §7.0a). It writes:
//   - an HTTPError (*Error, module errors) as its own status and envelope;
//   - *http.MaxBytesError as 413 PAYLOAD_TOO_LARGE;
//   - context.DeadlineExceeded as 503 UNAVAILABLE;
//   - anything else as 500 INTERNAL with the generic message.
//
// The client never sees the error text. For every 5xx the error is handed to
// the access log (logged with the request id; see AccessLog). PostgreSQL
// constraint violations are named by db.MapError in that log line; services
// turn the ones a route expects (ALREADY_ENTERED with its entryId, a name
// clash, a cycle overlap) into envelope errors themselves, because the
// envelope needs data only the service has, so one that reaches WriteError
// unmapped is a 500.
func WriteError(w http.ResponseWriter, err error) {
	status, body := envelope(err)
	if status >= 500 {
		cause := err
		if cause == nil {
			cause = errors.New("httpx: WriteError called with a nil error")
		}
		recordError(w, db.MapError(cause))
	}
	if rec := recorderOf(w); rec != nil && rec.wroteHeader {
		return // the response has started; the access log still gets the error
	}
	if body.Code == CodeRefused {
		if body.Refusals == nil {
			body.Refusals = []Refusal{}
		}
		if body.Warnings == nil {
			body.Warnings = []Warning{}
		}
	}
	if body.RetryAfterSeconds != nil {
		w.Header().Set("Retry-After", strconv.Itoa(*body.RetryAfterSeconds))
	}
	b, merr := Marshal(ErrorBody{Error: body})
	if merr != nil {
		recordError(w, merr)
		status = http.StatusInternalServerError
		b, _ = Marshal(ErrorBody{Error: APIError{Code: CodeInternal, Message: messages.Internal()}})
	}
	WriteRaw(w, status, b)
}

func envelope(err error) (int, APIError) {
	var he HTTPError
	var mbe *http.MaxBytesError
	switch {
	case err == nil:
	case errors.As(err, &he):
		return he.HTTPError()
	case errors.As(err, &mbe):
		return PayloadTooLarge().HTTPError()
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusServiceUnavailable, APIError{Code: CodeUnavailable, Message: messages.Unavailable()}
	}
	return http.StatusInternalServerError, APIError{Code: CodeInternal, Message: messages.Internal()}
}

// recordError gives a 5xx cause to the access log of this request, or logs it
// directly when the handler runs without AccessLog.
func recordError(w http.ResponseWriter, err error) {
	if rec := recorderOf(w); rec != nil {
		rec.err = err
		return
	}
	slog.Default().Error("httpx: error outside the access log", "error", err.Error())
}
