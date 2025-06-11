package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"net/http"
	"path/filepath"
)

// Page renders the page editing dialog.
func Page(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setTemplateNames(page)
	lang := tpl.GetLanguage(r)
	tpl.Get().Execute(w, "page.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Page          *cms.Content
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page",
			TitleTpl:   "page-window-title",
			ContentTpl: "page-window-content",
			MinWidth:   300,
			Lang:       lang,
		},
		Lang: lang,
		Path: path,
		Page: page,
	})
}

func getPagePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, path)
}

func setTemplateNames(content *cms.Content) {
	for k, v := range content.Content {
		for i := range v {
			var name TemplateConfig
			var found bool

			if content.Content[k][i].Tpl != "" {
				name, found = tplCache.Get(content.Content[k][i].Tpl)
			} else {
				name, found = tplCache.Get(content.Content[k][i].Ref)
			}

			if found {
				content.Content[k][i].Tpl = name.Label
			}

			setTemplateNames(&content.Content[k][i])
		}
	}
}
