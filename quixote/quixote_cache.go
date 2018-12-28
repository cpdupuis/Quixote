package quixote

import (
	"fmt"
	"sync"
	"time"
	"math/rand"
)

// QuixoteCache is just a basic cache.
// Eviction policy is oldest-first. Supports hard limit on age of items, as well as soft
// limit. When the soft limit is reached, the cache will attempt to refresh the item from the
// source, but it will return the cached item if the refresh fails.
type cacheItem struct {
	value string
	createTime time.Time
	id uint64
}

// QuixoteCache is the public interface, returned by MakeQuixoteCache
type QuixoteCache interface {
	Get(string) (string,bool)
	Dump()
	Stats() Stats
}

// timelineItem is an entry in the timeline circular buffer
type timelineItem struct {
	key *string // Pointer to the item's key
	id uint64 // id matches a particular item value
}

type cache struct {
	queryFunc func(string) (string,bool) // returns the content and whether there was any content
	index map[string]*cacheItem // protected by mutex.
	timeline []timelineItem // circular buffer for managing hard expiration
	timelineHead int // start of timeline
	timelineTail int // one past end of timeline
	timelineLen int // max number of elements
	count int // number of items currently in the cache
	stats Stats // cache statistics
	softLimit time.Duration // after the softLimit is passed, Get will query to get a fresh value, though it will return a cached value if an error occurs in the query
	hardLimit time.Duration // after the hardLimit, the cached value is removed.
	mutex sync.RWMutex // mutex guarding index, timeline, timelineHead, timelineTail, count, and stats
}

// MakeQuixoteCache creates a new cache that calls queryFunc to get results from the source.
// Parameters:
// - queryFunc: the func that calls the source service to get current values.
// - softLimit: the age at which an item is considered stale and needing to be refreshed.
// - hardLimit: the age at which a cached item will be removed from the cache.
// - maxCount: the maximum number of items that can be stored in this cache.
func MakeQuixoteCache(queryFunc func(string) (string,bool), softLimit time.Duration, hardLimit time.Duration, maxCount int) QuixoteCache {
	if maxCount < 2 {
		panic("maxCount must be at least 2.")
	}
	if softLimit > hardLimit {
		panic("hardLimit must be longer than softLimit")
	}
	return &cache{
		index: make(map[string]*cacheItem), 
		queryFunc: queryFunc, 
		softLimit: softLimit, 
		hardLimit: hardLimit,
		timelineLen: maxCount,
		timeline: make([]timelineItem, maxCount),
	}
}


// freeHead is an internal method to free the item currently referenced in the first entry of the timeline circular buffer. Must
// be called inside the write-lock
func (c *cache) freeHead(freeItem bool) {
	if freeItem {
		delete(c.index, *(c.timeline[c.timelineHead].key))
	}
	c.timeline[c.timelineHead] = timelineItem{key:nil,id:0}
	c.timelineHead = (c.timelineHead + 1) % c.timelineLen
	c.count--
}

// makeSpace is an internal method to free up space in the cache. Must be called inside the write-lock.
func (c *cache) makeSpace(now time.Time) {
	// first, clean up any entries past the hard limit
	for {
		// stop if the timeline is empty
		if c.count == 0 {
			return
		}
		key := c.timeline[c.timelineHead].key
		ci := c.index[*key]
		if ci == nil {
			c.freeHead(false)
		} else if ci.id != c.timeline[c.timelineHead].id {
			// don't free the item! This is an obsoleve timeline item
			c.freeHead(false)
		} else if ci.createTime.Add(c.hardLimit).Before(now) {
			// item is past its limit. Let's free it up!
			c.freeHead(true)
		} else {
			break
		}
	}
	// If there is still no space, let's make some.
	if c.timelineTail == c.timelineHead {
		c.stats.UnexpiredEvictionCount++
		c.freeHead(true)
	}
}

// addToTimeline is an internal method to write a reference to the given key and id as the new last element of the timeline
// circular buffer. Must be called witin the write-lock.
func (c *cache) addToTimeline(key *string, id uint64) {
	c.timeline[c.timelineTail].key = key
	c.timeline[c.timelineTail].id = id
	c.timelineTail = (c.timelineTail + 1) % c.timelineLen
	c.count++
}

// refresh is an internal function to retrieve a fresh value and update the cache with it.
func (c *cache) refresh(key string, now time.Time) (string,bool) {
	val,ok := c.queryFunc(key)
	if ok {
		id := rand.Uint64()
		ci := &cacheItem{value:val, createTime: now, id: id}
		c.mutex.Lock()
		c.makeSpace(now)
		c.index[key] = ci
		c.addToTimeline(&key, id)
		c.mutex.Unlock()
		return val, true
	} else {
		// no cached result, sorry
		return "",false
	}
}

// Get is the public interface for Quixote. Get fetches fresh values from the source
// as needed and stores them to satisfy future requests, and it returns a value if available. 
func (c *cache) Get(key string) (string, bool) {
	c.mutex.RLock()
	cacheVal := c.index[key]
	c.mutex.RUnlock()
	now := time.Now()
	if cacheVal != nil {
		since := now.Sub(cacheVal.createTime)
		if since < c.softLimit {
			// Just return the cached value
			c.stats.CacheHitCount++
			return cacheVal.value,true
		} else if since < c.hardLimit {
			val,ok := c.refresh(key, now)
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
	return c.refresh(key, now)
}

// Dump prints out details of the cache state for debugging purposes
func (c *cache) Dump() {
	fmt.Printf("Index:\n")
	for k,v := range(c.index) {
		fmt.Printf("  %s : %s\n", k, v.value)
	}
	fmt.Printf("timelineLen: %d\n", c.timelineLen)
	fmt.Printf("timelineHead: %d\n", c.timelineHead)
	fmt.Printf("timelineTail: %d\n", c.timelineTail)
	fmt.Printf("timeline len %d\n", len(c.timeline))
	fmt.Printf("timeline %v\n", c.timeline)
}

// Stats returns current statistics about cache performance.
func (c *cache) Stats() Stats {
	return c.stats
}
