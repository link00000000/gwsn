package cache_test

import (
	"testing"
	"time"

	"github.com/link00000000/gwsn/internal/cache"
	"golang.org/x/oauth2"
)

func TestTokenSource_CacheMiss(t *testing.T) {
	c := cache.NewMemory[*oauth2.Token](nil)
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test", Expiry: now.Add(time.Hour)})
	src = cache.NewTokenSource(c, src)
	src.(*cache.TokenSource).SetTimeSource(func() time.Time { return now })

	tok, err := src.Token()
	if err != nil {
		t.Error(err)
	}

	if tok.AccessToken != "test" {
		t.Errorf("access token is not the same as what was provided. expected %v, got %v", "test", tok.AccessToken)
	}
}

func TestTokenSource_CacheHit(t *testing.T) {
	c := cache.NewMemory[*oauth2.Token](nil)

	first := true
	var s TestTokenSource = func() (*oauth2.Token, error) {
		if first {
			first = false
			return &oauth2.Token{AccessToken: "first", Expiry: now.Add(time.Hour)}, nil
		}

		return &oauth2.Token{AccessToken: "second", Expiry: now.Add(time.Hour)}, nil
	}

	src := cache.NewTokenSource(c, s)
	src.SetTimeSource(func() time.Time { return now })

	// first time populates the cache
	if _, err := src.Token(); err != nil {
		t.Error(err)
	}

	// second time should use the cached value
	tok, err := src.Token()
	if err != nil {
		t.Error(err)
	}

	if tok.AccessToken != "first" {
		t.Errorf("access token is not the same as what was provided when the cache was populated. expected %v, got %v", "first", tok.AccessToken)
	}
}

func TestTokenSource_CacheExpired(t *testing.T) {
	c := cache.NewMemory[*oauth2.Token](nil)

	first := true
	var s TestTokenSource = func() (*oauth2.Token, error) {
		if first {
			first = false
			return &oauth2.Token{AccessToken: "first", Expiry: now.Add(-time.Hour)}, nil
		}

		return &oauth2.Token{AccessToken: "second", Expiry: now.Add(time.Hour)}, nil
	}

	src := cache.NewTokenSource(c, s)
	src.SetTimeSource(func() time.Time { return now })

	// first time populates the cache
	if _, err := src.Token(); err != nil {
		t.Error(err)
	}

	// second time should read through because cache is expired
	tok, err := src.Token()
	if err != nil {
		t.Error(err)
	}

	if tok.AccessToken != "second" {
		t.Errorf("access token is not the same as what is provided when the cache is read through more than once. expected %v, got %v", "second", tok.AccessToken)
	}
}
