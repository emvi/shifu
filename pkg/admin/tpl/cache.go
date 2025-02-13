package tpl

import (
	"github.com/emvi/shifu/static"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
)

// Cache caches HTML templates.
type Cache struct {
	temp template.Template
}

// NewCache creates a new HTML cache.
func NewCache() *Cache {
	cache := &Cache{}

	if err := cache.loadTemplate(); err != nil {
		slog.Error("Error loading admin template files", "error", err)
	}

	return cache
}

// Execute runs the HTML template for given name and sends it to the client.
func (cache *Cache) Execute(w http.ResponseWriter, name string, data any) {
	if err := cache.temp.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Error executing admin template", "error", err, "name", name)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (cache *Cache) loadTemplate() error {
	t := template.New("").Funcs(funcMap)

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
