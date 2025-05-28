package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DeletePage deletes a page.
func DeletePage(w http.ResponseWriter, r *http.Request) {
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
		if err := os.Remove(fullPath); err != nil {
			slog.Error("Error while deleting page", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tpl.Get().Execute(w, "pages-tree.html", struct {
			Entries []Entry
		}{
			Entries: listEntries(w),
		})
		return
	}

	tpl.Get().Execute(w, "pages-page-delete.html", struct {
		WindowOptions ui.WindowOptions
		Page          string
		Path          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-delete",
			TitleTpl:   "pages-page-delete-window-title",
			ContentTpl: "pages-page-delete-window-content",
			Overlay:    true,
		},
		Page: getPageName(filepath.Base(path)),
		Path: path,
	})
}
