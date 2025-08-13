package cms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/analytics"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"html/template"
	"io/fs"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	notFoundPath      = "/404"
	experimentsCookie = "experiments"
)

// CMS manages pages and content.
type CMS struct {
	baseDir         string
	hotReload       bool
	notFound        map[string]string
	source          source.Provider
	sitemap         *sitemap.Sitemap
	tpl             *Cache
	pages           []Route
	refs            map[string]Content
	handler         map[string]Handler
	pageExperiments map[string][]string
	pageCache       map[string][]byte
	m               sync.RWMutex
}

// Options is the CMS configuration.
type Options struct {
	Ctx       context.Context
	BaseDir   string
	HotReload bool
	NotFound  map[string]string
	FuncMap   template.FuncMap
	Source    source.Provider
	Sitemap   *sitemap.Sitemap
}

// NewCMS sets up a new CMS instance for the given configuration.
func NewCMS(options Options) *CMS {
	if len(options.NotFound) == 0 {
		options.NotFound = map[string]string{"en": notFoundPath}
	}

	cms := &CMS{
		baseDir:         options.BaseDir,
		hotReload:       options.HotReload,
		notFound:        options.NotFound,
		source:          options.Source,
		sitemap:         options.Sitemap,
		pages:           make([]Route, 0),
		refs:            make(map[string]Content),
		handler:         make(map[string]Handler),
		pageExperiments: make(map[string][]string),
		pageCache:       make(map[string][]byte),
	}
	cms.tpl = NewCache(filepath.Join(options.BaseDir, "tpl"), options.FuncMap, options.HotReload)
	cms.source.Watch(options.Ctx, func() {
		slog.Info("Updating website templates, content, and sitemap...")
		cms.tpl.Clear()
		cms.updateContent()
		cms.sitemap.Update()
		slog.Info("Done updating website templates, content, and sitemap")
	})
	return cms
}

