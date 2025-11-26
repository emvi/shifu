package content

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
)

// Reference renders the reference JSON editor.
func Reference(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	fullPath := getReferencePath(path)
	saved := false
	ref, err := os.ReadFile(fullPath)

	if err != nil {
		slog.Warn("Reference not found", "error", err, "path", path)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == http.MethodPost {
		data := r.FormValue("json")
		var reference cms.Content

		if err := json.Unmarshal([]byte(data), &reference); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := shared.SavePage(&reference, fullPath); err != nil {
			slog.Error("Error while saving reference", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		go content.Update()
		saved = true
		ref = []byte(data)
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "refs-ref.html", struct {
		Lang  string
		Path  string
		JSON  string
		Saved bool
	}{
		Lang:  lang,
		Path:  path,
		JSON:  string(ref),
		Saved: saved,
	})
}

func getReferencePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, refsDir, path)
}
