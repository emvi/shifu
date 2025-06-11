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

			// login
			"login_meta_title":       "Shifu Login",
			"login_meta_description": "Sign in to Shifu CMS.",
			"login_headline":         "Login",
			"login_copyright":        "All rights reserved.",
			"login_website_link":     "Visit Website",
			"login_form_email":       "Email",
			"login_form_password":    "Password",
			"login_form_stay":        "Keep me logged in",
			"login_form_submit":      "Sign in",

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

			// login
			"login_meta_title":       "Shifu Anmeldung",
			"login_meta_description": "Bei Shifu CMS anmelden.",
			"login_headline":         "Anmeldung",
			"login_copyright":        "Alle Rechte vorbehalten.",
			"login_website_link":     "Website besuchen",
			"login_form_email":       "Email",
			"login_form_password":    "Passwort",
			"login_form_stay":        "Angemeldet bleiben",
			"login_form_submit":      "Anmelden",

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
