package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"net/http"
	"strings"
)

type SelectionField struct {
	Type, Label, Field, Lang, Prefix, Value string
}

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
				SelectionField  SelectionField
				Files           []File
			}{
				Path:            path,
				Selection:       true,
				SelectionTarget: target,
				SelectionField: SelectionField{
					Type:   r.URL.Query().Get("type"),
					Label:  r.URL.Query().Get("label"),
					Field:  r.URL.Query().Get("field"),
					Lang:   r.URL.Query().Get("lang"),
					Prefix: r.URL.Query().Get("prefix"),
				},
				Files: listFiles(path),
			})
			return
		}

		tpl.Get().Execute(w, "page-element-edit-field-file-data.html", SelectionField{
			Type:   r.URL.Query().Get("type"),
			Label:  r.URL.Query().Get("label"),
			Field:  r.URL.Query().Get("field"),
			Lang:   r.URL.Query().Get("lang"),
			Prefix: r.URL.Query().Get("prefix"),
			Value:  file,
		})
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
		SelectionField  SelectionField
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
		SelectionField: SelectionField{
			Type:   r.URL.Query().Get("type"),
			Label:  r.URL.Query().Get("label"),
			Field:  r.URL.Query().Get("field"),
			Lang:   r.URL.Query().Get("lang"),
			Prefix: r.URL.Query().Get("prefix"),
		},
	})
}
