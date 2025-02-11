package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"syscall"
)

func sendJSON(w http.ResponseWriter, v any) {
	resp, err := json.Marshal(v)

	if err != nil {
		slog.Error("Error marshalling response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(resp); err != nil && !errors.Is(err, syscall.EPIPE) {
		slog.Error("Error writing response", "error", err)
	}
}

func sendError(w http.ResponseWriter, err error) {
	sendJSON(w, struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
