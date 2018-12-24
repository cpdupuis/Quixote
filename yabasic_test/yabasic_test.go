package yabasic_test

import (
	"github.com/cpdupuis/Quixote/yabasic"
	"testing"
	"time"
)

func TestInitialRefresh(t *testing.T) {
	value := "value"
	ok := true
	query := func (key string) (string,bool) {
		return value, ok
	}
	ybc := yabasic.MakeYabasicCache(query, time.Minute, time.Minute, 2)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Minute, 2)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Minute, 2)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Microsecond, 2)
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
	ybc := yabasic.MakeYabasicCache(query, time.Minute, time.Minute, 2)
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
	ybc := yabasic.MakeYabasicCache(query, time.Minute, time.Minute, 2)
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