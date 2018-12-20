package trivial

import (
	"github.com/cpdupuis/Quixote/common"
)

type trivialCache struct {
	cache map[string]string
}

func MakeTrivialCache() common.Cache {
	return trivialCache{cache: make(map[string]string)}
}

func (tc trivialCache) Put(key string, value string) bool {
	tc.cache[key] = value
	return true
}

func (tc trivialCache) Get(key string) (string,bool) {
	val, ok := tc.cache[key]
	return val,ok
}

func (tc trivialCache) ClearWhere(predicate func(string,string)bool) {
	for k,v := range tc.cache {
		if predicate(k,v) {
			delete(tc.cache,k)
		}
	}
}
