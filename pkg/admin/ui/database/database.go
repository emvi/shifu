package database

import (
	"net/http"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
)

// Database renders the database management dialog.
func Database(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "database.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-database",
			TitleTpl:   "database-window-title",
			ContentTpl: "database-window-content",
			MinWidth:   500,
			Lang:       tpl.GetUILanguage(r),
		},
		Lang: tpl.GetUILanguage(r),
	})
}
