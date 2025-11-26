package content

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/cfg"
)

// Get returns a content file.
func Get(path string) ([]byte, error) {
	path = strings.TrimSpace(strings.TrimPrefix(path, "/"))

	if path == "" {
		return nil, errors.New("path empty")
	}

	if !strings.HasPrefix(path, "content") {
		return nil, errors.New("path does not start with content")
	}

	file, err := os.ReadFile(filepath.Join(cfg.Get().BaseDir, path))

	if err != nil {
		return nil, errors.New("file not found")
	}

	return file, nil
}
