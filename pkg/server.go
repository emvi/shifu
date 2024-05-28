package pkg

import (
	"context"
	"fmt"
	"github.com/emvi/shifu/pkg/analytics"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"github.com/emvi/shifu/pkg/js"
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
	"time"
)

// Start starts the Shifu server for given directory.
// The second argument is an optional template.FuncMap that will be merged with Shifu's funcmap.
func Start(dir string, funcMap template.FuncMap) error {
	slog.Info("Starting Shifu", "version", version, "directory", dir)
	funcMap = cms.Merge(funcMap)
	ctx, cancel := context.WithCancel(context.Background())

	if err := cfg.Watch(ctx, dir, funcMap); err != nil {
		cancel()
		return err
	}

	if err := sass.Watch(ctx, dir); err != nil {
		cancel()
		return err
	}

	if err := js.Watch(ctx, dir); err != nil {
		cancel()
		return err
	}

	provider := source.NewFS(dir, 0) // TODO provider from config
	sm := sitemap.New()
	content := cms.NewCMS(cms.Options{
		Ctx:       ctx,
		BaseDir:   dir,
		HotReload: cfg.Get().Dev,
		FuncMap:   funcMap,
		Source:    provider,
		Sitemap:   sm,
	})
	analytics.Init()
	router := setupRouter(dir, content, sm)
	<-startServer(router, cancel)
	return nil
}

func setupRouter(dir string, cms *cms.CMS, sm *sitemap.Sitemap) chi.Router {
	router := chi.NewRouter()
	/*router.Use( TODO middlewares
		server.Cors(),
		server.Gzip(),
	)
	serveRobotsTxt(router)*/
	sm.Serve(router)
	serveStaticDir(router, dir)
	router.Handle("/*", http.HandlerFunc(cms.Serve))
	return router
}

func serveStaticDir(router chi.Router, dir string) {
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(dir, "static"))))
	router.Handle("/static/*", gzhttp.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})))
}

func startServer(handler http.Handler, cancel context.CancelFunc) chan struct{} {
	slog.Info("Starting server...")
	config := cfg.Get()
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	server := &http.Server{
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

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Error shutting down server gracefully", "error", err)
			panic(err)
		}

		cancel()
	}()

	done := make(chan struct{})

	go func() {
		if config.Server.TLSCertFile != "" && config.Server.TLSKeyFile != "" {
			if err := server.ListenAndServeTLS(config.Server.TLSCertFile, config.Server.TLSKeyFile); err != nil {
				slog.Error("Error starting server", "error", err)
				panic(err)
			}
		} else {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("Error starting server", "error", err)
				panic(err)
			}
		}

		done <- struct{}{}
	}()

	slog.Info("Server started!", "address", addr)
	return done
}
