package cms

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/emvi/shifu/pkg/admin/db"
	"github.com/emvi/shifu/pkg/admin/model"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"log/slog"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	tt "text/template"
	"time"
)

var (
	defaultFuncMap = template.FuncMap{
		"config":        func() *cfg.Config { return cfg.Get() },
		"hostname":      func() string { return cfg.Get().Server.Hostname },
		"copy":          getCopy,
		"get":           get,
		"markdown":      renderMarkdown,
		"markdownBlock": renderMarkdownBlock,
		"int":           func(v string) int { i, _ := strconv.Atoi(v); return i },
		"uint64":        func(i int) uint64 { return uint64(i) },
		"shuffle":       shuffle,
		"fmt":           fmt.Sprintf,
		"dict":          dict,
		"default":       defaultVar,
		"year":          func() int { return time.Now().Year() },
		"formatFloat":   formatFloat,
		"formatInt":     formatInt,
		"formatDate":    func(d time.Time, layout string) string { return d.Format(layout) },
		"gtFloat":       func(a, b float64) bool { return a > b },
		"ltFloat":       func(a, b float64) bool { return a < b },
		"html":          func(str string) template.HTML { return template.HTML(str) },
		"htmlAttr":      func(str string) template.HTMLAttr { return template.HTMLAttr(str) },
		"loggedIn":      isLoggedIn,
		"adminHead":     adminHead,
		"adminBody":     adminBody,
	}
)

// Merge merges FuncMaps.
func Merge(maps ...template.FuncMap) template.FuncMap {
	out := make(map[string]any)

	for k, v := range sprig.FuncMap() {
		out[k] = v
	}

	for k, v := range defaultFuncMap {
		out[k] = v
	}

	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				out[k] = v
			}
		}
	}

	return out
}

func getCopy(page, content *Content, key string) any {
	kv, ok := content.Copy[page.Language]

	if !ok {
		kv, ok = content.Copy["en"]

		if !ok {
			return nil
		}
	}

	value, ok := kv[key]

	if !ok {
		kv, ok = content.Copy["en"]

		if ok {
			value, ok = kv[key]

			if ok {
				return value
			}
		}

		return nil
	}

	return value
}

func get(content *Content, key string) any {
	value, ok := content.Data[key]

	if !ok {
		return nil
	}

	return value
}

func renderMarkdown(file string, data any) template.HTML {
	content, err := os.ReadFile(filepath.Join(cfg.Get().BaseDir, file))

	if err != nil {
		slog.Error("Error loading markdown file", "file", file, "error", err)
		return ""
	}

	return renderMarkdownContent(file, string(content), "", data)
}

func renderMarkdownBlock(file, block string, data any) template.HTML {
	content, err := os.ReadFile(filepath.Join(cfg.Get().BaseDir, file))

	if err != nil {
		slog.Error("Error loading markdown file", "file", file, "error", err)
		return ""
	}

	return renderMarkdownContent(file, string(content), block, data)
}

func renderMarkdownContent(file, content, block string, data any) template.HTML {
	tpl, err := tt.New("").Funcs(cfg.Get().FuncMap).Parse(content)

	if err != nil {
		slog.Error("Error parsing markdown file", "file", file, "error", err)
		return ""
	}

	var buffer, out bytes.Buffer
	converter := goldmark.New(
		goldmark.WithExtensions(
			extension.NewFootnote(),
			extension.NewTable(),
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	if block != "" {
		if _, err := tpl.Parse(fmt.Sprintf(`{{template "%s" .}}`, block)); err != nil {
			slog.Error("Error parsing markdown block", "block", block, "file", file, "error", err)
			return ""
		}

		if err := tpl.Execute(&buffer, data); err != nil {
			slog.Error("Error rendering markdown file", "file", file, "error", err)
			return ""
		}

		if err := converter.Convert(buffer.Bytes(), &out); err != nil {
			panic(err)
		}

		return template.HTML(out.String())
	}

	if err := tpl.Execute(&buffer, data); err != nil {
		slog.Error("Error rendering markdown file", "file", file, "error", err)
		return ""
	}

	if err := converter.Convert(buffer.Bytes(), &out); err != nil {
		panic(err)
	}

	return template.HTML(out.String())
}

func shuffle(list []any, n int) []any {
	listCopy := make([]any, len(list))
	copy(listCopy, list)

	rand.Shuffle(len(listCopy), func(i, j int) {
		listCopy[i], listCopy[j] = listCopy[j], listCopy[i]
	})

	if n > 0 && len(listCopy) > n {
		listCopy = listCopy[:n]
	}

	return listCopy
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

func defaultVar(vars ...any) any {
	for _, v := range vars {
		if v != nil {
			return v
		}
	}

	return nil
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

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatInt(i int) string {
	str := strconv.Itoa(i)

	if i < 1000 {
		return str
	}

	out := ""

	for i := len(str); i > 0; i -= 3 {
		start := i - 3

		if start < 0 {
			start = 0
		}

		out = str[start:i] + "," + out
	}

	return out[:len(out)-1]
}

func isLoggedIn(page *Content) bool {
	cookie, err := page.Request.Cookie("session")

	if err != nil || cookie.Value == "" {
		return false
	}

	s := new(model.Session)

	if err := db.Get().Get(s, `SELECT * FROM "session" WHERE session = ?`, cookie.Value); err != nil {
		return false
	}

	if s == nil || s.Expires.Before(time.Now()) {
		return false
	}

	return true
}

func adminHead(page *Content) template.HTML {
	if isLoggedIn(page) {
		path := cfg.Get().UI.Path
		return template.HTML(fmt.Sprintf(`<link rel="prefetch" href="%s/static/fonts/Inter-Medium.woff2" as="font" type="font/woff2" />
			<link rel="prefetch" href="%s/static/fonts/Inter-Regular.woff2" as="font" type="font/woff2" />
			<link rel="prefetch" href="%s/static/fonts/InterDisplay-Medium.woff2" as="font" type="font/woff2" />
			<link rel="stylesheet" type="text/css" href="%s/static/admin.css" />
			<link rel="stylesheet" type="text/css" href="%s/static/trix/trix.css" />
			<link rel="stylesheet" type="text/css" href="%s/static/jsoneditor/jsoneditor.css" />
			<script defer src="%s/static/trix/trix.min.js"></script>
			<script defer src="%s/static/jsoneditor/jsoneditor.js"></script>
			<script defer src="%s/static/htmx.min.js"></script>
			<script defer src="%s/static/htmx-ext-response-targets.min.js"></script>
			<script defer src="%s/static/admin.js"></script>`, path, path, path, path, path, path, path, path, path, path, path))
	}

	return ""
}

func adminBody(page *Content) template.HTML {
	if isLoggedIn(page) {
		path := cfg.Get().UI.Path
		return template.HTML(fmt.Sprintf(`<div hx-get="%s/toolbar?path=%s&language=%s"
			hx-swap="outerHTML"
			hx-trigger="load"></div>`, path, page.File, page.Language))
	}

	return ""
}
