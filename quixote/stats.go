package quixote

import (
	"fmt"
)


type Stats struct {
	CacheHitCount int // Count of requests that were fulfilled from cached items before their soft expiration
	CacheMissCount int // Count of items that were fulfilled by calling the queryFunc
	CacheRescueCount int // Count of requests that were fulfilled from cached items after their soft expiration
	UnexpiredEvictionCount int // Count of items that were evicted before reaching their hard expiration.
}

// Output the statistics.
func (s Stats) Dump() {
	calls := s.CacheHitCount + s.CacheMissCount + s.CacheRescueCount
	fmt.Printf("Cache hit count: %d Percent: %d\n", s.CacheHitCount, (s.CacheHitCount*100)/calls)
	fmt.Printf("Cache miss count: %d Percent: %d\n", s.CacheMissCount, (s.CacheMissCount*100)/calls)
	fmt.Printf("Cache rescue count: %d Percent: %d\n", s.CacheRescueCount, (s.CacheRescueCount*100)/calls)
	fmt.Printf("Unexpired eviction count: %d\n", s.UnexpiredEvictionCount)
}
