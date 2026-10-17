package content

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReferenceTree(t *testing.T) {
	tplCfgCache = NewTemplateCfgCache() // loads the test configuration from ./admin/tpl
	wd, err := os.Getwd()
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Join("content", "refs"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join("content", "pages"), 0755))
	defer os.RemoveAll(filepath.Join(wd, "content"))
	assert.NoError(t, os.WriteFile(filepath.Join("content", "refs", "slider.json"), []byte(`{"tpl":"text","copy":{"de":{"headline":"Hallo"}},"content":{"content":[{"tpl":"img"},{"tpl":"img"}]}}`), 0644))
	assert.NoError(t, os.WriteFile(filepath.Join("content", "pages", "home.json"), []byte(`{"path":{"de":"/","en":"/en"},"content":{"content":[{"ref":"slider"}]}}`), 0644))

	page, err := loadDocument(httptest.NewRequest("GET", "/", nil), getPagePath("content/refs/slider.json"), "")
	assert.NoError(t, err)
	assert.True(t, isVirtualDocument(page))
	assert.Equal(t, "slider", page.Content["content"][0].Ref)
	assert.Len(t, getPageLanguages(page), 2)

	// the reference can be edited, the virtual page cannot be saved
	target, err := resolveElementTarget(page, "content.0", true)
	assert.NoError(t, err)
	assert.Equal(t, "slider", target.ref)
	target, err = resolveElementTarget(page, "content.0", false)
	assert.NoError(t, err)
	assert.Error(t, target.save())

	w := httptest.NewRecorder()
	Reference(w, httptest.NewRequest("GET", "/refs/ref?path=/slider.json", nil))
	body := w.Body.String()
	assert.Equal(t, 200, w.Code)
	assert.True(t, strings.Contains(body, `id="shifu-ref-tree"`))
	assert.True(t, strings.Contains(body, `element=content.0.content.1`))
	assert.True(t, strings.Contains(body, `hx-target="#shifu-ref-tree"`))
	assert.False(t, strings.Contains(body, `content/element/delete?path=content/refs/slider.json&element=content.0&`))
	assert.False(t, strings.Contains(body, "shifuMoveElement("))
	assert.True(t, strings.Contains(body, `content/element/delete?path=content/refs/slider.json&element=content.0.content.0&`), body)
}
