package middleware

import (
	"github.com/emvi/shifu/pkg/admin/db"
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

		if err := db.Get().Get(s, `SELECT * FROM "session" WHERE session = ?`, session.Value); err != nil {
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
				if _, err := db.Get().Exec(`DELETE FROM "session" WHERE session = ?`, session.Value); err != nil {
					slog.Error("Error deleting session", "error", err)
				}
			}()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// IsAdmin returns whether the request is made by the administrator.
func IsAdmin(r *http.Request) bool {
	session, err := r.Cookie("session")

	if err != nil {
		return false
	}

	s := new(model.Session)

	if err := db.Get().Get(s, `SELECT * FROM "session" WHERE session = ?`, session.Value); err != nil {
		return false
	}

	if s == nil || s.Expires.Before(time.Now()) {
		return false
	}

	var email string

	if err := db.Get().Get(&email, `SELECT email FROM "user" WHERE id = ?`, s.UserID); err != nil {
		return false
	}

	return email == "admin"
}

// GetUser returns the signed in user for the request.
func GetUser(r *http.Request) *model.User {
	session, err := r.Cookie("session")

	if err != nil {
		return nil
	}

	s := new(model.Session)

	if err := db.Get().Get(s, `SELECT * FROM "session" WHERE session = ?`, session.Value); err != nil {
		return nil
	}

	if s == nil || s.Expires.Before(time.Now()) {
		return nil
	}

	user := new(model.User)

	if err := db.Get().Get(user, `SELECT * FROM "user" WHERE id = ?`, s.UserID); err != nil {
		return nil
	}

	return user
}
