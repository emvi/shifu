package js

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	esbuild "github.com/evanw/esbuild/pkg/api"
	"github.com/fsnotify/fsnotify"
)

// Compile compiles the entrypoint JS/TS for given base directory.
func Compile(dir string) {
	in := filepath.Join(dir, cfg.Get().JS.Dir, cfg.Get().JS.Entrypoint)
	slog.Info("Compiling js file", "file", in)
	sourceMap := esbuild.SourceMapNone

	if cfg.Get().JS.SourceMap {
		sourceMap = esbuild.SourceMapExternal
	}

	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(cfg.Get().JS.Out)), 0744); err != nil {
		slog.Error("Error creating js output directory", "error", err)
		return
	}

	result := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:       []string{in},
		Outfile:           filepath.Join(dir, cfg.Get().JS.Out),
		Sourcemap:         sourceMap,
		Bundle:            true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Write:             true,
	})

	if len(result.Errors) > 0 {
		slog.Error("Error compiling js", "errors", result.Errors)
	}
}

// Watch watches the entrypoint JS/TS for changes and recompiles if required.
func Watch(ctx context.Context, dir string) error {
	if cfg.Get().JS.Entrypoint != "" {
		Compile(dir)

		if cfg.Get().JS.Watch {
			watcher, err := fsnotify.NewWatcher()

			if err != nil {
				return err
			}

			go func() {
				out := filepath.Join(dir, cfg.Get().JS.Out)

				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							continue
						}

						if event.Op == fsnotify.Write && event.Name != out {
							ext := strings.ToLower(filepath.Ext(event.Name))

							if ext == ".js" || ext == ".ts" || ext == ".tsx" || ext == ".mts" || ext == ".cts" {
								Compile(dir)
							}
						}
					case <-ctx.Done():
						if err := watcher.Close(); err != nil {
							slog.Error("Error closing watcher", "error", err)
						}

						return
					}
				}
			}()

			if err := watcher.Add(filepath.Join(dir, cfg.Get().JS.Dir)); err != nil {
				return err
			}
		}
	}

	return nil
}
