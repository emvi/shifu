package content

import (
	"fmt"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
)

// PageTree is the data required to render the page tree.
type PageTree struct {
	Language  string
	Lang      string
	Path      string
	Page      *cms.Content
	Positions map[string]string

	ParentElement    string
	ElementPosition  string
	ElementDirection string
	AddElement       template.HTML
	MoveElement      string
	DeleteElement    string
}

// Page renders the page editing dialog.
func Page(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, "")

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setTemplateNames(page)
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page.html", struct {
		WindowOptions ui.WindowOptions
		PageTree
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page",
			TitleTpl:   "page-window-title",
			ContentTpl: "page-window-content",
			MinWidth:   300,
			Lang:       lang,
		},
		PageTree: PageTree{
			Language:  shared.GetLanguage(r),
			Lang:      lang,
			Path:      path,
			Page:      page,
			Positions: tplCfgCache.GetPositions(),
		},
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
				name, found = tplCfgCache.GetTemplate(content.Content[k][i].Tpl)

				if found {
					content.Content[k][i].Tpl = name.Label
				}
			} else {
				// FIXME optimize
				ref, err := loadRef(content.Content[k][i].Ref)

				if err != nil {
					slog.Error("Error loading reference file", "error", err)
					continue
				}

				name, found = tplCfgCache.GetTemplate(ref.Tpl)

				if found {
					content.Content[k][i].Tpl = fmt.Sprintf("%s (%s)", name.Label, name.Name)
				}
			}

			setTemplateNames(&content.Content[k][i])
		}
	}
}
