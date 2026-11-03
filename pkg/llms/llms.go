package llms

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"sync"

	"github.com/emvi/shifu/pkg/cfg"
)

// Page is a page listed in the llms.txt.
type Page struct {
	Language string
	Path     string
}

// RenderFunc renders the HTML of the page for given path.
// It returns false if the page does not exist (or is excluded) and whether the result can be cached.
type RenderFunc func(w http.ResponseWriter, r *http.Request, path string) (found bool, cache bool)

// LLMs builds and serves the llms.txt and the Markdown versions of pages.
type LLMs struct {
	render RenderFunc
	pages  []Page
	cache  map[string]*markdownPage
	txt    []byte
	m      sync.RWMutex
}

// New creates a new LLMs instance.
func New() *LLMs {
	return &LLMs{
		cache: make(map[string]*markdownPage),
	}
}

// SetRenderer sets the function used to render pages.
func (llms *LLMs) SetRenderer(render RenderFunc) {
	llms.m.Lock()
	defer llms.m.Unlock()
	llms.render = render
}

// Update sets the pages listed in the llms.txt and clears the cache.
func (llms *LLMs) Update(pages []Page) {
	llms.m.Lock()
	defer llms.m.Unlock()
	llms.pages = pages
	llms.cache = make(map[string]*markdownPage)
	llms.txt = nil
}

// ServeMarkdown serves the Markdown version of the page requested by the Markdown path.
func (llms *LLMs) ServeMarkdown(w http.ResponseWriter, r *http.Request) {
	path := PagePath(r.URL.Path)
	md, ok := llms.getPage(r, path)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found"))
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("X-Robots-Tag", "noindex")
	w.Header().Set("Link", fmt.Sprintf(`<%s%s>; rel="canonical"`, cfg.Get().Server.Hostname, path))

	if _, err := w.Write(md.Markdown); err != nil {
		slog.Debug("Error sending markdown", "path", path, "error", err)
	}
}

// ServeTxt serves the llms.txt listing all pages with a link to their Markdown version.
func (llms *LLMs) ServeTxt(w http.ResponseWriter, r *http.Request) {
	llms.m.RLock()
	data := llms.txt
	llms.m.RUnlock()

	if data == nil {
		data = llms.generate(r)
		llms.m.Lock()
		llms.txt = data
		llms.m.Unlock()
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if _, err := w.Write(data); err != nil {
		slog.Debug("Error sending llms.txt", "error", err)
	}
}

func (llms *LLMs) generate(r *http.Request) []byte {
	type entry struct {
		path string
		page *markdownPage
	}

	llms.m.RLock()
	paths := make(map[string][]string)

	for _, page := range llms.pages {
		paths[page.Language] = append(paths[page.Language], page.Path)
	}

	llms.m.RUnlock()
	languages := make([]string, 0, len(paths))
	entries := make(map[string][]entry)
	title := cfg.Get().LLMs.Title

	for language, list := range paths {
		slices.Sort(list)

		for _, path := range list {
			md, ok := llms.getPage(r, path)

			if ok {
				entries[language] = append(entries[language], entry{path, md})

				if title == "" && path == "/" {
					title = md.Title
				}
			}
		}

		if len(entries[language]) > 0 {
			languages = append(languages, language)
		}
	}

	slices.Sort(languages)

	if title == "" {
		title = "Website"
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("# %s\n", oneLine(title)))

	if description := cfg.Get().LLMs.Description; description != "" {
		buffer.WriteString(fmt.Sprintf("\n> %s\n", oneLine(description)))
	}

	hostname := cfg.Get().Server.Hostname

	for _, language := range languages {
		if len(languages) > 1 {
			buffer.WriteString(fmt.Sprintf("\n## %s\n\n", language))
		} else {
			buffer.WriteString("\n## Pages\n\n")
		}

		for _, e := range entries[language] {
			label := e.page.Title

			if label == "" {
				label = e.path
			}

			buffer.WriteString(fmt.Sprintf("- [%s](%s%s)", oneLine(label), hostname, MarkdownPath(e.path)))

			if e.page.Description != "" {
				buffer.WriteString(": " + oneLine(e.page.Description))
			}

			buffer.WriteString("\n")
		}
	}

	return buffer.Bytes()
}

func (llms *LLMs) getPage(r *http.Request, path string) (*markdownPage, bool) {
	llms.m.RLock()
	render := llms.render
	md, found := llms.cache[path]
	llms.m.RUnlock()

	if found {
		return md, true
	}

	if render == nil {
		return nil, false
	}

	out := newBufferResponseWriter()
	ok, cache := render(out, markdownRequest(r, path), path)

	if !ok || out.status >= http.StatusMultipleChoices {
		return nil, false
	}

	md, err := htmlToMarkdown(out.buffer.Bytes(), cfg.Get().LLMs.Selector, cfg.Get().Server.Hostname)

	if err != nil {
		slog.Error("Error converting page to markdown", "path", path, "error", err)
		return nil, false
	}

	if cache {
		llms.m.Lock()
		llms.cache[path] = md
		llms.m.Unlock()
	}

	return md, true
}
