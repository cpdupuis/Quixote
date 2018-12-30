package quixote

import (
	"time"
)


// A train of expiryItems. When you add/refresh a cached item, you put it in the
// front of the train. Then the front of the train moves on as time progresses.
// Every time the front of the train progresses (once it is full), the back becomes
// the front, and all of its entries are purged.

type expiryItem struct {
	itemSet map[string]bool
}

type addrTime int64

type ExpiryTimeline struct {
	expiryItems []expiryItem // circular buffer. Going forward in buffer order is going back in time.
	newestItem int
	newestTime addrTime
	count int // number of items in the circular buffer, constant value
	timeResolutionNanos int64 // at what resolution (in time) does the timeline chunk time?
}


func calcAddrTime(t time.Time, resolutionNanos int64) addrTime {
	return addrTime(t.UnixNano() / resolutionNanos)
}

func (et *ExpiryTimeline) addressableTime(t time.Time) addrTime {
	return calcAddrTime(t, et.timeResolutionNanos)
}

func (et *ExpiryTimeline) findExpiryItem(at addrTime) *expiryItem {
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


func (et *ExpiryTimeline) mutator(key string, timeOld time.Time, timeNew time.Time, doAdd bool, doDelete bool) bool {
	if doDelete {
		addrTimeOld := et.addressableTime(timeOld)
		if expiryItemOld := et.findExpiryItem(addrTimeOld); expiryItemOld != nil {
			delete(expiryItemOld.itemSet, key)
		}
	}
	if doAdd {
		addrTimeNew := et.addressableTime(timeNew)
		if expiryItemNew := et.findExpiryItem(addrTimeNew); expiryItemNew != nil {
			expiryItemNew.itemSet[key] = true
		} else {
			// Sorry, we can't store a value that we don't know how to expire.
			return false
		}
	}
	return true

}

func (et *ExpiryTimeline) AddItem(key string, timeNew time.Time) bool {
	return et.mutator(key, timeNew, timeNew, true, false)
}

func (et *ExpiryTimeline) ReplaceItem(key string, timeOld time.Time, timeNew time.Time) bool {
	return et.mutator(key, timeOld, timeNew, true, true)
}

func (et *ExpiryTimeline) DeleteItem(key string, timeOld time.Time) {
	et.mutator(key, timeOld, timeOld, false, true)
}

// Move the head of the buffer up to now.
func (et *ExpiryTimeline) ExpireItems(now time.Time, invalidator func(string)) {
	addrNow := et.addressableTime(now)
	for addrNow > et.newestTime {
		et.newestTime++
		et.newestItem = (et.newestItem + 1) % et.count
		// This is the new head. Clear out old items.
		itemSet := et.expiryItems[et.newestItem].itemSet
		for k := range itemSet {
			invalidator(k)
			delete(itemSet, k)
		}
	}
	// All set!
}

func MakeExpiryTimeline(count int, expiryLifetime time.Duration) *ExpiryTimeline {
	resolutionNanos := expiryLifetime.Nanoseconds() / int64(count)
	now := time.Now()
	addrNow := calcAddrTime(now, resolutionNanos)
	et := &ExpiryTimeline{
		count: count,
		timeResolutionNanos: resolutionNanos,
		newestItem: 0,
		newestTime: addrNow,
		expiryItems: make([]expiryItem, count),
	}
	for i:=0; i<count; i++ {
		et.expiryItems[i].itemSet = make(map[string]bool)
	}
	return et
}

