package js

import (
	"github.com/emvi/shifu/pkg/cfg"
	esbuild "github.com/evanw/esbuild/pkg/api"
	"log/slog"
	"os"
	"path/filepath"
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
