package cms

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func BenchmarkCMS(b *testing.B) {
	cfg.Get().BaseDir = "../../demo"
	c := NewCMS(Options{
		Ctx:       context.Background(),
		BaseDir:   "../../demo",
		HotReload: true,
		FuncMap:   defaultFuncMap,
		Source:    source.NewFS("../../demo", 1),
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
