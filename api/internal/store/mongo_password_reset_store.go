package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"yearplanner/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoPasswordResetStore struct {
	collection *mongo.Collection
}

func NewMongoPasswordResetStore(collection *mongo.Collection) *MongoPasswordResetStore {
	return &MongoPasswordResetStore{collection: collection}
}

func (s *MongoPasswordResetStore) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tokenHash", Value: 1}},
			Options: options.Index().SetName("password_reset_token_hash_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetName("password_reset_expiration_ttl").SetExpireAfterSeconds(0),
		},
		{
			Keys:    bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().SetName("password_reset_user_id_unique").SetUnique(true),
		},
	}
	_, err := s.collection.Indexes().CreateMany(ctx, models)
	return err
}

func (s *MongoPasswordResetStore) Create(ctx context.Context, userID string, tokenHash string, ttl time.Duration) error {
	now := time.Now().UTC()
	filter := bson.D{{Key: "userId", Value: userID}}
	replacement := bson.D{
		{Key: "userId", Value: userID},
		{Key: "tokenHash", Value: tokenHash},
		{Key: "createdAt", Value: now},
		{Key: "expiresAt", Value: now.Add(ttl)},
	}
	_, err := s.collection.ReplaceOne(ctx, filter, replacement, options.Replace().SetUpsert(true))
	if mongo.IsDuplicateKeyError(err) {
		_, err = s.collection.ReplaceOne(ctx, filter, replacement)
	}
	if err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}
	return nil
}

func (s *MongoPasswordResetStore) Consume(ctx context.Context, tokenHash string) (string, error) {
	var token struct {
		UserID    string    `bson:"userId"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}
	err := s.collection.FindOneAndDelete(
		ctx,
		bson.D{{Key: "tokenHash", Value: tokenHash}},
	).Decode(&token)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", auth.ErrInvalidResetToken
	}
	if err != nil {
		return "", fmt.Errorf("consume password reset token: %w", err)
	}
	if !time.Now().UTC().Before(token.ExpiresAt) {
		return "", auth.ErrInvalidResetToken
	}
	return token.UserID, nil
}

var _ auth.PasswordResetStore = (*MongoPasswordResetStore)(nil)
