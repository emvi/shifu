package content

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
)

// DeleteReference deletes a reference.
func DeleteReference(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getReferencePath(path)

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodDelete {
		if err := os.Remove(fullPath); err != nil {
			slog.Error("Error while deleting reference", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go content.Update()
		tpl.Get().Execute(w, "refs-tree.html", struct {
			Lang    string
			Entries []Entry
		}{
			Lang:    lang,
			Entries: listReferences(w),
		})
		return
	}

	tpl.Get().Execute(w, "refs-ref-delete.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Name          string
		Path          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-refs-ref-delete",
			TitleTpl:   "refs-ref-delete-window-title",
			ContentTpl: "refs-ref-delete-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang: lang,
		Name: getReferenceName(path),
		Path: path,
	})
}
