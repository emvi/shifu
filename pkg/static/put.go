package static

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
	path = strings.TrimPrefix(path, "static")
	dir := cfg.Get().BaseDir

	if err := os.MkdirAll(filepath.Join(dir, "static", filepath.Dir(path)), 0744); err != nil {
		slog.Error("Error creating static directory", "error", err)
		return errors.New("error creating static directory")
	}

	if err := os.WriteFile(filepath.Join(dir, "static", path), content, 0644); err != nil {
		slog.Error("Error writing static file", "error", err)
		return errors.New("error writing file")
	}

	return nil
}
