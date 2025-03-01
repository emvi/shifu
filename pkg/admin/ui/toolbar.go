package ui

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"net/http"
)

// Toolbar renders the CMS toolbar.
func Toolbar(w http.ResponseWriter, _ *http.Request) {
	tpl.Get().Execute(w, "toolbar.html", nil)
}
