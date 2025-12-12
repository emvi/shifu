package pages

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	iso6391 "github.com/emvi/iso-639-1"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/middleware"
)

// Page renders the page details.
func Page(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(strings.TrimSpace(r.URL.Query().Get("path")), "/")
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, "")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sitemapPriority, _ := strconv.ParseFloat(page.Sitemap.Priority, 64)
	tpl.Get().Execute(w, "pages-page-save-form.html", SavePageData{
		Admin:          middleware.IsAdmin(r),
		Lang:           tpl.GetUILanguage(r),
		Directories:    shared.ListDirectories(w, contentDir, true),
		SelectionField: "parent",
		SelectionID:    "page",
		Name:           strings.TrimSuffix(filepath.Base(path), ".json"),
		PagePath:       page.Path,
		Cache:          page.DisableCache,
		Sitemap:        sitemapPriority,
		Handler:        page.Handler,
		Selected:       getDirectory(filepath.Dir(path)),
		Path:           path,
		Header:         page.Header,
		Languages:      iso6391.Languages,
	})
}

func getDirectory(path string) string {
	if path == "/" {
		return ""
	}

	return path
}
