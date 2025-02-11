package sass

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"path/filepath"
	"strings"
)

// Watch watches the entrypoint Sass for changes and recompiles if required.
func Watch(ctx context.Context) error {
	if cfg.Get().Sass.Entrypoint != "" {
		dir := cfg.Get().BaseDir
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
