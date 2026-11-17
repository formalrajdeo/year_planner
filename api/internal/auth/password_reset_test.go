package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryPasswordResetStoreConsumesTokenOnce(t *testing.T) {
	store := NewMemoryPasswordResetStore()
	if err := store.Create(context.Background(), "user-1", "hash-1", time.Hour); err != nil {
		t.Fatal(err)
	}

	userID, err := store.Consume(context.Background(), "hash-1")
	if err != nil || userID != "user-1" {
		t.Fatalf("first consume userID = %q, error = %v", userID, err)
	}
	if _, err := store.Consume(context.Background(), "hash-1"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("second consume error = %v; want invalid token", err)
	}
}

func TestMemoryPasswordResetStoreExpiresAndReplacesTokens(t *testing.T) {
	store := NewMemoryPasswordResetStore()
	if err := store.Create(context.Background(), "user-1", "old-hash", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := store.Create(context.Background(), "user-1", "new-hash", time.Hour); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume(context.Background(), "old-hash"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("replaced token error = %v; want invalid token", err)
	}
	if err := store.Create(context.Background(), "user-2", "expired-hash", -time.Second); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Consume(context.Background(), "expired-hash"); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("expired token error = %v; want invalid token", err)
	}
}
