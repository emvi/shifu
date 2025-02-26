package admin

import "net/http"

func Database(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "database.html", struct {
		WindowOptions WindowOptions
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-database",
			TitleTpl:   "database-window-title",
			ContentTpl: "database-window-content",
			MinWidth:   500,
		},
	})
}
