package store

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestPasswordResetEmailQueueEncryptsLinksAndBindsRecipient(t *testing.T) {
	queue := &MongoPasswordResetEmailQueue{key: sha256.Sum256([]byte("test-secret"))}
	link := "https://planner.example.com/reset-password?token=secret-token"
	encrypted, err := queue.encrypt("user@example.com", link)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(encrypted, []byte(link)) {
		t.Fatal("encrypted queue payload contains the reset link in plaintext")
	}

	decrypted, err := queue.decrypt("user@example.com", encrypted)
	if err != nil || decrypted != link {
		t.Fatalf("decrypted link = %q, error = %v", decrypted, err)
	}
	if _, err := queue.decrypt("other@example.com", encrypted); err == nil {
		t.Fatal("queue payload decrypted for a different recipient")
	}

	encrypted[len(encrypted)-1] ^= 1
	if _, err := queue.decrypt("user@example.com", encrypted); err == nil {
		t.Fatal("tampered queue payload decrypted successfully")
	}
}
