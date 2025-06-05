package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"strings"
)

// TODO
// - pre-fill fields
// - save nested elements

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
		if err := r.ParseForm(); err != nil {
			slog.Error("Error parsing element form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		element = shared.FindElement(page, elementPath)
		element.Copy = getCopyFromRequest(r)
		element.Data = getDataFromRequest(r)

		if err := shared.SavePage(page, fullPath); err != nil {
			slog.Error("Error saving page while updating element", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	tpl.Get().Execute(w, "page-element-edit.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		ElementPath   string
		Element       *cms.Content
		Config        TemplateConfig
		Languages     []string
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
		Languages:   getPageLanguages(page),
	})
}

func getPageLanguages(page *cms.Content) []string {
	keys := make([]string, 0, len(page.Path))

	for k := range page.Path {
		keys = append(keys, k)
	}

	return keys
}

func getCopyFromRequest(r *http.Request) cms.Copy {
	c := make(cms.Copy)

	for k, v := range r.Form {
		if strings.HasPrefix(k, "copy.") && len(v) > 0 {
			l, key, found := strings.Cut(strings.TrimPrefix(k, "copy."), ".")

			if found {
				if c[l] == nil {
					c[l] = make(map[string]any)
				}

				c[l][key] = v[0]
			}
		}
	}

	return c
}

func getDataFromRequest(r *http.Request) map[string]any {
	data := make(map[string]any)

	for k, v := range r.Form {
		if strings.HasPrefix(k, "data.") && len(v) > 0 {
			data[strings.TrimPrefix(k, "data.")] = v[0]
		}
	}

	return data
}
