package pages

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
)

// DeleteDirectory deletes a directory.
func DeleteDirectory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getPagePath(path)

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		if err := os.RemoveAll(fullPath); err != nil {
			slog.Error("Error while deleting directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tpl.Get().Execute(w, "pages-tree.html", struct {
			Lang    string
			Entries []Entry
		}{
			Lang:    tpl.GetUILanguage(r),
			Entries: listEntries(w),
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages-directory-delete.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Directory     string
		Path          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-directory-delete",
			TitleTpl:   "pages-directory-delete-window-title",
			ContentTpl: "pages-directory-delete-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang:      lang,
		Directory: filepath.Base(path),
		Path:      path,
	})
}

func getPagePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, contentDir, path)
}
