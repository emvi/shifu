package media

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/stretchr/testify/assert"
)

func TestUnzip(t *testing.T) {
	dir := t.TempDir()
	data := createZip(t, map[string]string{
		"logo.svg":                "logo",
		"Bilder/Foto 1.jpg":       "photo",
		"Bilder/Sub/a.png":        "a",
		"empty/":                  "",
		"../../evil.txt":          "evil",
		"/absolute/abs.txt":       "abs",
		"__MACOSX/Bilder/._a.png": "junk",
		"Bilder/.DS_Store":        "junk",
		"windows\\path\\file.txt": "win",
	})
	existing, err := unzip(bytes.NewReader(data), int64(len(data)), dir, false)
	assert.NoError(t, err)
	assert.Empty(t, existing)
	assert.Equal(t, "logo", readFile(t, filepath.Join(dir, "logo.svg")))
	assert.Equal(t, "photo", readFile(t, filepath.Join(dir, "Bilder", "Foto_1.jpg")))
	assert.Equal(t, "a", readFile(t, filepath.Join(dir, "Bilder", "Sub", "a.png")))
	assert.DirExists(t, filepath.Join(dir, "empty"))
	assert.Equal(t, "evil", readFile(t, filepath.Join(dir, "evil.txt")))
	assert.Equal(t, "abs", readFile(t, filepath.Join(dir, "absolute", "abs.txt")))
	assert.Equal(t, "win", readFile(t, filepath.Join(dir, "windows", "path", "file.txt")))
	assert.NoDirExists(t, filepath.Join(dir, "__MACOSX"))
	assert.NoFileExists(t, filepath.Join(dir, "Bilder", "DS_Store"))
	assert.NoFileExists(t, filepath.Join(filepath.Dir(filepath.Dir(dir)), "evil.txt"))

	// nothing is extracted if a file exists already, unless it should be overwritten
	data = createZip(t, map[string]string{"logo.svg": "new", "Bilder/Foto 1.jpg": "new", "new.txt": "new"})
	existing, err = unzip(bytes.NewReader(data), int64(len(data)), dir, false)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"logo.svg", "Bilder/Foto_1.jpg"}, existing)
	assert.Equal(t, "logo", readFile(t, filepath.Join(dir, "logo.svg")))
	assert.NoFileExists(t, filepath.Join(dir, "new.txt"))
	existing, err = unzip(bytes.NewReader(data), int64(len(data)), dir, true)
	assert.NoError(t, err)
	assert.Empty(t, existing)
	assert.Equal(t, "new", readFile(t, filepath.Join(dir, "logo.svg")))
	assert.Equal(t, "new", readFile(t, filepath.Join(dir, "new.txt")))

	// existing directories are fine
	data = createZip(t, map[string]string{"Bilder/": "", "Bilder/other.jpg": "other"})
	existing, err = unzip(bytes.NewReader(data), int64(len(data)), dir, false)
	assert.NoError(t, err)
	assert.Empty(t, existing)

	// invalid zip
	_, err = unzip(bytes.NewReader([]byte("no zip")), 6, dir, false)
	assert.Error(t, err)
}

func TestUnzipLimits(t *testing.T) {
	dir := t.TempDir()
	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)

	for i := range maxUnzipFiles + 1 {
		_, err := w.Create(filepath.Join("dir", string(rune('a'+i%26)), "f"+string(rune('a'+i%26))))
		assert.NoError(t, err)
	}

	assert.NoError(t, w.Close())
	_, err := unzip(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()), dir, false)
	assert.ErrorIs(t, err, errUnzipFiles)

	// the uncompressed size is limited
	limit := maxUnzipSize
	maxUnzipSize = 10
	defer func() {
		maxUnzipSize = limit
	}()
	data := createZip(t, map[string]string{"big.txt": "more than ten bytes"})
	_, err = unzip(bytes.NewReader(data), int64(len(data)), dir, false)
	assert.ErrorIs(t, err, errUnzipSize)
	assert.NoFileExists(t, filepath.Join(dir, "big.txt"))

	// the actual size is checked while extracting, in case the size in the header is wrong
	n, err := extractFile(zipEntry(t, createZip(t, map[string]string{"big.txt": "more than ten bytes"})), filepath.Join(dir, "big.txt"), 10)
	assert.ErrorIs(t, err, errUnzipSize)
	assert.Equal(t, int64(11), n)
	assert.NoFileExists(t, filepath.Join(dir, "big.txt"))
}

