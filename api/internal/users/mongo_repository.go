package users

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(collection *mongo.Collection) *MongoRepository {
	return &MongoRepository{collection: collection}
}

func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetName("users_email_unique").SetUnique(true),
	})
	return err
}

func (r *MongoRepository) Create(ctx context.Context, user User) error {
	_, err := r.collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return ErrEmailExists
	}
	return err
}

func (r *MongoRepository) GetByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *MongoRepository) GetByID(ctx context.Context, userID string) (User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.D{{Key: "_id", Value: userID}}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *MongoRepository) UpdateProfile(ctx context.Context, userID string, displayName string) (User, error) {
	if displayName == "" {
		return User{}, errors.New("display name cannot be empty")
	}

	var user User
	err := r.collection.FindOneAndUpdate(
		ctx,
		bson.D{{Key: "_id", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "displayName", Value: displayName},
			{Key: "updatedAt", Value: time.Now().UTC()},
		}}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrNotFound
	}
	return user, err
}

func (r *MongoRepository) UpdatePassword(ctx context.Context, userID string, passwordHash string) error {
	if passwordHash == "" {
		return errors.New("password hash cannot be empty")
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.D{{Key: "_id", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "passwordHash", Value: passwordHash},
			{Key: "updatedAt", Value: time.Now().UTC()},
		}}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
