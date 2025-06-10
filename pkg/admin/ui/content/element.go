package content

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

func addElement(content *cms.Content, position, template string, positions []string) bool {
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
		Tpl:     template,
		Content: children,
	})
	return true
}

func addReference(content *cms.Content, position, reference string) bool {
	if content.Content == nil {
		content.Content = make(map[string][]cms.Content)
	}

	if _, found := content.Content[position]; !found {
		content.Content[position] = make([]cms.Content, 0)
	}

	content.Content[position] = append(content.Content[position], cms.Content{
		Ref: reference,
	})
	return true
}

func moveElement(content *cms.Content, path string, direction int) bool {
	parentElement, key, index := findParentElement(content, path)

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

func deleteElement(content *cms.Content, path string) bool {
	parentElement, key, index := findParentElement(content, path)

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}

func findParentElement(content *cms.Content, path string) (*cms.Content, string, int) {
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

func findElement(content *cms.Content, path string) *cms.Content {
	parentElement, key, index := findParentElement(content, path)

	if parentElement == nil {
		return nil
	}

	return &parentElement.Content[key][index]
}

func setElement(content *cms.Content, path string, element *cms.Content) bool {
	parentElement, key, index := findParentElement(content, path)

	if parentElement == nil {
		return false
	}

	parentElement.Content[key][index] = *element
	return true
}

func loadRef(name string) (*cms.Content, error) {
	name = fmt.Sprintf("%s.json", name)
	path := ""

	if err := filepath.WalkDir(filepath.Join(cfg.Get().BaseDir, contentDir), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Base(p) == name {
			path = p
			return fs.SkipAll
		}

		return nil
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

func saveRef(element *cms.Content, path string) error {
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
