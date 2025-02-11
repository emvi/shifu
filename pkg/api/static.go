package api

import (
	"github.com/emvi/shifu/pkg/static"
	"net/http"
)

// ListStaticFiles lists all files in the static directory.
func ListStaticFiles(w http.ResponseWriter, _ *http.Request) {
	sendJSON(w, struct {
		Files []static.File `json:"files"`
	}{
		Files: static.List(),
	})
}
