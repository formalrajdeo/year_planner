package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"yearplanner/internal/store"
	"yearplanner/internal/users"
	"yearplanner/internal/workspaces"
)

type workspaceRepositoryStub struct {
	items     []workspaces.Workspace
	createErr error
	listErr   error
	getErr    error
	updateErr error
}

func (r *workspaceRepositoryStub) Create(_ context.Context, workspace workspaces.Workspace) error {
	if r.createErr != nil {
		return r.createErr
	}
	for _, existing := range r.items {
		if existing.Slug == workspace.Slug {
			return workspaces.ErrSlugExists
		}
	}
	r.items = append(r.items, workspace)
	return nil
}

func (r *workspaceRepositoryStub) GetByIDForOwner(_ context.Context, workspaceID string, ownerUserID string) (workspaces.Workspace, error) {
	if r.getErr != nil {
		return workspaces.Workspace{}, r.getErr
	}
	for _, workspace := range r.items {
		if workspace.ID == workspaceID && workspace.OwnerUserID == ownerUserID && workspace.DeletedAt == nil {
			return workspace, nil
		}
	}
	return workspaces.Workspace{}, workspaces.ErrNotFound
}

func (r *workspaceRepositoryStub) ListByOwner(_ context.Context, ownerUserID string) ([]workspaces.Workspace, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	items := make([]workspaces.Workspace, 0)
	for _, workspace := range r.items {
		if workspace.OwnerUserID == ownerUserID && workspace.DeletedAt == nil {
			items = append(items, workspace)
		}
	}
	return items, nil
}

func (r *workspaceRepositoryStub) UpdateNameForOwner(_ context.Context, workspaceID string, ownerUserID string, name string) (workspaces.Workspace, error) {
	if r.updateErr != nil {
		return workspaces.Workspace{}, r.updateErr
	}
	for index, workspace := range r.items {
		if workspace.ID == workspaceID && workspace.OwnerUserID == ownerUserID && workspace.DeletedAt == nil {
			workspace.Name = strings.TrimSpace(name)
			workspace.UpdatedAt = time.Now().UTC()
			r.items[index] = workspace
			return workspace, nil
		}
	}
	return workspaces.Workspace{}, workspaces.ErrNotFound
}

func newWorkspaceHandlerTestServer(t *testing.T, repository workspaces.Repository) (*Server, string, string) {
	t.Helper()
	session := store.NewSessionStore()
	sessionToken, err := session.Create(context.Background(), "trusted-owner", time.Hour)
	if err != nil {
		t.Fatalf("create test session: %v", err)
	}
	server := NewServerWithDependencies(testConfig(), users.NewMemoryRepository(), session, nil)
	server.workspaceRepo = repository
	return server, sessionToken, "csrf-test-token"
}

func workspaceRequest(server *Server, method string, path string, body string, sessionToken string, csrfToken string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if method != http.MethodGet {
		request.Header.Set("Origin", server.cfg.AppBaseURL)
		request.Header.Set("X-CSRF-Token", csrfToken)
		request.AddCookie(&http.Cookie{Name: csrfCookieName, Value: csrfToken})
	}
	if sessionToken != "" {
		request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
	}
	response := httptest.NewRecorder()
	server.Router().ServeHTTP(response, request)
	return response
}

func TestWorkspaceCreateUsesAuthenticatedOwner(t *testing.T) {
	repository := &workspaceRepositoryStub{}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, repository)
	response := workspaceRequest(
		server,
		http.MethodPost,
		"/api/v1/workspaces",
		`{"name":" Team One ","slug":" team-one "}`,
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	var workspace workspaces.Workspace
	if err := json.Unmarshal(response.Body.Bytes(), &workspace); err != nil {
		t.Fatalf("decode created workspace: %v", err)
	}
	if workspace.Name != "Team One" || workspace.Slug != "team-one" || workspace.OwnerUserID != "trusted-owner" {
		t.Fatalf("unexpected created workspace: %#v", workspace)
	}
	if workspace.Settings != (workspaces.WorkspaceSettings{}) {
		t.Fatalf("client settings were unexpectedly accepted: %#v", workspace.Settings)
	}
	if got, want := response.Header().Get("Location"), "/api/v1/workspaces/"+workspace.ID; got != want {
		t.Fatalf("Location = %q, want %q", got, want)
	}
}

func TestWorkspaceCreateRejectsClientControlledFields(t *testing.T) {
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{})
	response := workspaceRequest(
		server,
		http.MethodPost,
		"/api/v1/workspaces",
		`{"name":"Team","slug":"team","ownerUserId":"attacker"}`,
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("create with client-controlled owner status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestWorkspaceRoutesRequireAuthenticationAndCSRF(t *testing.T) {
	server, _, _ := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{})

	unauthenticated := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces", "", "", "")
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list status = %d, want %d", unauthenticated.Code, http.StatusUnauthorized)
	}

	invalidCSRF := httptest.NewRequest(http.MethodPost, "/api/v1/workspaces", strings.NewReader(`{"name":"Team","slug":"team"}`))
	invalidCSRF.Header.Set("Origin", server.cfg.AppBaseURL)
	invalidCSRF.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "invalid-session"})
	invalidCSRFResponse := httptest.NewRecorder()
	server.Router().ServeHTTP(invalidCSRFResponse, invalidCSRF)
	if invalidCSRFResponse.Code != http.StatusForbidden {
		t.Fatalf("invalid CSRF status = %d, want %d", invalidCSRFResponse.Code, http.StatusForbidden)
	}
}

