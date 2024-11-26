package analytics

import (
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

var (
	client Analytics
)

// Init initializes the analytics client if configured.
func Init() {
	provider := strings.ToLower(cfg.Get().Analytics.Provider)

	if provider == "pirsch" && cfg.Get().Analytics.ClientSecret != "" {
		slog.Info("Using analytics provider Pirsch")
		client = NewPirschAnalytics(cfg.Get().Analytics.ClientID, cfg.Get().Analytics.ClientSecret)
	} else {
		slog.Info("No analytics provider found")
	}

	loadIPHeader()
	loadSubnets()
}

// PageView sends a page view using the configured Analytics provider.
func PageView(r *http.Request, tags map[string]string) {
	go func() {
		if client != nil {
			if err := client.PageView(r, tags); err != nil {
				slog.Error("Error sending page view", "error", err)
			}
		}
	}()
}

// Event sends a custom event using the configured Analytics provider.
func Event(r *http.Request, name string, duration int, meta, tags map[string]string) {
	go func() {
		if client != nil {
			if err := client.Event(r, name, duration, meta, tags); err != nil {
				slog.Error("Error sending custom event", "error", err)
			}
		}
	}()
}

func loadIPHeader() {
	for _, header := range cfg.Get().Analytics.Header {
		found := false

		for _, parser := range allIPHeader {
			if strings.ToLower(header) == strings.ToLower(parser.Header) {
				ipHeader = append(ipHeader, parser)
				found = true
				break
			}
		}

		if !found {
			slog.Error("Header invalid", "header", header)
			panic(errors.New(fmt.Sprintf("header invalid: %s", header)))
		}
	}
}

func loadSubnets() {
	for _, subnet := range cfg.Get().Analytics.Subnets {
		_, n, err := net.ParseCIDR(subnet)

		if err != nil {
			slog.Error("Error parsing subnet", "subnet", subnet, "error", err)
			panic(errors.New(fmt.Sprintf("Error parsing subnet '%s': %s", subnet, err)))
		}

		allowedSubnets = append(allowedSubnets, *n)
	}
}
