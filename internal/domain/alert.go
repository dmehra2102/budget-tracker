package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertType string

const (
	AlertTypePercentage AlertType = "percentage"
	AlertTypeAmount     AlertType = "amount"
)

type Alert struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	BudgetID   primitive.ObjectID `bson:"budget_id" json:"budget_id"`
	Type       AlertType          `bson:"type" json:"type"`
	Threshold  float64            `bson:"threshold" json:"threshold"`
	IsEnabled  bool               `bson:"is_enabled" json:"is_enabled"`
	LastSentAt *time.Time         `bson:"last_sent_at,omitempty" json:"last_sent_at"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateAlertRequest struct {
	BudgetID  string    `json:"budget_id" validate:"required"`
	Type      AlertType `json:"type" validate:"required,oneof=percentage amount"`
	Threshold float64   `json:"threshold" validate:"required,gt=0"`
}

type AlertNotification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	AlertID   primitive.ObjectID `bson:"alert_id" json:"alert_id"`
	BudgetID  primitive.ObjectID `bson:"budget_id" json:"budget_id"`
	Message   string             `bson:"message" json:"message"`
	SentAt    time.Time          `bson:"sent_at" json:"sent_at"`
	IsSuccess bool               `bson:"is_success" json:"is_success"`
}
