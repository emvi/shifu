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
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

const (
	maxSize      = 100_000_000   // 100 MB per file
	maxDrainSize = 1_000_000_000 // 1 GB
)

var (
	errFileTooLarge = errors.New("the file exceeds the maximum upload size")

	validFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
)

// UploadFiles uploads one or more files.
func UploadFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	existingFiles := make([]string, 0)

	if r.Method == http.MethodPost {
		reader, err := r.MultipartReader()
		overwrite, unpack, updateTree := false, false, false
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

				// the options are sent before the files, because they come first in the form
				if part.FileName() == "" {
					switch part.FormName() {
					case "overwrite":
						overwrite = readCheckbox(part)
					case "unzip":
						unpack = readCheckbox(part)
					}
				} else if unpack && isZip(part.FileName()) {
					existing, dirs, err := unzipUpload(part, getDirectoryPath(path), overwrite)
					updateTree = updateTree || dirs
					existingFiles = append(existingFiles, existing...)

					if err != nil {
						slog.Error("Error extracting zip file", "error", err, "file", part.FileName())

						if errors.Is(err, errFileTooLarge) || errors.Is(err, errUnzipSize) || errors.Is(err, errUnzipFiles) {
							errs["files"] = err.Error()
						} else {
							errs["files"] = "error extracting zip file"
						}

						break
					}
				} else {
					existing, err := uploadFile(part, path, overwrite)

					if err != nil {
						slog.Error("Error uploading file", "error", err, "file", part.FileName())

						if errors.Is(err, errFileTooLarge) {
							errs["files"] = err.Error()
						} else {
							errs["files"] = "error uploading file"
						}

						break
					}

					if existing != "" {
						existingFiles = append(existingFiles, existing)
					}
				}

				if err := part.Close(); err != nil {
					slog.Warn("Error closing multipart file", "error", err)
				}
			}
		}

		// drain the connection before replying to the client
		if len(errs) > 0 {
			if _, err := io.Copy(io.Discard, io.LimitReader(r.Body, maxDrainSize)); err != nil {
				slog.Debug("Error reading the rest of the upload request", "error", err)
			}
		}

		// existing files are skipped unless they should be overwritten, the other files are uploaded anyway
		if len(existingFiles) > 0 && errs["files"] == "" {
			errs["files"] = "file already exists"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "media-file-upload-form.html", struct {
				Lang          string
				Path          string
				Overwrite     bool
				Unzip         bool
				Errors        map[string]string
				ExistingFiles []string
				MaxSize       int64
			}{
				Lang:          tpl.GetUILanguage(r),
				Path:          path,
				Overwrite:     overwrite,
				Unzip:         unpack,
				Errors:        errs,
				ExistingFiles: existingFiles,
				MaxSize:       maxSize,
			})

			if updateTree {
				renderTreeUpdate(w, r)
			}

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

		if updateTree {
			renderTreeUpdate(w, r)
		}

		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "media-file-upload.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Path          string
		Overwrite     bool
		Unzip         bool
		Errors        map[string]string
		ExistingFiles []string
		MaxSize       int64
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media-file-upload",
			TitleTpl:   "media-file-upload-window-title",
			ContentTpl: "media-file-upload-window-content",
			Overlay:    true,
			MinWidth:   460,
			Lang:       lang,
		},
		Lang:    lang,
		Path:    path,
		MaxSize: maxSize,
	})
}

func uploadFile(part *multipart.Part, path string, overwrite bool) (string, error) {
	filename := cleanFilename(part.FileName())
	p := filepath.Join(getDirectoryPath(path), filename)

	if !overwrite && fileExists(p) {
		return filename, nil
	}

	f, err := os.Create(p)

	if err != nil {
		return "", err
	}

	written, err := io.CopyN(f, part, maxSize+1)
	_ = f.Close()

	if err != nil && err != io.EOF {
		_ = os.Remove(p)
		return "", err
	}

	if written > maxSize {
		_ = os.Remove(p)
		return "", errFileTooLarge
	}

	return "", nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func cleanFilename(filename string) string {
	filename = validFilenameChars.ReplaceAllString(filename, "_")
	filename = strings.Trim(filename, "._")

	if filename == "" {
		filename = "untitled"
	}

	return filename
}

func renderTreeUpdate(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "media-tree-oob.html", struct {
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
}

func readCheckbox(part *multipart.Part) bool {
	value, err := io.ReadAll(io.LimitReader(part, 16))
	return err == nil && strings.ToLower(strings.TrimSpace(string(value))) == "on"
}
