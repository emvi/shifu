package tpl

import (
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"html/template"
	"time"
)

var funcMap = template.FuncMap{
	"config":         func() *cfg.Config { return cfg.Get() },
	"path":           func() string { return cfg.Get().UI.Path },
	"year":           func() int { return time.Now().Year() },
	"formatFileSize": formatFileSize,
}

func formatFileSize(size int64) string {
	value := float64(size)

	if size >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", value/1024*1024*1024)
	} else if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", value/1024*1024)
	} else if size >= 1024 {
		return fmt.Sprintf("%.1f KB", value/1024)
	}

	return fmt.Sprintf("%.1f B", value)
}
