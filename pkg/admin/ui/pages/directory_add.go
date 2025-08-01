package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// AddDirectory creates a new subdirectory.
func AddDirectory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		fullPath := getPagePath(filepath.Join(path, name))
		errs := make(map[string]string)

		if name == "" {
			errs["name"] = "the name is required"
		} else if !isValidDirectoryName(name) {
			errs["name"] = "the name contains invalid characters"
		} else if info, _ := os.Stat(fullPath); info != nil {
			errs["name"] = "the directory already exists"
		} else {
			if err := os.Mkdir(fullPath, 0755); err != nil {
				slog.Error("Error while creating directory", "error", err)
				errs["name"] = "error creating directory"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "pages-directory-create-form.html", struct {
				Lang   string
				Name   string
				Path   string
				Errors map[string]string
			}{
				Lang:   tpl.GetUILanguage(r),
				Name:   name,
				Path:   path,
				Errors: errs,
			})
			return
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "pages-tree.html", struct {
			Lang    string
			Entries []Entry
		}{
			Lang:    tpl.GetUILanguage(r),
			Entries: listEntries(w),
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages-directory-create.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Name          string
		Path          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-directory-create",
			TitleTpl:   "pages-directory-create-window-title",
			ContentTpl: "pages-directory-create-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang: lang,
		Path: path,
	})
}

func isValidDirectoryName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' && c != ' ' && c != '.' {
			return false
		}
	}

	return true
}
