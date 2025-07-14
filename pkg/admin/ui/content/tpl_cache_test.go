package content

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestTemplateCfgCache(t *testing.T) {
	n := 10000
	cache := NewTemplateCfgCache()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			go cache.Load()
			assert.Len(t, cache.List(nil), 5)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestTemplateCfgCacheFiltered(t *testing.T) {
	n := 10000
	cache := NewTemplateCfgCache()
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func() {
			go cache.Load()
			assert.Len(t, cache.List([]string{"head", "end"}), 2)
			wg.Done()
		}()
	}

	wg.Wait()
}
