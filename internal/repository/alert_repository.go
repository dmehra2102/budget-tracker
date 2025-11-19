package repository

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlertRepository interface {
	Create(ctx context.Context, alert *domain.Alert) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Alert, error)
	FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Alert, error)
	FindByBudgetID(ctx context.Context, budgetID primitive.ObjectID) ([]*domain.Alert, error)
	Update(ctx context.Context, alert *domain.Alert) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	FindActiveAlerts(ctx context.Context) ([]*domain.Alert, error)
	CreateNotification(ctx context.Context, notification *domain.AlertNotification) error
}

type alertRepository struct {
	collection             *mongo.Collection
	notificationCollection *mongo.Collection
}

func NewAlertRepository(db *mongo.Database) AlertRepository {
	return &alertRepository{
		collection:             db.Collection("alerts"),
		notificationCollection: db.Collection("alert_notifications"),
	}
}

func (r *alertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	alert.CreatedAt = time.Now()
	alert.UpdatedAt = time.Now()
	alert.IsEnabled = true

	result, err := r.collection.InsertOne(ctx, alert)
	if err != nil {
		return err
	}

	alert.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *alertRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Alert, error) {
	var alert domain.Alert
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&alert)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrAlertNotFound
		}
		return nil, err
	}
	return &alert, nil
}

func (r *alertRepository) FindByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.Alert, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []*domain.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *alertRepository) FindByBudgetID(ctx context.Context, budgetID primitive.ObjectID) ([]*domain.Alert, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"budget_id": budgetID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []*domain.Alert
	if err := cursor.All(ctx, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *alertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	alert.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": alert.ID},
		bson.M{"$set": alert},
	)
	return err
}

func (r *alertRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (r *alertRepository) FindActiveAlerts(ctx context.Context) ([]*domain.Alert, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_enabled": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var alerts []*domain.Alert
	if err := cursor.All(ctx, alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *alertRepository) CreateNotification(ctx context.Context, notification *domain.AlertNotification) error {
	notification.SentAt = time.Now()

	_, err := r.notificationCollection.InsertOne(ctx, notification)
	return err
}
