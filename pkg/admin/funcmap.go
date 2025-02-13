package admin

import (
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"year": func() int { return time.Now().Year() },
}
