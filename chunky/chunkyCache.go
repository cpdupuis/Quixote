package chunky

import (
	"github.com/cpdupuis/Quixote/common"
	"time"
)



// The idea of a chunky cache is that every item gets assigned to a "chunk", which
// can hold a large number of items, all of which will be purged when the
// chunk's lifetime ends. Each item has an expiry time, and each chunk has an expiry
// time.
type ChunkyCache struct {
	ChunkDuration time.Duration
	NumChunks int

}

func MakeChunkyCache() common.Cache {
	return ChunkyCache{}
}

func (cc ChunkyCache) Put(key string, value string) bool {
	return true
}

func (cc ChunkyCache) Get(key string) (string,bool) {
	return "",false
}

func (cc ChunkyCache) ClearWhere(predicate func(string,string)bool) {
}
