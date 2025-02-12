package admin

import "net/http"

// Login serves the login page.
func Login(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.Write([]byte("<h1>Login</h1>"))
}
