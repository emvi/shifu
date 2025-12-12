package media

import (
	"net/http"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// SelectionField is the data for the media selection.
type SelectionField struct {
	Lang, Type, Label, Field, Language, Prefix, Value string
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
				Lang            string
				Path            string
				Selection       bool
				SelectionTarget string
				SelectionField  SelectionField
				Files           []File
			}{
				Lang:            tpl.GetUILanguage(r),
				Path:            path,
				Selection:       true,
				SelectionTarget: target,
				SelectionField: SelectionField{
					Lang:     tpl.GetUILanguage(r),
					Type:     r.URL.Query().Get("type"),
					Label:    r.URL.Query().Get("label"),
					Field:    r.URL.Query().Get("field"),
					Language: r.URL.Query().Get("lang"),
					Prefix:   r.URL.Query().Get("prefix"),
				},
				Files: listFiles(path),
			})
			return
		}

		tpl.Get().Execute(w, "page-element-edit-field-file-data.html", SelectionField{
			Lang:     tpl.GetUILanguage(r),
			Type:     r.URL.Query().Get("type"),
			Label:    r.URL.Query().Get("label"),
			Field:    r.URL.Query().Get("field"),
			Language: r.URL.Query().Get("lang"),
			Prefix:   r.URL.Query().Get("prefix"),
			Value:    file,
		})
		return
	}

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
			ID:         "shifu-media-selection",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
			Overlay:    true,
			Lang:       lang,
		},
		Lang:            lang,
		Directories:     shared.ListDirectories(w, mediaDir, false),
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
