package content

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/stretchr/testify/assert"
)

func TestResolveElementTarget(t *testing.T) {
	dir := t.TempDir()
	baseDir := cfg.Get().BaseDir
	cfg.Get().BaseDir = dir
	defer func() {
		cfg.Get().BaseDir = baseDir
	}()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "content", "refs"), 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "content", "refs", "slider.json"), []byte(`{"tpl":"slider","content":{"slides":[{"tpl":"slide"},{"ref":"inner"}]}}`), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "content", "refs", "inner.json"), []byte(`{"tpl":"text","content":{"content":[{"tpl":"img"}]}}`), 0644))
	page := &cms.Content{
		File: "page.json",
		Content: map[string][]cms.Content{
			"content": {
				{Tpl: "head"},
				{Ref: "slider"},
			},
		},
	}

	// regular element on the page
	target, err := resolveElementTarget(page, "content.0", true)
	assert.NoError(t, err)
	assert.False(t, target.isRef())
	assert.Equal(t, "content.0", target.path)
	assert.Equal(t, "head", target.element().Tpl)

	// the reference element itself (container on the page)
	target, err = resolveElementTarget(page, "content.1", false)
	assert.NoError(t, err)
	assert.False(t, target.isRef())
	assert.Equal(t, "slider", target.element().Ref)

	// the reference root
	target, err = resolveElementTarget(page, "content.1", true)
	assert.NoError(t, err)
	assert.Equal(t, "slider", target.ref)
	assert.Empty(t, target.path)
	assert.Equal(t, "content.1", target.prefix)
	assert.Equal(t, "slider", target.element().Tpl)

	// child of the reference
	target, err = resolveElementTarget(page, "content.1.slides.0", false)
	assert.NoError(t, err)
	assert.Equal(t, "slider", target.ref)
	assert.Equal(t, "slides.0", target.path)
	assert.Equal(t, "slide", target.element().Tpl)
	assert.Equal(t, "content.1.slides.0", target.fullPath(target.path))

	// child of a nested reference
	target, err = resolveElementTarget(page, "content.1.slides.1.content.0", true)
	assert.NoError(t, err)
	assert.Equal(t, "inner", target.ref)
	assert.Equal(t, "content.0", target.path)
	assert.Equal(t, "img", target.element().Tpl)

	// out of range
	_, err = resolveElementTarget(page, "content.1.slides.5", false)
	assert.Error(t, err)

	assert.True(t, refContains("slider", "inner"))
	assert.False(t, refContains("inner", "slider"))
}
