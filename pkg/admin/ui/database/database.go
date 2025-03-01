package database

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
)

// Database renders the database management dialog.
func Database(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "database.html", struct {
		WindowOptions ui.WindowOptions
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-database",
			TitleTpl:   "database-window-title",
			ContentTpl: "database-window-content",
			MinWidth:   500,
		},
	})
}
