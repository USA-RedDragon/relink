package relink

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
)

func Walk(root string) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		// Check if root exists before walking
		if _, err := os.Stat(root); os.IsNotExist(err) {
			return
		}

		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if d.IsDir() {
				return nil
			}

			if err != nil {
				return err
			}

			if path == root {
				return nil
			}

			absPath, err := filepath.Abs(path)
			if !yield(absPath, err) {
				return fs.SkipAll
			}

			return nil
		})

		if err != nil {
			yield("", err)
		}
	}
}
