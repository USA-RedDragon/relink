package relink_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/USA-RedDragon/relink/internal/relink"
)

func TestGetSafeTempFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dir       string
		prefix    string
		wantErr   bool
		setupFn   func() error
		cleanupFn func() error
	}{
		{
			name:    "successful temp file creation",
			dir:     os.TempDir(),
			prefix:  "test-",
			wantErr: false,
		},
		{
			name:    "invalid directory",
			dir:     "/nonexistent/directory",
			prefix:  "test-",
			wantErr: true,
		},
		{
			name:    "empty prefix",
			dir:     os.TempDir(),
			prefix:  "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.setupFn != nil {
				if err := tt.setupFn(); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			if tt.cleanupFn != nil {
				defer func() {
					if err := tt.cleanupFn(); err != nil {
						t.Errorf("cleanup failed: %v", err)
					}
				}()
			}

			got, err := relink.GetSafeTempFile(tt.dir, tt.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSafeTempFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the file doesn't exist (since it should be removed)
				if _, err := os.Stat(got); err == nil {
					t.Errorf("getSafeTempFile() returned path %s still exists", got)
				}

				// Verify the path is in the correct directory
				dir := filepath.Dir(got)
				if dir != tt.dir {
					t.Errorf("getSafeTempFile() returned path in wrong directory, got %s, want %s", dir, tt.dir)
				}

				// Verify the prefix is correct
				base := filepath.Base(got)
				if tt.prefix != "" && !strings.HasPrefix(base, tt.prefix) {
					t.Errorf("getSafeTempFile() returned path with wrong prefix, got %s, want prefix %s", base, tt.prefix)
				}
			}
		})
	}
}
