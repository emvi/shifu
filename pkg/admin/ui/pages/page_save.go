package pages

import (
	"fmt"
	"github.com/emvi/iso-639-1"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// SavePage creates or updates a page.
func SavePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		cache := strings.ToLower(strings.TrimSpace(r.FormValue("cache"))) == "on"
		sitemap := strings.TrimSpace(r.FormValue("sitemap"))
		handler := strings.TrimSpace(r.FormValue("handler"))
		languages := make([]string, 0)
		paths := make([]string, 0)
		pagePath := make(map[string]string)
		errs := make(map[string]string)

		for k, v := range r.Form {
			if k == "language[]" {
				languages = v
			} else if k == "path[]" {
				paths = v
			}
		}

		if len(languages) != len(paths) {
			errs["path"] = "the selected languages do not match the number of paths"
		} else {
			for i, l := range languages {
				pagePath[l] = strings.TrimSpace(paths[i])
			}
		}

		if len(pagePath) == 0 {
			errs["path"] = "no path selected"
		} else {
			for k, v := range pagePath {
				if iso6391.FromCode(k).Code == "" {
					errs["path"] = "the language code does not exist"
				} else if v == "" {
					errs["path"] = "the path must be set"
				} else if v[0] != '/' {
					errs["path"] = "the first character of a path must be a forward slash"
				} else {
					for l, p := range pagePath {
						if k != l && v == p {
							errs["path"] = "the path already exists"
							break
						}
					}
				}
			}
		}

		if name == "" {
			errs["name"] = "the name is required"
		} else if !isValidPageName(name) {
			errs["name"] = "the name contains invalid characters"
		} else if pageExists(name) {
			errs["name"] = "the page already exists"
		}

		sitemapf, err := strconv.ParseFloat(sitemap, 64)

		if err != nil {
			errs["sitemap"] = "the number is invalid"
		} else if sitemapf < 0 {
			errs["sitemap"] = "the sitemap must be a positive number"
		} else if sitemapf > 1 {
			errs["sitemap"] = "the sitemap must be less than 1"
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "pages-page-save-form.html", struct {
				Name      string
				PagePath  map[string]string
				Cache     bool
				Sitemap   float64
				Handler   string
				Path      string
				Errors    map[string]string
				New       bool
				Languages map[string]iso6391.Language
			}{
				Name:      name,
				PagePath:  pagePath,
				Cache:     cache,
				Sitemap:   sitemapf,
				Handler:   handler,
				Path:      path,
				Errors:    errs,
				New:       true,
				Languages: iso6391.Languages,
			})
			return
		} else {
			// TODO create page
			//fullPath := getPagePath(filepath.Join(path, name))
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "pages-tree.html", struct {
			Entries []Entry
		}{
			Entries: listEntries(w),
		})
		return
	}

	tpl.Get().Execute(w, "pages-page-save.html", struct {
		WindowOptions ui.WindowOptions
		Name          string
		PagePath      map[string]string
		Cache         bool
		Sitemap       float64
		Handler       string
		Path          string
		Errors        map[string]string
		New           bool
		Languages     map[string]iso6391.Language
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-save",
			TitleTpl:   "pages-page-save-window-title",
			ContentTpl: "pages-page-save-window-content",
			Overlay:    true,
			MinWidth:   520,
		},
		PagePath:  map[string]string{"de": "/"},
		Path:      path,
		New:       true,
		Languages: iso6391.Languages,
	})
}

func isValidPageName(name string) bool {
	for _, c := range name {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' && c != '.' {
			return false
		}
	}

	return true
}

func pageExists(name string) bool {
	name = fmt.Sprintf("%s.json", name)
	files := make([]string, 0)

	if err := filepath.WalkDir(filepath.Join(cfg.Get().BaseDir, "content"), func(_ string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".json" {
			files = append(files, entry.Name())
		}

		return err
	}); err != nil {
		slog.Error("Error reading content directory", "error", err)
		return false
	}

	log.Println(files)

	for _, file := range files {
		if file == name {
			return true
		}
	}

	return false
}
