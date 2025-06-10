package tpl

import (
	"fmt"
	"github.com/emvi/shifu/pkg/cfg"
	"html/template"
	"strings"
	"time"
)

var funcMap = template.FuncMap{
	"config":         func() *cfg.Config { return cfg.Get() },
	"path":           func() string { return cfg.Get().UI.Path },
	"year":           func() int { return time.Now().Year() },
	"formatFileSize": formatFileSize,
	"dict":           dict,
	"fmt":            fmt.Sprintf,
	"replace":        strings.ReplaceAll,
	"i18n":           getTranslation,
}

func formatFileSize(size int64) string {
	value := float64(size)

	if size >= 1024*1024*1024 {
		return fmt.Sprintf("%.1f GB", value/(1024*1024*1024))
	} else if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", value/(1024*1024))
	} else if size >= 1024 {
		return fmt.Sprintf("%.1f KB", value/1024)
	}

	return fmt.Sprintf("%.1f B", value)
}

func dict(v ...any) map[string]any {
	dict := map[string]any{}
	lenv := len(v)

	for i := 0; i < lenv; i += 2 {
		key := strval(v[i])

		if i+1 >= lenv {
			dict[key] = ""
			continue
		}

		dict[key] = v[i+1]
	}

	return dict
}

func strval(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
