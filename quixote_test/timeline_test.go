package quixote_test

import (
	"github.com/cpdupuis/Quixote/quixote"
	"testing"
	"time"
)

func TestAddOneExpireOne(t *testing.T) {
	timeline := quixote.MakeExpiryTimeline(1, time.Second)
	var isCalled = false
	var isExpired = false
	invalidator := func(key *string) {
		isCalled = true
		if *key == "foo" {
			isExpired = true
		}
	}
	key := "foo"
	now := time.Now()
	ok := timeline.AddItem(&key, now)
	if !ok {
		t.Errorf("Failed to add")
	}
	later := now.Add(time.Second)
	timeline.ExpireItems(later, invalidator)
	if !isCalled {
		t.Errorf("Not called")
	}
	if !isExpired {
		t.Errorf("Not correct")
	}
}

func TestAddOneDeleteOne(t *testing.T) {
	timeline := quixote.MakeExpiryTimeline(1, time.Second)
	var isExpired = false
	invalidator := func(key *string) {
		if *key == "foo" {
			isExpired = true
		}
	}
	key := "foo"
	now := time.Now()
	timeline.AddItem(&key, now)
	timeline.DeleteItem(&key, now)
	later := now.Add(time.Second)
	timeline.ExpireItems(later, invalidator)
	if isExpired {
		t.Errorf("Item was already supposed to be deleted")
	}
}
