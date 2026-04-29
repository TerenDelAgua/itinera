package handlers

import (
	"backend/internal/models"
	"backend/internal/services"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {

	cats, err := h.ExpensesRepo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}

func (h *Handlers) CreateExpense(w http.ResponseWriter, r *http.Request) {
	tripID, _ := uuid.Parse(chi.URLParam(r, "id"))
	placeIDParam := chi.URLParam(r, "placeId")

	var input services.CreateExpenseInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if placeIDParam != "" {
		if _, err := uuid.Parse(placeIDParam); err != nil {
			http.Error(w, "Invalid place ID", http.StatusBadRequest)
			return
		}

		if input.PlaceID != nil && *input.PlaceID != "" && *input.PlaceID != placeIDParam {
			http.Error(w, "Place ID mismatch", http.StatusBadRequest)
			return
		}

		input.PlaceID = &placeIDParam
	}

	if input.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
		return
	}

	exp, err := h.ExpenseSvc.CalculateAndCreateExpense(r.Context(), tripID, input)
	if err != nil {
		http.Error(w, "Failed to create expense: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(exp)
}

func (h *Handlers) GetExpenseSummary(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	summary, err := h.ExpensesRepo.GetExpensesSummary(r.Context(), tripId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

func (h *Handlers) ListExpenses(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	expenses, err := h.ExpensesRepo.GetExpensesByTrip(r.Context(), tripId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func (h *Handlers) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	expenseId := chi.URLParam(r, "expenseId")
	if _, err := uuid.Parse(expenseId); err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}
	if _, err := h.ExpensesRepo.Pool.Exec(r.Context(),
		"DELETE FROM Expenses WHERE id = $1", expenseId); err != nil {
		http.Error(w, "Failed to delete expense", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *Handlers) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	expenseId, _ := uuid.Parse(chi.URLParam(r, "expenseId"))
	var input models.Expense
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	updated, err := h.ExpensesRepo.UpdateExpense(r.Context(), expenseId, input)
	if err != nil {
		log.Printf("Error updating expense: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

func (h *Handlers) ListPlaceExpenses(w http.ResponseWriter, r *http.Request) {
	placeId, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	expenses, err := h.ExpensesRepo.ListPlaceExpenses(r.Context(), placeId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func (h *Handlers) GetPlaceExpenseSummary(w http.ResponseWriter, r *http.Request) {
	placeId, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	summary, err := h.ExpensesRepo.GetPlaceExpensesSummary(r.Context(), placeId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
