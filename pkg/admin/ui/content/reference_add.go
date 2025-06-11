package content

import (
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"strings"
)

// AddReference adds an existing reference to the parent element.
func AddReference(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var parent *cms.Content
	positions := make(map[string]string)

	if elementPath != "" {
		var key string
		var index int
		parent, key, index = findParentElement(page, elementPath)

		if parent == nil {
			slog.Debug("Parent element not found", "element", elementPath)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		parent = &parent.Content[key][index]
		parentName := parent.Tpl
		parentTpl, found := tplCache.Get(parentName)

		if !found {
			slog.Debug("Parent template not found", "name", parentName)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		positions = parentTpl.Content
	} else {
		parent = page
	}

	if r.Method == http.MethodPost {
		reference := strings.TrimSpace(r.FormValue("reference"))
		position := strings.TrimSpace(r.FormValue("position"))
		errors := make(map[string]string)
		found := false

		if err := db.Get().Get(&found, `SELECT EXISTS (SELECT 1 FROM "reference" WHERE name = ?)`, reference); err != nil {
			slog.Error("Error checking existence of reference", "name", reference)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !found {
			errors["reference"] = "the reference does not exist"
		}

		if position != "" {
			if _, found := positions[position]; !found {
				errors["position"] = "the position does not exist"
			}
		} else {
			position = "content"
		}

		if len(errors) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "page-reference-add-form.html", struct {
				Lang       string
				Path       string
				Element    string
				References []string
				Positions  map[string]string
				Reference  string
				Position   string
				Errors     map[string]string
			}{
				Lang:       tpl.GetLanguage(r),
				Path:       path,
				Element:    elementPath,
				References: getReferences(),
				Positions:  positions,
				Reference:  reference,
				Position:   position,
				Errors:     errors,
			})
			return
		}

		if addReference(parent, position, reference) {
			if err := shared.SavePage(page, fullPath); err != nil {
				slog.Error("Error while saving page", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		setTemplateNames(page)
		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "page-tree.html", struct {
			Lang string
			Path string
			Page *cms.Content
		}{
			Lang: tpl.GetLanguage(r),
			Path: path,
			Page: page,
		})
		return
	}

	lang := tpl.GetLanguage(r)
	tpl.Get().Execute(w, "page-element-add.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Element       string
		References    []string
		Positions     map[string]string
		Reference     string
		Position      string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-reference-add",
			TitleTpl:   "page-reference-add-window-title",
			ContentTpl: "page-reference-add-window-content",
			MinWidth:   300,
			Lang:       lang,
		},
		Lang:       lang,
		Path:       path,
		Element:    elementPath,
		References: getReferences(),
		Positions:  positions,
	})
}

func getReferences() []string {
	rows, err := db.Get().Query(`SELECT "name" FROM "reference" ORDER BY "name"`)

	if err != nil {
		slog.Error("Error reading references", "error", err)
	}

	references := make([]string, 0)

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			slog.Error("Error reading reference", "error", err)
		}

		references = append(references, name)
	}

	return references
}
