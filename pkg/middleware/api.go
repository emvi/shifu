package middleware

import (
	"net/http"
	"strings"

	"github.com/emvi/shifu/pkg/cfg"
)

// APISecret checks the Authorization header against the configured API secret.
// The schema must be "Key".
func APISecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		schema, key, found := strings.Cut(auth, " ")

		if !found || schema != "Key" || key != cfg.Get().API.Secret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
