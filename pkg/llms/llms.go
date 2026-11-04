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

var (
	// generateDelay is the time to wait after an update before the llms.txt is generated,
	// so that multiple updates in a short period of time only generate it once.
	generateDelay = time.Second
)

// pageMeta is the title and description of a page listed in the llms.txt.
type pageMeta struct {
	title       string
	description string
}

// Page is a page listed in the llms.txt.
type Page struct {
	Language string
	Path     string
}

// RenderFunc renders the HTML of the page for given path.
// It returns false if the page does not exist (or is excluded) and whether the result can be cached.
type RenderFunc func(w http.ResponseWriter, r *http.Request, path string) (found bool, cache bool)

// LLMs builds and serves the llms.txt and the Markdown versions of pages.
// The llms.txt is generated in the background after the content has changed.
type LLMs struct {
	render     RenderFunc
	pages      []Page
	version    string
	cache      map[string]*markdownPage
	meta       map[string]pageMeta
	txt        []byte
	ready      chan struct{}
	generation uint64
	timer      *time.Timer
	m          sync.RWMutex
}

// New creates a new LLMs instance.
func New() *LLMs {
	return &LLMs{
		cache: make(map[string]*markdownPage),
		meta:  make(map[string]pageMeta),
		ready: make(chan struct{}),
	}
}

// SetRenderer sets the function used to render pages.
func (llms *LLMs) SetRenderer(render RenderFunc) {
	llms.m.Lock()
	defer llms.m.Unlock()
	llms.render = render
}

// Update sets the pages listed in the llms.txt for given content version.
// If the version has changed, the cache is cleared and the llms.txt is generated in the background.
// The previous llms.txt is served until the new one is ready.
func (llms *LLMs) Update(pages []Page, version string) {
	llms.m.Lock()
	defer llms.m.Unlock()

	if version != "" && version == llms.version {
		return
	}

	llms.version = version
	llms.pages = pages
	llms.cache = make(map[string]*markdownPage)
	llms.meta = make(map[string]pageMeta)
	llms.schedule()
}

// Refresh clears the cache and generates the llms.txt in the background, even if the content hasn't changed.
// This is required if the output of pages changes without changing the content (e.g. by adding a page handler).
func (llms *LLMs) Refresh() {
	llms.m.Lock()
	defer llms.m.Unlock()
	llms.cache = make(map[string]*markdownPage)
	llms.meta = make(map[string]pageMeta)
	llms.schedule()
}

// Generate generates the llms.txt immediately and blocks until it's done.
func (llms *LLMs) Generate() {
	llms.m.Lock()
	llms.generation++
	generation := llms.generation

	if llms.timer != nil {
		llms.timer.Stop()
	}

	llms.m.Unlock()
	llms.generate(generation)
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
	ready := llms.ready
	llms.m.RUnlock()

	// only the first request after startup has to wait, later requests get the previous version while a new one is generated
	if data == nil {
		select {
		case <-ready:
		case <-r.Context().Done():
			return
		}

		llms.m.RLock()
		data = llms.txt
		llms.m.RUnlock()
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if _, err := w.Write(data); err != nil {
		slog.Debug("Error sending llms.txt", "error", err)
	}
}

func (llms *LLMs) schedule() {
	llms.generation++
	generation := llms.generation

	if llms.timer != nil {
		llms.timer.Stop()
	}

	llms.timer = time.AfterFunc(generateDelay, func() {
		llms.generate(generation)
	})
}

func (llms *LLMs) generate(generation uint64) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)

	if err != nil {
		slog.Error("Error creating request to generate llms.txt", "error", err)
		return
	}

	data := llms.generateLLMsTxt(r, generation)

	if data == nil {
		return
	}

	llms.m.Lock()
	defer llms.m.Unlock()

	if generation == llms.generation {
		if llms.txt == nil {
			close(llms.ready)
		}

		llms.txt = data
	}
}

func (llms *LLMs) isCurrent(generation uint64) bool {
	llms.m.RLock()
	defer llms.m.RUnlock()
	return generation == llms.generation
}

// generateLLMsTxt generates the llms.txt. It returns nil if a newer generation has been started in the meantime.
func (llms *LLMs) generateLLMsTxt(r *http.Request, generation uint64) []byte {
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

	// leave half of the CPUs to serve regular page views
	for range max(1, runtime.NumCPU()/2) {
		wg.Go(func() {
			for e := range jobs {
				if llms.isCurrent(generation) {
					e.meta, e.ok = llms.getMeta(r, e.path)
				}
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

	if !llms.isCurrent(generation) {
		slog.Debug("Canceled generating llms.txt, because a newer version is being generated")
		return nil
	}

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
