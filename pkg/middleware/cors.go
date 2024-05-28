package middleware

import (
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/go-chi/cors"
	"net/http"
	"strings"
)

// Cors configures CORS log level and origins.
// The configuration is not restricted by default.
func Cors() func(next http.Handler) http.Handler {
	origins := strings.Split(cfg.Get().CORS.Origins, ",")
	handler := cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400, // one day
		Debug:            strings.ToLower(cfg.Get().CORS.Loglevel) == "debug",
	})
	return handler
}
