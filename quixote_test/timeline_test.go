package quixote_test

import (
	"github.com/cpdupuis/Quixote/quixote"
	"testing"
	"time"
)

func Test(t *testing.T) {
	timeline := quixote.MakeExpiryTimeline(1, time.Second)
	var isCalled = false
	var isCorrect = false
	invalidator := func(key *string) {
		isCalled = true
		if *key == "foo" {
			isCorrect = true
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
	if !isCorrect {
		t.Errorf("Not correct")
	}
}
