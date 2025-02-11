package content

import (
	"github.com/emvi/shifu/pkg/cfg"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

// File is a content file.
type File struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
}

// List returns a list of all static files except for invisible files (starting with a dot).
func List() []File {
	dir := cfg.Get().BaseDir
	files := make([]File, 0)

	if err := filepath.WalkDir(filepath.Join(dir, "content"), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.ToLower(filepath.Ext(path)) == ".json" {
			info, e := d.Info()

			if e != nil {
				return e
			}

			files = append(files, File{
				Path:         strings.TrimPrefix(path, dir),
				Size:         info.Size(),
				LastModified: info.ModTime(),
			})
		}

		return err
	}); err != nil {
		slog.Error("Error listing content files", "error", err)
	}

	return files
}
