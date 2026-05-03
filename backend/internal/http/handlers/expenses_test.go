package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"backend/internal/http/handlers"
)

func TestCreateExpense_Validation(t *testing.T) {
	// Solo instanciamos el handler vacío, ya que la validación debe fallar
	// ANTES de llamar al servicio o a la base de datos.
	h := &handlers.Handlers{}

	router := chi.NewRouter()
	router.Post("/trips/{id}/expenses", h.CreateExpense)

	t.Run("Should reject invalid JSON payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/trips/"+uuid.New().String()+"/expenses", bytes.NewBuffer([]byte(`{bad json`)))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Invalid JSON")
	})

	t.Run("Should reject negative amount", func(t *testing.T) {
		payload := []byte(`{
			"amount": -10.5,
			"original_currency": "EUR",
			"category_id": "` + uuid.New().String() + `",
			"date": "2023-01-01"
		}`)

		req := httptest.NewRequest(http.MethodPost, "/trips/"+uuid.New().String()+"/expenses", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Amount must be positive")
	})

	t.Run("Should reject zero amount", func(t *testing.T) {
		payload := []byte(`{
			"amount": 0,
			"original_currency": "EUR",
			"category_id": "` + uuid.New().String() + `",
			"date": "2023-01-01"
		}`)

		req := httptest.NewRequest(http.MethodPost, "/trips/"+uuid.New().String()+"/expenses", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "Amount must be positive")
	})
}
