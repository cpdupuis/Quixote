package common

type Invalidator interface {
	Invalidate(predicate func(string,string) bool)
}