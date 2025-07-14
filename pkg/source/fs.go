package source

import (
	"context"
	"sync"
	"time"
)

// FileSystem loads the website data from the file system.
type FileSystem struct {
	dir           string
	updateSeconds int
	lastUpdate    time.Time
	m             sync.RWMutex
}

// NewFileSystem creates a new Provider for the file system.
func NewFileSystem(dir string, updateSeconds int) *FileSystem {
	if updateSeconds == 0 {
		updateSeconds = 60
	}

	return &FileSystem{
		dir:           dir,
		updateSeconds: updateSeconds,
	}
}

// Watch implements the Provider interface.
func (provider *FileSystem) Watch(ctx context.Context, update func()) {
	go func() {
		timerDuration := time.Second * time.Duration(provider.updateSeconds)
		timer := time.NewTimer(timerDuration)
		defer timer.Stop()

		for {
			timer.Reset(timerDuration)

			select {
			case <-timer.C:
				provider.Update(update)
			case <-ctx.Done():
				return
			}
		}
	}()

	update()
}

// Update implements the Provider interface.
func (provider *FileSystem) Update(update func()) {
	provider.m.Lock()
	defer provider.m.Unlock()
	update()
	provider.lastUpdate = time.Now().UTC()
}

// LastUpdate implements the Provider interface.
func (provider *FileSystem) LastUpdate() time.Time {
	provider.m.RLock()
	defer provider.m.RUnlock()
	return provider.lastUpdate
}
