package stats

import (
	"sync"
	"testing"
)

// TestIncrement_ConcurrentSafe hammers the same key from many goroutines
// at once. Run with -race: this must produce ZERO race warnings, and
// the final count must exactly equal goroutines * incrementsEach —
// proving no increments were silently lost.
func TestIncrement_ConcurrentSafe(t *testing.T) {
	const goroutines = 100
	const incrementsEach = 1000

	counter := NewExactCounter()
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < incrementsEach; j++ {
				counter.Increment("samekey", "value-from-goroutine")
			}
		}(i)
	}

	wg.Wait()

	got := counter.Count("samekey")
	want := int64(goroutines * incrementsEach)

	if got != want {
		t.Fatalf("expected count %d, got %d — increments were lost", want, got)
	}
}