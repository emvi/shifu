package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
)

const (
	contentDir = "content"
)

func addElement(content *cms.Content, parentPath, position, template string, positions []string) *cms.Content {
	if content.Content == nil {
		content.Content = make(map[string][]cms.Content)
	}

	if _, found := content.Content[position]; !found {
		content.Content[position] = make([]cms.Content, 0)
	}

	children := make(map[string][]cms.Content)

	for _, p := range positions {
		children[p] = make([]cms.Content, 0)
	}

	index := findElementPositionToInsert(content, position)
	n := len(content.Content[position])

	if index > -1 {
		n = index
	}

	pos := ""

	if parentPath != "" {
		pos = fmt.Sprintf("%s.%s.%d", parentPath, position, n)
	} else {
		pos = fmt.Sprintf("%s.%d", position, n)
	}

	element := cms.Content{
		Tpl:      template,
		Content:  children,
		Position: pos,
	}

	if index < 0 {
		content.Content[position] = append(content.Content[position], element)
	} else {
		content.Content[position] = slices.Insert(content.Content[position], index, element)
	}

	return &element
}

func addReference(content *cms.Content, parentPath, position, reference string) *cms.Content {
	if content.Content == nil {
		content.Content = make(map[string][]cms.Content)
	}

	if _, found := content.Content[position]; !found {
		content.Content[position] = make([]cms.Content, 0)
	}

	index := findElementPositionToInsert(content, position)
	n := len(content.Content[position])

	if index > -1 {
		n = index
	}

	pos := ""

	if parentPath != "" {
		pos = fmt.Sprintf("%s.%s.%d", parentPath, position, n)
	} else {
		pos = fmt.Sprintf("%s.%d", position, n)
	}

	element := cms.Content{
		Ref:      reference,
		Position: pos,
	}

	if index < 0 {
		content.Content[position] = append(content.Content[position], element)
	} else {
		content.Content[position] = slices.Insert(content.Content[position], index, element)
	}

	return &element
}

func moveElement(content *cms.Content, path string, direction int) bool {
	parentElement, _, key, index := findParentElement(content, path)

	if parentElement != nil {
		if index == 0 && direction == -1 || index == len(parentElement.Content[key])-1 && direction == 1 {
			return false
		}

		if direction == -1 {
			config, found := findTemplateConfig(parentElement, key, index-1)

			if !found || config.Layout {
				return false
			}

			parentElement.Content[key][index], parentElement.Content[key][index-1] = parentElement.Content[key][index-1], parentElement.Content[key][index]
		} else if direction == 1 {
			config, found := findTemplateConfig(parentElement, key, index+1)

			if !found || config.Layout {
				return false
			}

			parentElement.Content[key][index], parentElement.Content[key][index+1] = parentElement.Content[key][index+1], parentElement.Content[key][index]
		}

		return true
	}

	return false
}

func findElementPositionToInsert(content *cms.Content, position string) int {
	n := len(content.Content[position])

	if n > 0 {
		config, found := findTemplateConfig(content, position, n-1)

		if found && config.Layout {
			return n - 1
		}

		return n
	}

	return -1
}

func findTemplateConfig(content *cms.Content, position string, i int) (TemplateConfig, bool) {
	config, found := tplCfgCache.GetTemplate(content.Content[position][i].Tpl)

	if !found {
		ref, err := loadRef(content.Content[position][i].Ref)

		if err != nil {
			slog.Error("Error loading reference file", "error", err)
			return TemplateConfig{}, false
		}

		config, found = tplCfgCache.GetTemplate(ref.Tpl)
	}

	if !found {
		slog.Warn("Error loading template configuration", "tpl", content.Content[position][i].Tpl, "ref", content.Content[position][i].Ref)
		return TemplateConfig{}, false
	}

	slog.Debug("Found template configuration", "name", config.Name, "layout", config.Layout)
	return config, true
}

func deleteElement(content *cms.Content, path string) bool {
	parentElement, _, key, index := findParentElement(content, path)

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}

func findParentElement(content *cms.Content, path string) (*cms.Content, string, string, int) {
	if path == "" {
		return nil, "", "", 0
	}

	parts := strings.Split(path, ".")
	var elements []cms.Content
	element := content
	var index int
	var key string
	var parentElement *cms.Content
	parentPath := make([]string, 0, len(parts))

	for i, part := range parts {
		parentPath = append(parentPath, part)

		if i%2 != 0 {
			var err error
			index, err = strconv.Atoi(part)

			if err != nil {
				return nil, "", "", 0
			}

			element = &elements[index]
		} else {
			elements = element.Content[part]
			key = part
			parentElement = element
		}
	}

	return parentElement, strings.Join(parentPath, "."), key, index
}

func findElement(content *cms.Content, path string) *cms.Content {
	parentElement, _, key, index := findParentElement(content, path)

	if parentElement == nil {
		return nil
	}

	return &parentElement.Content[key][index]
}

func setElement(content *cms.Content, path string, element *cms.Content) bool {
	parentElement, _, key, index := findParentElement(content, path)

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

	c, err := os.ReadFile(path)

	if err != nil {
		slog.Error("Error reading reference file", "error", err)
		return nil, err
	}

	var element cms.Content

	if err := json.Unmarshal(c, &element); err != nil {
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
