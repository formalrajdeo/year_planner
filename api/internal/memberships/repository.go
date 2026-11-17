package memberships

import (
	"context"
	"errors"
)

var (
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipExists   = errors.New("membership already exists")
)

type Repository interface {
	Create(ctx context.Context, membership Membership) error
	// GetByWorkspaceAndUser may return removed memberships; callers must check Status before authorizing access.
	GetByWorkspaceAndUser(ctx context.Context, workspaceID string, userID string) (Membership, error)
	ListActiveByWorkspace(ctx context.Context, workspaceID string) ([]Membership, error)
}
