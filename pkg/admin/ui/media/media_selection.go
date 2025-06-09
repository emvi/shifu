package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
	"strings"
)

// Selection renders the media selection dialog.
func Selection(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSpace(r.URL.Query().Get("target"))

	if r.Method == http.MethodPost {
		path := strings.TrimSpace(r.URL.Query().Get("path"))
		file := strings.TrimSpace(r.FormValue("file"))

		if file == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("HX-Reswap", "innerHTML")
			tpl.Get().Execute(w, "media-files.html", struct {
				Path            string
				Selection       bool
				SelectionTarget string
				Files           []File
			}{
				Path:            path,
				Selection:       true,
				SelectionTarget: target,
				Files:           listFiles(path),
			})
			return
		}

		// TODO
		return
	}

	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions   ui.WindowOptions
		Path            string
		Directories     []Directory
		Files           []File
		Interactive     bool
		Selection       bool
		SelectionTarget string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-selection",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
			Overlay:    true,
		},
		Directories:     listDirectories(w),
		Selection:       true,
		SelectionTarget: target,
	})
}
