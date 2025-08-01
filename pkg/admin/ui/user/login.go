package user

import (
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/model"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/util"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// LoginForm is the login form data and errors.
type LoginForm struct {
	Lang  string
	Email string
	Error string
}

// Login serves the login page.
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			slog.Error("Error parsing form", "error", err)
			tpl.Get().Execute(w, "login-form.html", LoginForm{
				Error: "error parsing form",
			})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		email := strings.ToLower(r.FormValue("email"))
		password := r.FormValue("password")
		stayLoggedIn := r.FormValue("stay_logged_in")
		var user model.User

		if err := db.Get().Get(&user, `SELECT * FROM "user" WHERE lower(email) = ?`, email); err != nil {
			tpl.Get().Execute(w, "login-form.html", LoginForm{
				Lang:  tpl.GetUILanguage(r),
				Email: email,
				Error: "user not found",
			})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !util.ComparePassword(password+user.PasswordSalt, user.Password) {
			tpl.Get().Execute(w, "login-form.html", LoginForm{
				Lang:  tpl.GetUILanguage(r),
				Email: email,
				Error: "user not found",
			})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var session string
		exists := true

		for exists {
			session = util.GenRandomString(40)

			if err := db.Get().Get(&exists, `SELECT EXISTS (SELECT 1 FROM "session" WHERE session = ?)`, session); err != nil {
				slog.Error("Error reading session", "error", err)
				tpl.Get().Execute(w, "login-form.html", LoginForm{
					Lang:  tpl.GetUILanguage(r),
					Email: email,
					Error: "error creating session",
				})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		days := 1

		if stayLoggedIn == "on" {
			days = 7
		}

		expires := time.Now().UTC().Add(time.Hour * 24 * time.Duration(days))

		if _, err := db.Get().Exec(`INSERT INTO "session" (user_id, session, expires) VALUES (?, ?, ?)`, user.ID, session, expires); err != nil {
			slog.Error("Error storing session", "error", err)
			tpl.Get().Execute(w, "login-form.html", LoginForm{
				Lang:  tpl.GetUILanguage(r),
				Email: email,
				Error: "error creating session",
			})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    session,
			Path:     "/",
			Secure:   cfg.Get().Server.SecureCookies,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Expires:  expires,
		})
		w.Header().Add("HX-Redirect", "/")
		return
	}

	tpl.Get().Execute(w, "login.html", LoginForm{
		Lang: tpl.GetUILanguage(r),
	})
}
