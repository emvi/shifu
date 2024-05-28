package analytics

import (
	"net/http"
)

// Analytics is the interface for an analytics provider.
type Analytics interface {
	// PageView sends a page view for given request and optional tags.
	PageView(*http.Request, map[string]string) error

	// Event sends a custom event for given request, name, duration, meta data, and optional tags.
	Event(*http.Request, string, int, map[string]string, map[string]string) error
}
