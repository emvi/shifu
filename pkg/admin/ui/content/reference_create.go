package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	refsDir = "content/refs"
)

// CreateReference creates a new reference for the given element.
func CreateReference(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, "")

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	element := findElement(page, elementPath)

	if element == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		name := strings.ToLower(strings.TrimSpace(r.FormValue("name")))
		errs := make(map[string]string)

		if name == "" {
			errs["name"] = "the name is required"
		} else if !isValidFileName(name) {
			errs["name"] = "the name contains invalid characters"
		} else if templateExists(name) {
			errs["name"] = "a template or reference with this name already exists"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "page-element-reference-form.html", struct {
				Language string
				Lang     string
				Path     string
				Element  string
				Name     string
				Errors   map[string]string
			}{
				Language: shared.GetLanguage(r),
				Lang:     tpl.GetUILanguage(r),
				Path:     path,
				Element:  elementPath,
				Name:     name,
				Errors:   errs,
			})
			return
		}

		if err := createReference(element, name); err != nil {
			slog.Error("Error creating reference file", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if setElement(page, elementPath, &cms.Content{Ref: name}) {
			if err := shared.SavePage(page, fullPath); err != nil {
				slog.Error("Error while saving page", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		content.Update()
		setTemplateNames(page)
		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "page-tree.html", PageTree{
			Language:  shared.GetLanguage(r),
			Lang:      tpl.GetUILanguage(r),
			Path:      path,
			Page:      page,
			Positions: tplCfgCache.GetPositions(),
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "page-element-reference.html", struct {
		WindowOptions ui.WindowOptions
		Language      string
		Lang          string
		Path          string
		Element       string
		Name          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page-element-reference",
			TitleTpl:   "page-element-reference-window-title",
			ContentTpl: "page-element-reference-window-content",
			MinWidth:   300,
			Overlay:    true,
			Lang:       lang,
		},
		Language: shared.GetLanguage(r),
		Lang:     lang,
		Path:     path,
		Element:  elementPath,
	})
}

func isValidFileName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' && c != ' ' && c != '.' && c != '/' {
			return false
		}
	}

	return true
}

func templateExists(name string) bool {
	name = fmt.Sprintf("%s.json", name)
	found := false

	if err := filepath.WalkDir(filepath.Join(cfg.Get().BaseDir, contentDir), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Base(path) == name {
			found = true
			return fs.SkipAll
		}

		return nil
	}); err != nil && !errors.Is(err, fs.SkipAll) {
		slog.Error("Error listing templates", "error", err)
		return false
	}

	return found
}

func createReference(element *cms.Content, name string) error {
	if _, err := os.Stat(filepath.Join(cfg.Get().BaseDir, refsDir)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(cfg.Get().BaseDir, refsDir), os.ModePerm); err != nil {
			return err
		}
	}

	out, err := json.Marshal(element)

	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(cfg.Get().BaseDir, refsDir, fmt.Sprintf("%s.json", name)), out, 0644); err != nil {
		return err
	}

	if _, err := db.Get().Exec(`INSERT INTO "reference" ("name") VALUES (?)`, name); err != nil {
		return err
	}

	return nil
}
