package sass

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/storage"
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

func TestCompileSass(t *testing.T) {
	fs := storage.NewFileStorage()
	assert.NoError(t, os.RemoveAll(inDir))
	assert.NoError(t, os.Mkdir(inDir, 0777))
	in := filepath.Join(inDir, "test.scss")
	assert.NoError(t, os.WriteFile(in, []byte(".class{&-name{color:red}}"), 0777))
	time.Sleep(time.Millisecond * 100)
	out := filepath.Join(outDir, "test.css")
	cfg.Get().Sass.Dir = inDir
	cfg.Get().Sass.Entrypoint = "test.scss"
	cfg.Get().Sass.Out = filepath.Join(outDir, "test.css")
	cfg.Get().Sass.OutSourceMap = filepath.Join(outDir, "test.css.map")
	Compile("", fs)
	assert.FileExists(t, out)
	assert.FileExists(t, cfg.Get().Sass.OutSourceMap)
	content, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Contains(t, string(content), ".class-name{color:red}")
}

func TestWatchSass(t *testing.T) {
	fs := storage.NewFileStorage()
	assert.NoError(t, os.RemoveAll(inDir))
	assert.NoError(t, os.Mkdir(inDir, 0777))
	in := filepath.Join(inDir, "test.scss")
	assert.NoError(t, os.WriteFile(in, []byte(".class{&-name{color:red}}"), 0777))
	time.Sleep(time.Millisecond * 100)
	out := filepath.Join(outDir, "test.css")
	cfg.Get().Sass.Dir = inDir
	cfg.Get().Sass.Entrypoint = "test.scss"
	cfg.Get().Sass.Out = filepath.Join(outDir, "test.css")
	cfg.Get().Sass.Watch = true
	ctx, cancel := context.WithCancel(context.Background())
	assert.NoError(t, Watch(ctx, "", fs))
	time.Sleep(time.Millisecond * 100)
	assert.FileExists(t, out)
	content, err := os.ReadFile(out)
	assert.NoError(t, err)
	assert.Contains(t, string(content), ".class-name{color:red}")
	assert.NoError(t, os.WriteFile(in, []byte(".class{&-name{color:blue}}"), 0777))
	time.Sleep(time.Second)
	assert.FileExists(t, out)
	content, err = os.ReadFile(out)
	assert.NoError(t, err)
	assert.Contains(t, string(content), ".class-name{color:blue}")
	cancel()
}
