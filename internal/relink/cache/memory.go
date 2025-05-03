package cache

import "github.com/puzpuzpuz/xsync/v4"

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

func (m *MemoryCache) Get(key string) ([]byte, error) {
	value, ok := m.cache.Load(key)
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (m *MemoryCache) Delete(key string) error {
	m.cache.Delete(key)
	return nil
}

func (m *MemoryCache) Exists(key string) (bool, error) {
	_, ok := m.cache.Load(key)
	return ok, nil
}

func (m *MemoryCache) Close() error {
	// No-op for memory cache
	return nil
}
