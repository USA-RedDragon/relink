package relink_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/USA-RedDragon/relink/internal/relink"
)

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "walk-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create test files and directories
	files := []string{
		"file1.txt",
		"file2.txt",
		"subdir/file3.txt",
		"subdir/nested/file4.txt",
		"empty_dir",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if file == "empty_dir" {
			err = os.MkdirAll(path, 0755)
		} else {
			err = os.MkdirAll(filepath.Dir(path), 0755)
			if err == nil {
				err = os.WriteFile(path, []byte("test"), 0644)
			}
		}
		if err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestWalk(t *testing.T) {
	t.Parallel()

	t.Run("walks all files in directory", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := setupTestDir(t)
		defer cleanup()

		expectedFiles := map[string]bool{
			filepath.Join(tmpDir, "file1.txt"):               true,
			filepath.Join(tmpDir, "file2.txt"):               true,
			filepath.Join(tmpDir, "subdir/file3.txt"):        true,
			filepath.Join(tmpDir, "subdir/nested/file4.txt"): true,
		}

		foundFiles := make(map[string]bool)
		for path, err := range relink.Walk(tmpDir) {
			if err != nil {
				t.Errorf("Unexpected error while walking: %v", err)
				continue
			}
			foundFiles[path] = true
		}

		if len(foundFiles) != len(expectedFiles) {
			t.Errorf("Expected %d files, got %d", len(expectedFiles), len(foundFiles))
		}

		for path := range expectedFiles {
			if !foundFiles[path] {
				t.Errorf("Expected file not found: %s", path)
			}
		}
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		t.Parallel()
		nonExistentDir := filepath.Join(t.TempDir(), "does-not-exist")

		foundAny := false
		for _, err := range relink.Walk(nonExistentDir) {
			foundAny = true
			if err == nil {
				t.Error("Expected error for non-existent directory, got nil")
			}
		}

		if foundAny {
			t.Error("Expected no files to be found for non-existent directory")
		}
	})

	t.Run("skips root directory", func(t *testing.T) {
		t.Parallel()
		tmpDir, cleanup := setupTestDir(t)
		defer cleanup()

		// Create a file with the same name as the root directory
		rootFile := tmpDir + ".txt"
		err := os.WriteFile(rootFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create root file: %v", err)
		}
		defer os.Remove(rootFile)

		foundRoot := false
		for path, err := range relink.Walk(tmpDir) {
			if err != nil {
				t.Errorf("Unexpected error while walking: %v", err)
				continue
			}
			if path == tmpDir {
				foundRoot = true
			}
		}

		if foundRoot {
			t.Error("Root directory should be skipped")
		}
	})
}
