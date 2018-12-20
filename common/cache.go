package common

// At this level, a cache is a mapping from string to string.
// Any serialization/deserialization lives at a higher level.
type Cache interface {
	Put(key string, value string) bool
	Get(key string) (string,bool)
	ClearWhere(func(string,string)bool)
}

