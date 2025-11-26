package cms

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptedLanguages(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	languages := GetAcceptedLanguages(req)
	assert.Len(t, languages, 0)

	req.Header.Set("Accept-Language", "en")
	languages = GetAcceptedLanguages(req)
	assert.Len(t, languages, 1)
	assert.Equal(t, "en", languages[0])

	req.Header.Set("Accept-Language", "fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7, *;q=0.5")
	languages = GetAcceptedLanguages(req)
	assert.Len(t, languages, 3)
	assert.Equal(t, "fr", languages[0])
	assert.Equal(t, "en", languages[1])
	assert.Equal(t, "de", languages[2])
}
