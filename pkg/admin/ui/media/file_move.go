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

// MoveFile moves a file to a different directory.
func MoveFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	names := r.URL.Query()["name[]"]

	if len(names) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for i, name := range names {
		names[i] = strings.TrimSpace(name)
		fullPath := filepath.Join(getDirectoryPath(path), names[i])

		if path == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if r.Method == http.MethodPost {
		newPath := getDirectoryPath(strings.TrimSpace(r.FormValue("path")))
		errs := make(map[string]string)

		if newPath == "" {
			errs["path"] = "the path is required"
		} else if _, err := os.Stat(newPath); os.IsNotExist(err) {
			errs["path"] = "the directory does not exist"
		} else {
			for _, name := range names {
				newFilePath := filepath.Join(newPath, name)

				if info, _ := os.Stat(newFilePath); info != nil {
					errs["path"] = "the file already exists"
					break
				}
			}

			if len(errs) == 0 {
				for _, name := range names {
					fullPath := filepath.Join(getDirectoryPath(path), name)
					newFilePath := filepath.Join(newPath, name)
					slog.Info("Moving file", "from", fullPath, "to", newFilePath)

					if err := os.Rename(fullPath, newFilePath); err != nil {
						slog.Error("Error while moving file", "error", err)
						errs["path"] = "error moving file"
						break
					}
				}
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "media-file-move-form.html", struct {
				Lang        string
				Directories []Directory
				Interactive bool
				Path        string
				Name        []string
				Errors      map[string]string
			}{
				Lang:        tpl.GetUILanguage(r),
				Directories: listDirectories(w),
				Interactive: false,
				Path:        path,
				Name:        names,
				Errors:      errs,
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
	tpl.Get().Execute(w, "media-file-move.html", struct {
		WindowOptions   ui.WindowOptions
		Lang            string
		Directories     []Directory
		Interactive     bool
		Selection       bool
		SelectionTarget string
		SelectionField  SelectionField
		Path            string
		Name            []string
		Errors          map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-move",
			TitleTpl:   "media-file-move-window-title",
			ContentTpl: "media-file-move-window-content",
			Overlay:    true,
			MinWidth:   400,
			Lang:       lang,
		},
		Lang:        lang,
		Directories: listDirectories(w),
		Interactive: false,
		Path:        path,
		Name:        names,
	})
}
