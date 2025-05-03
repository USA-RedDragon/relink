package cache

type Cache interface {
	Put(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	Exists(key string) (bool, error)
	Close() error
}
