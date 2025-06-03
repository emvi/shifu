package cms

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Cache caches an HTML template directory.
type Cache struct {
	temp     template.Template
	funcMap  template.FuncMap
	dir      string
	disabled bool
	loaded   bool
	m        sync.RWMutex
}

// NewCache creates a new template cache for the given directory and function map.
// Disable is used to disable the cache for testing.
func NewCache(dir string, funcMap template.FuncMap, disable bool) *Cache {
	cache := &Cache{
		funcMap:  funcMap,
		dir:      dir,
		disabled: disable,
	}

	if err := cache.loadTemplate(); err != nil {
		slog.Error("Error loading template files from directory", "error", err, "directory", cache.dir)
	}

	return cache
}

// Execute executes the template for the given name. It logs errors and returns an error code if something goes wrong.
func (cache *Cache) Execute(w http.ResponseWriter, name string, data any) {
	if err := cache.Get().ExecuteTemplate(w, name, data); err != nil {
		slog.Error("Error executing template", "error", err, "name", name)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Render executes the template for the given name and returns the output.
func (cache *Cache) Render(name string, data any) ([]byte, error) {
	var buffer bytes.Buffer

	if err := cache.Get().ExecuteTemplate(&buffer, name, data); err != nil {
		slog.Error("Error rendering template", "error", err, "name", name)
		return nil, err
	}

	return buffer.Bytes(), nil
}

// Get returns the HTML template or loads it in case the cache is disabled, or it hasn't been loaded yet.
func (cache *Cache) Get() *template.Template {
	cache.m.RLock()
	load := cache.disabled || !cache.loaded
	cache.m.RUnlock()

	if load {
		if err := cache.loadTemplate(); err != nil {
			slog.Error("Error refreshing template files from directory", "error", err, "directory", cache.dir)
			panic(err)
		}
	}

	cache.m.RLock()
	defer cache.m.RUnlock()
	t := cache.temp
	return &t
}

func (cache *Cache) loadTemplate() error {
	cache.m.Lock()
	defer cache.m.Unlock()
	t := template.New("").Funcs(cache.funcMap)

	if err := filepath.Walk(cache.dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.Contains(path, ".html") {
			if _, err = t.ParseFiles(path); err != nil {
				return err
			}
		}

		return err
	}); err != nil {
		return err
	}

	cache.temp = *t
	cache.loaded = true
	return nil
}

// Enable enables caching.
func (cache *Cache) Enable() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.disabled = false
}

// Disable disables caching.
func (cache *Cache) Disable() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.disabled = true
}

// Clear clears the cache.
func (cache *Cache) Clear() {
	cache.m.Lock()
	defer cache.m.Unlock()
	cache.loaded = false
}
