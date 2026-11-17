package memberships

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidMembership = errors.New("invalid membership")

type MembershipStatus string

const (
	MembershipStatusActive  MembershipStatus = "active"
	MembershipStatusRemoved MembershipStatus = "removed"
)

type Membership struct {
	ID          string           `json:"id" bson:"_id"`
	WorkspaceID string           `json:"workspaceId" bson:"workspaceId"`
	UserID      string           `json:"userId" bson:"userId"`
	RoleKey     string           `json:"roleKey" bson:"roleKey"`
	Grants      []string         `json:"grants" bson:"grants"`
	Denies      []string         `json:"denies" bson:"denies"`
	Status      MembershipStatus `json:"status" bson:"status"`
	InvitedBy   string           `json:"invitedBy,omitempty" bson:"invitedBy,omitempty"`
	AcceptedAt  time.Time        `json:"acceptedAt" bson:"acceptedAt"`
	CreatedAt   time.Time        `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt" bson:"updatedAt"`
}

func NewMembership(workspaceID string, userID string, roleKey string, invitedBy string) (Membership, error) {
	now := time.Now().UTC()
	membership := Membership{
		ID:          primitive.NewObjectID().Hex(),
		WorkspaceID: strings.TrimSpace(workspaceID),
		UserID:      strings.TrimSpace(userID),
		RoleKey:     strings.TrimSpace(roleKey),
		Grants:      []string{},
		Denies:      []string{},
		Status:      MembershipStatusActive,
		InvitedBy:   strings.TrimSpace(invitedBy),
		AcceptedAt:  now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := membership.Validate(); err != nil {
		return Membership{}, err
	}
	return membership, nil
}

func (m Membership) Validate() error {
	if m.ID == "" ||
		strings.TrimSpace(m.WorkspaceID) == "" ||
		strings.TrimSpace(m.UserID) == "" ||
		strings.TrimSpace(m.RoleKey) == "" ||
		(m.Status != MembershipStatusActive && m.Status != MembershipStatusRemoved) ||
		m.Grants == nil ||
		m.Denies == nil ||
		m.AcceptedAt.IsZero() ||
		m.CreatedAt.IsZero() ||
		m.UpdatedAt.IsZero() ||
		m.AcceptedAt.After(m.UpdatedAt) ||
		m.CreatedAt.After(m.UpdatedAt) {
		return ErrInvalidMembership
	}
	if m.InvitedBy != "" && strings.TrimSpace(m.InvitedBy) == "" {
		return ErrInvalidMembership
	}
	for _, permissions := range [][]string{m.Grants, m.Denies} {
		for _, permission := range permissions {
			if strings.TrimSpace(permission) == "" {
				return ErrInvalidMembership
			}
		}
	}
	return nil
}
