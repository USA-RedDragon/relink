package relink_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/USA-RedDragon/relink/internal/config"
	"github.com/USA-RedDragon/relink/internal/relink"
)

func setupTestDirs(t *testing.T) (string, string, func()) {
	t.Helper()

	// Create temporary directories for source and target
	sourceDir, err := os.MkdirTemp(".", "relink-source-*")
	if err != nil {
		t.Fatalf("Failed to create source temp dir: %v", err)
	}

	targetDir, err := os.MkdirTemp(".", "relink-target-*")
	if err != nil {
		os.RemoveAll(sourceDir)
		t.Fatalf("Failed to create target temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(sourceDir)
		os.RemoveAll(targetDir)
	}

	return sourceDir, targetDir, cleanup
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("successfully links identical files", func(t *testing.T) {
		t.Parallel()
		sourceDir, targetDir, cleanup := setupTestDirs(t)
		defer cleanup()

		// Create identical files in both directories
		files := map[string]string{
			"file1.txt":               "testing",
			"subdir/file2.txt":        "another test",
			"subdir/nested/file3.txt": "more tests",
		}

		for file, contents := range files {
			// Create source file
			sourcePath := filepath.Join(sourceDir, file)
			err := os.MkdirAll(filepath.Dir(sourcePath), 0755)
			if err != nil {
				t.Fatalf("Failed to create source directory: %v", err)
			}
			err = os.WriteFile(sourcePath, []byte(contents), 0600)
			if err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			// Create identical target file
			targetPath := filepath.Join(targetDir, file)
			err = os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				t.Fatalf("Failed to create target directory: %v", err)
			}
			err = os.WriteFile(targetPath, []byte(contents), 0600)
			if err != nil {
				t.Fatalf("Failed to create target file: %v", err)
			}
		}

		// Run relink
		cfg := &config.Config{
			Source:     sourceDir,
			Target:     targetDir,
			HashJobs:   4,
			CacheType:  config.CacheTypeMemory,
			BufferSize: 4096,
		}
		err := relink.Run(cfg)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		for file, _ := range files {
			sourcePath := filepath.Join(sourceDir, file)
			targetPath := filepath.Join(targetDir, file)

			sourceInfo, err := os.Stat(sourcePath)
			if err != nil {
				t.Fatalf("Failed to stat source file: %v", err)
			}
			targetInfo, err := os.Stat(targetPath)
			if err != nil {
				t.Fatalf("Failed to stat target file: %v", err)
			}

			if !os.SameFile(sourceInfo, targetInfo) {
				t.Errorf("Files %s and %s are not hard linked", sourcePath, targetPath)
			}
		}
	})

	t.Run("skips files with different content", func(t *testing.T) {
		t.Parallel()
		sourceDir, targetDir, cleanup := setupTestDirs(t)
		defer cleanup()

		// Create files with different content
		sourcePath := filepath.Join(sourceDir, "file.txt")
		targetPath := filepath.Join(targetDir, "file.txt")

		err := os.WriteFile(sourcePath, []byte("source content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		err = os.WriteFile(targetPath, []byte("different content"), 0600)
		if err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}

		// Run relink
		cfg := &config.Config{
			Source:     sourceDir,
			Target:     targetDir,
			HashJobs:   4,
			BufferSize: 4096,
			CacheType:  config.CacheTypeMemory,
		}
		err = relink.Run(cfg)
		if err != nil {
			t.Fatalf("Run failed: %v", err)
		}

		sourceInfo, err := os.Stat(sourcePath)
		if err != nil {
			t.Fatalf("Failed to stat source file: %v", err)
		}
		targetInfo, err := os.Stat(targetPath)
		if err != nil {
			t.Fatalf("Failed to stat target file: %v", err)
		}

		if os.SameFile(sourceInfo, targetInfo) {
			t.Error("Files should not be hard linked due to different content")
		}
	})
}
