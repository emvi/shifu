package tpl

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	tt "text/template"
)

var (
	defaultFuncMap = template.FuncMap{
		"config":        func() *cfg.Config { return cfg.Get() },
		"markdown":      renderMarkdown,
		"markdownBlock": renderMarkdownBlock,
		"int":           func(v string) int { i, _ := strconv.Atoi(v); return i },
		"uint64":        func(i int) uint64 { return uint64(i) },
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
