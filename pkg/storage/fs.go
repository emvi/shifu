package storage

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// FileStorage serves files from the file system.
type FileStorage struct{}

// NewFileStorage creates a new FileStorage provider.
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// List implements the Storage interface.
func (storage *FileStorage) List(string, bool) ([]string, error) {
	return nil, errors.New("not implemented")
}

// Exists implements the Storage interface.
func (storage *FileStorage) Exists(path string) (bool, string) {
	if _, err := os.Stat(path); err != nil {
		return false, ""
	}

	return true, path
}

// Read implements the Storage interface.
func (storage *FileStorage) Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		slog.Debug("Error reading file", "error", err, "path", path)
		return nil, err
	}

	return data, nil
}

// Write implements the Storage interface.
func (storage *FileStorage) Write(path string, data []byte, _ *WriteOptions) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		slog.Error("Error creating directory", "error", err, "path", path)
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		slog.Error("Error writing file", "error", err, "path", path)
		return "", err
	}

	return path, nil
}

// WriteStream implements the Storage interface.
func (storage *FileStorage) WriteStream(path string, data io.Reader) (string, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0744); err != nil {
		slog.Error("Error creating directory", "error", err, "path", path)
		return "", err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)

	if err != nil {
		return "", err
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Error writing file", "error", err, "path", path)
		}
	}()

	for {
		buffer := make([]byte, 104857600)
		n, err := data.Read(buffer)

		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}

		if _, err := file.Write(buffer[:n]); err != nil {
			slog.Error("Error writing file", "error", err, "path", path)
			return "", err
		}
	}

	return path, nil
}

// Delete implements the Storage interface.
func (storage *FileStorage) Delete(path string) error {
	if err := os.Remove(path); err != nil {
		slog.Error("Error deleting file",
			"error", err, "path", path)
		return err
	}

	return nil
}
