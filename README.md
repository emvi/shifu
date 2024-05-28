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
* Automatically watch files for changes
* Build and minify JavaScript/TypeScript and Sass 
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

Or through Docker:

```yaml
version: "3"

volumes:
  shifu-data:

services:
  shifu:
    image: emvicom/shifu
    container_name: shifu
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - shifu-data:/app/data
```

## Configuration

Shifu is configured using a single `config.json` file inside the project directory.

```json
{
    "server": {
        "host": "localhost", // leave empty for production
        "port": 8080,
        "shutdown_time": 30, // time before the server is forcefully shut down (optional)
        "write_timeout": 5, // request write timeout
        "read_timeout": 5 // request read timeout
    },
    "sass": { // optional configuration to compile sass
        "dir": "assets", // asset directory path
        "entrypoint": "style.scss", // main sass file
        "out": "static/style.css", // compiled output css file path
        "out_source_map": "static/style.css.map", // css map file (optional)
        "watch": true // re-compile files when changed
    },
    "js": { // optional configuration to compile js/ts (see sass configuration for reference)
        "dir": "assets",
        "entrypoint": "entrypoint.js",
        "out": "static/bundle.js",
        "out_source_map": "static/bundle.js.map",
        "watch": true
    },
    "pirsch": { // optional configuration for Pirsch Analytics (pirsch.io)
        "client_id": "...", // optional when using an access key (recommended) instead of oAuth
        "client_secret": "..." // required
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

The JSON structure for a content file is as follows:

```json
{
    "path": {
        "en": "/", // /404 is a special case serving the 404 not found page
        "de": "/de"
    },
    "sitemap": {
        "priority": "1.0" // default is 1.0
    },
    "header": { // optional list of headers
        "X-Frame-Options": "deny"
    },
    "handler": "custom_handler", // sets a custom handler defined on the backend
    "analytics": { // optional analytics meta data
        "tags": {
            "key": "value"
        },
        "experiment": {
            "name": "landing", // A/B testing page variant
            "variant": "a"
        }
    },
    "content": {
        "content": [
            {
                "ref": "head", // references to a standalone element (JSON file without extension, always lowercase)
                "data": {
                    // overwrites whatever is set in head.json
                },
                "copy": {
                    "en": {
                        "title": "Home" // overrides the copy "title" with the value "Home"
                    }
                }
            },
            {
                "tpl": "text", // template file (without extension, always lowercase)
                "analytics": {
                    "experiment": {
                        "name": "experiment", // A/B testing experiment name and variant
                        "variant": "a"
                    }
                },
                "data": { // optional generic data object
                    "numbers": [1, 2, 3]
                },
                "copy": { // optional data used in the template
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
                    "children": [ /* ... */ ]
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
        // ...
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
cms.SetHandler("blog", func Blog(c *cms.CMS, page cms.Content, w http.ResponseWriter, r *http.Request) {
	// ...
    c.RenderPage(w, r, strings.ToLower(r.URL.Path), &page)
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

```go
package main

import (
	shifu "github.com/emvi/shifu/pkg"
	"log/slog"
	"html/template"
)

func loadAndRenderBlogArticle() string {
    return "TODO"
}

func main() {
	// Define a custom FuncMap to load and render blog articles from an external source.
	customFuncMap := template.FuncMap{
		"blogArticle": loadAndRenderBlogArticle,
	}
		
    // Start Shifu from the content/dir directory and pass your own template.FuncMap.
    // The FuncMap will be merged with the default FuncMap of Shifu.
	if err := shifu.Start("content/dir", customFuncMap); err != nil {
		slog.Error("Error starting Shifu", "error", err)
	}
}
```

## License

MIT
