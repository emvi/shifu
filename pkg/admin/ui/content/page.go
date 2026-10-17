package content

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
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
	UpdateElement    template.HTML
	MoveElement      string
	DeleteElement    string
}

// Page renders the page editing dialog.
func Page(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath := getPagePath(path)
	page, err := loadDocument(r, fullPath, "")

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
		Language:  shared.GetLanguage(r),
		Lang:      lang,
		Path:      path,
		Page:      page,
		Positions: tplCfgCache.GetPositions(),
	})
}

func getPagePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, path)
}

func setTemplateNames(content *cms.Content) {
	setTemplateNamesDepth(content, 0)
}

func setTemplateNamesDepth(content *cms.Content, depth int) {
	for k, v := range content.Content {
		for i := range v {
			element := &content.Content[k][i]

			if element.Ref != "" {
				ref, err := loadRef(element.Ref)

				if err != nil {
					slog.Error("Error loading reference file", "error", err, "ref", element.Ref)
					continue
				}

				name, found := tplCfgCache.GetTemplate(ref.Tpl)

				if found {
					element.Tpl = fmt.Sprintf("%s (%s)", name.Label, element.Ref)
				}

				if element.DisplayName == "" {
					element.DisplayName = ref.DisplayName
				}

				// show the children of the reference
				if depth < maxRefDepth {
					element.Content = ref.Content
				} else {
					element.Content = nil
				}

				setTemplateNamesDepth(element, depth+1)
			} else {
				name, found := tplCfgCache.GetTemplate(element.Tpl)

				if found {
					element.Tpl = name.Label
				}

				setTemplateNamesDepth(element, depth)
			}
		}
	}
}
