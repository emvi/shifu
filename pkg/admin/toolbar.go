package admin

import "net/http"

func Toolbar(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, "toolbar.html", nil)
}
