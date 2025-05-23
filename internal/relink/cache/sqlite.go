package cache

import (
	"database/sql"
	"sync"

	_ "github.com/glebarez/go-sqlite"
)

type SQLiteCache struct {
	db         *sql.DB
	writeMutex sync.Mutex
}

func NewSQLiteCache(path string) (*SQLiteCache, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	err = migrate(db)
	if err != nil {
		return nil, err
	}
	return &SQLiteCache{
		db: db,
	}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS cache (
		key TEXT PRIMARY KEY,
		value BLOB
	);
	`)
	return err
}

func (s *SQLiteCache) Put(key string, value []byte) error {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	_, err := s.db.Exec("INSERT OR REPLACE INTO cache (key, value) VALUES (?, ?)", key, value)
	return err
}

func (s *SQLiteCache) GetByHash(hash []byte) (string, error) {
	var key string
	err := s.db.QueryRow("SELECT key FROM cache WHERE value = ? LIMIT 1", hash).Scan(&key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return key, nil
}

func (s *SQLiteCache) Exists(key string) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM cache WHERE key = ?)", key).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *SQLiteCache) Close() error {
	return s.db.Close()
}
