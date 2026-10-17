package content

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/emvi/shifu/pkg/admin"
	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
	"github.com/emvi/shifu/pkg/cfg"
)

// Reference renders the element tree to edit a reference.
// It reuses the page tree, with the reference being the only root element of a virtual page.
func Reference(w http.ResponseWriter, r *http.Request) {
	path := filepath.ToSlash(filepath.Join(admin.RefsDir, filepath.Clean("/"+r.URL.Query().Get("path"))))
	page, err := loadDocument(r, getPagePath(path), shared.GetLanguage(r))

	if err != nil || !isVirtualDocument(page) {
		slog.Warn("Reference not found", "error", err, "path", path)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setTemplateNames(page)
	tpl.Get().Execute(w, "refs-ref.html", PageTree{
		Language:  page.Language,
		Lang:      tpl.GetUILanguage(r),
		Path:      path,
		Page:      page,
		Positions: tplCfgCache.GetPositions(),
	})
}

func getReferencePath(path string) string {
	return filepath.Join(cfg.Get().BaseDir, admin.RefsDir, path)
}
