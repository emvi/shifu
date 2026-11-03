package cms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"github.com/stretchr/testify/assert"
)

func TestCMS(t *testing.T) {
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   "../../demo",
		HotReload: true,
		FuncMap:   defaultFuncMap,
		Source:    source.NewFileSystem("../../demo", 1),
		Sitemap:   sitemap.New(),
	})
	n := 100
	var wg sync.WaitGroup
	wg.Add(n)

	for range n {
		go func() {
			c.Update()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			c.Serve(w, r)
			c.LastUpdate()
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkCMS(b *testing.B) {
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   "../../demo",
		HotReload: true,
		FuncMap:   defaultFuncMap,
		Source:    source.NewFileSystem("../../demo", 1),
		Sitemap:   sitemap.New(),
	})
	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			c.Serve(w, r)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestLLMs(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"tpl/head.html": `<!DOCTYPE html><html><head><title>{{ copy .Page .Content "title" }}</title>` +
			`<meta name="description" content="{{ copy .Page .Content "description" }}" />{{ llms .Page }}</head><body><nav>Menu</nav><main>`,
		"tpl/text.html": `<p>{{ copy .Page .Content "text" }}{{ with index .Args "slug" }} {{ . }}{{ end }}</p>`,
		"tpl/end.html":  `</main></body></html>`,
		"content/home.json": `{"path":{"en":"/","de":"/de"},"content":{"content":[
			{"tpl":"head","copy":{"en":{"title":"Home","description":"Welcome."},"de":{"title":"Start","description":"Willkommen."}}},
			{"tpl":"text","copy":{"en":{"text":"Hello"},"de":{"text":"Hallo"}}},{"tpl":"end"}]}}`,
		"content/about.json": `{"path":{"en":"/about"},"content":{"content":[
			{"tpl":"head","copy":{"en":{"title":"About"}}},{"tpl":"text","copy":{"en":{"text":"About us"}}},{"tpl":"end"}]}}`,
		"content/secret.json": `{"path":{"en":"/secret"},"llms":{"exclude":true},"content":{"content":[
			{"tpl":"head","copy":{"en":{"title":"Secret"}}},{"tpl":"end"}]}}`,
		"content/blog.json": `{"path":{"en":"/blog/{slug:[a-z]+}"},"content":{"content":[
			{"tpl":"head","copy":{"en":{"title":"Blog"}}},{"tpl":"text","copy":{"en":{"text":"Article"}}},{"tpl":"end"}]}}`,
		"content/404.json": `{"path":{"en":"/404"},"content":{"content":[{"tpl":"head","copy":{"en":{"title":"Not found"}}},{"tpl":"end"}]}}`,
	}

	for name, content := range files {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0755))
		assert.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
	}

	llms := cfg.Get().LLMs
	hostname := cfg.Get().Server.Hostname
	cfg.Get().LLMs = cfg.LLMs{Enabled: true, Description: "A test website.", Selector: "main"}
	cfg.Get().Server.Hostname = "https://example.com"
	defer func() {
		cfg.Get().LLMs = llms
		cfg.Get().Server.Hostname = hostname
	}()
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   dir,
		HotReload: true,
		FuncMap:   Merge(),
		Source:    source.NewFileSystem(dir, 1),
		Sitemap:   sitemap.New(),
	})
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)

		if path == "/llms.txt" {
			c.ServeLLMs(w, r)
		} else {
			c.Serve(w, r)
		}

		return w
	}

	// HTML links to the Markdown version
	w := get("/about")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `<link rel="alternate" type="text/markdown" href="https://example.com/about.md" />`)

	w = get("/index.md")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/markdown; charset=utf-8", w.Header().Get("Content-Type"))
	assert.Equal(t, `<https://example.com/>; rel="canonical"`, w.Header().Get("Link"))
	assert.Equal(t, "# Home\n\n> Welcome.\n\nHello\n", w.Body.String())

	w = get("/de.md")
	assert.Equal(t, "# Start\n\n> Willkommen.\n\nHallo\n", w.Body.String())

	// dynamic routes
	w = get("/blog/hello.md")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "# Blog\n\nArticle hello\n", w.Body.String())

	// excluded and unknown pages
	assert.Equal(t, http.StatusNotFound, get("/secret.md").Code)
	assert.Equal(t, http.StatusNotFound, get("/unknown.md").Code)
	assert.NotContains(t, get("/secret").Body.String(), "text/markdown")

	w = get("/llms.txt")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `# Home

> A test website.

## de

- [Start](https://example.com/de.md): Willkommen.

## en

- [Home](https://example.com/index.md): Welcome.
- [About](https://example.com/about.md)
`, w.Body.String())
}
