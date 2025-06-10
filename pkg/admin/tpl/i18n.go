package tpl

import (
	"github.com/emvi/shifu/pkg/cms"
	"net/http"
)

var (
	i18n = map[string]map[string]string{
		"en": {
			// window
			"close_window": "Close window",

			// toolbar
			"edit_page": "Edit page",
			"pages":     "Pages",
			"media":     "Media",
			"database":  "Database",
			"user":      "User",
			"sign_out":  "Sign out",
		},
		"de": {
			// window
			"close_window": "Fenster schließen",

			// toolbar
			"edit_page": "Seite bearbeiten",
			"pages":     "Seiten",
			"media":     "Medien",
			"database":  "Datenbank",
			"user":      "Nutzer",
			"sign_out":  "Abmelden",
		},
	}
)

func getTranslation(language, key string) string {
	kv, found := i18n[language]

	if found {
		return kv[key]
	}

	return i18n["en"][key]
}

// GetLanguage returns the accepted language or "en" by default.
func GetLanguage(r *http.Request) string {
	languages := cms.GetAcceptedLanguages(r)

	if len(languages) == 0 {
		return "en"
	}

	return languages[0]
}
