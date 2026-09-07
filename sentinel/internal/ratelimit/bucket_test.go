package ratelimit

import (
	"sync"
	"testing"
)

// TestAllow_TOCTOU fires many goroutines at a bucket that starts with
// exactly 1 token, all at roughly the same instant. With the fix in
// place (check+decrement as one atomic step), exactly 1 should ever
// return true, no matter how many goroutines race for it.
func TestAllow_TOCTOU(t *testing.T) {
	const goroutines = 50

	// refillRate = 0 so no new tokens sneak in mid-test and confuse the count.
	bucket := NewTokenBucket(1, 0)

	var wg sync.WaitGroup
	start := make(chan struct{}) // closed to release all goroutines at once

	results := make([]bool, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start // block here until the signal fires — maximizes overlap
			results[idx] = bucket.Allow()
		}(i)
	}

	close(start) // release every goroutine at once
	wg.Wait()

	allowedCount := 0
	for _, r := range results {
		if r {
			allowedCount++
		}
	}

	if allowedCount != 1 {
		t.Fatalf("expected exactly 1 allowed request, got %d — TOCTOU bug present", allowedCount)
	}
}