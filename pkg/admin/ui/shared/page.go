package shared

import (
	"encoding/json"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"os"
)

// LoadPage loads a page for the given path and parses it into a cms.Content object.
func LoadPage(path string) (*cms.Content, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		slog.Error("Error reading content file", "error", err)
		return nil, err
	}

	var page cms.Content

	if err := json.Unmarshal(content, &page); err != nil {
		slog.Error("Error parsing content file", "error", err)
		return nil, err
	}

	page.File = path
	return &page, nil
}
