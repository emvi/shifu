package cms

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCMSAcceptedLanguages(t *testing.T) {
	cms := CMS{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	languages := cms.getAcceptedLanguages(req)
	assert.Len(t, languages, 0)

	req.Header.Set("Accept-Language", "en")
	languages = cms.getAcceptedLanguages(req)
	assert.Len(t, languages, 1)
	assert.Equal(t, "en", languages[0])

	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	languages = cms.getAcceptedLanguages(req)
	assert.Len(t, languages, 3)
	assert.Equal(t, "fr", languages[0])
	assert.Equal(t, "en", languages[1])
	assert.Equal(t, "de", languages[2])
}

func BenchmarkCMS(b *testing.B) {
	cfg.Get().BaseDir = "../../demo"
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
