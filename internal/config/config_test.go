package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/USA-RedDragon/relink/internal/config"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "relink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		config  config.Config
		wantErr error
	}{
		{
			name: "valid config",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: nil,
		},
		{
			name: "invalid log level",
			config: config.Config{
				LogLevel:   "invalid",
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrBadLogLevel,
		},
		{
			name: "missing source",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     "",
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrNoSource,
		},
		{
			name: "missing target",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     "",
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrNoTarget,
		},
		{
			name: "source and target same",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     tempDir,
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrSourceAndTargetSame,
		},
		{
			name: "source directory not found",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     filepath.Join(tempDir, "non-existent"),
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrSourceNotFound,
		},
		{
			name: "zero buffer size",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 0,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrZeroBufferSize,
		},
		{
			name: "zero hash jobs",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   0,
				BufferSize: 1024,
				CacheType:  config.CacheTypeMemory,
			},
			wantErr: config.ErrZeroHashJobs,
		},
		{
			name: "invalid cache type",
			config: config.Config{
				LogLevel:   config.LogLevelInfo,
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				HashJobs:   4,
				BufferSize: 1024,
				CacheType:  "invalid",
			},
			wantErr: config.ErrInvalidCacheType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantErr)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func TestLogLevelConstants(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "relink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		logLevel config.LogLevel
		valid    bool
	}{
		{"debug level", config.LogLevelDebug, true},
		{"info level", config.LogLevelInfo, true},
		{"warn level", config.LogLevelWarn, true},
		{"error level", config.LogLevelError, true},
		{"invalid level", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Config{
				LogLevel:   tt.logLevel,
				Source:     tempDir,
				Target:     filepath.Join(tempDir, "target"),
				BufferSize: 1024,
				HashJobs:   4,
				CacheType:  config.CacheTypeMemory,
			}
			err := cfg.Validate()
			if tt.valid {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			} else if !errors.Is(err, config.ErrBadLogLevel) {
				t.Errorf("Validate() error = %v, want %v", err, config.ErrBadLogLevel)
			}
		})
	}
}
