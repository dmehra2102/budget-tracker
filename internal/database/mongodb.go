package database

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/config"
	"go.mongodb.org/mongo-driver/bson"
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
		SetMaxConnIdleTime(5 * time.Minute).
		SetServerSelectionTimeout(10 * time.Second).
		SetRetryWrites(true).
		SetRetryReads(true)

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
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "reset_token", Value: 1}},
		},
	}
	if _, err := db.Collection("users").Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	// Budgets collection indexes
	budgetIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "start_date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}, {Key: "end_date", Value: 1}},
		},
	}
	if _, err := db.Collection("budgets").Indexes().CreateMany(ctx, budgetIndexes); err != nil {
		return err
	}

	// Expenses collection indexes
	expenseIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "budget_id", Value: 1}, {Key: "date", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "date", Value: -1}},
		},
	}
	if _, err := db.Collection("expenses").Indexes().CreateMany(ctx, expenseIndexes); err != nil {
		return err
	}

	// Alerts collection indexes
	alertIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "budget_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_enabled", Value: 1}},
		},
	}
	if _, err := db.Collection("alerts").Indexes().CreateMany(ctx, alertIndexes); err != nil {
		return err
	}

	return nil
}
