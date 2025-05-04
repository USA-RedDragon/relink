package relink_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/USA-RedDragon/relink/internal/relink"
	"golang.org/x/crypto/blake2b"
)

const testBuffer = 4096

func TestHashFileRegular(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for this test
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file with known content
	content := []byte("Hello, World!")
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Calculate expected hash
	expectedHash, err := blake2b.New512(nil)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}
	if _, err := expectedHash.Write(content); err != nil {
		t.Fatalf("Failed to write to hash: %v", err)
	}
	expectedSum := expectedHash.Sum(nil)

	// Test HashFile
	actualHash, err := relink.HashFile(filePath, testBuffer, nil)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}
	if len(actualHash) != len(expectedSum) {
		t.Errorf("Hash length mismatch: got %d, want %d", len(actualHash), len(expectedSum))
	}
	for i := range actualHash {
		if actualHash[i] != expectedSum[i] {
			t.Errorf("Hash mismatch at byte %d: got %x, want %x", i, actualHash[i], expectedSum[i])
		}
	}
}

func TestHashFileEmpty(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(filePath, []byte{}, 0600); err != nil {
		t.Fatalf("Failed to write empty file: %v", err)
	}

	expectedHash, err := blake2b.New512(nil)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}
	expectedSum := expectedHash.Sum(nil)

	actualHash, err := relink.HashFile(filePath, testBuffer, nil)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}
	if len(actualHash) != len(expectedSum) {
		t.Errorf("Hash length mismatch: got %d, want %d", len(actualHash), len(expectedSum))
	}
	for i := range actualHash {
		if actualHash[i] != expectedSum[i] {
			t.Errorf("Hash mismatch at byte %d: got %x, want %x", i, actualHash[i], expectedSum[i])
		}
	}
}

func TestHashFileLarge(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "large.txt")
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(filePath, largeContent, 0600); err != nil {
		t.Fatalf("Failed to write large file: %v", err)
	}

	expectedHash, err := blake2b.New512(nil)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}
	if _, err := expectedHash.Write(largeContent); err != nil {
		t.Fatalf("Failed to write to hash: %v", err)
	}
	expectedSum := expectedHash.Sum(nil)

	actualHash, err := relink.HashFile(filePath, testBuffer, nil)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}
	if len(actualHash) != len(expectedSum) {
		t.Errorf("Hash length mismatch: got %d, want %d", len(actualHash), len(expectedSum))
	}
	for i := range actualHash {
		if actualHash[i] != expectedSum[i] {
			t.Errorf("Hash mismatch at byte %d: got %x, want %x", i, actualHash[i], expectedSum[i])
		}
	}
}

func TestHashFileSpecialChars(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	content := []byte("Special chars test")
	filePath := filepath.Join(tmpDir, "test!@#$%^&*().txt")
	if err := os.WriteFile(filePath, content, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	expectedHash, err := blake2b.New512(nil)
	if err != nil {
		t.Fatalf("Failed to create hash: %v", err)
	}
	if _, err := expectedHash.Write(content); err != nil {
		t.Fatalf("Failed to write to hash: %v", err)
	}
	expectedSum := expectedHash.Sum(nil)

	actualHash, err := relink.HashFile(filePath, testBuffer, nil)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}
	if len(actualHash) != len(expectedSum) {
		t.Errorf("Hash length mismatch: got %d, want %d", len(actualHash), len(expectedSum))
	}
	for i := range actualHash {
		if actualHash[i] != expectedSum[i] {
			t.Errorf("Hash mismatch at byte %d: got %x, want %x", i, actualHash[i], expectedSum[i])
		}
	}
}

func TestHashFileNotExist(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = relink.HashFile(filepath.Join(tmpDir, "nonexistent.txt"), testBuffer, nil)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error, got %v", err)
	}
}

func TestHashFilePermissionDenied(t *testing.T) {
	t.Parallel()
	tmpDir, err := os.MkdirTemp("", "hash_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "no_perms.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0000); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err = relink.HashFile(filePath, testBuffer, nil)
	if err == nil {
		t.Error("Expected error for permission denied, got nil")
	}
	if !os.IsPermission(err) {
		t.Errorf("Expected IsPermission error, got %v", err)
	}
}
