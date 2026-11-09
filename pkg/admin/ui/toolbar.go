package ui

import (
	"net/http"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// Toolbar renders the CMS toolbar.
func Toolbar(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "toolbar.html", struct {
		Language string
		Lang     string
		Path     string
	}{
		Language: shared.GetLanguage(r),
		Lang:     tpl.GetUILanguage(r),
		Path:     r.URL.Query().Get("path"),
	})
}
