package content

import (
	htmlTpl "html/template"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
)

// Ref is a referenced element to be displayed in the selection.
type Ref struct {
	Name  string
	Label string
}

// AddReferenceData is the data required to render the reference dialog.
type AddReferenceData struct {
	Language   string
	Lang       string
	Path       string
	Element    string
	References []Ref
	Positions  map[string]TemplateContent
	Reference  string
	Position   string
	Errors     map[string]string
}

// AddReference adds an existing reference to the parent element.
func AddReference(w http.ResponseWriter, r *http.Request) {
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
		reference := strings.TrimSpace(r.FormValue("reference"))
		position := strings.TrimSpace(r.FormValue("position"))
		errs := make(map[string]string)
		found := false

		if err := db.Get().Get(&found, `SELECT EXISTS (SELECT 1 FROM "reference" WHERE name = ?)`, reference); err != nil {
			slog.Error("Error checking existence of reference", "name", reference)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !found {
			errs["reference"] = "the reference does not exist"
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
			tpl.Get().Execute(w, "page-reference-add-form.html", AddReferenceData{
				Language:   shared.GetLanguage(r),
				Lang:       tpl.GetUILanguage(r),
				Path:       path,
				Element:    elementPath,
				References: getReferences(filter),
				Positions:  positions,
				Reference:  reference,
				Position:   position,
				Errors:     errs,
			})
			return
		}

		element := addReference(parent, parentPath, position, reference)

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
		AddReferenceData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-reference-add",
			TitleTpl:   "page-reference-add-window-title",
			ContentTpl: "page-reference-add-window-content",
			MinWidth:   300,
			Overlay:    true,
			Lang:       lang,
		},
		AddReferenceData: AddReferenceData{
			Language:   shared.GetLanguage(r),
			Lang:       lang,
			Path:       path,
			Element:    elementPath,
			References: getReferences(filter),
			Positions:  positions,
		},
	})
}

func getReferences(filter []string) []Ref {
	rows, err := db.Get().Query(`SELECT "name" FROM "reference" ORDER BY "name"`)

	if err != nil {
		slog.Error("Error reading references", "error", err)
	}

	references := make([]Ref, 0)

	for rows.Next() {
		var entity Ref

		if err := rows.Scan(&entity.Name); err != nil {
			slog.Error("Error reading reference", "error", err)
		}

		// FIXME optimize
		ref, err := loadRef(entity.Name)

		if err != nil {
			slog.Error("Error loading reference file", "error", err)
			continue
		}

		name, found := tplCfgCache.GetTemplate(ref.Tpl)

		if found {
			entity.Label = name.Label
		}

		if len(filter) == 0 || slices.Contains(filter, name.Name) {
			references = append(references, entity)
		}
	}

	return references
}
