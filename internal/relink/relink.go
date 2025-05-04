package relink

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/USA-RedDragon/relink/internal/config"
	"github.com/USA-RedDragon/relink/internal/relink/cache"
	"github.com/USA-RedDragon/relink/internal/utils"
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

	slog.Info("Walking source files")

	totalFiles := 0
	totalSize := uint64(0)
	var completedFiles atomic.Uint64
	var completedSize atomic.Uint64

	for file := range Walk(absSource) {
		totalFiles++
		fileSize := file.Info.Size()
		totalSize += uint64(fileSize)
		go func() {
			grp.Go(func() error {
				defer func() { completedFiles.Add(1) }()
				relative, err := filepath.Rel(absSource, file.Path)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				// Check if the file is already in the cache
				exists, err := cc.Exists(relative)
				if err != nil {
					return fmt.Errorf("failed to check if file exists in cache: %w", err)
				}
				if exists {
					completedSize.Add(uint64(fileSize))
					return nil
				}

				readBytesChan := make(chan uint64)
				wg := errgroup.Group{}
				wg.Go(func() error {
					hash, err := HashFile(file.Path, cfg.BufferSize, readBytesChan)
					if err != nil {
						slog.Error("failed to hash file", "file", file, "error", err)
						return err
					}

					close(readBytesChan)

					return cc.Put(relative, hash)
				})

				for {
					select {
					case readBytes, ok := <-readBytesChan:
						if !ok {
							return wg.Wait()
						}
						completedSize.Add(readBytes)
					}
				}
			})
		}()
	}

	for int(completedFiles.Load()) < totalFiles {
		slog.Info("Hashing source files", "completed", int(completedFiles.Load()), "total", totalFiles, "completedSize", utils.HumanReadableSize(completedSize.Load()), "totalSize", utils.HumanReadableSize(totalSize))
		time.Sleep(time.Second)
	}
	slog.Info("Hashing source files", "completed", int(completedFiles.Load()), "total", totalFiles, "completedSize", utils.HumanReadableSize(completedSize.Load()), "totalSize", utils.HumanReadableSize(totalSize))

	err = grp.Wait()
	if err != nil {
		slog.Error("failed to process files", "error", err)
		return err
	}

	slog.Info("Walking target files")

	totalFiles = 0
	totalSize = 0
	completedFiles.Store(0)
	completedSize.Store(0)

	for file := range Walk(absTarget) {
		totalFiles++
		if file.Info.Mode()&os.ModeSymlink != 0 {
			slog.Debug("skipping symlink", "file", file)
			continue
		}
		fileSize := file.Info.Size()
		totalSize += uint64(fileSize)
		go func() {
			grp.Go(func() error {
				defer func() { completedFiles.Add(1) }()
				readBytesChan := make(chan uint64)
				wg := errgroup.Group{}
				wg.Go(func() error {
					hash, err := HashFile(file.Path, cfg.BufferSize, readBytesChan)
					if err != nil {
						slog.Error("failed to hash file", "file", file, "error", err)
						return err
					}
					close(readBytesChan)

					sourceRelative, err := cc.GetByHash(hash)
					if err != nil {
						return fmt.Errorf("failed to get hash from cache: %w", err)
					}
					if sourceRelative == "" {
						return nil
					}

					sourceFile := filepath.Join(absSource, sourceRelative)
					err = AtomicLink(sourceFile, file.Path)
					if err != nil {
						return fmt.Errorf("failed to create hardlink: %w", err)
					}
					slog.Info("file hashes match, hardlink created", "source", sourceFile, "target", file.Path)

					return nil
				})

				for {
					select {
					case readBytes, ok := <-readBytesChan:
						if !ok {
							return wg.Wait()
						}
						completedSize.Add(readBytes)
					}
				}
			})
		}()
	}

	for int(completedFiles.Load()) < totalFiles {
		slog.Info("Hashing target files", "completed", int(completedFiles.Load()), "total", totalFiles, "completedSize", utils.HumanReadableSize(completedSize.Load()), "totalSize", utils.HumanReadableSize(totalSize))
		time.Sleep(time.Second)
	}
	slog.Info("Hashing target files", "completed", int(completedFiles.Load()), "total", totalFiles, "completedSize", utils.HumanReadableSize(completedSize.Load()), "totalSize", utils.HumanReadableSize(totalSize))

	err = grp.Wait()
	if err != nil {
		slog.Error("failed to process target files", "error", err)
		return err
	}

	slog.Info("Hashing and hardlinking completed")

	return nil
}
