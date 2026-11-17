package store

import (
	"context"
	"errors"
	"time"

	"yearplanner/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSessionStore struct {
	collection *mongo.Collection
}

func NewMongoSessionStore(collection *mongo.Collection) *MongoSessionStore {
	return &MongoSessionStore{collection: collection}
}

func (s *MongoSessionStore) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tokenHash", Value: 1}},
			Options: options.Index().SetName("sessions_token_hash_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetName("sessions_expiration_ttl").SetExpireAfterSeconds(0),
		},
		{
			Keys:    bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().SetName("sessions_user_id"),
		},
	}
	_, err := s.collection.Indexes().CreateMany(ctx, models)
	return err
}

func (s *MongoSessionStore) Create(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}

	_, err = s.collection.InsertOne(ctx, bson.D{
		{Key: "tokenHash", Value: auth.HashToken(token)},
		{Key: "userId", Value: userID},
		{Key: "expiresAt", Value: time.Now().UTC().Add(ttl)},
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *MongoSessionStore) Validate(ctx context.Context, token string) (string, bool, error) {
	if token == "" {
		return "", false, nil
	}

	var session struct {
		UserID    string    `bson:"userId"`
		ExpiresAt time.Time `bson:"expiresAt"`
	}
	filter := bson.D{{Key: "tokenHash", Value: auth.HashToken(token)}}
	err := s.collection.FindOne(ctx, filter).Decode(&session)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if !time.Now().Before(session.ExpiresAt) {
		if err := s.Revoke(ctx, token); err != nil {
			return "", false, err
		}
		return "", false, nil
	}
	return session.UserID, true, nil
}

func (s *MongoSessionStore) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.collection.DeleteOne(ctx, bson.D{{Key: "tokenHash", Value: auth.HashToken(token)}})
	return err
}

func (s *MongoSessionStore) RevokeUser(ctx context.Context, userID string) error {
	_, err := s.collection.DeleteMany(ctx, bson.D{{Key: "userId", Value: userID}})
	return err
}
