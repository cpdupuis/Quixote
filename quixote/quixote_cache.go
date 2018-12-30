package quixote

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// QuixoteCache is just a basic cache.
// Eviction policy is oldest-first. Supports hard limit on age of items, as well as soft
// limit. When the soft limit is reached, the cache will attempt to refresh the item from the
// source, but it will return the cached item if the refresh fails.
type cacheItem struct {
	value      string
	createTime time.Time
	id         uint64
}

// QuixoteCache is the public interface, returned by MakeQuixoteCache
type QuixoteCache interface {
	Get(string) (string, bool)
	Dump()
	Stats() Stats
}

type Cache struct {
	queryFunc    func(string) (string, bool) // returns the content and whether there was any content
	index        map[string]*cacheItem       // protected by mutex.
	timeline	 *ExpiryTimeline
	count        int                         // number of items currently in the cache
	stats        Stats                       // cache statistics
	softLimit    time.Duration               // after the softLimit is passed, Get will query to get a fresh value, though it will return a cached value if an error occurs in the query
	hardLimit    time.Duration               // after the hardLimit, the cached value is removed.
	mutex        sync.RWMutex                // mutex guarding index, timeline, timelineHead, timelineTail, count, and stats
}

// MakeQuixoteCache creates a new cache that calls queryFunc to get results from the source.
// Parameters:
// - queryFunc: the func that calls the source service to get current values.
// - softLimit: the age at which an item is considered stale and needing to be refreshed.
// - hardLimit: the age at which a cached item will be removed from the cache.
// - maxCount: the maximum number of items that can be stored in this cache.
func MakeQuixoteCache(queryFunc func(string) (string, bool), softLimit time.Duration, hardLimit time.Duration, maxCount int) QuixoteCache {
	if maxCount < 2 {
		panic("maxCount must be at least 2.")
	}
	if softLimit > hardLimit {
		panic("hardLimit must be longer than softLimit")
	}
	return &Cache{
		index:       make(map[string]*cacheItem),
		queryFunc:   queryFunc,
		softLimit:   softLimit,
		hardLimit:   hardLimit,
		timeline:	MakeExpiryTimeline(256, hardLimit),
	}
}

// refresh is an internal function to retrieve a fresh value and update the cache with it.
func (c *Cache) refresh(key string, now time.Time, timeOld time.Time) (string, bool) {
	val, ok := c.queryFunc(key)
	if ok {
		id := rand.Uint64()
		ci := &cacheItem{value: val, createTime: now, id: id}
		c.mutex.Lock()
		c.index[key] = ci
		// TODO: update timeline!
		c.timeline.ReplaceItem(&key, timeOld, now)
		c.mutex.Unlock()
		return val, true
	} else {
		// no cached result, sorry
		return "", false
	}
}

// Get is the public interface for Quixote. Get fetches fresh values from the source
// as needed and stores them to satisfy future requests, and it returns a value if available.
func (c *Cache) Get(key string) (string, bool) {
	c.mutex.RLock()
	cacheVal := c.index[key]
	c.mutex.RUnlock()
	now := time.Now()
	if cacheVal != nil {
		since := now.Sub(cacheVal.createTime)
		if since < c.softLimit {
			// Just return the cached value
			c.stats.CacheHitCount++
			return cacheVal.value, true
		} else {
			val, ok := c.refresh(key, now, cacheVal.createTime)
			if ok {
				c.stats.CacheMissCount++
				// Hey, we refreshed successfully!
				return val, true
			} else {
				c.stats.CacheRescueCount++
				// We didn't refresh, but it's still OK.
				return cacheVal.value, true
			}
		}
	}
	c.stats.CacheMissCount++
	// fallthrough: the cache didn't help us
	return c.refresh(key, now, now)
}

// Dump prints out details of the cache state for debugging purposes
func (c *Cache) Dump() {
	fmt.Printf("Index:\n")
	for k, v := range c.index {
		fmt.Printf("  %s : %s\n", k, v.value)
	}
}

// Stats returns current statistics about cache performance.
func (c *Cache) Stats() Stats {
	return c.stats
}
