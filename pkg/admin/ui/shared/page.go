package shared

import (
	"encoding/json"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// LoadPage loads a page for the given path and parses it into a cms.Content object.
func LoadPage(r *http.Request, path, language string) (*cms.Content, error) {
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

	// The language must set when adding or updating an element to response in the displayed language.
	// Otherwise, setting the language explicitly isn't required. Extracting it from the request is just a "fallback".
	if language == "" {
		languages := cms.GetAcceptedLanguages(r)

		for _, lang := range languages {
			if _, ok := page.Path[lang]; ok {
				page.Language = lang
				break
			}
		}

		if page.Language == "" {
			page.Language = "en"
		}
	} else {
		page.Language = language
	}

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

// GetLanguage returns the language query parameter.
func GetLanguage(r *http.Request) string {
	return strings.ToLower(strings.TrimSpace(r.URL.Query().Get("language")))
}
