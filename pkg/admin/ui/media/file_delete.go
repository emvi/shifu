package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DeleteFile deletes a file.
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	names := r.URL.Query()["name[]"]

	if len(names) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	files := make([]string, 0, len(names))

	for i, name := range names {
		names[i] = strings.TrimSpace(name)
		fullPath := filepath.Join(getDirectoryPath(path), names[i])

		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		files = append(files, fullPath)
	}

	if r.Method == http.MethodDelete {
		for _, file := range files {
			if err := os.Remove(file); err != nil {
				slog.Error("Error while deleting file", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		tpl.Get().Execute(w, "media-files.html", struct {
			Path  string
			Files []File
		}{
			Path:  path,
			Files: listFiles(path),
		})
		return
	}

	tpl.Get().Execute(w, "media-file-delete.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Name          []string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-delete",
			TitleTpl:   "media-file-delete-window-title",
			ContentTpl: "media-file-delete-window-content",
			Overlay:    true,
		},
		Path: path,
		Name: names,
	})
}
