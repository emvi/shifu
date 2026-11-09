package media

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/stretchr/testify/assert"
)

func TestUploadFilesTooLarge(t *testing.T) {
	baseDir := cfg.Get().BaseDir
	cfg.Get().BaseDir = t.TempDir()
	defer func() {
		cfg.Get().BaseDir = baseDir
	}()
	assert.NoError(t, os.MkdirAll(getDirectoryPath("/"), 0755))
	srv := httptest.NewServer(http.HandlerFunc(UploadFiles))
	defer srv.Close()

	for _, unpack := range []bool{false, true} {
		pr, pw := io.Pipe()
		w := multipart.NewWriter(pw)
		complete := make(chan bool, 1)

		// stream more than the maximum file size, followed by a second file
		go func() {
			var n int64
			ok := false
			defer func() { complete <- ok }()

			if unpack {
				_ = w.WriteField("unzip", "on")
			}

			f, _ := w.CreateFormFile("files", "big.zip")
			chunk := bytes.Repeat([]byte("x"), 1<<20)

			for n <= maxSize {
				written, err := f.Write(chunk)
				n += int64(written)

				if err != nil {
					pw.CloseWithError(err)
					return
				}
			}

			f, _ = w.CreateFormFile("files", "small.txt")
			_, _ = f.Write([]byte("small"))
			ok = w.Close() == nil
			_ = pw.Close()
		}()
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/?path=/", pr)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", w.FormDataContentType())
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		<-complete
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		assert.Contains(t, string(body), "A file exceeds the maximum upload size of 100 MB.")
		assert.NoFileExists(t, filepath.Join(getDirectoryPath("/"), "big.zip"))
	}
}

func TestUploadFilesForm(t *testing.T) {
	w := httptest.NewRecorder()
	UploadFiles(w, httptest.NewRequest(http.MethodGet, "/media/file/upload?path=/", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "const maxSize =  100000000 ;")
	assert.Contains(t, w.Body.String(), `const tooLargeMessage = "A file exceeds the maximum upload size of 100 MB.";`)
}
