package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// serverGoVersion is the minimum Go version for custom servers (must match Shifu's go.mod).
	serverGoVersion = "1.27.0"

	// serverPublicDir is the directory containing the website (templates, content, static files, config).
	serverPublicDir = "public"

	serverGoMod = `module %s

go %s

require github.com/emvi/shifu v%s
`

	serverMainGo = `package main

import (
	"html/template"
	"log/slog"

	"%s/pkg/handler"
	shifu "github.com/emvi/shifu/pkg"
	"github.com/go-chi/chi/v5"
)

func main() {
	// Custom routes are merged with the Shifu router.
	router := chi.NewRouter()

	// The website (config.json, content, templates, static files) is located in the public directory.
	s, err := shifu.NewServer("public", shifu.ServerOptions{
		Router: router,

		// Custom template functions are merged with the Shifu template functions.
		FuncMap: template.FuncMap{},
	})

	if err != nil {
		slog.Error("Error setting up server", "error", err)
		return
	}

	handler.Register(s, router)

	if err := s.Start(nil); err != nil {
		slog.Error("Error starting server", "error", err)
	}
}
`

	serverHandlerGo = `package handler

import (
	"net/http"
	"strings"

	shifu "github.com/emvi/shifu/pkg"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/go-chi/chi/v5"
)

// Register adds the custom routes and page handlers to the server.
func Register(s *shifu.Server, router chi.Router) {
	// Custom route, independent of the CMS content.
	router.Get("/api/hello", Hello)

	// Page handler, used by pages with "handler": "example" in their content JSON.
	s.Content.SetHandler("example", Example)
}

// Hello is an example for a custom route.
func Hello(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("Hello from Shifu!"))
}

// Example is an example for a page handler. It renders the page as usual.
func Example(c *cms.CMS, page cms.Content, args map[string]string, w http.ResponseWriter, r *http.Request) {
	c.RenderPage(w, r, strings.ToLower(r.URL.Path), args, &page)
}
`

	serverMakefile = `.PHONY: run deps test build build_linux_amd64 build_linux_arm64 build_mac_amd64 build_mac_arm64 build_windows_amd64 build_windows_arm64

# cgo is required by the SQLite driver of the admin UI (cross-compiling needs a C cross-compiler)
BUILD = go build -a -ldflags "-s -w"

run:
	go run cmd/server/main.go

deps:
	go get -u -t ./...
	go mod tidy
	go mod vendor

test:
	go test -cover ./...

build:
	$(BUILD) -o server cmd/server/main.go

build_linux_amd64:
	GOOS=linux GOARCH=amd64 $(BUILD) -o server cmd/server/main.go

build_linux_arm64:
	GOOS=linux GOARCH=arm64 $(BUILD) -o server cmd/server/main.go

build_mac_amd64:
	GOOS=darwin GOARCH=amd64 $(BUILD) -o server cmd/server/main.go

build_mac_arm64:
	GOOS=darwin GOARCH=arm64 $(BUILD) -o server cmd/server/main.go

build_windows_amd64:
	GOOS=windows GOARCH=amd64 $(BUILD) -o server.exe cmd/server/main.go

build_windows_arm64:
	GOOS=windows GOARCH=arm64 $(BUILD) -o server.exe cmd/server/main.go
`

	serverGitIgnore = `# binaries
/server
/server.exe

# IDEs
.vscode/
.idea/

# secrets and runtime data
.secrets.env
shifu.db
`
)

// InitServer creates a new Go project for a custom server using Shifu as a library in given directory.
// The website itself is initialized in the "public" subdirectory. If module is empty, the directory name is used.
func InitServer(path, module string) error {
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return errors.New("the target directory already contains a Go project (go.mod)")
	}

	if err := Init(filepath.Join(path, serverPublicDir)); err != nil {
		return err
	}

	if module == "" {
		abs, err := filepath.Abs(path)

		if err != nil {
			return err
		}

		module = strings.ToLower(filepath.Base(abs))
	}

	if module == "" || strings.ContainsAny(module, " \t\n") {
		return errors.New("invalid module name")
	}

	for _, dir := range []string{"cmd/server", "pkg/handler"} {
		if err := os.MkdirAll(filepath.Join(path, dir), 0755); err != nil {
			return err
		}
	}

	files := map[string]string{
		"go.mod":                 fmt.Sprintf(serverGoMod, module, serverGoVersion, version),
		"cmd/server/main.go":     fmt.Sprintf(serverMainGo, module),
		"pkg/handler/handler.go": serverHandlerGo,
		"Makefile":               serverMakefile,
		".gitignore":             serverGitIgnore,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(path, name), []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}
