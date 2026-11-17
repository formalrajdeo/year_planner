package workspaces

import (
	"context"
	"errors"
)

var (
	ErrNotFound   = errors.New("workspace not found")
	ErrSlugExists = errors.New("workspace slug already exists")
)

type Repository interface {
	Create(ctx context.Context, workspace Workspace) error
	GetByIDForOwner(ctx context.Context, workspaceID string, ownerUserID string) (Workspace, error)
	ListByOwner(ctx context.Context, ownerUserID string) ([]Workspace, error)
	UpdateNameForOwner(ctx context.Context, workspaceID string, ownerUserID string, name string) (Workspace, error)
}
