package config

import (
	"errors"
	"os"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type CacheType string

const (
	CacheTypeMemory CacheType = "memory"
	CacheTypeSQLite CacheType = "sqlite"
)

type Config struct {
	LogLevel   LogLevel  `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Source     string    `name:"source" description:"Source directory to read the files from"`
	Target     string    `name:"target" description:"Target directory to write the relinked files to"`
	HashJobs   int       `name:"hash-jobs" description:"Number of jobs to use for hashing files" default:"4"`
	BufferSize int       `name:"buffer-size" description:"Buffer size for file checksum operations in bytes" default:"4096"`
	CacheType  CacheType `name:"cache-type" description:"Cache type to use for storing file hashes. One of memory or sqlite" default:"memory"`
	CachePath  string    `name:"cache-path" description:"Path to the SQLite database file for caching. Only used if cache-type is sqlite" default:":memory:"`
}

var (
	ErrBadLogLevel            = errors.New("invalid log level provided")
	ErrNoSource               = errors.New("no source directory provided")
	ErrNoTarget               = errors.New("no target directory provided")
	ErrSourceNotFound         = errors.New("source directory not found")
	ErrSourceAndTargetSame    = errors.New("source and target directories are the same")
	ErrZeroBufferSize         = errors.New("buffer size must be greater than 0 bytes")
	ErrZeroHashJobs           = errors.New("hash jobs must be greater than 0")
	ErrInvalidCacheType       = errors.New("invalid cache type provided")
	ErrCachePathWithoutSQLite = errors.New("cache path cannot be set without cache type being sqlite")
)

func (c Config) Validate() error {
	if c.LogLevel != LogLevelDebug &&
		c.LogLevel != LogLevelInfo &&
		c.LogLevel != LogLevelWarn &&
		c.LogLevel != LogLevelError {
		return ErrBadLogLevel
	}

	if c.Source == "" {
		return ErrNoSource
	}

	if c.Target == "" {
		return ErrNoTarget
	}

	if c.Source == c.Target {
		return ErrSourceAndTargetSame
	}

	if _, err := os.Stat(c.Source); errors.Is(err, os.ErrNotExist) {
		return ErrSourceNotFound
	}

	if c.BufferSize <= 0 {
		return ErrZeroBufferSize
	}

	if c.HashJobs <= 0 {
		return ErrZeroHashJobs
	}

	if c.CacheType != CacheTypeMemory &&
		c.CacheType != CacheTypeSQLite {
		return ErrInvalidCacheType
	}

	if c.CacheType == CacheTypeSQLite && c.CachePath == "" {
		return ErrCachePathWithoutSQLite
	}

	return nil
}
