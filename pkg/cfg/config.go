package cfg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	LogLevel  string    `json:"log_level"`
	Server    Server    `json:"server"`
	API       API       `json:"api"`
	Remote    Remote    `json:"remote"`
	Git       Git       `json:"git"`
	Content   Content   `json:"content"`
	CORS      CORS      `json:"cors"`
	Sass      Sass      `json:"sass"`
	JS        JS        `json:"js"`
	Analytics Analytics `json:"analytics"`
	UI        UI        `json:"ui"`
}

// Server is the HTTP server configuration.
type Server struct {
	Host             string `json:"host"`
	Port             int    `json:"port"`
	ShutdownTimeout  int    `json:"shutdown_time"`
	WriteTimeout     int    `json:"write_timeout"`
	ReadTimeout      int    `json:"read_timeout"`
	IdleTimeout      int    `json:"idle_timeout"`
	TLSCertFile      string `json:"tls_cert_file"`
	TLSKeyFile       string `json:"tls_key_file"`
	Hostname         string `json:"hostname"`
	SecureCookies    bool   `json:"secure_cookies"`
	CookieDomainName string `json:"cookie_domain_name"`
}

// API is the server API configuration.
type API struct {
	Secret string `json:"secret"`
}

// Remote is the configuration for a remote Shifu server.
type Remote struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// Git is the Git repository configuration for auto-updates.
type Git struct {
	UpdateSeconds int    `json:"update_seconds"`
	Repository    string `json:"repository"`
}

// Content is the content and source configuration.
type Content struct {
	Provider string            `json:"provider"`
	NotFound map[string]string `json:"not_found"`
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

// UI configures the admin user interface.
type UI struct {
	Path          string `json:"path"`
	AdminPassword string `json:"admin_password"`
}

// Load loads the configuration from config.json.
func Load(dir string, funcMap template.FuncMap) error {
	content, err := os.ReadFile(filepath.Join(dir, configFile))

	if err != nil {
		return fmt.Errorf("error loading config.json: %s", err)
	}

	contentStr := string(content)
	secrets := loadSecrets(dir)

	for key, value := range secrets {
		contentStr = strings.ReplaceAll(contentStr, fmt.Sprintf("${%s}", key), value)
	}

	content = []byte(contentStr)

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
		cfg.Server.WriteTimeout = 60
	}

	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 60
	}

	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 60
	}

	if cfg.CORS.Origins == "" {
		cfg.CORS.Origins = "*"
	}

	if cfg.Content.Provider == "" {
		cfg.Content.Provider = "fs"
	}

	cfg.BaseDir = dir
	cfg.FuncMap = funcMap
	return nil
}

func loadSecrets(dir string) map[string]string {
	vars, err := os.Open(filepath.Join(dir, ".secrets.env"))

	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		slog.Error("Error opening secrets file: ", "error", err)
		return nil
	}

	defer func() {
		if err := vars.Close(); err != nil {
			slog.Error("Error closing secrets file: ", "error", err)
		}
	}()
	scanner := bufio.NewScanner(vars)
	scanner.Split(bufio.ScanLines)
	substitute := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		key, value, found := strings.Cut(line, "=")

		if found {
			substitute[key] = value
		}
	}

	return substitute
}

// Get returns the configuration.
// This is not thread safe! In production, the watcher must be disabled!
func Get() *Config {
	return &cfg
}
