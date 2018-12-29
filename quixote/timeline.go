package quixote

import (
	"time"
)


// A train of expiryItems. When you add/refresh a cached item, you put it in the
// front of the train. Then the front of the train moves on as time progresses.
// Every time the front of the train progresses (once it is full), the back becomes
// the front, and all of its entries are purged.

type expiryItem struct {
	itemSet map[*string]bool
}

type addrTime int64

type expiryTimeline struct {
	expiryItems []expiryItem // circular buffer. Going forward in buffer order is going back in time.
	newestItem int
	newestTime addrTime
	count int // number of items in the circular buffer, constant value
	timeResolutionMillis int64 // at what resolution (in time) does the timeline chunk time?
}


func (et *expiryTimeline) addressableTime(t time.Time) addrTime {
	millis := t.UnixNano() / 1000000
	return addrTime(millis / et.timeResolutionMillis)
}

func (et *expiryTimeline) findExpiryItem(at addrTime) *expiryItem {
	offset := et.newestTime - at
	if offset < 0 {
		return nil
	} else if offset >= addrTime(et.count) {
		return nil
	} else {
		// OK, it's in our timeline somewhere.
		return &et.expiryItems[(et.newestItem + int(offset)) % et.count]
	}
}

func (et *expiryTimeline) ReplaceItem(key *string, timeOld time.Time, timeNew time.Time) bool {
	addrTimeOld := et.addressableTime(timeOld)
	addrTimeNew := et.addressableTime(timeNew)
	if expiryItemOld := et.findExpiryItem(addrTimeOld); expiryItemOld != nil {
		delete(expiryItemOld.itemSet, key)
	}
	if expiryItemNew := et.findExpiryItem(addrTimeNew); expiryItemNew != nil {
		expiryItemNew.itemSet[key] = true
		return true
	} else {
		// Sorry, we can't store a value that we don't know how to expire.
		return false
	}
}

// Move the head of the buffer up to now.
func (et *expiryTimeline) ExpireItems(now time.Time) {
	addrNow := et.addressableTime(now)
	for addrNow < et.newestTime {
		et.newestTime++
		et.newestItem = (et.newestItem + 1) % et.count
		// This is the new head. Clear out old items by replacing the entire map.
		et.expiryItems[et.newestItem].itemSet = make(map[*string]bool)
	}
	// All set!
}

