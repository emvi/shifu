package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	htmlTpl "html/template"
	"log/slog"
	"net/http"
	"strings"
)

// AddElementData is the data required to render the element dialog.
type AddElementData struct {
	Language  string
	Lang      string
	Path      string
	Element   string
	Templates []TemplateConfig
	Positions map[string]TemplateContent
	Template  string
	Position  string
	Errors    map[string]string
}

// AddElement adds a new element to the page or to a parent element.
func AddElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, shared.GetLanguage(r))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var parent *cms.Content
	var parentPath string
	filter := make([]string, 0)
	positions := make(map[string]TemplateContent)

	if elementPath != "" {
		var key string
		var index int
		parent, parentPath, key, index = findParentElement(page, elementPath)

		if parent == nil {
			slog.Debug("Parent element not found", "element", elementPath)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		parent = &parent.Content[key][index]
		parentName := parent.Tpl
		parentTpl, found := tplCfgCache.GetTemplate(parentName)

		if !found {
			slog.Debug("Parent template not found", "name", parentName)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		positions = parentTpl.Content

		for _, c := range parentTpl.Content {
			filter = append(filter, c.TplFilter...)
		}
	} else {
		parent = page
	}

	if r.Method == http.MethodPost {
		template := strings.TrimSpace(r.FormValue("template"))
		position := strings.TrimSpace(r.FormValue("position"))
		errs := make(map[string]string)
		t, found := tplCfgCache.GetTemplate(template)

		if !found {
			errs["template"] = "the template does not exist"
		}

		if position != "" {
			if _, found := positions[position]; !found {
				errs["position"] = "the position does not exist"
			}
		} else {
			position = "content"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "page-element-add-form.html", AddElementData{
				Language:  shared.GetLanguage(r),
				Lang:      tpl.GetUILanguage(r),
				Path:      path,
				Element:   elementPath,
				Templates: tplCfgCache.List(filter),
				Positions: positions,
				Template:  template,
				Position:  position,
				Errors:    errs,
			})
			return
		}

		element := addElement(parent, parentPath, position, template, t.Positions())

		if element != nil {
			if err := shared.SavePage(page, fullPath); err != nil {
				slog.Error("Error while saving page", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		newElement, err := content.RenderElement(w, r, page, element.Position, element)

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
			ParentElement:   parentPath,
			ElementPosition: position,
			AddElement:      htmlTpl.HTML(newElement),
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page-element-add.html", struct {
		WindowOptions ui.WindowOptions
		AddElementData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-add",
			TitleTpl:   "page-element-add-window-title",
			ContentTpl: "page-element-add-window-content",
			MinWidth:   300,
			Overlay:    true,
			Lang:       lang,
		},
		AddElementData: AddElementData{
			Language:  shared.GetLanguage(r),
			Lang:      lang,
			Path:      path,
			Element:   elementPath,
			Templates: tplCfgCache.List(filter),
			Positions: positions,
		},
	})
}
