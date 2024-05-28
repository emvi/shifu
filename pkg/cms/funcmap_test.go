package cms

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

func TestShuffle(t *testing.T) {
	list := shuffle([]any{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5)
	assert.Len(t, list, 5)
}

func TestFormatFloat(t *testing.T) {
	input := []float64{
		0,
		1,
		42,
		98.7,
		12.56,
		99.1949,
		123.4567,
		1234.56789,
	}
	output := []string{
		"0.00",
		"1.00",
		"42.00",
		"98.70",
		"12.56",
		"99.19",
		"123.46",
		"1234.57",
	}

	for i, in := range input {
		assert.Equal(t, output[i], formatFloat(in))
	}
}

func TestFormatInt(t *testing.T) {
	input := []int{
		-1,
		0,
		1,
		42,
		987,
		1000,
		1256,
		991949,
		1234567,
		123456789,
	}
	output := []string{
		"-1",
		"0",
		"1",
		"42",
		"987",
		"1,000",
		"1,256",
		"991,949",
		"1,234,567",
		"123,456,789",
	}

	for i, in := range input {
		assert.Equal(t, output[i], formatInt(in))
	}
}
