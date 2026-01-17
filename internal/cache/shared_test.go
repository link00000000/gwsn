package cache_test

import (
	"time"

	"github.com/link00000000/gwsn/internal/cache"
)

var now = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

type Struct struct {
	One string
	Two int32
}

func (s Struct) Equals(other Struct) bool {
	return s.One == other.One && s.Two == other.Two
}

type TestCache[T any] struct {
	LoadHandler  func(key string, now time.Time) (val *cache.Value[T], ok bool, err error)
	StoreHandler func(key string, value *cache.Value[T]) error
}

// implements [cache.Cache[T]]
func (c *TestCache[T]) Load(key string, now time.Time) (val *cache.Value[T], ok bool, err error) {
	if c.LoadHandler != nil {
		return c.LoadHandler(key, now)
	}

	return nil, false, nil
}

// implements [cache.Cache[T]]
func (c *TestCache[T]) Store(key string, value *cache.Value[T]) error {
	if c.StoreHandler != nil {
		return c.StoreHandler(key, value)
	}

	return nil
}
