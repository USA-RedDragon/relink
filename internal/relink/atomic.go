package relink

import (
	"fmt"
	"os"
	"path/filepath"
)

func AtomicLink(source, target string) error {
	tempName, err := GetSafeTempFile(filepath.Dir(target), ".relink-"+filepath.Base(target))
	if err != nil {
		return fmt.Errorf("failed to create temp file for %s: %w", target, err)
	}
	defer os.Remove(tempName)
	if err = os.Link(source, tempName); err != nil {
		return fmt.Errorf("failed to create hardlink from %s to %s: %w", source, tempName, err)
	}
	if err = os.Rename(tempName, target); err != nil {
		return fmt.Errorf("failed to move hardlink from %s to %s: %w", tempName, target, err)
	}
	return nil
}
