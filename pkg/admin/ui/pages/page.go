package pages

import (
	"encoding/json"
	iso6391 "github.com/emvi/iso-639-1"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/cms"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Page renders the page details.
func Page(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getPagePath(path)
	page, err := loadPage(fullPath)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sitemapPriority, _ := strconv.ParseFloat(page.Sitemap.Priority, 64)
	tpl.Get().Execute(w, "pages-page-save-form.html", SavePageData{
		Name:      strings.TrimSuffix(filepath.Base(path), ".json"),
		PagePath:  page.Path,
		Cache:     page.DisableCache,
		Sitemap:   sitemapPriority,
		Handler:   page.Handler,
		Path:      path,
		Header:    page.Header,
		Languages: iso6391.Languages,
	})
}

func loadPage(path string) (*cms.Content, error) {
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

	return &page, nil
}
