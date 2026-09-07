package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements the token-bucket rate-limiting algorithm.
// Picture a bucket that holds up to `capacity` tokens. Tokens refill
// at a steady rate over time. Every request that wants to proceed
// must take one token. If the bucket is empty, the request is denied.
type TokenBucket struct {
	mu sync.Mutex

	capacity     int64     // max tokens the bucket can ever hold
	tokens       int64     // tokens currently available
	refillRate   int64     // tokens added per second
	lastRefill   time.Time // last time we topped up the bucket
}

// NewTokenBucket creates a bucket that starts full.
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// refill adds tokens based on how much time has passed since the last
// refill. Must be called while holding the lock — it mutates shared state.
func (b *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	tokensToAdd := int64(elapsed * float64(b.refillRate))
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > b.capacity {
			b.tokens = b.capacity // never overflow past capacity
		}
		b.lastRefill = now
	}
}

// Allow reports whether a request may proceed right now, and if so,
// consumes one token.
//
// IMPORTANT: refill, the check (tokens > 0), and the decrement all
// happen under ONE lock acquisition. This is deliberate — see Day 2's
// TOCTOU discussion. If check and decrement were two separate locked
// sections, another goroutine could slip in between them and both
// callers could be allowed through on a single remaining token.
func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()

	if b.tokens > 0 {
		b.tokens--
		return true
	}
	return false
}

// Tokens returns the current token count — useful for tests/debugging.
func (b *TokenBucket) Tokens() int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill()
	return b.tokens
}