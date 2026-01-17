package cache

import (
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"
)

// implements [Cache[T]]
type fs[T any] struct {
	dir string
	src Cache[T]
}

var _ Cache[any] = (*fs[any])(nil)

func NewFS[T any](dir string, src Cache[T]) Cache[T] {
	return &fs[T]{dir: dir, src: src}
}

func (c *fs[T]) Load(key string, now time.Time) (*Value[T], bool, error) {
	return c.load(key, now)
}

func (c *fs[T]) Store(key string, val *Value[T]) error {
	return c.store(key, val)
}

func (c *fs[T]) load(key string, now time.Time) (*Value[T], bool, error) {
	if val, ok, err := c.loadFile(key, now); ok {
		return val, ok, err
	}

	val, ok, err := c.loadSrc(key, now)
	if ok {
		// TODO: handle error
		// update cache on the way back up from reading source
		c.storeFile(key, val)
	}

	return val, ok, err
}

func (c *fs[T]) loadFile(key string, now time.Time) (*Value[T], bool, error) {
	f, err := os.Open(c.filename(key))
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	val := Value[T]{}
	if err := gob.NewDecoder(f).Decode(&val); err != nil {
		return nil, false, err
	}

	if val.Expired(now) {
		return nil, false, nil
	}

	return &val, true, nil
}

func (c *fs[T]) loadSrc(key string, now time.Time) (*Value[T], bool, error) {
	if c.src == nil {
		return nil, false, nil
	}

	return c.src.Load(key, now)
}

func (c *fs[T]) store(key string, val *Value[T]) error {
	err := c.storeSrc(key, val)
	if err != nil {
		return err
	}

	// only update cache if there was no error writing to source
	return c.storeFile(key, val)
}

func (c *fs[T]) storeFile(key string, val *Value[T]) error {
	if val == nil {
		err := os.Remove(c.filename(key))
		if err != nil && !os.IsNotExist(err) {
			return err
		}

		return nil
	}

	f, err := os.Create(c.filename(key))
	if err != nil {
		return err
	}
	defer f.Close()

	return gob.NewEncoder(f).Encode(val)
}

func (c *fs[T]) storeSrc(key string, val *Value[T]) error {
	if c.src == nil {
		return nil
	}

	return c.src.Store(key, val)
}

func (c *fs[T]) filename(key string) string {
	hash := md5.Sum([]byte(key))
	name := hex.EncodeToString(hash[:])
	return filepath.Join(c.dir, name)
}
