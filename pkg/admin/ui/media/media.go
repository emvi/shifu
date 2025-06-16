package media

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
	mediaDir = "static/media"
)

// Directory is a media directory.
type Directory struct {
	Name     string
	Path     string
	Children []Directory
}

// Media renders the media management dialog.
func Media(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions   ui.WindowOptions
		Lang            string
		Path            string
		Directories     []Directory
		Files           []File
		Interactive     bool
		Selection       bool
		SelectionTarget string
		SelectionField  SelectionField
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
			Lang:       lang,
		},
		Lang:        lang,
		Directories: listDirectories(w),
		Interactive: true,
	})
}

func listDirectories(w http.ResponseWriter) []Directory {
	dir := filepath.Join(cfg.Get().BaseDir, mediaDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating media directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	tree, err := readDirectoryTree(dir, dir)

	if err != nil {
		slog.Error("Error reading media directory", "error", err)
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
