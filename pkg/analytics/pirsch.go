package analytics

import (
	pirsch "github.com/pirsch-analytics/pirsch-go-sdk/v2/pkg"
	"net/http"
)

// PirschAnalytics is the analytics provider for Pirsch (pirsch.io).
type PirschAnalytics struct {
	client *pirsch.Client
}

// NewPirschAnalytics creates a new Analytics provider for Pirsch.
func NewPirschAnalytics(clientID, clientSecret string) *PirschAnalytics {
	return &PirschAnalytics{
		client: pirsch.NewClient(clientID, clientSecret, nil),
	}
}

// PageView implements the Analytics interface.
func (provider *PirschAnalytics) PageView(r *http.Request, tags map[string]string) error {
	if err := provider.client.PageView(r, &pirsch.PageViewOptions{
		IP:   getIP(r),
		Tags: tags,
	}); err != nil {
		return err
	}

	return nil
}

// Event implements the Analytics interface.
func (provider *PirschAnalytics) Event(r *http.Request, name string, duration int, meta, tags map[string]string) error {
	if err := provider.client.Event(name, duration, meta, r, &pirsch.PageViewOptions{
		IP:   getIP(r),
		Tags: tags,
	}); err != nil {
		return err
	}

	return nil
}
