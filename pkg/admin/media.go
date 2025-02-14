package admin

import "net/http"

func Media(w http.ResponseWriter, _ *http.Request) {
	tpl.Execute(w, "media.html", nil)
}
