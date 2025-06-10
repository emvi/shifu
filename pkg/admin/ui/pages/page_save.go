package pages

import (
	"fmt"
	"github.com/emvi/iso-639-1"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

// SavePageData is the data for the page form.
type SavePageData struct {
	Name      string
	PagePath  map[string]string
	Cache     bool
	Sitemap   float64
	Handler   string
	Path      string
	Header    map[string]string
	Errors    map[string]string
	New       bool
	Saved     bool
	Languages map[string]iso6391.Language
}

// SavePage creates or updates a page.
func SavePage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))

	if r.Method == http.MethodPost {
		overwrite := strings.HasSuffix(path, ".json")
		name := strings.TrimSpace(r.FormValue("name"))
		cache := strings.ToLower(strings.TrimSpace(r.FormValue("cache"))) == "on"
		sitemap := strings.TrimSpace(r.FormValue("sitemap"))
		handler := strings.TrimSpace(r.FormValue("handler"))
		languages := make([]string, 0)
		paths := make([]string, 0)
		headerKeys := make([]string, 0)
		headerValues := make([]string, 0)
		pagePath := make(map[string]string)
		header := make(map[string]string)
		errs := make(map[string]string)

		for k, v := range r.Form {
			if k == "language[]" {
				languages = v
			} else if k == "path[]" {
				paths = v
			} else if k == "header[]" {
				headerKeys = v
			} else if k == "header_value[]" {
				headerValues = v
			}
		}

		if len(languages) != len(paths) {
			errs["path"] = "the selected languages do not match the number of paths"
		} else {
			for i, l := range languages {
				l = strings.TrimSpace(l)

				if l != "" {
					pagePath[l] = strings.TrimSpace(paths[i])
				}
			}
		}

		if len(headerKeys) != len(headerValues) {
			errs["header"] = "the headers do not match the number of values"
		} else {
			for i, k := range headerKeys {
				k = strings.TrimSpace(k)

				if k != "" {
					header[k] = strings.TrimSpace(headerValues[i])
				}
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
		} else if !overwrite && pageExists(name) {
			errs["name"] = "the page already exists"
		}

		sitemapFloat, err := strconv.ParseFloat(sitemap, 64)

		if err != nil {
			errs["sitemap"] = "the number is invalid"
		} else if sitemapFloat < 0 {
			errs["sitemap"] = "the sitemap must be a positive number"
		} else if sitemapFloat > 1 {
			errs["sitemap"] = "the sitemap must be less than 1"
		}

		outPath := getPagePath(filepath.Join(path, name+".json"))

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			tpl.Get().Execute(w, "pages-page-save-form.html", SavePageData{
				Name:      name,
				PagePath:  pagePath,
				Cache:     cache,
				Sitemap:   sitemapFloat,
				Handler:   handler,
				Path:      path,
				Header:    header,
				Errors:    errs,
				New:       true,
				Languages: iso6391.Languages,
			})
			return
		} else {
			var page *cms.Content

			if overwrite {
				outPath = getPagePath(path)
				page, err = shared.LoadPage(outPath)

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				page.DisableCache = cache
				page.Path = pagePath
				page.Sitemap = cms.Sitemap{
					Priority: fmt.Sprintf("%f", sitemapFloat),
				}
				page.Header = header
				page.Handler = handler
			} else {
				page = &cms.Content{
					DisableCache: cache,
					Path:         pagePath,
					Sitemap: cms.Sitemap{
						Priority: fmt.Sprintf("%f", sitemapFloat),
					},
					Header:  header,
					Handler: handler,
				}
			}

			if err := shared.SavePage(page, outPath); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// rename the file if name changed
			if overwrite && getPageName(filepath.Base(outPath)) != name {
				path = filepath.Join(filepath.Dir(path), name+".json")

				if err := os.Rename(outPath, getPagePath(path)); err != nil {
					slog.Error("Error renaming page", "error", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		tpl.Get().Execute(w, "pages-tree-save.html", struct {
			SavePageData
			WindowOptions ui.WindowOptions
			Entries       []Entry
		}{
			WindowOptions: ui.WindowOptions{
				ID:         "shifu-pages-page-save",
				TitleTpl:   "pages-page-save-window-title",
				ContentTpl: "pages-page-save-window-content",
				Overlay:    true,
				MinWidth:   520,
			},
			SavePageData: SavePageData{
				Name:      name,
				PagePath:  pagePath,
				Cache:     cache,
				Sitemap:   sitemapFloat,
				Handler:   handler,
				Path:      path,
				Header:    header,
				Saved:     true,
				Languages: iso6391.Languages,
			},
			Entries: listEntries(w),
		})
		return
	}

	tpl.Get().Execute(w, "pages-page-save.html", struct {
		SavePageData
		WindowOptions ui.WindowOptions
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-save",
			TitleTpl:   "pages-page-save-window-title",
			ContentTpl: "pages-page-save-window-content",
			Overlay:    true,
			MinWidth:   520,
		},
		SavePageData: SavePageData{
			PagePath:  map[string]string{"de": "/"},
			Path:      path,
			Sitemap:   1,
			New:       true,
			Languages: iso6391.Languages,
		},
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
		if err != nil {
			return err
		}

		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".json" {
			files = append(files, entry.Name())
		}

		return nil
	}); err != nil {
		slog.Error("Error reading content directory", "error", err)
		return false
	}

	for _, file := range files {
		if file == name {
			return true
		}
	}

	return false
}
