package cms

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrMatcherBrackets = "matching route '%s' must end with a closing bracket"
	ErrMatcherEmpty    = "matching route '%s' must not be empty"
	ErrMatcherVariable = "matching route '%s' must have a variable name (name:expression)"
	ErrParsingMatcher  = "error parsing matching route '%s': %v"
)

// Route matches raw or matching routes including regular expressions.
type Route struct {
	raw     string
	match   []Subroute
	isRaw   bool
	length  int
	content Content
}

// Subroute is a part of a route separated by slashes.
type Subroute struct {
	path     string
	regex    *regexp.Regexp
	variable string
}

// NewRoute parses and returns a route matcher for the given raw path.
func NewRoute(raw string, content Content) (*Route, error) {
	isRaw := true
	matcher := make([]Subroute, 0)

	if strings.Contains(raw, "{") || strings.Contains(raw, "}") {
		isRaw = false
		parts := strings.Split(raw, "/")

		for _, part := range parts {
			if strings.HasPrefix(part, "{") {
				if !strings.HasSuffix(part, "}") {
					return nil, fmt.Errorf(ErrMatcherBrackets, raw)
				}

				substr := part[1 : len(part)-1]

				if substr == "" {
					return nil, fmt.Errorf(ErrMatcherEmpty, raw)
				}

				name, expr, found := strings.Cut(substr, ":")

				if !found {
					name = strings.TrimSpace(substr)

					if name == "" {
						return nil, fmt.Errorf(ErrMatcherVariable, raw)
					}
				}

				var regex *regexp.Regexp

				if expr != "" {
					var err error
					regex, err = regexp.Compile(fmt.Sprintf("^%s$", expr))

					if err != nil {
						return nil, fmt.Errorf(ErrParsingMatcher, raw, err)
					}
				}

				matcher = append(matcher, Subroute{
					regex:    regex,
					variable: name,
				})
			} else {
				matcher = append(matcher, Subroute{path: part})
			}
		}
	}

	return &Route{
		raw:     raw,
		match:   matcher,
		isRaw:   isRaw,
		length:  len(matcher),
		content: content,
	}, nil
}

func (r *Route) Match(path string) (map[string]string, bool) {
	if r.isRaw {
		return nil, r.raw == path
	}

	parts := strings.Split(path, "/")

	if len(parts) != r.length {
		return nil, false
	}

	vars := make(map[string]string)

	for i, part := range parts {
		if r.match[i].regex != nil {
			if !r.match[i].regex.MatchString(part) {
				return nil, false
			}

			vars[r.match[i].variable] = part
		} else if r.match[i].path != part {
			return nil, false
		}
	}

	return vars, true
}
