package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"yearplanner/internal/workspaces"
)

type CreateWorkspaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *Server) HandleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	if s.workspaceRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace service unavailable")
		return
	}
	ownerUserID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var request CreateWorkspaceRequest
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "request body is required")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	if err := decodeWorkspaceJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	workspace, err := workspaces.NewWorkspace(request.Name, request.Slug, ownerUserID)
	if errors.Is(err, workspaces.ErrInvalidWorkspace) {
		writeError(w, http.StatusBadRequest, "name and slug are required")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create workspace")
		return
	}
	if err := s.workspaceRepo.Create(r.Context(), workspace); err != nil {
		if errors.Is(err, workspaces.ErrSlugExists) {
			writeError(w, http.StatusConflict, "workspace slug already exists")
			return
		}
		if errors.Is(err, workspaces.ErrInvalidWorkspace) {
			writeError(w, http.StatusBadRequest, "invalid workspace")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to create workspace")
		return
	}

	w.Header().Set("Location", "/api/v1/workspaces/"+workspace.ID)
	writeJSON(w, http.StatusCreated, workspace)
}

func (s *Server) HandleListWorkspaces(w http.ResponseWriter, r *http.Request) {
	if s.workspaceRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace service unavailable")
		return
	}
	ownerUserID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	workspaces, err := s.workspaceRepo.ListByOwner(r.Context(), ownerUserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to list workspaces")
		return
	}
	writeJSON(w, http.StatusOK, workspaces)
}

func (s *Server) HandleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	if s.workspaceRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "workspace service unavailable")
		return
	}
	ownerUserID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	workspace, err := s.workspaceRepo.GetByIDForOwner(r.Context(), chi.URLParam(r, "workspaceID"), ownerUserID)
	if errors.Is(err, workspaces.ErrNotFound) {
		writeError(w, http.StatusNotFound, "workspace not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load workspace")
		return
	}
	writeJSON(w, http.StatusOK, workspace)
}

func decodeWorkspaceJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain one JSON value")
		}
		return err
	}
	return nil
}
