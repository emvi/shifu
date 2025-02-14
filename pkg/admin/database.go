package admin

import "net/http"

func Database(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "database.html", nil)
}
