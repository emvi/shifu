package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	mediaDir = "static/media"
)

// Directory is a media directory.
type Directory struct {
	Name     string
	Path     string
	Children []Directory

	level int
}

// Media renders the media management dialog.
func Media(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "media.html", struct {
		WindowOptions ui.WindowOptions
		Directories   []Directory
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-media",
			TitleTpl:   "media-window-title",
			ContentTpl: "media-window-content",
			MinWidth:   800,
		},
		Directories: listDirectories(w),
	})
}

func listDirectories(w http.ResponseWriter) []Directory {
	dir := filepath.Join(cfg.Get().BaseDir, mediaDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating media directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	dirs := make([]Directory, 0)
	var last *Directory

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		path = strings.TrimPrefix(path, dir)

		if d.IsDir() && path != "" {
			level := strings.Count(path, "/")

			if last == nil || level <= last.level {
				dirs = append(dirs, Directory{
					Name:     d.Name(),
					Path:     path,
					Children: make([]Directory, 0),
					level:    level,
				})
				last = &dirs[len(dirs)-1]
			} else {
				last.Children = append(last.Children, Directory{
					Name:     d.Name(),
					Path:     path,
					Children: make([]Directory, 0),
					level:    level,
				})
				last = &last.Children[len(last.Children)-1]
			}
		}

		return err
	}); err != nil {
		slog.Error("Error reading media directory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	return dirs
}
