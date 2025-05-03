package relink

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"

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

	for file := range Walk(absSource) {
		grp.Go(func() error {
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
	}

	slog.Info("Walking target files")

	for file := range Walk(absTarget) {
		grp.Go(func() error {
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
	}

	slog.Info("Walking files completed")
	slog.Info("Waiting for all hashing jobs to finish")

	err = grp.Wait()
	if err != nil {
		slog.Error("failed to process files", "error", err)
		return err
	}

	slog.Info("Hashing completed")
	slog.Info("Comparing hashes")

	for k, targetHash := range hashedTargetFiles.Range {
		sourceHash, err := cc.Get(k)
		if err != nil {
			return fmt.Errorf("failed to get hash from cache: %w", err)
		}

		if slices.Equal(sourceHash, targetHash) {
			sourceFile := filepath.Join(absSource, k)
			targetFile := filepath.Join(absTarget, k)
			err = AtomicLink(sourceFile, targetFile)
			if err != nil {
				return fmt.Errorf("failed to create hardlink: %w", err)
			}
			slog.Info("file hashes match, hardlink created", "source", sourceFile, "target", targetFile)
		}
	}

	return nil
}
