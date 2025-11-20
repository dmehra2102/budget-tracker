package service

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/cache"
	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BudgetService interface {
	CreateBudget(ctx context.Context, userID primitive.ObjectID, req *domain.CreateBudgetRequest) (*domain.Budget, error)
	GetBudget(ctx context.Context, userID, budgetID primitive.ObjectID) (*domain.Budget, error)
	GetUserBudgets(ctx context.Context, userID primitive.ObjectID) ([]*domain.Budget, error)
	UpdateBudget(ctx context.Context, userID primitive.ObjectID, budgetID primitive.ObjectID, req *domain.UpdateBudgetRequest) (*domain.Budget, error)
	DeleteBudget(ctx context.Context, userID, budgetID primitive.ObjectID) error
}

type budgetService struct {
	budgetRepo repository.BudgetRepository
	cache      cache.CacheService
}

func NewBudgetService(
	budgetRepo repository.BudgetRepository,
	cache cache.CacheService,
) BudgetService {
	return &budgetService{
		budgetRepo: budgetRepo,
		cache:      cache,
	}
}

func (s *budgetService) CreateBudget(ctx context.Context, userID primitive.ObjectID, req *domain.CreateBudgetRequest) (*domain.Budget, error) {
	var endDate time.Time
	if req.Period == domain.BudgetPeriodWeekly {
		endDate = req.StartDate.AddDate(0, 0, 7)
	} else {
		endDate = req.StartDate.AddDate(0, 1, 0)
	}

	totalAmount := 0.0
	for _, category := range req.Categories {
		totalAmount += category.Amount
	}

	budget := &domain.Budget{
		UserID:      userID,
		Name:        req.Name,
		Period:      req.Period,
		StartDate:   req.StartDate,
		EndDate:     endDate,
		Categories:  req.Categories,
		TotalAmount: totalAmount,
	}

	if err := s.budgetRepo.Create(ctx, budget); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, "budgets:user:"+userID.Hex())

	return budget, nil
}

func (s *budgetService) GetBudget(ctx context.Context, userID, budgetID primitive.ObjectID) (*domain.Budget, error) {
	cacheKey := "budget:" + budgetID.Hex()
	var budget domain.Budget

	if err := s.cache.Get(ctx, cacheKey, &budget); err != nil {
		if budget.UserID != userID {
			return nil, domain.ErrUnauthorized
		}
		return &budget, nil
	}

	budgetPtr, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	if budgetPtr.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	s.cache.Set(ctx, cacheKey, budgetPtr, 1*time.Hour)

	return budgetPtr, nil
}

func (s *budgetService) GetUserBudgets(ctx context.Context, userID primitive.ObjectID) ([]*domain.Budget, error) {
	cacheKey := "budgets:user:" + userID.Hex()
	var budgets []*domain.Budget

	if err := s.cache.Get(ctx, cacheKey, &budgets); err == nil {
		return budgets, nil
	}

	budgets, err := s.budgetRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.cache.Set(ctx, cacheKey, budgets, 30*time.Minute)

	return budgets, nil
}

func (s *budgetService) UpdateBudget(ctx context.Context, userID primitive.ObjectID, budgetID primitive.ObjectID, req *domain.UpdateBudgetRequest) (*domain.Budget, error) {
	budget, err := s.GetBudget(ctx, userID, budgetID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		budget.Name = req.Name
	}
	if len(req.Categories) > 0 {
		budget.Categories = req.Categories
		totalAmount := 0.0
		for _, category := range req.Categories {
			totalAmount += category.Amount
		}
		budget.TotalAmount = totalAmount
	}

	if err := s.budgetRepo.Update(ctx, budget); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, "budget:"+budgetID.Hex())
	s.cache.Delete(ctx, "budget:user:"+userID.Hex())

	return budget, nil
}

func (s *budgetService) DeleteBudget(ctx context.Context, userID, budgetID primitive.ObjectID) error {
	budget, err := s.GetBudget(ctx, userID, budgetID)
	if err != nil {
		return nil
	}

	if budget.UserID != userID {
		return domain.ErrUnauthorized
	}

	if err := s.budgetRepo.Delete(ctx, budgetID); err != nil {
		return err
	}

	s.cache.Delete(ctx, "budget:"+budgetID.Hex())
	s.cache.Delete(ctx, "budget:user:"+userID.Hex())

	return nil
}
