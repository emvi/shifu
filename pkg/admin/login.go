package admin

import "net/http"

// Login serves the login page.
func Login(w http.ResponseWriter, _ *http.Request) {
	tpl.execute(w, "login.html", nil)
}
