package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
)

// Edit renders the page editing dialog.
func Edit(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "edit.html", struct {
		WindowOptions ui.WindowOptions
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-edit",
			TitleTpl:   "edit-window-title",
			ContentTpl: "edit-window-content",
		},
	})
}
