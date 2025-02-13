package admin

import (
	"github.com/emvi/shifu/pkg/admin/model"
	"net/http"
)

type LoginForm struct {
	Email string
	Error string
}

// Login serves the login page.
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			tpl.Execute(w, "login-form.html", LoginForm{
				Error: "error parsing form",
			})
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		var user model.User

		if err := db.Get(&user, `SELECT * FROM "user" WHERE email = ?`, email); err != nil {
			tpl.Execute(w, "login-form.html", LoginForm{
				Email: email,
				Error: "user not found",
			})
			return
		}

		if !ComparePassword(password+user.PasswordSalt, user.Password) {
			tpl.Execute(w, "login-form.html", LoginForm{
				Email: email,
				Error: "user not found",
			})
			return
		}

		w.Header().Add("HX-Redirect", "/")
		//http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	tpl.Execute(w, "login.html", nil)
}
