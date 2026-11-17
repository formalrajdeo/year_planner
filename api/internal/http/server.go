package http

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"math"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"yearplanner/internal/auth"
	"yearplanner/internal/config"
	"yearplanner/internal/store"
	"yearplanner/internal/users"
	"yearplanner/internal/workspaces"
	apischema "yearplanner/openapi"
)

type userContextKey struct{}

const (
	sessionCookieName = "yearplanner_session"
	csrfCookieName    = "yearplanner_csrf"
)

type Server struct {
	cfg           config.Config
	repo          users.Repository
	workspaceRepo workspaces.Repository
	session       store.SessionRepository
	database      store.Pinger
	loginLimiter  auth.LoginRateLimiter
	resetTokens   auth.PasswordResetStore
	emailQueue    auth.PasswordResetEmailQueue
	sessionTTL    time.Duration
	allowedOrigin map[string]struct{}
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

type PasswordResetRequest struct {
	Email string `json:"email"`
}

type PasswordChangeRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type AuthResponse struct {
	User PublicUser `json:"user"`
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
	return NewServerWithRateLimiter(cfg, repo, sessions, database, auth.NewMemoryLoginRateLimiter())
}

func NewServerWithRateLimiter(cfg config.Config, repo users.Repository, sessions store.SessionRepository, database store.Pinger, loginLimiter auth.LoginRateLimiter) *Server {
	return NewServerWithAuthDependencies(
		cfg,
		repo,
		sessions,
		database,
		loginLimiter,
		auth.NewMemoryPasswordResetStore(),
		nil,
	)
}

func NewServerWithAuthDependencies(
	cfg config.Config,
	repo users.Repository,
	sessions store.SessionRepository,
	database store.Pinger,
	loginLimiter auth.LoginRateLimiter,
	resetTokens auth.PasswordResetStore,
	emailQueue auth.PasswordResetEmailQueue,
) *Server {
	ttl := auth.DefaultSessionTTL()
	if cfg.SessionTTLMinutes > 0 {
		ttl = time.Duration(cfg.SessionTTLMinutes) * time.Minute
	}

	allowedOrigins := make(map[string]struct{}, len(cfg.AllowedOrigins)+2)
	for _, origin := range cfg.AllowedOrigins {
		allowedOrigins[origin] = struct{}{}
	}
	if cfg.AppBaseURL != "" {
		allowedOrigins[cfg.AppBaseURL] = struct{}{}
	}
	if cfg.AppEnv != "production" && cfg.AppPort != "" {
		allowedOrigins["http://localhost:"+cfg.AppPort] = struct{}{}
		allowedOrigins["http://127.0.0.1:"+cfg.AppPort] = struct{}{}
	}

	return &Server{
		cfg:           cfg,
		repo:          repo,
		session:       sessions,
		database:      database,
		loginLimiter:  loginLimiter,
		resetTokens:   resetTokens,
		emailQueue:    emailQueue,
		sessionTTL:    ttl,
		allowedOrigin: allowedOrigins,
	}
}

func NewServerWithWorkspaceRepository(
	cfg config.Config,
	repo users.Repository,
	sessions store.SessionRepository,
	database store.Pinger,
	loginLimiter auth.LoginRateLimiter,
	resetTokens auth.PasswordResetStore,
	emailQueue auth.PasswordResetEmailQueue,
	workspaceRepo workspaces.Repository,
) *Server {
	server := NewServerWithAuthDependencies(cfg, repo, sessions, database, loginLimiter, resetTokens, emailQueue)
	server.workspaceRepo = workspaceRepo
	return server
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(s.CSRFProtection)

	r.Get("/docs", s.HandleDocs)
	r.Get("/docs/", s.HandleDocs)
	r.Get("/openapi/auth.yaml", s.HandleOpenAPISpec)
	r.Get("/health/live", s.HandleLiveness)
	r.Get("/health/ready", s.HandleReadiness)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", s.HandleRegister)
		r.Post("/auth/login", s.HandleLogin)
		r.Post("/auth/password/forgot", s.HandleForgotPassword)
		r.Post("/auth/password/reset", s.HandleResetPassword)
		r.With(s.RequireAuth).Get("/auth/me", s.HandleMe)
		r.With(s.RequireAuth).Patch("/auth/me", s.HandleUpdateProfile)
		r.With(s.RequireAuth).Post("/auth/logout", s.HandleLogout)
		r.With(s.RequireAuth).Post("/workspaces", s.HandleCreateWorkspace)
		r.With(s.RequireAuth).Get("/workspaces", s.HandleListWorkspaces)
		r.With(s.RequireAuth).Get("/workspaces/{workspaceID}", s.HandleGetWorkspace)
		r.With(s.RequireAuth).Patch("/workspaces/{workspaceID}", s.HandleUpdateWorkspace)
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
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		writeError(w, http.StatusBadRequest, "email is invalid")
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

	if err := s.setAuthCookies(w, token); err != nil {
		_ = s.session.Revoke(r.Context(), token)
		writeError(w, http.StatusInternalServerError, "unable to create session")
		return
	}
	writeJSON(w, http.StatusCreated, AuthResponse{User: publicUser(user)})
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

	email := strings.ToLower(strings.TrimSpace(req.Email))
	retryAfter, err := s.loginLimiter.Check(r.Context(), clientIP(r), email)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "authentication temporarily unavailable")
		return
	}
	if retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(retryAfter.Seconds())))))
		writeError(w, http.StatusTooManyRequests, "too many login attempts; try again later")
		return
	}

	user, err := s.repo.GetByEmail(r.Context(), email)
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

	if err := s.setAuthCookies(w, token); err != nil {
		_ = s.session.Revoke(r.Context(), token)
		writeError(w, http.StatusInternalServerError, "unable to create session")
		return
	}
	writeJSON(w, http.StatusOK, AuthResponse{User: publicUser(user)})
}

