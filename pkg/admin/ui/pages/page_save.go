package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
	"strings"
)

// SavePage creates or updates a page.
func SavePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	tpl.Get().Execute(w, "pages-page-save.html", struct {
		WindowOptions ui.WindowOptions
		Name          string
		Path          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-save",
			TitleTpl:   "pages-page-save-window-title",
			ContentTpl: "pages-page-save-window-content",
			Overlay:    true,
			MinWidth:   520,
		},
		Path: path,
	})
}
