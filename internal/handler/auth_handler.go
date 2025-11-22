package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/service"
	"github.com/dmehra2102/budget-tracker/internal/utils"
	"github.com/dmehra2102/budget-tracker/pkg/response"
)

type AuthHandler struct {
	authService service.AuthService
	validator   *utils.Validator
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   utils.NewValidator(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			response.Error(w, err, http.StatusConflict)
			return
		}
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, result, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	result, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			response.Error(w, err, http.StatusUnauthorized)
			return
		}
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, result, http.StatusOK)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.authService.ForgotPassword(r.Context(), &req); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, map[string]string{
		"message": "Password reset email sent",
	}, http.StatusOK)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, domain.ErrInvalidInput, http.StatusBadRequest)
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		response.ValidationError(w, err)
		return
	}

	if err := h.authService.ResetPassword(r.Context(), &req); err != nil {
		if err == domain.ErrInvalidToken {
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.Success(w, map[string]string{
		"message": "Password reset successful",
	}, http.StatusOK)
}
