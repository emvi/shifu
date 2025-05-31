package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"net/http"
)

// DeleteElement deletes an element from a page.
func DeleteElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodDelete {
		// TODO
	}

	tpl.Get().Execute(w, "page-element-delete.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Page          *cms.Content
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-delete",
			TitleTpl:   "page-element-delete-window-title",
			ContentTpl: "page-element-delete-window-content",
		},
		Path: path,
		Page: page,
	})
}
