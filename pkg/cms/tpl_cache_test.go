package cms

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestCache(t *testing.T) {
	cache := NewCache("cache_test/*", nil, false)
	data := "World!"
	recorder := httptest.NewRecorder()
	cache.Execute(recorder, "test.html", data)
	body := recorder.Body.String()
	assert.Equal(t, "<p>Hello <span>World!</span>\n</p>\n", body)
	cache.Clear()
	assert.False(t, cache.loaded)
	cache.Disable()
	assert.True(t, cache.disabled)
	cache.Enable()
	assert.False(t, cache.disabled)
}
