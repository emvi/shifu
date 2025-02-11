package api

import (
	"github.com/emvi/shifu/pkg/content"
	"net/http"
)

// ListContentFiles lists all files in the content directory.
func ListContentFiles(w http.ResponseWriter, _ *http.Request) {
	sendJSON(w, struct {
		Files []content.File `json:"files"`
	}{
		Files: content.List(),
	})
}

// GetContentFile returns a content file.
func GetContentFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	file, err := content.Get(path)

	if err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, struct {
		Content string `json:"content"`
	}{
		Content: string(file),
	})
}
