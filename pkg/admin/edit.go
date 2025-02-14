package admin

import "net/http"

func Edit(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "edit.html", nil)
}
