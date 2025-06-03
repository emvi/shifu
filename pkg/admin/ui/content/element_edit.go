package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"log"
	"net/http"
	"strings"
)

// EditElement updates the copy and data for an element.
func EditElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	element := strings.TrimSpace(r.URL.Query().Get("element"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		// TODO
		log.Println(page)

		return
	}

	tpl.Get().Execute(w, "page-element-edit.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Element       string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-edit",
			TitleTpl:   "page-element-edit-window-title",
			ContentTpl: "page-element-edit-window-content",
			MinWidth:   680,
		},
		Path:    path,
		Element: element,
	})
}
