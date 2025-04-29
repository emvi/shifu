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
	"strings"
)

// UploadFiles uploads one or more files.
func UploadFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		errs := make(map[string]string)

		// 100 MB
		if err := r.ParseMultipartForm(100_000_000); err != nil {
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

		if len(errs) == 0 {
			for _, file := range files {
				p := filepath.Join(getDirectoryPath(path), file.Filename)
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
				Path   string
				Errors map[string]string
			}{
				Path:   path,
				Errors: errs,
			})
			return
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "media-files.html", struct {
			Path  string
			Files []File
		}{
			Path:  path,
			Files: listFiles(path),
		})
		return
	}

	tpl.Get().Execute(w, "media-file-upload.html", struct {
		WindowOptions ui.WindowOptions
		Path          string
		Errors        map[string]string
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-upload",
			TitleTpl:   "media-file-upload-window-title",
			ContentTpl: "media-file-upload-window-content",
			Overlay:    true,
			MinWidth:   460,
		},
		Path: path,
	})
}
