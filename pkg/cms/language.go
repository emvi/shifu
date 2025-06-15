package cms

import (
	"net/http"
	"slices"
	"strings"
)

// GetAcceptedLanguages returns the accepted languages for the given request.
func GetAcceptedLanguages(r *http.Request) []string {
	header := r.Header.Get("Accept-Language")
	parts := strings.Split(header, ",")
	languages := make([]string, 0)

	for _, part := range parts {
		left, _, _ := strings.Cut(strings.TrimSpace(part), ";")

		if strings.Contains(left, "-") {
			left, _, _ = strings.Cut(left, "-")
		}

		if left != "" && len(left) == 2 && !slices.Contains(languages, left) {
			languages = append(languages, left)
		}
	}

	return languages
}
