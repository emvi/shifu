package middleware

import (
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

// Gzip returns a gzip middleware.
func Gzip() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return gzhttp.GzipHandler(next)
	}
}
