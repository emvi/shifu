package source

import (
	"context"
	"github.com/emvi/shifu/pkg/storage"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// S3 loads the website data from a S3 bucket.
// It will only load the content once. Afterward, the content needs to be pulled manually.
type S3 struct {
	storage    storage.Storage
	dir        string
	pathPrefix string
	lastUpdate time.Time
	m          sync.RWMutex
}

// NewS3 creates a new Provider for S3.
func NewS3(storage storage.Storage, dir, pathPrefix string) *S3 {
	provider := &S3{
		storage:    storage,
		dir:        dir,
		pathPrefix: pathPrefix,
	}
	provider.pull()
	return provider
}

// Update implements the Provider interface.
func (provider *S3) Update(_ context.Context, update func()) {
	provider.m.Lock()
	update()
	provider.lastUpdate = time.Now().UTC()
	provider.m.Unlock()
}

// LastUpdate implements the Provider interface.
func (provider *S3) LastUpdate() time.Time {
	provider.m.RLock()
	defer provider.m.RUnlock()
	return provider.lastUpdate
}

// TODO optimize
func (provider *S3) pull() bool {
	files, err := provider.storage.List("content", true)

	if err != nil {
		return false
	}

	for _, file := range files {
		file = strings.TrimPrefix(file, provider.pathPrefix)
		data, err := provider.storage.Read(file)

		if err != nil {
			slog.Error("Error reading content from S3", "error", err, "file", file)
			continue
		}

		file = strings.TrimPrefix(file, "/content/")
		dir := filepath.Join(provider.dir, "content", filepath.Dir(file))

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0744); err != nil {
				slog.Error("Error creating directory", "error", err, "dir", dir)
				continue
			}
		}

		if err := os.WriteFile(filepath.Join(provider.dir, "content", file), data, 0644); err != nil {
			slog.Error("Error writing S3 content to disk", "error", err, "file", file)
		}
	}

	return false
}
