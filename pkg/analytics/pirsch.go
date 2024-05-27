package analytics

import (
	"github.com/emvi/shifu/pkg/cfg"
	pirsch "github.com/pirsch-analytics/pirsch-go-sdk/v2/pkg"
	"log"
	"net"
	"net/http"
	"strings"
)

var (
	pirschClient *pirsch.Client
)

// Init initializes the analytics client if configured.
func Init() {
	if cfg.Get().Pirsch.ClientSecret != "" {
		pirschClient = pirsch.NewClient(cfg.Get().Pirsch.ClientID, cfg.Get().Pirsch.ClientSecret, nil)
		loadIPHeader()
		loadSubnets()
	}
}

// PageView tracks a page view for given request.
func PageView(r *http.Request, path string) {
	if pirschClient != nil {
		url := r.URL

		if path != "" {
			url.Path = path
		}

		if err := pirschClient.PageView(r, &pirsch.PageViewOptions{
			IP:  getIP(r),
			URL: url.String(),
		}); err != nil {
			log.Printf("Error sending page view to Pirsch: %s", err)
		}
	}
}

// Event tracks a custom event for given request.
func Event(r *http.Request, path, name string, duration int, meta map[string]string) {
	if pirschClient != nil {
		url := r.URL

		if path != "" {
			url.Path = path
		}

		if err := pirschClient.Event(name, duration, meta, r, &pirsch.PageViewOptions{
			IP:  getIP(r),
			URL: url.String(),
		}); err != nil {
			log.Printf("Error sending page view to Pirsch: %s", err)
		}
	}
}

func loadIPHeader() {
	for _, header := range cfg.Get().Pirsch.Header {
		found := false

		for _, parser := range allIPHeader {
			if strings.ToLower(header) == strings.ToLower(parser.Header) {
				ipHeader = append(ipHeader, parser)
				found = true
				break
			}
		}

		if !found {
			log.Fatalf("Header invalid: %s", header)
		}
	}
}

func loadSubnets() {
	for _, subnet := range cfg.Get().Pirsch.Subnets {
		_, n, err := net.ParseCIDR(subnet)

		if err != nil {
			log.Fatalf("Error parsing subnet '%s': %s", subnet, err)
		}

		allowedSubnets = append(allowedSubnets, *n)
	}
}
