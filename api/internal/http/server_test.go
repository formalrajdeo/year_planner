package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"yearplanner/internal/auth"
	"yearplanner/internal/config"
	"yearplanner/internal/store"
	"yearplanner/internal/users"
)

func TestAuthFlow(t *testing.T) {
	cfg := testConfig()
	server := NewServer(cfg, users.NewMemoryRepository())
	router := server.Router()

	t.Run("register and fetch profile", func(t *testing.T) {
		body := `{"email":"alice@example.com","password":"secret123","displayName":"Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		req.Header.Set("Origin", cfg.AppBaseURL)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("register status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var authResp AuthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &authResp); err != nil {
			t.Fatalf("unable to decode register response: %v", err)
		}
		if strings.Contains(rec.Body.String(), `"token"`) {
			t.Fatal("response must not expose the session token")
		}
		if authResp.User.Email != "alice@example.com" {
			t.Fatalf("user email = %q, want %q", authResp.User.Email, "alice@example.com")
		}
		sessionCookie := responseCookie(t, rec, sessionCookieName)
		csrfCookie := responseCookie(t, rec, csrfCookieName)
		if !sessionCookie.HttpOnly {
			t.Fatal("session cookie must be HttpOnly")
		}
		if sessionCookie.Value == "" || csrfCookie.Value == "" {
			t.Fatal("expected session and CSRF cookies")
		}

		req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.AddCookie(sessionCookie)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("profile status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("reject invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "invalid-token"})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("invalid token status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("login and logout flow", func(t *testing.T) {
		body := `{"email":"alice@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		req.Header.Set("Origin", cfg.AppBaseURL)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("login status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var authResp AuthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &authResp); err != nil {
			t.Fatalf("unable to decode login response: %v", err)
		}
		if authResp.User.Email != "alice@example.com" {
			t.Fatalf("login email = %q, want %q", authResp.User.Email, "alice@example.com")
		}
		sessionCookie := responseCookie(t, rec, sessionCookieName)
		csrfCookie := responseCookie(t, rec, csrfCookieName)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Origin", cfg.AppBaseURL)
		req.Header.Set("X-CSRF-Token", csrfCookie.Value)
		req.AddCookie(sessionCookie)
		req.AddCookie(csrfCookie)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("logout status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.AddCookie(sessionCookie)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("revoked token status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}

func testConfig() config.Config {
	return config.Config{
		AppEnv:         "test",
		AppPort:        "8080",
		AppBaseURL:     "http://localhost:5173",
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8080"},
		SessionSecret:  "test-secret",
	}
}

func responseCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("response did not set cookie %q", name)
	return nil
}

func TestDocsRoutes(t *testing.T) {
	server := NewServer(config.Config{AppEnv: "test"}, users.NewMemoryRepository())
	router := server.Router()

	tests := []struct {
		path        string
		contentType string
		contains    string
	}{
		{path: "/docs", contentType: "text/html; charset=utf-8", contains: `url: "/openapi/auth.yaml"`},
		{path: "/openapi/auth.yaml", contentType: "application/yaml; charset=utf-8", contains: "/api/v1/auth/register:"},
		{path: "/openapi/auth.yaml", contentType: "application/yaml; charset=utf-8", contains: "/api/v1/auth/password/reset:"},
		{path: "/openapi/auth.yaml", contentType: "application/yaml; charset=utf-8", contains: "/api/v1/workspaces:"},
		{path: "/openapi/auth.yaml", contentType: "application/yaml; charset=utf-8", contains: "/api/v1/workspaces/{workspaceID}:"},
		{path: "/openapi/auth.yaml", contentType: "application/yaml; charset=utf-8", contains: "Rename a workspace owned by the authenticated user"},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", test.path, rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != test.contentType {
				t.Fatalf("GET %s Content-Type = %q, want %q", test.path, got, test.contentType)
			}
			if !strings.Contains(rec.Body.String(), test.contains) {
				t.Fatalf("GET %s response does not contain %q", test.path, test.contains)
			}
		})
	}
}

type testPinger struct {
	err error
}

