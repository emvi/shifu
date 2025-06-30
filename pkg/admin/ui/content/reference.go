package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"net/http"
)

// Reference renders the reference JSON editor.
func Reference(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	if r.Method == http.MethodPost {
		// TODO
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "refs-ref.html", struct {
		Lang  string
		Path  string
		JSON  string
		Saved bool
	}{
		Lang: lang,
		Path: path,
		JSON: "{}", // TODO
	})
}
