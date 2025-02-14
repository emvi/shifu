package admin

import "net/http"

func Pages(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "pages.html", nil)
}