func (p testPinger) Ping(context.Context) error {
	return p.err
}

func TestHealthEndpoints(t *testing.T) {
	server := NewServerWithDependencies(
		config.Config{AppEnv: "test"},
		users.NewMemoryRepository(),
		store.NewSessionStore(),
		testPinger{},
	)
	router := server.Router()

	live := httptest.NewRecorder()
	router.ServeHTTP(live, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if live.Code != http.StatusOK {
		t.Fatalf("liveness status = %d, want %d", live.Code, http.StatusOK)
	}

	ready := httptest.NewRecorder()
	router.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if ready.Code != http.StatusOK {
		t.Fatalf("readiness status = %d, want %d", ready.Code, http.StatusOK)
	}

	unavailableServer := NewServerWithDependencies(
		config.Config{AppEnv: "test"},
		users.NewMemoryRepository(),
		store.NewSessionStore(),
		testPinger{err: errors.New("database unavailable")},
	)
	unavailable := httptest.NewRecorder()
	unavailableServer.Router().ServeHTTP(unavailable, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if unavailable.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable readiness status = %d, want %d", unavailable.Code, http.StatusServiceUnavailable)
	}
}

func TestWriteErrorKeepsStableJSONShape(t *testing.T) {
	rec := httptest.NewRecorder()
	writeError(rec, http.StatusBadRequest, "invalid request")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body["error"] != "invalid request" {
		t.Fatalf("error = %q, want %q", body["error"], "invalid request")
	}
}

func TestCSRFProtectionRejectsUntrustedOriginAndMissingToken(t *testing.T) {
	server := NewServer(testConfig(), users.NewMemoryRepository())
	router := server.Router()

	untrustedOrigin := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"email":"user@example.com","password":"secret"}`),
	)
	untrustedOrigin.Header.Set("Origin", "https://evil.example")
	untrustedOriginResponse := httptest.NewRecorder()
	router.ServeHTTP(untrustedOriginResponse, untrustedOrigin)
	if untrustedOriginResponse.Code != http.StatusForbidden {
		t.Fatalf("untrusted origin status = %d, want %d", untrustedOriginResponse.Code, http.StatusForbidden)
	}

	missingCSRF := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	missingCSRF.Header.Set("Origin", "http://localhost:5173")
	missingCSRF.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session"})
	missingCSRFResponse := httptest.NewRecorder()
	router.ServeHTTP(missingCSRFResponse, missingCSRF)
	if missingCSRFResponse.Code != http.StatusForbidden {
		t.Fatalf("missing CSRF token status = %d, want %d", missingCSRFResponse.Code, http.StatusForbidden)
	}
}

func TestLoginRateLimitReturns429AndRetryAfter(t *testing.T) {
	server := NewServer(testConfig(), users.NewMemoryRepository())
	router := server.Router()

	var last *httptest.ResponseRecorder
	for attempt := 0; attempt <= 5; attempt++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/auth/login",
			strings.NewReader(`{"email":"user@example.com","password":"incorrect"}`),
		)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("X-Forwarded-For", "203.0.113."+strconv.Itoa(attempt+1))
		last = httptest.NewRecorder()
		router.ServeHTTP(last, req)
	}

	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("sixth login status = %d, want %d", last.Code, http.StatusTooManyRequests)
	}
	if last.Header().Get("Retry-After") == "" {
		t.Fatal("rate limited response must include Retry-After")
	}
}

type passwordResetQueueStub struct {
	email string
	link  string
}

func (s *passwordResetQueueStub) Enqueue(_ context.Context, email string, link string) error {
	s.email = email
	s.link = link
	return nil
}

func TestPasswordRecoveryResetsPasswordOnceAndRevokesSessions(t *testing.T) {
	cfg := testConfig()
	repository := users.NewMemoryRepository()
	passwordHash, err := auth.HashPassword("old-password")
	if err != nil {
		t.Fatal(err)
	}
	user := users.User{
		ID:           users.NextUserID(),
		Email:        "alice@example.com",
		DisplayName:  "Alice",
		PasswordHash: passwordHash,
	}
	if err := repository.Create(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	sessions := store.NewSessionStore()
	sessionToken, err := sessions.Create(context.Background(), user.ID, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	sender := &passwordResetQueueStub{}
	server := NewServerWithAuthDependencies(
		cfg,
		repository,
		sessions,
		nil,
		auth.NewMemoryLoginRateLimiter(),
		auth.NewMemoryPasswordResetStore(),
		sender,
	)
	router := server.Router()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/password/forgot",
		strings.NewReader(`{"email":"alice@example.com"}`),
	)
	request.Header.Set("Origin", cfg.AppBaseURL)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("forgot password status = %d, want %d: %s", response.Code, http.StatusAccepted, response.Body.String())
	}
	if sender.email != user.Email {
		t.Fatalf("recovery email = %q, want %q", sender.email, user.Email)
	}
	resetURL, err := url.Parse(sender.link)
	if err != nil {
		t.Fatal(err)
	}
	fragmentQuery := strings.TrimPrefix(resetURL.Fragment, "/reset-password?")
	fragmentValues, err := url.ParseQuery(fragmentQuery)
	if err != nil {
		t.Fatal(err)
	}
	token := fragmentValues.Get("token")
	if token == "" {
		t.Fatal("email link has no reset token")
	}

	body := `{"token":"` + token + `","password":"new-password"}`
	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", strings.NewReader(body))
	request.Header.Set("Origin", cfg.AppBaseURL)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("reset password status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}

	if _, valid, err := sessions.Validate(context.Background(), sessionToken); err != nil || valid {
		t.Fatalf("old session valid = %v, error = %v; want revoked", valid, err)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/reset", strings.NewReader(body))
	request.Header.Set("Origin", cfg.AppBaseURL)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("reused token status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	login := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"email":"alice@example.com","password":"new-password"}`),
	)
	login.Header.Set("Origin", cfg.AppBaseURL)
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, login)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login with reset password status = %d, want %d", loginResponse.Code, http.StatusOK)
	}
}

