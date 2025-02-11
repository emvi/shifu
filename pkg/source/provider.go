package source

import (
	"context"
	"time"
)

// Provider is the interface for a website data provider.
type Provider interface {
	// Watch regularly updates the website data.
	Watch(context.Context, func())

	// Update updates the website data once.
	Update(func())

	// LastUpdate returns the time the website data has last been updated.
	LastUpdate() time.Time
}
