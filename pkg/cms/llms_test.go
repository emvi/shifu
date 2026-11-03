package cms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"github.com/stretchr/testify/assert"
)

func TestMarkdownPath(t *testing.T) {
	assert.Equal(t, "/index.md", markdownPath("/"))
	assert.Equal(t, "/about.md", markdownPath("/about"))
	assert.Equal(t, "/about.md", markdownPath("/about/"))
	assert.Equal(t, "/de.md", markdownPath("/de"))
	assert.Equal(t, "/", pagePathFromMarkdown("/index.md"))
	assert.Equal(t, "/about", pagePathFromMarkdown("/about.md"))
	assert.Equal(t, "/de", pagePathFromMarkdown("/de.md"))
}

func TestHTMLToMarkdown(t *testing.T) {
	page, err := htmlToMarkdown([]byte(`<!DOCTYPE html><html><head>
		<title>About us</title>
		<meta name="description" content="Who we are." />
	</head><body>
		<header><nav><a href="/">Home</a></nav></header>
		<main>
			<h2>Team</h2>
			<p>We are <strong>three</strong> people. <a href="/contact">Contact</a></p>
			<script>alert(1)</script>
			<slot name="content.1/content"></slot>
			<form><input name="q" /><button>Search</button></form>
		</main>
		<footer>Footer</footer>
	</body></html>`), "main", "https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "About us", page.Title)
	assert.Equal(t, "Who we are.", page.Description)
	assert.Equal(t, "# About us\n\n> Who we are.\n\n## Team\n\nWe are **three** people. [Contact](https://example.com/contact)\n", string(page.Markdown))

	// falls back to the body, doesn't duplicate an existing h1
	page, err = htmlToMarkdown([]byte(`<html><head><title>T</title></head><body><h1>Hello</h1><nav>x</nav></body></html>`), "main", "")
	assert.NoError(t, err)
	assert.Equal(t, "# Hello\n", string(page.Markdown))

	// without <main>, the page header and footer are dropped, but not the ones of an article
	page, err = htmlToMarkdown([]byte(`<body><header>Logo</header><article><header><h2>Post</h2></header>Text</article><footer>Footer</footer></body>`), "main", "")
	assert.NoError(t, err)
	assert.Equal(t, "## Post\n\nText\n", string(page.Markdown))

	// id and class selectors
	page, err = htmlToMarkdown([]byte(`<body><div id="a">A</div><div class="x b">B</div></body>`), ".b", "")
	assert.NoError(t, err)
	assert.Equal(t, "B\n", string(page.Markdown))
	page, err = htmlToMarkdown([]byte(`<body><div id="a">A</div><div class="x b">B</div></body>`), "#a", "")
	assert.NoError(t, err)
	assert.Equal(t, "A\n", string(page.Markdown))
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
