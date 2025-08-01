package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxSize = 100_000_000 // 100 MB
)

// UploadFiles uploads one or more files.
func UploadFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		overwrite := strings.ToLower(r.FormValue("overwrite")) == "on"
		errs := make(map[string]string)

		if err := r.ParseMultipartForm(maxSize); err != nil {
			errs["files"] = "files exceed the maximum size"
		}

		var files []*multipart.FileHeader

		if r.MultipartForm == nil {
			errs["files"] = "multipart/form-data required"
		} else {
			files = r.MultipartForm.File["files"]

			if len(files) == 0 {
				errs["files"] = "no files selected"
			}
		}

		existingFiles := make([]string, 0)

		if len(errs) == 0 {
			for _, file := range files {
				filename := sanitizeFilename(path, file.Filename, overwrite)
				p := filepath.Join(getDirectoryPath(path), filename)

				if !overwrite {
					if _, err := os.Stat(p); !os.IsNotExist(err) {
						errs["files"] = "file already exists"
						existingFiles = append(existingFiles, filename)
						continue
					}
				}

				f, err := file.Open()

				if err != nil {
					slog.Error("Error opening file to upload", "error", err, "path", p)
					errs["files"] = "error uploading file"
					break
				}

				data, err := io.ReadAll(f)

				if err != nil {
					slog.Error("Error reading file to upload", "error", err, "path", p)
					_ = f.Close()
					errs["files"] = "error uploading file"
					break
				}

				if err := os.WriteFile(p, data, 0644); err != nil {
					slog.Error("Error writing uploaded file", "error", err, "path", p)
					_ = f.Close()
					errs["files"] = "error uploading file"
					break
				}

				_ = f.Close()
			}
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "media-file-upload-form.html", struct {
				Lang          string
				Path          string
				Overwrite     bool
				Errors        map[string]string
				ExistingFiles []string
			}{
				Lang:          tpl.GetUILanguage(r),
				Path:          path,
				Overwrite:     overwrite,
				Errors:        errs,
				ExistingFiles: existingFiles,
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
	tpl.Get().Execute(w, "media-file-upload.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Overwrite     bool
		Errors        map[string]string
		ExistingFiles []string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-upload",
			TitleTpl:   "media-file-upload-window-title",
			ContentTpl: "media-file-upload-window-content",
			Overlay:    true,
			MinWidth:   460,
			Lang:       lang,
		},
		Lang: lang,
		Path: path,
	})
}

func sanitizeFilename(path, filename string, overwrite bool) string {
	validChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = validChars.ReplaceAllString(filename, "_")
	filename = strings.Trim(filename, "._")

	if filename == "" {
		filename = "untitled"
	}

	if !overwrite {
		_, err := os.Stat(filepath.Join(getDirectoryPath(path), filename))

		for !os.IsNotExist(err) {
			filename = "_" + filename
			_, err = os.Stat(filepath.Join(getDirectoryPath(path), filename))
		}
	}

	return filename
}
