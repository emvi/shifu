package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/content"
	"github.com/emvi/shifu/pkg/static"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Pull pulls changed static and content files from a remote Shifu server.
func Pull(dir string) error {
	if err := cfg.Load(dir, nil); err != nil {
		return err
	}

	if err := hasRemoteConfig(); err != nil {
		return err
	}

	slog.Info("Pulling remote static files and content")

	if err := pullContent(); err != nil {
		return err
	}

	if err := pullStatic(); err != nil {
		return err
	}

	slog.Info("Done")
	return nil
}

func hasRemoteConfig() error {
	if cfg.Get().Remote.URL == "" || cfg.Get().Remote.Secret == "" {
		return errors.New("no remote server configured")
	}

	return nil
}

func pullContent() error {
	slog.Info("Pulling content files")

	if err := os.MkdirAll(filepath.Join(cfg.Get().BaseDir, "content"), 0744); err != nil {
		slog.Error("Failed to create content directory", "error", err)
		return err
	}

	remoteFiles, err := getRemoteContentFiles()

	if err != nil {
		return err
	}

	localFiles := content.List()
	download := make([]string, 0)

	for i, remoteFile := range remoteFiles {
		found := false

		for j, localFile := range localFiles {
			if remoteFile.Path == localFile.Path {
				if remoteFiles[i].LastModified.After(localFiles[j].LastModified) {
					download = append(download, remoteFile.Path)
				}

				found = true
				break
			}
		}

		if !found {
			download = append(download, remoteFile.Path)
		}
	}

	for _, path := range download {
		if err := downloadContentFile(path); err != nil {
			return err
		}
	}

	return nil
}

func getRemoteContentFiles() ([]content.File, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/content", cfg.Get().Remote.URL), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", cfg.Get().Remote.Secret))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		slog.Error("Error listing remote content files", "error", err)
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing content response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Error listing remote content files", "error", err, "status", resp.StatusCode)
		return nil, errors.New("error listing remote content files")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		slog.Error("Error reading content response body", "error", err)
		return nil, err
	}

	var remoteFiles struct {
		Files []content.File `json:"files"`
	}

	if err := json.Unmarshal(body, &remoteFiles); err != nil {
		slog.Error("Error unmarshalling remote content files", "error", err)
		return nil, err
	}

	for i := range remoteFiles.Files {
		remoteFiles.Files[i].Path = strings.TrimPrefix(remoteFiles.Files[i].Path, "/")
	}

	return remoteFiles.Files, nil
}

func downloadContentFile(path string) error {
	slog.Info("Downloading remote file", "path", path)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/content/file?path=%s", cfg.Get().Remote.URL, url.QueryEscape(path)), nil)
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
		return errors.New("error downloading remote file")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var file struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(body, &file); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cfg.Get().BaseDir, filepath.Dir(path)), 0744); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(cfg.Get().BaseDir, path), []byte(file.Content), 0644); err != nil {
		return err
	}

	return nil
}

func pullStatic() error {
	slog.Info("Pulling static files")

	if err := os.MkdirAll(filepath.Join(cfg.Get().BaseDir, "static"), 0744); err != nil {
		slog.Error("Failed to create static directory", "error", err)
		return err
	}

	remoteFiles, err := getRemoteStaticFiles()

	if err != nil {
		return err
	}

	localFiles := static.List()
	download := make([]string, 0)

	for i, remoteFile := range remoteFiles {
		found := false

		for j, localFile := range localFiles {
			if remoteFile.Path == localFile.Path {
				if remoteFiles[i].LastModified.After(localFiles[j].LastModified) {
					download = append(download, remoteFile.Path)
				}

				found = true
				break
			}
		}

		if !found {
			download = append(download, remoteFile.Path)
		}
	}

	for _, path := range download {
		if err := downloadStaticFile(path); err != nil {
			return err
		}
	}

	return nil
}

func getRemoteStaticFiles() ([]static.File, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/static", cfg.Get().Remote.URL), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Key %s", cfg.Get().Remote.Secret))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		slog.Error("Error listing remote static files", "error", err)
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing static response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Error listing remote static files", "error", err, "status", resp.StatusCode)
		return nil, errors.New("error listing remote static files")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		slog.Error("Error reading static response body", "error", err)
		return nil, err
	}

	var remoteFiles struct {
		Files []static.File `json:"files"`
	}

	if err := json.Unmarshal(body, &remoteFiles); err != nil {
		slog.Error("Error unmarshalling remote static files", "error", err)
		return nil, err
	}

	for i := range remoteFiles.Files {
		remoteFiles.Files[i].Path = strings.TrimPrefix(remoteFiles.Files[i].Path, "/")
	}

	return remoteFiles.Files, nil
}

func downloadStaticFile(path string) error {
	slog.Info("Downloading remote file", "path", path)
	resp, err := http.Get(fmt.Sprintf("%s/%s", cfg.Get().Remote.URL, path))

	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Warn("Error closing response body", "error", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New("error downloading remote file")
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cfg.Get().BaseDir, filepath.Dir(path)), 0744); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(cfg.Get().BaseDir, path), body, 0644); err != nil {
		return err
	}

	return nil
}
