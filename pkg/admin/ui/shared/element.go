package shared

import (
	"github.com/emvi/shifu/pkg/cms"
	"strconv"
	"strings"
)

// FindElement finds an element by JSON path.
func FindElement(content *cms.Content, path string) *cms.Content {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	var elements []cms.Content
	element := content

	for i, part := range parts {
		if i%2 != 0 {
			index, err := strconv.Atoi(part)

			if err != nil {
				return nil
			}

			element = &elements[index]
		} else {
			elements = element.Content[part]
		}
	}

	return element
}

// DeleteElement finds an element by JSON path and removes it from the tree.
func DeleteElement(content *cms.Content, path string) bool {
	if path == "" {
		return false
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
				return false
			}

			element = &elements[index]
		} else {
			elements = element.Content[part]
			key = part
			parentElement = element
		}
	}

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}
