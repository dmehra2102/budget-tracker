package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"log"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"github.com/dmehra2102/budget-tracker/internal/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthService interface {
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
	ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
}

type authService struct {
	userRepo     repository.UserRepository
	emailService EmailService
	cfg          *config.Config
	jwtAuth      *utils.JWTAuth
}

func NewAuthService(
	userRepo repository.UserRepository,
	emailService EmailService,
	cfg *config.Config,
	jwtAuth *utils.JWTAuth,
) AuthService {
	return &authService{
		userRepo:     userRepo,
		emailService: emailService,
		cfg:          cfg,
		jwtAuth:      jwtAuth,
	}
}

func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:           req.Email,
		Password:        hashedPassword,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		IsEmailVerified: false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	go s.emailService.SendWelcomeEmail(context.Background(), user.Email, user.FirstName)

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.userRepo.Update(ctx, user); err != nil {
		log.Printf("error while updating user from login function : %v", err)
	}

	accessToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Generate reset token
	token := generateSecureToken(32)
	expiry := time.Now().Add(1 * time.Hour)

	// Save reset token
	if err := s.userRepo.UpdateResetToken(ctx, user.Email, token, expiry); err != nil {
		return err
	}

	go s.emailService.SendPasswordResetEmail(context.Background(), user.Email, user.FirstName, token)

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	user, err := s.userRepo.FindByResetToken(ctx, req.Token)
	if err != nil {
		return err
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.Password = hashedPassword
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if err := s.userRepo.ClearResetToken(ctx, user.ID); err != nil {
		return err
	}

	go s.emailService.SendPasswordChangedEmail(ctx, user.Email, user.FirstName)

	return nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	claims, err := s.jwtAuth.ValidateToken(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	// Get user ID from claims
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtAuth.GenerateToken(user.ID.Hex(), user.Email)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
