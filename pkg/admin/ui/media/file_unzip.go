package media

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUnzipFiles = 10_000
)

var (
	maxUnzipSize int64 = 1_000_000_000 // 1 GB uncompressed

	errUnzipSize  = errors.New("the zip file exceeds the maximum uncompressed size")
	errUnzipFiles = errors.New("the zip file contains too many files")
)

func isZip(filename string) bool {
	return strings.ToLower(filepath.Ext(filename)) == ".zip"
}

func unzipUpload(r io.Reader, dir string, overwrite bool) ([]string, bool, error) {
	tmp, err := os.CreateTemp("", "shifu-upload-*.zip")

	if err != nil {
		return nil, false, err
	}

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()
	written, err := io.CopyN(tmp, r, maxSize+1)

	if err != nil && err != io.EOF {
		return nil, false, err
	}

	if written > maxSize {
		return nil, false, errFileTooLarge
	}

	return unzip(tmp, written, dir, overwrite)
}

func unzip(r io.ReaderAt, size int64, dir string, overwrite bool) ([]string, bool, error) {
	archive, err := zip.NewReader(r, size)

	if err != nil {
		return nil, false, err
	}

	if len(archive.File) > maxUnzipFiles {
		return nil, false, errUnzipFiles
	}

	var total uint64

	for _, file := range archive.File {
		total += file.UncompressedSize64

		if total > uint64(maxUnzipSize) {
			return nil, false, errUnzipSize
		}
	}

	root, err := filepath.Abs(dir)

	if err != nil {
		return nil, false, err
	}

	type entry struct {
		file   *zip.File
		target string
	}

	// resolve and check all paths before anything is extracted
	entries := make([]entry, 0, len(archive.File))
	existing := make([]string, 0)

	for _, file := range archive.File {
		parts := unzipPath(file.Name)

		if len(parts) == 0 || !(file.Mode().IsRegular() || file.FileInfo().IsDir()) {
			continue
		}

		target := filepath.Join(append([]string{root}, parts...)...)

		if target != root && !strings.HasPrefix(target, root+string(filepath.Separator)) {
			return nil, false, fmt.Errorf("invalid path in zip file: %s", file.Name)
		}

		if !file.FileInfo().IsDir() && !overwrite && fileExists(target) {
			existing = append(existing, strings.Join(parts, "/"))
		}

		entries = append(entries, entry{file, target})
	}

	if len(existing) > 0 {
		return existing, false, nil
	}

	var extracted int64
	dirs := false
	mkdir := func(path string) error {
		if !fileExists(path) {
			dirs = true
		}

		return os.MkdirAll(path, 0755)
	}

	for _, e := range entries {
		if e.file.FileInfo().IsDir() {
			if err := mkdir(e.target); err != nil {
				return nil, dirs, err
			}

			continue
		}

		if err := mkdir(filepath.Dir(e.target)); err != nil {
			return nil, dirs, err
		}

		n, err := extractFile(e.file, e.target, maxUnzipSize-extracted)
		extracted += n

		if err != nil {
			return nil, dirs, err
		}
	}

	return nil, dirs, nil
}

func extractFile(file *zip.File, target string, limit int64) (int64, error) {
	in, err := file.Open()

	if err != nil {
		return 0, err
	}

	defer func() {
		if err := in.Close(); err != nil {
			slog.Debug("Error closing zip file entry", "error", err)
		}
	}()
	out, err := os.Create(target)

	if err != nil {
		return 0, err
	}

	// the size in the zip header can be manipulated, so the actual size is checked while extracting
	written, err := io.CopyN(out, in, limit+1)
	_ = out.Close()

	if err != nil && err != io.EOF {
		_ = os.Remove(target)
		return written, err
	}

	if written > limit {
		_ = os.Remove(target)
		return written, errUnzipSize
	}

	return written, nil
}

func unzipPath(name string) []string {
	parts := make([]string, 0)

	for part := range strings.SplitSeq(strings.ReplaceAll(name, "\\", "/"), "/") {
		if part == "" || part == "." || part == ".." {
			continue
		}

		if part == "__MACOSX" || part == ".DS_Store" || part == "Thumbs.db" {
			return nil
		}

		parts = append(parts, cleanFilename(part))
	}

	return parts
}
