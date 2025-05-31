package shared

import (
	"github.com/emvi/shifu/pkg/cms"
	"strconv"
	"strings"
)

// MoveElement finds an element by JSON path and moves it up or down.
func MoveElement(content *cms.Content, path string, direction int) bool {
	parentElement, key, index := findElement(content, path)

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
	parentElement, key, index := findElement(content, path)

	if parentElement != nil {
		parentElement.Content[key] = append(parentElement.Content[key][:index], parentElement.Content[key][index+1:]...)
		return true
	}

	return false
}

func findElement(content *cms.Content, path string) (*cms.Content, string, int) {
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
