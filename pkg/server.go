package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/analytics"
	"github.com/emvi/shifu/pkg/api"
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
	funcMap template.FuncMap
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewServer creates a new Shifu server for given directory.
// The second argument is an optional template.FuncMap that will be merged with Shifu's funcmap.
func NewServer(dir string, options ServerOptions) (*Server, error) {
	options.FuncMap = cms.Merge(options.FuncMap)

	if err := cfg.Load(dir, options.FuncMap); err != nil {
		return nil, err
	}

	if cfg.Get().Dev {
		slog.Warn("Dev mode is enabled!")
	}

	switch cfg.Get().LogLevel {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
		break
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
		break
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
		break
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	var provider source.Provider

	switch strings.ToLower(strings.TrimSpace(cfg.Get().Content.Provider)) {
	case "git":
		gitConfig := cfg.Get().Git

		if gitConfig.Repository == "" {
			return nil, errors.New("git repository URL is empty")
		}

		provider = source.NewGit(dir, gitConfig.Repository, gitConfig.UpdateSeconds)
		break
	default:
		provider = source.NewFileSystem(dir, cfg.Get().Git.UpdateSeconds)
	}

	sm := sitemap.New()
	ctx, cancel := context.WithCancel(context.Background())
	content := cms.NewCMS(cms.Options{
		Ctx:       ctx,
		BaseDir:   dir,
		HotReload: cfg.Get().Dev,
		NotFound:  cfg.Get().Content.NotFound,
		FuncMap:   options.FuncMap,
		Source:    provider,
		Sitemap:   sm,
	})
	return &Server{
		Content: content,
		Sitemap: sm,
		router:  options.Router,
		funcMap: options.FuncMap,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Start starts the Shifu server.
// The context.CancelFunc is optional and will be called on server shutdown or error if set.
func (server *Server) Start(cancel context.CancelFunc) error {
	slog.Info("Starting Shifu", "version", version, "directory", cfg.Get().BaseDir)
	stop := func() {
		server.cancel()

		if cancel != nil {
			cancel()
		}
	}

	if err := sass.Watch(server.ctx); err != nil {
		stop()
		return err
	}

	if err := js.Watch(server.ctx); err != nil {
		stop()
		return err
	}

	analytics.Init()
	server.setupRouter(cfg.Get().BaseDir)
	<-server.startServer(server.router, stop)
	return nil
}

func (server *Server) setupRouter(dir string) {
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

	if cfg.Get().API.Secret != "" {
		server.serveAPI(router)
	}

	server.Sitemap.Serve(router)
	server.serveRobotsTxt(router)
	server.serveStaticDir(router, dir)
	router.HandleFunc("/*", server.Content.Serve)
	server.router = router

	for _, route := range router.Routes() {
		slog.Info("Added route", "route", route.Pattern)
	}
}

func (server *Server) serveAPI(router chi.Router) {
	router.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.APISecret)
		r.Get("/health", func(http.ResponseWriter, *http.Request) {})
		r.Get("/static", api.ListStaticFiles)
		r.Put("/static", api.PutStaticFile)
		r.Get("/cms", func(http.ResponseWriter, *http.Request) { server.Content.Update() })
		r.Get("/content", api.ListContentFiles)
		r.Get("/content/file", api.GetContentFile)
		r.Put("/content/file", api.PutContentFile)
	})
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
