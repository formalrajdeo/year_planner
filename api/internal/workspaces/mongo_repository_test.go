package workspaces

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestMongoRepositoryTenantScopingAndGlobalSlugUniqueness(t *testing.T) {
	uri := os.Getenv("WORKSPACE_TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("WORKSPACE_TEST_MONGODB_URI is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("connect to MongoDB: %v", err)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		if err := client.Disconnect(disconnectCtx); err != nil {
			t.Errorf("disconnect from MongoDB: %v", err)
		}
	}()

	collection := client.Database("yearplanner_workspace_repo_test").Collection(
		fmt.Sprintf("workspaces_%s", primitive.NewObjectID().Hex()),
	)
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Errorf("drop test collection: %v", err)
		}
	}()

	repository := NewMongoRepository(collection)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("create workspace indexes: %v", err)
	}

	workspace, err := NewWorkspace("Team One", "team-one", "owner-one")
	if err != nil {
		t.Fatal(err)
	}
	workspace.UpdatedAt = time.Now().Add(-time.Hour).UTC()
	if err := repository.Create(ctx, workspace); err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	found, err := repository.GetByIDForOwner(ctx, workspace.ID, "owner-one")
	if err != nil || found.ID != workspace.ID {
		t.Fatalf("owner-scoped get = (%#v, %v), want workspace %q", found, err, workspace.ID)
	}
	if _, err := repository.GetByIDForOwner(ctx, workspace.ID, "owner-two"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner get error = %v, want %v", err, ErrNotFound)
	}

	otherWorkspace, err := NewWorkspace("Team Two", "team-two", "owner-two")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, otherWorkspace); err != nil {
		t.Fatalf("create second workspace: %v", err)
	}

	ownerWorkspaces, err := repository.ListByOwner(ctx, "owner-one")
	if err != nil || len(ownerWorkspaces) != 1 || ownerWorkspaces[0].ID != workspace.ID {
		t.Fatalf("owner-scoped list = (%#v, %v), want only workspace %q", ownerWorkspaces, err, workspace.ID)
	}

	duplicateSlug, err := NewWorkspace("Duplicate", workspace.Slug, "owner-two")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, duplicateSlug); !errors.Is(err, ErrSlugExists) {
		t.Fatalf("duplicate global slug error = %v, want %v", err, ErrSlugExists)
	}

	renamed, err := repository.UpdateNameForOwner(ctx, workspace.ID, "owner-one", "  Renamed Team  ")
	if err != nil {
		t.Fatalf("rename workspace: %v", err)
	}
	if renamed.Name != "Renamed Team" || renamed.Slug != workspace.Slug || renamed.OwnerUserID != workspace.OwnerUserID {
		t.Fatalf("rename changed unexpected workspace fields: %#v", renamed)
	}
	if !renamed.UpdatedAt.After(workspace.UpdatedAt) {
		t.Fatalf("renamed UpdatedAt = %s, want after %s", renamed.UpdatedAt, workspace.UpdatedAt)
	}
	if _, err := repository.UpdateNameForOwner(ctx, workspace.ID, "owner-two", "Unauthorized"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner rename error = %v, want %v", err, ErrNotFound)
	}
	if _, err := repository.UpdateNameForOwner(ctx, workspace.ID, "owner-one", "  "); !errors.Is(err, ErrInvalidWorkspace) {
		t.Fatalf("empty rename error = %v, want %v", err, ErrInvalidWorkspace)
	}

	deletedAt := time.Now().UTC()
	if _, err := collection.UpdateOne(ctx, bson.D{{Key: "_id", Value: workspace.ID}}, bson.D{
		{Key: "$set", Value: bson.D{{Key: "deletedAt", Value: deletedAt}}},
	}); err != nil {
		t.Fatalf("soft-delete workspace: %v", err)
	}
	if _, err := repository.GetByIDForOwner(ctx, workspace.ID, "owner-one"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get soft-deleted workspace error = %v, want %v", err, ErrNotFound)
	}
	if _, err := repository.UpdateNameForOwner(ctx, workspace.ID, "owner-one", "After deletion"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("rename soft-deleted workspace error = %v, want %v", err, ErrNotFound)
	}
	if err := repository.Create(ctx, duplicateSlug); !errors.Is(err, ErrSlugExists) {
		t.Fatalf("reused soft-deleted slug error = %v, want %v", err, ErrSlugExists)
	}
}