// Serve matches the path and renders the page for the given request.
// If no page is found, it will redirect to the 404-page.
func (cms *CMS) Serve(w http.ResponseWriter, r *http.Request) {
	if cms.hotReload {
		cms.updateContent()
	}

	start := time.Now()
	path := r.URL.Path
	page, args, ok := cms.getPage(path)

	if !ok {
		slog.Debug("Page not found", "path", path)
		w.WriteHeader(http.StatusNotFound)
		page, args, ok = cms.getPage(cms.getNotFoundPath(r))

		if !ok {
			_, _ = w.Write([]byte("404 page not found"))
			return
		}
	}

	if page.Handler != "" {
		handler, ok := cms.getHandler(page.Handler)

		if !ok {
			slog.Error("Page handler not found", "path", path, "handler", page.Handler)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		handler(cms, page, args, w, r)
		return
	}

	cms.RenderPage(w, r, path, args, &page)
	slog.Debug("Served page", "time_ms", time.Now().Sub(start).Milliseconds())
}

// RenderPage renders the given page and returns it to the client.
// If no content is present, a default head and body section are served.
func (cms *CMS) RenderPage(w http.ResponseWriter, r *http.Request, path string, args map[string]string, page *Content) {
	page.Request = r
	cms.selectExperiments(w, r, page)

	if cms.redirectExperiment(w, r, page) {
		return
	}

	cms.pageView(r, page)

	for k, v := range page.Header {
		w.Header().Add(k, v)
	}

	if !page.DisableCache && !cms.isLoggedIn(r) {
		cms.m.RLock()
		data, ok := cms.pageCache[path]
		cms.m.RUnlock()

		if ok {
			if _, err := w.Write(data); err != nil {
				slog.Debug("Error sending response", "path", path, "error", err)
			}

			return
		}
	}

	var data []byte

	if cms.hasContent(page) {
		var buffer bytes.Buffer

		for position, content := range page.Content {
			out, err := cms.renderContent(args, page, position, content)

			if err != nil {
				slog.Error("Error rendering template", "path", path, "error", err)
				return
			}

			buffer.Write(out)
		}

		data = buffer.Bytes()
	} else {
		data = []byte(fmt.Sprintf(defaultPageContent, adminBody(page)))
	}

	if _, err := w.Write(data); err != nil {
		slog.Debug("Error sending response", "path", path, "error", err)
	}

	if !page.DisableCache && !cms.isLoggedIn(r) {
		cms.m.Lock()
		cms.pageCache[path] = data
		cms.m.Unlock()
	}
}

// Render404 renders the 404 page for the given path and language if it exists.
// The language will fall back to en if not found or empty.
func (cms *CMS) Render404(w http.ResponseWriter, r *http.Request, path string) {
	slog.Debug("Page not found", "path", path)
	w.WriteHeader(http.StatusNotFound)
	page, args, ok := cms.getPage(cms.getNotFoundPath(r))

	if ok {
		page.Request = r
		cms.RenderPage(w, r, path, args, &page)
	}
}

// Render renders and returns the content for the given page.
func (cms *CMS) Render(args map[string]string, page *Content, position string, content []Content) template.HTML {
	out, err := cms.renderContent(args, page, position, content)

	if err != nil {
		slog.Error("Error rendering template", "error", err)
		return template.HTML(err.Error())
	}

	return template.HTML(out)
}

// RenderElement renders the given element and returns the content.
func (cms *CMS) RenderElement(w http.ResponseWriter, r *http.Request, page *Content, position string, element *Content) ([]byte, error) {
	page.Request = r
	cms.selectExperiments(w, r, page)
	var buffer bytes.Buffer

	if err := cms.renderElement(&buffer, nil, page, position, -1, element); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// SetHandler sets the handler function for the given name.
func (cms *CMS) SetHandler(name string, handler Handler) {
	cms.m.Lock()
	defer cms.m.Unlock()
	cms.handler[name] = handler
}

// Update updates the templates, content, and sitemap.
func (cms *CMS) Update() {
	slog.Info("Updating CMS")
	cms.source.Update(func() {
		slog.Info("Updating website templates, content, and sitemap...")
		cms.tpl.Clear()
		cms.updateContent()
		cms.sitemap.Update()
		slog.Info("Done updating website templates, content, and sitemap")
	})
}

// LastUpdate returns the time the website data has last been updated.
func (cms *CMS) LastUpdate() string {
	return cms.source.LastUpdate().Format(time.RFC3339)
}

func (cms *CMS) renderContent(args map[string]string, page *Content, position string, content []Content) ([]byte, error) {
	var buffer bytes.Buffer

	for i, c := range content {
		if err := cms.renderElement(&buffer, args, page, position, i, &c); err != nil {
			return nil, err
		}
	}

	return buffer.Bytes(), nil
}

func (cms *CMS) renderElement(buffer *bytes.Buffer, args map[string]string, page *Content, position string, index int, c *Content) error {
	elementPosition := ""

	if index > -1 {
		elementPosition = fmt.Sprintf("%s.%d", position, index)
	} else {
		elementPosition = position
	}

	if c.Ref != "" {
		cms.m.RLock()
		ref, ok := cms.refs[c.Ref]

		if !ok {
			cms.m.RUnlock()
			buffer.WriteString(fmt.Sprintf(`Reference "%s" not found!`, c.Ref))
			return nil
		}

		ref = ref.Clone()
		cms.m.RUnlock()

		// overwrite data
		if ref.Data == nil {
			ref.Data = make(map[string]any)
		}

		for k, v := range c.Data {
			if !cms.isZeroValue(v) {
				ref.Data[k] = v
			}
		}

		if ref.Copy == nil {
			ref.Copy = make(Copy)
		}

		// overwrite copy
		for language, vars := range c.Copy {
			if _, ok := ref.Copy[language]; !ok {
				ref.Copy[language] = make(map[string]any)
			}

			for k, v := range vars {
				if v != "" {
					ref.Copy[language][k] = v
				}
			}
		}

		// overwrite analytics
		if ref.Analytics.Tags == nil {
			ref.Analytics.Tags = make(map[string]string)
		}

		for k, v := range c.Analytics.Tags {
			ref.Analytics.Tags[k] = v
		}

		if c.Analytics.Experiment.Name != "" {
			ref.Analytics.Experiment.Name = c.Analytics.Experiment.Name
		}

		if c.Analytics.Experiment.Variant != "" {
			ref.Analytics.Experiment.Variant = c.Analytics.Experiment.Variant
		}

		// render reference
		ref.Position = elementPosition
		out, err := cms.render(ref.Tpl, args, page, &ref)

		if err != nil {
			return err
		}

		buffer.Write(out)
	} else {
		c.Position = elementPosition
		out, err := cms.render(c.Tpl, args, page, c)

		if err != nil {
			return err
		}

		buffer.Write(out)
	}

	return nil
}

func (cms *CMS) render(tpl string, args map[string]string, page *Content, content *Content) ([]byte, error) {
	if content.Analytics.Experiment.Name != "" && page.SelectedExperiments[content.Analytics.Experiment.Name] != content.Analytics.Experiment.Variant {
		return nil, nil
	}

	out, err := cms.tpl.Render(fmt.Sprintf("%s.html", tpl), struct {
		CMS     *CMS
		Args    map[string]string
		Page    *Content
		Content *Content
	}{
		cms,
		args,
		page,
		content,
	})

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (cms *CMS) selectExperiments(w http.ResponseWriter, r *http.Request, page *Content) {
	selected := make(map[string]string)
	updateCookie := true
	cookie, err := r.Cookie(experimentsCookie)

	if err == nil {
		updateCookie = false
		kv := strings.Split(cookie.Value, ",")

		for _, v := range kv {
			if v != "" {
				left, right, ok := strings.Cut(v, ":")

				if ok {
					selected[left] = right
				}
			}
		}
	}

	for experiment, variants := range page.Experiments {
		if len(variants) < 2 {
			continue
		}

		selectedVariant, ok := selected[experiment]

		if !ok || slices.Index(variants, selectedVariant) == -1 {
			selected[experiment] = variants[rand.Intn(len(variants))]
			updateCookie = true
		}
	}

	if page.Analytics.Experiment.Name != "" {
		cms.m.RLock()
		variants, ok := cms.pageExperiments[page.Analytics.Experiment.Name]
		cms.m.RUnlock()

		if ok && len(variants) > 1 {
			selectedVariant, ok := selected[page.Analytics.Experiment.Name]

			if !ok || slices.Index(variants, selectedVariant) == -1 {
				variant := variants[rand.Intn(len(variants))]
				selected[page.Analytics.Experiment.Name] = variant
				page.SelectedPageExperiment = variant
				updateCookie = true
			} else {
				page.SelectedPageExperiment = selectedVariant
			}
		}
	}

	if updateCookie {
		var builder strings.Builder

		for k, v := range selected {
			builder.WriteString(fmt.Sprintf("%s:%s,", k, v))
		}

		http.SetCookie(w, &http.Cookie{
			Name:     experimentsCookie,
			Value:    builder.String(),
			Expires:  time.Now().UTC().Add(time.Hour * 24),
			Secure:   cfg.Get().Server.SecureCookies,
			Domain:   cfg.Get().Server.CookieDomainName,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	}

	page.SelectedExperiments = selected
}

func (cms *CMS) redirectExperiment(w http.ResponseWriter, r *http.Request, page *Content) bool {
	if page.SelectedPageExperiment != "" && page.Analytics.Experiment.Variant != page.SelectedPageExperiment {
		cms.m.RLock()
		defer cms.m.RUnlock()

		for _, route := range cms.pages {
			if route.content.Analytics.Experiment.Name == page.Analytics.Experiment.Name && route.content.Analytics.Experiment.Variant == page.SelectedPageExperiment {
				redirect, ok := route.content.Path[page.Language]

				if ok && r.URL.Path != redirect {
					http.Redirect(w, r, redirect, http.StatusFound)
					return true
				}
			}
		}
	}

	return false
}

func (cms *CMS) pageView(r *http.Request, page *Content) {
	if page.Analytics.Tags == nil {
		page.Analytics.Tags = make(map[string]string)
	}

	for k, v := range page.SelectedExperiments {
		page.Analytics.Tags[k] = v
	}

	go analytics.PageView(r, page.Analytics.Tags)
}

func (cms *CMS) updateContent() {
	pages := make([]Route, 0)
	refs := make(map[string]Content)
	pageExperiments := make(map[string][]string)
	websiteHost := cfg.Get().Server.Hostname
	contentDir := filepath.Join(cms.baseDir, "content")
	_, err := os.Stat(contentDir)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(contentDir, 0744); err != nil {
			slog.Error("Error creating content directory", "error", err, "path", contentDir)
			return
		}
	} else if err != nil {
		slog.Error("Error reading content directory", "error", err, "path", contentDir)
		return
	}

	// extract refs
	if err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, err := cms.getContent(path)

			if err != nil {
				return err
			}

			// when the template is specified on the first level, this is a standalone component, otherwise, it's a page
			if content.Tpl != "" {
				name := strings.ToLower(strings.TrimSuffix(d.Name(), filepath.Ext(d.Name())))
				refs[name] = *content
			}
		}

		return nil
	}); err != nil {
		slog.Error("Error reading website content directory", "error", err)
	}

	// extract pages and experiments
	if err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			content, err := cms.getContent(path)

			if err != nil {
				return err
			}

			if content.Tpl == "" {
				for language, p := range content.Path {
					info, err := d.Info()

					if err != nil {
						return errors.New(fmt.Sprintf("error reading website content file info '%s': %s", path, err))
					}

					content.Language = language
					content.CanonicalLink = websiteHost + p
					cms.updateExperiments(refs, pageExperiments, content)
					route, err := NewRoute(p, *content)

					if err != nil {
						return err
					}

					pages = append(pages, *route)
					priority := content.Sitemap.Priority

					if priority == "" {
						priority = "1.0"
					}

					cms.sitemap.Set(p, priority, info.ModTime().Format(sitemap.SitemapLastModFormat))
				}
			}
		}

		return nil
	}); err != nil {
		slog.Error("Error reading website content directory", "error", err)
	}

	// the longest path goes first
	slices.SortFunc(pages, func(a, b Route) int {
		if len(a.raw) > len(b.raw) {
			return -1
		}

		return 1
	})

	if len(pages) == 0 {
		pages = append(pages, Route{
			raw:    "/",
			isRaw:  true,
			length: 1,
			content: Content{
				Path: map[string]string{
					"en": "/",
				},
			},
		})
	}

	cms.m.Lock()
	defer cms.m.Unlock()
	cms.pages = pages
	cms.refs = refs
	cms.pageExperiments = pageExperiments
	cms.pageCache = make(map[string][]byte)
}

