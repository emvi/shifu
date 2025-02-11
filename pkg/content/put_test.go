package content

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPut(t *testing.T) {
	assert.NoError(t, os.RemoveAll("content"))
	assert.NoError(t, os.Mkdir("content", 0744))
	assert.NoError(t, os.WriteFile("content/test.json", []byte("{}"), 0644))
	assert.NoError(t, Put("/content/test.json", []byte(`{"foo":"bar"}`)))
	assert.FileExists(t, "content/test.json")
	content, err := os.ReadFile("content/test.json")
	assert.NoError(t, err)
	assert.Equal(t, `{"foo":"bar"}`, string(content))
}
