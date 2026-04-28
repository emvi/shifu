package content

import (
	htmlTpl "html/template"
	"log/slog"
	"maps"
	"net/http"
	"reflect"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
)

// RemoveReferenceData is the data to remove a reference.
type RemoveReferenceData struct {
	WindowOptions ui.WindowOptions
	Language      string
	Lang          string
	Path          string
	Element       string
}

// RemoveReference turns a reference into a copy of an element on a page.
// It won't touch the original reference.
func RemoveReference(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))

	if r.Method == http.MethodPost {
		// load page and find existing reference
		fullPath := getPagePath(path)
		page, err := shared.LoadPage(r, fullPath, shared.GetLanguage(r))

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		element := findElement(page, elementPath)

		if element == nil || element.Ref == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ref := content.GetReference(element.Ref)

		if ref == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// overwrite data, copy, and configuration
		if ref.Data == nil {
			ref.Data = make(map[string]any)
		}

		for k, v := range element.Data {
			if !isZeroValue(v) {
				ref.Data[k] = v
			}
		}

		if ref.Copy == nil {
			ref.Copy = make(cms.Copy)
		}

		// overwrite copy
		for language, vars := range element.Copy {
			if _, ok := ref.Copy[language]; !ok {
				ref.Copy[language] = make(map[string]any)
			}

			for k, v := range vars {
				if v != "" {
					ref.Copy[language][k] = v
				}
			}
		}

		// overwrite analytics
		if ref.Analytics.Tags == nil {
			ref.Analytics.Tags = make(map[string]string)
		}

		maps.Copy(ref.Analytics.Tags, element.Analytics.Tags)

		if element.Analytics.Experiment.Name != "" {
			ref.Analytics.Experiment.Name = element.Analytics.Experiment.Name
		}

		if element.Analytics.Experiment.Variant != "" {
			ref.Analytics.Experiment.Variant = element.Analytics.Experiment.Variant
		}

		ref.DisplayName = element.DisplayName
		ref.Position = elementPath

		// replace element and save page
		if !replaceElement(page, elementPath, ref) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := shared.SavePage(page, fullPath); err != nil {
			slog.Error("Error saving page while replacing reference", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		newElement, err := content.RenderElement(w, r, page, ref.Position, ref)

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
			UpdateElement:   htmlTpl.HTML(newElement),
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page-reference-remove.html", struct {
		WindowOptions ui.WindowOptions
		RemoveReferenceData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-reference-remove",
			TitleTpl:   "page-reference-remove-window-title",
			ContentTpl: "page-reference-remove-window-content",
			MinWidth:   300,
			MaxWidth:   500,
			Overlay:    true,
			Lang:       lang,
		},
		RemoveReferenceData: RemoveReferenceData{
			Language: shared.GetLanguage(r),
			Lang:     lang,
			Path:     path,
			Element:  elementPath,
		},
	})
}

func isZeroValue(v any) bool {
	if v == nil {
		return true
	}

	return reflect.ValueOf(v).IsZero()
}
