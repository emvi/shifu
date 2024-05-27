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

	Server ServerConfig `json:"server"`
	Sass   SassConfig   `json:"sass"`
	JS     JSConfig     `json:"js"`
	Pirsch PirschConfig `json:"pirsch"`
}

// ServerConfig is the HTTP server configuration.
type ServerConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	ShutdownTimeout int    `json:"shutdown_time"`
	WriteTimeout    int    `json:"write_timeout"`
	ReadTimeout     int    `json:"read_timeout"`
	TLSCertFile     string `json:"tls_cert_file"`
	TLSKeyFile      string `json:"tls_key_file"`
}

// SassConfig is the sass compiler configuration.
type SassConfig struct {
	Entrypoint   string `json:"entrypoint"`
	Dir          string `json:"dir"`
	Watch        bool   `json:"watch"`
	Out          string `json:"out"`
	OutSourceMap string `json:"out_source_map"`
}

// JSConfig is the JavaScript compiler configuration.
type JSConfig struct {
	Entrypoint string `json:"entrypoint"`
	Dir        string `json:"dir"`
	Watch      bool   `json:"watch"`
	Out        string `json:"out"`
	SourceMap  bool   `json:"source_map"`
}

// PirschConfig is the configuration for pirsch.io.
type PirschConfig struct {
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