func TestForgotPasswordDoesNotRevealUnknownEmail(t *testing.T) {
	cfg := testConfig()
	repository := users.NewMemoryRepository()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(context.Background(), users.User{
		ID:           users.NextUserID(),
		Email:        "known@example.com",
		DisplayName:  "Known",
		PasswordHash: hash,
	}); err != nil {
		t.Fatal(err)
	}
	server := NewServerWithAuthDependencies(
		cfg,
		repository,
		store.NewSessionStore(),
		nil,
		auth.NewMemoryLoginRateLimiter(),
		auth.NewMemoryPasswordResetStore(),
		&passwordResetQueueStub{},
	)
	router := server.Router()

	responses := make([]*httptest.ResponseRecorder, 2)
	for index, email := range []string{"known@example.com", "unknown@example.com"} {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/auth/password/forgot",
			strings.NewReader(`{"email":"`+email+`"}`),
		)
		request.Header.Set("Origin", cfg.AppBaseURL)
		responses[index] = httptest.NewRecorder()
		router.ServeHTTP(responses[index], request)
	}
	if responses[0].Code != http.StatusAccepted || responses[1].Code != http.StatusAccepted {
		t.Fatalf("forgot-password statuses = %d and %d; want both %d", responses[0].Code, responses[1].Code, http.StatusAccepted)
	}
	if responses[0].Body.String() != responses[1].Body.String() {
		t.Fatal("forgot-password responses must not reveal whether the email exists")
	}
}

func TestPasswordRecoveryIsUnavailableWithoutEmailSender(t *testing.T) {
	server := NewServer(testConfig(), users.NewMemoryRepository())
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/password/forgot",
		strings.NewReader(`{"email":"user@example.com"}`),
	)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	server.Router().ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("forgot password without sender status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestRegistrationRejectsInvalidEmail(t *testing.T) {
	server := NewServer(testConfig(), users.NewMemoryRepository())
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"email":"not-an-email","password":"password123","displayName":"User"}`),
	)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()
	server.Router().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("register status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
