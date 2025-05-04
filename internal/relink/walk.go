package relink

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Path string
	Info fs.FileInfo
}

func Walk(root string) iter.Seq2[FileInfo, error] {
	return func(yield func(FileInfo, error) bool) {
		// Check if root exists before walking
		if _, err := os.Stat(root); os.IsNotExist(err) {
			return
		}

		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				return nil
			}

			if err != nil {
				return err
			}

			if path == root {
				return nil
			}

			absPath, err := filepath.Abs(path)
			if !yield(FileInfo{absPath, info}, err) {
				return fs.SkipAll
			}

			return nil
		})

		if err != nil {
			yield(FileInfo{}, err)
		}
	}
}
