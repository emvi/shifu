package sass

import (
	"context"
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/fsnotify/fsnotify"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compile compiles the entrypoint Sass for given base directory.
func Compile(dir string) {
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(cfg.Get().Sass.Out)), 0744); err != nil {
		slog.Error("Error creating css output directory", "error", err)
		return
	}

	in := filepath.Join(dir, cfg.Get().Sass.Dir, cfg.Get().Sass.Entrypoint)
	out := filepath.Join(dir, cfg.Get().Sass.Out)
	slog.Info("Compiling sass file", "in", in, "out", out)
	dirs, err := getDirs(filepath.Join(dir, cfg.Get().Sass.Dir))

	if err != nil {
		slog.Error("Error reading sass directory", "error", err)
		return
	}

	args := make([]string, 0)

	for _, d := range dirs {
		args = append(args, fmt.Sprintf("--load-path=%s", d))
	}

	if cfg.Get().Sass.OutSourceMap == "" {
		args = append(args, "--no-source-map")
	} else {
		args = append(args, "--source-map")
	}

	args = append(args, "--style=compressed")
	args = append(args, in)
	args = append(args, out)
	cmd := exec.Command("sass", args...)

	if err := cmd.Run(); err != nil {
		slog.Error("Error compiling sass", "error", err)
		return
	}
}

// Watch watches the entrypoint Sass for changes and recompiles if required.
func Watch(ctx context.Context, dir string) error {
	if cfg.Get().Sass.Entrypoint != "" {
		Compile(dir)

		if cfg.Get().Sass.Watch {
			watcher, err := fsnotify.NewWatcher()

			if err != nil {
				return err
			}

			go func() {
				out := filepath.Join(dir, cfg.Get().Sass.Out)

				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							continue
						}

						if event.Op == fsnotify.Write && event.Name != out {
							ext := strings.ToLower(filepath.Ext(event.Name))

							if ext == ".scss" || ext == ".sass" {
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

			if err := watcher.Add(filepath.Join(dir, cfg.Get().Sass.Dir)); err != nil {
				return err
			}
		}
	}

	return nil
}

func getDirs(root string) ([]string, error) {
	dirs := make([]string, 0)

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			dirs = append(dirs, path)
		}

		return err
	}); err != nil {
		return nil, err
	}

	return dirs, nil
}
