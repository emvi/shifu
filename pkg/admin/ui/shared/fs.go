package shared

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/cfg"
)

// Directory is a file system directory.
type Directory struct {
	Name     string
	Path     string
	Children []Directory
}

// ListDirectories returns the directory tree for given path.
func ListDirectories(w http.ResponseWriter, path string, includeRoot bool) []Directory {
	dir := filepath.Join(cfg.Get().BaseDir, path)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	tree, err := readDirectoryTree(dir, dir, includeRoot)

	if err != nil {
		slog.Error("Error reading directory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	return tree
}

func readDirectoryTree(prefix, dir string, includeRoot bool) ([]Directory, error) {
	files, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	dirs := make([]Directory, 0)

	if includeRoot {
		dirs = append(dirs, Directory{
			Name: "/",
		})
	}

	for _, file := range files {
		if file.IsDir() {
			path := filepath.Join(dir, file.Name())
			children, err := readDirectoryTree(prefix, path, false)

			if err != nil {
				return nil, err
			}

			dirs = append(dirs, Directory{
				Name:     file.Name(),
				Path:     strings.TrimPrefix(path, prefix),
				Children: children,
			})
		}
	}

	sortDirectories(dirs)
	return dirs, nil
}

func sortDirectories(dirs []Directory) {
	slices.SortFunc(dirs, func(a, b Directory) int {
		if strings.ToLower(a.Name) > strings.ToLower(b.Name) {
			return 1

		} else if strings.ToLower(a.Name) < strings.ToLower(b.Name) {
			return -1
		}

		return 0
	})
}

// GetParentDirectory returns the parent directory for given path or an empty string if it is root.
func GetParentDirectory(path string) string {
	parent := filepath.Dir(path)

	if parent == "/" {
		return ""
	}

	return parent
}
