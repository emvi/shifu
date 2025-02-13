package tpl

import (
	"github.com/emvi/shifu/pkg/cfg"
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"config": func() *cfg.Config { return cfg.Get() },
	"path":   func() string { return cfg.Get().UI.Path },
	"year":   func() int { return time.Now().Year() },
}
