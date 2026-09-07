package stats

import "sync"

// ExactCounter tracks exact hit counts and exact distinct-value counts
// per key. This is our ground truth: on Day 4 we build a Count-Min
// Sketch and a HyperLogLog that approximate these same two things
// using far less memory — and we measure their error against
// whatever ExactCounter says is really true.
type ExactCounter struct {
	mu sync.RWMutex

	// counts["endpoint:/login"] = 4821  → total hits for that key
	counts map[string]int64

	// uniques["endpoint:/login"] = {ip1, ip2, ...} → distinct values seen
	uniques map[string]map[string]struct{}
}

func NewExactCounter() *ExactCounter {
	return &ExactCounter{
		counts:  make(map[string]int64),
		uniques: make(map[string]map[string]struct{}),
	}
}

// Increment records one hit for `key`, and marks `value` as seen under
// that key. Example: Increment("endpoint:/login", "203.0.113.7")
// means "the /login endpoint got hit once more, and this particular
// IP is one of the distinct callers we've seen."
//
// Write path — takes the full (exclusive) lock because it mutates
// both maps.
func (c *ExactCounter) Increment(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.counts[key]++

	if c.uniques[key] == nil {
		c.uniques[key] = make(map[string]struct{})
	}
	c.uniques[key][value] = struct{}{}
}

// Count returns the exact hit count for key.
// Read path — takes a read lock, so many callers can read concurrently.
func (c *ExactCounter) Count(key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.counts[key]
}

// UniqueCount returns the exact number of distinct values seen for key.
func (c *ExactCounter) UniqueCount(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.uniques[key])
}

// Snapshot returns a copy of all counts. We copy rather than hand back
// the live map so callers (like a JSON encoder) can't read the map
// while a concurrent Increment is mutating it, and so they don't need
// to hold our lock themselves.
func (c *ExactCounter) Snapshot() map[string]int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		out[k] = v
	}
	return out
}