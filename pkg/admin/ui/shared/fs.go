package shared

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// GetDirectories returns all relative child directory paths for given path including the path itself.
func GetDirectories(path string) []string {
	dirs := make([]string, 0)

	if err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if d != nil && d.IsDir() {
			dirs = append(dirs, strings.TrimPrefix(p, path))
		}

		return err
	}); err != nil {
		return nil
	}

	slices.SortFunc(dirs, func(a, b string) int {
		if a > b {
			return -1
		} else if a < b {
			return 1
		}

		return 0
	})
	return dirs
}
