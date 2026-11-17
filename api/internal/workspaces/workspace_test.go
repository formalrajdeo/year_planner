package workspaces

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewWorkspaceCreatesValidModel(t *testing.T) {
	workspace, err := NewWorkspace("  Product Team  ", " product-team ", "owner-1")
	if err != nil {
		t.Fatalf("NewWorkspace() error = %v", err)
	}
	if err := workspace.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := primitive.ObjectIDFromHex(workspace.ID); err != nil {
		t.Fatalf("workspace ID is not a MongoDB ObjectID: %v", err)
	}
	if workspace.Name != "Product Team" || workspace.Slug != "product-team" {
		t.Fatalf("workspace values were not trimmed: %#v", workspace)
	}
	if workspace.OwnerUserID != "owner-1" || workspace.SchemaVersion != 1 {
		t.Fatalf("workspace owner or schema version is incorrect: %#v", workspace)
	}
	if workspace.CreatedAt.Location() != time.UTC || !workspace.CreatedAt.Equal(workspace.UpdatedAt) {
		t.Fatalf("workspace timestamps are not initialized consistently: %#v", workspace)
	}
}

func TestNewWorkspaceRejectsMissingRequiredValues(t *testing.T) {
	if _, err := NewWorkspace(" ", "team", "owner-1"); !errors.Is(err, ErrInvalidWorkspace) {
		t.Fatalf("NewWorkspace() error = %v, want %v", err, ErrInvalidWorkspace)
	}
}

func TestWorkspaceValidationRequiresIdentityAndOwnershipFields(t *testing.T) {
	valid, err := NewWorkspace("Team", "team", "owner-1")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*Workspace)
	}{
		{name: "id", mutate: func(w *Workspace) { w.ID = "" }},
		{name: "name", mutate: func(w *Workspace) { w.Name = "  " }},
		{name: "slug", mutate: func(w *Workspace) { w.Slug = "" }},
		{name: "owner", mutate: func(w *Workspace) { w.OwnerUserID = "" }},
		{name: "schema version", mutate: func(w *Workspace) { w.SchemaVersion = 0 }},
		{name: "created timestamp", mutate: func(w *Workspace) { w.CreatedAt = time.Time{} }},
		{name: "updated timestamp", mutate: func(w *Workspace) { w.UpdatedAt = time.Time{} }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workspace := valid
			test.mutate(&workspace)
			if err := workspace.Validate(); !errors.Is(err, ErrInvalidWorkspace) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidWorkspace)
			}
		})
	}
}

func TestWorkspaceSerializesWithPublicAndMongoFieldNames(t *testing.T) {
	workspace, err := NewWorkspace("Team", "team", "owner-1")
	if err != nil {
		t.Fatal(err)
	}

	jsonBytes, err := json.Marshal(workspace)
	if err != nil {
		t.Fatalf("marshal workspace JSON: %v", err)
	}
	var public map[string]any
	if err := json.Unmarshal(jsonBytes, &public); err != nil {
		t.Fatalf("unmarshal workspace JSON: %v", err)
	}
	if public["ownerUserId"] != "owner-1" || public["id"] != workspace.ID {
		t.Fatalf("unexpected public workspace fields: %#v", public)
	}
	if settings, ok := public["settings"].(map[string]any); !ok || len(settings) != 0 {
		t.Fatalf("public settings = %#v; want an empty typed settings object", public["settings"])
	}
	if _, exists := public["_id"]; exists {
		t.Fatal("public workspace JSON must not expose MongoDB _id")
	}

	bsonBytes, err := bson.Marshal(workspace)
	if err != nil {
		t.Fatalf("marshal workspace BSON: %v", err)
	}
	var stored bson.M
	if err := bson.Unmarshal(bsonBytes, &stored); err != nil {
		t.Fatalf("unmarshal workspace BSON: %v", err)
	}
	if stored["_id"] != workspace.ID || stored["ownerUserId"] != "owner-1" {
		t.Fatalf("unexpected persisted workspace fields: %#v", stored)
	}
	if settings, ok := stored["settings"].(bson.M); !ok || len(settings) != 0 {
		t.Fatalf("persisted settings = %#v; want an empty BSON document", stored["settings"])
	}
}
