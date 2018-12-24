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
	ybc := yabasic.MakeYabasicCache(query, time.Minute, time.Minute)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Minute)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Minute)
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
	ybc := yabasic.MakeYabasicCache(query, time.Nanosecond, time.Microsecond)
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

