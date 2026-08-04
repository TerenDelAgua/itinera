package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteError_SimpleShape checks the basic 3-field envelope.
func TestWriteError_SimpleShape(t *testing.T) {
	t.Parallel()
	rw := httptest.NewRecorder()
	WriteError(rw, http.StatusBadRequest, CodeValidationError, "Email is required")

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, "application/json; charset=utf-8", rw.Header().Get("Content-Type"))

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
	assert.Equal(t, "Email is required", out.Error.Message)
	assert.Nil(t, out.Error.Fields)
}

// TestWriteError_WithFields pins the per-field validation contract.
func TestWriteError_WithFields(t *testing.T) {
	t.Parallel()
	rw := httptest.NewRecorder()
	WriteErrorWithFields(rw, http.StatusBadRequest, CodeValidationError,
		"Check the highlighted fields",
		map[string]any{"email": "INVALID_FORMAT", "password": "TOO_SHORT"})

	var out JSONErrorBody
	require.NoError(t, json.NewDecoder(rw.Body).Decode(&out))
	assert.Equal(t, CodeValidationError, out.Error.Code)
	require.NotNil(t, out.Error.Fields)
	assert.Equal(t, "INVALID_FORMAT", out.Error.Fields["email"])
	assert.Equal(t, "TOO_SHORT", out.Error.Fields["password"])
}

// TestWriteError_NoInternalLeak documents the rule: callers must NOT
// pass err.Error() in production. The helper happily serialises whatever
// string is given, so this is enforced at code-review time, not at the
// helper level. The test exists to pin that decision for the next
// contributor.
func TestWriteError_NoInternalLeak_Documented(t *testing.T) {
	t.Parallel()
	// If a future change moves the rule into the helper itself, it can
	// live here. For now, smoke-test that the helper is permissive on input.
	rw := httptest.NewRecorder()
	WriteError(rw, http.StatusInternalServerError, CodeInternalError, "Something went wrong")
	assert.Contains(t, rw.Body.String(), "Something went wrong")
}

// TestWriteJSON_HappyPath covers the success wrapper shape. The pilot
// handlers use this rather than encoding inline so the Content-Type
// stays consistent with the error path.
func TestWriteJSON_HappyPath(t *testing.T) {
	t.Parallel()
	rw := httptest.NewRecorder()
	WriteJSON(rw, http.StatusOK, map[string]any{"hello": "world"})
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "application/json; charset=utf-8", rw.Header().Get("Content-Type"))
	assert.Contains(t, rw.Body.String(), `"hello":"world"`)
}

// TestErrorCodes_Stable pins a few of the wire-level code constants.
// Frontend lookups depend on the exact string values.
func TestErrorCodes_Stable(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		CodeValidationError:    "VALIDATION_ERROR",
		CodeEmailAlreadyExists: "EMAIL_ALREADY_EXISTS",
		CodeInvalidCredentials: "INVALID_CREDENTIALS",
		CodeWeakPassword:       "WEAK_PASSWORD",
		CodeRateLimited:        "RATE_LIMITED",
		CodeReuseDetected:      "TOKEN_REUSE_DETECTED",
	}
	for got, want := range cases {
		assert.Equal(t, want, got)
	}
}
