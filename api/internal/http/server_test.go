package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yearplanner/internal/config"
	"yearplanner/internal/store"
	"yearplanner/internal/users"
)

func TestAuthFlow(t *testing.T) {
	cfg := config.Config{AppEnv: "test", AppPort: "8080", SessionSecret: "test-secret"}
	server := NewServer(cfg, users.NewMemoryRepository())
	router := server.Router()

	t.Run("register and fetch profile", func(t *testing.T) {
		body := `{"email":"alice@example.com","password":"secret123","displayName":"Alice"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("register status = %d, want %d: %s", rec.Code, http.StatusCreated, rec.Body.String())
		}

		var authResp AuthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &authResp); err != nil {
			t.Fatalf("unable to decode register response: %v", err)
		}
		if authResp.Token == "" {
			t.Fatal("expected token in register response")
		}
		if authResp.User.Email != "alice@example.com" {
			t.Fatalf("user email = %q, want %q", authResp.User.Email, "alice@example.com")
		}

		req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+authResp.Token)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("profile status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("reject invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("invalid token status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("login and logout flow", func(t *testing.T) {
		body := `{"email":"alice@example.com","password":"secret123"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("login status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		var authResp AuthResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &authResp); err != nil {
			t.Fatalf("unable to decode login response: %v", err)
		}
		if authResp.Token == "" {
			t.Fatal("expected login token")
		}

		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+authResp.Token)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("logout status = %d, want %d: %s", rec.Code, http.StatusOK, rec.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+authResp.Token)
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("revoked token status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
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
