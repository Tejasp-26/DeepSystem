package ratelimit

import "sync"

// Registry holds one TokenBucket per key (e.g. per API client),
// created lazily on first use.
type Registry struct {
	mu      sync.Mutex
	buckets map[string]*TokenBucket

	defaultCapacity float64
	defaultRefill   float64
}

func NewRegistry(capacity, refillRate float64) *Registry {
	return &Registry{
		buckets:         make(map[string]*TokenBucket),
		defaultCapacity: capacity,
		defaultRefill:   refillRate,
	}
}

func (r *Registry) getOrCreate(key string) *TokenBucket {
	r.mu.Lock()
	defer r.mu.Unlock()

	tb, exists := r.buckets[key]
	if !exists {
		tb = NewTokenBucket(r.defaultCapacity, r.defaultRefill)
		r.buckets[key] = tb
	}
	return tb
}

// Allow checks whether a request for `key` should be permitted.
func (r *Registry) Allow(key string) bool {
	return r.getOrCreate(key).Allow()
}