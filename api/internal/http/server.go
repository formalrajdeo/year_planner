package http

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"yearplanner/internal/auth"
	"yearplanner/internal/config"
	"yearplanner/internal/store"
	"yearplanner/internal/users"
	apischema "yearplanner/openapi"
)

type userContextKey struct{}

type Server struct {
	cfg        config.Config
	repo       users.Repository
	session    store.SessionRepository
	database   store.Pinger
	sessionTTL time.Duration
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ProfileUpdateRequest struct {
	DisplayName string `json:"displayName"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  PublicUser `json:"user"`
}

type PublicUser struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewServer(cfg config.Config, repo users.Repository) *Server {
	return NewServerWithDependencies(cfg, repo, store.NewSessionStore(), nil)
}

func NewServerWithDependencies(cfg config.Config, repo users.Repository, sessions store.SessionRepository, database store.Pinger) *Server {
	ttl := auth.DefaultSessionTTL()
	if cfg.SessionTTLMinutes > 0 {
		ttl = time.Duration(cfg.SessionTTLMinutes) * time.Minute
	}

	return &Server{
		cfg:        cfg,
		repo:       repo,
		session:    sessions,
		database:   database,
		sessionTTL: ttl,
	}
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/docs", s.HandleDocs)
	r.Get("/docs/", s.HandleDocs)
	r.Get("/openapi/auth.yaml", s.HandleOpenAPISpec)
	r.Get("/health/live", s.HandleLiveness)
	r.Get("/health/ready", s.HandleReadiness)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.HandleRegister)
		r.Post("/auth/login", s.HandleLogin)
		r.With(s.RequireAuth).Get("/auth/me", s.HandleMe)
		r.With(s.RequireAuth).Patch("/auth/me", s.HandleUpdateProfile)
		r.With(s.RequireAuth).Post("/auth/logout", s.HandleLogout)
	})

	return r
}

func (s *Server) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	if s.database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.database.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "database unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) HandleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Year Planner API</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.17.14/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        SwaggerUIBundle({
          url: "/openapi/auth.yaml",
          dom_id: "#swagger-ui",
          presets: [SwaggerUIBundle.presets.apis],
          layout: "BaseLayout"
        });
      };
    </script>
  </body>
</html>`))
}

func (s *Server) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec, err := fs.ReadFile(apischema.Files, "auth.yaml")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "OpenAPI specification unavailable")
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}

func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	displayName := strings.TrimSpace(req.DisplayName)
	if email == "" || req.Password == "" || displayName == "" {
		writeError(w, http.StatusBadRequest, "email, password and displayName are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters long")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	user := users.User{
		ID:           users.NextUserID(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: hash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(r.Context(), user); err != nil {
		if errors.Is(err, users.ErrEmailExists) {
			writeError(w, http.StatusConflict, users.ErrEmailExists.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to create user")
		return
	}

	token, err := s.session.Create(r.Context(), user.ID, s.sessionTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create session")
		return
	}

	writeJSON(w, http.StatusCreated, AuthResponse{
		Token: token,
		User:  publicUser(user),
	})
}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := s.repo.GetByEmail(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)))
	if errors.Is(err, users.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to authenticate")
		return
	}

	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := s.session.Create(r.Context(), user.ID, s.sessionTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to create session")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		Token: token,
		User:  publicUser(user),
	})
}

func (s *Server) HandleMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := s.repo.GetByID(r.Context(), userID)
	if errors.Is(err, users.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to load user")
		return
	}

	writeJSON(w, http.StatusOK, publicUser(user))
}

func (s *Server) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req ProfileUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.DisplayName) == "" {
		writeError(w, http.StatusBadRequest, "displayName is required")
		return
	}

	user, err := s.repo.UpdateProfile(r.Context(), userID, strings.TrimSpace(req.DisplayName))
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "unable to update profile")
		return
	}

	writeJSON(w, http.StatusOK, publicUser(user))
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	token, err := bearerTokenFromRequest(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := s.session.Revoke(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to revoke session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *Server) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := bearerTokenFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		userID, ok, err := s.session.Validate(r.Context(), token)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "unable to validate session")
			return
		}
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userContextKey{}).(string)
	return userID, ok
}

func publicUser(user users.User) PublicUser {
	return PublicUser{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func bearerTokenFromRequest(r *http.Request) (string, error) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return "", errors.New("missing Authorization header")
	}
	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("authorization header must be 'Bearer <token>'")
	}
	return strings.TrimSpace(parts[1]), nil
}
