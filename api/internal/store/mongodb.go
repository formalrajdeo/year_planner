package store

import (
	"context"
	"fmt"

	"yearplanner/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDatabase struct {
	client *mongo.Client
	db     *mongo.Database
}

func ConnectMongo(ctx context.Context, cfg config.Config) (*MongoDatabase, error) {
	clientOptions := options.Client().
		ApplyURI(cfg.MongoURI).
		SetMaxPoolSize(cfg.MongoMaxPoolSize).
		SetConnectTimeout(cfg.MongoTimeout).
		SetServerSelectionTimeout(cfg.MongoTimeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("initialize MongoDB client: %w", err)
	}

	db := &MongoDatabase{
		client: client,
		db:     client.Database(cfg.MongoDatabase),
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.MongoTimeout)
	defer cancel()
	if err := db.Ping(pingCtx); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("connect to MongoDB: %w", err)
	}

	return db, nil
}

func (db *MongoDatabase) Collection(name string) *mongo.Collection {
	return db.db.Collection(name)
}

func (db *MongoDatabase) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, readpref.Primary())
}

func (db *MongoDatabase) Disconnect(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}
