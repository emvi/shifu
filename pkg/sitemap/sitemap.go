package sitemap

import (
	"encoding/xml"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/go-chi/chi/v5"
	"github.com/klauspost/compress/gzhttp"
)

const (
	LastModFormat = "2006-01-02"

	header = `<?xml version="1.0" encoding="UTF-8"?>`
	xmlns  = "http://www.sitemaps.org/schemas/sitemap/0.9"
)

// URLSet is a set of urls for the sitemap.xml.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL
}

// URL is a single entry inside a URLSet.
type URL struct {
	XMLName    xml.Name `xml:"url"`
	Loc        string   `xml:"loc"`
	Lastmod    string   `xml:"lastmod"`
	Changefreq string   `xml:"changefreq,omitempty"`
	Priority   string   `xml:"priority,omitempty"`
}

// Sitemap builds and serves a sitemap.
type Sitemap struct {
	sitemap map[string]URL
	content []byte
	m       sync.RWMutex
}

// New creates a new sitemap.
func New() *Sitemap {
	return &Sitemap{
		sitemap: make(map[string]URL),
	}
}

// Set sets a page in the sitemap.
func (sitemap *Sitemap) Set(loc, priority, lastmod string) {
	sitemap.m.Lock()
	defer sitemap.m.Unlock()
	sitemap.sitemap[loc] = URL{
		Loc:      cfg.Get().Server.Hostname + loc,
		Priority: priority,
		Lastmod:  lastmod,
	}
}

// Update updates the sitemap content.
func (sitemap *Sitemap) Update() {
	sitemap.m.Lock()
	defer sitemap.m.Unlock()
	list := make([]URL, 0, len(sitemap.sitemap))

	for _, entry := range sitemap.sitemap {
		list = append(list, entry)
	}

	var err error
	sitemap.content, err = sitemap.generate(list)

	if err != nil {
		slog.Error("Error generating sitemap", "error", err)
	}
}

// Serve adds a handler to server the sitemap.
func (sitemap *Sitemap) Serve(router chi.Router) {
	router.Handle("/sitemap.xml", gzhttp.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/xml")
		w.Header().Add("Cache-Control", "max-age=3600")
		sitemap.m.RLock()
		defer sitemap.m.RUnlock()

		if _, err := w.Write(sitemap.content); err != nil {
			slog.Debug("Error serving sitemap.xml", "error", err)
		}
	})))
}

func (sitemap *Sitemap) generate(urls []URL) ([]byte, error) {
	now := time.Now().Format(LastModFormat)

	for i := range urls {
		if urls[i].Lastmod == "" {
			urls[i].Lastmod = now
		}
	}

	s := URLSet{
		XMLNS: xmlns,
		URLs:  urls,
	}
	out, err := xml.Marshal(&s)

	if err != nil {
		return nil, err
	}

	return []byte(header + string(out)), nil
}
