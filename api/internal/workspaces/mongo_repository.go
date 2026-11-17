package workspaces

import (
	"context"
	"errors"
	"fmt"

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
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetName("workspaces_slug_unique").SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "ownerUserId", Value: 1},
				{Key: "deletedAt", Value: 1},
				{Key: "createdAt", Value: -1},
				{Key: "_id", Value: -1},
			},
			Options: options.Index().SetName("workspaces_owner_user_created"),
		},
	}
	_, err := r.collection.Indexes().CreateMany(ctx, models)
	return err
}

func (r *MongoRepository) Create(ctx context.Context, workspace Workspace) error {
	if err := workspace.Validate(); err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, workspace); err != nil {
		if !mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("create workspace: %w", err)
		}
		existingCount, lookupErr := r.collection.CountDocuments(ctx, bson.D{{Key: "slug", Value: workspace.Slug}})
		if lookupErr != nil {
			return fmt.Errorf("check workspace slug conflict: %w", lookupErr)
		}
		if existingCount > 0 {
			return ErrSlugExists
		}
		return fmt.Errorf("create workspace: %w", err)
	}
	return nil
}

func (r *MongoRepository) GetByIDForOwner(ctx context.Context, workspaceID string, ownerUserID string) (Workspace, error) {
	if workspaceID == "" || ownerUserID == "" {
		return Workspace{}, ErrNotFound
	}
	var workspace Workspace
	err := r.collection.FindOne(ctx, bson.D{
		{Key: "_id", Value: workspaceID},
		{Key: "ownerUserId", Value: ownerUserID},
		{Key: "deletedAt", Value: nil},
	}).Decode(&workspace)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Workspace{}, ErrNotFound
	}
	if err != nil {
		return Workspace{}, fmt.Errorf("get workspace for owner: %w", err)
	}
	return workspace, nil
}

func (r *MongoRepository) ListByOwner(ctx context.Context, ownerUserID string) ([]Workspace, error) {
	if ownerUserID == "" {
		return nil, ErrInvalidWorkspace
	}
	cursor, err := r.collection.Find(
		ctx,
		bson.D{
			{Key: "ownerUserId", Value: ownerUserID},
			{Key: "deletedAt", Value: nil},
		},
		options.Find().SetSort(bson.D{
			{Key: "createdAt", Value: -1},
			{Key: "_id", Value: -1},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("list workspaces for owner: %w", err)
	}
	defer cursor.Close(ctx)

	workspaces := make([]Workspace, 0)
	if err := cursor.All(ctx, &workspaces); err != nil {
		return nil, fmt.Errorf("decode workspaces for owner: %w", err)
	}
	return workspaces, nil
}

var _ Repository = (*MongoRepository)(nil)
