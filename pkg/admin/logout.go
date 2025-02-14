package admin

import (
	"github.com/emvi/shifu/pkg/cfg"
	"net/http"
)

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
