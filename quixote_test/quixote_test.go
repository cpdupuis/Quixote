package quixote_test

import (
	"fmt"
	"github.com/cpdupuis/Quixote/quixote"
	"testing"
	"time"
)

func TestInitialRefresh(t *testing.T) {
	value := "value"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}
	ybc := quixote.MakeQuixoteCache(query, time.Minute, time.Minute, 2)
	str,ok := ybc.Get("anything")
	if !ok {
		t.Fail()
	}
	if str != "value" {
		t.Fail()
	}
}

func TestSuccessfulSoftRefresh(t *testing.T) {
	value := "value"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}
	ybc := quixote.MakeQuixoteCache(query, time.Nanosecond, time.Minute, 2)
	str,ok := ybc.Get("anything")
	if !ok {
		t.Errorf("not ok")
	}
	if str != "value" {
		t.Errorf("not value")
	}
	time.Sleep(2 * time.Nanosecond)
	value = "new value"
	str,ok = ybc.Get("anything")
	if str != "new value" {
		t.Errorf("not new value %s", str)
	}
}

func TestFailedSoftRefresh(t *testing.T) {
	value := "value"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}
	ybc := quixote.MakeQuixoteCache(query, time.Nanosecond, time.Minute, 2)
	str,ok := ybc.Get("anything")
	if !ok {
		t.Errorf("not ok")
	}
	if str != "value" {
		t.Errorf("not value")
	}
	time.Sleep(2 * time.Nanosecond)
	value = ""
	ok = false
	str,ok = ybc.Get("anything")
	if str != "value" {
		t.Errorf("not value %s", str)
	}
}

func TestFailedHardRefresh(t *testing.T) {
	value := "value"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}
	ybc := quixote.MakeQuixoteCache(query, time.Nanosecond, time.Microsecond, 2)
	str,ok := ybc.Get("anything")
	if !ok {
		t.Errorf("not ok")
	}
	if str != "value" {
		t.Errorf("not value")
	}
	time.Sleep(2 * time.Microsecond)
	value = ""
	ok = false
	str,ok = ybc.Get("anything")
	if ok {
		t.Fail()
	}
	if str != "" {
		t.Errorf("not value %s", str)
	}

}

func TestCacheCapacity(t *testing.T) {
	value := "un"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}	
	ybc := quixote.MakeQuixoteCache(query, time.Minute, time.Minute, 2)
	str,ok := ybc.Get("one")
	if !ok || str != "un" {
		t.Errorf("Expected un, got %s", str)
	}
	value = "deux"
	str,ok = ybc.Get("two")
	if !ok || str != "deux" {
		t.Errorf("expected deux, got %s", str)
	}
	value = "trois"
	str,ok = ybc.Get("one")
	if !ok || str != "un" {
		t.Errorf("expected un, got: %s", str)
	}
}

func TestCacheEviction(t *testing.T) {
	value := "un"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}	
	ybc := quixote.MakeQuixoteCache(query, time.Minute, time.Minute, 2)
	_,_ = ybc.Get("one")
	value = "deux"
	_,_ = ybc.Get("two")
	value = "trois"
	_,_ = ybc.Get("three")

	value = "new"
	str,_ := ybc.Get("three")
	if str != "trois" {
		t.Fail()
	}
	str,_ = ybc.Get("two")
	if str != "deux" {
		t.Fail()
	}
	str,_ = ybc.Get("one")
	if str != "new" {
		// old value should have been evicted
		t.Errorf("Wrong value: %s", str)
	}
}

func TestPerfNoOverflow(t *testing.T) {
	query := func (key string) (string,bool) {
		return key, true
	}
	ybc := quixote.MakeQuixoteCache(query, 	5* time.Millisecond, 2 * time.Minute, 1024)

	for j:=0; j<4096; j++ {
		for i:=0; i<1024; i++ {
			key := fmt.Sprintf("key:%d", i)
			res,ok := ybc.Get(key)
			if !ok {
				t.Errorf("Not OK!")
				return
			}
			if res != key {
				t.Errorf("Wrong res, expected %s, got %s", key, res)
			}
		}
	}
	stats := ybc.Stats()
	fmt.Printf("Stats: %v\n", stats)
	if stats.CacheHitCount < 94 {
		t.Errorf("Low cache hit count. Expected 94, got %d", stats.CacheHitCount)
	}
	if stats.UnexpiredEvictionCount > 3000000 {
		t.Errorf("Expected less eviction: %d", stats.UnexpiredEvictionCount)
	}
}
