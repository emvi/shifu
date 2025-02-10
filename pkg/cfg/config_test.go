package cfg

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

const (
	sampleConfig = `{
	"dev": true,
	"log_level": "debug",
    "server": {
        "host": "localhost",
        "port": 8080,
        "shutdown_time": 9,
        "write_timeout": 10,
        "read_timeout": 11,
        "tls_cert_file": "cert/file.pem",
        "tls_key_file": "key/file.pem",
		"hostname": "example.com",
		"secure_cookies": true,
		"cookie_domain_name": "example.com"
    },
	"git": {
	    "update_seconds": 600,
		"repository": "https//github.com/foo/bar"
	},
	"content": {
		"provider": "git",
		"not_found": {
			"en": "/not-found"
		}
	},
	"cors": {
		"origins": "*",
		"loglevel": "info"
	},
    "sass": {
        "entrypoint": "style.scss",
        "dir": "assets/scss",
        "watch": true,
        "out": "assets/style.css",
        "out_source_map": "assets/style.css.map"
    },
    "js": {
        "entrypoint": "main.js",
        "dir": "assets/js",
        "watch": true,
        "out": "assets/bundle.js",
        "source_map": true
    },
    "analytics": {
		"provider": "pirsch",
        "client_id": "id",
        "client_secret": "secret",
        "subnets": [
            "10.1.0.0/16",
            "10.2.0.0/8"
        ],
        "header": [
            "X-Forwarded-For",
            "Forwarded"
        ]
    }
}
`
)

func TestLoadConfig(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.json"))
	assert.NoError(t, os.WriteFile("config.json", []byte(sampleConfig), 0644))
	assert.NoError(t, Load(".", nil))
	assert.True(t, cfg.Dev)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 9, cfg.Server.ShutdownTimeout)
	assert.Equal(t, 10, cfg.Server.WriteTimeout)
	assert.Equal(t, 11, cfg.Server.ReadTimeout)
	assert.Equal(t, "cert/file.pem", cfg.Server.TLSCertFile)
	assert.Equal(t, "key/file.pem", cfg.Server.TLSKeyFile)
	assert.Equal(t, "example.com", cfg.Server.Hostname)
	assert.True(t, cfg.Server.SecureCookies)
	assert.Equal(t, "example.com", cfg.Server.CookieDomainName)
	assert.Equal(t, 600, cfg.Git.UpdateSeconds)
	assert.Equal(t, "https//github.com/foo/bar", cfg.Git.Repository)
	assert.Equal(t, "git", cfg.Content.Provider)
	assert.Equal(t, "/not-found", cfg.Content.NotFound["en"])
	assert.Equal(t, "*", cfg.CORS.Origins)
	assert.Equal(t, "info", cfg.CORS.Loglevel)
	assert.Equal(t, "style.scss", cfg.Sass.Entrypoint)
	assert.Equal(t, "assets/scss", cfg.Sass.Dir)
	assert.True(t, cfg.Sass.Watch)
	assert.Equal(t, "assets/style.css", cfg.Sass.Out)
	assert.Equal(t, "assets/style.css.map", cfg.Sass.OutSourceMap)
	assert.Equal(t, "main.js", cfg.JS.Entrypoint)
	assert.Equal(t, "assets/js", cfg.JS.Dir)
	assert.True(t, cfg.JS.Watch)
	assert.Equal(t, "assets/bundle.js", cfg.JS.Out)
	assert.True(t, cfg.JS.SourceMap)
	assert.Equal(t, "pirsch", cfg.Analytics.Provider)
	assert.Equal(t, "id", cfg.Analytics.ClientID)
	assert.Equal(t, "secret", cfg.Analytics.ClientSecret)
	assert.Len(t, cfg.Analytics.Subnets, 2)
	assert.Equal(t, "10.1.0.0/16", cfg.Analytics.Subnets[0])
	assert.Equal(t, "10.2.0.0/8", cfg.Analytics.Subnets[1])
	assert.Len(t, cfg.Analytics.Header, 2)
	assert.Equal(t, "X-Forwarded-For", cfg.Analytics.Header[0])
	assert.Equal(t, "Forwarded", cfg.Analytics.Header[1])
}

func TestLoadConfigNotExists(t *testing.T) {
	assert.NoError(t, os.RemoveAll("config.json"))
	err := Load(".", nil)
	assert.NotNil(t, err)
	assert.Equal(t, "error loading config.json: open config.json: no such file or directory", err.Error())
}
