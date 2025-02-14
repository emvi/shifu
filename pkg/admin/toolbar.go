package admin

import "net/http"

func Toolbar(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "toolbar.html", nil)
}
