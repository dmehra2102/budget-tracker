package service

import (
	"context"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	// emailService Ema
}
