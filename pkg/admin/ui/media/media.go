package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
)

// Media renders the media management dialog.
func Media(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions ui.WindowOptions
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   500,
		},
	})
}