func TestWorkspaceListAndGetAreOwnerScoped(t *testing.T) {
	owned, err := workspaces.NewWorkspace("Owned", "owned", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	other, err := workspaces.NewWorkspace("Other", "other", "different-owner")
	if err != nil {
		t.Fatal(err)
	}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{
		items: []workspaces.Workspace{owned, other},
	})

	list := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces", "", sessionToken, csrfToken)
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d: %s", list.Code, http.StatusOK, list.Body.String())
	}
	var listed []workspaces.Workspace
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode workspace list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != owned.ID {
		t.Fatalf("listed workspaces = %#v, want only owned workspace %q", listed, owned.ID)
	}

	get := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces/"+owned.ID, "", sessionToken, csrfToken)
	if get.Code != http.StatusOK {
		t.Fatalf("get own workspace status = %d, want %d", get.Code, http.StatusOK)
	}
	crossOwnerGet := workspaceRequest(server, http.MethodGet, "/api/v1/workspaces/"+other.ID, "", sessionToken, csrfToken)
	if crossOwnerGet.Code != http.StatusNotFound {
		t.Fatalf("cross-owner get status = %d, want %d", crossOwnerGet.Code, http.StatusNotFound)
	}
}

func TestWorkspaceUpdateRenamesOnlyOwnedWorkspace(t *testing.T) {
	owned, err := workspaces.NewWorkspace("Old Name", "stable-slug", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	owned.UpdatedAt = time.Now().Add(-time.Hour).UTC()
	other, err := workspaces.NewWorkspace("Other", "other", "different-owner")
	if err != nil {
		t.Fatal(err)
	}
	repository := &workspaceRepositoryStub{items: []workspaces.Workspace{owned, other}}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, repository)

	response := workspaceRequest(
		server,
		http.MethodPatch,
		"/api/v1/workspaces/"+owned.ID,
		`{"name":" New Name "}`,
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var updated workspaces.Workspace
	if err := json.Unmarshal(response.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated workspace: %v", err)
	}
	if updated.Name != "New Name" || updated.Slug != owned.Slug || updated.OwnerUserID != owned.OwnerUserID {
		t.Fatalf("workspace update changed immutable fields or failed to rename: %#v", updated)
	}
	if !updated.UpdatedAt.After(owned.UpdatedAt) {
		t.Fatalf("updatedAt = %s, want later than %s", updated.UpdatedAt, owned.UpdatedAt)
	}

	crossOwner := workspaceRequest(
		server,
		http.MethodPatch,
		"/api/v1/workspaces/"+other.ID,
		`{"name":"Attempted Rename"}`,
		sessionToken,
		csrfToken,
	)
	if crossOwner.Code != http.StatusNotFound {
		t.Fatalf("cross-owner update status = %d, want %d", crossOwner.Code, http.StatusNotFound)
	}
}

func TestWorkspaceUpdateRejectsEmptyNameAndSlugChanges(t *testing.T) {
	workspace, err := workspaces.NewWorkspace("Team", "team", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{
		items: []workspaces.Workspace{workspace},
	})
	for _, body := range []string{
		`{"name":"  "}`,
		`{"name":"Renamed","slug":"renamed"}`,
	} {
		response := workspaceRequest(
			server,
			http.MethodPatch,
			"/api/v1/workspaces/"+workspace.ID,
			body,
			sessionToken,
			csrfToken,
		)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("update body %s status = %d, want %d", body, response.Code, http.StatusBadRequest)
		}
	}
}

func TestWorkspaceUpdateMapsRepositoryFailureToInternalError(t *testing.T) {
	workspace, err := workspaces.NewWorkspace("Team", "team", "trusted-owner")
	if err != nil {
		t.Fatal(err)
	}
	server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, &workspaceRepositoryStub{
		items:     []workspaces.Workspace{workspace},
		updateErr: errors.New("database unavailable"),
	})
	response := workspaceRequest(
		server,
		http.MethodPatch,
		"/api/v1/workspaces/"+workspace.ID,
		`{"name":"Renamed"}`,
		sessionToken,
		csrfToken,
	)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("update status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}

func TestWorkspaceCreateMapsSlugConflictAndRepositoryFailure(t *testing.T) {
	for _, test := range []struct {
		name       string
		createErr  error
		wantStatus int
	}{
		{name: "duplicate slug", createErr: workspaces.ErrSlugExists, wantStatus: http.StatusConflict},
		{name: "storage failure", createErr: errors.New("database unavailable"), wantStatus: http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &workspaceRepositoryStub{createErr: test.createErr}
			server, sessionToken, csrfToken := newWorkspaceHandlerTestServer(t, repository)
			response := workspaceRequest(
				server,
				http.MethodPost,
				"/api/v1/workspaces",
				`{"name":"Team","slug":"team"}`,
				sessionToken,
				csrfToken,
			)
			if response.Code != test.wantStatus {
				t.Fatalf("create status = %d, want %d: %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}
