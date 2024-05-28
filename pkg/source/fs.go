package source

import (
	"context"
	"sync"
	"time"
)

// FS loads the website data from the file system.
type FS struct {
	dir           string
	updateSeconds int
	lastUpdate    time.Time
	m             sync.RWMutex
}

// NewFS creates a new Provider for the file system.
func NewFS(dir string, updateSeconds int) *FS {
	if updateSeconds == 0 {
		updateSeconds = 15
	}

	return &FS{
		dir:           dir,
		updateSeconds: updateSeconds,
	}
}

// Update implements the Provider interface.
func (provider *FS) Update(ctx context.Context, update func()) {
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
}

// LastUpdate implements the Provider interface.
func (provider *FS) LastUpdate() time.Time {
	provider.m.RLock()
	defer provider.m.RUnlock()
	return provider.lastUpdate
}
