package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

// TmpDir returns the path to the cache directory.
// It is created if necessary which might result in an error.
func TmpDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "tmpc")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create tmp dir: %w", err)
	}
	return dir, nil
}