func (cms *CMS) getContent(path string) (*Content, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("error reading website content file '%s': %s", path, err))
	}

	var content Content

	if err := json.Unmarshal(data, &content); err != nil {
		return nil, errors.New(fmt.Sprintf("error parsing website content file '%s': %s", path, err))
	}

	content.File = strings.TrimPrefix(path, cms.baseDir)
	return &content, nil
}

func (cms *CMS) updateExperiments(refs map[string]Content, pageExperiments map[string][]string, page *Content) {
	if page.Analytics.Experiment.Name != "" && page.Analytics.Experiment.Variant != "" {
		if pageExperiments[page.Analytics.Experiment.Name] == nil {
			pageExperiments[page.Analytics.Experiment.Name] = make([]string, 0)
		}

		if slices.Index(pageExperiments[page.Analytics.Experiment.Name], page.Analytics.Experiment.Variant) == -1 {
			pageExperiments[page.Analytics.Experiment.Name] = append(pageExperiments[page.Analytics.Experiment.Name], page.Analytics.Experiment.Variant)
		}
	}

	experiments := make(map[string][]string)
	cms.extractExperiments(refs, page, experiments)
	page.Experiments = experiments
}

func (cms *CMS) extractExperiments(refs map[string]Content, content *Content, experiments map[string][]string) {
	for _, elements := range content.Content {
		for _, element := range elements {
			if element.Ref != "" && (element.Analytics.Experiment.Name == "" || element.Analytics.Experiment.Variant == "") {
				ref, ok := refs[element.Ref]

				if ok {
					if element.Analytics.Experiment.Name == "" {
						element.Analytics.Experiment.Name = ref.Analytics.Experiment.Name
					}

					if element.Analytics.Experiment.Variant == "" {
						element.Analytics.Experiment.Variant = ref.Analytics.Experiment.Variant
					}
				}
			}

			if element.Analytics.Experiment.Name != "" {
				if experiments[element.Analytics.Experiment.Name] == nil {
					experiments[element.Analytics.Experiment.Name] = make([]string, 0)
				}

				if slices.Index(experiments[element.Analytics.Experiment.Name], element.Analytics.Experiment.Variant) == -1 {
					experiments[element.Analytics.Experiment.Name] = append(experiments[element.Analytics.Experiment.Name], element.Analytics.Experiment.Variant)
				}
			}

			cms.extractExperiments(refs, &element, experiments)
		}
	}
}

