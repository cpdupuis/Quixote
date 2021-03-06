/*

Quixote is an in-process cache supporting oldest-first cache eviction with a two-stage
invalidation policy. The first stage is a "soft" invalidation, in which the cached value
will be used only if no fresh value can be produced by the service dependency. The
second stage is "hard" invalidation, in which the cached value is purged from the cache.

Using two stages like this allows cached values to be kept arbitrarily fresh while the
service dependency is available. In the extreme case, the soft invalidation time can be
set to zero, so that Quixote will return cached values only in the case of the service
dependency being unavailable.

Creating a cache:

	query := func(context Context, key string) {...}
	softLimit := 15 * time.Second
	hardLimit := 30 * time.Minute
	maxCount := 65536
	cache := quixote.MakeQuixoteCache(query, softLimit, hardLimit, maxCount)

Using the cache:

	key := "{customerId=12345,orderId=67890}"
	now := time.Now()
	context := "something required for a particular service call"
	result,ok := cache.Get(context, key)
	if ok {
		// Do something with result.
	}

*/
package quixote
