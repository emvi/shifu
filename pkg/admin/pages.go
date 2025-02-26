package admin

import "net/http"

func Pages(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "pages.html", struct {
		WindowOptions WindowOptions
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-pages",
			TitleTpl:   "pages-window-title",
			ContentTpl: "pages-window-content",
		},
	})
}
