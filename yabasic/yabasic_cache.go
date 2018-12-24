package yabasic

import (
	"sync"
	"time"
)

type cacheItem struct {
	value string
	createTime time.Time
}

type YabasicCache interface {
	Get(string) (string,bool)
}

type cache struct {
	index map[string]*cacheItem // protected by mutex.
	queryFunc func(string) (string,bool) // returns the content and whether there was any content
	softLimit time.Duration // after the softLimit is passed, Get will query to get a fresh value, though it will return a cached value if an error occurs in the query
	hardLimit time.Duration // after the hardLimit, the cached value is removed.
	mutex sync.Mutex
}

func MakeYabasicCache(queryFunc func(string) (string,bool), softLimit time.Duration, hardLimit time.Duration) YabasicCache {
	return cache{index: make(map[string]*cacheItem), queryFunc: queryFunc, softLimit: softLimit, hardLimit: hardLimit}
}

func (c cache) refresh(key string, now time.Time) (string,bool) {
	val,ok := c.queryFunc(key)
	if ok {
		ci := &cacheItem{value:val, createTime: now}
		c.mutex.Lock()
		c.index[key] = ci
		c.mutex.Unlock()
		return val, true
	} else {
		// no cached result, sorry
		return "",false
	}
}


func (c cache) Get(key string) (string, bool) {
	c.mutex.Lock()
	cacheVal := c.index[key]
	c.mutex.Unlock()
	now := time.Now()
	if cacheVal != nil {
		since := now.Sub(cacheVal.createTime)
		if since < c.softLimit {
			// Just return the cached value
			return cacheVal.value,true
		} else if since < c.hardLimit {
			val,ok := c.refresh(key, now)
			if ok {
				// Hey, we refreshed successfully!
				return val, true
			} else {
				// We didn't refresh, but it's still OK.
				return cacheVal.value, true
			}
		}
	}
	// fallthrough: the cache didn't help us
	return c.refresh(key, now)
}
