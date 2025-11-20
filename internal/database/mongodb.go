package database

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.Database.URI).
		SetMaxPoolSize(cfg.Database.MaxPoolSize).
		SetMinPoolSize(cfg.Database.MinPoolSize).
		SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	db := client.Database(cfg.Database.Database)

	if err := createIndexes(ctx, db); err != nil {
		return nil, err
	}

	return &MongoDB{
		client: client,
		db:     db,
	}, nil
}

func (m *MongoDB) DB() *mongo.Database {
	return m.db
}

func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Users collection indexes
	userIndexes := []mongo.IndexModel{
		{
			Keys:    map[string]any{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]any{"reset_token": 1},
		},
	}
	if _, err := db.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	// Budgets collection indexes
	budgetIndexes := []mongo.IndexModel{
		{
			Keys: map[string]any{"user_id": 1, "start_date": -1},
		},
		{
			Keys: map[string]any{"is_active": 1, "end_date": 1},
		},
	}
	if _, err := db.Collection("budgets").Indexes().CreateMany(ctx, budgetIndexes); err != nil {
		return err
	}

	// Expenses collection indexes
	expenseIndexes := []mongo.IndexModel{
		{
			Keys: map[string]any{"budget_id": 1, "date": -1},
		},
		{
			Keys: map[string]any{"user_id": 1, "date": -1},
		},
	}
	if _, err := db.Collection("expenses").Indexes().CreateMany(ctx, expenseIndexes); err != nil {
		return err
	}

	// Alerts collection indexes
	alertIndexes := []mongo.IndexModel{
		{
			Keys: map[string]any{"user_id": 1},
		},
		{
			Keys: map[string]any{"budget_id": 1},
		},
		{
			Keys: map[string]any{"is_enabled": 1},
		},
	}
	if _, err := db.Collection("alerts").Indexes().CreateMany(ctx, alertIndexes); err != nil {
		return err
	}

	return nil
}
