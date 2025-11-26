package content

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// DeleteElement deletes an element from a page.
func DeleteElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))

	if r.Method == http.MethodDelete {
		fullPath := getPagePath(path)
		page, err := shared.LoadPage(r, fullPath, shared.GetLanguage(r))

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
		tpl.Get().Execute(w, "page-tree.html", PageTree{
			Language:      shared.GetLanguage(r),
			Lang:          tpl.GetUILanguage(r),
			Path:          path,
			Page:          page,
			Positions:     tplCfgCache.GetPositions(),
			DeleteElement: elementPath,
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page-element-delete.html", struct {
		WindowOptions ui.WindowOptions
		Language      string
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
		Language: shared.GetLanguage(r),
		Lang:     lang,
		Path:     path,
		Element:  elementPath,
	})
}
