package llms

import (
	"context"
	"net/http"
	"strings"
)

const (
	markdownExt   = ".md"
	markdownIndex = "/index"
)

type markdownRequestKey struct{}

// IsMarkdownPath returns true if the path requests the Markdown version of a page.
func IsMarkdownPath(path string) bool {
	return strings.HasSuffix(path, markdownExt)
}

// MarkdownPath returns the path of the Markdown version for given page path.
func MarkdownPath(path string) string {
	path = strings.TrimSuffix(path, "/")

	if path == "" {
		path = markdownIndex
	}

	return path + markdownExt
}

// PagePath returns the page path for given Markdown path.
func PagePath(path string) string {
	path = strings.TrimSuffix(path, markdownExt)

	if path == markdownIndex || path == "" {
		return "/"
	}

	return path
}

// IsMarkdownRequest returns true if the page is rendered to create its Markdown version.
func IsMarkdownRequest(r *http.Request) bool {
	if r == nil {
		return false
	}

	v, _ := r.Context().Value(markdownRequestKey{}).(bool)
	return v
}

// markdownRequest returns a copy of the request for the page path, marked as Markdown request.
func markdownRequest(r *http.Request, path string) *http.Request {
	req := r.Clone(context.WithValue(r.Context(), markdownRequestKey{}, true))
	req.URL.Path = path
	return req
}
