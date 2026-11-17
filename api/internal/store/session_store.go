package store

import (
	"context"
	"sync"
	"time"

	"yearplanner/internal/auth"
)

type Session struct {
	UserID    string    `bson:"userId"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type SessionRepository interface {
	Create(ctx context.Context, userID string, ttl time.Duration) (string, error)
	Validate(ctx context.Context, token string) (string, bool, error)
	Revoke(ctx context.Context, token string) error
	RevokeUser(ctx context.Context, userID string) error
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]Session)}
}

func (s *SessionStore) Create(_ context.Context, userID string, ttl time.Duration) (string, error) {
	token, err := auth.NewToken()
	if err != nil {
		return "", err
	}

	hash := auth.HashToken(token)
	expiresAt := time.Now().Add(ttl)

	s.mu.Lock()
	s.sessions[hash] = Session{UserID: userID, ExpiresAt: expiresAt}
	s.mu.Unlock()

	return token, nil
}

func (s *SessionStore) Validate(ctx context.Context, token string) (string, bool, error) {
	if token == "" {
		return "", false, nil
	}

	hash := auth.HashToken(token)

	s.mu.RLock()
	session, ok := s.sessions[hash]
	s.mu.RUnlock()
	if !ok {
		return "", false, nil
	}

	if !time.Now().Before(session.ExpiresAt) {
		if err := s.Revoke(ctx, token); err != nil {
			return "", false, err
		}
		return "", false, nil
	}

	return session.UserID, true, nil
}

func (s *SessionStore) Revoke(_ context.Context, token string) error {
	if token == "" {
		return nil
	}

	s.mu.Lock()
	delete(s.sessions, auth.HashToken(token))
	s.mu.Unlock()
	return nil
}

func (s *SessionStore) RevokeUser(_ context.Context, userID string) error {
	s.mu.Lock()
	for key, session := range s.sessions {
		if session.UserID == userID {
			delete(s.sessions, key)
		}
	}
	s.mu.Unlock()
	return nil
}
