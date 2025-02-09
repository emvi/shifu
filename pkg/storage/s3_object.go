package storage

import (
	"github.com/minio/minio-go/v7"
	"os"
	"time"
)

// S3Object contains the file meta information for an S3 object.
type S3Object struct {
	info minio.ObjectInfo
}

// Name implements the os.FileInfo interface.
func (object *S3Object) Name() string {
	return object.info.Key
}

// Size implements the os.FileInfo interface.
func (object *S3Object) Size() int64 {
	return object.info.Size
}

// Mode implements the os.FileInfo interface.
func (object *S3Object) Mode() os.FileMode {
	return 0
}

// ModTime implements the os.FileInfo interface.
func (object *S3Object) ModTime() time.Time {
	return object.info.LastModified
}

// IsDir implements the os.FileInfo interface.
func (object *S3Object) IsDir() bool {
	return false
}

// Sys implements the os.FileInfo interface.
func (object *S3Object) Sys() any {
	return nil
}
