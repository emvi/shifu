package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// S3 serves files from an S3 bucket.
type S3 struct {
	client     *minio.Client
	bucket     string
	baseDir    string
	pathPrefix string
}

// NewS3 creates a new S3 Storage provider.
func NewS3(baseDir, pathPrefix string) *S3 {
	staticCfg := cfg.Get().Storage
	client, err := minio.New(staticCfg.URL, &minio.Options{
		Secure: true,
		Creds: credentials.NewStaticV4(
			staticCfg.AccessKey,
			staticCfg.Secret,
			"",
		),
	})

	if err != nil {
		slog.Error("Error creating client", "error", err)
		panic(err)
	}

	return &S3{
		client:     client,
		bucket:     staticCfg.Bucket,
		baseDir:    baseDir,
		pathPrefix: pathPrefix,
	}
}

// List implements the Storage interface.
func (storage *S3) List(prefix string, recursive bool) ([]string, error) {
	if storage.pathPrefix != "" && storage.pathPrefix != "." {
		prefix = filepath.Join(storage.pathPrefix, prefix)
	}

	objects := storage.client.ListObjects(context.Background(), storage.bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})
	files := make([]string, 0)

	for object := range objects {
		if object.Err != nil {
			slog.Error("Error listing objects", "error", object.Err)
			return nil, object.Err
		}

		files = append(files, object.Key)
	}

	return files, nil
}

// Exists implements the Storage interface.
func (storage *S3) Exists(path string) (bool, string) {
	path = storage.getPath(path)
	info, err := storage.client.StatObject(context.Background(),
		storage.bucket,
		path,
		minio.StatObjectOptions{})

	if err != nil {
		return false, ""
	}

	return true, storage.getPublicURL(info.Key)
}

// Stat implements the Storage interface.
func (storage *S3) Stat(path string) (os.FileInfo, error) {
	path = storage.getPath(path)
	info, err := storage.client.StatObject(context.Background(),
		storage.bucket,
		path,
		minio.StatObjectOptions{})

	if err != nil {
		return nil, err
	}

	return &S3Object{info}, nil
}

// Read implements the Storage interface.
func (storage *S3) Read(path string) ([]byte, error) {
	path = storage.getPath(path)
	object, err := storage.client.GetObject(context.Background(),
		storage.bucket,
		path,
		minio.GetObjectOptions{})

	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(object)

	if err != nil {
		slog.Error("Error reading object", "error", err, "path", path)
		return nil, err
	}

	return data, nil
}

// Write implements the Storage interface.
func (storage *S3) Write(path string, data []byte, _ *WriteOptions) (string, error) {
	path = storage.getPath(path)
	info, err := storage.client.PutObject(context.Background(),
		storage.bucket,
		path,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{})

	if err != nil {
		return "", err
	}

	return storage.getPublicURL(info.Key), nil
}

// WriteStream implements the Storage interface.
func (storage *S3) WriteStream(string, io.Reader) (string, error) {
	return "", errors.New("not implemented")
}

// Delete implements the Storage interface.
func (storage *S3) Delete(path string) error {
	path = storage.getPath(path)

	if err := storage.client.RemoveObject(context.Background(),
		storage.bucket,
		path,
		minio.RemoveObjectOptions{}); err != nil {
		return err
	}

	return nil
}

func (storage *S3) getPath(path string) string {
	return filepath.Join(storage.pathPrefix, strings.TrimPrefix(path, storage.baseDir))
}

func (storage *S3) getPublicURL(path string) string {
	return fmt.Sprintf("https://%s.fsn1.your-objectstorage.com/%s", storage.bucket, path)
}
