package content

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// MoveElement moves an element to a new position.
func MoveElement(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	elementPath := strings.TrimSpace(r.URL.Query().Get("element"))
	direction := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("direction")))
	fullPath := getPagePath(path)
	page, err := loadDocument(r, fullPath, shared.GetLanguage(r))

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	d := 0

	if direction == "up" {
		d = -1
	} else if direction == "down" {
		d = 1
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	target, err := resolveElementTarget(page, elementPath, false)

	if err != nil {
		slog.Debug("Error resolving element", "error", err, "path", elementPath)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if moveElement(target.Root, target.Path, d) {
		if err := target.Save(); err != nil {
			slog.Error("Error while saving page", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	setTemplateNames(page)
	tpl.Get().Execute(w, "page-tree.html", PageTree{
		Language:         shared.GetLanguage(r),
		Lang:             tpl.GetUILanguage(r),
		Path:             path,
		Page:             page,
		Positions:        tplCfgCache.GetPositions(),
		MoveElement:      elementPath,
		ElementDirection: direction,
	})
	go content.Update()
}
