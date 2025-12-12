package media

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// AddDirectoryData is the data for the directory form.
type AddDirectoryData struct {
	Lang           string
	Directories    []shared.Directory
	SelectionField string
	SelectionID    string
	Name           string
	Selected       string
	Errors         map[string]string
}

// AddDirectory creates a new subdirectory.
func AddDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		parent := strings.TrimSpace(r.FormValue("parent"))
		name := strings.TrimSpace(r.FormValue("name"))
		fullPath := getDirectoryPath(filepath.Join(parent, name))
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
			tpl.Get().Execute(w, "media-directory-create-form.html", AddDirectoryData{
				Lang:           tpl.GetUILanguage(r),
				Directories:    shared.ListDirectories(w, mediaDir, true),
				SelectionField: "parent",
				SelectionID:    "media-directory-add",
				Name:           name,
				Selected:       parent,
				Errors:         errs,
			})
			return
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "media-tree.html", struct {
			Lang            string
			Directories     []shared.Directory
			Interactive     bool
			Selection       bool
			SelectionTarget string
			SelectionField  SelectionField
		}{
			Lang:        tpl.GetUILanguage(r),
			Directories: shared.ListDirectories(w, mediaDir, false),
			Interactive: true,
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media-directory-create.html", struct {
		WindowOptions ui.WindowOptions
		AddDirectoryData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-directory-create",
			TitleTpl:   "media-directory-create-window-title",
			ContentTpl: "media-directory-create-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		AddDirectoryData: AddDirectoryData{
			Lang:           lang,
			Directories:    shared.ListDirectories(w, mediaDir, true),
			SelectionField: "parent",
			SelectionID:    "media-directory-add",
			Selected:       strings.TrimSuffix(strings.TrimSpace(r.URL.Query().Get("path")), "/"),
		},
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
