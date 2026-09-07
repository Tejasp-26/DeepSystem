package ratelimit

import "sync"

// Registry holds one TokenBucket per key (e.g. per API key, per client IP).
// This lets each caller have their own independent rate limit.
type Registry struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket

	// defaults used when creating a bucket for a key we haven't seen yet
	defaultCapacity   int64
	defaultRefillRate int64
}

func NewRegistry(defaultCapacity, defaultRefillRate int64) *Registry {
	return &Registry{
		buckets:           make(map[string]*TokenBucket),
		defaultCapacity:   defaultCapacity,
		defaultRefillRate: defaultRefillRate,
	}
}

// GetBucket returns the bucket for `key`, creating one with default
// settings if this is the first time we've seen this key.
func (r *Registry) GetBucket(key string) *TokenBucket {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, exists := r.buckets[key]
	if !exists {
		b = NewTokenBucket(r.defaultCapacity, r.defaultRefillRate)
		r.buckets[key] = b
	}
	return b
}

// Allow is a convenience wrapper: look up (or create) the bucket for
// key, then check if it allows the request.
func (r *Registry) Allow(key string) bool {
	bucket := r.GetBucket(key)
	return bucket.Allow()
}