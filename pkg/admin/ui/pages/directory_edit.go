package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// EditDirectory changes the name of a directory.
func EditDirectory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		errs := make(map[string]string)

		if name == "" {
			errs["name"] = "the name is required"
		} else if !isValidDirectoryName(name) {
			errs["name"] = "the name contains invalid characters"
		} else if info, _ := os.Stat(getPagePath(path)); info == nil {
			errs["name"] = "the directory does not exist"
		} else if info, _ := os.Stat(getPagePath(filepath.Join(filepath.Dir(path), name))); info != nil {
			errs["name"] = "the directory already exists"
		} else {
			if err := os.Rename(getPagePath(path), getPagePath(filepath.Join(filepath.Dir(path), name))); err != nil {
				slog.Error("Error while creating directory", "error", err)
				errs["name"] = "error renaming directory"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "pages-directory-edit-form.html", struct {
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
	tpl.Get().Execute(w, "pages-directory-edit.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Name          string
		Path          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-directory-edit",
			TitleTpl:   "pages-directory-edit-window-title",
			ContentTpl: "pages-directory-edit-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang: lang,
		Name: filepath.Base(path),
		Path: path,
	})
}
