package quixote

import (
	"strings"
	"encoding/json"
)


type Stats struct {
	CacheHitCount int // Count of requests that were fulfilled from cached items before their soft expiration
	CacheMissCount int // Count of items that were fulfilled by calling the queryFunc
	CacheRescueCount int // Count of requests that were fulfilled from cached items after their soft expiration
	UnexpiredEvictionCount int // Count of items that were evicted before reaching their hard expiration.
}

// String encodes the Stats object as JSON.
func (s Stats) String() string {
	sb := &strings.Builder{}
	encoder := json.NewEncoder(sb)
	err := encoder.Encode(s)
	if err != nil {
		panic("Error: failed to encode stats")
	} else {
		return sb.String()
	}
}
