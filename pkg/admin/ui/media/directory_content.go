package media

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"net/http"
)

// DirectoryContent returns all files inside a media directory.
func DirectoryContent(w http.ResponseWriter, r *http.Request) {
	tpl.Get().Execute(w, "media-files.html", struct{}{})
}
