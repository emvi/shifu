package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
)

// Pages renders the pages management dialog.
func Pages(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "pages.html", struct {
		WindowOptions ui.WindowOptions
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages",
			TitleTpl:   "pages-window-title",
			ContentTpl: "pages-window-content",
		},
	})
}
