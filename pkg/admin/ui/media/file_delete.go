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
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	fullPath := filepath.Join(getDirectoryPath(path), name)

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		if err := os.Remove(fullPath); err != nil {
			slog.Error("Error while deleting file", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
		Name          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-delete",
			TitleTpl:   "media-file-delete-window-title",
			ContentTpl: "media-file-delete-window-content",
			Overlay:    true,
		},
		Path: path,
		Name: name,
	})
}
