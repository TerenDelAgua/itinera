package handlers

import (
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {

	cats, err := h.DB.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}

func (h *Handlers) CreateExpense(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	var input models.Expense
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	exp, err := h.DB.CreateExpense(r.Context(), tripId, input)
	if err != nil {
		http.Error(w, "Failed to save", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exp)
}

func (h *Handlers) GetExpenseSummary(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	summary, err := h.DB.GetExpensesSummary(r.Context(), tripId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

func (h *Handlers) ListExpenses(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	expenses, err := h.DB.GetExpensesByTrip(r.Context(), tripId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
