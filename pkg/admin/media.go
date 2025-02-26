package admin

import "net/http"

func Media(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "media.html", struct {
		WindowOptions WindowOptions
	}{
		WindowOptions: WindowOptions{
			ID:         "shifu-media",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   500,
		},
	})
}
