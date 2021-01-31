package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

func tmpDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "tmpc")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("could not create tmp dir: %w", err)
	}
	return dir, nil
}
