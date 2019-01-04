# Quixote: Caching at windmills

[![GoDoc](https://godoc.org/github.com/cpdupuis/Quixote?status.svg)](https://godoc.org/github.com/cpdupuis/Quixote)

Sometimes you want a cache to reduce the number of calls you make to
a service dependency, and sometimes you want a cache to reduce the impact of
an unreliable service dependency. Sometimes you want both.

Quixote is an in-process cache supporting oldest-first cache eviction with a two-stage
invalidation policy. The first stage is a "soft" invalidation, in which the cached value
will be used only if no fresh value can be produced by the service dependency. The
second stage is "hard" invalidation, in which the cached value is purged from the cache.

Using two stages like this allows cached values to be kept arbitrarily fresh while the
service dependency is available. In the extreme case, the soft invalidation time can be
set to zero, so that Quixote will return cached values only in the case of the service
dependency being unavailable.

## Installation

`go get github.com/cpdupuis/Quixote/quixote`

## Usage

This example shows a program that prints the result of calling a service. The result is
cached, with a soft expiry after 10 seconds, and a hard expiry after 15 minutes. The
cache has a maximum size of 256 items.

```
func callService(params string) (string,bool) {
    // Call a service, return the result as a string, as well as a boolean ok value
}

func main() {
    cache := quixote.MakeQuixoteCache(callService, 10 * time.Second, 15 * time.Minute, 256)
    for {
        result,ok := cache.Get("testing 123")
        if ok {
            // Use result
        }
    }
}
```
