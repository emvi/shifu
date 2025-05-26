package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	contentDir = "content"
)

// Directory is a content directory.
type Directory struct {
	Name     string
	Path     string
	Children []Directory
}

// Pages renders the pages management dialog.
func Pages(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "pages.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Directories   []Directory
		Interactive   bool
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages",
			TitleTpl:   "pages-window-title",
			ContentTpl: "pages-window-content",
			MinWidth:   800,
		},
		Directories: listDirectories(w),
		Interactive: true,
	})
}

func listDirectories(w http.ResponseWriter) []Directory {
	dir := filepath.Join(cfg.Get().BaseDir, contentDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating content directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	tree, err := readDirectoryTree(dir, dir)

	if err != nil {
		slog.Error("Error reading content directory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	return tree
}

func readDirectoryTree(prefix, dir string) ([]Directory, error) {
	files, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	dirs := make([]Directory, 0)

	for _, file := range files {
		if file.IsDir() {
			path := filepath.Join(dir, file.Name())
			children, err := readDirectoryTree(prefix, path)

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
