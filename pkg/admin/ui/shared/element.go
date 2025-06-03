package shared

import (
	"github.com/emvi/shifu/pkg/cms"
	"strconv"
	"strings"
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
	parentElement, key, index := FindElement(content, path)

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
	parentElement, key, index := FindElement(content, path)

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}

// FindElement finds an element by JSON path.
// The path is a dot-separated list of keys and indices.
func FindElement(content *cms.Content, path string) (*cms.Content, string, int) {
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
