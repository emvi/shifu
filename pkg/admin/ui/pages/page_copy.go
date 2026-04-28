package pages

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/emvi/shifu/pkg/admin/tpl"
	"github.com/emvi/shifu/pkg/admin/ui"
	"github.com/emvi/shifu/pkg/admin/ui/shared"
)

// CopyPageData is the data for the copy form.
type CopyPageData struct {
	Lang   string
	Path   string
	Name   string
	Errors map[string]string
}

// CopyPage copies a page.
func CopyPage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	fullPath := getPagePath(path)

	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		name := strings.TrimSpace(r.FormValue("name"))
		errs := make(map[string]string)

		if err := validatePageName(name, false); err != nil {
			errs["name"] = err.Error()
		}

		if len(errs) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			lang := tpl.GetUILanguage(r)
			tpl.Get().Execute(w, "pages-page-copy-form.html", CopyPageData{
				Lang:   lang,
				Path:   path,
				Name:   name,
				Errors: errs,
			})
			return
		}

		// load the existing page
		page, err := shared.LoadPage(r, fullPath, "")

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for k := range page.Path {
			if strings.HasSuffix(page.Path[k], "/") {
				page.Path[k] = page.Path[k] + "copy"
			} else {
				page.Path[k] = page.Path[k] + "/copy"
			}
		}

		// write to a new file
		outPath := getPagePath(filepath.Join(filepath.Dir(path), name+".json"))

		if err := shared.SavePage(page, outPath); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("HX-Reswap", "innerHTML")
		tpl.Get().Execute(w, "pages-tree.html", struct {
			Lang    string
			Entries []Entry
		}{
			Lang:    tpl.GetUILanguage(r),
			Entries: listEntries(w),
		})
		go content.Update()
		return
	}

	lang := tpl.GetUILanguage(r)
	tpl.Get().Execute(w, "pages-page-copy.html", struct {
		WindowOptions ui.WindowOptions
		CopyPageData
	}{
		WindowOptions: ui.WindowOptions{
			ID:         "shifu-pages-page-copy",
			TitleTpl:   "pages-page-copy-window-title",
			ContentTpl: "pages-page-copy-window-content",
			Overlay:    true,
			Lang:       lang,
		},
		CopyPageData: CopyPageData{
			Lang:   lang,
			Path:   path,
			Name:   getPageName(filepath.Base(path)),
			Errors: make(map[string]string),
		},
	})
}
