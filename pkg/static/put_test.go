package static

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPut(t *testing.T) {
	assert.NoError(t, os.RemoveAll("static"))
	assert.NoError(t, os.Mkdir("static", 0744))
	assert.NoError(t, os.WriteFile("static/test.txt", []byte("text"), 0644))
	assert.NoError(t, Put("/static/test.txt", []byte("Hello world!")))
	assert.FileExists(t, "static/test.txt")
	content, err := os.ReadFile("static/test.txt")
	assert.NoError(t, err)
	assert.Equal(t, "Hello world!", string(content))
	assert.NoError(t, Put("/static/sub/test.txt", []byte("Hello world!")))
	assert.FileExists(t, "static/sub/test.txt")
}
