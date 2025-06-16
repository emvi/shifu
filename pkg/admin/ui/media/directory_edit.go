package media

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
		} else if info, _ := os.Stat(getDirectoryPath(path)); info == nil {
			errs["name"] = "the directory does not exist"
		} else if info, _ := os.Stat(getDirectoryPath(filepath.Join(filepath.Dir(path), name))); info != nil {
			errs["name"] = "the directory already exists"
		} else {
			if err := os.Rename(getDirectoryPath(path), getDirectoryPath(filepath.Join(filepath.Dir(path), name))); err != nil {
				slog.Error("Error while creating directory", "error", err)
				errs["name"] = "error renaming directory"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "media-directory-edit-form.html", struct {
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
		tpl.Get().Execute(w, "media-tree.html", struct {
			Lang            string
			Directories     []Directory
			Interactive     bool
			Selection       bool
			SelectionTarget string
			SelectionField  SelectionField
		}{
			Lang:        tpl.GetUILanguage(r),
			Directories: listDirectories(w),
			Interactive: true,
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media-directory-edit.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Name          string
		Path          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-directory-edit",
			TitleTpl:   "media-directory-edit-window-title",
			ContentTpl: "media-directory-edit-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		Lang: lang,
		Name: filepath.Base(path),
		Path: path,
	})
}
