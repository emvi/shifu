package content

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/cfg"
)

// Entry is a reference.
type Entry struct {
	Name string
	Path string
}

// References renders the references management dialog.
func References(w http.ResponseWriter, r *http.Request) {
	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "refs.html", struct {
		WindowOptions ui.WindowOptions
		Lang          string
		Entries       []Entry
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-refs",
			TitleTpl:   "refs-window-title",
			ContentTpl: "refs-window-content",
			MinWidth:   900,
			Lang:       lang,
		},
		Lang:    lang,
		Entries: listReferences(w),
	})
}

func listReferences(w http.ResponseWriter) []Entry {
	dir := filepath.Join(cfg.Get().BaseDir, refsDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			slog.Error("Error creating references directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return nil
		}
	}

	tree, err := readReferences(dir)

	if err != nil {
		slog.Error("Error reading references directory", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	return tree
}

func readReferences(dir string) ([]Entry, error) {
	files, err := os.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	refs := make([]Entry, 0)

	for _, file := range files {
		refs = append(refs, Entry{
			Name: getReferenceName(file.Name()),
			Path: strings.TrimPrefix(filepath.Join(dir, file.Name()), dir),
		})
	}

	sortEntries(refs)
	return refs, nil
}

func sortEntries(entries []Entry) {
	slices.SortFunc(entries, func(a, b Entry) int {
		nameA := strings.ToLower(a.Name)
		nameB := strings.ToLower(b.Name)

		if nameA > nameB {
			return 1
		} else if nameA < nameB {
			return -1
		}

		return 0
	})
}

func getReferenceName(filename string) string {
	name, _, _ := strings.Cut(filename, ".")
	return strings.TrimPrefix(name, "/")
}
