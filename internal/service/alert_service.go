package service

import (
	"context"
	"fmt"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertService interface {
	CreateAlert(ctx context.Context, userID primitive.ObjectID, req *domain.CreateAlertRequest) (*domain.Alert, error)
	GetAlert(ctx context.Context, userID, alertID primitive.ObjectID) (*domain.Alert, error)
	GetUserAlerts(ctx context.Context, userID primitive.ObjectID) ([]*domain.Alert, error)
	UpdateAlert(ctx context.Context, userID, alertID primitive.ObjectID, req *domain.CreateAlertRequest) (*domain.Alert, error)
	DeleteAlert(ctx context.Context, userID, alertID primitive.ObjectID) error
	CheckAndSendAlerts(ctx context.Context) error
}

type alertService struct {
	alertRepo    repository.AlertRepository
	budgetRepo   repository.BudgetRepository
	userRepo     repository.UserRepository
	emailService EmailService
}

func NewAlertService(
	alertRepo repository.AlertRepository,
	budgetRepo repository.BudgetRepository,
	userRepo repository.UserRepository,
	emailService EmailService,
) AlertService {
	return &alertService{
		alertRepo:    alertRepo,
		budgetRepo:   budgetRepo,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (s *alertService) CreateAlert(ctx context.Context, userID primitive.ObjectID, req *domain.CreateAlertRequest) (*domain.Alert, error) {
	budgetID, err := primitive.ObjectIDFromHex(req.BudgetID)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	if budget.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	alert := &domain.Alert{
		UserID:    userID,
		BudgetID:  budgetID,
		Type:      req.Type,
		Threshold: req.Threshold,
	}

	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *alertService) GetAlert(ctx context.Context, userID, alertID primitive.ObjectID) (*domain.Alert, error) {
	alert, err := s.alertRepo.FindByID(ctx, alertID)
	if err != nil {
		return nil, err
	}

	if alert.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	return alert, nil
}

func (s *alertService) GetUserAlerts(ctx context.Context, userID primitive.ObjectID) ([]*domain.Alert, error) {
	return s.alertRepo.FindByUserID(ctx, userID)
}

func (s *alertService) UpdateAlert(ctx context.Context, userID, alertID primitive.ObjectID, req *domain.CreateAlertRequest) (*domain.Alert, error) {
	alert, err := s.GetAlert(ctx, userID, alertID)
	if err != nil {
		return nil, err
	}

	alert.Type = req.Type
	alert.Threshold = req.Threshold

	if err := s.alertRepo.Update(ctx, alert); err != nil {
		return nil, err
	}

	return alert, nil
}

func (s *alertService) DeleteAlert(ctx context.Context, userID, alertID primitive.ObjectID) error {
	alert, err := s.GetAlert(ctx, userID, alertID)
	if err != nil {
		return err
	}

	if alert.UserID != userID {
		return domain.ErrUnauthorized
	}

	return s.alertRepo.Delete(ctx, alertID)
}

func (s *alertService) CheckAndSendAlerts(ctx context.Context) error {
	alerts, err := s.alertRepo.FindActiveAlerts(ctx)
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		budget, err := s.budgetRepo.FindByID(ctx, alert.BudgetID)
		if err != nil {
			continue
		}

		percentageSpent := (budget.SpentAmount / budget.TotalAmount) * 100

		shouldTrigger := false
		switch alert.Type {
		case domain.AlertTypePercentage:
			shouldTrigger = percentageSpent >= alert.Threshold
		case domain.AlertTypeAmount:
			shouldTrigger = budget.SpentAmount >= alert.Threshold
		}

		if shouldTrigger {
			if alert.LastSentAt != nil && time.Since(*alert.LastSentAt) < 24*time.Hour {
				continue
			}

			user, err := s.userRepo.FindByID(ctx, alert.UserID)
			if err != nil {
				continue
			}

			if err := s.emailService.SendBudgetAlertEmail(
				ctx,
				user.Email,
				user.FirstName,
				budget.Name,
				percentageSpent,
			); err != nil {
				continue
			}

			now := time.Now()
			alert.LastSentAt = &now
			s.alertRepo.Update(ctx, alert)

			notification := &domain.AlertNotification{
				UserID:    alert.UserID,
				AlertID:   alert.ID,
				BudgetID:  alert.BudgetID,
				Message:   fmt.Sprintf("Budget '%s' has reached %.0f%% of limit", budget.Name, percentageSpent),
				IsSuccess: true,
			}
			s.alertRepo.CreateNotification(ctx, notification)
		}
	}

	return nil
}