func TestUploadFilesUnzip(t *testing.T) {
	baseDir := cfg.Get().BaseDir
	cfg.Get().BaseDir = t.TempDir()
	defer func() {
		cfg.Get().BaseDir = baseDir
	}()
	assert.NoError(t, os.MkdirAll(getDirectoryPath("/target"), 0755))
	upload := func(unpack bool) *httptest.ResponseRecorder {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)

		if unpack {
			assert.NoError(t, w.WriteField("unzip", "on"))
		}

		f, err := w.CreateFormFile("files", "archive.zip")
		assert.NoError(t, err)
		_, _ = f.Write(createZip(t, map[string]string{"sub/file.txt": "content"}))
		f, err = w.CreateFormFile("files", "plain.txt")
		assert.NoError(t, err)
		_, _ = f.Write([]byte("plain"))
		assert.NoError(t, w.Close())
		r := httptest.NewRequest(http.MethodPost, "/media/file/upload?path=/target", &body)
		r.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		UploadFiles(rec, r)
		return rec
	}

	assert.Equal(t, http.StatusOK, upload(true).Code)
	assert.Equal(t, "content", readFile(t, filepath.Join(getDirectoryPath("/target"), "sub", "file.txt")))
	assert.Equal(t, "plain", readFile(t, filepath.Join(getDirectoryPath("/target"), "plain.txt")))
	assert.NoFileExists(t, filepath.Join(getDirectoryPath("/target"), "archive.zip"))

	// without the checkbox, the zip file is uploaded as is (plain.txt exists already and is skipped)
	rec := upload(false)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<li>plain.txt</li>")
	assert.FileExists(t, filepath.Join(getDirectoryPath("/target"), "archive.zip"))

	// unpacking again lists the existing files from the zip file
	rec = upload(true)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<li>sub/file.txt</li>")
}

func TestUploadFilesOverwrite(t *testing.T) {
	baseDir := cfg.Get().BaseDir
	cfg.Get().BaseDir = t.TempDir()
	defer func() {
		cfg.Get().BaseDir = baseDir
	}()
	assert.NoError(t, os.MkdirAll(getDirectoryPath("/"), 0755))
	upload := func(overwrite bool, files map[string]string) *httptest.ResponseRecorder {
		var body bytes.Buffer
		w := multipart.NewWriter(&body)

		if overwrite {
			assert.NoError(t, w.WriteField("overwrite", "on"))
		}

		for name, content := range files {
			f, err := w.CreateFormFile("files", name)
			assert.NoError(t, err)
			_, _ = f.Write([]byte(content))
		}

		assert.NoError(t, w.Close())
		r := httptest.NewRequest(http.MethodPost, "/media/file/upload?path=/", &body)
		r.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		UploadFiles(rec, r)
		return rec
	}

	assert.Equal(t, http.StatusOK, upload(false, map[string]string{"file.txt": "1"}).Code)

	// the existing file is skipped and listed, the other file is uploaded anyway
	rec := upload(false, map[string]string{"file.txt": "2", "other.txt": "other"})
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "<li>file.txt</li>")
	assert.NotContains(t, rec.Body.String(), "<li>other.txt</li>")
	assert.Equal(t, "1", readFile(t, filepath.Join(getDirectoryPath("/"), "file.txt")))
	assert.Equal(t, "other", readFile(t, filepath.Join(getDirectoryPath("/"), "other.txt")))
	assert.NoFileExists(t, filepath.Join(getDirectoryPath("/"), "_file.txt"))

	assert.Equal(t, http.StatusOK, upload(true, map[string]string{"file.txt": "3"}).Code)
	assert.Equal(t, "3", readFile(t, filepath.Join(getDirectoryPath("/"), "file.txt")))
}

func createZip(t *testing.T, files map[string]string) []byte {
	var buffer bytes.Buffer
	w := zip.NewWriter(&buffer)

	for name, content := range files {
		f, err := w.Create(name)
		assert.NoError(t, err)
		_, err = f.Write([]byte(content))
		assert.NoError(t, err)
	}

	assert.NoError(t, w.Close())
	return buffer.Bytes()
}

func readFile(t *testing.T, path string) string {
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	return string(content)
}

func zipEntry(t *testing.T, data []byte) *zip.File {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	assert.NoError(t, err)
	return r.File[0]
}
