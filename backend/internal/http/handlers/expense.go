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

// GetCategories godoc
// @Summary      Get expense categories
// @Description  Get a list of available expense categories and their icons
// @Tags         expenses
// @Produce      json
// @Success      200  {array}   models.ExpenseCategory
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /expenses/categories [get]
func (h *Handlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.ExpensesRepo.GetCategories(r.Context())
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}

// CreateExpense godoc
// @Summary      Create an expense
// @Description  Create a new expense for a trip, optionally associated with a place
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "Trip ID (UUID)"
// @Param        placeId  path      string                        false "Place ID (UUID)"
// @Param        expense  body      services.CreateExpenseInput   true  "Expense data"
// @Success      201      {object}  models.Expense
// @Failure      400      {string}  string "Invalid input"
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{id}/expenses [post]
// @Router       /trips/{id}/places/{placeId}/expenses [post]
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

// GetExpenseSummary godoc
// @Summary      Get trip expense summary
// @Description  Get a detailed summary of expenses by category and place for a trip
// @Tags         expenses
// @Produce      json
// @Param        id   path      string  true  "Trip ID (UUID)"
// @Success      200  {object}  models.TripExpenseSummary
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips/{id}/expenses/summary [get]
func (h *Handlers) GetExpenseSummary(w http.ResponseWriter, r *http.Request) {
	tripId, _ := uuid.Parse(chi.URLParam(r, "id"))
	summary, err := h.ExpensesRepo.GetExpensesSummary(r.Context(), tripId)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

// ListExpenses godoc
// @Summary      List trip expenses
// @Description  Get all expenses associated with a specific trip
// @Tags         expenses
// @Produce      json
// @Param        id   path      string  true  "Trip ID (UUID)"
// @Success      200  {array}   models.Expense
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips/{id}/expenses [get]
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

// DeleteExpense godoc
// @Summary      Delete an expense
// @Description  Delete an expense by ID
// @Tags         expenses
// @Param        id         path      string  true  "Trip ID (UUID)"
// @Param        expenseId  path      string  true  "Expense ID (UUID)"
// @Success      200        {string}  string "OK"
// @Failure      400        {string}  string "Invalid ID"
// @Failure      500        {string}  string "Internal Server Error"
// @Router       /trips/{id}/expenses/{expenseId} [delete]
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

// UpdateExpense godoc
// @Summary      Update an expense
// @Description  Update details of an existing expense
// @Tags         expenses
// @Accept       json
// @Produce      json
// @Param        id         path      string          true  "Trip ID (UUID)"
// @Param        expenseId  path      string          true  "Expense ID (UUID)"
// @Param        expense    body      models.Expense  true  "Updated expense data"
// @Success      200        {object}  models.Expense
// @Failure      400        {string}  string "Invalid input"
// @Failure      500        {string}  string "Internal Server Error"
// @Router       /trips/{id}/expenses/{expenseId} [put]
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

// ListPlaceExpenses godoc
// @Summary      List place expenses
// @Description  Get all expenses associated with a specific place
// @Tags         expenses
// @Produce      json
// @Param        id       path      string  true  "Trip ID (UUID)"
// @Param        placeId  path      string  true  "Place ID (UUID)"
// @Success      200      {array}   models.Expense
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{id}/places/{placeId}/expenses [get]
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

// GetPlaceExpenseSummary godoc
// @Summary      Get place expense summary
// @Description  Get a summary of expenses for a specific place
// @Tags         expenses
// @Produce      json
// @Param        id       path      string  true  "Trip ID (UUID)"
// @Param        placeId  path      string  true  "Place ID (UUID)"
// @Success      200      {array}   models.CategorySummary
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{id}/places/{placeId}/expenses/summary [get]
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
