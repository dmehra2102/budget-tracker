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

type ExpenseRepository interface {
	Create(ctx context.Context, expense *domain.Expense) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Expense, error)
	FindByBudgetID(ctx context.Context, budgetID primitive.ObjectID) ([]*domain.Expense, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Expense, error)
	Update(ctx context.Context, expense *domain.Expense) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type expenseRepository struct {
	collection *mongo.Collection
}

func NewExpenseRepository(db *mongo.Database) ExpenseRepository {
	return &expenseRepository{
		collection: db.Collection("expenses"),
	}
}

func (r *expenseRepository) Create(ctx context.Context, expense *domain.Expense) error {
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, expense)
	if err != nil {
		return err
	}

	expense.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *expenseRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Expense, error) {
	var expense domain.Expense
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&expense)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrExpenseNotFound
		}
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) FindByBudgetID(ctx context.Context, budgetID primitive.ObjectID) ([]*domain.Expense, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"budget_id": budgetID}, options.Find().SetSort(bson.D{{Key: "date", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var expenses []*domain.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Expense, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.D{{Key: "date", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var expenses []*domain.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}
	return expenses, nil
}

func (r *expenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
	expense.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": expense.ID}, bson.M{
		"$set": expense,
	})

	return err
}

func (r *expenseRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
