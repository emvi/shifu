package js

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"path/filepath"
	"strings"
)

// Watch watches the entrypoint JS/TS for changes and recompiles if required.
func Watch(ctx context.Context) error {
	if cfg.Get().JS.Entrypoint != "" {
		dir := cfg.Get().BaseDir
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
