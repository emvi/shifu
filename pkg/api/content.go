package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/emvi/shifu/pkg/content"
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

// PutContentFile uploads a content file.
func PutContentFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	body, err := io.ReadAll(r.Body)

	if err != nil {
		sendError(w, errors.New("error reading request body"))
	}

	if err := content.Put(path, body); err != nil {
		sendError(w, err)
	}
}
