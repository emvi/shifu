package admin

import (
	"log/slog"
	"net/http"
)

// Auth authenticates the admin session.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		found := false

		if err := db.Get(&found, `SELECT EXISTS (SELECT 1 FROM "session" WHERE session = ?)`, session.Value); err != nil {
			slog.Error("Error checking session", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
