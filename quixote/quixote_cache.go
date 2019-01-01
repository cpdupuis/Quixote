package quixote

import (
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
}


// Cache is the main entry point for QuixoteCache.
type Cache struct {
	queryFunc    func(string) (string, bool) // returns the content and whether there was any content
	index        map[string]*cacheItem       // protected by mutex.
	timeline	 ExpiryTimeline	// timeline for expiring items in the cache
	count        int                         // number of items currently in the cache
	maxCount	int	// maximum number of items in the cache
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
func MakeQuixoteCache(queryFunc func(string) (string, bool), softLimit time.Duration, hardLimit time.Duration, maxCount int) *Cache {
	if maxCount < 2 {
		panic("maxCount must be at least 2.")
	}
	if softLimit > hardLimit {
		panic("hardLimit must be longer than softLimit")
	}
	return &Cache{
		index:       make(map[string]*cacheItem),
		queryFunc:   queryFunc,
		maxCount: maxCount,
		softLimit:   softLimit,
		hardLimit:   hardLimit,
		timeline:	MakeExpiryTimeline(256, hardLimit),
	}
}

// refresh is an internal function to retrieve a fresh value and update the cache with it.
func (c *Cache) refresh(key string, now time.Time, timeOld time.Time) (string, bool) {
	val, ok := c.queryFunc(key)
	if ok {
		ci := &cacheItem{value: val, createTime: now}
		c.mutex.Lock()
		c.index[key] = ci
		// Clear expired items
		c.timeline.ExpireItems(now, func(akey string) {
			delete(c.index, akey)
		})
		// And replace!
		c.timeline.ReplaceItem(key, timeOld, now)
		c.mutex.Unlock()
		return val, true
	} else {
		// Clear expired items
		c.mutex.Lock()
		c.timeline.ExpireItems(now, func(akey string) {
			delete(c.index, akey)
		})
		c.mutex.Unlock()
		
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
	var cacheMiss int
	var cacheHit int
	var cacheRescue int
	var cacheNoRoom int

	var result string
	var resOk bool
	if cacheVal != nil {
		since := now.Sub(cacheVal.createTime)
		if since < c.softLimit {
			// Just return the cached value
			cacheHit = 1
			result,resOk =  cacheVal.value, true
		} else {
			val, ok := c.refresh(key, now, cacheVal.createTime)
			if ok {
				cacheMiss = 1
				// Hey, we refreshed successfully!
				result,resOk = val, true
			} else if since < c.hardLimit {
				cacheRescue = 1
				// We didn't refresh, but it's still OK.
				result,resOk = cacheVal.value, true
			} else {
				// Cached value is too old. (Will be cleaned up eventually by the expiryTimeline.)
				cacheMiss = 1
				result,resOk = val,ok
			}
		}
	} else {
	// The cache didn't help us
		cacheMiss = 1
		if c.count < c.maxCount {
			// We have space, let's put it in our cache!
			result,resOk = c.refresh(key, now, now)
		} else {
			// We're out of space, just return the result without caching.
			cacheNoRoom = 1
			result,resOk = c.queryFunc(key)
		}
	}
	c.mutex.Lock()
	c.stats.CacheHitCount += cacheHit
	c.stats.CacheMissCount += cacheMiss
	c.stats.CacheNoRoomCount += cacheNoRoom
	c.stats.CacheRescueCount += cacheRescue
	c.mutex.Unlock()
	return result,resOk
}


// Stats returns current statistics about cache performance.
func (c *Cache) Stats() Stats {
	return c.stats
}
