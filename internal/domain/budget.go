package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BudgetPeriod string

const (
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

type Budget struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Name        string             `bson:"name" json:"name"`
	Period      BudgetPeriod       `bson:"period" json:"period"`
	StartDate   time.Time          `bson:"start_date" json:"start_date"`
	EndDate     time.Time          `bson:"end_date" json:"end_date"`
	Categories  []BudgetCategory   `bson:"categories" json:"categories"`
	TotalAmount float64            `bson:"total_amount" json:"total_amount"`
	SpentAmount float64            `bson:"spent_amount" json:"spent_amount"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type BudgetCategory struct {
	Name        string  `bson:"name" json:"name"`
	Amount      float64 `bson:"amount" json:"amount"`
	SpentAmount float64 `bson:"spent_amount" json:"spent_amount"`
}

type CreateBudgetRequest struct {
	Name       string           `json:"name" validate:"required"`
	Period     BudgetPeriod     `json:"period" validate:"required,oneof=weekly monthly"`
	StartDate  time.Time        `json:"start_date" validate:"required"`
	Categories []BudgetCategory `json:"categories" validate:"required,min=1,dive"`
}

type UpdateBudgetRequest struct {
	Name       string           `json:"name"`
	Categories []BudgetCategory `json:"categories"`
}
