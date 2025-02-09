package source

import (
	"context"
	"sync"
	"time"
)

// FileStorage loads the website data from the file system.
type FileStorage struct {
	dir           string
	updateSeconds int
	lastUpdate    time.Time
	m             sync.RWMutex
}

// NewFileStorage creates a new Provider for the file system.
func NewFileStorage(dir string, updateSeconds int) *FileStorage {
	if updateSeconds == 0 {
		updateSeconds = 60
	}

	return &FileStorage{
		dir:           dir,
		updateSeconds: updateSeconds,
	}
}

// Update implements the Provider interface.
func (provider *FileStorage) Update(ctx context.Context, update func()) {
	go func() {
		timerDuration := time.Second * time.Duration(provider.updateSeconds)
		timer := time.NewTimer(timerDuration)
		defer timer.Stop()

		for {
			timer.Reset(timerDuration)

			select {
			case <-timer.C:
				provider.m.Lock()
				update()
				provider.lastUpdate = time.Now().UTC()
				provider.m.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	update()
}

// LastUpdate implements the Provider interface.
func (provider *FileStorage) LastUpdate() time.Time {
	provider.m.RLock()
	defer provider.m.RUnlock()
	return provider.lastUpdate
}
