package content

import (
	"context"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/fsnotify/fsnotify"
	"log/slog"
	"path/filepath"
	"strings"
)

var (
	tplCache *TemplateCache
)

// Init initializes the template cache and watches for configuration changes.
func Init(ctx context.Context) error {
	tplCache = NewTemplateCache()
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

				if event.Op == fsnotify.Write && strings.ToLower(filepath.Ext(event.Name)) == ".json" {
					tplCache.Load()
				}
			case <-ctx.Done():
				if err := watcher.Close(); err != nil {
					slog.Error("Error closing watcher", "error", err)
				}

				return
			}
		}
	}()

	if err := watcher.Add(filepath.Join(cfg.Get().BaseDir, templateConfigDir)); err != nil {
		return err
	}

	return nil
}
