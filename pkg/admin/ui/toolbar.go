package ui

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"net/http"
)

// Toolbar renders the CMS toolbar.
func Toolbar(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "toolbar.html", struct {
		Path string
	}{
		Path: r.URL.Query().Get("path"),
	})
}
