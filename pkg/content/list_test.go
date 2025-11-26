package content

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	assert.NoError(t, os.RemoveAll("content"))
	assert.NoError(t, os.Mkdir("content", 0744))
	assert.NoError(t, os.WriteFile("content/test.json", []byte("{}"), 0644))
	assert.NoError(t, os.WriteFile("content/test.txt", []byte("txt"), 0644))
	files := List()
	assert.Equal(t, 1, len(files))
	assert.Equal(t, "content/test.json", files[0].Path)
	assert.Equal(t, int64(2), files[0].Size)
	assert.InDelta(t, time.Now().Unix(), files[0].LastModified.Unix(), 2)
}
