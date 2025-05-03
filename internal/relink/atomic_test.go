package relink_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/USA-RedDragon/relink/internal/relink"
)

func setupTestFiles(t *testing.T) (string, string, string) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "relink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create source file with some content
	sourcePath := filepath.Join(tempDir, "source.txt")
	sourceContent := []byte("test content")
	if err := os.WriteFile(sourcePath, sourceContent, 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create target path
	targetPath := filepath.Join(tempDir, "target.txt")

	return tempDir, sourcePath, targetPath
}

func cleanupTestFiles(t *testing.T, tempDir string) {
	if err := os.RemoveAll(tempDir); err != nil {
		t.Errorf("Failed to cleanup test directory: %v", err)
	}
}

func TestAtomicLink_Success(t *testing.T) {
	t.Parallel()
	tempDir, sourcePath, targetPath := setupTestFiles(t)
	defer cleanupTestFiles(t, tempDir)

	// Perform the atomic link
	err := relink.AtomicLink(sourcePath, targetPath)
	if err != nil {
		t.Errorf("AtomicLink failed: %v", err)
	}

	// Verify the target file exists and has the same content
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		t.Error("Target file was not created")
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("Failed to stat target file: %v", err)
	}

	if !os.SameFile(sourceInfo, targetInfo) {
		t.Error("Source and target files are not hard linked")
	}
}

func TestAtomicLink_SourceDoesNotExist(t *testing.T) {
	t.Parallel()
	tempDir, _, targetPath := setupTestFiles(t)
	defer cleanupTestFiles(t, tempDir)

	nonExistentSource := filepath.Join(tempDir, "nonexistent.txt")
	err := relink.AtomicLink(nonExistentSource, targetPath)
	if err == nil {
		t.Error("Expected error when source file does not exist")
	}
}

func TestAtomicLink_TargetDirectoryDoesNotExist(t *testing.T) {
	t.Parallel()
	tempDir, sourcePath, _ := setupTestFiles(t)
	defer cleanupTestFiles(t, tempDir)

	nonExistentDir := filepath.Join(tempDir, "nonexistent", "target.txt")
	err := relink.AtomicLink(sourcePath, nonExistentDir)
	if err == nil {
		t.Error("Expected error when target directory does not exist")
	}
}
