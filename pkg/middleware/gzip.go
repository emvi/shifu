package middleware

import (
	"github.com/klauspost/compress/gzhttp"
	"net/http"
)

// Gzip returns a gzip middleware.
func Gzip() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return gzhttp.GzipHandler(next)
	}
}
