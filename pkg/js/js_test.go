package js

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	inDir  = "in"
	outDir = "out"
)

func TestCompile(t *testing.T) {
	assert.NoError(t, os.RemoveAll(inDir))
	assert.NoError(t, os.Mkdir(inDir, 0777))
	in := filepath.Join(inDir, "test.js")
	assert.NoError(t, os.WriteFile(in, []byte("console.log('Hello World')"), 0777))
	time.Sleep(time.Millisecond * 10)
	out := filepath.Join(outDir, "bundle.js")
	cfg.Get().JS.Dir = inDir
	cfg.Get().JS.Entrypoint = "test.js"
	cfg.Get().JS.Out = filepath.Join(outDir, "bundle.js")
	cfg.Get().JS.SourceMap = true
	Compile("")
	assert.FileExists(t, out)
	assert.FileExists(t, filepath.Join(outDir, "bundle.js.map"))
	content, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Equal(t, "(()=>{console.log(\"Hello World\");})();\n", string(content))
}

func TestWatch(t *testing.T) {
	assert.NoError(t, os.RemoveAll(inDir))
	assert.NoError(t, os.Mkdir(inDir, 0777))
	in := filepath.Join(inDir, "test.js")
	assert.NoError(t, os.WriteFile(in, []byte("console.log('Hello World')"), 0777))
	time.Sleep(time.Millisecond * 10)
	out := filepath.Join(outDir, "bundle.js")
	cfg.Get().JS.Dir = inDir
	cfg.Get().JS.Entrypoint = "test.js"
	cfg.Get().JS.Out = filepath.Join(outDir, "bundle.js")
	cfg.Get().JS.Watch = true
	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, Watch(ctx))
	time.Sleep(time.Millisecond * 10)
	assert.FileExists(t, out)
	content, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Equal(t, "(()=>{console.log(\"Hello World\");})();\n", string(content))
	assert.NoError(t, os.WriteFile(in, []byte("console.log('Foo bar')"), 0777))
	time.Sleep(time.Millisecond * 10)
	assert.FileExists(t, out)
	content, err = os.ReadFile(out)
	assert.NoError(t, err)
	assert.Equal(t, "(()=>{console.log(\"Foo bar\");})();\n", string(content))
	cancel()
}
