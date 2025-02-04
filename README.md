# Shifu

Shifu is a simple Git-based content management system (CMS) and framework using the Go template syntax.

Content is managed in JSON files and can be automatically updated from Git, the file system, or a database.
Templates are updated automatically and JavaScript/TypeScript and Sass can be compiled on the fly, allowing for fast local development.
Shifu can also be used as a library in your Go application to add template functionality and custom behavior.

## Features

* JSON-based content management
* Reference commonly used elements
* Support for i18n (translations)
* Serve static files (from file system or S3)
* Reusable Golang template files
* Automatically watch files for changes
* Build and minify JavaScript/TypeScript and Sass
* Static content cache
* 404 error fallback page
* Automatic sitemap generation
* Integrated analytics using [Pirsch](https://pirsch.io)
* A/B testing
* Custom Golang handlers for advanced functionality
* Simple configuration and easy deployment
* Standalone server or library

## Installation and Setup

Download the latest release for your platform from the releases section on GitHub.
Move the binary to a directory in your $PATH (like `/usr/local/bin`).
For Sass, you need to install the `sass` command globally (`sudo npm i -g sass`).
After that you can run Shifu from the command line with the `shifu` command.

* `shifu run <path>` will run Shifu in the given directory.
* `shifu init <path>` will initialize a new project in the given directory.
* `shifu version` will print the version number of Shifu.

systemd sample configuration:

```ini
[Unit]
Description=Shifu Website

[Service]
Type=simple
ExecStart=/root/example.com/shifu
WorkingDirectory=/root/example.com

[Install]
WantedBy=multi-user.target
```

## Configuration

Shifu is configured using a single `config.json` file inside the project's root directory. A `.secrets.env` file can be added to store secrets. Secrets are key-value pairs, one per line.

```env
SECRET=value
```

You can then replace them inside `config.json` using `${SECRET}`.

Below is the entire configuration. Keys starting with `_` are comments.

```json
{
    "dev": false,
    "_log_level": "debug, info, warn, error",
    "log_level": "info",
    "server": {
        "_host": "Leave empty for production",
        "host": "localhost",
        "port": 8080,
        "_shutdown_time": "Time before the server is forcefully shut down (optional).",
        "shutdown_time": 30,
        "_write_timeout": "Request write timeout.",
        "write_timeout": 5,
        "_read_timeout": "Request read timeout.",
        "read_timeout": 5,
        "tls_cert_file": "cert/file.pem",
        "tls_key_file": "key/file.pem",
        "hostname": "example.com",
        "secure_cookies": true,
        "cookie_domain_name": "example.com"
    },
    "content": {
        "_provider": "git, fs",
        "provider": "git",
        "update_seconds": 600,
        "repository": "https://...",
        "_not_found": "Overrides the default 404-page path (/404).",
        "not_found": {
            "en": "/not-found",
            "de": "/de/nicht-gefunden"
        }
    },
    "_static": "Provides an optional storage configuration. It uses the file system by default",
    "static": {
        "_provider": "fs, s3",
        "provider": "s3",
        "url": "fsn1.your-objectstorage.com",
        "bucket": "hetzner-bucket",
        "access_key": "...",
        "secret": "..."
    },
    "cors": {
        "origins": "*",
        "loglevel": "info"
    },
    "_sass": "Optional configuration to compile Sass.",
    "sass": {
        "_dir": "Asset directory path.",
        "dir": "assets",
        "_entrypoint": "Main sass file.",
        "entrypoint": "style.scss",
        "_out": "Compiled output CSS file path.",
        "out": "static/style.css",
        "_out_source_map": "CSS map file (optional).",
        "out_source_map": "static/style.css.map",
        "_watch": "Re-compile files when changed.",
        "watch": true
    },
    "_js": "Optional configuration to compile js/ts (see Sass configuration for reference).",
    "js": {
        "dir": "assets",
        "entrypoint": "entrypoint.js",
        "out": "static/bundle.js",
        "source_map": true,
        "watch": true
    },
    "_analytics": "Optional analytics configuration.",
    "analytics": {
        "provider": "pirsch",
        "_client_id": "Optional when using an access key (recommended) instead of oAuth.",
        "client_id": "...",
        "_client_secret": "Required.",
        "client_secret": "...",
        "_subnets": "Optional subnet configuration.",
        "subnets": [
            "10.1.0.0/16",
            "10.2.0.0/8"
        ],
        "_header": "Optional IP header configuration.",
        "header": [
            "X-Forwarded-For",
            "Forwarded"
        ]
    }
}
```

## Structuring Your Website

The directory structure is as follows:

| Directory | Description                                          |
|-----------|------------------------------------------------------|
| content/  | Recursive content files in JSON format.              |
| static/   | Static content (will be served as is on `/static/`). |
| tpl/      | Recursive Golang template files.                     |
| tmp/      | For temporary files.                                 |

The JSON structure for a content file is as follows:

```json
{
    "_disable_cache": "Disables the static content cache (does not affect custom handlers).",
    "disable_cache": false,
    "_path": "A list of routes on which this page should be served. Routes support regular expressions and variables, such as /p/{var} or /p/{var:[0-9]+}. They are available as a string map called 'args'.",
    "path": {
        "_en": "Use /404 (default) as a special case serving the 404 not found page.",
        "en": "/",
        "de": "/de"
    },
    "sitemap": {
        "_priority": "Default is 1.0.",
        "priority": "1.0"
    },
    "_header": "Optional list of headers.",
    "header": {
        "X-Frame-Options": "deny"
    },
    "_handler": "Sets a custom handler defined on the backend.",
    "handler": "custom_handler",
    "_analytics": "Optional analytics meta data.",
    "analytics": {
        "tags": {
            "key": "value"
        },
        "experiment": {
            "_name": "A/B testing page variant.",
            "name": "landing",
            "variant": "a"
        }
    },
    "content": {
        "content": [
            {
                "_ref": "References to a standalone element (JSON file without extension, always lowercase).",
                "ref": "head",
                "data": {
                    "_": "Overwrites whatever is set in head.json."
                },
                "copy": {
                    "en": {
                        "_title": "Overrides the copy 'title' with the value 'Home'.",
                        "title": "Home"
                    }
                }
            },
            {
                "_tpl": "Template file (without extension, always lowercase).",
                "tpl": "text",
                "analytics": {
                    "experiment": {
                        "_name": "A/B testing experiment name and variant.",
                        "name": "experiment",
                        "variant": "a"
                    }
                },
                "_data": "Optional generic data object.",
                "data": {
                    "numbers": [1, 2, 3]
                },
                "_copy": "Optional data used in the template.",
                "copy": {
                    "en": {
                        "headline": "Welcome!",
                        "text": "To Shifu."
                    },
                    "de": {
                        "headline": "Willkommen!",
                        "text": "Bei Shifu."
                    }
                },
                "content": {
                    "children": [
                        {
                            "_": "..."
                        }
                    ]
                }
            }
        ]
    }
}
```

Standalone elements are use the same structure as pages, but do not specify paths. Here is an example:

```json
{
    "tpl": "head",
    "data": {
        "_": "..."
    },
    "copy": {
        "en": {
            "title": "Welcome to Shifu"
        },
        "de": {
            "title": "Willkommen bei Shifu"
        }
    }
}
```

Custom handlers are implemented like this:

```go
cms := shifu.NewCMS(cms.Options{
	// ...
})
cms.SetHandler("blog", func(c *cms.CMS, page cms.Content, args map[string]string, w http.ResponseWriter, r *http.Request) {
	// ...
	c.RenderPage(w, r, strings.ToLower(r.URL.Path), args, &page)
})
```

## Template Functions

Shifu comes with a number of template functions that can be used within templates.

| Function      | Description                                                                                                        | Example                                                    |
|---------------|--------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------|
| config        | Exposes the Oogway configuration.                                                                                  | `{{config.Server.Host}}`                                   |
| hostname      | Returns the hostname from configuration.                                                                           | `{{hostname}}`                                             |
| copy          | Returns the copy (text) for given page, content, and key.                                                          | `{{copy .Page .Content "meta_description"}}`               |
| get           | Returns the variable for given content and key.                                                                    | `{{get .Content "img"}}`                                   |
| markdown      | Renders given markdown file as HTML using Go text templates. Use the full path for the template name.              | `{{markdown "static/blog/article.md" .}}`                  |
| markdownBlock | Renders a block from given markdown file as HTML using Go text templates. Use the full path for the template name. | `{{markdownBlock "static/blog/article.md" "blockName" .}}` |
| int           | Converts given string to an integer.                                                                               | `{{int "123"}}`                                            |
| uint64        | Converts given int to an uint64.                                                                                   | `{{uint64 123}}`                                           |
| shuffle       | Shuffles given list and returns up to n results if n > 0.                                                          | `{{shuffle .List 10}}`                                     |

For more template functions, see the [Sprig documentation](github.com/Masterminds/sprig).

## Using Shifu as a Library

Shifu is designed to be used as a standalone server, but also as a library for Go (Golang).
You can add your own template functions for more advanced functionality and use cases and embed them in your application.
Just `go get` it and call it anywhere in your application to start a web server.

### Basic Example

```go
package main

import (
	"log/slog"

	shifu "github.com/emvi/shifu/pkg"
)

func main() {
	// Set up Shifu from the content/dir directory.
	s, err := shifu.NewServer("content/dir", shifu.ServerOptions{})

	if err != nil {
		panic(err)
	}

	// Start the server. The cancel function is optional.
	if err := s.Start(nil); err != nil {
		slog.Error("Error starting server", "error", err)
	}
}
```

### Advanced Example

```go
package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	shifu "github.com/emvi/shifu/pkg"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/go-chi/chi/v5"
)

func loadAndRenderBlogArticle() string {
	return "TODO"
}

func main() {
	// Optional custom router. The routes will be merged with the Shifu router.
	router := chi.NewRouter()
	router.Get("/example", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	// Set up Shifu from the content/dir directory.
	s, err := shifu.NewServer("content/dir", shifu.ServerOptions{
		Router: router,

		// Define a custom FuncMap to load and render blog articles from an external source.
		FuncMap: template.FuncMap{
			"blogArticle": loadAndRenderBlogArticle,
		},
	})

	if err != nil {
		panic(err)
	}

	// Add a custom handler.
	s.Content.SetHandler("blog", func(c *cms.CMS, page cms.Content, args map[string]string, w http.ResponseWriter, r *http.Request) {
		// ...
		c.RenderPage(w, r, strings.ToLower(r.URL.Path), args, &page)
	})

	// Start the server. The cancel function is optional.
	if err := s.Start(nil); err != nil {
		slog.Error("Error starting server", "error", err)
	}
}
```

## License

MIT
