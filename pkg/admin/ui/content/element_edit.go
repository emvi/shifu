package content

import (
	"fmt"
	htmlTpl "html/template"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
)

// EditElement updates the copy and data for an element.
func EditElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	override := strings.TrimSpace(r.URL.Query().Get("override")) != ""
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, shared.GetLanguage(r))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var element *cms.Content

	if override {
		element = findElement(page, elementPath)

		if element == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		element, err = loadRef(element.Ref)

		if err != nil {
			slog.Error("Error loading reference file", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} else {
		element = findElement(page, elementPath)
	}

	if element == nil {
		slog.Error("Element not found", "path", elementPath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var config TemplateConfig
	var found bool

	if element.Tpl != "" {
		config, found = tplCfgCache.GetTemplate(element.Tpl)
	} else {
		ref, err := loadRef(element.Ref)

		if err != nil {
			slog.Error("Error loading reference file", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		config, found = tplCfgCache.GetTemplate(ref.Tpl)
	}

	if !found {
		slog.Error("Template configuration not found", "name", element.Tpl, "ref", element.Ref)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			slog.Error("Error parsing element form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		element.DisplayName = r.FormValue("display_name")
		element.Copy = getCopyFromRequest(r)
		element.Data = getDataFromRequest(r)

		if override {
			if err := saveRef(element, element.File); err != nil {
				slog.Error("Error saving reference file", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			if setElement(page, elementPath, element) {
				if err := shared.SavePage(page, fullPath); err != nil {
					slog.Error("Error saving page while updating element", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		updatedElement, err := content.RenderElement(w, r, page, elementPath, element)

		if err != nil {
			slog.Error("Error rendering updated element", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		setTemplateNames(page)
		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "page-tree.html", PageTree{
			Language:        shared.GetLanguage(r),
			Lang:            tpl.GetUILanguage(r),
			Path:            path,
			Page:            page,
			Positions:       tplCfgCache.GetPositions(),
			ElementPosition: elementPath,
			UpdateElement:   htmlTpl.HTML(updatedElement),
		})
		go content.Update()
		return
	}

	windowID := "shifu-page-element-edit"
	windowTitleTpl := "page-element-edit-window-title"

	if override {
		windowID = "shifu-page-element-edit-override"
		windowTitleTpl = "page-element-edit-window-title-override"
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page-element-edit.html", struct {
		WindowOptions ui.WindowOptions
		Language      string
		Lang          string
		Path          string
		ElementPath   string
		Element       *cms.Content
		Config        TemplateConfig
		Languages     []string
		Copy          map[string]any
		Data          map[string]any
		Override      bool
	}{
		WindowOptions: ui.WindowOptions{
			ID:         windowID,
			TitleTpl:   windowTitleTpl,
			ContentTpl: "page-element-edit-window-content",
			MinWidth:   680,
			Overlay:    true,
			Lang:       lang,
		},
		Language:    shared.GetLanguage(r),
		Lang:        lang,
		Path:        path,
		ElementPath: elementPath,
		Element:     element,
		Config:      config,
		Languages:   getPageLanguages(page),
		Copy:        getCopy(element),
		Data:        getData(element),
		Override:    override,
	})
}

func getCopy(element *cms.Content) map[string]any {
	c := make(map[string]any)

	for l, t := range element.Copy {
		for k, v := range t {
			c[fmt.Sprintf("copy.%s.%s", l, k)] = v
		}
	}

	return c
}

func getData(element *cms.Content) map[string]any {
	c := make(map[string]any)

	for k, v := range element.Data {
		c[fmt.Sprintf("data.%s", k)] = v
	}

	return c
}

func getPageLanguages(page *cms.Content) []string {
	keys := make([]string, 0, len(page.Path))

	for k := range page.Path {
		keys = append(keys, k)
	}

	slices.Sort(keys)
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
