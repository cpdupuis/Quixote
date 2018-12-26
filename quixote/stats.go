package quixote

import (
	"fmt"
)
type Stats struct {
	UnexpiredEvictionCount int
	CacheHitCount int
	CacheMissCount int
	CacheRescueCount int
}

func (s Stats) Dump() {
	fmt.Printf("Unexpired eviction count: %d\n", s.UnexpiredEvictionCount)
	calls := s.CacheHitCount + s.CacheMissCount + s.CacheRescueCount
	fmt.Printf("Cache hit count: %d Percent: %d\n", s.CacheHitCount, (s.CacheHitCount*100)/calls)
	fmt.Printf("Cache miss count: %d Percent: %d\n", s.CacheMissCount, (s.CacheMissCount*100)/calls)
	fmt.Printf("Cache rescue count: %d Percent: %d\n", s.CacheRescueCount, (s.CacheRescueCount*100)/calls)
}
