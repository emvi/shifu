package static

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.NoError(t, os.RemoveAll("static"))
	assert.NoError(t, os.Mkdir("static", 0744))
	assert.NoError(t, os.WriteFile("static/test.txt", []byte("test"), 0644))
	assert.NoError(t, os.WriteFile("static/.secrets.env", []byte("secret"), 0644))
	files := List()
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "static/test.txt", files[0].Path)
	assert.Equal(t, int64(4), files[0].Size)
	assert.InDelta(t, time.Now().Unix(), files[0].LastModified.Unix(), 2)
}
