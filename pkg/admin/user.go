package admin

import "net/http"

func User(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "user.html", nil)
}
