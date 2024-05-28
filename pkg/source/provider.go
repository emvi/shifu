package source

import (
	"context"
	"time"
)

// Provider is the interface for a website data provider.
type Provider interface {
	// Update regularly updates the website data from Git.
	Update(context.Context, func())

	// LastUpdate returns the time the website data has last been updated.
	LastUpdate() time.Time
}
