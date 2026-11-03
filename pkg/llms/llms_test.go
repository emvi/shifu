package llms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/stretchr/testify/assert"
)

func TestMarkdownPath(t *testing.T) {
	assert.Equal(t, "/index.md", MarkdownPath("/"))
	assert.Equal(t, "/about.md", MarkdownPath("/about"))
	assert.Equal(t, "/about.md", MarkdownPath("/about/"))
	assert.Equal(t, "/de.md", MarkdownPath("/de"))
	assert.Equal(t, "/", PagePath("/index.md"))
	assert.Equal(t, "/about", PagePath("/about.md"))
	assert.Equal(t, "/de", PagePath("/de.md"))
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

func TestLLMsCache(t *testing.T) {
	selector := cfg.Get().LLMs.Selector
	cfg.Get().LLMs.Selector = "main"
	defer func() {
		cfg.Get().LLMs.Selector = selector
	}()
	renders := 0
	l := New()
	l.SetRenderer(func(w http.ResponseWriter, r *http.Request, path string) (bool, bool) {
		if path == "/missing" {
			return false, false
		}

		renders++
		assert.True(t, IsMarkdownRequest(r))
		assert.Equal(t, path, r.URL.Path)
		_, _ = w.Write([]byte("<title>Page</title><main><p>Text</p></main>"))
		return true, path != "/nocache"
	})
	l.Update([]Page{{Language: "en", Path: "/"}, {Language: "en", Path: "/nocache"}})
	get := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		l.ServeMarkdown(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}

	assert.Equal(t, "# Page\n\nText\n", get("/index.md").Body.String())
	get("/index.md")
	assert.Equal(t, 1, renders)
	get("/nocache.md")
	get("/nocache.md")
	assert.Equal(t, 3, renders)
	assert.Equal(t, http.StatusNotFound, get("/missing.md").Code)

	// the cache is cleared on update
	l.Update(nil)
	get("/index.md")
	assert.Equal(t, 4, renders)
	assert.False(t, IsMarkdownRequest(httptest.NewRequest(http.MethodGet, "/", nil)))
}
