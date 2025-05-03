package relink

import (
	"fmt"
	"log/slog"
	"math"
	"path/filepath"
	"slices"

	"github.com/USA-RedDragon/relink/internal/config"
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
	if cfg.HashJobs > math.MaxInt {
		grp.SetLimit(math.MaxInt)
	} else {
		grp.SetLimit(int(cfg.HashJobs))
	}

	hashedSourceFiles := xsync.NewMap[string, []byte]()
	hashedTargetFiles := xsync.NewMap[string, []byte]()

	slog.Info("Walking source files")

	for file := range Walk(absSource) {
		grp.Go(func() error {
			hash, err := HashFile(file)
			if err != nil {
				slog.Error("failed to hash file", "file", file, "error", err)
				return err
			}
			relative, err := filepath.Rel(absSource, file)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %w", err)
			}
			hashedSourceFiles.Store(relative, hash)
			return nil
		})
	}

	slog.Info("Walking target files")

	for file := range Walk(absTarget) {
		grp.Go(func() error {
			hash, err := HashFile(file)
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
		if sourceHash, ok := hashedSourceFiles.Load(k); ok {
			if slices.Equal(sourceHash, targetHash) {
				sourceFile := filepath.Join(absSource, k)
				targetFile := filepath.Join(absTarget, k)
				err = AtomicLink(sourceFile, targetFile)
				if err != nil {
					return fmt.Errorf("failed to create hardlink: %w", err)
				}
				slog.Info("file hashes match, hardlink created", "source", sourceFile, "target", targetFile)
			}
			continue
		}
	}

	return nil
}
