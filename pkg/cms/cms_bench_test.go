package cms

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
)

func BenchmarkLLMsTxt(b *testing.B) {
	dir := b.TempDir()
	write := func(name, content string) {
		_ = os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0755)
		_ = os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	}
	write("tpl/head.html", `<!DOCTYPE html><html><head><title>{{ copy .Page .Content "title" }}</title><meta name="description" content="{{ copy .Page .Content "description" }}" /></head><body><main>`)
	write("tpl/text.html", `<section><h2>{{ copy .Page .Content "headline" }}</h2><div>{{ html (copy .Page .Content "text") }}</div></section>`)
	write("tpl/end.html", `</main></body></html>`)
	text := strings.Repeat("<p>Lorem ipsum <strong>dolor</strong> sit amet, <a href=\"/x\">consectetur</a> adipiscing elit.</p><ul><li>One</li><li>Two</li></ul>", 20)
	sections := make([]string, 10)

	for i := range sections {
		sections[i] = fmt.Sprintf(`{"tpl":"text","copy":{"en":{"headline":"Section %d","text":%q}}}`, i, text)
	}

	for i := range 500 {
		write(fmt.Sprintf("content/page%d.json", i), fmt.Sprintf(`{"path":{"en":"/page-%d"},"content":{"content":[{"tpl":"head","copy":{"en":{"title":"Page %d","description":"Description %d"}}},%s,{"tpl":"end"}]}}`, i, i, i, strings.Join(sections, ",")))
	}

	llms := cfg.Get().LLMs
	cfg.Get().LLMs = cfg.LLMs{Enabled: true, Selector: "main"}
	defer func() {
		cfg.Get().LLMs = llms
	}()
	c := NewCMS(Options{
		Ctx:     context.Background(),
		BaseDir: dir,
		FuncMap: Merge(),
		Source:  source.NewFileSystem(dir, 0),
		Sitemap: sitemap.New(),
	})
	c.updateContent()
	b.ResetTimer()

	for b.Loop() {
		c.updateLLMsTxt() // clears the cache, like a content update
		w := httptest.NewRecorder()
		c.ServeLLMs(w, httptest.NewRequest(http.MethodGet, "/llms.txt", nil))
	}
}
