package js

import (
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/storage"
	esbuild "github.com/evanw/esbuild/pkg/api"
	"log/slog"
	"os"
	"path/filepath"
)

// Compile compiles the entrypoint JS/TS for given base directory.
func Compile(dir string, store storage.Storage) {
	in := filepath.Join(dir, cfg.Get().JS.Dir, cfg.Get().JS.Entrypoint)
	slog.Info("Compiling js file", "file", in)
	sourceMap := esbuild.SourceMapNone

	if cfg.Get().JS.SourceMap {
		sourceMap = esbuild.SourceMapExternal
	}

	if err := os.MkdirAll(filepath.Join(dir, "tmp", filepath.Dir(cfg.Get().JS.Out)), 0744); err != nil {
		slog.Error("Error creating js output directory", "error", err)
		return
	}

	out := filepath.Join(dir, "tmp", cfg.Get().JS.Out)
	result := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:       []string{in},
		Outfile:           out,
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

	data, err := os.ReadFile(out)

	if err != nil {
		slog.Error("Error reading temporary js output", "error", err)
		return
	}

	if _, err := store.Write(filepath.Join(dir, cfg.Get().JS.Out), data, nil); err != nil {
		slog.Error("Error saving js file", "error", err)
	}

	if cfg.Get().JS.SourceMap {
		data, err = os.ReadFile(out + ".map")

		if err != nil {
			slog.Error("Error reading temporary js source map output", "error", err)
			return
		}

		if _, err := store.Write(filepath.Join(dir, cfg.Get().JS.Out)+".map", data, nil); err != nil {
			slog.Error("Error saving js source map file", "error", err)
		}
	}
}
