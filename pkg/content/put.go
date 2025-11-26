package content

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/cfg"
)

// Put uploads a content file.
func Put(path string, content []byte) error {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "content")
	dir := cfg.Get().BaseDir

	if err := os.MkdirAll(filepath.Join(dir, "content", filepath.Dir(path)), 0744); err != nil {
		slog.Error("Error creating content directory", "error", err)
		return errors.New("error creating content directory")
	}

	if err := os.WriteFile(filepath.Join(dir, "content", path), content, 0644); err != nil {
		slog.Error("Error writing content file", "error", err)
		return errors.New("error writing file")
	}

	return nil
}
