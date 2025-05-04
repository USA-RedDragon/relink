package cmd

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/USA-RedDragon/relink/internal/relink"
	"github.com/spf13/cobra"
)

func NewBufferBenchCommand(version, commit string) *cobra.Command {
	return &cobra.Command{
		Use:     "bufferbench",
		Version: fmt.Sprintf("%s - %s", version, commit),
		Annotations: map[string]string{
			"version": version,
			"commit":  commit,
		},
		RunE:              runBufferBench,
		SilenceErrors:     true,
		DisableAutoGenTag: true,
	}
}

func runBufferBench(cmd *cobra.Command, _ []string) error {
	fmt.Printf("relink - %s (%s)\n", cmd.Annotations["version"], cmd.Annotations["commit"])

	// Create temporary 1gib file in the current directory
	f, err := os.CreateTemp(".", "bufferbench-")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	// Fill it with 1 gib of random data
	size := 1024 * 1024 * 1024
	buf := make([]byte, 4096)
	for i := 0; i < size/4096; i++ {
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("failed to fill buffer: %w", err)
		}
		if _, err := f.Write(buf); err != nil {
			return fmt.Errorf("failed to write buffer: %w", err)
		}
	}
	// Sync the file to ensure all data is written
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Hash it with buffer sizes in powers of 2
	sizes := []int{
		1024, 2048, 4096, 8192, 16384, 32768, 65536,
	}
	times := make([]float64, len(sizes))

	for i, bufsize := range sizes {
		start := time.Now()
		if _, err := relink.HashFile(f.Name(), bufsize, nil); err != nil {
			return fmt.Errorf("failed to hash file: %w", err)
		}
		duration := time.Since(start)
		mibsPerSecond := float64(size) / (1024 * 1024) / duration.Seconds()
		times[i] = mibsPerSecond
	}

	optimalIndex := 0
	optimalSpeed := times[0]
	for i, t := range times {
		if t > optimalSpeed {
			optimalSpeed = t
			optimalIndex = i
		}
	}

	fmt.Printf("Optimal buffer size: %d bytes\n", sizes[optimalIndex])
	fmt.Printf("Optimal speed: %.2f MiB/s\n", optimalSpeed)

	// Print the results in a table with 2 columns, size and MiB/s
	fmt.Printf("%-10s %-10s\n", "Size", "MiB/s")
	for i, bufsize := range sizes {
		fmt.Printf("%-10d %-10.2f\n", bufsize, times[i])
	}

	return nil
}
