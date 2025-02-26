package admin

import "net/http"

func Edit(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "edit.html", struct {
		WindowOptions WindowOptions
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-edit",
			TitleTpl:   "edit-window-title",
			ContentTpl: "edit-window-content",
		},
	})
}
