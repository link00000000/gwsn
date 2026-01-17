package cache_test

import (
	"testing"
	"time"

	"github.com/link00000000/gwsn/internal/cache"
)

func TestFS_Load_MissingKey(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[string](dir, nil)

	_, ok, err := c.Load("key", now)

	if ok {
		t.Errorf("expected not ok, but was")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestFS_Load_Expired(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[string](dir, nil)

	if err := c.Store("key", cache.NewValue("value", now.Add(-time.Hour))); err != nil {
		t.Error(err)
	}

	_, ok, err := c.Load("key", now)
	if ok {
		t.Errorf("expected not ok, but was")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestFS_Load_ExpiredFallbackToSrc(t *testing.T) {
	dir := t.TempDir()

	src := &TestCache[string]{
		LoadHandler: func(key string, now time.Time) (val *cache.Value[string], ok bool, err error) {
			return cache.NewValue("value", now.Add(time.Hour)), true, nil
		},
	}

	c := cache.NewFS(dir, src)

	val, ok, err := c.Load("key", now)
	if !ok {
		t.Errorf("expected ok, but was not")
	}

	if err != nil {
		t.Error(err)
	}

	if data := val.Data(); data != "value" {
		t.Errorf("value read from cache is not the same as what was written. expected %v, got %v", "value", data)
	}
}

func TestFS_Store_NewSimpleValue(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[string](dir, nil)

	if err := c.Store("key", cache.NewValue("value", now.Add(time.Hour))); err != nil {
		t.Error(err)
	}

	val, ok, err := c.Load("key", now)

	if !ok {
		t.Errorf("expected ok, but was not")
	}

	if err != nil {
		t.Error(err)
	}

	if data := val.Data(); data != "value" {
		t.Errorf("value read from cache is not the same as what was written. expected %v, got %v", "value", data)
	}
}

func TestFS_Store_NewStructValue(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[Struct](dir, nil)

	v := Struct{One: "value", Two: 100}
	if err := c.Store("key", cache.NewValue(v, now)); err != nil {
		t.Error(err)
	}

	val, ok, err := c.Load("key", now)

	if !ok {
		t.Errorf("expected ok, but was not")
	}

	if err != nil {
		t.Error(err)
	}

	if data := val.Data(); !data.Equals(v) {
		t.Errorf("value read from cache is not the same as what was written. expected %v, got %v", v, data)
	}
}

func TestFS_Store_Overwrite(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[string](dir, nil)

	if err := c.Store("key", cache.NewValue("value", now)); err != nil {
		t.Error(err)
	}

	if err := c.Store("key", cache.NewValue("new value", now)); err != nil {
		t.Error(err)
	}

	val, ok, err := c.Load("key", now)

	if !ok {
		t.Errorf("expected ok, but was not")
	}

	if err != nil {
		t.Error(err)
	}

	if data := val.Data(); data != "new value" {
		t.Errorf("value read from cache is not the same as what was last written. expected %v, got %v", "new value", data)
	}
}

func TestFS_Store_Nil(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewFS[string](dir, nil)

	if err := c.Store("key", cache.NewValue("value", now)); err != nil {
		t.Error(err)
	}

	if err := c.Store("key", nil); err != nil {
		t.Error(err)
	}

	_, ok, err := c.Load("key", now)

	if ok {
		t.Errorf("expected not ok, but was")
	}

	if err != nil {
		t.Error(err)
	}
}

func TestFS_Store_WritebackToSrc(t *testing.T) {
	dir := t.TempDir()

	var writebackVal *cache.Value[string] = nil
	src := &TestCache[string]{
		StoreHandler: func(key string, value *cache.Value[string]) error {
			writebackVal = value
			return nil
		},
	}

	c := cache.NewFS(dir, src)

	if err := c.Store("key", cache.NewValue("value", now.Add(time.Hour))); err != nil {
		t.Error(err)
	}

	if writebackVal == nil {
		t.Errorf("did not perform writeback to src")
	}

	if data := writebackVal.Data(); data != "value" {
		t.Errorf("value written to the src is not the same as what was last written. expected %v, got %v", "new value", data)
	}
}
