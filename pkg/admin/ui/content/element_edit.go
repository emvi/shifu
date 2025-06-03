package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"log"
	"log/slog"
	"net/http"
	"strings"
)

// EditElement updates the copy and data for an element.
func EditElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	element := shared.FindElement(page, elementPath)

	if element == nil {
		slog.Error("Element not found", "path", elementPath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	config, found := tplCache.Get(element.Tpl)

	if !found {
		slog.Error("Template configuration not found", "name", element.Tpl)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		// TODO
		log.Println(page)
		return
	}

	tpl.Get().Execute(w, "page-element-edit.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		ElementPath   string
		Element       *cms.Content
		Config        TemplateConfig
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-edit",
			TitleTpl:   "page-element-edit-window-title",
			ContentTpl: "page-element-edit-window-content",
			MinWidth:   680,
		},
		Path:        path,
		ElementPath: elementPath,
		Element:     element,
		Config:      config,
	})
}
