package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
)

// Selection renders the media selection dialog.
func Selection(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Directories   []Directory
		Files         []File
		Interactive   bool
		Selection     bool
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-selection",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
			Overlay:    true,
		},
		Directories: listDirectories(w),
		Selection:   true,
	})
}
