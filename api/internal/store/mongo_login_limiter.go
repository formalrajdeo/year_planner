package store

import (
	"context"
	"fmt"
	"time"

	"yearplanner/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoLoginRateLimiter struct {
	collection *mongo.Collection
	secret     string
}

func NewMongoLoginRateLimiter(collection *mongo.Collection, secret string) *MongoLoginRateLimiter {
	return &MongoLoginRateLimiter{collection: collection, secret: secret}
}

func (l *MongoLoginRateLimiter) EnsureIndexes(ctx context.Context) error {
	_, err := l.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetName("auth_rate_limit_expiration_ttl").SetExpireAfterSeconds(0),
	})
	return err
}

func (l *MongoLoginRateLimiter) Check(ctx context.Context, ipAddress string, email string) (time.Duration, error) {
	return l.check(ctx, ipAddress, email, "login")
}

func (l *MongoLoginRateLimiter) CheckPasswordRecovery(ctx context.Context, ipAddress string, email string) (time.Duration, error) {
	return l.check(ctx, ipAddress, email, "recovery")
}

func (l *MongoLoginRateLimiter) check(ctx context.Context, ipAddress string, email string, scope string) (time.Duration, error) {
	now := time.Now().UTC()
	windowEnd := now.Truncate(auth.LoginWindowSize).Add(auth.LoginWindowSize)
	ipKey := auth.RateLimitKeyWithSecret(l.secret, scope+"-ip", ipAddress, windowEnd)
	pairKey := auth.RateLimitKeyWithSecret(l.secret, scope+"-pair", ipAddress+"\x00"+email, windowEnd)

	ipCount, err := l.increment(ctx, ipKey, windowEnd)
	if err != nil {
		return 0, err
	}
	if ipCount > auth.LoginIPLimit {
		return windowEnd.Sub(now), nil
	}

	pairCount, err := l.increment(ctx, pairKey, windowEnd)
	if err != nil {
		return 0, err
	}
	if pairCount > auth.LoginPairLimit {
		return windowEnd.Sub(now), nil
	}
	return 0, nil
}

func (l *MongoLoginRateLimiter) increment(ctx context.Context, key string, windowEnd time.Time) (int, error) {
	var result struct {
		Count int `bson:"count"`
	}
	filter := bson.D{{Key: "_id", Value: key}}
	update := bson.D{
		{Key: "$inc", Value: bson.D{{Key: "count", Value: 1}}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "expiresAt", Value: windowEnd.Add(time.Minute)}}},
	}
	err := l.collection.FindOneAndUpdate(
		ctx,
		filter,
		update,
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&result)
	if mongo.IsDuplicateKeyError(err) {
		err = l.collection.FindOneAndUpdate(
			ctx,
			filter,
			update,
			options.FindOneAndUpdate().SetReturnDocument(options.After),
		).Decode(&result)
	}
	if err != nil {
		return 0, fmt.Errorf("increment auth rate limit: %w", err)
	}
	return result.Count, nil
}

var _ auth.LoginRateLimiter = (*MongoLoginRateLimiter)(nil)