func (cms *CMS) getPage(path string) (Content, map[string]string, bool) {
	cms.m.RLock()
	defer cms.m.RUnlock()

	for i := range cms.pages {
		args, found := cms.pages[i].Match(path)

		if found {
			return cms.pages[i].content, args, true
		}
	}

	return Content{}, nil, false
}

func (cms *CMS) getHandler(name string) (Handler, bool) {
	cms.m.RLock()
	defer cms.m.RUnlock()
	handler, found := cms.handler[name]
	return handler, found
}

func (cms *CMS) getNotFoundPath(r *http.Request) string {
	languages := GetAcceptedLanguages(r)

	for _, l := range languages {
		p, found := cms.notFound[l]

		if found {
			return p
		}
	}

	if !slices.Contains(languages, "en") {
		return cms.notFound["en"]
	}

	return ""
}

func (cms *CMS) hasContent(page *Content) bool {
	for _, content := range page.Content {
		if len(content) > 0 {
			return true
		}
	}

	return false
}

func (cms *CMS) isZeroValue(v any) bool {
	if v == nil {
		return true
	}

	return reflect.ValueOf(v).IsZero()
}

func (cms *CMS) isLoggedIn(r *http.Request) bool {
	session, err := r.Cookie("session")
	return err == nil && session.Value != ""
}
