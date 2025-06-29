package pages

import (
	"encoding/json"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// SaveJSON renders a JSON editor for the given page and saves it.
func SaveJSON(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getPagePath(path)
	c, err := os.ReadFile(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		data := r.FormValue("json")
		var page cms.Content

		if err := json.Unmarshal([]byte(data), &page); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := shared.SavePage(&page, fullPath); err != nil {
			slog.Error("Error while saving page", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages-page-json.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Saved         bool
		Lang          string
		JSON          string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-json",
			TitleTpl:   "pages-page-json-window-title",
			ContentTpl: "pages-page-json-window-content",
			Overlay:    true,
			MinWidth:   900,
			Lang:       lang,
		},
		Path: path,
		Lang: lang,
		JSON: string(c),
	})
}
