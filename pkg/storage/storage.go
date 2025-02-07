package storage

import (
	"io"
)

// WriteOptions are options to write a file.
type WriteOptions struct {
	ContentType  string
	PublicAccess bool
}

// Storage is an interface abstracting the static file handling.
type Storage interface {
	// List lists all files.
	List(string, bool) ([]string, error)

	// Exists tests if the file for given path exists and returns the paths if it does.
	Exists(string) (bool, string)

	// Read reads a file for given path.
	Read(string) ([]byte, error)

	// Write writes a file for given path and returns the path.
	Write(string, []byte, *WriteOptions) (string, error)

	// WriteStream writes a file stream for given path and returns the path.
	WriteStream(string, io.Reader) (string, error)

	// Delete deletes a file for given path.
	Delete(string) error
}
