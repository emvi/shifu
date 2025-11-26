package pages

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
)

const (
	contentDir = "content/pages"
)

// Entry is a content directory or page.
type Entry struct {
	Name     string
	Path     string
	IsDir    bool
	Children []Entry
}

// Pages renders the pages management dialog.
func Pages(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Entries       []Entry
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages",
			TitleTpl:   "pages-window-title",
			ContentTpl: "pages-window-content",
			MinWidth:   680,
			Lang:       lang,
		},
		Lang:    lang,
		Entries: listEntries(w),
	})
}

func listEntries(w http.ResponseWriter) []Entry {
	dir := filepath.Join(cfg.Get().BaseDir, contentDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating content directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	tree, err := readContentTree(dir, dir)

	if err != nil {
		slog.Error("Error reading content directory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	return tree
}

func readContentTree(prefix, dir string) ([]Entry, error) {
	files, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	dirs := make([]Entry, 0)

	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		children := make([]Entry, 0)

		if file.IsDir() {
			children, err = readContentTree(prefix, path)

			if err != nil {
				return nil, err
			}
		}

		dirs = append(dirs, Entry{
			Name:     getPageName(file.Name()),
			Path:     strings.TrimPrefix(path, prefix),
			IsDir:    file.IsDir(),
			Children: children,
		})
	}

	sortEntries(dirs)
	return dirs, nil
}

func sortEntries(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		nameA := strings.ToLower(a.Name)
		nameB := strings.ToLower(b.Name)

		if nameA > nameB {
			return 1
		} else if nameA < nameB {
			return -1
		}

		return 0
	})
}

func getPageName(filename string) string {
	name, _, _ := strings.Cut(filename, ".")
	return name
}
