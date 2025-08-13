package pages

import (
	iso6391 "github.com/emvi/iso-639-1"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/middleware"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// Page renders the page details.
func Page(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getPagePath(path)
	page, err := shared.LoadPage(r, fullPath, "")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sitemapPriority, _ := strconv.ParseFloat(page.Sitemap.Priority, 64)
	tpl.Get().Execute(w, "pages-page-save-form.html", SavePageData{
		Admin:     middleware.IsAdmin(r),
		Lang:      tpl.GetUILanguage(r),
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
