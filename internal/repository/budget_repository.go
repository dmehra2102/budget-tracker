package repository

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BudgetRepository interface {
	Create(ctx context.Context, budget *domain.Budget) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Budget, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Budget, error)
	Update(ctx context.Context, budget *domain.Budget) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpdateSpendAmount(ctx context.Context, id primitive.ObjectID, amount float64) error
	FindActiveBudgets(ctx context.Context) ([]*domain.Budget, error)
}

type budgetRepository struct {
	collection *mongo.Collection
}

func NewBudgetRepository(db *mongo.Database) BudgetRepository {
	return &budgetRepository{
		collection: db.Collection("budgets"),
	}
}

func (r *budgetRepository) Create(ctx context.Context, budget *domain.Budget) error {
	budget.CreatedAt = time.Now()
	budget.UpdatedAt = time.Now()
	budget.IsActive = true
	budget.SpentAmount = 0

	result, err := r.collection.InsertOne(ctx, budget)
	if err != nil {
		return err
	}

	budget.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *budgetRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Budget, error) {
	var budget domain.Budget
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&budget)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return &budget, nil
}

func (r *budgetRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Budget, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *budgetRepository) Update(ctx context.Context, budget *domain.Budget) error {
	budget.UpdatedAt = time.Now()

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": budget.ID},
		bson.M{"$set": budget},
	)

	if result.MatchedCount == 0 {
		return domain.ErrBudgetNotFound
	}

	return err
}

func (r *budgetRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return domain.ErrBudgetNotFound
	}

	return nil
}

func (r *budgetRepository) UpdateSpendAmount(ctx context.Context, id primitive.ObjectID, amount float64) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{"spend_amount": amount},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)

	if result.MatchedCount == 0 {
		return domain.ErrBudgetNotFound
	}

	return err
}

func (r *budgetRepository) FindActiveBudgets(ctx context.Context) ([]*domain.Budget, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"is_active": true,
		"end_date":  bson.M{"$gte": time.Now()},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var budgets []*domain.Budget
	if err = cursor.All(ctx, &budgets); err != nil {
		return nil, err
	}

	return budgets, nil
}
