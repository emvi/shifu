package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"strings"
)

// DeleteElement deletes an element from a page.
func DeleteElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))

	if r.Method == http.MethodDelete {
		fullPath := getPagePath(path)
		page, err := shared.LoadPage(fullPath)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if deleteElement(page, elementPath) {
			if err := shared.SavePage(page, fullPath); err != nil {
				slog.Error("Error while saving page", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		setTemplateNames(page)
		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "page-tree.html", struct {
			Lang string
			Path string
			Page *cms.Content
		}{
			Lang: tpl.GetLanguage(r),
			Path: path,
			Page: page,
		})
		return
	}

	lang := tpl.GetLanguage(r)
	tpl.Get().Execute(w, "page-element-delete.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Element       string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-delete",
			TitleTpl:   "page-element-delete-window-title",
			ContentTpl: "page-element-delete-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang:    lang,
		Path:    path,
		Element: elementPath,
	})
}
