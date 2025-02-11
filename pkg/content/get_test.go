package content

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	assert.NoError(t, os.RemoveAll("content"))
	assert.NoError(t, os.Mkdir("content", 0744))
	assert.NoError(t, os.WriteFile("content/test.json", []byte("{}"), 0644))
	file, err := Get("content/test.json")
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(file))
	file, err = Get("/content/test.json")
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(file))
	_, err = Get("content/foobar.json")
	assert.Equal(t, "file not found", err.Error())
	_, err = Get("foo/bar.json")
	assert.Equal(t, "path does not start with content", err.Error())
	_, err = Get(" ")
	assert.Equal(t, "path empty", err.Error())
}
