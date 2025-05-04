package cache

type Cache interface {
	Put(key string, value []byte) error
	GetByHash(hash []byte) (string, error)
	Exists(key string) (bool, error)
	Close() error
}
