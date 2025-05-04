package relink

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/USA-RedDragon/relink/internal/config"
	"github.com/USA-RedDragon/relink/internal/relink/cache"
	"github.com/puzpuzpuz/xsync/v4"
	"golang.org/x/sync/errgroup"
)

func Run(cfg *config.Config) error {
	absSource, err := filepath.Abs(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for source: %w", err)
	}
	absTarget, err := filepath.Abs(cfg.Target)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for target: %w", err)
	}

	grp := errgroup.Group{}
	grp.SetLimit(cfg.HashJobs)

	var cc cache.Cache

	switch cfg.CacheType {
	case config.CacheTypeMemory:
		slog.Info("Using memory cache")
		cc = cache.NewMemoryCache()
	case config.CacheTypeSQLite:
		slog.Info("Using SQLite cache")
		cc, err = cache.NewSQLiteCache(cfg.CachePath)
		if err != nil {
			return fmt.Errorf("failed to create SQLite cache: %w", err)
		}
		defer cc.Close()
	default:
		return fmt.Errorf("invalid cache type: %s", cfg.CacheType)
	}

	hashedTargetFiles := xsync.NewMap[string, []byte]()

	slog.Info("Walking source files")

	totalFiles := 0
	var completedFiles atomic.Uint64

	for file := range Walk(absSource) {
		totalFiles++
		go func() {
			grp.Go(func() error {
				defer func() { completedFiles.Add(1) }()
				relative, err := filepath.Rel(absSource, file)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				// Check if the file is already in the cache
				exists, err := cc.Exists(relative)
				if err != nil {
					return fmt.Errorf("failed to check if file exists in cache: %w", err)
				}
				if exists {
					return nil
				}

				hash, err := HashFile(file, cfg.BufferSize)
				if err != nil {
					slog.Error("failed to hash file", "file", file, "error", err)
					return err
				}

				return cc.Put(relative, hash)
			})
		}()
	}

	for int(completedFiles.Load()) < totalFiles {
		slog.Info("Hashing source files", "completed", int(completedFiles.Load()), "total", totalFiles)
		time.Sleep(time.Second)
	}
	slog.Info("Hashing source files", "completed", int(completedFiles.Load()), "total", totalFiles)

	err = grp.Wait()
	if err != nil {
		slog.Error("failed to process files", "error", err)
		return err
	}

	slog.Info("Walking target files")

	totalFiles = 0
	completedFiles.Store(0)

	for file := range Walk(absTarget) {
		totalFiles++
		go func() {
			grp.Go(func() error {
				defer func() { completedFiles.Add(1) }()
				hash, err := HashFile(file, cfg.BufferSize)
				if err != nil {
					slog.Error("failed to hash file", "file", file, "error", err)
					return err
				}
				relative, err := filepath.Rel(absTarget, file)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}
				hashedTargetFiles.Store(relative, hash)
				return nil
			})
		}()
	}

	for int(completedFiles.Load()) < totalFiles {
		slog.Info("Hashing target files", "completed", int(completedFiles.Load()), "total", totalFiles)
		time.Sleep(time.Second)
	}
	slog.Info("Hashing target files", "completed", int(completedFiles.Load()), "total", totalFiles)

	err = grp.Wait()
	if err != nil {
		slog.Error("failed to process target files", "error", err)
		return err
	}

	slog.Info("Hashing completed")
	slog.Info("Comparing hashes")

	for k, targetHash := range hashedTargetFiles.Range {
		sourceRelative, err := cc.GetByHash(targetHash)
		if err != nil {
			return fmt.Errorf("failed to get hash from cache: %w", err)
		}
		if sourceRelative == "" {
			continue
		}

		sourceFile := filepath.Join(absSource, sourceRelative)
		targetFile := filepath.Join(absTarget, k)
		err = AtomicLink(sourceFile, targetFile)
		if err != nil {
			return fmt.Errorf("failed to create hardlink: %w", err)
		}
		slog.Info("file hashes match, hardlink created", "source", sourceFile, "target", targetFile)
	}

	return nil
}
