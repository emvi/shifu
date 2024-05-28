# Shifu

Shifu is a simple Git-based content management system (CMS) and framework using the Go template syntax.

Content is managed in JSON files and can be automatically updated from Git, the file system, or a database.
Templates are updated automatically and JavaScript/TypeScript and Sass can be compiled on the fly, allowing for fast local development.
Shifu can also be used as a library in your Go application to add template functionality and custom behavior.

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
        "host": "localhost", # leave empty for production
        "port": 8080,
        "shutdown_time": 30, # time before the server is forcefully shut down (optional)
        "write_timeout": 5, # request write timeout
        "read_timeout": 5 # request read timeout
    },
    "sass": { # optional configuration to compile sass
        "dir": "assets", # asset directory path
        "entrypoint": "style.scss", # main sass file
        "out": "static/style.css", # compiled output css file path
        "out_source_map": "static/style.css.map", # css map file (optional)
        "watch": true # re-compile files when changed
    },
    "js": { # optional configuration to compile js/ts (see sass configuration for reference)
        "dir": "assets",
        "entrypoint": "entrypoint.js",
        "out": "static/bundle.js",
        "out_source_map": "static/bundle.js.map",
        "watch": true
    },
    "pirsch": { # optional configuration for Pirsch Analytics (pirsch.io)
        "client_id": "...", # optional when using an access key (recommended) instead of oAuth
        "client_secret": "..." # required
    }
}
```

## Structuring Your Website

*TODO*

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
