package content

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
)

// loadDocument loads the page or reference for given file path.
// References use a virtual page to edit them without updating an actual page.
func loadDocument(r *http.Request, path, language string) (*cms.Content, error) {
	page, err := shared.LoadPage(r, path, language)

	if err != nil || page.Tpl == "" {
		return page, err
	}

	languages := getSiteLanguages()

	for l := range page.Copy {
		languages[l] = ""
	}

	if len(languages) == 0 {
		languages["en"] = ""
	}

	ref := &cms.Content{
		Path: languages,
		Content: map[string][]cms.Content{
			"content": {{Ref: strings.ToLower(getReferenceName(filepath.Base(path)))}},
		},
	}
	shared.SetLanguage(r, ref, language)
	return ref, nil
}

func isVirtualDocument(page *cms.Content) bool {
	return page.File == ""
}

func getSiteLanguages() map[string]string {
	languages := make(map[string]string)
	dir := filepath.Join(cfg.Get().BaseDir, contentDir)

	if err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".json" {
			return nil
		}

		// the language is set explicitly, so no request is required
		page, err := shared.LoadPage(nil, path, "en")

		if err != nil {
			return nil
		}

		for l := range page.Path {
			languages[l] = ""
		}

		return nil
	}); err != nil {
		slog.Error("Error reading website languages", "error", err)
	}

	return languages
}
