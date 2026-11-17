package memberships

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		Keys: bson.D{
			{Key: "workspaceId", Value: 1},
			{Key: "userId", Value: 1},
		},
		Options: options.Index().SetName("memberships_workspace_user_unique").SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("create membership indexes: %w", err)
	}
	return nil
}

func (r *MongoRepository) Create(ctx context.Context, membership Membership) error {
	if err := membership.Validate(); err != nil {
		return err
	}
	if _, err := r.collection.InsertOne(ctx, membership); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrMembershipExists
		}
		return fmt.Errorf("create membership: %w", err)
	}
	return nil
}

func (r *MongoRepository) GetByWorkspaceAndUser(ctx context.Context, workspaceID string, userID string) (Membership, error) {
	if strings.TrimSpace(workspaceID) == "" || strings.TrimSpace(userID) == "" {
		return Membership{}, ErrInvalidMembership
	}

	var membership Membership
	err := r.collection.FindOne(ctx, bson.D{
		{Key: "workspaceId", Value: workspaceID},
		{Key: "userId", Value: userID},
	}).Decode(&membership)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return Membership{}, ErrMembershipNotFound
	}
	if err != nil {
		return Membership{}, fmt.Errorf("get membership by workspace and user: %w", err)
	}
	return membership, nil
}

var _ Repository = (*MongoRepository)(nil)
