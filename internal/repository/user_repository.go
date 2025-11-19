package repository

import (
	"context"
	"time"

	"github.com/dmehra2102/budget-tracker/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdateResetToken(ctx context.Context, email, token string, expiry time.Time) error
	FindByResetToken(ctx context.Context, token string) (*domain.User, error)
	ClearResetToken(ctx context.Context, id primitive.ObjectID) error
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": user})
	return err
}

func (r *userRepository) UpdateResetToken(ctx context.Context, email, token string, expiry time.Time) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{
			"reset_token":        token,
			"reset_token_expiry": expiry,
			"updated_at":         time.Now(),
		}},
	)
	return err
}

func (r *userRepository) FindByResetToken(ctx context.Context, token string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{
		"reset_token":        token,
		"reset_token_expiry": bson.M{"$gt": time.Now()},
	}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ClearResetToken(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$unset": bson.M{
				"reset_token":        "",
				"reset_token_expiry": "",
			},
			"$set": bson.M{
				"updated_at": time.Now(),
			},
		},
	)

	return err
}
