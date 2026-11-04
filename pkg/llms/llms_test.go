package llms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

	// the description follows an existing h1
	page, err = htmlToMarkdown([]byte(`<html><head><title>T</title><meta name="description" content="D" /></head><body><main><h1>Hello</h1><p>Text</p></main></body></html>`), "main", "")
	assert.NoError(t, err)
	assert.Equal(t, "# Hello\n\n> D\n\nText\n", string(page.Markdown))

	// id and class selectors
	page, err = htmlToMarkdown([]byte(`<body><div id="a">A</div><div class="x b">B</div></body>`), ".b", "")
	assert.NoError(t, err)
	assert.Equal(t, "B\n", string(page.Markdown))
	page, err = htmlToMarkdown([]byte(`<body><div id="a">A</div><div class="x b">B</div></body>`), "#a", "")
	assert.NoError(t, err)
	assert.Equal(t, "A\n", string(page.Markdown))
}

func TestExtractMeta(t *testing.T) {
	title, description := extractMeta([]byte(`<!DOCTYPE html><html><head>
		<meta charset="UTF-8" />
		<title> Caf&eacute; &amp; Bar </title>
		<META NAME="Description" CONTENT="Coffee &amp; drinks." />
		<meta name="description" content="second" />
	</head><body><title>Ignored</title></body></html>`))
	assert.Equal(t, "Café & Bar", title)
	assert.Equal(t, "Coffee & drinks.", description)

	// stops at the body
	title, description = extractMeta([]byte(`<html><head></head><body><title>T</title><meta name="description" content="D" /></body></html>`))
	assert.Empty(t, title)
	assert.Empty(t, description)

	title, description = extractMeta([]byte(`<title>Only title</title>`))
	assert.Equal(t, "Only title", title)
	assert.Empty(t, description)
}

func TestLLMsCache(t *testing.T) {
	// no background generation of the llms.txt during this test
	delay := generateDelay
	generateDelay = time.Hour
	defer func() {
		generateDelay = delay
	}()
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
	l.Update([]Page{{Language: "en", Path: "/"}, {Language: "en", Path: "/nocache"}}, "1")
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

	// the cache is kept if the content version didn't change, and cleared otherwise
	l.Update(nil, "1")
	get("/index.md")
	assert.Equal(t, 3, renders)
	l.Update(nil, "2")
	get("/index.md")
	assert.Equal(t, 4, renders)
	assert.False(t, IsMarkdownRequest(httptest.NewRequest(http.MethodGet, "/", nil)))
}

func TestLLMsBackground(t *testing.T) {
	delay := generateDelay
	generateDelay = time.Millisecond
	defer func() {
		generateDelay = delay
	}()
	var renders atomic.Int32
	var title atomic.Value
	title.Store("One")
	block := make(chan struct{})
	close(block)
	var blockM sync.Mutex
	l := New()
	l.SetRenderer(func(w http.ResponseWriter, r *http.Request, path string) (bool, bool) {
		blockM.Lock()
		b := block
		blockM.Unlock()
		<-b
		renders.Add(1)
		_, _ = fmt.Fprintf(w, "<title>%s</title><main>Text</main>", title.Load())
		return true, true
	})
	get := func() string {
		w := httptest.NewRecorder()
		l.ServeLLMsTxt(w, httptest.NewRequest(http.MethodGet, "/llms.txt", nil))
		return w.Body.String()
	}

	// the first request waits for the llms.txt to be generated
	l.Update([]Page{{Language: "en", Path: "/"}}, "1")
	assert.Contains(t, get(), "- [One](/index.md)")
	assert.Equal(t, int32(1), renders.Load())

	// unchanged content isn't generated again
	l.Update([]Page{{Language: "en", Path: "/"}}, "1")
	time.Sleep(time.Millisecond * 20)
	assert.Contains(t, get(), "- [One](/index.md)")
	assert.Equal(t, int32(1), renders.Load())

	// while a new version is generated, the previous one is served without waiting
	blockM.Lock()
	block = make(chan struct{})
	blockM.Unlock()
	title.Store("Two")
	l.Update([]Page{{Language: "en", Path: "/"}}, "2")
	time.Sleep(time.Millisecond * 20)
	assert.Contains(t, get(), "- [One](/index.md)")
	blockM.Lock()
	close(block)
	blockM.Unlock()
	assert.Eventually(t, func() bool {
		return strings.Contains(get(), "- [Two](/index.md)")
	}, time.Second, time.Millisecond*5)

	// refresh generates it again without changing the content
	title.Store("Three")
	l.Refresh()
	assert.Eventually(t, func() bool {
		return strings.Contains(get(), "- [Three](/index.md)")
	}, time.Second, time.Millisecond*5)
}
