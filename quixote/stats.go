package quixote

import (
	"strings"
	"encoding/json"
)


type Stats struct {
	CacheHitCount int		// Count of requests that were fulfilled from cached items before their soft expiration
	CacheMissCount int		// Count of items that were fulfilled by calling the queryFunc
	CacheRescueCount int	// Count of requests that were fulfilled from cached items after their soft expiration
	CacheNoRoomCount int	// Count of responses that were not cached due to the cache being full.
	CacheNewItemCount int // Count of new items that were added to the cache 
	CacheRequestFailCount int // Count of times the request has simply failed
	ExplicitInvalidationCount int // Count of items that were explicitly invalidated
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
