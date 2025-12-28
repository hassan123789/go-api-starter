// Package resilience provides rate limiting mechanisms.
package resilience

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides rate limiting using the token bucket algorithm.
type RateLimiter struct {
	rate       float64   // tokens per second
	burst      int       // maximum burst size
	tokens     float64   // current tokens
	lastUpdate time.Time // last token update time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter.
// rate: tokens per second
// burst: maximum tokens that can accumulate
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so.
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if n requests are allowed.
func (rl *RateLimiter) AllowN(n int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Calculate new tokens
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens += elapsed * rl.rate
	if rl.tokens > float64(rl.burst) {
		rl.tokens = float64(rl.burst)
	}
	rl.lastUpdate = now

	// Check if we have enough tokens
	if rl.tokens < float64(n) {
		return false
	}

	rl.tokens -= float64(n)
	return true
}

// Wait blocks until a token is available or context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.WaitN(ctx, 1)
}

// WaitN blocks until n tokens are available.
func (rl *RateLimiter) WaitN(ctx context.Context, n int) error {
	for {
		if rl.AllowN(n) {
			return nil
		}

		// Calculate wait time
		rl.mu.Lock()
		needed := float64(n) - rl.tokens
		waitTime := time.Duration(needed/rl.rate*1000) * time.Millisecond
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// Tokens returns the current number of available tokens.
func (rl *RateLimiter) Tokens() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	tokens := rl.tokens + elapsed*rl.rate
	if tokens > float64(rl.burst) {
		return float64(rl.burst)
	}
	return tokens
}

// KeyedRateLimiter provides per-key rate limiting.
type KeyedRateLimiter struct {
	rate     float64
	burst    int
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	cleanup  time.Duration // how often to clean up old entries
}

// NewKeyedRateLimiter creates a new keyed rate limiter.
func NewKeyedRateLimiter(rate float64, burst int, cleanup time.Duration) *KeyedRateLimiter {
	krl := &KeyedRateLimiter{
		rate:     rate,
		burst:    burst,
		limiters: make(map[string]*RateLimiter),
		cleanup:  cleanup,
	}

	// Start cleanup goroutine
	go krl.cleanupLoop()

	return krl
}

// Allow checks if a request for the given key is allowed.
func (krl *KeyedRateLimiter) Allow(key string) bool {
	return krl.AllowN(key, 1)
}

// AllowN checks if n requests for the given key are allowed.
func (krl *KeyedRateLimiter) AllowN(key string, n int) bool {
	limiter := krl.getLimiter(key)
	return limiter.AllowN(n)
}

// getLimiter gets or creates a rate limiter for the given key.
func (krl *KeyedRateLimiter) getLimiter(key string) *RateLimiter {
	krl.mu.RLock()
	limiter, ok := krl.limiters[key]
	krl.mu.RUnlock()

	if ok {
		return limiter
	}

	krl.mu.Lock()
	defer krl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok = krl.limiters[key]; ok {
		return limiter
	}

	limiter = NewRateLimiter(krl.rate, krl.burst)
	krl.limiters[key] = limiter
	return limiter
}

// cleanupLoop periodically removes idle rate limiters.
func (krl *KeyedRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(krl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		krl.mu.Lock()
		// Remove limiters that are at full capacity (idle)
		for key, limiter := range krl.limiters {
			if limiter.Tokens() >= float64(krl.burst) {
				delete(krl.limiters, key)
			}
		}
		krl.mu.Unlock()
	}
}

// Size returns the number of active rate limiters.
func (krl *KeyedRateLimiter) Size() int {
	krl.mu.RLock()
	defer krl.mu.RUnlock()
	return len(krl.limiters)
}
