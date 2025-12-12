package media

import (
	"net/http"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

const (
	mediaDir = "static/media"
)

// Media renders the media management dialog.
func Media(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions   ui.WindowOptions
		Lang            string
		Path            string
		Directories     []shared.Directory
		Files           []File
		Interactive     bool
		Selection       bool
		SelectionTarget string
		SelectionField  SelectionField
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
			Lang:       lang,
		},
		Lang:        lang,
		Directories: shared.ListDirectories(w, mediaDir, false),
		Interactive: true,
	})
}
