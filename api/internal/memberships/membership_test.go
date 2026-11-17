package memberships

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewMembershipCreatesValidActiveMember(t *testing.T) {
	membership, err := NewMembership(" workspace-1 ", " user-1 ", " editor ", " inviter-1 ")
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if err := membership.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if _, err := primitive.ObjectIDFromHex(membership.ID); err != nil {
		t.Fatalf("membership ID is not a MongoDB ObjectID: %v", err)
	}
	if membership.WorkspaceID != "workspace-1" ||
		membership.UserID != "user-1" ||
		membership.RoleKey != "editor" ||
		membership.InvitedBy != "inviter-1" {
		t.Fatalf("membership identifiers were not normalized: %#v", membership)
	}
	if membership.Status != MembershipStatusActive {
		t.Fatalf("membership status = %q, want %q", membership.Status, MembershipStatusActive)
	}
	if membership.Grants == nil || len(membership.Grants) != 0 ||
		membership.Denies == nil || len(membership.Denies) != 0 {
		t.Fatalf("membership permission overrides must be initialized empty arrays: %#v", membership)
	}
	if membership.CreatedAt.Location() != time.UTC ||
		!membership.CreatedAt.Equal(membership.AcceptedAt) ||
		!membership.CreatedAt.Equal(membership.UpdatedAt) {
		t.Fatalf("membership timestamps are not initialized consistently: %#v", membership)
	}
}

func TestNewMembershipAllowsNoInviter(t *testing.T) {
	membership, err := NewMembership("workspace-1", "user-1", "owner", "")
	if err != nil {
		t.Fatalf("NewMembership() error = %v", err)
	}
	if membership.InvitedBy != "" {
		t.Fatalf("InvitedBy = %q, want empty for non-invited membership", membership.InvitedBy)
	}
}

func TestMembershipValidationAllowsRemovedStatus(t *testing.T) {
	membership, err := NewMembership("workspace-1", "user-1", "member", "")
	if err != nil {
		t.Fatal(err)
	}
	membership.Status = MembershipStatusRemoved
	if err := membership.Validate(); err != nil {
		t.Fatalf("Validate() for removed membership error = %v", err)
	}
}

func TestMembershipValidationRejectsInvalidFields(t *testing.T) {
	valid, err := NewMembership("workspace-1", "user-1", "member", "")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		mutate func(*Membership)
	}{
		{name: "id", mutate: func(m *Membership) { m.ID = "" }},
		{name: "workspace", mutate: func(m *Membership) { m.WorkspaceID = " " }},
		{name: "user", mutate: func(m *Membership) { m.UserID = "" }},
		{name: "role", mutate: func(m *Membership) { m.RoleKey = " " }},
		{name: "status", mutate: func(m *Membership) { m.Status = "pending" }},
		{name: "nil grants", mutate: func(m *Membership) { m.Grants = nil }},
		{name: "nil denies", mutate: func(m *Membership) { m.Denies = nil }},
		{name: "blank grant", mutate: func(m *Membership) { m.Grants = []string{" "} }},
		{name: "blank deny", mutate: func(m *Membership) { m.Denies = []string{" "} }},
		{name: "zero accepted timestamp", mutate: func(m *Membership) { m.AcceptedAt = time.Time{} }},
		{name: "zero created timestamp", mutate: func(m *Membership) { m.CreatedAt = time.Time{} }},
		{name: "zero updated timestamp", mutate: func(m *Membership) { m.UpdatedAt = time.Time{} }},
		{name: "accepted after updated", mutate: func(m *Membership) { m.AcceptedAt = m.UpdatedAt.Add(time.Second) }},
		{name: "created after updated", mutate: func(m *Membership) { m.CreatedAt = m.UpdatedAt.Add(time.Second) }},
		{name: "blank inviter", mutate: func(m *Membership) { m.InvitedBy = " " }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			membership := valid
			test.mutate(&membership)
			if err := membership.Validate(); !errors.Is(err, ErrInvalidMembership) {
				t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidMembership)
			}
		})
	}
}

func TestMembershipSerializesWithPublicAndMongoFieldNames(t *testing.T) {
	membership, err := NewMembership("workspace-1", "user-1", "member", "inviter-1")
	if err != nil {
		t.Fatal(err)
	}
	membership.Grants = []string{"plans.read"}
	membership.Denies = []string{"plans.delete"}

	jsonBytes, err := json.Marshal(membership)
	if err != nil {
		t.Fatalf("marshal membership JSON: %v", err)
	}
	var public map[string]any
	if err := json.Unmarshal(jsonBytes, &public); err != nil {
		t.Fatalf("unmarshal membership JSON: %v", err)
	}
	if public["id"] != membership.ID ||
		public["workspaceId"] != membership.WorkspaceID ||
		public["userId"] != membership.UserID ||
		public["roleKey"] != membership.RoleKey ||
		public["status"] != string(MembershipStatusActive) {
		t.Fatalf("unexpected public membership fields: %#v", public)
	}
	if _, exists := public["_id"]; exists {
		t.Fatal("public membership JSON must not expose MongoDB _id")
	}
	if _, ok := public["grants"].([]any); !ok {
		t.Fatalf("public grants = %#v, want an array", public["grants"])
	}
	if _, ok := public["denies"].([]any); !ok {
		t.Fatalf("public denies = %#v, want an array", public["denies"])
	}

	bsonBytes, err := bson.Marshal(membership)
	if err != nil {
		t.Fatalf("marshal membership BSON: %v", err)
	}
	var stored bson.M
	if err := bson.Unmarshal(bsonBytes, &stored); err != nil {
		t.Fatalf("unmarshal membership BSON: %v", err)
	}
	if stored["_id"] != membership.ID ||
		stored["workspaceId"] != membership.WorkspaceID ||
		stored["userId"] != membership.UserID ||
		stored["roleKey"] != membership.RoleKey ||
		stored["status"] != string(MembershipStatusActive) {
		t.Fatalf("unexpected persisted membership fields: %#v", stored)
	}
	if _, ok := stored["grants"].(primitive.A); !ok {
		t.Fatalf("persisted grants = %#v, want BSON array", stored["grants"])
	}
	if _, ok := stored["denies"].(primitive.A); !ok {
		t.Fatalf("persisted denies = %#v, want BSON array", stored["denies"])
	}
}
