package main

import (
	"fmt"
	"time"
	"math/rand"
	"github.com/cpdupuis/Quixote/quixote"
)

type simulatedService struct {
	failureRate float64 // Between 0 (no failures) and 1 (100% failure)
	callLatency time.Duration
	calls int
	failures int
}

func (serv *simulatedService) GetAnswer(question string) (string,bool) {
	serv.calls++
	time.Sleep(serv.callLatency)
	roll := rand.Float64()
	if roll <= serv.failureRate {
		serv.failures++
		return "",false
	} else {
		return fmt.Sprintf("%v", roll), true
	}
}

func main() {
	failures := 0
	noData := 0
	serv := &simulatedService{failureRate:0.1, callLatency: 1 * time.Millisecond}
	cacheFunc := func(ctx quixote.Context, key string) (string,bool) {
		if sv,ok := ctx.(*simulatedService); ok {
			return sv.GetAnswer(key)
		} else {
			return "",false
		}
	}
	cache := quixote.MakeQuixoteCache(cacheFunc, 8 * time.Second, 24 * time.Second, 4096)
	start := time.Now()
	for i:=0; i<1000000; i++ {
		keyNum := rand.Intn(4096)
		str,ok := cache.Get(serv, fmt.Sprintf("%d", keyNum))
		if !ok {
			failures++
		} else if len(str) == 0 {
			noData++
		}
		if i % 1000 == 0 {
			fmt.Printf("So far: %v", cache.GetAndResetStats().String())
		}
	}
	end := time.Now()
	diff := end.Sub(start)
	fmt.Printf("end user failures: %d\n", failures)
	fmt.Printf("end user noData: %d\n", noData)
	fmt.Printf("serv failures: %d\n", serv.failures)
	fmt.Printf("serv calls: %d\n", serv.calls)
	fmt.Printf("cache stats: %v", cache.Stats().String())
	fmt.Printf("Elapsed time: %+v", diff)
}
