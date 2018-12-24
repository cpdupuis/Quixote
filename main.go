package main

import (
	"fmt"
	"github.com/cpdupuis/Quixote/common"
	"github.com/cpdupuis/Quixote/trivial"
)

var cache common.Cache
func main() {
	cache = trivial.MakeTrivialCache()
	cache.Put("schuyler sisters", "we hold these truths to be self-evident")
	str,ok := cache.Get("schuyler sisters")
	if ok {
		fmt.Printf("Cached: %s\n", str)
	}
	cache.ClearWhere(func (k string, v string) bool {
		return true
	})
	str,ok = cache.Get("schuler sisters")
	if ok {
		fmt.Printf("Oops, didn't clear")
	} else {
		fmt.Printf("Cleared!")
	}
}
