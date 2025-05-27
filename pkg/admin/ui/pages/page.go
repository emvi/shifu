package pages

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"net/http"
	"strings"
)

// Page renders the page details.
func Page(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getDirectoryPath(path)

	tpl.Get().Execute(w, "pages-page.html", struct {
		Path string
	}{
		Path: fullPath,
	})
}
