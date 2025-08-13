package ui

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/middleware"
	"net/http"
)

// Toolbar renders the CMS toolbar.
func Toolbar(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "toolbar.html", struct {
		Admin    bool
		Language string
		Lang     string
		Path     string
	}{
		Admin:    middleware.IsAdmin(r),
		Language: shared.GetLanguage(r),
		Lang:     tpl.GetUILanguage(r),
		Path:     r.URL.Query().Get("path"),
	})
}
