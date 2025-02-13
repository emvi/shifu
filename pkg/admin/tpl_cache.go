package admin

import (
	"github.com/emvi/shifu/static"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

type cache struct {
	temp template.Template
}

func newCache() *cache {
	cache := &cache{}

	if err := cache.loadTemplate(); err != nil {
		slog.Error("Error loading admin template files", "error", err)
	}

	return cache
}

func (cache *cache) execute(w http.ResponseWriter, name string, data any) {
	if err := cache.temp.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Error executing admin template", "error", err, "name", name)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cache *cache) loadTemplate() error {
	// TODO .Funcs(cache.funcMap)
	t := template.New("")

	if err := fs.WalkDir(static.AdminTpl, "admin/tpl", func(path string, info fs.DirEntry, err error) error {
		if !info.IsDir() && strings.Contains(path, ".html") {
			if _, err = t.ParseFS(static.AdminTpl, path); err != nil {
				return err
			}
		}

		return err
	}); err != nil {
		return err
	}

	cache.temp = *t
	return nil
}
