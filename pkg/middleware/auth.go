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
		if GetSession(r) == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetSession returns the session if the request was made by a signed-in user.
// Otherwise, nil is returned.
func GetSession(r *http.Request) *model.Session {
	session, err := r.Cookie("session")

	if err != nil {
		return nil
	}

	s := new(model.Session)

	if err := db.Get().Get(s, `SELECT * FROM "session" WHERE session = ?`, session.Value); err != nil {
		slog.Error("Error checking session", "error", err)
		return nil
	}

	if s == nil {
		return nil
	}

	if s.Expires.Before(time.Now()) {
		go func() {
			if _, err := db.Get().Exec(`DELETE FROM "session" WHERE session = ?`, session.Value); err != nil {
				slog.Error("Error deleting session", "error", err)
			}
		}()
		return nil
	}

	return s
}

// IsAdmin returns whether the request is made by the administrator.
func IsAdmin(r *http.Request) bool {
	session := GetSession(r)

	if session == nil {
		return false
	}

	var email string

	if err := db.Get().Get(&email, `SELECT email FROM "user" WHERE id = ?`, session.UserID); err != nil {
		return false
	}

	return email == "admin"
}

// GetUser returns the signed-in user for the request.
func GetUser(r *http.Request) *model.User {
	session := GetSession(r)

	if session == nil {
		return nil
	}

	user := new(model.User)

	if err := db.Get().Get(user, `SELECT * FROM "user" WHERE id = ?`, session.UserID); err != nil {
		return nil
	}

	return user
}
