package tpl

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testDir = "test"
)

func TestMerge(t *testing.T) {
	a := template.FuncMap{
		"a": func() int { return 42 },
	}
	b := template.FuncMap{
		"a": func() int { return 43 },
		"b": func() int { return 44 },
	}
	out := Merge(a, b)
	assert.NotNil(t, out["a"])
	assert.NotNil(t, out["b"])
	assert.Nil(t, out["c"])
}

func TestRenderMarkdown(t *testing.T) {
	assert.NoError(t, os.RemoveAll(testDir))
	assert.NoError(t, os.Mkdir(testDir, 0777))
	file := filepath.Join(testDir, "test.md")
	assert.NoError(t, os.WriteFile(file, []byte("# Test {{.Var}}"), 0777))
	time.Sleep(time.Millisecond * 10)
	out := renderMarkdown(file, struct {
		Var string
	}{
		"var",
	})
	assert.Equal(t, template.HTML("<h1>Test var</h1>\n"), out)
}

func TestRenderMarkdownBlock(t *testing.T) {
	assert.NoError(t, os.RemoveAll(testDir))
	assert.NoError(t, os.Mkdir(testDir, 0777))
	file := filepath.Join(testDir, "test.md")
	assert.NoError(t, os.WriteFile(file, []byte(`{{define "foo"}}# Test {{.Var}}{{end}}{{define "bar"}}## Hello World {{.Var}}{{end}}`), 0777))
	time.Sleep(time.Millisecond * 10)
	out := renderMarkdownBlock(file, "bar", struct {
		Var string
	}{
		"var",
	})
	assert.Equal(t, template.HTML("<h2>Hello World var</h2>\n"), out)
}
