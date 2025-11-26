package sync

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/content"
	"github.com/emvi/shifu/pkg/static"
)

// Push pushes changed static and content files to a remote Shifu server.
func Push(dir string) error {
	if err := cfg.Load(dir, nil); err != nil {
		return err
	}

	if err := hasRemoteConfig(); err != nil {
		return err
	}

	slog.Info("Pushing static files and content")

	if err := pushContent(); err != nil {
		return err
	}

	if err := pushStatic(); err != nil {
		return err
	}

	if err := updateContent(); err != nil {
		return err
	}

	slog.Info("Done")
	return nil
}

func pushContent() error {
	slog.Info("Pushing content files")
	remoteFiles, err := getRemoteContentFiles()

	if err != nil {
		return err
	}

	localFiles := content.List()
	upload := make([]string, 0)

	for i, localFile := range localFiles {
		found := false

		for j, remoteFile := range remoteFiles {
			if remoteFile.Path == localFile.Path {
				if remoteFiles[j].LastModified.Before(localFiles[i].LastModified) {
					upload = append(upload, localFile.Path)
				}

				found = true
				break
			}
		}

		if !found {
			upload = append(upload, localFile.Path)
		}
	}

	for _, path := range upload {
		if err := uploadContentFile(path); err != nil {
			return err
		}
	}

	return nil
}

func uploadContentFile(path string) error {
	slog.Info("Uploading file", "path", path)
	body, err := os.ReadFile(filepath.Join(cfg.Get().BaseDir, path))

	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/content/file?path=%s", cfg.Get().Remote.URL, url.QueryEscape(path)), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", cfg.Get().Remote.Secret))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("error uploading file")
	}

	return nil
}

func pushStatic() error {
	slog.Info("Pushing static files")
	remoteFiles, err := getRemoteStaticFiles()

	if err != nil {
		return err
	}

	localFiles := static.List()
	upload := make([]string, 0)

	for i, localFile := range localFiles {
		found := false

		for j, remoteFile := range remoteFiles {
			if remoteFile.Path == localFile.Path {
				if remoteFiles[j].LastModified.Before(localFiles[i].LastModified) {
					upload = append(upload, localFile.Path)
				}

				found = true
				break
			}
		}

		if !found {
			upload = append(upload, localFile.Path)
		}
	}

	for _, path := range upload {
		if err := uploadStaticFile(path); err != nil {
			return err
		}
	}

	return nil
}

func uploadStaticFile(path string) error {
	slog.Info("Uploading file", "path", path)
	body, err := os.ReadFile(filepath.Join(cfg.Get().BaseDir, path))

	if err != nil {
		return err
	}

	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/static?path=%s", cfg.Get().Remote.URL, url.QueryEscape(path)), bytes.NewReader(body))
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", cfg.Get().Remote.Secret))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("error uploading file")
	}

	return nil
}

func updateContent() error {
	slog.Info("Updating remote content")
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/cms", cfg.Get().Remote.URL), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", cfg.Get().Remote.Secret))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		slog.Error("Error updating remote content", "error", err)
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing content update response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Error updating remote content", "error", err, "status", resp.StatusCode)
		return errors.New("error updating remote content")
	}

	return nil
}
