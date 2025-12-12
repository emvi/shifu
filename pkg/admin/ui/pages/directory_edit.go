package pages

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// EditDirectoryData is the data for the directory form.
type EditDirectoryData struct {
	Lang           string
	Directories    []shared.Directory
	SelectionField string
	SelectionID    string
	Name           string
	Selected       string
	Path           string
	Errors         map[string]string
}

// EditDirectory changes the name of a directory.
func EditDirectory(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		parent := strings.TrimSpace(r.FormValue("parent"))
		name := strings.TrimSpace(r.FormValue("name"))
		parentChanged := shared.GetParentDirectory(path) != parent
		errs := make(map[string]string)

		if parentChanged {
			p := getPagePath(filepath.Join(parent, name))

			if _, err := os.Stat(p); !errors.Is(err, fs.ErrNotExist) {
				errs["parent"] = "a different directory with this name exists already"
			}
		}

		if name == "" {
			errs["name"] = "the name is required"
		} else if !isValidDirectoryName(name) {
			errs["name"] = "the name contains invalid characters"
		} else if info, _ := os.Stat(getPagePath(path)); info == nil {
			errs["name"] = "the directory does not exist"
		} else {
			p := getPagePath(filepath.Join(filepath.Dir(path), name))

			if parentChanged {
				p = getPagePath(filepath.Join(parent, name))
			}

			if info, _ := os.Stat(p); info != nil {
				errs["name"] = "the directory already exists"
			} else if err := os.Rename(getPagePath(path), p); err != nil {
				slog.Error("Error while creating directory", "error", err)
				errs["name"] = "error renaming directory"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "pages-directory-edit-form.html", EditDirectoryData{
				Lang:           tpl.GetUILanguage(r),
				Directories:    shared.ListDirectories(w, contentDir, true),
				SelectionField: "parent",
				SelectionID:    "page-directory-edit",
				Name:           name,
				Selected:       shared.GetParentDirectory(path),
				Path:           path,
				Errors:         errs,
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
		EditDirectoryData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-directory-edit",
			TitleTpl:   "pages-directory-edit-window-title",
			ContentTpl: "pages-directory-edit-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		EditDirectoryData: EditDirectoryData{
			Lang:           lang,
			Directories:    shared.ListDirectories(w, contentDir, true),
			SelectionField: "parent",
			SelectionID:    "page-directory-edit",
			Name:           filepath.Base(path),
			Selected:       shared.GetParentDirectory(path),
			Path:           path,
		},
	})
}
