package auth

import (
	"context"
	"strconv"
	"testing"
	"time"
)

func TestMemoryLoginRateLimiterLimitsEmailAndIP(t *testing.T) {
	limiter := NewMemoryLoginRateLimiter()

	for attempt := 1; attempt <= LoginPairLimit; attempt++ {
		retry, err := limiter.Check(context.Background(), "192.0.2.1", "user@example.com")
		if err != nil || retry != 0 {
			t.Fatalf("attempt %d retry = %s, error = %v; want allowed", attempt, retry, err)
		}
	}

	retry, err := limiter.Check(context.Background(), "192.0.2.1", "user@example.com")
	if err != nil || retry <= 0 {
		t.Fatalf("over-limit retry = %s, error = %v; want positive retry delay", retry, err)
	}

	ipLimiter := NewMemoryLoginRateLimiter()
	for attempt := 1; attempt <= LoginIPLimit; attempt++ {
		retry, err = ipLimiter.Check(context.Background(), "192.0.2.2", strconv.Itoa(attempt)+"@example.com")
		if err != nil || retry != 0 {
			t.Fatalf("IP attempt %d retry = %s, error = %v; want allowed", attempt, retry, err)
		}
	}

	retry, err = ipLimiter.Check(context.Background(), "192.0.2.2", "over-limit@example.com")
	if err != nil || retry <= 0 {
		t.Fatalf("over-IP-limit retry = %s, error = %v; want positive retry delay", retry, err)
	}
}

func TestRateLimitKeyIsDeterministicAndWindowScoped(t *testing.T) {
	window := time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC)
	first := RateLimitKey("pair", "user@example.com", window)
	if first != RateLimitKey("pair", "user@example.com", window) {
		t.Fatal("same values produced different rate limit keys")
	}
	if first == RateLimitKey("pair", "other@example.com", window) {
		t.Fatal("different email values produced the same rate limit key")
	}
	keyed := RateLimitKeyWithSecret("secret-one", "pair", "user@example.com", window)
	if keyed == RateLimitKeyWithSecret("secret-two", "pair", "user@example.com", window) {
		t.Fatal("different secrets produced the same keyed rate limit key")
	}
}

func TestMemoryLoginRateLimiterResetsAtWindowBoundary(t *testing.T) {
	limiter := NewMemoryLoginRateLimiter()
	now := time.Date(2026, 11, 17, 15, 0, 0, 0, time.UTC)
	limiter.now = func() time.Time { return now }

	if _, err := limiter.Check(context.Background(), "192.0.2.1", "user@example.com"); err != nil {
		t.Fatal(err)
	}
	now = now.Add(LoginWindowSize)
	retry, err := limiter.Check(context.Background(), "192.0.2.1", "user@example.com")
	if err != nil || retry != 0 {
		t.Fatalf("first attempt after window rollover retry = %s, error = %v; want allowed", retry, err)
	}
}

func TestPasswordRecoveryRateLimitIsIndependentFromLogin(t *testing.T) {
	limiter := NewMemoryLoginRateLimiter()
	for attempt := 1; attempt <= LoginPairLimit; attempt++ {
		retry, err := limiter.CheckPasswordRecovery(context.Background(), "192.0.2.1", "user@example.com")
		if err != nil || retry != 0 {
			t.Fatalf("recovery attempt %d retry = %s, error = %v; want allowed", attempt, retry, err)
		}
	}
	retry, err := limiter.CheckPasswordRecovery(context.Background(), "192.0.2.1", "user@example.com")
	if err != nil || retry <= 0 {
		t.Fatalf("over-limit recovery retry = %s, error = %v; want positive retry delay", retry, err)
	}
	if retry, err := limiter.Check(context.Background(), "192.0.2.1", "user@example.com"); err != nil || retry != 0 {
		t.Fatalf("login limit was affected by recovery attempts: retry = %s, error = %v", retry, err)
	}
}
