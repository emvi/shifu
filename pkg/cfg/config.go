package cfg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
)

const (
	configFile = "config.json"
)

var (
	cfg Config
)

// Config is the Oogway application config.
type Config struct {
	BaseDir string           `json:"-"`
	FuncMap template.FuncMap `json:"-"`

	Dev       bool      `json:"dev"`
	Server    Server    `json:"server"`
	CORS      CORS      `json:"cors"`
	Sass      Sass      `json:"sass"`
	JS        JS        `json:"js"`
	Analytics Analytics `json:"analytics"`
}

// Server is the HTTP server configuration.
type Server struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	ShutdownTimeout  int    `json:"shutdown_time"`
	WriteTimeout     int    `json:"write_timeout"`
	ReadTimeout      int    `json:"read_timeout"`
	TLSCertFile      string `json:"tls_cert_file"`
	TLSKeyFile       string `json:"tls_key_file"`
	Hostname         string `json:"hostname"`
	SecureCookies    bool   `json:"secure_cookies"`
	CookieDomainName string `json:"cookie_domain_name"`
}

// CORS is the HTTP CORS configuration.
type CORS struct {
	Origins  string `json:"origins"`
	Loglevel string `json:"loglevel"`
}

// Sass is the sass compiler configuration.
type Sass struct {
	Entrypoint   string `json:"entrypoint"`
	Dir          string `json:"dir"`
	Watch        bool   `json:"watch"`
	Out          string `json:"out"`
	OutSourceMap string `json:"out_source_map"`
}

// JS is the JavaScript compiler configuration.
type JS struct {
	Entrypoint string `json:"entrypoint"`
	Dir        string `json:"dir"`
	Watch      bool   `json:"watch"`
	Out        string `json:"out"`
	SourceMap  bool   `json:"source_map"`
}

// Analytics is the web analytics configuration.
type Analytics struct {
	Provider     string   `json:"provider"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Subnets      []string `json:"subnets"`
	Header       []string `json:"header"`
}

// Load loads the configuration from config.json.
func Load(dir string, funcMap template.FuncMap) error {
	content, err := os.ReadFile(filepath.Join(dir, configFile))

	if err != nil {
		return fmt.Errorf("error loading config.json: %s", err)
	}

	if err := json.Unmarshal(content, &cfg); err != nil {
		return fmt.Errorf("error loading config.json: %s", err)
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}

	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30
	}

	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 5
	}

	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 5
	}

	cfg.BaseDir = dir
	cfg.FuncMap = funcMap
	return nil
}

// Watch watches config.json for changes and automatically reloads the settings.
func Watch(ctx context.Context, dir string, funcMap template.FuncMap) error {
	if err := Load(dir, funcMap); err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					continue
				}

				if event.Op == fsnotify.Write {
					if err := Load(dir, funcMap); err != nil {
						slog.Error("Error updating config.json", "err", err)
					}
				}
			case <-ctx.Done():
				if err := watcher.Close(); err != nil {
					slog.Error("Error closing config.json watcher", "error", err)
				}

				return
			}
		}
	}()

	if err := watcher.Add(filepath.Join(dir, configFile)); err != nil {
		return err
	}

	return nil
}

// Get returns the configuration.
// This is not thread safe! In production, the watcher must be disabled!
func Get() *Config {
	return &cfg
}
