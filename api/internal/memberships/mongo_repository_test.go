package memberships

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

func TestMongoRepositoryMembershipPersistenceAndUniqueness(t *testing.T) {
	uri := os.Getenv("MEMBERSHIP_TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("MEMBERSHIP_TEST_MONGODB_URI is not set")
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

	collection := client.Database("yearplanner_membership_repo_test").Collection(
		fmt.Sprintf("memberships_%s", primitive.NewObjectID().Hex()),
	)
	defer func() {
		if err := collection.Drop(ctx); err != nil {
			t.Errorf("drop test collection: %v", err)
		}
	}()

	repository := NewMongoRepository(collection)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("create membership indexes: %v", err)
	}

	membership, err := NewMembership("workspace-1", "user-1", "member", "inviter-1")
	if err != nil {
		t.Fatal(err)
	}
	membership.Grants = []string{"plans.read"}
	membership.Denies = []string{"plans.delete"}
	if err := repository.Create(ctx, membership); err != nil {
		t.Fatalf("create membership: %v", err)
	}

	found, err := repository.GetByWorkspaceAndUser(ctx, "workspace-1", "user-1")
	if err != nil || found.ID != membership.ID {
		t.Fatalf("get membership = (%#v, %v), want membership %q", found, err, membership.ID)
	}
	if len(found.Grants) != 1 || found.Grants[0] != "plans.read" ||
		len(found.Denies) != 1 || found.Denies[0] != "plans.delete" {
		t.Fatalf("membership permission overrides were not persisted: %#v", found)
	}
	if _, err := repository.GetByWorkspaceAndUser(ctx, "workspace-2", "user-1"); !errors.Is(err, ErrMembershipNotFound) {
		t.Fatalf("cross-workspace lookup error = %v, want %v", err, ErrMembershipNotFound)
	}

	duplicateMembership, err := NewMembership("workspace-1", "user-1", "member", "inviter-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, duplicateMembership); !errors.Is(err, ErrMembershipExists) {
		t.Fatalf("duplicate membership error = %v, want %v", err, ErrMembershipExists)
	}

	otherWorkspaceMembership, err := NewMembership("workspace-2", "user-1", "member", "inviter-2")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(ctx, otherWorkspaceMembership); err != nil {
		t.Fatalf("create membership for another workspace: %v", err)
	}

	if _, err := collection.UpdateOne(ctx, bson.D{
		{Key: "workspaceId", Value: membership.WorkspaceID},
		{Key: "userId", Value: membership.UserID},
	}, bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: MembershipStatusRemoved}}}}); err != nil {
		t.Fatalf("mark membership removed: %v", err)
	}
	removed, err := repository.GetByWorkspaceAndUser(ctx, "workspace-1", "user-1")
	if err != nil || removed.Status != MembershipStatusRemoved {
		t.Fatalf("get removed membership = (%#v, %v), want removed status", removed, err)
	}
	if err := repository.Create(ctx, duplicateMembership); !errors.Is(err, ErrMembershipExists) {
		t.Fatalf("duplicate removed membership error = %v, want %v", err, ErrMembershipExists)
	}
}

func TestMongoRepositoryRejectsInvalidMembership(t *testing.T) {
	repository := NewMongoRepository(nil)
	if err := repository.Create(context.Background(), Membership{}); !errors.Is(err, ErrInvalidMembership) {
		t.Fatalf("create invalid membership error = %v, want %v", err, ErrInvalidMembership)
	}
	if _, err := repository.GetByWorkspaceAndUser(context.Background(), " ", "user-1"); !errors.Is(err, ErrInvalidMembership) {
		t.Fatalf("lookup with blank workspace ID error = %v, want %v", err, ErrInvalidMembership)
	}
}