func (s *Server) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req PasswordResetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}
	if s.emailQueue == nil || s.resetTokens == nil {
		writeError(w, http.StatusServiceUnavailable, "password recovery is not configured")
		return
	}
	retryAfter, err := s.loginLimiter.CheckPasswordRecovery(r.Context(), clientIP(r), email)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "password recovery is temporarily unavailable")
		return
	}
	if retryAfter > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(max(1, int(math.Ceil(retryAfter.Seconds())))))
		writeError(w, http.StatusTooManyRequests, "too many recovery requests; try again later")
		return
	}

	user, err := s.repo.GetByEmail(r.Context(), email)
	if errors.Is(err, users.ErrNotFound) {
		writePasswordResetAccepted(w)
		return
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "password recovery is temporarily unavailable")
		return
	}

	token, err := auth.NewToken()
	if err != nil {
		slog.Error("password reset token generation failed", "error", err)
		writePasswordResetAccepted(w)
		return
	}
	if err := s.resetTokens.Create(r.Context(), user.ID, auth.HashToken(token), auth.PasswordResetTTL); err != nil {
		slog.Error("password reset token storage failed", "error", err)
		writePasswordResetAccepted(w)
		return
	}

	resetURL, err := url.Parse(s.cfg.AppBaseURL)
	if err != nil {
		slog.Error("password reset URL creation failed", "error", err)
		writePasswordResetAccepted(w)
		return
	}
	resetURL.Fragment = "/reset-password?" + url.Values{"token": []string{token}}.Encode()
	if err := s.emailQueue.Enqueue(r.Context(), user.Email, resetURL.String()); err != nil {
		slog.Error("password reset email queueing failed", "error", err)
	}
	writePasswordResetAccepted(w)
}

func writePasswordResetAccepted(w http.ResponseWriter) {
	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "If an account exists for that email, password reset instructions will be sent.",
	})
}

func (s *Server) HandleResetPassword(w http.ResponseWriter, r *http.Request) {
	if s.resetTokens == nil {
		writeError(w, http.StatusServiceUnavailable, "password recovery is not configured")
		return
	}
	var req PasswordChangeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	token := strings.TrimSpace(req.Token)
	if len(token) != 64 {
		writeError(w, http.StatusBadRequest, "invalid or expired password reset token")
		return
	}
	if _, err := hex.DecodeString(token); err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired password reset token")
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, "password does not meet security requirements")
		return
	}
	userID, err := s.resetTokens.Consume(r.Context(), auth.HashToken(token))
	if errors.Is(err, auth.ErrInvalidResetToken) {
		writeError(w, http.StatusBadRequest, "invalid or expired password reset token")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unable to reset password")
		return
	}
	if err := s.session.RevokeUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to reset password")
		return
	}
	if err := s.repo.UpdatePassword(r.Context(), userID, passwordHash); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to reset password")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "password updated"})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
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
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	if err := s.session.Revoke(r.Context(), cookie.Value); err != nil {
		writeError(w, http.StatusInternalServerError, "unable to revoke session")
		return
	}
	s.clearAuthCookies(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func (s *Server) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		userID, ok, err := s.session.Validate(r.Context(), cookie.Value)
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

func (s *Server) CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if _, ok := s.allowedOrigin[origin]; !ok {
			writeError(w, http.StatusForbidden, "request origin is not allowed")
			return
		}

		if _, err := r.Cookie(sessionCookieName); err == nil {
			csrfCookie, cookieErr := r.Cookie(csrfCookieName)
			headerToken := r.Header.Get("X-CSRF-Token")
			if cookieErr != nil || headerToken == "" ||
				subtle.ConstantTimeCompare([]byte(csrfCookie.Value), []byte(headerToken)) != 1 {
				writeError(w, http.StatusForbidden, "valid CSRF token is required")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) setAuthCookies(w http.ResponseWriter, token string) error {
	csrfToken, err := auth.NewToken()
	if err != nil {
		return err
	}

	secure := s.cfg.AppEnv != "development" && s.cfg.AppEnv != "test"
	maxAge := int(s.sessionTTL.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/api/v1",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
		Expires:  time.Now().Add(s.sessionTTL),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
		Expires:  time.Now().Add(s.sessionTTL),
	})
	return nil
}

func (s *Server) clearAuthCookies(w http.ResponseWriter) {
	secure := s.cfg.AppEnv != "development" && s.cfg.AppEnv != "test"
	for _, cookie := range []*http.Cookie{
		{Name: sessionCookieName, Path: "/api/v1", HttpOnly: true},
		{Name: csrfCookieName, Path: "/", HttpOnly: false},
	} {
		cookie.Value = ""
		cookie.Secure = secure
		cookie.SameSite = http.SameSiteStrictMode
		cookie.MaxAge = -1
		cookie.Expires = time.Unix(1, 0)
		http.SetCookie(w, cookie)
	}
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
