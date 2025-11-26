package media

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/cfg"
)

var (
	previewExt = []string{
		".jpg",
		".jpeg",
		".png",
		".gif",
		".svg",
		".webp",
	}
)

// File is a media file.
type File struct {
	Name    string
	Path    string
	Preview string
	Size    int64
}

// DirectoryContent returns all files inside a media directory.
func DirectoryContent(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	tpl.Get().Execute(w, "media-files.html", struct {
		Lang            string
		Path            string
		Selection       bool
		SelectionTarget string
		SelectionField  SelectionField
		Files           []File
	}{
		Lang:            tpl.GetUILanguage(r),
		Path:            path,
		Selection:       strings.TrimSpace(r.URL.Query().Get("selection")) != "",
		SelectionTarget: strings.TrimSpace(r.URL.Query().Get("target")),
		SelectionField: SelectionField{
			Type:   r.URL.Query().Get("type"),
			Label:  r.URL.Query().Get("label"),
			Field:  r.URL.Query().Get("field"),
			Lang:   r.URL.Query().Get("lang"),
			Prefix: r.URL.Query().Get("prefix"),
		},
		Files: listFiles(path),
	})
}

func listFiles(path string) []File {
	files, err := os.ReadDir(filepath.Join(cfg.Get().BaseDir, mediaDir, path))

	if err != nil {
		slog.Warn("Error reading media directory", "error", err, "path", path)
		return nil
	}

	list := make([]File, 0)

	for _, f := range files {
		if !f.IsDir() {
			info, err := f.Info()

			if err != nil {
				slog.Error("Error reading file info", "error", err, "path", path, "file", f.Name())
				continue
			}

			p := filepath.Join("/", mediaDir, path, f.Name())
			list = append(list, File{
				Name:    f.Name(),
				Path:    p,
				Preview: getFilePreview(p),
				Size:    info.Size(),
			})
		}
	}

	sortFiles(list)
	return list
}

func getFilePreview(path string) string {
	if slices.Contains(previewExt, strings.ToLower(filepath.Ext(path))) {
		return path
	}

	return ""
}

func sortFiles(list []File) {
	slices.SortFunc(list, func(a, b File) int {
		if strings.ToLower(a.Name) > strings.ToLower(b.Name) {
			return 1

		} else if strings.ToLower(a.Name) < strings.ToLower(b.Name) {
			return -1
		}

		return 0
	})
}
