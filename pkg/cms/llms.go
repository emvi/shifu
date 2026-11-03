package cms

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/emvi/shifu/pkg/cfg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	markdownExt   = ".md"
	markdownIndex = "/index"
)

var (
	markdownRemove = []atom.Atom{
		atom.Script,
		atom.Style,
		atom.Noscript,
		atom.Template,
		atom.Slot,
		atom.Nav,
		atom.Form,
		atom.Iframe,
		atom.Svg,
		atom.Button,
	}
)

type markdownRequestKey struct{}

type markdownPage struct {
	Title       string
	Description string
	Markdown    []byte
}

func (cms *CMS) serveMarkdown(w http.ResponseWriter, r *http.Request) {
	path := pagePathFromMarkdown(r.URL.Path)
	md, ok := cms.getMarkdownPage(r, path)

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

func (cms *CMS) generateLLMs(r *http.Request) []byte {
	type entry struct {
		path string
		page *markdownPage
	}

	// static pages only, dynamic routes can't be listed
	cms.m.RLock()
	paths := make(map[string][]string)

	for _, route := range cms.pages {
		if route.isRaw && !route.content.LLMs.Exclude && !cms.isNotFoundPath(route.raw) {
			paths[route.content.Language] = append(paths[route.content.Language], route.raw)
		}
	}

	cms.m.RUnlock()
	languages := make([]string, 0, len(paths))
	entries := make(map[string][]entry)
	title := cfg.Get().LLMs.Title

	for language, list := range paths {
		slices.Sort(list)

		for _, path := range list {
			md, ok := cms.getMarkdownPage(r, path)

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

			buffer.WriteString(fmt.Sprintf("- [%s](%s%s)", oneLine(label), hostname, markdownPath(e.path)))

			if e.page.Description != "" {
				buffer.WriteString(": " + oneLine(e.page.Description))
			}

			buffer.WriteString("\n")
		}
	}

	return buffer.Bytes()
}

func (cms *CMS) getMarkdownPage(r *http.Request, path string) (*markdownPage, bool) {
	page, args, ok := cms.getPage(path)

	if !ok || page.LLMs.Exclude {
		return nil, false
	}

	cache := !page.DisableCache && !cms.isLoggedIn(r)

	if cache {
		cms.m.RLock()
		md, found := cms.markdownCache[path]
		cms.m.RUnlock()

		if found {
			return md, true
		}
	}

	req := r.Clone(context.WithValue(r.Context(), markdownRequestKey{}, true))
	req.URL.Path = path
	out := newBufferResponseWriter()

	if page.Handler != "" {
		handler, found := cms.getHandler(page.Handler)

		if !found {
			slog.Error("Page handler not found", "path", path, "handler", page.Handler)
			return nil, false
		}

		handler(cms, page, args, out, req)
	} else {
		cms.RenderPage(out, req, path, args, &page)
	}

	if out.status >= http.StatusMultipleChoices {
		return nil, false
	}

	md, err := htmlToMarkdown(out.buffer.Bytes(), cfg.Get().LLMs.Selector, cfg.Get().Server.Hostname)

	if err != nil {
		slog.Error("Error converting page to markdown", "path", path, "error", err)
		return nil, false
	}

	if cache {
		cms.m.Lock()
		cms.markdownCache[path] = md
		cms.m.Unlock()
	}

	return md, true
}

func (cms *CMS) isNotFoundPath(path string) bool {
	for _, p := range cms.notFound {
		if p == path {
			return true
		}
	}

	return false
}

func isMarkdownRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	v, _ := r.Context().Value(markdownRequestKey{}).(bool)
	return v
}

func markdownPath(path string) string {
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		path = markdownIndex
	}

	return path + markdownExt
}

func pagePathFromMarkdown(path string) string {
	path = strings.TrimSuffix(path, markdownExt)

	if path == markdownIndex || path == "" {
		return "/"
	}

	return path
}

func htmlToMarkdown(data []byte, selector, domain string) (*markdownPage, error) {
	doc, err := html.Parse(bytes.NewReader(data))

	if err != nil {
		return nil, err
	}

	page := new(markdownPage)

	if title := findNode(doc, func(n *html.Node) bool { return n.DataAtom == atom.Title }); title != nil {
		page.Title = strings.TrimSpace(textContent(title))
	}

	if meta := findNode(doc, func(n *html.Node) bool {
		return n.DataAtom == atom.Meta && strings.EqualFold(attr(n, "name"), "description")
	}); meta != nil {
		page.Description = strings.TrimSpace(attr(meta, "content"))
	}

	node := findNode(doc, selectorMatcher(selector))

	if node == nil {
		node = findNode(doc, func(n *html.Node) bool { return n.DataAtom == atom.Body })

		// without the selected element, the page header and footer (direct children of the body) are dropped
		if node != nil {
			for c := node.FirstChild; c != nil; {
				next := c.NextSibling

				if c.DataAtom == atom.Header || c.DataAtom == atom.Footer {
					node.RemoveChild(c)
				}

				c = next
			}
		}
	}

	if node == nil {
		node = doc
	}

	removeNodes(node, func(n *html.Node) bool { return slices.Contains(markdownRemove, n.DataAtom) })
	content, err := htmltomarkdown.ConvertNode(node, converter.WithDomain(domain))

	if err != nil {
		return nil, err
	}

	content = bytes.TrimSpace(content)
	var buffer bytes.Buffer

	if page.Title != "" && !bytes.HasPrefix(content, []byte("# ")) {
		buffer.WriteString(fmt.Sprintf("# %s\n\n", oneLine(page.Title)))
	}

	if page.Description != "" {
		buffer.WriteString(fmt.Sprintf("> %s\n\n", oneLine(page.Description)))
	}

	buffer.Write(content)
	buffer.WriteString("\n")
	page.Markdown = buffer.Bytes()
	return page, nil
}

func selectorMatcher(selector string) func(*html.Node) bool {
	selector = strings.TrimSpace(selector)

	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}

		if id, ok := strings.CutPrefix(selector, "#"); ok {
			return attr(n, "id") == id
		}

		if class, ok := strings.CutPrefix(selector, "."); ok {
			return slices.Contains(strings.Fields(attr(n, "class")), class)
		}

		return strings.EqualFold(n.Data, selector)
	}
}

func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if n.Type == html.ElementNode && match(n) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, match); found != nil {
			return found
		}
	}

	return nil
}

func removeNodes(n *html.Node, match func(*html.Node) bool) {
	for c := n.FirstChild; c != nil; {
		next := c.NextSibling

		if c.Type == html.ElementNode && match(c) {
			n.RemoveChild(c)
		} else {
			removeNodes(c, match)
		}

		c = next
	}
}

func textContent(n *html.Node) string {
	var sb strings.Builder

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		} else {
			sb.WriteString(textContent(c))
		}
	}

	return sb.String()
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}

	return ""
}

func oneLine(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

type bufferResponseWriter struct {
	header http.Header
	buffer bytes.Buffer
	status int
}

func newBufferResponseWriter() *bufferResponseWriter {
	return &bufferResponseWriter{
		header: make(http.Header),
		status: http.StatusOK,
	}
}

func (w *bufferResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferResponseWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *bufferResponseWriter) WriteHeader(status int) {
	w.status = status
}
