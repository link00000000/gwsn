package cache

import (
	"time"

	"golang.org/x/oauth2"
)

type TimeSource func() time.Time

type defaultTimeSource func() time.Time

func (defaultTimeSource) Now() time.Time {
	return time.Now()
}

type TokenSource struct {
	cache   Cache[*oauth2.Token]
	key     string
	src     oauth2.TokenSource
	timeSrc TimeSource
}

var _ oauth2.TokenSource = (*TokenSource)(nil)

func NewTokenSource(cache Cache[*oauth2.Token], key string, src oauth2.TokenSource) *TokenSource {
	return &TokenSource{cache: cache, key: key, src: src}
}

func (s *TokenSource) SetTimeSource(timeSrc TimeSource) {
	s.timeSrc = timeSrc
}

// implements [oauth2.TokenSource]
func (s *TokenSource) Token() (*oauth2.Token, error) {
	if val, ok, err := s.cache.Load(s.key, s.now()); ok && err == nil {
		return val.Data(), nil
	}
	// TODO: log error

	t, err := s.src.Token()
	if err == nil {
		s.cache.Store(s.key, NewValue(t, t.Expiry))
	}
	// TODO: log error

	return t, err
}

func (s *TokenSource) now() time.Time {
	if s.timeSrc != nil {
		return s.timeSrc()
	}

	return time.Now()
}
