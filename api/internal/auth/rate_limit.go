package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

const (
	LoginIPLimit    = 20
	LoginPairLimit  = 5
	LoginWindowSize = 15 * time.Minute
)

type LoginRateLimiter interface {
	Check(ctx context.Context, ipAddress string, email string) (time.Duration, error)
	CheckPasswordRecovery(ctx context.Context, ipAddress string, email string) (time.Duration, error)
}

type memoryRateBucket struct {
	count     int
	expiresAt time.Time
}

type MemoryLoginRateLimiter struct {
	mu      sync.Mutex
	buckets map[string]memoryRateBucket
	now     func() time.Time
}

func NewMemoryLoginRateLimiter() *MemoryLoginRateLimiter {
	return &MemoryLoginRateLimiter{
		buckets: make(map[string]memoryRateBucket),
		now:     time.Now,
	}
}

func (l *MemoryLoginRateLimiter) Check(_ context.Context, ipAddress string, email string) (time.Duration, error) {
	return l.check(ipAddress, email, "login")
}

func (l *MemoryLoginRateLimiter) CheckPasswordRecovery(_ context.Context, ipAddress string, email string) (time.Duration, error) {
	return l.check(ipAddress, email, "recovery")
}

func (l *MemoryLoginRateLimiter) check(ipAddress string, email string, scope string) (time.Duration, error) {
	now := l.now()
	windowEnd := now.Truncate(LoginWindowSize).Add(LoginWindowSize)
	ipKey := RateLimitKey(scope+"-ip", ipAddress, windowEnd)
	pairKey := RateLimitKey(scope+"-pair", ipAddress+"\x00"+email, windowEnd)

	l.mu.Lock()
	defer l.mu.Unlock()
	for key, bucket := range l.buckets {
		if !now.Before(bucket.expiresAt) {
			delete(l.buckets, key)
		}
	}

	ipCount := l.increment(ipKey, windowEnd)
	if ipCount > LoginIPLimit {
		return windowEnd.Sub(now), nil
	}
	pairCount := l.increment(pairKey, windowEnd)
	if pairCount > LoginPairLimit {
		return windowEnd.Sub(now), nil
	}
	return 0, nil
}

func (l *MemoryLoginRateLimiter) increment(key string, expiresAt time.Time) int {
	bucket := l.buckets[key]
	bucket.count++
	bucket.expiresAt = expiresAt
	l.buckets[key] = bucket
	return bucket.count
}

func RateLimitKey(scope string, value string, windowEnd time.Time) string {
	sum := sha256.Sum256([]byte(value))
	return scope + ":" + hex.EncodeToString(sum[:]) + ":" + windowEnd.UTC().Format(time.RFC3339)
}

func RateLimitKeyWithSecret(secret string, scope string, value string, windowEnd time.Time) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(value))
	return scope + ":" + hex.EncodeToString(mac.Sum(nil)) + ":" + windowEnd.UTC().Format(time.RFC3339)
}
