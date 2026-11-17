package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrInvalidResetToken = errors.New("invalid or expired password reset token")

const PasswordResetTTL = time.Hour

type PasswordResetStore interface {
	Create(ctx context.Context, userID string, tokenHash string, ttl time.Duration) error
	Consume(ctx context.Context, tokenHash string) (string, error)
}

type PasswordResetEmailSender interface {
	SendPasswordReset(ctx context.Context, email string, link string) error
}

type PasswordResetEmailQueue interface {
	Enqueue(ctx context.Context, email string, link string) error
}

type resetToken struct {
	userID    string
	expiresAt time.Time
}

type MemoryPasswordResetStore struct {
	mu     sync.Mutex
	tokens map[string]resetToken
}

func NewMemoryPasswordResetStore() *MemoryPasswordResetStore {
	return &MemoryPasswordResetStore{tokens: make(map[string]resetToken)}
}

func (s *MemoryPasswordResetStore) Create(_ context.Context, userID string, tokenHash string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for hash, token := range s.tokens {
		if token.userID == userID {
			delete(s.tokens, hash)
		}
	}
	s.tokens[tokenHash] = resetToken{userID: userID, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (s *MemoryPasswordResetStore) Consume(_ context.Context, tokenHash string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token, exists := s.tokens[tokenHash]
	delete(s.tokens, tokenHash)
	if !exists || !time.Now().Before(token.expiresAt) {
		return "", ErrInvalidResetToken
	}
	return token.userID, nil
}
