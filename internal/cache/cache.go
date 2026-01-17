package cache

import (
	"bytes"
	"encoding/gob"
	"time"
)

type Value[T any] struct {
	data   T
	expiry time.Time
}

var _ gob.GobEncoder = (*Value[any])(nil)
var _ gob.GobDecoder = (*Value[any])(nil)

func NewValue[T any](data T, expiry time.Time) *Value[T] {
	return &Value[T]{data: data, expiry: expiry}
}

func (v *Value[T]) Data() T {
	return v.data
}

func (v *Value[T]) Expired(now time.Time) bool {
	return now.After(v.expiry)
}

// implements [gob.GobEncoder]
func (v *Value[T]) GobEncode() ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	encodable := struct {
		Data   T
		Expiry time.Time
	}{
		Data:   v.data,
		Expiry: v.expiry,
	}

	if err := enc.Encode(encodable); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// implements [gob.GobDecoder]
func (v *Value[T]) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	encodable := struct {
		Data   T
		Expiry time.Time
	}{}

	if err := dec.Decode(&encodable); err != nil {
		return nil
	}

	v.data = encodable.Data
	v.expiry = encodable.Expiry
	return nil
}

type Cache[T any] interface {
	Load(key string, now time.Time) (val *Value[T], ok bool, err error)
	Store(key string, val *Value[T]) error
}
