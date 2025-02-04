package cfg

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"html/template"
	"log/slog"
	"path/filepath"
)

// Watch watches config.json for changes and automatically reloads the settings.
func Watch(ctx context.Context, dir string, funcMap template.FuncMap) error {
	if err := Load(dir, funcMap); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				if event.Op == fsnotify.Write {
					if err := Load(dir, funcMap); err != nil {
						slog.Error("Error updating config.json", "error", err)
					}
				}
			case <-ctx.Done():
				if err := watcher.Close(); err != nil {
					slog.Error("Error closing config.json watcher", "error", err)
				}

				return
			}
		}
	}()

	if err := watcher.Add(filepath.Join(dir, configFile)); err != nil {
		return err
	}

	return nil
}
