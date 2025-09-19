package media

import (
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
)

const (
	maxSize = 100_000_000 // 100 MB
)

// UploadFiles uploads one or more files.
func UploadFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	existingFiles := make([]string, 0)

	if r.Method == http.MethodPost {
		reader, err := r.MultipartReader()
		overwrite := strings.ToLower(r.FormValue("overwrite")) == "on"
		errs := make(map[string]string)

		if err != nil {
			slog.Error("Error creating multipart reader", "error", err)
			errs["files"] = "error reading multipart request"
		} else {
			for {
				part, err := reader.NextPart()

				if err == io.EOF {
					break
				}

				if err != nil {
					errs["files"] = "error reading multipart request part"
					break
				}

				if err := uploadFile(part, path, overwrite, errs, existingFiles); err != nil {
					errs["files"] = "error uploading file"
					break
				}

				if err := part.Close(); err != nil {
					slog.Warn("Error closing multipart file", "error", err)
				}
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

func uploadFile(part *multipart.Part, path string, overwrite bool, errs map[string]string, existingFiles []string) error {
	filename := part.FileName()

	if filename == "" {
		return nil
	}

	filename = sanitizeFilename(path, filename, overwrite)
	p := filepath.Join(getDirectoryPath(path), filename)

	if !overwrite {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			errs["files"] = "file already exists"
			existingFiles = append(existingFiles, filename)
			return nil
		}
	}

	f, err := os.Create(p)

	if err != nil {
		slog.Error("Error opening file to upload", "error", err, "path", p)
		errs["files"] = "error uploading file"
		return err
	}

	written, err := io.CopyN(f, part, maxSize+1)

	if err != nil && err != io.EOF {
		_ = f.Close()
		_ = os.Remove(p)
		return err
	}

	if written > maxSize {
		_ = f.Close()
		_ = os.Remove(p)
		return errors.New("max file upload size exceeded")
	}

	_ = f.Close()
	return nil
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
