package service

import (
	"context"

	"github.com/dmehra2102/budget-tracker/internal/cache"
	"github.com/dmehra2102/budget-tracker/internal/domain"
	"github.com/dmehra2102/budget-tracker/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExpenseService interface {
	CreateExpense(ctx context.Context, userID primitive.ObjectID, req *domain.CreateExpenseRequest) (*domain.Expense, error)
	GetExpense(ctx context.Context, userID, expenseID primitive.ObjectID) (*domain.Expense, error)
	GetExpensesByBudget(ctx context.Context, userID, budgetID primitive.ObjectID) ([]*domain.Expense, error)
	UpdateExpense(ctx context.Context, userID, expenseID primitive.ObjectID, req *domain.CreateExpenseRequest) (*domain.Expense, error)
	DeleteExpense(ctx context.Context, userID, expenseID primitive.ObjectID) error
}

type expenseService struct {
	expenseRepo repository.ExpenseRepository
	budgetRepo  repository.BudgetRepository
	cache       cache.CacheService
}

func NewExpenseService(
	expenseRepo repository.ExpenseRepository,
	budgetRepo repository.BudgetRepository,
	cache cache.CacheService,
) ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
		budgetRepo:  budgetRepo,
		cache:       cache,
	}
}

func (s *expenseService) CreateExpense(ctx context.Context, userID primitive.ObjectID, req *domain.CreateExpenseRequest) (*domain.Expense, error) {
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

	expense := &domain.Expense{
		UserID:      userID,
		BudgetID:    budgetID,
		Category:    req.Category,
		Amount:      req.Amount,
		Description: req.Description,
		Date:        req.Date,
	}

	if err := s.expenseRepo.Create(ctx, expense); err != nil {
		return nil, err
	}

	if err := s.budgetRepo.UpdateSpendAmount(ctx, budgetID, req.Amount); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, "budget:"+budgetID.Hex())
	s.cache.Delete(ctx, "budget:user:"+userID.Hex())

	return expense, nil
}

func (s *expenseService) GetExpense(ctx context.Context, userID, expenseID primitive.ObjectID) (*domain.Expense, error) {
	expense, err := s.expenseRepo.FindByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	if expense.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	return expense, nil
}

func (s *expenseService) GetExpensesByBudget(ctx context.Context, userID, budgetID primitive.ObjectID) ([]*domain.Expense, error) {
	budget, err := s.budgetRepo.FindByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	if budget.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	return s.expenseRepo.FindByBudgetID(ctx, budgetID)
}

func (s *expenseService) UpdateExpense(ctx context.Context, userID, expenseID primitive.ObjectID, req *domain.CreateExpenseRequest) (*domain.Expense, error) {
	expense, err := s.GetExpense(ctx, userID, expenseID)
	if err != nil {
		return nil, err
	}

	oldAmount := expense.Amount

	expense.Category = req.Category
	expense.Amount = req.Amount
	expense.Description = req.Description
	expense.Date = req.Date

	if err := s.expenseRepo.Update(ctx, expense); err != nil {
		return nil, err
	}

	diff := req.Amount - oldAmount
	if diff != 0 {
		if err := s.budgetRepo.UpdateSpendAmount(ctx, expense.BudgetID, diff); err != nil {
			return nil, err
		}
	}

	s.cache.Delete(ctx, "budget:"+expense.BudgetID.Hex())
	s.cache.Delete(ctx, "budgets:user:"+userID.Hex())

	return expense, nil
}

func (s *expenseService) DeleteExpense(ctx context.Context, userID, expenseID primitive.ObjectID) error {
	expense, err := s.GetExpense(ctx, userID, expenseID)
	if err != nil {
		return err
	}

	if err := s.expenseRepo.Delete(ctx, expenseID); err != nil {
		return err
	}

	if err := s.budgetRepo.UpdateSpendAmount(ctx, expense.BudgetID, -expense.Amount); err != nil {
		return err
	}

	s.cache.Delete(ctx, "budget:"+expense.BudgetID.Hex())
	s.cache.Delete(ctx, "budgets:user:"+userID.Hex())

	return nil
}
