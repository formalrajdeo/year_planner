package workspaces

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidWorkspace = errors.New("invalid workspace")

type Workspace struct {
	ID            string            `json:"id" bson:"_id"`
	Name          string            `json:"name" bson:"name"`
	Slug          string            `json:"slug" bson:"slug"`
	OwnerUserID   string            `json:"ownerUserId" bson:"ownerUserId"`
	Settings      WorkspaceSettings `json:"settings" bson:"settings"`
	SchemaVersion int               `json:"schemaVersion" bson:"schemaVersion"`
	CreatedAt     time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt" bson:"updatedAt"`
	DeletedAt     *time.Time        `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

// WorkspaceSettings stays empty until workspace-level settings are defined in product scope.
type WorkspaceSettings struct{}

func NewWorkspace(name string, slug string, ownerUserID string) (Workspace, error) {
	now := time.Now().UTC()
	workspace := Workspace{
		ID:            primitive.NewObjectID().Hex(),
		Name:          strings.TrimSpace(name),
		Slug:          strings.TrimSpace(slug),
		OwnerUserID:   strings.TrimSpace(ownerUserID),
		Settings:      WorkspaceSettings{},
		SchemaVersion: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := workspace.Validate(); err != nil {
		return Workspace{}, err
	}
	return workspace, nil
}

func (w Workspace) Validate() error {
	if w.ID == "" ||
		strings.TrimSpace(w.Name) == "" ||
		strings.TrimSpace(w.Slug) == "" ||
		strings.TrimSpace(w.OwnerUserID) == "" ||
		w.SchemaVersion < 1 ||
		w.CreatedAt.IsZero() ||
		w.UpdatedAt.IsZero() {
		return ErrInvalidWorkspace
	}
	return nil
}
