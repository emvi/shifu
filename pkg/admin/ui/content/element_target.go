package content

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cms"
)

const maxRefDepth = 16

type elementTarget struct {
	root   *cms.Content
	file   string
	ref    string
	prefix string
	path   string
}

func (t *elementTarget) element() *cms.Content {
	if t.path == "" {
		return t.root
	}

	return findElement(t.root, t.path)
}

func (t *elementTarget) fullPath(path string) string {
	return joinElementPath(t.prefix, path)
}

func (t *elementTarget) isRef() bool {
	return t.ref != ""
}

func (t *elementTarget) save() error {
	if t.isRef() {
		if err := shared.SavePage(t.root, t.file); err != nil {
			return err
		}

		content.SetReference(t.ref, t.root)
		return nil
	}

	if t.file == "" {
		return errors.New("the document cannot be saved")
	}

	return shared.SavePage(t.root, t.file)
}

func resolveElementTarget(page *cms.Content, elementPath string, refRoot bool) (*elementTarget, error) {
	target := &elementTarget{
		root: page,
		file: page.File,
		path: elementPath,
	}

	if elementPath == "" {
		return target, nil
	}

	parts := strings.Split(elementPath, ".")

	if len(parts)%2 != 0 {
		return nil, errors.New("invalid element path")
	}

	start := 0
	depth := 0

	for i := 0; i < len(parts); i += 2 {
		index, err := strconv.Atoi(parts[i+1])

		if err != nil {
			return nil, err
		}

		elements, found := target.root.Content[parts[i]]

		if !found || index < 0 || index >= len(elements) {
			return nil, errors.New("element not found")
		}

		element := elements[index]
		last := i+2 == len(parts)

		if element.Ref != "" && (!last || refRoot) {
			depth++

			if depth > maxRefDepth {
				return nil, errors.New("references nested too deep")
			}

			ref, err := loadRef(element.Ref)

			if err != nil {
				return nil, err
			}

			target.root = ref
			target.file = ref.File
			target.ref = element.Ref
			start = i + 2
			target.prefix = strings.Join(parts[:start], ".")
			break
		} else {
			target.root = &element
		}
	}

	// reset target if we couldn't find a reference
	if target.ref == "" {
		return &elementTarget{
			root: page,
			file: page.File,
			path: elementPath,
		}, nil
	}

	target.path = strings.Join(parts[start:], ".")
	return target, nil
}

func joinElementPath(prefix, path string) string {
	if prefix == "" {
		return path
	}

	if path == "" {
		return prefix
	}

	return fmt.Sprintf("%s.%s", prefix, path)
}

func refContains(name, search string) bool {
	return refContainsDepth(name, search, 0)
}

func refContainsDepth(name, search string, depth int) bool {
	if strings.EqualFold(name, search) {
		return true
	}

	if depth > maxRefDepth {
		return true
	}

	ref, err := loadRef(name)

	if err != nil {
		return false
	}

	return contentContainsRef(ref, search, depth+1)
}

func contentContainsRef(element *cms.Content, search string, depth int) bool {
	for _, elements := range element.Content {
		for i := range elements {
			if elements[i].Ref != "" {
				if refContainsDepth(elements[i].Ref, search, depth) {
					return true
				}
			} else if contentContainsRef(&elements[i], search, depth) {
				return true
			}
		}
	}

	return false
}
