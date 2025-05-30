package content

import (
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/emvi/shifu/pkg/cms"
	"net/http"
	"path/filepath"
)

// Page renders the page editing dialog.
func Page(w http.ResponseWriter, r *http.Request) {
	path := getPagePath(r.URL.Query().Get("path"))
	page, err := shared.LoadPage(path)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tpl.Get().Execute(w, "page.html", struct {
		WindowOptions ui.WindowOptions
		Page          *cms.Content
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-page",
			TitleTpl:   "page-window-title",
			ContentTpl: "page-window-content",
		},
		Page: page,
	})
}

func getPagePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, path)
}
