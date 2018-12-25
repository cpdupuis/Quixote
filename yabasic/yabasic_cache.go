package yabasic

import (
	"fmt"
	"sync"
	"time"
)

// It's just a basic cache.
// Eviction policy is oldest-first. Supports hard limit on age of items, as well as soft
// limit. When the soft limit is reached, the cache will attempt to refresh the item from the
// source, but it will return the cached item if the refresh fails.
type cacheItem struct {
	value string
	createTime time.Time
}

type YabasicCache interface {
	Get(string) (string,bool)
	Dump()
	Stats()
}

type cache struct {
	index map[string]*cacheItem // protected by mutex.
	timeline []*string // circular buffer
	timelineHead int // start of timeline
	timelineTail int // one past end of timeline
	timelineLen int // max number of elements
	queryFunc func(string) (string,bool) // returns the content and whether there was any content
	softLimit time.Duration // after the softLimit is passed, Get will query to get a fresh value, though it will return a cached value if an error occurs in the query
	hardLimit time.Duration // after the hardLimit, the cached value is removed.
	mutex sync.RWMutex
	count int
	unexpiredEvictionCount int
	cacheHitCount int
	cacheMissCount int
	cacheRescueCount int
}

func MakeYabasicCache(queryFunc func(string) (string,bool), softLimit time.Duration, hardLimit time.Duration, maxCount int) YabasicCache {
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
		timeline: make([]*string, maxCount),
	}
}


func (c *cache) freeHead() {
	delete(c.index, *c.timeline[c.timelineHead])
	c.timeline[c.timelineHead] = nil
	c.timelineHead = (c.timelineHead + 1) % c.timelineLen
	c.count--
}

// Must be called inside the write-lock
func (c *cache) makeSpace(now time.Time) {
	// first, clean up any entries past the hard limit
	for {
		// stop if the timeline is empty
		if c.count == 0 {
			return
		}
		key := c.timeline[c.timelineHead]
		ci := c.index[*key]
		if ci == nil {
			c.freeHead()
		} else if ci.createTime.Add(c.hardLimit).Before(now) {
			// item is past its limit. Let's free it up!
			c.freeHead()
		} else {
			break
		}
	}
	// If there is still no space, let's make some.
	if c.timelineTail == c.timelineHead {
		c.unexpiredEvictionCount++
		c.freeHead()
	}
}

func (c *cache) addToTimeline(key *string) {
	c.timeline[c.timelineTail] = key
	c.timelineTail = (c.timelineTail + 1) % c.timelineLen
	c.count++
}

func (c *cache) refresh(key string, now time.Time) (string,bool) {
	val,ok := c.queryFunc(key)
	if ok {
		ci := &cacheItem{value:val, createTime: now}
		c.mutex.Lock()
		c.makeSpace(now)
		c.index[key] = ci
		c.addToTimeline(&key)
		c.mutex.Unlock()
		return val, true
	} else {
		// no cached result, sorry
		return "",false
	}
}


func (c *cache) Get(key string) (string, bool) {
	c.mutex.RLock()
	cacheVal := c.index[key]
	c.mutex.RUnlock()
	now := time.Now()
	if cacheVal != nil {
		since := now.Sub(cacheVal.createTime)
		if since < c.softLimit {
			// Just return the cached value
			c.cacheHitCount++
			return cacheVal.value,true
		} else if since < c.hardLimit {
			val,ok := c.refresh(key, now)
			if ok {
				c.cacheMissCount++
				// Hey, we refreshed successfully!
				return val, true
			} else {
				c.cacheRescueCount++
				// We didn't refresh, but it's still OK.
				return cacheVal.value, true
			}
		}
	}
	c.cacheMissCount++
	// fallthrough: the cache didn't help us
	return c.refresh(key, now)
}

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

func (c *cache) Stats() {
	fmt.Printf("Cache count: %d\n", c.count)
	fmt.Printf("Unexpired eviction count: %d\n", c.unexpiredEvictionCount)
	calls := c.cacheHitCount + c.cacheMissCount + c.cacheRescueCount
	fmt.Printf("Cache hit count: %d Percent: %d\n", c.cacheHitCount, (c.cacheHitCount*100)/calls)
	fmt.Printf("Cache miss count: %d Percent: %d\n", c.cacheMissCount, (c.cacheMissCount*100)/calls)
	fmt.Printf("Cache rescue count: %d Percent: %d\n", c.cacheRescueCount, (c.cacheRescueCount*100)/calls)

}
