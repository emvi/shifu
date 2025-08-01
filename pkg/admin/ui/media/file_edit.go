package media

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

// EditFileData is the data for the file editing form.
type EditFileData struct {
	Lang    string
	Path    string
	Name    string
	NewName string
	Errors  map[string]string
}

// EditFile renames a file.
func EditFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	name := strings.TrimSpace(r.URL.Query().Get("name"))
	fullPath := filepath.Join(getDirectoryPath(path), name)

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		newName := strings.TrimSpace(r.FormValue("name"))
		newPath := filepath.Join(getDirectoryPath(path), newName)
		errs := make(map[string]string)

		if newName == "" {
			errs["name"] = "the name is required"
		} else if !isValidFileName(newName) {
			errs["name"] = "the name contains invalid characters"
		} else if info, _ := os.Stat(newPath); info != nil {
			errs["name"] = "the file already exists"
		} else {
			if err := os.Rename(fullPath, newPath); err != nil {
				slog.Error("Error while renaming file", "error", err)
				errs["name"] = "error renaming file"
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "media-file-edit-form.html", EditFileData{
				Lang:    tpl.GetUILanguage(r),
				Path:    path,
				Name:    name,
				NewName: newName,
				Errors:  errs,
			})
			return
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "media-files.html", struct {
			Lang            string
			Path            string
			Selection       bool
			SelectionTarget string
			SelectionField  SelectionField
			Files           []File
		}{
			Lang:  tpl.GetUILanguage(r),
			Path:  path,
			Files: listFiles(path),
		})
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media-file-edit.html", struct {
		WindowOptions ui.WindowOptions
		EditFileData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-edit",
			TitleTpl:   "media-file-edit-window-title",
			ContentTpl: "media-file-edit-window-content",
			Overlay:    true,
			MinWidth:   400,
			Lang:       lang,
		},
		EditFileData: EditFileData{
			Lang:    lang,
			Path:    path,
			Name:    name,
			NewName: name,
		},
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
