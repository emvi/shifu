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
	// Root is the document to modify (either the page or a reference).
	Root *cms.Content

	// File is the path to the file of Root.
	File string

	// Ref is the name of the reference if Root is a reference, empty for pages.
	Ref string

	// Prefix is the position of the reference element on the page (empty for pages).
	Prefix string

	// Path is the element path relative to Root (empty if the element is Root itself).
	Path string
}

// Element returns the element the target points to.
func (t *elementTarget) Element() *cms.Content {
	if t.Path == "" {
		return t.Root
	}

	return findElement(t.Root, t.Path)
}

// FullPath returns the position on the page for given path relative to Root.
func (t *elementTarget) FullPath(path string) string {
	return joinElementPath(t.Prefix, path)
}

// IsRef returns true if the target document is a reference.
func (t *elementTarget) IsRef() bool {
	return t.Ref != ""
}

// Save writes the target document to disk.
// References are updated in memory as well, so that they can be rendered right away.
func (t *elementTarget) Save() error {
	if t.IsRef() {
		if err := saveRef(t.Root, t.File); err != nil {
			return err
		}

		content.SetReference(t.Ref, t.Root)
		return nil
	}

	return shared.SavePage(t.Root, t.File)
}

func resolveElementTarget(page *cms.Content, elementPath string, refRoot bool) (*elementTarget, error) {
	target := &elementTarget{
		Root: page,
		File: page.File,
		Path: elementPath,
	}

	if elementPath == "" {
		return target, nil
	}

	parts := strings.Split(elementPath, ".")

	if len(parts)%2 != 0 {
		return nil, errors.New("invalid element path")
	}

	start := 0 // index into parts where the path relative to the current root starts
	depth := 0

	for i := 0; i < len(parts); i += 2 {
		index, err := strconv.Atoi(parts[i+1])

		if err != nil {
			return nil, err
		}

		elements, found := target.Root.Content[parts[i]]

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

			target.Root = ref
			target.File = ref.File
			target.Ref = element.Ref
			start = i + 2
			target.Prefix = strings.Join(parts[:start], ".")
		}
	}

	target.Path = strings.Join(parts[start:], ".")
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
