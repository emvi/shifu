# Shifu

Shifu is a simple Git-based content management system (CMS) and framework using the Go template syntax.

Content is managed in JSON files and can be automatically updated from Git, the file system, or a database.
Templates are updated automatically and JavaScript/TypeScript and Sass can be compiled on the fly, allowing for fast local development.
Shifu can also be used as a library in your Go application to add template functionality and custom behavior.

## Features

* JSON-based content management
* Reference commonly used elements
* Support for i18n (translations)
* Serve static files
* Reusable Golang template files
* Build and minify JavaScript/TypeScript and Sass on save
* Static content cache
* Configurable 404 error fallback page
* Automatic sitemap generation
* Integrated analytics using [Pirsch](https://pirsch.io)
  * Track page views
  * Track custom events
  * Integrated A/B testing
* Custom Golang handlers for advanced functionality
* Simple configuration and easy deployment
* Standalone server or library
* Push/Pull files to/from a remote server
* Admin UI

## Installation and Setup

Download the latest release for your platform from the releases section on GitHub.
Move the binary to a directory in your $PATH (like `/usr/local/bin`).
For Sass, you need to install the `sass` command globally (`sudo npm i -g sass`).
After that you can run Shifu from the command line with the `shifu` command.

* `shifu run <path>` will run Shifu in the given directory
* `shifu init <path>` will initialize a new project in the given directory
* `shifu version` will print the version number of Shifu
* `shifu pull` will pull changed `static` and `content` files from a remote server if configured
* `shifu push` will push changed `static` and `content` files to a remote server if configured

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
    "_api": "Enables the API",
    "api": {
        "secret": "secret"
    },
    "_remote": "Configures a remote server for synchronization (push/pull) using the API secret",
    "remote": {
        "url": "https://example.com",
        "secret": "secret"
    },
    "_git": "Pulls files from a Git repository regularly",
    "git": {
        "update_seconds": 600,
        "repository": "https://..."
    },
    "content": {
        "_provider": "git, fs",
        "provider": "git",
        "_not_found": "Overrides the default 404-page path (/404).",
        "not_found": {
            "en": "/not-found",
            "de": "/de/nicht-gefunden"
        }
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
    },
    "ui": {
        "_path": "The path where you can sign in to the admin UI.",
        "path": "/admin",
        "_admin_password": "The default admin user password.",
        "admin_password": "${ADMIN_PASSWORD}"
    }
}
```

## Structuring Your Website

The directory structure is as follows:

| Directory  | Description                                          |
|------------|------------------------------------------------------|
| admin/tpl/ | The template configuration for the admin UI.         |
| content/   | Recursive content files in JSON format.              |
| static/    | Static content (will be served as is on `/static/`). |
| tpl/       | Recursive Golang template files.                     |

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

## Build-in Variables

When rendering a template, the following variables are always available. Click one of the links below to see the struct definitions.

```
CMS     *CMS
Args    map[string]string
Page    *Content
Content *Content
```

* [CMS](pkg/cms/cms.go) is the CMS main struct that can be used to render elements and access general configuration
* `Args` are the router parameters when using dynamic route matching
* [Page](pkg/cms/model.go) is the page object, including the routes and other page attributes
* [Content](pkg/cms/model.go) is the current element that is being rendered

## Template Functions

Shifu comes with a number of template functions that can be used within templates.

| Function      | Description                                                                                                        | Example                                                      |
|---------------|--------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| config        | Exposes the Oogway configuration.                                                                                  | `{{ config.Server.Host }}`                                   |
| hostname      | Returns the hostname from configuration.                                                                           | `{{ hostname }}`                                             |
| copy          | Returns the copy (text) for given page, content, and key.                                                          | `{{ copy .Page .Content "meta_description" }}`               |
| get           | Returns the variable for given content and key.                                                                    | `{{ get .Content "img" }}`                                   |
| markdown      | Renders given markdown file as HTML using Go text templates. Use the full path for the template name.              | `{{ markdown "static/blog/article.md" . }}`                  |
| markdownBlock | Renders a block from given markdown file as HTML using Go text templates. Use the full path for the template name. | `{{ markdownBlock "static/blog/article.md" "blockName" . }}` |
| int           | Converts given string to an integer.                                                                               | `{{ int "123" }}`                                            |
| uint64        | Converts given int to an uint64.                                                                                   | `{{ uint64 123 }}`                                           |
| shuffle       | Shuffles given list and returns up to n results if n > 0.                                                          | `{{ shuffle .List 10 }}`                                     |
| fmt           | Formats a string.                                                                                                  | `{{ fmt "foo %s" "bar" }}`                                   |
| dict          | Creates a map from given key value pairs.                                                                          | `{{ dict "key" "value" "answer" 42 }}`                       |
| default       | Returns the first value that's not nil.                                                                            | `{{ default .Val0 .Val1 }}`                                  |
| year          | Returns the current year                                                                                           | `{{ year }}`                                                 |
| formatFloat   | Formats the given value with two decimal places.                                                                   | `{{ formatFloat 42.34567 }}`                                 |
| formatInt     | Formats the given value with separators (comma).                                                                   | `{{ formatInt 4213465576 }}`                                 |
| formatDate    | Formats a data for given layout.                                                                                   | `{{ formatDate .Date "2006-01-02" }}`                        |
| gtFloat       | Checks whether the first value is bigger than the second.                                                          | `{{ gtFloat 6 3 }}`                                          |
| ltFloat       | Checks whether the first value is smaller than the second.                                                         | `{{ gtFloat 6 3 }}`                                          |
| html          | Returns given string as valid HTML. This is unsafe if the input is user provided.                                  | `{{ html "<p>Text</p>" }}`                                   |
| htmlAttr      | Returns given string as a valid HTML attribute. This is unsafe if the input is user provided.                      | `{{ htmlAttr "value" }}`                                     |
| loggedIn      | Returns true if the visitor is signed in as admin.                                                                 | `{{ loggedIn .Page }}`                                       |
| adminHead     | Returns the HTML for the `<head>` section for the admin UI.                                                        | `{{ adminHead .Page }}`                                      |
| adminBody     | Returns the HTML for the `<body>` section for the admin UI.                                                        | `{{ adminBody .Page }}`                                      |

For more template functions, see the [Sprig documentation](https://github.com/Masterminds/sprig).

## Admin UI

Shifu has a build-in admin UI. To enable it, add the `ui` section to the `config.json`. Static files are provided under the configured path and always under `/shifu-admin`.

Elements and references on the page can only be edited if you add a template configuration file for them. They need to be created inside `admin/tpl/` and look like this:

```json
{
    "_label": "The display name for the template file.",
    "label": "Label",
    "_layout": "Layout marks the element as structural, like for the head section, disabling certain actions.",
    "layout": false,
    "_content": "The configuration for the content areas.",
    "content": {
        "content": {
            "label": "Inhalt",
            "tpl_filter": ["filter", "templates"]
        }
    },
    "_copy": "The display name and field types for the copy section. The type can be html, img, file, boolean, or select.",
    "copy": {
        "headline": {
            "label": "Headline"
        },
        "text": {
            "label": "Main Text",
            "type": "html"
        }
    },
    "_data": "The same as for copy, but for the data fields instead.",
    "data": {
        "headline_size": {
            "label": "Headline Size",
            "type": "select",
            "_options": "A list of options for a select field.",
            "options": {
                "h1": "H1",
                "h2": "H2",
                "h3": "H3"
            }
        }
    }
}
```

In order for the on-page editor to work, the HTML for the elements need to be structured like this:

```html
<section data-shifu-element="{{ .Content.Position }}">
    <p>An element...</p>

    {{ .CMS.Render .Args .Page (fmt "%s.content" .Content.Position) (index .Content.Content "content") }}
    <slot name="{{ .Content.Position }}/content"></slot>
</section>
```

Where `data-shifu-element="{{ .Content.Position }}"` must be added to the top-level element, so that it can be modified. `<slot name="{{ .Content.Position }}/content"></slot>` can be used to allow nesting, where the part after the slash is the name of the position in the template.

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
			"blogArticle": func() string {
				// ...
				return "TODO"
            },
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
