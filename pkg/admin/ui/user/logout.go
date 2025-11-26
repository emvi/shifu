package user

import (
	"net/http"

	"github.com/emvi/shifu/pkg/cfg"
)

// Logout signs out the user and redirects to the home page.
func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Path:     "/",
		Secure:   cfg.Get().Server.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
