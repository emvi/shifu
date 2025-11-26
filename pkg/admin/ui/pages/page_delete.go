package pages

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
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
			Lang    string
			Entries []Entry
		}{
			Lang:    tpl.GetUILanguage(r),
			Entries: listEntries(w),
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages-page-delete.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Page          string
		Path          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-delete",
			TitleTpl:   "pages-page-delete-window-title",
			ContentTpl: "pages-page-delete-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang: lang,
		Page: getPageName(filepath.Base(path)),
		Path: path,
	})
}
