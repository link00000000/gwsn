package cache

import (
	"sync"
	"time"
)

// implements [Cache[T]]
type memory[T any] struct {
	mu  sync.Mutex
	m   map[string]*Value[T]
	src Cache[T]
}

var _ Cache[any] = (*memory[any])(nil)

func NewMemory[T any](src Cache[T]) Cache[T] {
	return &memory[T]{
		m:   make(map[string]*Value[T]),
		src: src,
	}
}

func (c *memory[T]) Load(key string, now time.Time) (*Value[T], bool, error) {
	return c.load(key, now)
}

func (c *memory[T]) Store(key string, val *Value[T]) error {
	return c.store(key, val)
}

func (c *memory[T]) load(key string, now time.Time) (*Value[T], bool, error) {
	if val, ok, err := c.loadMemory(key, now); ok {
		return val, ok, err
	}

	val, ok, err := c.loadSrc(key, now)
	if ok {
		// TODO: handle error
		// update cache on the way back up from reading source
		c.storeMemory(key, val)
	}

	return val, ok, err
}

func (c *memory[T]) loadMemory(key string, now time.Time) (*Value[T], bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	val, ok := c.m[key]
	if !ok || val.Expired(now) {
		return nil, false, nil
	}

	return val, ok, nil
}

func (c *memory[T]) loadSrc(key string, now time.Time) (*Value[T], bool, error) {
	if c.src == nil {
		return nil, false, nil
	}

	return c.src.Load(key, now)
}

func (c *memory[T]) store(key string, val *Value[T]) error {
	err := c.storeSrc(key, val)
	if err != nil {
		return err
	}

	// only update cache if there was no error writing to source
	return c.storeMemory(key, val)
}

func (c *memory[T]) storeMemory(key string, val *Value[T]) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if val == nil {
		delete(c.m, key)
		return nil
	}

	c.m[key] = val
	return nil
}

func (c *memory[T]) storeSrc(key string, val *Value[T]) error {
	if c.src == nil {
		return nil
	}

	return c.src.Store(key, val)
}
