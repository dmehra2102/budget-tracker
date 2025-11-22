package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/middleware"
	"github.com/dmehra2102/budget-tracker/internal/service"
	"github.com/dmehra2102/budget-tracker/internal/utils"
	"github.com/dmehra2102/budget-tracker/pkg/logger"
	"github.com/dmehra2102/budget-tracker/pkg/response"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BudgetHandler struct {
	budgetService service.BudgetService
	validator     *utils.Validator
	logger        *logger.Logger
}

func NewBudgetHandler(budgetService service.BudgetService) *BudgetHandler {
	return &BudgetHandler{
		budgetService: budgetService,
		validator:     utils.NewValidator(),
	}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, err, http.StatusUnauthorized)
		return
	}

	var req domain.CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	budget, err := h.budgetService.CreateBudget(r.Context(), userID, &req)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, budget, http.StatusCreated)
}

func (h *BudgetHandler) GetBudgets(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, err, http.StatusUnauthorized)
		return
	}

	budgets, err := h.budgetService.GetUserBudgets(r.Context(), userID)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, budgets, http.StatusOK)
}

func (h *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, err, http.StatusUnauthorized)
		return
	}

	budgetIDStr := r.PathValue("id")
	if budgetIDStr == "" {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	budgetID, err := primitive.ObjectIDFromHex(budgetIDStr)
	if err != nil {
		response.Error(w, domain.ErrInvalidObjectID, http.StatusBadRequest)
		return
	}

	budget, err := h.budgetService.GetBudget(r.Context(), userID, budgetID)
	if err != nil {
		if err == domain.ErrBudgetNotFound {
			response.Error(w, err, http.StatusNotFound)
		} else {
			response.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	response.Success(w, budget, http.StatusOK)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, err, http.StatusUnauthorized)
		return
	}

	budgetIDStr := r.PathValue("id")
	if budgetIDStr == "" {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	budgetID, err := primitive.ObjectIDFromHex(budgetIDStr)
	if err != nil {
		response.Error(w, domain.ErrInvalidObjectID, http.StatusBadRequest)
		return
	}

	var req domain.UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	updatedBudget, err := h.budgetService.UpdateBudget(r.Context(), userID, budgetID, &req)
	if err != nil {
		if err == domain.ErrBudgetNotFound {
			response.Error(w, err, http.StatusNotFound)
		} else {
			response.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	response.Success(w, updatedBudget, http.StatusOK)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.Error(w, err, http.StatusUnauthorized)
		return
	}

	budgetIDStr := r.PathValue("id")
	if budgetIDStr == "" {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	budgetID, err := primitive.ObjectIDFromHex(budgetIDStr)
	if err != nil {
		response.Error(w, domain.ErrInvalidObjectID, http.StatusBadRequest)
		return
	}

	if err := h.budgetService.DeleteBudget(r.Context(), userID, budgetID); err != nil {
		if err == domain.ErrBudgetNotFound {
			response.Error(w, err, http.StatusNotFound)
		} else {
			response.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	response.Success(w, map[string]string{
		"message": "Budget deleted successfully",
	}, http.StatusOK)
}
