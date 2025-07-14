package cms

import (
	"context"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCMS(t *testing.T) {
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   "../../demo",
		HotReload: true,
		FuncMap:   defaultFuncMap,
		Source:    source.NewFileSystem("../../demo", 1),
		Sitemap:   sitemap.New(),
	})
	n := 100
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			c.Update()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			c.Serve(w, r)
			c.LastUpdate()
			wg.Done()
		}()
	}

	wg.Wait()
}

func BenchmarkCMS(b *testing.B) {
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   "../../demo",
		HotReload: true,
		FuncMap:   defaultFuncMap,
		Source:    source.NewFileSystem("../../demo", 1),
		Sitemap:   sitemap.New(),
	})
	var wg sync.WaitGroup
	wg.Add(b.N)

	for i := 0; i < b.N; i++ {
		go func() {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			c.Serve(w, r)
			wg.Done()
		}()
	}

	wg.Wait()
}
