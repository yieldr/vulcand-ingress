package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	NoExpiration      = cache.NoExpiration
	DefaultExpiration = cache.DefaultExpiration
)

type Cache interface {
	Set(key string, value interface{}, duration time.Duration)
	Get(key string) (interface{}, bool)
}

func NewCache(expiration time.Duration) Cache {
	return cache.New(expiration, expiration/10)
}

type mockCache struct{}

func (m *mockCache) Set(key string, value interface{}, duration time.Duration) {}

func (m *mockCache) Get(key string) (interface{}, bool) { return nil, false }

func NewMockCache() Cache {
	return new(mockCache)
}
