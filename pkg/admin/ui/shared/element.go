package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	contentDir = "content"
)

// AddElement adds a new empty element to the tree.
func AddElement(content *cms.Content, position, template string, positions []string) bool {
	if content.Content == nil {
		content.Content = make(map[string][]cms.Content)
	}

	if _, found := content.Content[position]; !found {
		content.Content[position] = make([]cms.Content, 0)
	}

	children := make(map[string][]cms.Content)

	for _, position := range positions {
		children[position] = make([]cms.Content, 0)
	}

	content.Content[position] = append(content.Content[position], cms.Content{
		Tpl: template,
		// TODO ref
		Content: children,
	})
	return true
}

// MoveElement finds an element by JSON path and moves it up or down.
func MoveElement(content *cms.Content, path string, direction int) bool {
	parentElement, key, index := FindParentElement(content, path)

	if parentElement != nil {
		if index == 0 && direction == -1 || index == len(parentElement.Content[key])-1 && direction == 1 {
			return false
		}

		if direction == -1 {
			parentElement.Content[key][index], parentElement.Content[key][index-1] = parentElement.Content[key][index-1], parentElement.Content[key][index]
		} else if direction == 1 {
			parentElement.Content[key][index], parentElement.Content[key][index+1] = parentElement.Content[key][index+1], parentElement.Content[key][index]
		}

		return true
	}

	return false
}

// DeleteElement finds an element by JSON path and removes it from the tree.
func DeleteElement(content *cms.Content, path string) bool {
	parentElement, key, index := FindParentElement(content, path)

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}

// FindParentElement finds an element by JSON path.
// The path is a dot-separated list of keys and indices.
func FindParentElement(content *cms.Content, path string) (*cms.Content, string, int) {
	if path == "" {
		return nil, "", 0
	}

	parts := strings.Split(path, ".")
	var elements []cms.Content
	element := content
	var index int
	var key string
	var parentElement *cms.Content

	for i, part := range parts {
		if i%2 != 0 {
			var err error
			index, err = strconv.Atoi(part)

			if err != nil {
				return nil, "", 0
			}

			element = &elements[index]
		} else {
			elements = element.Content[part]
			key = part
			parentElement = element
		}
	}

	return parentElement, key, index
}

// FindElement finds an element by JSON path.
// The path is a dot-separated list of keys and indices.
func FindElement(content *cms.Content, path string) *cms.Content {
	parentElement, key, index := FindParentElement(content, path)

	if parentElement == nil {
		return nil
	}

	return &parentElement.Content[key][index]
}

// SetElement finds an element by JSON path and updates it.
// The path is a dot-separated list of keys and indices.
func SetElement(content *cms.Content, path string, element *cms.Content) bool {
	parentElement, key, index := FindParentElement(content, path)

	if parentElement == nil {
		return false
	}

	parentElement.Content[key][index] = *element
	return true
}

// LoadRef loads a reference for the given name and parses it into a cms.Content object.
func LoadRef(name string) (*cms.Content, error) {
	name = fmt.Sprintf("%s.json", name)
	path := ""

	if err := filepath.WalkDir(filepath.Join(cfg.Get().BaseDir, contentDir), func(p string, d fs.DirEntry, err error) error {
		if filepath.Base(p) == name {
			path = p
			return fs.SkipAll
		}

		return err
	}); err != nil && !errors.Is(err, fs.SkipAll) {
		return nil, err
	}

	if path == "" {
		return nil, errors.New("file not found")
	}

	content, err := os.ReadFile(path)

	if err != nil {
		slog.Error("Error reading reference file", "error", err)
		return nil, err
	}

	var element cms.Content

	if err := json.Unmarshal(content, &element); err != nil {
		slog.Error("Error parsing reference file", "error", err)
		return nil, err
	}

	element.File = path
	return &element, nil
}

// SaveRef saves a reference to the given path.
func SaveRef(element *cms.Content, path string) error {
	elementJson, err := json.Marshal(element)

	if err != nil {
		slog.Error("Error marshalling reference data", "error", err)
		return err
	}

	if err := os.WriteFile(path, elementJson, 0644); err != nil {
		slog.Error("Error writing reference data", "error", err)
		return err
	}

	return nil
}
