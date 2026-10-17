package shared

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/emvi/shifu/pkg/cms"
)

// LoadPage loads a page for the given path and parses it into a cms.Content object.
func LoadPage(r *http.Request, path, language string) (*cms.Content, error) {
	content, err := os.ReadFile(path)

	if err != nil {
		slog.Error("Error reading content file", "path", path, "error", err)
		return nil, err
	}

	var page cms.Content

	if err := json.Unmarshal(content, &page); err != nil {
		slog.Error("Error parsing content file", "path", path, "error", err)
		return nil, err
	}

	page.File = path
	SetLanguage(r, &page, language)
	return &page, nil
}

// SetLanguage sets the language of the page.
// The language must set when adding or updating an element to response in the displayed language.
// Otherwise, setting the language explicitly isn't required. Extracting it from the request is just a "fallback".
func SetLanguage(r *http.Request, page *cms.Content, language string) {
	if language != "" {
		page.Language = language
		return
	}

	page.Language = ""

	for _, lang := range cms.GetAcceptedLanguages(r) {
		if _, ok := page.Path[lang]; ok {
			page.Language = lang
			return
		}
	}

	page.Language = "en"
}

// SavePage saves a page to the given path.
func SavePage(page *cms.Content, path string) error {
	pageJson, err := json.Marshal(page)

	if err != nil {
		slog.Error("Error marshalling page data", "path", path, "error", err)
		return err
	}

	if err := os.WriteFile(path, pageJson, 0644); err != nil {
		slog.Error("Error writing page data", "path", path, "error", err)
		return err
	}

	return nil
}
