package shared

import (
	"net/http"
	"strings"
)

// GetLanguage returns the language query parameter.
func GetLanguage(r *http.Request) string {
	return strings.ToLower(strings.TrimSpace(r.URL.Query().Get("language")))
}
