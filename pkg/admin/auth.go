package admin

import (
	"github.com/emvi/shifu/pkg/admin/model"
	"log/slog"
	"net/http"
	"time"
)

// Auth authenticates the admin session.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := r.Cookie("session")

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		s := new(model.Session)

		if err := db.Get(s, `SELECT * FROM "session" WHERE session = ?`, session.Value); err != nil {
			slog.Error("Error checking session", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if s == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if s.Expires.Before(time.Now()) {
			go func() {
				if _, err := db.Exec(`DELETE FROM "session" WHERE session = ?`, session.Value); err != nil {
					slog.Error("Error deleting session", "error", err)
				}
			}()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
