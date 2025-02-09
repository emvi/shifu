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

func (provider *S3) pull() bool {
	files, err := provider.storage.List("content", true)

	if err != nil {
		return false
	}

	for _, file := range files {
		file = strings.TrimPrefix(file, provider.pathPrefix)
		withoutPrefix := strings.TrimPrefix(file, "/content/")
		dir := filepath.Join(provider.dir, "content", filepath.Dir(withoutPrefix))
		out := filepath.Join(provider.dir, "content", withoutPrefix)

		// check if file needs to be updated
		info, err := provider.storage.Stat(file)

		if err == nil {
			stat, err := os.Stat(out)

			if err == nil && stat != nil &&
				(stat.ModTime().Equal(info.ModTime()) || stat.ModTime().After(info.ModTime())) {
				continue
			}
		}

		// create directory if required
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0744); err != nil {
				slog.Error("Error creating directory", "error", err, "dir", dir)
				continue
			}
		}

		// download and write to disk
		data, err := provider.storage.Read(file)

		if err != nil {
			slog.Error("Error reading content from S3", "error", err, "file", file)
			continue
		}

		if err := os.WriteFile(out, data, 0644); err != nil {
			slog.Error("Error writing S3 content to disk", "error", err, "file", out)
		}
	}

	return false
}
