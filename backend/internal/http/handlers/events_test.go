package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"backend/internal/http/handlers"
)

func TestTrackEvent_Validation(t *testing.T) {
	// Solo instanciamos el handler con un pool nil, ya que la validación debe fallar
	// ANTES de llamar a la base de datos si el tipo es inválido.
	h := handlers.TrackEvent(nil)

	t.Run("Should reject invalid event type", func(t *testing.T) {
		payload := []byte(`{
			"type": "invalid_event_type",
			"metadata": {"foo": "bar"}
		}`)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Should reject malformed JSON", func(t *testing.T) {
		payload := []byte(`{bad json}`)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
