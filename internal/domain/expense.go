package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Expense struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	BudgetID    primitive.ObjectID `bson:"budget_id" json:"budget_id"`
	Category    string             `bson:"category" json:"category"`
	Amount      float64            `bson:"amount" json:"amount"`
	Description string             `bson:"description" json:"description"`
	Date        time.Time          `bson:"date" json:"date"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateExpenseRequest struct {
	BudgetID    string    `json:"budget_id" validate:"required"`
	Category    string    `json:"category" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"  validate:"required"`
}
