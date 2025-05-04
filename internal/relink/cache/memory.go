package cache

import (
	"slices"

	"github.com/puzpuzpuz/xsync/v4"
)

type MemoryCache struct {
	cache *xsync.Map[string, []byte]
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: xsync.NewMap[string, []byte](),
	}
}

func (m *MemoryCache) Put(key string, value []byte) error {
	m.cache.Store(key, value)
	return nil
}

func (m *MemoryCache) GetByHash(hash []byte) (string, error) {
	var foundKey string
	m.cache.Range(func(key string, value []byte) bool {
		if slices.Equal(value, hash) {
			foundKey = key
			return false // Stop iteration
		}
		return true // Continue iteration
	})
	if foundKey == "" {
		return "", nil // Not found
	}
	return foundKey, nil
}

func (m *MemoryCache) Exists(key string) (bool, error) {
	_, ok := m.cache.Load(key)
	return ok, nil
}

func (m *MemoryCache) Close() error {
	// No-op for memory cache
	return nil
}
