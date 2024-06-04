package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/analytics"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/emvi/shifu/pkg/js"
	"github.com/emvi/shifu/pkg/middleware"
	"github.com/emvi/shifu/pkg/sass"
	"github.com/emvi/shifu/pkg/sitemap"
	"github.com/emvi/shifu/pkg/source"
	"github.com/go-chi/chi/v5"
	"github.com/klauspost/compress/gzhttp"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

// ServerOptions are the options for the Shifu Server.
type ServerOptions struct {
	// Router is the router used by the server. If set, it will be used to attach the Shifu handlers.
	// Otherwise, a new router will be created.
	Router chi.Router

	// FuncMap will be merged with the default Shifu template function map.
	FuncMap template.FuncMap
}

// Server is the Shifu server.
type Server struct {
	// Content is the CMS content.
	Content *cms.CMS

	// Sitemap is the sitemap generated from the content.
	Sitemap *sitemap.Sitemap

	router  chi.Router
	dir     string
	funcMap template.FuncMap
}

// NewServer creates a new Shifu server for given directory.
// The second argument is an optional template.FuncMap that will be merged with Shifu's funcmap.
func NewServer(dir string, options ServerOptions) *Server {
	return &Server{
		Sitemap: sitemap.New(),
		router:  options.Router,
		dir:     dir,
		funcMap: options.FuncMap,
	}
}

// Start starts the Shifu server.
// The context.CancelFunc is optional and will be called on server shutdown or error if set.
func (server *Server) Start(cancel context.CancelFunc) error {
	slog.Info("Starting Shifu", "version", version, "directory", server.dir)
	server.funcMap = cms.Merge(server.funcMap)
	ctx, cancelServer := context.WithCancel(context.Background())
	stop := func() {
		cancelServer()

		if cancel != nil {
			cancel()
		}
	}

	if err := cfg.Watch(ctx, server.dir, server.funcMap); err != nil {
		stop()
		return err
	}

	if err := sass.Watch(ctx, server.dir); err != nil {
		stop()
		return err
	}

	if err := js.Watch(ctx, server.dir); err != nil {
		stop()
		return err
	}

	config := cfg.Get().Content
	var provider source.Provider

	switch strings.ToLower(config.Provider) {
	case "fs":
		provider = source.NewFS(server.dir, config.UpdateSeconds)
		break
	case "git":
		provider = source.NewGit(server.dir, config.Repository, config.UpdateSeconds)
		break
	default:
		stop()
		return errors.New("content provider not found")
	}

	server.Content = cms.NewCMS(cms.Options{
		Ctx:       ctx,
		BaseDir:   server.dir,
		HotReload: cfg.Get().Dev,
		FuncMap:   server.funcMap,
		Source:    provider,
		Sitemap:   server.Sitemap,
	})
	analytics.Init()
	server.setupRouter()
	<-server.startServer(server.router, stop)
	return nil
}

func (server *Server) setupRouter() {
	router := chi.NewRouter()
	router.Use(
		middleware.Cors(),
		middleware.Gzip(),
	)

	if server.router != nil {
		slog.Info("Merging router with Shifu router...")

		for _, route := range server.router.Routes() {
			for method, handler := range route.Handlers {
				if method == "*" {
					router.Handle(route.Pattern, handler)
				} else {
					router.Method(method, route.Pattern, handler)
				}
			}
		}
	}

	server.Sitemap.Serve(router)
	server.serveRobotsTxt(router)
	server.serveStaticDir(router, server.dir)
	router.Handle("/*", http.HandlerFunc(server.Content.Serve))
	server.router = router

	for _, route := range router.Routes() {
		slog.Info("Added route", "route", route.Pattern)
	}
}

func (server *Server) serveRobotsTxt(router chi.Router) {
	robotsTxt := fmt.Sprintf("User-agent: *\nDisallow:\n\nSitemap: %s/sitemap.xml\n", cfg.Get().Server.Hostname)
	router.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		w.Header().Add("Cache-Control", "max-age=86400")

		if _, err := w.Write([]byte(robotsTxt)); err != nil {
			slog.Debug("Error serving robots.txt", "error", err)
		}
	}))
}

func (server *Server) serveStaticDir(router chi.Router, dir string) {
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(dir, "static"))))
	router.Handle("/static/*", gzhttp.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})))
}

func (server *Server) startServer(handler http.Handler, cancel context.CancelFunc) chan struct{} {
	slog.Info("Starting server...")
	config := cfg.Get()
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	httpServer := &http.Server{
		Handler:      handler,
		Addr:         addr,
		WriteTimeout: time.Second * time.Duration(config.Server.WriteTimeout),
		ReadTimeout:  time.Second * time.Duration(config.Server.ReadTimeout),
	}

	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		slog.Info("Shutting down server...")
		cancel()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(config.Server.ShutdownTimeout))

		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("Error shutting down server gracefully", "error", err)
			panic(err)
		}

		cancel()
	}()

	done := make(chan struct{})

	go func() {
		if config.Server.TLSCertFile != "" && config.Server.TLSKeyFile != "" {
			if err := httpServer.ListenAndServeTLS(config.Server.TLSCertFile, config.Server.TLSKeyFile); err != nil {
				slog.Error("Error starting server", "error", err)
				panic(err)
			}
		} else {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Error starting server", "error", err)
				panic(err)
			}
		}

		done <- struct{}{}
	}()

	slog.Info("Server started!", "address", addr)
	return done
}
