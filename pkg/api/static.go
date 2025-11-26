package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/emvi/shifu/pkg/static"
)

// ListStaticFiles lists all files in the static directory.
func ListStaticFiles(w http.ResponseWriter, _ *http.Request) {
	sendJSON(w, struct {
		Files []static.File `json:"files"`
	}{
		Files: static.List(),
	})
}

// PutStaticFile uploads a static file.
func PutStaticFile(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	body, err := io.ReadAll(r.Body)

	if err != nil {
		sendError(w, errors.New("error reading request body"))
	}

	if err := static.Put(path, body); err != nil {
		sendError(w, err)
	}
}
