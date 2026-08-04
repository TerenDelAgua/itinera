package handlers

import (
	"encoding/json"
	"net/http"
)

// Stable, machine-readable error codes.
// Keep these aligned with the spec table; the frontend uses the `code` field
// for human-readable i18n lookups (`error.email_already_exists`) and for
// branching on retry policies (e.g. RESEND_TOO_FREQUENT).
//
// Naming convention: SCREAMING_SNAKE_CASE. The frontend normalises them
// lower-case + dot-separated, so changing case here is a no-op for clients.
const (
	// 4xx — client errors.
	CodeValidationError    = "VALIDATION_ERROR"
	CodeEmailAlreadyExists = "EMAIL_ALREADY_EXISTS"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeWeakPassword       = "WEAK_PASSWORD"
	CodeUnauthenticated    = "UNAUTHENTICATED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeRateLimited        = "RATE_LIMITED"
	CodeTermsNotAccepted   = "TERMS_NOT_ACCEPTED"
	CodeInvalidToken       = "INVALID_RESET_TOKEN"
	CodeLockedToken        = "RESET_TOKEN_LOCKED"
	CodeAccountDeleted     = "ACCOUNT_DELETED"
	CodeSessionExpired     = "SESSION_EXPIRED"
	CodeReuseDetected      = "TOKEN_REUSE_DETECTED"

	// 5xx — server errors. Codes are paired with generic messages so the
	// handler NEVER leaks err.Error() to the client (Spec 017 §9.3).
	CodeInternalError       = "INTERNAL_ERROR"
	CodeUpstreamUnavailable = "UPSTREAM_UNAVAILABLE"
	CodeEmailSendFailed     = "EMAIL_SEND_FAILED"
)

// APIError is the wire contract every error response in Itinera follows.
// The shape `{ "error": { ... } }` is intentional: when the body is
// non-error (e.g. trips, places), the same root can be used as the
// successful payload without ever keying on whether `error` is present.
// Spec 017 §9.3 expands on this rationale.
type APIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// JSONErrorBody is the wrapper around APIError. We render it
// explicitly as `{"error": {...}}` rather than the convention
// `{"code": "...", "message": "..."}` so future top-level fields
// (e.g. pagination meta, request_id) don't collide with error keys.
type JSONErrorBody struct {
	Error APIError `json:"error"`
}

// WriteError serialises an APIError to the response and applies the
// matching HTTP status. It ALWAYS writes the Content-Type header
// (text/plain from http.Error conflicts with the JSON shape the frontend
// expects from this codebase). Use this for every handler error response
// going forward; the old `http.Error` calls are being migrated gradually
// in spec §6.4 fase 1.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteErrorWithFields(w, status, code, message, nil)
}

// WriteErrorWithFields is the same as WriteError, with an extra map for
// per-field validation errors. `fields` keys are field names like `email`.
// Example call site: WriteErrorWithFields(w, 400, CodeValidationError,
// "Check the highlighted fields", map[string]any{"email": "INVALID_FORMAT"}).
func WriteErrorWithFields(w http.ResponseWriter, status int, code, message string, fields map[string]any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(JSONErrorBody{Error: APIError{
		Code:    code,
		Message: message,
		Fields:  fields,
	}})
}

// WriteJSON is the success-side counterpart. It mirrors WriteError's
// Content-Type handling and accepts any json-marshalable value. Used
// by the pilot migration in handlers/auth.go.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
