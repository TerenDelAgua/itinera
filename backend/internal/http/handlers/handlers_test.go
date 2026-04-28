package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	// 1. Arrange: Preparar el handler (sin necesidad de DB para este endpoint)
	h := &Handlers{}

	// Crear una request simulada
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Crear un ResponseRecorder para capturar la respuesta
	rr := httptest.NewRecorder()

	// 2. Act: Llamar al handler directamente
	handler := http.HandlerFunc(h.healthCheck)
	handler.ServeHTTP(rr, req)

	// 3. Assert: Verificar que el status code es 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Verificar el cuerpo de la respuesta
	expected := `{"Message":"Service is running"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
