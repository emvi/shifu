package pkg

import (
	"context"
	"fmt"
	"github.com/emvi/shifu/pkg/analytics"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/js"
	"github.com/emvi/shifu/pkg/sass"
	"github.com/emvi/shifu/pkg/tpl"
	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzhttp"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

const (
	staticDir = "static"
)

// Start starts the Shifu server for given directory.
// The second argument is an optional template.FuncMap that will be merged with Shifu's funcmap.
func Start(dir string, funcMap template.FuncMap) error {
	slog.Info("Starting Shifu", "version", version, "directory", dir)
	ctx, cancel := context.WithCancel(context.Background())

	if err := cfg.Watch(ctx, dir, tpl.Merge(funcMap)); err != nil {
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

	// TODO
	/*if err := watchPartials(ctx, dir, tplFuncMap); err != nil {
		cancel()
		return err
	}

	if err := watchContent(ctx, dir, tplFuncMap); err != nil {
		cancel()
		return err
	}*/

	analytics.Init()
	router := setupRouter(dir)
	<-startServer(router, cancel)
	return nil
}

func setupRouter(dir string) *mux.Router {
	router := mux.NewRouter()
	// TODO
	//serveSitemap(router)
	serveStaticDir(router, dir)
	//servePage(router)
	return router
}

func serveStaticDir(router *mux.Router, dir string) {
	fs := http.StripPrefix("/static/", http.FileServer(http.Dir(filepath.Join(dir, staticDir))))
	router.PathPrefix("/static/").Handler(gzhttp.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
