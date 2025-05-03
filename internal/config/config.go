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

type Config struct {
	LogLevel LogLevel `name:"log-level" description:"Logging level for the application. One of debug, info, warn, or error" default:"info"`
	Source   string   `name:"source" description:"Source directory to read the files from"`
	Target   string   `name:"target" description:"Target directory to write the relinked files to"`
	HashJobs uint     `name:"hash-jobs" description:"Number of jobs to use for hashing files" default:"4"`
}

var (
	ErrBadLogLevel         = errors.New("invalid log level provided")
	ErrNoSource            = errors.New("no source directory provided")
	ErrNoTarget            = errors.New("no target directory provided")
	ErrSourceNotFound      = errors.New("source directory not found")
	ErrSourceAndTargetSame = errors.New("source and target directories are the same")
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

	return nil
}
