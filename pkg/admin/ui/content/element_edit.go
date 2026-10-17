package content

import (
	"fmt"
	htmlTpl "html/template"
	"log/slog"
	"maps"
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
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, shared.GetLanguage(r))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// references are edited directly, including their children
	target, err := resolveElementTarget(page, elementPath, true)

	if err != nil {
		slog.Error("Error resolving element", "error", err, "path", elementPath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	element := target.Element()

	if element == nil {
		slog.Error("Element not found", "path", elementPath)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// for references, the reference element on the page (or the parent reference) can override the copy and data
	var refElement *elementTarget

	if target.Path == "" && target.IsRef() {
		refElement, err = resolveElementTarget(page, elementPath, false)

		if err != nil || refElement.Element() == nil {
			slog.Error("Reference element not found", "error", err, "path", elementPath)
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	config, found := tplCfgCache.GetTemplate(element.Tpl)

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

		element.DisplayName = r.FormValue("display_name")
		element.Copy = getCopyFromRequest(r, "copy.")
		element.Data = getDataFromRequest(r, "data.")

		if err := target.Save(); err != nil {
			slog.Error("Error saving element", "error", err, "file", target.File)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// render the element as it appears on the page (for references, the reference element on the page)
		renderElement := element

		if refElement != nil {
			renderElement = refElement.Element()
			renderElement.Copy = removeEmptyCopy(getCopyFromRequest(r, "override.copy."))
			renderElement.Data = removeEmptyData(getDataFromRequest(r, "override.data."))

			if err := refElement.Save(); err != nil {
				slog.Error("Error saving reference overrides", "error", err, "file", refElement.File)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		updatedElement, err := content.RenderElement(w, r, page, elementPath, renderElement)

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

	var overrideCopy, overrideData map[string]any

	if refElement != nil {
		overrideCopy = getCopy(refElement.Element(), "override.copy.")
		overrideData = getData(refElement.Element(), "override.data.")
	}

	windowTitleTpl := "page-element-edit-window-title"

	if target.IsRef() {
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
		ShowOverride  bool
		OverrideCopy  map[string]any
		OverrideData  map[string]any
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-edit",
			TitleTpl:   windowTitleTpl,
			ContentTpl: "page-element-edit-window-content",
			MinWidth:   680,
			Overlay:    true,
			Lang:       lang,
		},
		Language:     shared.GetLanguage(r),
		Lang:         lang,
		Path:         path,
		ElementPath:  elementPath,
		Element:      element,
		Config:       config,
		Languages:    getPageLanguages(page),
		Copy:         getCopy(element, "copy."),
		Data:         getData(element, "data."),
		Override:     target.IsRef(),
		ShowOverride: refElement != nil,
		OverrideCopy: overrideCopy,
		OverrideData: overrideData,
	})
}

func getCopy(element *cms.Content, prefix string) map[string]any {
	c := make(map[string]any)

	for l, t := range element.Copy {
		for k, v := range t {
			c[fmt.Sprintf("%s%s.%s", prefix, l, k)] = v
		}
	}

	return c
}

func getData(element *cms.Content, prefix string) map[string]any {
	c := make(map[string]any)

	for k, v := range element.Data {
		c[prefix+k] = v
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

func getCopyFromRequest(r *http.Request, prefix string) cms.Copy {
	c := make(cms.Copy)

	for k, v := range r.Form {
		if strings.HasPrefix(k, prefix) && len(v) > 0 {
			l, key, found := strings.Cut(strings.TrimPrefix(k, prefix), ".")

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

func getDataFromRequest(r *http.Request, prefix string) map[string]any {
	data := make(map[string]any)

	for k, v := range r.Form {
		if strings.HasPrefix(k, prefix) && len(v) > 0 {
			data[strings.TrimPrefix(k, prefix)] = v[0]
		}
	}

	return data
}

func removeEmptyCopy(c cms.Copy) cms.Copy {
	for l, vars := range c {
		maps.DeleteFunc(vars, func(_ string, v any) bool {
			return isZeroValue(v)
		})

		if len(vars) == 0 {
			delete(c, l)
		}
	}

	if len(c) == 0 {
		return nil
	}

	return c
}

func removeEmptyData(data map[string]any) map[string]any {
	maps.DeleteFunc(data, func(_ string, v any) bool {
		return isZeroValue(v)
	})

	if len(data) == 0 {
		return nil
	}

	return data
}
