package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"yearplanner/internal/memberships"
	"yearplanner/internal/workspaces"
)

type membershipRepositoryStub struct {
	items       []memberships.Membership
	listErr     error
	listCalls   int
	workspaceID string
}

func (r *membershipRepositoryStub) Create(_ context.Context, membership memberships.Membership) error {
	r.items = append(r.items, membership)
	return nil
}

func (r *membershipRepositoryStub) GetByWorkspaceAndUser(_ context.Context, workspaceID string, userID string) (memberships.Membership, error) {
	for _, membership := range r.items {
		if membership.WorkspaceID == workspaceID && membership.UserID == userID {
			return membership, nil
		}
	}
	return memberships.Membership{}, memberships.ErrMembershipNotFound
}

func (r *membershipRepositoryStub) ListActiveByWorkspace(_ context.Context, workspaceID string) ([]memberships.Membership, error) {
	r.listCalls++
	r.workspaceID = workspaceID
	if r.listErr != nil {
		return nil, r.listErr
	}
	active := make([]memberships.Membership, 0)
	for _, membership := range r.items {
		if membership.WorkspaceID == workspaceID && membership.Status == memberships.MembershipStatusActive {
			active = append(active, membership)
		}
	}
	return active, nil
}

func TestWorkspaceMemberListReturnsActiveMembershipsForOwner(t *testing.T) {
	workspace, err := workspaces.NewWorkspace("Team", "team", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	otherWorkspace, err := workspaces.NewWorkspace("Other", "other", "different-owner")
	if err != nil {
		t.Fatal(err)
	}
	activeMembership, err := memberships.NewMembership(workspace.ID, "member-1", "member", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	removedMembership, err := memberships.NewMembership(workspace.ID, "member-2", "member", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	removedMembership.Status = memberships.MembershipStatusRemoved
	otherWorkspaceMembership, err := memberships.NewMembership(otherWorkspace.ID, "member-3", "member", "different-owner")
	if err != nil {
		t.Fatal(err)
	}
	repository := &membershipRepositoryStub{items: []memberships.Membership{
		activeMembership,
		removedMembership,
		otherWorkspaceMembership,
	}}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{
		items: []workspaces.Workspace{workspace, otherWorkspace},
	})
	server.membershipRepo = repository

	response := workspaceRequest(
		server,
		http.MethodGet,
		"/api/v1/workspaces/"+workspace.ID+"/members",
		"",
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusOK {
		t.Fatalf("member list status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var listed []memberships.Membership
	if err := json.Unmarshal(response.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode member list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != activeMembership.ID {
		t.Fatalf("listed memberships = %#v, want only active membership %q", listed, activeMembership.ID)
	}
	if repository.workspaceID != workspace.ID || repository.listCalls != 1 {
		t.Fatalf("membership repository queried workspace %q %d times, want %q once", repository.workspaceID, repository.listCalls, workspace.ID)
	}

	crossOwner := workspaceRequest(
		server,
		http.MethodGet,
		"/api/v1/workspaces/"+otherWorkspace.ID+"/members",
		"",
		sessionToken,
		csrfToken,
	)
	if crossOwner.Code != http.StatusNotFound {
		t.Fatalf("cross-owner member list status = %d, want %d", crossOwner.Code, http.StatusNotFound)
	}
	if repository.listCalls != 1 {
		t.Fatalf("membership repository called for unauthorized workspace; calls = %d", repository.listCalls)
	}
}

func TestWorkspaceMemberListRequiresAuthenticationAndConfiguredRepositories(t *testing.T) {
	workspace, err := workspaces.NewWorkspace("Team", "team", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{items: []workspaces.Workspace{workspace}})

	unauthenticated := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces/"+workspace.ID+"/members", "", "", "")
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated member list status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	unavailable := workspaceRequest(
		server,
		http.MethodGet,
		"/api/v1/workspaces/"+workspace.ID+"/members",
		"",
		sessionToken,
		csrfToken,
	)
	if unavailable.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured member list status = %d, want %d", unavailable.Code, http.StatusServiceUnavailable)
	}

	server.membershipRepo = &membershipRepositoryStub{}
	response := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces/"+workspace.ID+"/members", "", sessionToken, csrfToken)
	if response.Code != http.StatusOK {
		t.Fatalf("configured membership list status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestWorkspaceMemberListMapsRepositoryFailures(t *testing.T) {
	workspace, err := workspaces.NewWorkspace("Team", "team", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{items: []workspaces.Workspace{workspace}})
	server.membershipRepo = &membershipRepositoryStub{listErr: errors.New("database unavailable")}

	response := workspaceRequest(
		server,
		http.MethodGet,
		"/api/v1/workspaces/"+workspace.ID+"/members",
		"",
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("repository failure status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}
