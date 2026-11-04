package llms

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"slices"
	"sync"
	"time"

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

// pageMeta is the title and description of a page listed in the llms.txt.
type pageMeta struct {
	title       string
	description string
}

// LLMs builds and serves the llms.txt and the Markdown versions of pages.
type LLMs struct {
	render RenderFunc
	pages  []Page
	cache  map[string]*markdownPage
	meta   map[string]pageMeta
	txt    []byte
	m      sync.RWMutex
}

// New creates a new LLMs instance.
func New() *LLMs {
	return &LLMs{
		cache: make(map[string]*markdownPage),
		meta:  make(map[string]pageMeta),
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
	llms.meta = make(map[string]pageMeta)
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

// ServeLLMsTxt serves the llms.txt listing all pages with a link to their Markdown version.
func (llms *LLMs) ServeLLMsTxt(w http.ResponseWriter, r *http.Request) {
	llms.m.RLock()
	data := llms.txt
	llms.m.RUnlock()

	if data == nil {
		data = llms.generateLLMsTxt(r)
		llms.m.Lock()
		llms.txt = data
		llms.m.Unlock()
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if _, err := w.Write(data); err != nil {
		slog.Debug("Error sending llms.txt", "error", err)
	}
}

func (llms *LLMs) generateLLMsTxt(r *http.Request) []byte {
	slog.Debug("Generating llms.txt...")
	now := time.Now()
	type entry struct {
		path string
		meta pageMeta
		ok   bool
	}

	llms.m.RLock()
	paths := make(map[string][]string)

	for _, page := range llms.pages {
		paths[page.Language] = append(paths[page.Language], page.Path)
	}

	llms.m.RUnlock()
	languages := make([]string, 0, len(paths))
	entries := make(map[string][]entry)
	jobs := make(chan *entry)
	var wg sync.WaitGroup

	for range runtime.NumCPU() {
		wg.Go(func() {
			for e := range jobs {
				e.meta, e.ok = llms.getMeta(r, e.path)
			}
		})
	}

	for language, list := range paths {
		slices.Sort(list)
		entries[language] = make([]entry, len(list))

		for i, path := range list {
			entries[language][i].path = path
			jobs <- &entries[language][i]
		}
	}

	close(jobs)
	wg.Wait()
	title := cfg.Get().LLMs.Title

	for language, list := range entries {
		list = slices.DeleteFunc(list, func(e entry) bool {
			return !e.ok
		})
		entries[language] = list

		for _, e := range list {
			if title == "" && e.path == "/" {
				title = e.meta.title
			}
		}

		if len(list) > 0 {
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
			label := e.meta.title

			if label == "" {
				label = e.path
			}

			buffer.WriteString(fmt.Sprintf("- [%s](%s%s)", oneLine(label), hostname, MarkdownPath(e.path)))

			if e.meta.description != "" {
				buffer.WriteString(": " + oneLine(e.meta.description))
			}

			buffer.WriteString("\n")
		}
	}

	slog.Debug("Done generating llms.txt", "time_ms", time.Since(now).Milliseconds())
	return buffer.Bytes()
}

func (llms *LLMs) getPage(r *http.Request, path string) (*markdownPage, bool) {
	llms.m.RLock()
	md, found := llms.cache[path]
	llms.m.RUnlock()

	if found {
		return md, true
	}

	data, cache, ok := llms.renderHTML(r, path)

	if !ok {
		return nil, false
	}

	md, err := htmlToMarkdown(data, cfg.Get().LLMs.Selector, cfg.Get().Server.Hostname)

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

func (llms *LLMs) getMeta(r *http.Request, path string) (pageMeta, bool) {
	llms.m.RLock()
	meta, found := llms.meta[path]
	md, foundMarkdown := llms.cache[path]
	llms.m.RUnlock()

	if found {
		return meta, true
	}

	if foundMarkdown {
		return pageMeta{md.Title, md.Description}, true
	}

	data, cache, ok := llms.renderHTML(r, path)

	if !ok {
		return pageMeta{}, false
	}

	meta.title, meta.description = extractMeta(data)

	if cache {
		llms.m.Lock()
		llms.meta[path] = meta
		llms.m.Unlock()
	}

	return meta, true
}

func (llms *LLMs) renderHTML(r *http.Request, path string) ([]byte, bool, bool) {
	llms.m.RLock()
	render := llms.render
	llms.m.RUnlock()

	if render == nil {
		return nil, false, false
	}

	out := newBufferResponseWriter()
	ok, cache := render(out, markdownRequest(r, path), path)

	if !ok || out.status >= http.StatusMultipleChoices {
		return nil, false, false
	}

	return out.buffer.Bytes(), cache, true
}
