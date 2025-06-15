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

// SavePage saves a page to the given path.
func SavePage(page *cms.Content, path string) error {
	pageJson, err := json.Marshal(page)

	if err != nil {
		slog.Error("Error marshalling page data", "error", err)
		return err
	}

	if err := os.WriteFile(path, pageJson, 0644); err != nil {
		slog.Error("Error writing page data", "error", err)
		return err
	}

	return nil
}
